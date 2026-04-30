package secret

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// ENC_REGEX matches ENC(base64_blob)
var ENC_REGEX = regexp.MustCompile(`ENC\(([^)]+)\)`)

func getPrivateKey() (*rsa.PrivateKey, error) {
	keyPath := os.Getenv("BASTIEN_PRIVATE_KEY_PATH")
	if keyPath == "" {
		keyPath = "/etc/bastien/private.pem"
		// Fallback for local sandbox testing
		if _, err := os.Stat("./private.pem"); err == nil {
			keyPath = "./private.pem"
		}
	}

	data, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("private key not found at %s: %w", keyPath, err)
	}

	block, _ := pem.Decode(data)
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		return nil, fmt.Errorf("failed to decode PEM block containing RSA private key")
	}

	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

// Encrypt encrypts the plaintext using the provided RSA Public Key (PEM format).
// Returns ENC(base64(ciphertext))
// -----------------------------------------------------------------------------

func Encrypt(plaintext, publicKeyPEM string) (string, error) {
	block, _ := pem.Decode([]byte(publicKeyPEM))
	if block == nil {
		return "", fmt.Errorf("failed to parse PEM block containing public key")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return "", err
	}

	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return "", fmt.Errorf("not an RSA public key")
	}

	ciphertext, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, rsaPub, []byte(plaintext), nil)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("ENC(%s)", base64.StdEncoding.EncodeToString(ciphertext)), nil
}

// Decrypt decrypts the ciphertext using the stored RSA Private Key.
// It expects the format ENC(base64_ciphertext).
// -----------------------------------------------------------------------------

func Decrypt(ciphertext string) (string, error) {
	priv, err := getPrivateKey()
	if err != nil {
		return "", err
	}

	if !strings.HasPrefix(ciphertext, "ENC(") || !strings.HasSuffix(ciphertext, ")") {
		return "", fmt.Errorf("invalid ciphertext format")
	}

	rawB64 := ciphertext[4 : len(ciphertext)-1]
	data, err := base64.StdEncoding.DecodeString(rawB64)
	if err != nil {
		return "", err
	}

	plaintext, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, priv, data, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// ProcessConfigSecrets takes the raw YAML content, finds all "ENC(...)" strings,
// decrypts them using the RSA Private Key, and returns the modified content.
// -----------------------------------------------------------------------------

func ProcessConfigSecrets(content []byte) ([]byte, error) {
	// We only try to process if we can find the private key
	_, err := getPrivateKey()
	if err != nil {
		// If key is missing, we assume no secrets need to be decrypted
		return content, nil
	}

	result := ENC_REGEX.ReplaceAllFunc(content, func(match []byte) []byte {
		decrypted, err := Decrypt(string(match))
		if err != nil {
			fmt.Printf("Warning: Failed to decrypt secret: %v\n", err)
			return match
		}
		return []byte(decrypted)
	})

	return result, nil
}
