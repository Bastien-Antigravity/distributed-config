package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

func main() {
	// Generate Private Key
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		fmt.Printf("Error generating key: %v\n", err)
		os.Exit(1)
	}

	// Save Private Key
	privFile, err := os.Create("private.pem")
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
		fmt.Printf("Error encoding private.pem: %v\n", err)
		os.Exit(1)
	}

	// Save Public Key
	pubFile, err := os.Create("public.pem")
	if err != nil {
		fmt.Printf("Error creating public.pem: %v\n", err)
		os.Exit(1)
	}
	defer pubFile.Close()

	pubBytes, err := x509.MarshalPKIXPublicKey(&privKey.PublicKey)
	if err != nil {
		fmt.Printf("Error marshaling public key: %v\n", err)
		os.Exit(1)
	}
	pubPEM := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubBytes,
	}
	if err := pem.Encode(pubFile, pubPEM); err != nil {
		fmt.Printf("Error encoding public.pem: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("----------------------------------------------------------------")
	fmt.Println("RSA Key Pair Generated Successfully!")
	fmt.Println(" - private.pem (KEEP THIS SECRET, mount it in Docker)")
	fmt.Println(" - public.pem  (Use this to encrypt your tokens)")
	fmt.Println("----------------------------------------------------------------")
}
