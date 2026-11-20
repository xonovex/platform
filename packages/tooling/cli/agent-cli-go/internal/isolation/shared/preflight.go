package shared

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
)

// ResolveDirectory returns the canonical form of the absolute path after
// verifying it is a directory.
func ResolveDirectory(fsys FileSystem, path, purpose string) (string, error) {
	if err := requireAbsolute(path, purpose); err != nil {
		return "", err
	}
	abs := filepath.Clean(path)
	info, err := fsys.Stat(abs)
	if err != nil {
		return "", fmt.Errorf("inspect %s %q: %w", purpose, abs, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("%s %q is not a directory", purpose, abs)
	}
	resolved, err := fsys.EvalSymlinks(abs)
	if err != nil {
		return "", fmt.Errorf("resolve canonical %s %q: %w", purpose, abs, err)
	}
	return resolved, nil
}

// ResolveExistingPath returns the clean form of the absolute path after
// verifying it exists.
func ResolveExistingPath(fsys FileSystem, path, purpose string) (string, error) {
	if err := requireAbsolute(path, purpose); err != nil {
		return "", err
	}
	abs := filepath.Clean(path)
	if _, err := fsys.Stat(abs); err != nil {
		return "", fmt.Errorf("inspect %s %q: %w", purpose, abs, err)
	}
	return abs, nil
}

// ResolveContainedOptionalPath resolves an optional home-relative path and
// rejects roots that escape the canonical home directory through traversal or
// symlinks. The returned source is canonical and safe to bind read-only.
func ResolveContainedOptionalPath(fsys FileSystem, homeDir, relative, purpose string) (string, bool, error) {
	if filepath.IsAbs(relative) {
		return "", false, fmt.Errorf("%s %q must be relative to the home directory", purpose, relative)
	}
	clean := filepath.Clean(relative)
	if clean == "." || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", false, fmt.Errorf("%s %q escapes the home directory", purpose, relative)
	}
	home, err := ResolveDirectory(fsys, homeDir, "home directory")
	if err != nil {
		return "", false, err
	}
	candidate := filepath.Join(home, clean)
	resolved, err := fsys.EvalSymlinks(candidate)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
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

// requireAbsolute rejects a relative path: the planner resolves every path
// against the working directory of the CLI, so a relative one here would bind
// whatever directory the isolator happens to run in.
func requireAbsolute(path, purpose string) error {
	if !filepath.IsAbs(path) {
		return fmt.Errorf("%s %q must be an absolute path", purpose, path)
	}
	return nil
}
