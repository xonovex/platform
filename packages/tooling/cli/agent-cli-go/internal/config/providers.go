package config

import (
	"fmt"
	"path/filepath"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/providers"
)

// providersDirectory is the name of the directory that holds provider files
// beside a configuration file, and below the project directory.
const providersDirectory = "providers"

// ProviderFiles is the port the provider loader reads through.
type ProviderFiles interface {
	providers.Reader
	UserHomeDir() (string, error)
	// ManagedConfigDir is the directory of the system-wide configuration.
	ManagedConfigDir() string
}

// ProviderDir is one directory the loader read, for a listing.
type ProviderDir struct {
	Layer Layer
	Path  string
}

// ProviderSet is every provider document of the layers.
type ProviderSet struct {
	// Documents are in ascending precedence: user, project, managed. A later
	// document replaces an earlier one of the same name, so a project file never
	// changes a managed provider.
	Documents []providers.Document
	// Project is the part of Documents the project supplies, which takes effect
	// only after the user approves its digest.
	Project []providers.Document
	// Dirs are the directories the loader read, managed first.
	Dirs []ProviderDir
}

// ProviderDirs returns the provider directory of each layer: next to the user
// configuration file, below the project directory, and next to the managed
// configuration. userPath is the file --config names, empty for the default file.
func ProviderDirs(files ProviderFiles, userPath, projectDir string) ([]ProviderDir, error) {
	if userPath == "" {
		home, err := files.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("resolve user home directory: %w", err)
		}
		userPath = DefaultPath(home)
	}
	dirs := []ProviderDir{}
	if managed := files.ManagedConfigDir(); managed != "" {
		dirs = append(dirs, ProviderDir{Layer: LayerManaged, Path: filepath.Join(managed, providersDirectory)})
	}
	dirs = append(dirs, ProviderDir{Layer: LayerUser, Path: filepath.Join(filepath.Dir(userPath), providersDirectory)})
	if projectDir != "" {
		dirs = append(dirs, ProviderDir{Layer: LayerProject, Path: filepath.Join(projectDir, projectBase, providersDirectory)})
	}
	return dirs, nil
}

// LoadProviders reads the provider files of the managed, user and project layers.
// A layer without a directory contributes nothing; a file that does not parse
// fails the load and names the file.
func LoadProviders(files ProviderFiles, userPath, projectDir string) (ProviderSet, error) {
	dirs, err := ProviderDirs(files, userPath, projectDir)
	if err != nil {
		return ProviderSet{}, err
	}
	loaded := map[Layer][]providers.Document{}
	for _, dir := range dirs {
		documents, err := providers.LoadDir(files, dir.Path, string(dir.Layer))
		if err != nil {
			return ProviderSet{}, fmt.Errorf("%s providers: %w", dir.Layer, err)
		}
		loaded[dir.Layer] = documents
	}
	set := ProviderSet{Dirs: dirs, Project: loaded[LayerProject]}
	for _, layer := range []Layer{LayerUser, LayerProject, LayerManaged} {
		set.Documents = append(set.Documents, loaded[layer]...)
	}
	return set, nil
}

// ProjectModels returns the "slot=id" overrides the project layer sets.
func ProjectModels(inputs []Input) []string {
	for _, input := range inputs {
		if input.Layer == LayerProject {
			return input.Entries[KeyModels].Items
		}
	}
	return nil
}
