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
// 1. [targetName].yaml (in CWD config/, EXE config/, CWD, EXE)
// 2. [exeName].yaml (in CWD config/, EXE config/, CWD, EXE)
// -----------------------------------------------------------------------------

func ResolveConfigPath(targetName string) string {
	// 1. Determine Executable Context
	exePath, err := os.Executable()
	exeName := ""
	exeDir := ""
	if err == nil {
		exePathAbs, _ := filepath.Abs(exePath)
		exeName = filepath.Base(exePathAbs)
		exeName = strings.TrimSuffix(exeName, filepath.Ext(exeName))
		exeDir = filepath.Dir(exePathAbs)
	}

	cwd, _ := os.Getwd()

	// 2. Build Search Candidates (STRICT ORDER)
	var candidates []string

	// Steps 1 & 2: config/[targetName].yaml (CWD then EXEDIR)
	if targetName != "" {
		if cwd != "" {
			candidates = append(candidates, filepath.Join(cwd, "config", targetName+".yaml"))
		}
		if exeDir != "" {
			candidates = append(candidates, filepath.Join(exeDir, "config", targetName+".yaml"))
		}
	}

	// Steps 3 & 4: config/[exeName].yaml (CWD then EXEDIR)
	if exeName != "" {
		if cwd != "" {
			candidates = append(candidates, filepath.Join(cwd, "config", exeName+".yaml"))
		}
		if exeDir != "" {
			candidates = append(candidates, filepath.Join(exeDir, "config", exeName+".yaml"))
		}
	}

	// Steps 5 & 6: [exeName].yaml (CWD then EXEDIR)
	if exeName != "" {
		if cwd != "" {
			candidates = append(candidates, filepath.Join(cwd, exeName+".yaml"))
		}
		if exeDir != "" {
			candidates = append(candidates, filepath.Join(exeDir, exeName+".yaml"))
		}
	}

	// 3. Search Loop
	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	// 4. Default Path: if nothing found, return [exeName].yaml or [targetName].yaml as a guess
	finalDefault := exeName
	if finalDefault == "" {
		finalDefault = targetName
	}
	return finalDefault + ".yaml"
}
