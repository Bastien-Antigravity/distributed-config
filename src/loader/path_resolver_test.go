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
	defer os.RemoveAll(tempDir)

	oldCwd, _ := os.Getwd()
	_ = os.Chdir(tempDir)
	defer func() { _ = os.Chdir(oldCwd) }()

	exePath, _ := os.Executable()
	exePathAbs, _ := filepath.Abs(exePath)
	exeName := filepath.Base(exePathAbs)
	exeName = strings.TrimSuffix(exeName, filepath.Ext(exeName))

	t.Run("Priority-1-ConfigSubfolder-Target", func(t *testing.T) {
		tempRun, _ := os.MkdirTemp(tempDir, "run-*")
		_ = os.Chdir(tempRun)
		defer func() { _ = os.Chdir(tempDir) }()

		configDir := filepath.Join(tempRun, "config")
		_ = os.MkdirAll(configDir, 0755)

		// Create target file
		targetPath := filepath.Join(configDir, "staging.yaml")
		_ = os.WriteFile(targetPath, []byte("name: staging"), 0644)

		// Create inferior priority file to ensure it's bypassed
		exePathLocal := filepath.Join(configDir, exeName+".yaml")
		_ = os.WriteFile(exePathLocal, []byte("name: wrong"), 0644)

		path := ResolveConfigPath("staging")
		
		path, _ = filepath.Abs(path)
		path, _ = filepath.EvalSymlinks(path)
		expected, _ := filepath.Abs(targetPath)
		expected, _ = filepath.EvalSymlinks(expected)
		
		if path != expected {
			t.Errorf("Expected path %s, got %s", expected, path)
		}
	})

	t.Run("Priority-3-ConfigSubfolder-ExeFallback", func(t *testing.T) {
		tempRun, _ := os.MkdirTemp(tempDir, "run-*")
		_ = os.Chdir(tempRun)
		defer func() { _ = os.Chdir(tempDir) }()

		configDir := filepath.Join(tempRun, "config")
		_ = os.MkdirAll(configDir, 0755)

		// Create specific file based on exeName
		exePathLocal := filepath.Join(configDir, exeName+".yaml")
		_ = os.WriteFile(exePathLocal, []byte("name: exefallback"), 0644)
		
		// Target 'production' is missing, fallback should catch exeName in config/
		path := ResolveConfigPath("production")
		
		path, _ = filepath.Abs(path)
		path, _ = filepath.EvalSymlinks(path)
		expected, _ := filepath.Abs(exePathLocal)
		expected, _ = filepath.EvalSymlinks(expected)

		if path != expected {
			t.Errorf("Expected config/[exe].yaml path %s, got %s", expected, path)
		}
	})

	t.Run("Priority-5-CWD-ExeFallback", func(t *testing.T) {
		tempRun, _ := os.MkdirTemp(tempDir, "run-*")
		_ = os.Chdir(tempRun)
		defer func() { _ = os.Chdir(tempDir) }()

		// Create specific file based on exeName in CWD
		exePathLocal := filepath.Join(tempRun, exeName+".yaml")
		_ = os.WriteFile(exePathLocal, []byte("name: exefallback_cwd"), 0644)
		
		// Target 'production' is missing, fallback should catch exeName in CWD
		path := ResolveConfigPath("production")
		
		path, _ = filepath.Abs(path)
		path, _ = filepath.EvalSymlinks(path)
		expected, _ := filepath.Abs(exePathLocal)
		expected, _ = filepath.EvalSymlinks(expected)

		if path != expected {
			t.Errorf("Expected [exe].yaml path %s, got %s", expected, path)
		}
	})
	
	t.Run("Priority-7-CompleteFallback", func(t *testing.T) {
		tempRun, _ := os.MkdirTemp(tempDir, "run-*")
		_ = os.Chdir(tempRun)
		defer func() { _ = os.Chdir(tempDir) }()

		// Nothing created
		path := ResolveConfigPath("production")
		// Should just default to [exeName].yaml strings
		if path != exeName+".yaml" {
			t.Errorf("Expected fallback %s.yaml, got %s", exeName, path)
		}
	})
}
