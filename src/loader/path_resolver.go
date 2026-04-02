package loader

import (
	"os"
	"path/filepath"
	"strings"
)

// ResolveConfigPath handles the search logic for configuration files.
// It prioritizes the explicitly provided 'fallbackName' (usually profile or target name)
// but also falls back to the executable name if neither is found.
//
// SEARCH ORDER:
// 1. [targetName].yaml (in CWD config/, EXE config/, then CWD, then EXE)
// 2. default.yaml (in CWD config/ then EXE config/)
// -----------------------------------------------------------------------------

func ResolveConfigPath(targetName string) string {
	// 1. Determine Target Name
	exePath, err := os.Executable()
	exeName := ""
	exeDir := ""
	if err == nil {
		exeName = filepath.Base(exePath)
		exeName = strings.TrimSuffix(exeName, filepath.Ext(exeName))
		exeDir = filepath.Dir(exePath)
	}

	primaryName := targetName
	if primaryName == "" {
		primaryName = exeName
	}

	// 2. Build Search Candidates (Priority Ordered)
	var candidates []string
	cwd, _ := os.Getwd()

	// Priority 1: Specifically named file in config/
	if cwd != "" {
		candidates = append(candidates, filepath.Join(cwd, "config", primaryName+".yaml"))
	}
	if exeDir != "" {
		candidates = append(candidates, filepath.Join(exeDir, "config", primaryName+".yaml"))
	}

	// Priority 2: default.yaml in config/
	if cwd != "" {
		candidates = append(candidates, filepath.Join(cwd, "config", "default.yaml"))
	}
	if exeDir != "" {
		candidates = append(candidates, filepath.Join(exeDir, "config", "default.yaml"))
	}

	// Priority 3: Specifically named file in root
	if cwd != "" {
		candidates = append(candidates, filepath.Join(cwd, primaryName+".yaml"))
	}
	if exeDir != "" {
		candidates = append(candidates, filepath.Join(exeDir, primaryName+".yaml"))
	}

	// Priority 4: Magic Exe fallbacks (if different from targetName)
	if exeName != "" && exeName != primaryName {
		candidates = append(candidates,
			filepath.Join(cwd, "config", exeName+".yaml"),
			filepath.Join(exeDir, "config", exeName+".yaml"),
			filepath.Join(cwd, exeName+".yaml"),
			filepath.Join(exeDir, exeName+".yaml"),
		)
	}

	// 3. Search Loop
	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	// 4. Default Path: if nothing found, return [primaryName].yaml in current dir
	return primaryName + ".yaml"
}
