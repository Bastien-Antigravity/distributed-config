package main

// cli:allow-print

// =============================================================================
// ESSENTIAL PROCESS:
// Pure-Go RSA-2048 X.509 Fleet Root CA and per-service leaf TLS certificate
// generator with zero external OpenSSL dependency.
//
// DATA FLOW:
// 1. Input: CLI flags (--dir, --ca, --service, --all, --hosts, --days, --force).
// 2. Logic: Generates distinct RSA-2048 keypairs, constructs X.509 CA and SAN-rich
//    leaf templates, signs leaf certificates with ca.crt/ca.key, and persists PEMs.
// 3. Output: PEM-encoded ca.crt/ca.key and <service>.crt/<service>.key files on disk.
//
// KEY PARAMETERS:
// - --dir: Output directory outside Git for TLS certificates and keys (mandatory).
// - --ca / --service / --all: Certificate generation target flags.
// - --hosts / --days / --force: Extra SANs, validity days, and overwrite toggle.
// =============================================================================

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/spf13/pflag"
)

// -----------------------------------------------------------------------------

const (
	defaultCADays   = 3650
	defaultLeafDays = 365
	rsaKeyBits      = 2048
)

var defaultFleetTLSServices = []string{
	"tele-remote",
	"config-server",
	"notif-server",
	"log-server",
	"watchdog-agent",
	"web-interface",
	"rag-engine",
	"squad-agent",
}

// -----------------------------------------------------------------------------

func handleTLSGen() {
	flags := pflag.NewFlagSet("tlsgen", pflag.ExitOnError)
	dir := flags.String("dir", "", "Output directory outside Git for TLS certificates and keys (required)")
	genCA := flags.Bool("ca", false, "Generate Fleet Root CA (ca.crt and ca.key)")
	service := flags.String("service", "", "Service name to generate a signed leaf cert/key for")
	genAll := flags.Bool("all", false, "Generate CA (if missing) and leaf certs/keys for all default fleet services")
	hosts := flags.String("hosts", "", "Comma-separated extra SAN IPs or DNS hostnames")
	days := flags.Int("days", 0, "Validity override in days (CA defaults to 3650, leaf defaults to 365)")
	force := flags.Bool("force", false, "Overwrite existing certificate and key files")

	if err := flags.Parse(os.Args[2:]); err != nil {
		fmt.Printf("Error parsing flags: %v\n", err)
		os.Exit(1)
	}

	outDir := strings.TrimSpace(*dir)
	svcName := strings.TrimSpace(*service)
	if outDir == "" || (!*genCA && svcName == "" && !*genAll) {
		fmt.Println("Error: --dir is required, and at least one of --ca, --service, or --all must be specified.")
		flags.Usage()
		os.Exit(1)
	}

	outDir = expandHomePrefix(outDir)
	extraHosts := splitHosts(*hosts)

	fmt.Println("----------------------------------------------------------------")
	if *genAll || *genCA {
		certPath, keyPath, created, err := generateFleetCA(outDir, *days, *force)
		if err != nil {
			fmt.Printf("Error generating Fleet Root CA: %v\n", err)
			os.Exit(1)
		}
		printCertResult("ca", certPath, keyPath, created)
	}

	if *genAll {
		for _, fleetSvc := range defaultFleetTLSServices {
			certPath, keyPath, created, err := generateServiceCert(outDir, fleetSvc, extraHosts, *days, *force)
			if err != nil {
				fmt.Printf("Error generating certificate for %s: %v\n", fleetSvc, err)
				os.Exit(1)
			}
			printCertResult(fleetSvc, certPath, keyPath, created)
		}
	}

	if svcName != "" && (!*genAll || !slices.Contains(defaultFleetTLSServices, svcName)) {
		certPath, keyPath, created, err := generateServiceCert(outDir, svcName, extraHosts, *days, *force)
		if err != nil {
			fmt.Printf("Error generating certificate for %s: %v\n", svcName, err)
			os.Exit(1)
		}
		printCertResult(svcName, certPath, keyPath, created)
	}
	fmt.Println("----------------------------------------------------------------")
}

// -----------------------------------------------------------------------------

