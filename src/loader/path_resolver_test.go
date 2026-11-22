package loader

// =============================================================================
// ESSENTIAL PROCESS:
// Unit test suite verifying configuration path resolution priority across CWD,
// config/ subdirectory, executable directory, and fallback binary names.
//
// DATA FLOW:
// 1. Input: Mock directory structures created inside os.MkdirTemp.
// 2. Logic: Creates sample YAML files at various priority levels and invokes ResolveConfigPath.
// 3. Output: Test assertion results verifying path resolution precedence.
//
// KEY PARAMETERS:
// - t: Standard testing harness handle.
// =============================================================================

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// -----------------------------------------------------------------------------

func TestResolveConfigPath(t *testing.T) {
	// Create a dummy test environment in a temp directory
	tempDir_raw, _ := os.MkdirTemp("", "distconf-test-*")
	tempDir, _ := filepath.EvalSymlinks(tempDir_raw)
	defer func() {
		_ = os.RemoveAll(tempDir)
	}()

	oldCwd, _ := os.Getwd()
	_ = os.Chdir(tempDir)
	defer func() { _ = os.Chdir(oldCwd) }()

	t.Run("Priority-1-CWD-Target", func(t *testing.T) {
		tempRun, _ := os.MkdirTemp(tempDir, "run-*")
		_ = os.Chdir(tempRun)
		defer func() { _ = os.Chdir(tempDir) }()

		// Create target file directly in CWD
		targetPath := filepath.Join(tempRun, "staging.yaml")
		_ = os.WriteFile(targetPath, []byte("name: staging"), 0644)

		path := ResolveConfigPath("staging")

		path, _ = filepath.Abs(path)
		path, _ = filepath.EvalSymlinks(path)
		expected, _ := filepath.Abs(targetPath)
		expected, _ = filepath.EvalSymlinks(expected)

		if path != expected {
			t.Errorf("Expected path %s, got %s", expected, path)
		}
	})

	t.Run("Priority-2-CompleteFallback", func(t *testing.T) {
		tempRun, _ := os.MkdirTemp(tempDir, "run-*")
		_ = os.Chdir(tempRun)
		defer func() { _ = os.Chdir(tempDir) }()

		// Nothing created, should default to [targetName].yaml in base dir (CWD / exeDir)
		path := ResolveConfigPath("production")
		if !strings.HasSuffix(path, "production.yaml") {
			t.Errorf("Expected fallback to end with production.yaml, got %s", path)
		}
	})

	t.Run("Priority-0-SharedConfigPath-Directory", func(t *testing.T) {
		sharedDir, _ := os.MkdirTemp(tempDir, "shared-cfg-*")
		stagingFile := filepath.Join(sharedDir, "staging.yaml")
		standaloneFile := filepath.Join(sharedDir, "standalone.yaml")
		_ = os.WriteFile(stagingFile, []byte("common:\n  name: staging-shared\n"), 0644)
		_ = os.WriteFile(standaloneFile, []byte("common:\n  name: standalone-shared\n"), 0644)

		t.Setenv("SHARED_CONFIG_PATH", sharedDir)

		gotStaging := ResolveConfigPath("staging")
		if gotStaging != stagingFile {
			t.Errorf("Expected SHARED_CONFIG_PATH directory to resolve %s, got %s", stagingFile, gotStaging)
		}

		gotFallback := ResolveConfigPath("nonexistent_profile")
		if gotFallback != standaloneFile {
			t.Errorf("Expected SHARED_CONFIG_PATH directory fallback to resolve %s, got %s", standaloneFile, gotFallback)
		}
	})
}
