package utils

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/magefile/mage/sh"
)

// FileExists checks if a file exists at the given path.
// Returns true if the file exists, false otherwise.
func FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return err == nil
}

// FileExistsInDirectory checks if a file with the given name exists in the specified directory.
// Returns true if the file exists, false otherwise.
func FileExistsInDirectory(directory, filename string) bool {
	fullPath := filepath.Join(directory, filename)
	return FileExists(fullPath)
}

// GetRepoRoot returns the absolute path to the repository root directory.
// It uses git to find the top-level directory of the repository.
// If git is not available or not in a git repository, it falls back to the current working directory.
func GetRepoRoot() (string, error) {
	// Try to get repo root from git
	repoRoot, err := sh.Output("git", "rev-parse", "--show-toplevel")
	if err == nil {
		return strings.TrimSpace(repoRoot), nil
	}

	// Fallback to current working directory if not in a git repo
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return cwd, nil
}
