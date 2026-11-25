package secret

// =============================================================================
// ESSENTIAL PROCESS:
// Unit test suite verifying RSA keypair generation, token encryption, and
// round-trip ciphertext decryption using temporary test keys and env overrides.
//
// DATA FLOW:
// 1. Input: Randomly generated test RSA keys and plaintext test secrets.
// 2. Logic: Encrypts with Encrypt; decrypts with Decrypt and ProcessConfigSecrets.
// 3. Output: Test assertion results verifying plaintext integrity.
//
// KEY PARAMETERS:
// - t: Testing harness handle.
// =============================================================================

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"
)

// -----------------------------------------------------------------------------

func isolateSystemPrivateKey(t *testing.T) {
	t.Helper()
	prev := defaultSystemPrivateKeyPath
	defaultSystemPrivateKeyPath = filepath.Join(t.TempDir(), "nonexistent-system-private.pem")
	t.Setenv("BASTIEN_PRIVATE_KEY", "")
	t.Setenv("BASTIEN_PRIVATE_KEY_PATH", "")
	ResetKeyCache()
	t.Cleanup(func() {
		defaultSystemPrivateKeyPath = prev
		ResetKeyCache()
	})
}

func TestEncryptionRoundTrip(t *testing.T) {
	isolateSystemPrivateKey(t)
	// 1. Generate a temporary key pair for the test
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}

	pubBytes, err := x509.MarshalPKIXPublicKey(&privKey.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
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
	isolateSystemPrivateKey(t)
	// Setup keys
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	pubBytes, err := x509.MarshalPKIXPublicKey(&privKey.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	pubPEM := string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes}))
	privPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privKey)})

	_ = os.WriteFile("./private.pem", privPEM, 0600)
	defer func() {
		_ = os.Remove("./private.pem")
	}()

	// Create encrypted value
	secret := "my-db-password"
	encValue, err := Encrypt(secret, pubPEM)
	if err != nil {
		t.Fatal(err)
	}

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

// -----------------------------------------------------------------------------

func TestProcessConfigSecrets_FailFastAndCaching(t *testing.T) {
	isolateSystemPrivateKey(t)
	_ = os.Remove("./private.pem")

	// 1. Plain config without ENC(...) does not require a private key
	plain := []byte("common:\n  name: plain-service\n")
	out, err := ProcessConfigSecrets(plain)
	if err != nil {
		t.Fatalf("Expected plain config without ENC to succeed without key, got: %v", err)
	}
	if string(out) != string(plain) {
		t.Fatalf("Expected unchanged plain content, got: %s", string(out))
	}

	// 2. Config with ENC(...) MUST fail fast when private key is missing
	withEnc := []byte("secret: ENC(dGhpcyBpcyBub3QgYSB2YWxpZCBjaXBoZXJ0ZXh0)")
	if _, err := ProcessConfigSecrets(withEnc); err == nil {
		t.Fatal("Expected ProcessConfigSecrets to fail fast when ENC token is present but private key is missing")
	}

	// 3. Provide private key via BASTIEN_PRIVATE_KEY env var and verify fail-fast on bad ciphertext
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	privPEM := string(pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privKey),
	}))
	t.Setenv("BASTIEN_PRIVATE_KEY", privPEM)
	defer ResetKeyCache()

	if _, err := ProcessConfigSecrets(withEnc); err == nil {
		t.Fatal("Expected ProcessConfigSecrets to fail fast on invalid ENC token ciphertext")
	}

	// 4. Valid ENC token decrypts and uses cached key on repeated calls
	pubBytes, err := x509.MarshalPKIXPublicKey(&privKey.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	pubPEM := string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes}))
	encToken, err := Encrypt("cached-secret-val", pubPEM)
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 3; i++ {
		dec, err := ProcessConfigSecrets([]byte("token: " + encToken))
		if err != nil {
			t.Fatalf("Iteration %d failed: %v", i, err)
		}
		if string(dec) != "token: cached-secret-val" {
			t.Fatalf("Unexpected decrypted output: %s", string(dec))
		}
	}
}

// -----------------------------------------------------------------------------

func TestPrivateKeyLookupPrecedence(t *testing.T) {
	isolateSystemPrivateKey(t)

	// Generate two distinct key pairs: system key and local CWD key
	sysKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	sysPrivPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(sysKey)})
	sysPubBytes, _ := x509.MarshalPKIXPublicKey(&sysKey.PublicKey)
	sysPubPEM := string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: sysPubBytes}))

	localKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	localPrivPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(localKey)})
	localPubBytes, _ := x509.MarshalPKIXPublicKey(&localKey.PublicKey)
	localPubPEM := string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: localPubBytes}))

	workDir := t.TempDir()
	oldCwd, _ := os.Getwd()
	if err := os.Chdir(workDir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(oldCwd) }()

	if err := os.WriteFile(filepath.Join(workDir, "private.pem"), localPrivPEM, 0600); err != nil {
		t.Fatal(err)
	}

	localCipher, err := Encrypt("local-secret", localPubPEM)
	if err != nil {
		t.Fatal(err)
	}
	sysCipher, err := Encrypt("system-secret", sysPubPEM)
	if err != nil {
		t.Fatal(err)
	}

	// 1. In production, ./private.pem MUST NOT be used when system key is missing
	t.Setenv("APP_ENV", "production")
	ResetKeyCache()
	if _, err := Decrypt(localCipher); err == nil {
		t.Fatal("Expected APP_ENV=production to ignore ./private.pem fallback")
	}

	// 2. Outside production, ./private.pem works when system key does not exist
	t.Setenv("APP_ENV", "")
	ResetKeyCache()
	gotLocal, err := Decrypt(localCipher)
	if err != nil || gotLocal != "local-secret" {
		t.Fatalf("Expected ./private.pem fallback outside production, got %q (err=%v)", gotLocal, err)
	}

	// 3. When system private key exists, it MUST take precedence over ./private.pem
	sysPath := filepath.Join(t.TempDir(), "system-private.pem")
	if err := os.WriteFile(sysPath, sysPrivPEM, 0600); err != nil {
		t.Fatal(err)
	}
	defaultSystemPrivateKeyPath = sysPath
	ResetKeyCache()
	gotSys, err := Decrypt(sysCipher)
	if err != nil || gotSys != "system-secret" {
		t.Fatalf("Expected system key precedence over ./private.pem, got %q (err=%v)", gotSys, err)
	}
	if _, err := Decrypt(localCipher); err == nil {
		t.Fatal("Expected localCipher not to decrypt when system key takes precedence")
	}
}