func generateFleetCA(dir string, days int, force bool) (certPath, keyPath string, created bool, err error) {
	if strings.TrimSpace(dir) == "" {
		return "", "", false, fmt.Errorf("output directory must not be empty")
	}
	if days <= 0 {
		days = defaultCADays
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", "", false, fmt.Errorf("failed to create TLS directory %s: %w", dir, err)
	}

	certPath = filepath.Join(dir, "ca.crt")
	keyPath = filepath.Join(dir, "ca.key")

	certExists := fileExists(certPath)
	keyExists := fileExists(keyPath)
	if !force {
		if certExists && keyExists {
			return certPath, keyPath, false, nil
		}
		if certExists || keyExists {
			return "", "", false, fmt.Errorf("incomplete CA state in %s: pass --force to regenerate", dir)
		}
	}

	privKey, err := rsa.GenerateKey(rand.Reader, rsaKeyBits)
	if err != nil {
		return "", "", false, fmt.Errorf("failed to generate CA RSA key: %w", err)
	}

	serial, err := randomSerialNumber()
	if err != nil {
		return "", "", false, err
	}

	now := time.Now().UTC()
	tmpl := &x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			CommonName:   "Bastien-Antigravity Fleet Root CA",
			Organization: []string{"Bastien-Antigravity"},
		},
		NotBefore:             now.Add(-5 * time.Minute),
		NotAfter:              now.Add(time.Duration(days) * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
		IsCA:                  true,
		SubjectKeyId:          subjectKeyID(&privKey.PublicKey),
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &privKey.PublicKey, privKey)
	if err != nil {
		return "", "", false, fmt.Errorf("failed to create CA certificate: %w", err)
	}

	if err := writePEMFile(certPath, "CERTIFICATE", derBytes, 0644); err != nil {
		return "", "", false, err
	}
	if err := writePEMFile(keyPath, "RSA PRIVATE KEY", x509.MarshalPKCS1PrivateKey(privKey), 0600); err != nil {
		return "", "", false, err
	}

	return certPath, keyPath, true, nil
}

// -----------------------------------------------------------------------------

func generateServiceCert(dir, serviceName string, extraHosts []string, days int, force bool) (certPath, keyPath string, created bool, err error) {
	if strings.TrimSpace(dir) == "" {
		return "", "", false, fmt.Errorf("output directory must not be empty")
	}
	if err := validateServiceName(serviceName); err != nil {
		return "", "", false, err
	}
	if days <= 0 {
		days = defaultLeafDays
	}

	caCert, caKey, err := loadFleetCA(dir)
	if err != nil {
		return "", "", false, err
	}

	certPath = filepath.Join(dir, serviceName+".crt")
	keyPath = filepath.Join(dir, serviceName+".key")

	certExists := fileExists(certPath)
	keyExists := fileExists(keyPath)
	if !force {
		if certExists && keyExists {
			return certPath, keyPath, false, nil
		}
		if certExists || keyExists {
			return "", "", false, fmt.Errorf("incomplete certificate state for %s in %s: pass --force to regenerate", serviceName, dir)
		}
	}

	leafKey, err := rsa.GenerateKey(rand.Reader, rsaKeyBits)
	if err != nil {
		return "", "", false, fmt.Errorf("failed to generate RSA key for %s: %w", serviceName, err)
	}

	serial, err := randomSerialNumber()
	if err != nil {
		return "", "", false, err
	}

	dnsNames, ipAddrs := buildSANs(serviceName, extraHosts)
	now := time.Now().UTC()
	tmpl := &x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			CommonName:   serviceName,
			Organization: []string{"Bastien-Antigravity"},
		},
		NotBefore:             now.Add(-5 * time.Minute),
		NotAfter:              now.Add(time.Duration(days) * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		IsCA:                  false,
		DNSNames:              dnsNames,
		IPAddresses:           ipAddrs,
		SubjectKeyId:          subjectKeyID(&leafKey.PublicKey),
		AuthorityKeyId:        caCert.SubjectKeyId,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, tmpl, caCert, &leafKey.PublicKey, caKey)
	if err != nil {
		return "", "", false, fmt.Errorf("failed to sign certificate for %s: %w", serviceName, err)
	}

	if err := writePEMFile(certPath, "CERTIFICATE", derBytes, 0644); err != nil {
		return "", "", false, err
	}
	if err := writePEMFile(keyPath, "RSA PRIVATE KEY", x509.MarshalPKCS1PrivateKey(leafKey), 0600); err != nil {
		return "", "", false, err
	}

	return certPath, keyPath, true, nil
}

// -----------------------------------------------------------------------------

func validateServiceName(serviceName string) error {
	if serviceName == "" {
		return fmt.Errorf("service name must not be empty")
	}
	if serviceName == "ca" {
		return fmt.Errorf("service name 'ca' is reserved for the Fleet Root CA")
	}
	for _, r := range serviceName {
		isLower := r >= 'a' && r <= 'z'
		isDigit := r >= '0' && r <= '9'
		isSep := r == '-' || r == '_'
		if !isLower && !isDigit && !isSep {
			return fmt.Errorf("invalid service name %q: must contain only lowercase alphanumeric characters, '-' or '_'", serviceName)
		}
	}
	return nil
}

