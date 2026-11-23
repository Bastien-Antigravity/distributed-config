package main

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"testing"
	"time"
)

func TestTLSGen_CAAndServiceLeafChain(t *testing.T) {
	dir := t.TempDir()

	// Calling generateServiceCert without a CA must fail fast.
	if _, _, _, err := generateServiceCert(dir, "tele-remote", nil, 0, false); err == nil {
		t.Fatalf("expected error when generating leaf cert without CA, got nil")
	}

	// Generate Fleet Root CA.
	caCertPath, caKeyPath, created, err := generateFleetCA(dir, 0, false)
	if err != nil {
		t.Fatalf("generateFleetCA failed: %v", err)
	}
	if !created {
		t.Fatalf("expected initial CA generation to report created=true")
	}

	// Generate two distinct service leaf certs with custom extra hosts.
	teleCertPath, teleKeyPath, created, err := generateServiceCert(
		dir,
		"tele-remote",
		[]string{"10.0.0.5", "tele.example.com"},
		0,
		false,
	)
	if err != nil {
		t.Fatalf("generateServiceCert(tele-remote) failed: %v", err)
	}
	if !created {
		t.Fatalf("expected tele-remote leaf generation to report created=true")
	}

	notifCertPath, notifKeyPath, created, err := generateServiceCert(dir, "notif-server", nil, 0, false)
	if err != nil {
		t.Fatalf("generateServiceCert(notif-server) failed: %v", err)
	}
	if !created {
		t.Fatalf("expected notif-server leaf generation to report created=true")
	}

	// Verify 0600 permissions on all private keys.
	for _, keyPath := range []string{caKeyPath, teleKeyPath, notifKeyPath} {
		info, err := os.Stat(keyPath)
		if err != nil {
			t.Fatalf("stat(%s) failed: %v", keyPath, err)
		}
		if mode := info.Mode().Perm(); mode != 0600 {
			t.Fatalf("expected mode 0600 on %s, got %04o", keyPath, mode)
		}
	}

	// Verify per-service key isolation (distinct RSA moduli).
	caRSA := mustReadRSAPrivateKey(t, caKeyPath)
	teleRSA := mustReadRSAPrivateKey(t, teleKeyPath)
	notifRSA := mustReadRSAPrivateKey(t, notifKeyPath)

	if caRSA.N.Cmp(teleRSA.N) == 0 || caRSA.N.Cmp(notifRSA.N) == 0 || teleRSA.N.Cmp(notifRSA.N) == 0 {
		t.Fatalf("expected distinct RSA moduli across ca, tele-remote, and notif-server keys")
	}

	// Verify both leaf certificates against the CA for ServerAuth and ClientAuth.
	caCert := mustReadCert(t, caCertPath)
	roots := x509.NewCertPool()
	roots.AddCert(caCert)

	teleCert := mustReadCert(t, teleCertPath)
	notifCert := mustReadCert(t, notifCertPath)

	verifyLeaf := func(name string, cert *x509.Certificate, dnsToCheck []string) {
		t.Helper()
		for _, usage := range []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth} {
			for _, dns := range dnsToCheck {
				if _, err := cert.Verify(x509.VerifyOptions{
					Roots:       roots,
					DNSName:     dns,
					CurrentTime: time.Now(),
					KeyUsages:   []x509.ExtKeyUsage{usage},
				}); err != nil {
					t.Fatalf("%s failed x509 verification for dns=%q usage=%v: %v", name, dns, usage, err)
				}
			}
		}
	}

	verifyLeaf("tele-remote", teleCert, []string{
		"localhost",
		"127.0.0.1",
		"tele-remote",
		"tele_remote",
		"10.0.0.5",
		"tele.example.com",
	})
	verifyLeaf("notif-server", notifCert, []string{
		"localhost",
		"127.0.0.1",
		"notif-server",
		"notif_server",
	})

	// Verify idempotency (force=false preserves key bytes and returns created=false).
	origTeleKeyBytes, err := os.ReadFile(teleKeyPath)
	if err != nil {
		t.Fatalf("read tele key failed: %v", err)
	}
	_, _, created, err = generateServiceCert(dir, "tele-remote", []string{"10.0.0.5", "tele.example.com"}, 0, false)
	if err != nil {
		t.Fatalf("idempotent generateServiceCert failed: %v", err)
	}
	if created {
		t.Fatalf("expected created=false on idempotent generateServiceCert call")
	}
	afterTeleKeyBytes, err := os.ReadFile(teleKeyPath)
	if err != nil {
		t.Fatalf("re-read tele key failed: %v", err)
	}
	if string(origTeleKeyBytes) != string(afterTeleKeyBytes) {
		t.Fatalf("expected existing key bytes to be preserved when force=false")
	}

	// Verify force=true rotates the key (created=true and different key bytes).
	_, _, created, err = generateServiceCert(dir, "tele-remote", []string{"10.0.0.5", "tele.example.com"}, 0, true)
	if err != nil {
		t.Fatalf("forced generateServiceCert failed: %v", err)
	}
	if !created {
		t.Fatalf("expected created=true when force=true")
	}
	rotatedTeleKeyBytes, err := os.ReadFile(teleKeyPath)
	if err != nil {
		t.Fatalf("read rotated tele key failed: %v", err)
	}
	if string(origTeleKeyBytes) == string(rotatedTeleKeyBytes) {
		t.Fatalf("expected key rotation when force=true")
	}
}

func TestTLSGen_ValidateServiceName(t *testing.T) {
	dir := t.TempDir()
	if _, _, _, err := generateFleetCA(dir, 0, false); err != nil {
		t.Fatalf("generateFleetCA failed: %v", err)
	}

	invalidNames := []string{"", "../evil", "foo/bar", "foo\\bar", "UpperCase", "ca", "bad.service"}
	for _, name := range invalidNames {
		if _, _, _, err := generateServiceCert(dir, name, nil, 0, false); err == nil {
			t.Fatalf("expected validation error for serviceName=%q, got nil", name)
		}
	}
}

func mustReadCert(t *testing.T, path string) *x509.Certificate {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read cert %s: %v", path, err)
	}
	block, _ := pem.Decode(raw)
	if block == nil {
		t.Fatalf("failed to decode PEM cert %s", path)
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Fatalf("failed to parse x509 cert %s: %v", path, err)
	}
	return cert
}

func mustReadRSAPrivateKey(t *testing.T, path string) *rsa.PrivateKey {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read key %s: %v", path, err)
	}
	block, _ := pem.Decode(raw)
	if block == nil {
		t.Fatalf("failed to decode PEM key %s", path)
	}
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		t.Fatalf("failed to parse RSA key %s: %v", path, err)
	}
	return key
}
