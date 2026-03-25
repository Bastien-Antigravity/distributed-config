package secret

import "fmt"

// Encrypt encrypts the plaintext using the provided key.
// Format: ENC(base64_ciphertext)
// -----------------------------------------------------------------------------

func Encrypt(plaintext, key string) (string, error) {
	panic("not implemented")
}

// Decrypt decrypts the ciphertext using the provided key.
// It expects the format ENC(base64_ciphertext).
// -----------------------------------------------------------------------------

func Decrypt(ciphertext, key string) (string, error) {
	panic("not implemented")
}

// ProcessConfigSecrets takes the raw YAML content, finds all "ENC(...)" strings,
// decrypts them using the key from Environment, and returns the modified content.
// -----------------------------------------------------------------------------

func ProcessConfigSecrets(content []byte) ([]byte, error) {
	// panic("not implemented") // Uncomment when ready to enforce
	fmt.Println("Warning: Secrets processing is not implemented.")
	return content, nil
}
