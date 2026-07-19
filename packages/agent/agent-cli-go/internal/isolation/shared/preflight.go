package shared

import (
	"fmt"
	"os"
	"path/filepath"
)

// ResolveHomeDir returns an absolute, existing home directory. An explicit
// directory takes precedence over the operating-system user home.
func ResolveHomeDir(explicit string) (string, error) {
	homeDir := explicit
	if homeDir == "" {
		var err error
		homeDir, err = os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolve user home directory: %w", err)
		}
	}
	return ResolveDirectory(homeDir, "home directory")
}

// ResolveDirectory returns an absolute path after verifying it is a directory.
func ResolveDirectory(path, purpose string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve %s %q: %w", purpose, path, err)
	}
	info, err := os.Stat(abs)
	if err != nil {
		return "", fmt.Errorf("inspect %s %q: %w", purpose, abs, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("%s %q is not a directory", purpose, abs)
	}
	return abs, nil
}

// ResolveExistingPath returns an absolute path after verifying it exists.
func ResolveExistingPath(path, purpose string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve %s %q: %w", purpose, path, err)
	}
	if _, err := os.Stat(abs); err != nil {
		return "", fmt.Errorf("inspect %s %q: %w", purpose, abs, err)
	}
	return abs, nil
}

// OptionalPath reports whether an optional path exists while preserving
// permission and other filesystem errors.
func OptionalPath(path, purpose string) (bool, error) {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("inspect %s %q: %w", purpose, path, err)
	}
	return true, nil
}
