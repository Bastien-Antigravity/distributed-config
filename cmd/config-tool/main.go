package main

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

	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		fmt.Printf("Error creating directory: %v\n", err)
		os.Exit(1)
	}

	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		fmt.Printf("Error generating key: %v\n", err)
		os.Exit(1)
	}

	privPath := filepath.Join(*outputDir, "private.pem")
	privFile, err := os.Create(privPath)
	if err != nil {
		fmt.Printf("Error creating private.pem: %v\n", err)
		os.Exit(1)
	}
	defer privFile.Close()

	privPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privKey),
	}
	if err := pem.Encode(privFile, privPEM); err != nil {
		fmt.Printf("Error encoding private key: %v\n", err)
		os.Exit(1)
	}

	pubPath := filepath.Join(*outputDir, "public.pem")
	pubFile, err := os.Create(pubPath)
	if err != nil {
		fmt.Printf("Error creating public.pem: %v\n", err)
		os.Exit(1)
	}
	defer pubFile.Close()

	pubBytes, _ := x509.MarshalPKIXPublicKey(&privKey.PublicKey)
	pubPEM := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubBytes,
	}
	if err := pem.Encode(pubFile, pubPEM); err != nil {
		fmt.Printf("Error encoding public key: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("----------------------------------------------------------------")
	fmt.Println("RSA Key Pair Generated Successfully!")
	fmt.Printf(" - %s (KEEP THIS SECRET)\n", privPath)
	fmt.Printf(" - %s (Non-sensitive, used for encryption)\n", pubPath)
	fmt.Println("----------------------------------------------------------------")
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
