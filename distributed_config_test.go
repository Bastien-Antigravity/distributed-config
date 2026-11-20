package distributed_config

// =============================================================================
// ESSENTIAL PROCESS:
// Top-level entrypoint test suite validating facade constructors, node
// processing, path resolution, and secret decryption helpers.
//
// DATA FLOW:
// 1. Input: Profile names, YAML nodes, and target file identifiers.
// 2. Logic: Invokes top-level package exported functions and verifies output.
// 3. Output: Test assertion results confirming top-level library behavior.
//
// KEY PARAMETERS:
// - t: Standard Go testing harness pointer.
// =============================================================================

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

// -----------------------------------------------------------------------------

func TestTopLevel_New(t *testing.T) {
	profiles := []string{"standalone", "test", "staging", "production"}
	for _, p := range profiles {
		cfg := New(p)
		if cfg == nil {
			t.Fatalf("Expected non-nil Config for profile %s", p)
		}
	}
}

// -----------------------------------------------------------------------------

func TestTopLevel_ResolveConfigPath(t *testing.T) {
	// Should resolve standalone.yaml in current repository root
	path := ResolveConfigPath("standalone")
	if path == "" {
		t.Fatalf("Expected ResolveConfigPath('standalone') to return path, got empty")
	}
	if filepath.Base(path) != "standalone.yaml" {
		t.Errorf("Expected base name 'standalone.yaml', got '%s'", filepath.Base(path))
	}
}

// -----------------------------------------------------------------------------

func TestTopLevel_ProcessNode(t *testing.T) {
	t.Setenv("TOP_LEVEL_TEST_VAR", "HelloUniverse")
	node := &yaml.Node{
		Kind:  yaml.ScalarNode,
		Value: "${TOP_LEVEL_TEST_VAR:default_val}",
	}
	ProcessNode(node)
	if node.Value != "HelloUniverse" {
		t.Errorf("Expected expanded value 'HelloUniverse', got '%s'", node.Value)
	}
}

// -----------------------------------------------------------------------------

func TestTopLevel_Secrets(t *testing.T) {
	// Without private key set, Decrypt should return an explicit error
	_, err := Decrypt("ENC(invalid_token)")
	if err == nil {
		t.Errorf("Expected error when decrypting invalid token without private key, got nil")
	}

	// ProcessConfigSecrets on plain text with no ENC should return unchanged
	plain := []byte("plain text content without enc")
	processed, err := ProcessConfigSecrets(plain)
	if err != nil {
		t.Fatalf("Expected ProcessConfigSecrets to succeed on plain text, got: %v", err)
	}
	if string(processed) != string(plain) {
		t.Errorf("Expected '%s', got '%s'", string(plain), string(processed))
	}
}

// -----------------------------------------------------------------------------

func TestTopLevel_EnvironmentOverride(t *testing.T) {
	tempDir := t.TempDir()
	dummyConfig := filepath.Join(tempDir, "custom.yaml")
	if err := os.WriteFile(dummyConfig, []byte("common:\n  name: custom-app\n"), 0644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("CONFIG_PATH", dummyConfig)
	resolved := ResolveConfigPath("standalone")
	if resolved != dummyConfig {
		t.Errorf("Expected CONFIG_PATH override '%s', got '%s'", dummyConfig, resolved)
	}
}
