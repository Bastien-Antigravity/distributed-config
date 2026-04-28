package main

import (
	"fmt"
	"os"

	"github.com/Bastien-Antigravity/distributed-config/src/secret"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run encrypt_token.go <public_key_path> <plaintext_token>")
		os.Exit(1)
	}

	pubKeyPath := os.Args[1]
	token := os.Args[2]

	pubKeyBytes, err := os.ReadFile(pubKeyPath)
	if err != nil {
		fmt.Printf("Error reading public key: %v\n", err)
		os.Exit(1)
	}

	encrypted, err := secret.Encrypt(token, string(pubKeyBytes))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("----------------------------------------------------------------")
	fmt.Println("Encrypted Token (Paste this into your config.yaml):")
	fmt.Println(encrypted)
	fmt.Println("----------------------------------------------------------------")
}
