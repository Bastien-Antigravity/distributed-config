package secret

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"testing"
)

func TestEncryptionRoundTrip(t *testing.T) {
	// 1. Generate a temporary key pair for the test
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}

	pubBytes, _ := x509.MarshalPKIXPublicKey(&privKey.PublicKey)
	pubPEM := string(pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubBytes,
	}))

	privPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privKey),
	})

	// 2. Setup local private.pem for getPrivateKey() fallback
	tempKeyFile := "./private.pem"
	_ = os.WriteFile(tempKeyFile, privPEM, 0600)
	defer func() {
		_ = os.Remove(tempKeyFile)
	}()

	// 3. Test Encrypt
	original := "bastien-secret-123"
	encrypted, err := Encrypt(original, pubPEM)
	if err != nil {
		t.Fatalf("Encryption failed: %v", err)
	}

	// 4. Test Decrypt
	decrypted, err := Decrypt(encrypted)
	if err != nil {
		t.Fatalf("Decryption failed: %v", err)
	}

	if decrypted != original {
		t.Errorf("Expected %s, got %s", original, decrypted)
	}
}

func TestProcessConfigSecrets(t *testing.T) {
	// Setup keys
	privKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	pubBytes, _ := x509.MarshalPKIXPublicKey(&privKey.PublicKey)
	pubPEM := string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes}))
	privPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privKey)})

	_ = os.WriteFile("./private.pem", privPEM, 0600)
	defer func() {
		_ = os.Remove("./private.pem")
	}()

	// Create encrypted value
	secret := "my-db-password"
	encValue, _ := Encrypt(secret, pubPEM)

	yamlContent := []byte("database:\n  password: " + encValue + "\n  user: admin")

	// Process
	processed, err := ProcessConfigSecrets(yamlContent)
	if err != nil {
		t.Fatal(err)
	}

	if string(processed) == string(yamlContent) {
		t.Error("Content was not decrypted")
	}

	if !testing.Short() {
		expected := "database:\n  password: " + secret + "\n  user: admin"
		if string(processed) != expected {
			t.Errorf("Expected decrypted content, got:\n%s", string(processed))
		}
	}
}
