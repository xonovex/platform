package config

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
)

// managedBase and projectBase name the files of the managed and the project
// layer; the extension selects the format, and the first one present is read.
const (
	managedBase = "managed"
	projectBase = ".agent-cli"
)

var layerExtensions = []string{".yaml", ".yml", ".toml"}

// LoadLayers reads the managed, user and project layers. The user layer is the
// default configuration file, which may be absent, or the file userPath names,
// which must exist. The project layer is the file in projectDir. A layer without
// a file contributes nothing, and the command line is not read here.
func LoadLayers(files Files, userPath, projectDir string) ([]Input, error) {
	var inputs []Input

	managedDir := files.ManagedConfigDir()
	managed, found, err := loadFirst(files, managedDir, managedBase, LayerManaged, true)
	if err != nil {
		return nil, err
	}
	if found {
		inputs = append(inputs, managed)
	}

	user, err := loadUser(files, userPath)
	if err != nil {
		return nil, err
	}
	if user != nil {
		inputs = append(inputs, *user)
	}

	project, found, err := loadFirst(files, projectDir, projectBase, LayerProject, false)
	if err != nil {
		return nil, err
	}
	if found {
		inputs = append(inputs, project)
	}
	return inputs, nil
}

// loadUser reads the user layer. An empty path is the default file, which may be
// absent.
func loadUser(files Files, path string) (*Input, error) {
	if path == "" {
		home, err := files.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("resolve user home directory: %w", err)
		}
		path = DefaultPath(home)
		if _, err := files.Stat(path); errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
	}
	config, err := Load(files, path)
	if err != nil {
		return nil, fmt.Errorf("user configuration %s: %w", path, err)
	}
	return &Input{Layer: LayerUser, Origin: path, Entries: config.Entries()}, nil
}

// loadFirst reads the first file of dir named base with a known extension. A
// managed file that others can write is refused, because the layer is only as
// strong as the write protection of its file.
func loadFirst(files Files, dir, base string, layer Layer, guarded bool) (Input, bool, error) {
	if dir == "" {
		return Input{}, false, nil
	}
	for _, extension := range layerExtensions {
		path := filepath.Join(dir, base+extension)
		info, err := files.Stat(path)
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		if err != nil {
			return Input{}, false, fmt.Errorf("%s configuration %s: %w", layer, path, err)
		}
		if guarded && info.Mode().Perm()&0o022 != 0 {
			return Input{}, false, fmt.Errorf("%s configuration %s is writable by its group or by others; refuse a policy that anyone can change", layer, path)
		}
		data, err := files.ReadFile(path)
		if err != nil {
			return Input{}, false, fmt.Errorf("%s configuration %s: %w", layer, path, err)
		}
		config, err := Parse(data, extension)
		if err != nil {
			return Input{}, false, fmt.Errorf("%s configuration %s: %w", layer, path, err)
		}
		return Input{Layer: layer, Origin: path, Entries: config.Entries()}, true, nil
	}
	return Input{}, false, nil
}
