package providers

import (
	"errors"
	"fmt"
	"io/fs"
	"path"
	"strings"
)

// Reader is the file port the loader reads through.
type Reader interface {
	ReadDir(dir string) ([]fs.DirEntry, error)
	ReadFile(file string) ([]byte, error)
}

// Extension is the suffix of a provider file.
const Extension = ".json"

// LoadDir reads every provider file in dir, in file name order, and labels each
// document with layer. A missing directory holds no provider. A file that does
// not parse, or whose name differs from the provider name it defines, fails the
// load and names the file.
func LoadDir(reader Reader, dir, layer string) ([]Document, error) {
	entries, err := reader.ReadDir(dir)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read providers directory %s: %w", dir, err)
	}
	var documents []Document
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), Extension) {
			continue
		}
		file := path.Join(dir, entry.Name())
		data, err := reader.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("read provider file %s: %w", file, err)
		}
		preset, err := ParsePreset(data)
		if err != nil {
			return nil, fmt.Errorf("provider file %s: %w", file, err)
		}
		if want := strings.TrimSuffix(entry.Name(), Extension); want != preset.Name {
			return nil, fmt.Errorf("provider file %s defines %q; name the file %s%s", file, preset.Name, preset.Name, Extension)
		}
		documents = append(documents, Document{Preset: preset, Layer: layer, Origin: file, Raw: data})
	}
	return documents, nil
}
