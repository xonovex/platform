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
	"sort"
	"strings"

	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/discover"
	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/layer"
	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/spec"
	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/trust"
)

// Extension is the suffix of a shard file.
const Extension = ".json"

// Lister is the port that names the entries of a directory. ReadDir wraps
// io/fs.ErrNotExist for a missing directory.
type Lister interface {
	ReadDir(path string) ([]string, error)
}

// Env holds the ports of a read.
type Env struct {
	Files  discover.Files
	Places discover.Places
	// Lister names the shards of a directory for a request without names; a
	// request with names needs none.
	Lister    Lister
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
	// Names are the shards to look for, each in <Dir>/<name>.json. Without
	// names every <name>.json of each directory is a shard.
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
	// Digest is the sha256 digest of Data, the value a user approves.
	Digest string
}

// Read returns the shards that exist, in the order they apply: user, then
// project, then managed, so the managed shard of a name wins. A project shard
// passes the trust gate first.
func Read(env Env, request Request) ([]Shard, error) {
	shards, err := Find(env, request)
	if err != nil {
		return nil, err
	}
	for _, shard := range shards {
		if shard.Layer != spec.Project {
			continue
		}
		document := layer.Document{Layer: shard.Layer, Path: shard.Path, Data: shard.Data, Digest: shard.Digest}
		gate := trust.Request{Document: document, Sections: []string{request.Kind}, Flag: request.ApproveDigest}
		if err := trust.Gate(gate, env.Approvals, env.Prompter, env.Out); err != nil {
			return nil, err
		}
	}
	return shards, nil
}

// Find returns the shards that exist in the order Read applies them, without
// the trust gate, for a tool that gates a project shard itself or only reports
// on it.
func Find(env Env, request Request) ([]Shard, error) {
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
		names, err := shardNames(env.Lister, request.Names, layerName, dir)
		if err != nil {
			return nil, err
		}
		for _, name := range names {
			shard, found, err := read(env.Files, layerName, name, filepath.Join(dir, name+Extension))
			if err != nil {
				return nil, err
			}
			if found {
				shards = append(shards, shard)
			}
		}
	}
	return shards, nil
}

// shardNames returns the names a request looks for, or, without names, the
// name of every shard file in dir in sorted order.
func shardNames(lister Lister, names []string, layerName spec.Layer, dir string) ([]string, error) {
	if names != nil {
		return names, nil
	}
	if lister == nil {
		return nil, errors.New("a read of every shard needs a Lister")
	}
	entries, err := lister.ReadDir(dir)
	if errors.Is(err, iofs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("%s shards %s: %w", layerName, dir, err)
	}
	var found []string
	for _, entry := range entries {
		if name, ok := strings.CutSuffix(entry, Extension); ok && name != "" {
			found = append(found, name)
		}
	}
	sort.Strings(found)
	return found, nil
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
	return Shard{Layer: layerName, Name: name, Path: path, Data: data, Digest: layer.Digest(data)}, true, nil
}
