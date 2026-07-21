package shared

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
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
	resolved, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return "", fmt.Errorf("resolve canonical %s %q: %w", purpose, abs, err)
	}
	return resolved, nil
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

// ResolveContainedOptionalPath resolves an optional home-relative path and
// rejects roots that escape the canonical home directory through traversal or
// symlinks. The returned source is canonical and safe to bind read-only.
func ResolveContainedOptionalPath(homeDir, relative, purpose string) (string, bool, error) {
	if filepath.IsAbs(relative) {
		return "", false, fmt.Errorf("%s %q must be relative to the home directory", purpose, relative)
	}
	clean := filepath.Clean(relative)
	if clean == "." || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", false, fmt.Errorf("%s %q escapes the home directory", purpose, relative)
	}
	home, err := ResolveDirectory(homeDir, "home directory")
	if err != nil {
		return "", false, err
	}
	candidate := filepath.Join(home, clean)
	resolved, err := filepath.EvalSymlinks(candidate)
	if err != nil {
		if os.IsNotExist(err) {
			return "", false, nil
		}
		return "", false, fmt.Errorf("resolve %s %q: %w", purpose, candidate, err)
	}
	relativeToHome, err := filepath.Rel(home, resolved)
	if err != nil {
		return "", false, fmt.Errorf("verify %s containment: %w", purpose, err)
	}
	if relativeToHome == ".." || strings.HasPrefix(relativeToHome, ".."+string(filepath.Separator)) || filepath.IsAbs(relativeToHome) {
		return "", false, fmt.Errorf("%s %q resolves outside home directory %q", purpose, relative, home)
	}
	return resolved, true, nil
}
