package secret

// =============================================================================
// ESSENTIAL PROCESS:
// Cryptographic subsystem handling RSA-OAEP secret encryption and decryption,
// supporting inline ENC(...) token extraction across configurations.
//
// DATA FLOW:
// 1. Input: Plaintext secrets, PEM keys, or ENC(...) ciphertext strings/byte buffers.
// 2. Logic: Uses RSA-2048 with SHA-256 OAEP padding; resolves private key from env or disk.
// 3. Output: Decrypted plaintext strings or parsed configuration payloads.
//
// KEY PARAMETERS:
// - BASTIEN_PRIVATE_KEY: Raw PEM string environment override.
// - BASTIEN_PRIVATE_KEY_PATH: Filesystem path to private key PEM (default: /etc/bastien/private.pem).
// =============================================================================

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
	"sync"
	"time"
)

// -----------------------------------------------------------------------------

// ENC_REGEX matches ENC(base64_blob)
var ENC_REGEX = regexp.MustCompile(`ENC\(([^)]+)\)`)

var defaultSystemPrivateKeyPath = "/etc/bastien/private.pem"

var (
	keyCacheMu    sync.RWMutex
	cachedKey     *rsa.PrivateKey
	cachedEnvHash [32]byte
	cachedPath    string
	cachedModTime time.Time
	cachedSize    int64
)

// ResetKeyCache clears the in-memory memoized RSA private key.
func ResetKeyCache() {
	keyCacheMu.Lock()
	defer keyCacheMu.Unlock()
	cachedKey = nil
	cachedEnvHash = [32]byte{}
	cachedPath = ""
	cachedModTime = time.Time{}
	cachedSize = 0
}

func parsePrivateKeyPEM(data []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(data)
	if block == nil || (block.Type != "RSA PRIVATE KEY" && block.Type != "PRIVATE KEY") {
		return nil, fmt.Errorf("failed to decode PEM block containing RSA private key")
	}

	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return key, nil
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}
	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA private key")
	}
	return rsaKey, nil
}

func getPrivateKey() (*rsa.PrivateKey, error) {
	// 1. Check for direct key content in environment variable (Highest Priority)
	keyContent := os.Getenv("BASTIEN_PRIVATE_KEY")
	if keyContent != "" {
		envHash := sha256.Sum256([]byte(keyContent))
		keyCacheMu.RLock()
		if cachedKey != nil && cachedPath == "" && cachedEnvHash == envHash {
			k := cachedKey
			keyCacheMu.RUnlock()
			return k, nil
		}
		keyCacheMu.RUnlock()

		rsaKey, err := parsePrivateKeyPEM([]byte(keyContent))
		if err != nil {
			return nil, err
		}

		keyCacheMu.Lock()
		cachedKey = rsaKey
		cachedEnvHash = envHash
		cachedPath = ""
		cachedModTime = time.Time{}
		cachedSize = 0
		keyCacheMu.Unlock()
		return rsaKey, nil
	}

	// 2. Fallback to file path
	keyPath := os.Getenv("BASTIEN_PRIVATE_KEY_PATH")
	if keyPath == "" {
		keyPath = defaultSystemPrivateKeyPath
		if _, err := os.Stat(keyPath); os.IsNotExist(err) && os.Getenv("APP_ENV") != "production" {
			// Local sandbox fallback only when system key does not exist outside production
			if _, localErr := os.Stat("./private.pem"); localErr == nil {
				keyPath = "./private.pem"
			}
		}
	}

	info, err := os.Stat(keyPath)
	if err != nil {
		return nil, fmt.Errorf("private key not found in BASTIEN_PRIVATE_KEY or at %s: %w", keyPath, err)
	}

	keyCacheMu.RLock()
	if cachedKey != nil && cachedPath == keyPath && cachedModTime.Equal(info.ModTime()) && cachedSize == info.Size() {
		k := cachedKey
		keyCacheMu.RUnlock()
		return k, nil
	}
	keyCacheMu.RUnlock()

	data, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("private key not found in BASTIEN_PRIVATE_KEY or at %s: %w", keyPath, err)
	}

	rsaKey, err := parsePrivateKeyPEM(data)
	if err != nil {
		return nil, err
	}

	keyCacheMu.Lock()
	cachedKey = rsaKey
	cachedEnvHash = [32]byte{}
	cachedPath = keyPath
	cachedModTime = info.ModTime()
	cachedSize = info.Size()
	keyCacheMu.Unlock()

	return rsaKey, nil
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
	return decryptWithKey(ciphertext, priv)
}

func decryptWithKey(ciphertext string, priv *rsa.PrivateKey) (string, error) {
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
// Fails fast if any ENC(...) token is present and the key is missing or decryption fails.
// -----------------------------------------------------------------------------

func ProcessConfigSecrets(content []byte) ([]byte, error) {
	if !ENC_REGEX.Match(content) {
		return content, nil
	}

	priv, err := getPrivateKey()
	if err != nil {
		return nil, fmt.Errorf("secret decryption failed: %w", err)
	}

	var decryptErr error
	result := ENC_REGEX.ReplaceAllFunc(content, func(match []byte) []byte {
		if decryptErr != nil {
			return match
		}
		decrypted, err := decryptWithKey(string(match), priv)
		if err != nil {
			decryptErr = fmt.Errorf("failed to decrypt secret token: %w", err)
			return match
		}
		return []byte(decrypted)
	})
	if decryptErr != nil {
		return nil, decryptErr
	}

	return result, nil
}