func loadFleetCA(dir string) (*x509.Certificate, *rsa.PrivateKey, error) {
	caCertPath := filepath.Join(dir, "ca.crt")
	caKeyPath := filepath.Join(dir, "ca.key")

	certPEM, err := os.ReadFile(caCertPath)
	if err != nil {
		return nil, nil, fmt.Errorf("missing or unreadable Fleet Root CA cert %s: %w", caCertPath, err)
	}
	certBlock, _ := pem.Decode(certPEM)
	if certBlock == nil || certBlock.Type != "CERTIFICATE" {
		return nil, nil, fmt.Errorf("invalid PEM block in %s", caCertPath)
	}
	caCert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse CA certificate %s: %w", caCertPath, err)
	}

	keyPEM, err := os.ReadFile(caKeyPath)
	if err != nil {
		return nil, nil, fmt.Errorf("missing or unreadable Fleet Root CA key %s: %w", caKeyPath, err)
	}
	keyBlock, _ := pem.Decode(keyPEM)
	if keyBlock == nil {
		return nil, nil, fmt.Errorf("invalid PEM block in %s", caKeyPath)
	}

	var caKey *rsa.PrivateKey
	switch keyBlock.Type {
	case "RSA PRIVATE KEY":
		caKey, err = x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
	case "PRIVATE KEY":
		parsed, parseErr := x509.ParsePKCS8PrivateKey(keyBlock.Bytes)
		err = parseErr
		if err == nil {
			var ok bool
			caKey, ok = parsed.(*rsa.PrivateKey)
			if !ok {
				err = fmt.Errorf("CA private key is not RSA")
			}
		}
	default:
		err = fmt.Errorf("unexpected PEM key type %q", keyBlock.Type)
	}
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse CA private key %s: %w", caKeyPath, err)
	}

	return caCert, caKey, nil
}

func buildSANs(serviceName string, extraHosts []string) ([]string, []net.IP) {
	var dnsNames []string
	var ipAddrs []net.IP
	seenDNS := make(map[string]struct{})
	seenIP := make(map[string]struct{})

	addDNS := func(name string) {
		if name == "" {
			return
		}
		if _, ok := seenDNS[name]; ok {
			return
		}
		seenDNS[name] = struct{}{}
		dnsNames = append(dnsNames, name)
	}

	addIP := func(ip net.IP) {
		if ip == nil {
			return
		}
		key := ip.String()
		if _, ok := seenIP[key]; ok {
			return
		}
		seenIP[key] = struct{}{}
		ipAddrs = append(ipAddrs, ip)
	}

	addIP(net.ParseIP("127.0.0.1"))
	addIP(net.ParseIP("127.0.0.2"))
	addIP(net.ParseIP("::1"))

	addDNS("localhost")
	addDNS(serviceName)
	addDNS(strings.ReplaceAll(serviceName, "-", "_"))
	addDNS(strings.ReplaceAll(serviceName, "_", "-"))

	for _, raw := range extraHosts {
		for _, host := range splitHosts(raw) {
			if ip := net.ParseIP(host); ip != nil {
				addIP(ip)
			} else {
				addDNS(host)
			}
		}
	}

	return dnsNames, ipAddrs
}

func splitHosts(raw string) []string {
	var hosts []string
	for _, part := range strings.Split(raw, ",") {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			hosts = append(hosts, trimmed)
		}
	}
	return hosts
}

func randomSerialNumber() (*big.Int, error) {
	limit := new(big.Int).Lsh(big.NewInt(1), 128)
	serial, err := rand.Int(rand.Reader, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to generate certificate serial number: %w", err)
	}
	if serial.Sign() == 0 {
		serial.SetInt64(1)
	}
	return serial, nil
}

func subjectKeyID(pub *rsa.PublicKey) []byte {
	sum := sha256.Sum256(x509.MarshalPKCS1PublicKey(pub))
	return sum[:20]
}

func writePEMFile(path, blockType string, derBytes []byte, perm os.FileMode) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return fmt.Errorf("failed to open %s for writing: %w", path, err)
	}
	defer func() {
		_ = f.Close()
	}()

	if err := pem.Encode(f, &pem.Block{Type: blockType, Bytes: derBytes}); err != nil {
		return fmt.Errorf("failed to encode PEM %s: %w", path, err)
	}
	if err := os.Chmod(path, perm); err != nil {
		return fmt.Errorf("failed to set permissions on %s: %w", path, err)
	}
	return nil
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func expandHomePrefix(path string) string {
	if path == "~" || strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err == nil && home != "" {
			if path == "~" {
				return home
			}
			return filepath.Join(home, path[2:])
		}
	}
	return path
}

func printCertResult(label, certPath, keyPath string, created bool) {
	status := "SKIPPED (already exists)"
	if created {
		status = "CREATED"
	}
	fmt.Printf("[%s] %s\n", label, status)
	fmt.Printf(" - cert: %s\n", certPath)
	fmt.Printf(" - key:  %s\n", keyPath)
}
