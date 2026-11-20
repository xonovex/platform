package legacy

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/spec"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/runconfig"
)

// Names of the previous release: the file of each layer and the directory of
// its provider files.
const (
	managedBase  = "managed"
	projectBase  = ".agent-cli"
	providersDir = "providers"
)

// extensions are the formats of the managed and the project file, in the
// order they are looked for.
var extensions = []string{".yaml", ".yml", ".toml"}

// place is where one layer of the previous release keeps its file and its
// provider files. Candidates are the possible files, in the order they are
// looked for; the first that exists is the file of the layer.
type place struct {
	layer      spec.Layer
	candidates []string
	providers  string
	// guarded refuses a file that its group or others can write.
	guarded bool
}

// places lists the layers of the previous release, managed first, and the
// project only when projectDir names one.
func places(host runconfig.Host, projectDir string) ([]place, error) {
	var out []place
	if managed := managedDir(host); managed != "" {
		out = append(out, place{
			layer: spec.Managed, candidates: candidates(managed, managedBase),
			providers: filepath.Join(managed, providersDir), guarded: true,
		})
	}
	home, err := host.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("find the home directory: %w", err)
	}
	user := filepath.Join(home, ".config", "sandboxed-claude")
	out = append(out, place{
		layer: spec.User, candidates: []string{filepath.Join(user, "config")},
		providers: filepath.Join(user, providersDir),
	})
	if projectDir == "" {
		return out, nil
	}
	return append(out, place{
		layer: spec.Project, candidates: candidates(projectDir, projectBase),
		providers: filepath.Join(projectDir, projectBase, providersDir),
	}), nil
}

// managedDir is the managed directory of the previous release; empty on a
// Windows host without ProgramData.
func managedDir(host runconfig.Host) string {
	switch host.GOOS() {
	case "darwin":
		return "/Library/Application Support/agent-cli"
	case "windows":
		if data := host.Getenv("ProgramData"); data != "" {
			return filepath.Join(data, "agent-cli")
		}
		return ""
	default:
		return "/etc/agent-cli"
	}
}

func candidates(dir, base string) []string {
	out := make([]string, len(extensions))
	for i, extension := range extensions {
		out[i] = filepath.Join(dir, base+extension)
	}
	return out
}

// find returns the first candidate that exists, or an empty path.
func (p place) find(host runconfig.Host) (string, error) {
	for _, candidate := range p.candidates {
		info, err := host.Stat(candidate)
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		if err != nil {
			return "", fmt.Errorf("%s configuration %s: %w", p.layer, candidate, err)
		}
		if p.guarded && info.Mode().Perm()&0o022 != 0 {
			return "", fmt.Errorf("%s configuration %s is writable by its group or by others; refuse a policy that anyone can change", p.layer, candidate)
		}
		return candidate, nil
	}
	return "", nil
}
