// Package runconfig reads the configuration of a run: the run and providers
// sections of each agentic.json layer, the provider shards in providers.d and
// the harness overrides in harnesses.d beside each file. The rule of each key and the layers that may set it are
// data in the run section of shared-config-go. Content a project supplies
// takes effect only after the user approves its digest through the one trust
// gate of shared-config-go.
package runconfig

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"sort"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/agents"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/providers"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/types"
	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/layer"
	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/merge"
	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/resolve"
	providerssection "github.com/xonovex/platform/packages/library/shared-config-go/pkg/section/providers"
	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/shard"
	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/spec"
	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/trust"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/jsonschema"
)

// ProvidersDir is the directory of the provider shards beside each
// agentic.json, and ProviderKind names a provider shard in the approval
// prompt.
const (
	ProvidersDir = "providers.d"
	ProviderKind = "provider"
)

// Env is what the configuration is read through.
type Env struct {
	Host      Host
	Validator jsonschema.Validator
	// Out receives the review and the question of the trust gate.
	Out io.Writer
}

// Request locates the configuration of one command.
type Request struct {
	// Start is the directory the project search begins in, relative to the
	// working directory.
	Start string
	// Project names the project agentic.json instead of the search.
	Project string
	// ApproveDigest is the digest the user approved on the command line.
	ApproveDigest string
	// Flags are the run keys the command line sets.
	Flags []merge.Flag
}

// Pending is project content that takes effect only after the user approves
// its digest.
type Pending struct {
	Subject trust.Subject
	// Origins are the paths of the documents the content defines.
	Origins []string
}

// ProviderDir is the provider shard directory of one layer.
type ProviderDir struct {
	Layer spec.Layer
	Path  string
}

// Loaded is the configuration of a run.
type Loaded struct {
	Resolved Resolved
	// Providers are the provider documents in ascending precedence: user,
	// project, managed. A later document replaces an earlier one of the same
	// name, so a project file never changes a managed provider.
	Providers []providers.Document
	// ProviderDirs are the provider shard directories, managed first.
	ProviderDirs []ProviderDir
	// Harnesses are the facts of every harness with the overrides of every
	// layer applied: user, then project, then managed.
	Harnesses agents.Catalog
	// Pending is the project content behind the trust gate.
	Pending []Pending
}

// UserProviderDir is the provider shard directory of the user layer, empty
// when the user has none.
func (l Loaded) UserProviderDir() string {
	for _, dir := range l.ProviderDirs {
		if dir.Layer == spec.User {
			return dir.Path
		}
	}
	return ""
}

// Load reads the configuration and lets the project content through the
// trust gate.
func Load(env Env, request Request) (Loaded, error) {
	loaded, err := Read(env, request)
	if err != nil {
		return Loaded{}, err
	}
	if err := Gate(env, loaded.Pending, request.ApproveDigest); err != nil {
		return Loaded{}, err
	}
	return loaded, nil
}

// Read reads, checks and merges the configuration without the trust gate, for
// a command that reports on the project content or approves it.
func Read(env Env, request Request) (Loaded, error) {
	files := Files{Host: env.Host}
	places := Places{Host: env.Host}
	dirs, err := providerDirs(places)
	if err != nil {
		return Loaded{}, err
	}
	start, err := absolute(env.Host, request.Start)
	if err != nil {
		return Loaded{}, err
	}
	documents, err := resolve.Load(files, places, env.Validator, resolve.Request{Sections: Sections(), Start: start, Project: request.Project})
	if err != nil {
		return Loaded{}, err
	}
	resolved, err := Merge(documents, request.Flags)
	if err != nil {
		return Loaded{}, err
	}
	shardEnv := shard.Env{Files: files, Places: places, Lister: files}
	shards, err := shard.Find(shardEnv, shard.Request{
		Dir: ProvidersDir, Kind: ProviderKind, Start: start, Project: request.Project,
	})
	if err != nil {
		return Loaded{}, err
	}
	base := agents.Embedded()
	harnessShards, err := shard.Find(shardEnv, shard.Request{
		Dir: agents.OverrideDir, Names: harnessNames(base), Kind: agents.OverrideKind, Start: start, Project: request.Project,
	})
	if err != nil {
		return Loaded{}, err
	}
	overrides := make([]agents.Override, len(harnessShards))
	for i, s := range harnessShards {
		overrides[i] = agents.Override{Name: types.AgentType(s.Name), Origin: s.Path, Data: s.Data}
	}
	harnesses, err := base.Apply(env.Validator, overrides...)
	if err != nil {
		return Loaded{}, err
	}
	loaded := Loaded{Resolved: resolved, ProviderDirs: dirs, Harnesses: harnesses}
	for _, document := range documents {
		if document.Layer == spec.Project {
			loaded.ProviderDirs = append(loaded.ProviderDirs, ProviderDir{Layer: spec.Project, Path: filepath.Join(filepath.Dir(document.Path), ProvidersDir)})
		}
	}
	for _, l := range []spec.Layer{spec.User, spec.Project, spec.Managed} {
		documentsOf, err := layerProviders(l, documents, shards)
		if err != nil {
			return Loaded{}, err
		}
		loaded.Providers = append(loaded.Providers, documentsOf...)
	}
	loaded.Pending = pending(documents, shards, harnessShards)
	return loaded, nil
}

