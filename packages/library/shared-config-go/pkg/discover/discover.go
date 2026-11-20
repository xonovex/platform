// Package discover finds the files of the layers: the managed file of the
// host, the file of the user, and the nearest agentic.json of the project.
package discover

import (
	"errors"
	"fmt"
	iofs "io/fs"
	"path/filepath"

	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/spec"
)

// FileName is the name of the file in every layer.
const FileName = "agentic.json"

// Info describes a path.
type Info struct {
	Mode iofs.FileMode
}

// Files is the port the layers are read through. Stat wraps io/fs.ErrNotExist
// for a missing path.
type Files interface {
	ReadFile(path string) ([]byte, error)
	Stat(path string) (Info, error)
}

// Places is the port that names the directories of the host.
type Places interface {
	// ManagedDir is the directory of the managed file, empty on a host that has
	// none.
	ManagedDir() string
	// UserDir is the directory of the user file.
	UserDir() (string, error)
	// StateDir is the directory of the state of the user, such as approvals.
	StateDir() (string, error)
	// Getwd is the working directory.
	Getwd() (string, error)
}

// Candidate is a file that may hold one layer.
type Candidate struct {
	Layer spec.Layer
	Path  string
	// Required is true for a file the user named, which must exist.
	Required bool
}

// repositoryMarkers mark the root of a git or a jj repository.
var repositoryMarkers = [...]string{".git", ".jj"}

// Find returns the candidates of the managed, user and project layers. start is
// the directory the project search begins in, relative to the working
// directory. explicit names the project file instead of the search, relative to
// start.
func Find(files Files, places Places, start, explicit string) ([]Candidate, error) {
	if !filepath.IsAbs(start) {
		wd, err := places.Getwd()
		if err != nil {
			return nil, fmt.Errorf("find the working directory: %w", err)
		}
		start = filepath.Join(wd, start)
	}
	var candidates []Candidate
	if dir := places.ManagedDir(); dir != "" {
		candidates = append(candidates, Candidate{Layer: spec.Managed, Path: filepath.Join(dir, FileName)})
	}
	userDir, err := places.UserDir()
	if err != nil {
		return nil, fmt.Errorf("find the user configuration directory: %w", err)
	}
	candidates = append(candidates, Candidate{Layer: spec.User, Path: filepath.Join(userDir, FileName)})
	if explicit != "" {
		if !filepath.IsAbs(explicit) {
			explicit = filepath.Join(start, explicit)
		}
		return append(candidates, Candidate{Layer: spec.Project, Path: filepath.Clean(explicit), Required: true}), nil
	}
	project, found, err := Project(files, start)
	if err != nil || !found {
		return candidates, err
	}
	return append(candidates, Candidate{Layer: spec.Project, Path: project}), nil
}

// Project returns the nearest agentic.json from start up to the root of the
// repository that holds start. Outside a repository only start itself is
// searched, so a file that another user left in a parent directory never
// becomes a layer.
func Project(files Files, start string) (string, bool, error) {
	dirs, err := searchDirs(files, filepath.Clean(start))
	if err != nil {
		return "", false, err
	}
	for _, dir := range dirs {
		path := filepath.Join(dir, FileName)
		found, err := exists(files, path)
		if err != nil {
			return "", false, err
		}
		if found {
			return path, true, nil
		}
	}
	return "", false, nil
}

// searchDirs lists start and its parents up to the repository root, or only
// start when no parent is a repository root.
func searchDirs(files Files, start string) ([]string, error) {
	var dirs []string
	for dir := start; ; dir = filepath.Dir(dir) {
		dirs = append(dirs, dir)
		root, err := isRepositoryRoot(files, dir)
		if err != nil {
			return nil, err
		}
		if root {
			return dirs, nil
		}
		if filepath.Dir(dir) == dir {
			return dirs[:1], nil
		}
	}
}

func isRepositoryRoot(files Files, dir string) (bool, error) {
	for _, marker := range repositoryMarkers {
		found, err := exists(files, filepath.Join(dir, marker))
		if err != nil || found {
			return found, err
		}
	}
	return false, nil
}

func exists(files Files, path string) (bool, error) {
	_, err := files.Stat(path)
	switch {
	case err == nil:
		return true, nil
	case errors.Is(err, iofs.ErrNotExist):
		return false, nil
	default:
		return false, fmt.Errorf("stat %s: %w", path, err)
	}
}
