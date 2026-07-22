package loader

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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

	exePath, _ := os.Executable()
	exePathAbs, _ := filepath.Abs(exePath)
	exeName := filepath.Base(exePathAbs)
	exeName = strings.TrimSuffix(exeName, filepath.Ext(exeName))

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
}
