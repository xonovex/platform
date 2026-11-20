// Package shard reads the override files of one kind, such as
// harnesses.d/<name>.json, that sit beside the agentic.json of each layer: the
// managed directory, the user directory and the directory of the project file.
// A project shard takes effect only after the user approves its digest through
// the trust gate, because it can change the program a tool starts.
package shard

import (
	"errors"
	"fmt"
	"io"
	iofs "io/fs"
	"path/filepath"

	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/discover"
	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/layer"
	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/spec"
	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/trust"
)

// Extension is the suffix of a shard file.
const Extension = ".json"

// Env holds the ports of a read.
type Env struct {
	Files     discover.Files
	Places    discover.Places
	Approvals trust.Approvals
	Prompter  trust.Prompter
	// Out receives the approval prompt.
	Out io.Writer
}

// Request names the shards one tool reads.
type Request struct {
	// Dir is the directory of the shards beside each agentic.json, such as
	// harnesses.d.
	Dir string
	// Names are the shards to look for, each in <Dir>/<name>.json.
	Names []string
	// Kind names the shards in the approval prompt, such as harness.
	Kind string
	// Start is the directory the project search begins in, and Project names
	// the project file instead of the search, as in resolve.Request.
	Start, Project string
	// ApproveDigest is the digest the user approved on the command line.
	ApproveDigest string
}

// Shard is one shard file.
type Shard struct {
	Layer spec.Layer
	Name  string
	Path  string
	// Data is the content of the file, exactly as read.
	Data []byte
}

// Read returns the shards that exist, in the order they apply: user, then
// project, then managed, so the managed shard of a name wins.
func Read(env Env, request Request) ([]Shard, error) {
	candidates, err := discover.Find(env.Files, env.Places, request.Start, request.Project)
	if err != nil {
		return nil, err
	}
	dirs := map[spec.Layer]string{}
	for _, candidate := range candidates {
		dirs[candidate.Layer] = filepath.Join(filepath.Dir(candidate.Path), request.Dir)
	}
	var shards []Shard
	for _, layerName := range []spec.Layer{spec.User, spec.Project, spec.Managed} {
		dir, ok := dirs[layerName]
		if !ok {
			continue
		}
		for _, name := range request.Names {
			shard, found, err := read(env.Files, layerName, name, filepath.Join(dir, name+Extension))
			if err != nil {
				return nil, err
			}
			if !found {
				continue
			}
			if layerName == spec.Project {
				document := layer.Document{Layer: layerName, Path: shard.Path, Data: shard.Data, Digest: layer.Digest(shard.Data)}
				gate := trust.Request{Document: document, Sections: []string{request.Kind}, Flag: request.ApproveDigest}
				if err := trust.Gate(gate, env.Approvals, env.Prompter, env.Out); err != nil {
					return nil, err
				}
			}
			shards = append(shards, shard)
		}
	}
	return shards, nil
}

// read reads one shard file. A missing file is no shard.
func read(files discover.Files, layerName spec.Layer, name, path string) (Shard, bool, error) {
	info, err := files.Stat(path)
	if errors.Is(err, iofs.ErrNotExist) {
		return Shard{}, false, nil
	}
	if err != nil {
		return Shard{}, false, fmt.Errorf("%s shard %s: %w", layerName, path, err)
	}
	if !info.Mode.IsRegular() {
		return Shard{}, false, fmt.Errorf("%s shard %s is not a regular file", layerName, path)
	}
	data, err := files.ReadFile(path)
	if err != nil {
		return Shard{}, false, fmt.Errorf("%s shard %s: %w", layerName, path, err)
	}
	return Shard{Layer: layerName, Name: name, Path: path, Data: data}, true, nil
}
