// Package legacy reads the configuration files of the previous release
// beside agentic.json, and migrates them to agentic.json: the managed
// managed.yaml, managed.yml or managed.toml, the user sandboxed-claude/config,
// the project .agent-cli.yaml, .agent-cli.yml or .agent-cli.toml, the
// providers directory beside each of them, and the approval ledger of the
// previous release. It is the only package that reads YAML or TOML.
package legacy

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"slices"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/providers"
	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/layer"
	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/shard"
	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/spec"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/jsonschema"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/schemas"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/runconfig"
)

// Adapter is the runconfig.Legacy port over the files of the host.
type Adapter struct{}

// found is the legacy content of one layer.
type found struct {
	place     place
	path      string
	config    fileConfig
	providers []providers.Document
}

// read reads the legacy file and the provider files of each layer.
func read(host runconfig.Host, projectDir string) ([]found, error) {
	all, err := places(host, projectDir)
	if err != nil {
		return nil, err
	}
	var out []found
	for _, p := range all {
		f := found{place: p}
		if f.path, err = p.find(host); err != nil {
			return nil, err
		}
		if f.path != "" {
			data, err := host.ReadFile(f.path)
			if err != nil {
				return nil, fmt.Errorf("%s configuration %s: %w", p.layer, f.path, err)
			}
			if f.config, err = parse(data, filepath.Ext(f.path)); err != nil {
				return nil, fmt.Errorf("%s configuration %s: %w", p.layer, f.path, err)
			}
		}
		if f.providers, err = providers.LoadDir(host, p.providers, string(p.layer)); err != nil {
			return nil, fmt.Errorf("%s providers: %w", p.layer, err)
		}
		if f.path != "" || len(f.providers) > 0 {
			out = append(out, f)
		}
	}
	return out, nil
}

// Read returns the legacy content of every layer as runconfig reads it: the
// run section of each file as a document, the provider files, and the
// project material the ledgers do not approve.
func (Adapter) Read(env runconfig.Env, projectDir string) (runconfig.LegacyContent, error) {
	layers, err := read(env.Host, projectDir)
	if err != nil {
		return runconfig.LegacyContent{}, err
	}
	var content runconfig.LegacyContent
	for _, f := range layers {
		var origins []string
		if f.path != "" {
			path := f.path + runconfig.LegacySuffix
			document, err := checked(env.Validator, f.place.layer, path, f.config.agentic())
			if err != nil {
				return runconfig.LegacyContent{}, err
			}
			content.Documents = append(content.Documents, document)
			content.Files = append(content.Files, runconfig.LegacyFile{Layer: f.place.layer, Path: f.path})
		}
		if len(f.providers) > 0 {
			content.Files = append(content.Files, runconfig.LegacyFile{Layer: f.place.layer, Path: f.place.providers})
		}
		for _, document := range f.providers {
			document.Origin += runconfig.LegacySuffix
			content.Providers = append(content.Providers, document)
			origins = append(origins, document.Origin)
		}
		if f.place.layer != spec.Project {
			continue
		}
		m := material{dir: projectDir, documents: f.providers, models: f.config.Models}
		pending, err := m.pending(env.Host, origins)
		if err != nil {
			return runconfig.LegacyContent{}, err
		}
		content.Pending = pending
	}
	return content, nil
}

// checked parses an agentic.json made from a legacy file and checks it
// against the schemas, so a legacy value the new file cannot hold names the
// legacy file.
func checked(validator jsonschema.Validator, l spec.Layer, path string, data []byte) (layer.Document, error) {
	document, err := layer.Parse(l, path, data)
	if err == nil {
		err = document.Check(validator, runconfig.Sections())
	}
	return document, err
}

// output is one file Migrate writes.
type output struct {
	path string
	data []byte
	from []string
}

// Migrate writes the legacy content of the user layer, of the project layer
// when the request names a project directory, and of the managed layer when
// the request asks for it, as agentic.json and providers.d shards beside it.
// It checks every file against the schemas and refuses to replace any file
// before it writes the first one.
func (Adapter) Migrate(env runconfig.Env, request runconfig.Migration) ([]runconfig.Migrated, error) {
	layers, err := read(env.Host, request.ProjectDir)
	if err != nil {
		return nil, err
	}
	var outputs []output
	for _, f := range layers {
		if f.place.layer == spec.Managed && !request.Managed {
			continue
		}
		dir, err := targetDir(env.Host, f.place.layer, request.ProjectDir)
		if err != nil {
			return nil, err
		}
		layerOutputs, err := migrateLayer(env.Validator, f, dir)
		if err != nil {
			return nil, err
		}
		outputs = append(outputs, layerOutputs...)
	}
	for _, o := range outputs {
		if _, err := env.Host.Stat(o.path); !errors.Is(err, fs.ErrNotExist) {
			if err == nil {
				return nil, fmt.Errorf("%s exists; migrate never replaces a file, so move its content by hand or remove it first", o.path)
			}
			return nil, fmt.Errorf("check %s: %w", o.path, err)
		}
	}
	migrated := make([]runconfig.Migrated, 0, len(outputs))
	for _, o := range outputs {
		if err := env.Host.MkdirAll(filepath.Dir(o.path), 0o755); err != nil {
			return migrated, fmt.Errorf("create the directory of %s: %w", o.path, err)
		}
		if err := env.Host.WriteFile(o.path, o.data, 0o644); err != nil {
			return migrated, fmt.Errorf("write %s: %w", o.path, err)
		}
		migrated = append(migrated, runconfig.Migrated{Path: o.path, From: o.from})
	}
	return migrated, nil
}

// targetDir is the directory of the agentic.json of a layer.
func targetDir(host runconfig.Host, l spec.Layer, projectDir string) (string, error) {
	places := runconfig.Places{Host: host}
	switch l {
	case spec.Managed:
		return places.ManagedDir(), nil
	case spec.User:
		dir, err := places.UserDir()
		if err != nil {
			return "", fmt.Errorf("find the user configuration directory: %w", err)
		}
		return dir, nil
	default:
		return projectDir, nil
	}
}

// migrateLayer returns the agentic.json and the provider shards of one layer.
func migrateLayer(validator jsonschema.Validator, f found, dir string) ([]output, error) {
	var outputs []output
	if f.path != "" {
		path := filepath.Join(dir, "agentic.json")
		data := f.config.agentic()
		if _, err := checked(validator, f.place.layer, path, data); err != nil {
			return nil, err
		}
		outputs = append(outputs, output{path: path, data: data, from: []string{f.path}})
	}
	for _, document := range f.providers {
		path := filepath.Join(dir, runconfig.ProvidersDir, document.Preset.Name+shard.Extension)
		violations, err := validator.Validate(schemas.Provider, document.Raw)
		if err != nil {
			return nil, fmt.Errorf("check %s: %w", document.Origin, err)
		}
		if len(violations) > 0 {
			return nil, fmt.Errorf("%s does not match the provider schema: %s %s", document.Origin, violations[0].Pointer, violations[0].Message)
		}
		outputs = append(outputs, output{path: path, data: slices.Clone(document.Raw), from: []string{document.Origin}})
	}
	return outputs, nil
}
