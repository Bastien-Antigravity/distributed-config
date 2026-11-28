package main

// =============================================================================
// ESSENTIAL PROCESS:
// Fleet administrative utility for RSA-2048 keypair generation and plaintext
// token encryption into canonical ENC(...) format for configuration files.
//
// DATA FLOW:
// 1. Input: CLI subcommands ("keygen", "encrypt", "help") and associated flags.
// 2. Logic: Uses crypto/rsa and x509 for keygen; invokes secret.Encrypt for tokens.
// 3. Output: Encrypted ENC(...) strings or public/private PEM keypair files on disk.
//
// KEY PARAMETERS:
// - --dir: Destination directory for generated PEM key files.
// - --key: Path to public key PEM for encryption.
// - --token: Plaintext secret string to be encrypted into ENC(...) format.
// =============================================================================

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Bastien-Antigravity/distributed-config/src/secret"
	"github.com/spf13/pflag"
)

// -----------------------------------------------------------------------------

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	switch command {
	case "keygen":
		handleKeygen()
	case "encrypt":
		handleEncrypt()
	case "tlsgen":
		handleTLSGen()
	case "help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: config-tool <command> [options]")
	fmt.Println("\nCommands:")
	fmt.Println("  keygen   Generate a new RSA key pair")
	fmt.Println("  encrypt  Encrypt a secret token using a public key")
	fmt.Println("  tlsgen   Generate Fleet Root CA and per-service mTLS certificates")
	fmt.Println("  help     Show this help message")
	fmt.Println("\nRun 'config-tool <command> --help' for more information on a command.")
}

func handleKeygen() {
	flags := pflag.NewFlagSet("keygen", pflag.ExitOnError)
	outputDir := flags.String("dir", ".", "Directory to save the generated .pem keys")
	if err := flags.Parse(os.Args[2:]); err != nil {
		fmt.Printf("Error parsing flags: %v\n", err)
		os.Exit(1)
	}

	privPath, pubPath, err := generateRSAKeyPair(*outputDir)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("----------------------------------------------------------------")
	fmt.Println("RSA Key Pair Generated Successfully!")
	fmt.Printf(" - %s (KEEP THIS SECRET)\n", privPath)
	fmt.Printf(" - %s (Non-sensitive, used for encryption)\n", pubPath)
	fmt.Println("----------------------------------------------------------------")
}

// -----------------------------------------------------------------------------

func generateRSAKeyPair(outputDir string) (privPath, pubPath string, err error) {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", "", fmt.Errorf("creating directory: %w", err)
	}

	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", fmt.Errorf("generating key: %w", err)
	}

	privPath = filepath.Join(outputDir, "private.pem")
	privFile, err := os.OpenFile(privPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return "", "", fmt.Errorf("creating private.pem: %w", err)
	}
	defer func() {
		_ = privFile.Close()
	}()

	privPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privKey),
	}
	if err := pem.Encode(privFile, privPEM); err != nil {
		return "", "", fmt.Errorf("encoding private key: %w", err)
	}
	if err := os.Chmod(privPath, 0600); err != nil {
		return "", "", fmt.Errorf("setting permissions on private.pem: %w", err)
	}

	pubPath = filepath.Join(outputDir, "public.pem")
	pubFile, err := os.OpenFile(pubPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return "", "", fmt.Errorf("creating public.pem: %w", err)
	}
	defer func() {
		_ = pubFile.Close()
	}()

	pubBytes, _ := x509.MarshalPKIXPublicKey(&privKey.PublicKey)
	pubPEM := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubBytes,
	}
	if err := pem.Encode(pubFile, pubPEM); err != nil {
		return "", "", fmt.Errorf("encoding public key: %w", err)
	}

	return privPath, pubPath, nil
}

func handleEncrypt() {
	flags := pflag.NewFlagSet("encrypt", pflag.ExitOnError)
	keyPath := flags.String("key", "", "Path to the RSA public key")
	token := flags.String("token", "", "The plaintext token to encrypt")
	if err := flags.Parse(os.Args[2:]); err != nil {
		fmt.Printf("Error parsing flags: %v\n", err)
		os.Exit(1)
	}

	if *keyPath == "" || *token == "" {
		fmt.Println("Error: --key and --token are mandatory.")
		flags.Usage()
		os.Exit(1)
	}

	pubKeyBytes, err := os.ReadFile(*keyPath)
	if err != nil {
		fmt.Printf("Error reading public key: %v\n", err)
		os.Exit(1)
	}

	encrypted, err := secret.Encrypt(*token, string(pubKeyBytes))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("----------------------------------------------------------------")
	fmt.Println("Encrypted Token (Paste this into your config.yaml):")
	fmt.Println(encrypted)
	fmt.Println("----------------------------------------------------------------")
}
