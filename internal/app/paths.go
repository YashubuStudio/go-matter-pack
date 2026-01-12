package app

import (
	"os"
	"path/filepath"
)

// StateDir returns the state directory for the application based on XDG conventions.
func StateDir(appName string) string {
	baseDir := os.Getenv("XDG_STATE_HOME")
	if baseDir == "" {
		homeDir, err := os.UserHomeDir()
		if err == nil && homeDir != "" {
			baseDir = filepath.Join(homeDir, ".local", "state")
		}
	}
	if baseDir == "" {
		return filepath.Join(".", appName)
	}
	return filepath.Join(baseDir, appName)
}
