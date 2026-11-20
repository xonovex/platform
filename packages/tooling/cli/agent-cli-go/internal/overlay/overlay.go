// Package overlay is the adapter that builds the per-run artifact overlay: it
// gathers the artifacts of installed plugins and of directories, lets the dialect
// of the agent place them, and writes the placement into a private directory the
// sandbox mounts. The dialects are pure; this package does the I/O through its
// ports.
package overlay

import (
	"errors"
	"fmt"
	"io/fs"
	"path"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/artifact"
	libplugin "github.com/xonovex/platform/packages/library/shared-agent-go/pkg/overlay"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/overlay/shared"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/launch"
	origin "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/origin/shared"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/plugin"
)

// dryRunRoot is the directory a dry run reports. Nothing is written there.
const dryRunRoot = "/tmp/agent-cli-overlay-dry-run"

// Host is the machine the builder consults.
type Host interface {
	plugin.Files
	Getenv(key string) string
	UserHomeDir() (string, error)
	MkdirTemp(pattern string) (string, error)
}

// Builder builds the overlay of a run. It implements launch.Overlays.
type Builder struct {
	Host    Host
	Origins origin.Registry
}

var _ launch.Overlays = Builder{}

// Prepare gathers the artifacts, places them with the dialect of the agent, and
// writes them. A request without a dry run creates the overlay directory; a dry
// run writes nothing.
func (b Builder) Prepare(request launch.OverlayRequest) (launch.OverlayResult, error) {
	dialect, err := libplugin.Resolve(request.Agent.Type)
	if err != nil {
		return launch.OverlayResult{}, err
	}
	set, sources, err := b.gather(request)
	if err != nil {
		return launch.OverlayResult{}, err
	}
	root, cleanup, err := b.root(request.DryRun)
	if err != nil {
		return launch.OverlayResult{}, err
	}
	result, err := b.place(dialect, set, request, root)
	if err != nil || result.Root == "" {
		cleanup()
	}
	if err != nil {
		return launch.OverlayResult{}, err
	}
	result.Cleanup = cleanup
	result.Sources = sources
	return result, nil
}

// gather reads the artifacts of the installed plugins and of the directories and
// merges them, and lists each source with the digest of its content. Two sources
// that define one artifact name fail the merge.
func (b Builder) gather(request launch.OverlayRequest) (artifact.Set, []launch.OverlaySource, error) {
	var sets []artifact.Set
	var sources []launch.OverlaySource
	if len(request.Plugins) > 0 {
		service, err := b.service(request.Store)
		if err != nil {
			return artifact.Set{}, nil, err
		}
		locked, err := service.Locked(request.Plugins)
		if err != nil {
			return artifact.Set{}, nil, err
		}
		for _, record := range locked {
			sources = append(sources, launch.OverlaySource{Kind: launch.SourcePlugin, Name: record.Name, Digest: record.Digest})
		}
		set, err := service.Load(request.Plugins)
		if err != nil {
			return artifact.Set{}, nil, err
		}
		sets = append(sets, set)
	}
	for _, dir := range request.Directories {
		tree, err := origin.ReadTree(b.Host, dir)
		if err != nil {
			return artifact.Set{}, nil, fmt.Errorf("overlay directory %s: %w", dir, err)
		}
		set, err := artifact.FromTree(dir, tree)
		if err != nil {
			return artifact.Set{}, nil, fmt.Errorf("overlay directory %s: %w", dir, err)
		}
		sets = append(sets, set)
		sources = append(sources, launch.OverlaySource{Kind: launch.SourceDirectory, Name: dir, Digest: tree.Digest()})
	}
	merged, err := artifact.Merge(sets...)
	if err != nil {
		return artifact.Set{}, nil, err
	}
	return merged, sources, nil
}

func (b Builder) service(dir string) (plugin.Service, error) {
	if dir == "" {
		home, err := b.Host.UserHomeDir()
		if err != nil {
			return plugin.Service{}, fmt.Errorf("resolve user home directory: %w", err)
		}
		dir = plugin.DefaultDir(b.Host.Getenv, home)
	}
	store, err := plugin.NewStore(b.Host, dir)
	if err != nil {
		return plugin.Service{}, err
	}
	return plugin.Service{Store: store, Files: b.Host, Origins: b.Origins}, nil
}

// root creates the private directory of the overlay, or names the placeholder of
// a dry run. The returned cleanup removes the directory.
func (b Builder) root(dryRun bool) (string, func(), error) {
	if dryRun {
		return dryRunRoot, func() {}, nil
	}
	root, err := b.Host.MkdirTemp("agent-cli-overlay-*")
	if err != nil {
		return "", nil, fmt.Errorf("create the overlay directory: %w", err)
	}
	return root, func() { _ = b.Host.RemoveAll(root) }, nil
}

// place runs the dialect and writes its files below root.
func (b Builder) place(dialect shared.Dialect, set artifact.Set, request launch.OverlayRequest, root string) (launch.OverlayResult, error) {
	native, err := b.nativeInstructions(dialect.Facts(), request.WorkDir)
	if err != nil {
		return launch.OverlayResult{}, err
	}
	placement, err := dialect.Place(shared.Input{Artifacts: set, Root: root, NativeInstructions: native})
	if err != nil {
		return launch.OverlayResult{}, err
	}
	summary := summarize(placement, request.DryRun)
	if len(placement.Files) == 0 {
		return launch.OverlayResult{Summary: summary}, nil
	}
	if !request.DryRun {
		if err := b.write(root, placement.Files); err != nil {
			return launch.OverlayResult{}, err
		}
	}
	result := launch.OverlayResult{Root: root, Environment: placement.Environment, Arguments: placement.Arguments, Summary: summary}
	if placement.RootWritable {
		result.BindPaths = []string{root}
	} else {
		result.RoBindPaths = []string{root}
	}
	return result, nil
}

// nativeInstructions lists the project instruction files of the agent that exist
// in the work directory.
func (b Builder) nativeInstructions(facts shared.Facts, workDir string) ([]string, error) {
	var found []string
	for _, name := range facts.NativeInstructions {
		if _, err := b.Host.Stat(path.Join(workDir, name)); err == nil {
			found = append(found, name)
		} else if !errors.Is(err, fs.ErrNotExist) {
			return nil, err
		}
	}
	return found, nil
}

func (b Builder) write(root string, files []shared.File) error {
	for _, file := range files {
		full := path.Join(root, file.Path)
		if err := b.Host.MkdirAll(path.Dir(full), 0o755); err != nil {
			return err
		}
		perm := fs.FileMode(0o644)
		if file.Executable {
			perm = 0o755
		}
		if err := b.Host.WriteFile(full, file.Content, perm); err != nil {
			return err
		}
	}
	return nil
}

// summarize describes a placement for the user: the number of files placed, and
// each skipped artifact with its reason.
func summarize(placement shared.Placement, dryRun bool) []string {
	var lines []string
	if len(placement.Files) > 0 {
		verb := "placed"
		if dryRun {
			verb = "would place"
		}
		lines = append(lines, fmt.Sprintf("%s %d files at the overlay root", verb, len(placement.Files)))
	}
	for _, skip := range placement.Skipped {
		lines = append(lines, fmt.Sprintf("skipped %s %s: %s", skip.Kind, skip.Name, skip.Reason))
	}
	return lines
}
