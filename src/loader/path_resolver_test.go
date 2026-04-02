package loader

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveConfigPath(t *testing.T) {
	// 1. Create a dummy test environment in a temp directory
	tempDir_raw, _ := os.MkdirTemp("", "distconf-test-*")
	tempDir, _ := filepath.EvalSymlinks(tempDir_raw)
	defer os.RemoveAll(tempDir)

	oldCwd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(oldCwd)

	t.Run("Priority-CWD-Config-Specific", func(t *testing.T) {
		configDir := filepath.Join(tempDir, "config")
		os.MkdirAll(configDir, 0755)

		specificPath := filepath.Join(configDir, "myapp.yaml")
		os.WriteFile(specificPath, []byte("name: specific"), 0644)

		path := ResolveConfigPath("myapp")
		
		path, _ = filepath.Abs(path)
		path, _ = filepath.EvalSymlinks(path)
		expected, _ := filepath.Abs(specificPath)
		expected, _ = filepath.EvalSymlinks(expected)
		
		if path != expected {
			t.Errorf("Expected path %s, got %s", expected, path)
		}
	})

	t.Run("Priority-CWD-Config-Default", func(t *testing.T) {
		tempDir2_raw, _ := os.MkdirTemp("", "distconf-test-default-*")
		tempDir2, _ := filepath.EvalSymlinks(tempDir2_raw)
		defer os.RemoveAll(tempDir2)
		
		os.Chdir(tempDir2)
		// No defer os.Chdir(tempDir) here as it's shared by other tests? 
		// Actually best to re-set CWD at end of subtest.
		defer os.Chdir(tempDir)

		configDir := filepath.Join(tempDir2, "config")
		os.MkdirAll(configDir, 0755)

		defaultPath := filepath.Join(configDir, "default.yaml")
		os.WriteFile(defaultPath, []byte("name: default"), 0644)

		path := ResolveConfigPath("nonexistent")
		
		path, _ = filepath.Abs(path)
		path, _ = filepath.EvalSymlinks(path)
		expected, _ := filepath.Abs(defaultPath)
		expected, _ = filepath.EvalSymlinks(expected)

		if path != expected {
			t.Errorf("Expected default path %s, got %s", expected, path)
		}
	})

	t.Run("Priority-CWD-Local", func(t *testing.T) {
		tempDir3_raw, _ := os.MkdirTemp("", "distconf-test-local-*")
		tempDir3, _ := filepath.EvalSymlinks(tempDir3_raw)
		defer os.RemoveAll(tempDir3)
		
		os.Chdir(tempDir3)
		defer os.Chdir(tempDir)

		localPath := filepath.Join(tempDir3, "myapp.yaml")
		os.WriteFile(localPath, []byte("name: local"), 0644)

		path := ResolveConfigPath("myapp")
		
		path, _ = filepath.Abs(path)
		path, _ = filepath.EvalSymlinks(path)
		expected, _ := filepath.Abs(localPath)
		expected, _ = filepath.EvalSymlinks(expected)

		if path != expected {
			t.Errorf("Expected local path %s, got %s", expected, path)
		}
	})
}
