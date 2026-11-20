// Package resolve turns the layers of the agentic configuration into the
// merged sections one tool reads: it finds the files, reads and checks each
// one, passes a project file through the trust gate, and merges the layers with
// the command line.
package resolve

import (
	"errors"
	"fmt"
	"io"
	iofs "io/fs"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/jsonschema"

	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/discover"
	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/layer"
	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/merge"
	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/spec"
	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/trust"
)

// Env holds the ports of a resolution.
type Env struct {
	Files     discover.Files
	Places    discover.Places
	Validator jsonschema.Validator
	Approvals trust.Approvals
	Prompter  trust.Prompter
	// Out receives the approval prompt.
	Out io.Writer
}

// Request is what one tool asks for.
type Request struct {
	// Sections are the sections the tool owns and reads.
	Sections []spec.Section
	// Start is the directory the project search begins in.
	Start string
	// Project names the project file instead of the search, empty to search.
	Project string
	// ApproveDigest is the digest the user approved on the command line.
	ApproveDigest string
	Flags         []merge.Flag
}

// Resolve reads, checks, gates and merges the layers.
func Resolve(env Env, request Request) (merge.Resolved, error) {
	candidates, err := discover.Find(env.Files, env.Places, request.Start, request.Project)
	if err != nil {
		return merge.Resolved{}, err
	}
	var documents []layer.Document
	for _, candidate := range candidates {
		document, found, err := read(env.Files, candidate)
		if err != nil {
			return merge.Resolved{}, err
		}
		if !found {
			continue
		}
		if err := document.Check(env.Validator, request.Sections); err != nil {
			return merge.Resolved{}, err
		}
		if document.Layer == spec.Project {
			gate := trust.Request{Document: document, Sections: gated(document, request.Sections), Flag: request.ApproveDigest}
			if err := trust.Gate(gate, env.Approvals, env.Prompter, env.Out); err != nil {
				return merge.Resolved{}, err
			}
		}
		documents = append(documents, document)
	}
	return merge.Merge(request.Sections, documents, request.Flags)
}

// read reads one candidate. A missing file is no layer unless the user named
// it.
func read(files discover.Files, candidate discover.Candidate) (layer.Document, bool, error) {
	info, err := files.Stat(candidate.Path)
	if errors.Is(err, iofs.ErrNotExist) && !candidate.Required {
		return layer.Document{}, false, nil
	}
	if err != nil {
		return layer.Document{}, false, fmt.Errorf("%s configuration %s: %w", candidate.Layer, candidate.Path, err)
	}
	if !info.Mode.IsRegular() {
		return layer.Document{}, false, fmt.Errorf("%s configuration %s is not a regular file", candidate.Layer, candidate.Path)
	}
	data, err := files.ReadFile(candidate.Path)
	if err != nil {
		return layer.Document{}, false, fmt.Errorf("%s configuration %s: %w", candidate.Layer, candidate.Path, err)
	}
	document, err := layer.Parse(candidate.Layer, candidate.Path, data)
	return document, err == nil, err
}

// gated lists the gated sections among sections that the document sets.
func gated(document layer.Document, sections []spec.Section) []string {
	var names []string
	for _, section := range sections {
		if _, set := document.Sections[section.Name]; set && section.Gated {
			names = append(names, section.Name)
		}
	}
	return names
}