// absolute resolves the start directory against the working directory once,
// so every reader of the project looks in the same directory.
func absolute(host Host, start string) (string, error) {
	if filepath.IsAbs(start) {
		return start, nil
	}
	wd, err := host.Getwd()
	if err != nil {
		return "", fmt.Errorf("find the working directory: %w", err)
	}
	return filepath.Join(wd, start), nil
}

// harnessNames lists the harnesses an override file may name: a file named
// after no harness is not an override.
func harnessNames(catalog agents.Catalog) []string {
	var names []string
	for _, agentType := range catalog.Types() {
		names = append(names, string(agentType))
	}
	return names
}

// providerDirs lists the provider shard directories of the managed and the
// user layer. The project has one only beside a project file, which Read adds.
func providerDirs(places Places) ([]ProviderDir, error) {
	dirs := []ProviderDir{}
	if managed := places.ManagedDir(); managed != "" {
		dirs = append(dirs, ProviderDir{Layer: spec.Managed, Path: filepath.Join(managed, ProvidersDir)})
	}
	user, err := places.UserDir()
	if err != nil {
		return nil, fmt.Errorf("find the user configuration directory: %w", err)
	}
	return append(dirs, ProviderDir{Layer: spec.User, Path: filepath.Join(user, ProvidersDir)}), nil
}

// layerProviders returns the provider documents of one layer: the entries of
// its providers section, then its shards. A name that both define is an
// error, because neither may silently replace the other.
func layerProviders(l spec.Layer, documents []layer.Document, shards []shard.Shard) ([]providers.Document, error) {
	var out []providers.Document
	defined := map[string]string{}
	for _, document := range documents {
		if document.Layer != l {
			continue
		}
		inline, err := inlineProviders(document)
		if err != nil {
			return nil, err
		}
		for _, provider := range inline {
			defined[provider.Preset.Name] = provider.Origin
		}
		out = append(out, inline...)
	}
	for _, s := range shards {
		if s.Layer != l {
			continue
		}
		preset, err := providers.ParsePreset(s.Data)
		if err != nil {
			return nil, fmt.Errorf("provider file %s: %w", s.Path, err)
		}
		if preset.Name != s.Name {
			return nil, fmt.Errorf("provider file %s defines %q; name the file %s%s", s.Path, preset.Name, preset.Name, shard.Extension)
		}
		if origin, taken := defined[preset.Name]; taken {
			return nil, fmt.Errorf("the provider %s is defined in %s and in %s; keep one", preset.Name, origin, s.Path)
		}
		defined[preset.Name] = s.Path
		out = append(out, providers.Document{Preset: preset, Layer: string(l), Origin: s.Path, Raw: s.Data})
	}
	return out, nil
}

// inlineProviders decodes the providers section of one file. The member name
// is the provider name, which the route may repeat but never contradict.
func inlineProviders(document layer.Document) ([]providers.Document, error) {
	section, ok := document.Sections[providerssection.Name]
	if !ok {
		return nil, nil
	}
	var entries map[string]map[string]json.RawMessage
	if err := json.Unmarshal(section, &entries); err != nil {
		return nil, fmt.Errorf("%s: the providers section: %w", document.Path, err)
	}
	names := make([]string, 0, len(entries))
	for name := range entries {
		names = append(names, name)
	}
	sort.Strings(names)
	out := make([]providers.Document, 0, len(names))
	for _, name := range names {
		entry := entries[name]
		quoted, _ := json.Marshal(name)
		if given, set := entry["name"]; set && !bytes.Equal(given, quoted) {
			return nil, fmt.Errorf("%s: the provider %s names itself %s", document.Path, name, given)
		}
		entry["name"] = quoted
		// The entry was decoded from JSON, so it encodes.
		raw, _ := json.Marshal(entry)
		preset, err := providers.ParsePreset(raw)
		if err != nil {
			return nil, fmt.Errorf("%s: the provider %s: %w", document.Path, name, err)
		}
		out = append(out, providers.Document{Preset: preset, Layer: string(document.Layer), Origin: document.Path, Raw: raw})
	}
	return out, nil
}

// pending lists the project content behind the trust gate: the project
// agentic.json when it sets a run or providers section, then each project
// provider shard, then each project harness override.
func pending(documents []layer.Document, providerShards, harnessShards []shard.Shard) []Pending {
	var out []Pending
	for _, document := range documents {
		if document.Layer != spec.Project {
			continue
		}
		if sections := resolve.Gated(document, Sections()); len(sections) > 0 {
			out = append(out, Pending{Subject: trust.FileSubject(document, sections), Origins: []string{document.Path}})
		}
	}
	out = append(out, projectShards(providerShards, ProviderKind)...)
	return append(out, projectShards(harnessShards, agents.OverrideKind)...)
}

func projectShards(shards []shard.Shard, kind string) []Pending {
	var out []Pending
	for _, s := range shards {
		if s.Layer == spec.Project {
			document := layer.Document{Layer: s.Layer, Path: s.Path, Data: s.Data, Digest: s.Digest}
			out = append(out, Pending{Subject: trust.FileSubject(document, []string{kind}), Origins: []string{s.Path}})
		}
	}
	return out
}
