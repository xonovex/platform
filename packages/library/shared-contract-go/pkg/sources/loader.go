package sources

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"sort"
	"strings"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/findings"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/fs"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/jsonschema"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/schemas"
)

// Loader reads a catalog through the file-system and schema ports.
type Loader struct {
	FS        fs.FileSystem
	Validator jsonschema.Validator
}

// Load reads dir/sources.json and every dir/sources.d/*.json, checks each file
// against the schema, merges the sources and runs Check over the result.
//
// Problems in the content are findings. The error is for a fault of the
// infrastructure, such as a directory that cannot be listed. A directory with
// neither a root catalog nor a shard yields a CodeMissing finding.
//
// A file that breaks the schema adds no sources, so the checks across sources
// run on what the valid files hold.
func (l Loader) Load(dir string) (Catalog, []findings.Finding, error) {
	files, err := l.catalogFiles(dir)
	if err != nil {
		return Catalog{}, nil, err
	}
	if len(files) == 0 {
		missing := finding(findings.Fail, CodeMissing, dir,
			fmt.Sprintf("no %s and no %s/*.json", CatalogFile, ShardDir))
		return Catalog{}, []findings.Finding{missing}, nil
	}

	var catalog Catalog
	var out []findings.Finding
	for _, file := range files {
		loaded, problems, err := l.loadFile(file)
		if err != nil {
			return Catalog{}, nil, err
		}
		out = append(out, problems...)
		catalog.Sources = append(catalog.Sources, loaded...)
	}
	out = append(out, Check(catalog)...)
	return catalog, dropOKIfProblems(out), nil
}

// dropOKIfProblems removes the pass finding of Check when a schema problem
// already failed the load, so a broken catalog never reports a pass.
func dropOKIfProblems(all []findings.Finding) []findings.Finding {
	if findings.Worst(all) == findings.Pass {
		return all
	}
	kept := all[:0:0]
	for _, f := range all {
		if f.Code != CodeOK {
			kept = append(kept, f)
		}
	}
	return kept
}

func (l Loader) catalogFiles(dir string) ([]string, error) {
	var files []string
	root := path.Join(dir, CatalogFile)
	exists, err := fs.Exists(l.FS, root)
	if err != nil {
		return nil, fmt.Errorf("stat %s: %w", root, err)
	}
	if exists {
		files = append(files, root)
	}

	shardDir := path.Join(dir, ShardDir)
	entries, err := l.FS.ReadDir(shardDir)
	if errors.Is(err, fs.ErrNotExist) {
		return files, nil
	}
	if err != nil {
		return nil, fmt.Errorf("list %s: %w", shardDir, err)
	}
	var shards []string
	for _, entry := range entries {
		if !entry.IsDir && strings.HasSuffix(entry.Name, ".json") {
			shards = append(shards, path.Join(shardDir, entry.Name))
		}
	}
	sort.Strings(shards)
	return append(files, shards...), nil
}

func (l Loader) loadFile(file string) ([]Source, []findings.Finding, error) {
	data, err := l.FS.ReadFile(file)
	if err != nil {
		return nil, nil, fmt.Errorf("read %s: %w", file, err)
	}
	violations, err := l.Validator.Validate(schemas.Sources, data)
	if err != nil {
		return nil, nil, fmt.Errorf("validate %s: %w", file, err)
	}
	if len(violations) > 0 {
		problems := make([]findings.Finding, 0, len(violations))
		for _, v := range violations {
			where := v.Pointer
			if where == "" {
				where = "/"
			}
			problems = append(problems, finding(findings.Fail, CodeSchema, file, where+": "+v.Message))
		}
		return nil, problems, nil
	}

	var document struct {
		Schema  string   `json:"$schema"`
		Version int      `json:"version"`
		Sources []Source `json:"sources"`
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&document); err != nil {
		// The schema accepted the file, so a decode failure means the Go
		// model and the schema disagree: a fault of the contract itself.
		return nil, nil, fmt.Errorf("decode %s: %w", file, err)
	}
	for i := range document.Sources {
		document.Sources[i].File = file
	}
	return document.Sources, nil, nil
}
