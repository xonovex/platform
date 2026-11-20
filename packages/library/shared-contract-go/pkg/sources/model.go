// Package sources is the typed model of the source catalog, the loader that
// reads sources.json and its sources.d shards, and the checks that hold across
// sources. Reading goes through the fs and jsonschema ports.
package sources

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"time"
)

const (
	// CatalogFile is the root catalog of a directory.
	CatalogFile = "sources.json"
	// ShardDir holds further catalog files, merged after the root in name order.
	ShardDir = "sources.d"
	// CheckName is the name findings of this package carry.
	CheckName = "sources"

	dateLayout = "2006-01-02"
)

// Kind says how an audit reads a source.
type Kind string

// Source kinds.
const (
	Git        Kind = "git"
	HTTP       Kind = "http"
	Filesystem Kind = "filesystem"
	Web        Kind = "web"
	Repository Kind = "repository"
	// Original marks material written in this repository, with no address.
	Original Kind = "original"
)

// Claim is one statement a source backs.
type Claim struct {
	ID        string `json:"id"`
	Statement string `json:"statement"`
	Anchor    string `json:"anchor,omitempty"`
	Evidence  string `json:"evidence,omitempty"`
}

// Watch names a path of a checkout whose changes since the pinned commit
// affect the files that draw on the source.
type Watch struct {
	Path  string   `json:"path"`
	Feeds []string `json:"feeds,omitempty"`
}

// Note is a prose section of the catalog that belongs to no single source.
type Note struct {
	Title string `json:"title"`
	Text  string `json:"text"`
	// After is the id of the source the section follows. It is empty for a
	// section that comes before the first source.
	After string `json:"after,omitempty"`
}

// Source is one entry of the catalog.
type Source struct {
	ID                  string   `json:"id"`
	Kind                Kind     `json:"kind"`
	Title               string   `json:"title"`
	Citation            string   `json:"citation,omitempty"`
	URLs                []string `json:"urls,omitempty"`
	URLNotes            []string `json:"urlNotes,omitempty"`
	Path                string   `json:"path,omitempty"`
	Ref                 string   `json:"ref,omitempty"`
	Checkout            string   `json:"checkout,omitempty"`
	Version             string   `json:"version,omitempty"`
	Commit              string   `json:"commit,omitempty"`
	Watch               []Watch  `json:"watch,omitempty"`
	ContentSHA256       string   `json:"contentSha256,omitempty"`
	OutputSHA256        string   `json:"outputSha256,omitempty"`
	Provenance          string   `json:"provenance,omitempty"`
	ReviewedAt          string   `json:"reviewedAt"`
	ReviewAfter         string   `json:"reviewAfter,omitempty"`
	SupersededBy        string   `json:"supersededBy,omitempty"`
	Feeds               []string `json:"feeds,omitempty"`
	CoversAllReferences bool     `json:"coversAllReferences,omitempty"`
	Usage               []string `json:"usage,omitempty"`
	Claims              []Claim  `json:"claims,omitempty"`
	Remarks             string   `json:"remarks,omitempty"`

	// File is the catalog file that declared the source. The loader sets it.
	File string `json:"-"`
}

// Document is one catalog file as it is written on disk: the root catalog or a
// shard of sources.d.
type Document struct {
	Schema   string   `json:"$schema,omitempty"`
	Version  int      `json:"version"`
	Title    string   `json:"title,omitempty"`
	Preamble string   `json:"preamble,omitempty"`
	Sources  []Source `json:"sources"`
	Notes    []Note   `json:"notes,omitempty"`
}

// Encode renders a document as indented JSON with a final newline. Equal
// documents give equal bytes.
func Encode(d Document) ([]byte, error) {
	if d.Sources == nil {
		d.Sources = []Source{}
	}
	return encodeJSON(d)
}

func encodeJSON(v any) ([]byte, error) {
	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(v); err != nil {
		return nil, fmt.Errorf("encode: %w", err)
	}
	return buffer.Bytes(), nil
}

// Decode reads one catalog file. Unknown fields are an error, so a file the
// schema accepts always decodes.
func Decode(data []byte) (Document, error) {
	var d Document
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&d); err != nil {
		return Document{}, fmt.Errorf("%w", err)
	}
	return d, nil
}

// Catalog is the merged content of the root catalog and its shards.
type Catalog struct {
	Sources []Source
}

// ByID returns the source with the given id.
func (c Catalog) ByID(id string) (Source, bool) {
	for _, s := range c.Sources {
		if s.ID == id {
			return s, true
		}
	}
	return Source{}, false
}

// Overdue returns the current sources whose review date has come, in the order
// of that date and then of the id. A source that another one supersedes needs
// no review, and a source without a review date is never overdue.
func (c Catalog) Overdue(now time.Time) []Source {
	today := now.UTC().Format(dateLayout)
	var due []Source
	for _, s := range c.Sources {
		if s.SupersededBy == "" && s.ReviewAfter != "" && s.ReviewAfter <= today {
			due = append(due, s)
		}
	}
	sort.Slice(due, func(i, j int) bool {
		if due[i].ReviewAfter != due[j].ReviewAfter {
			return due[i].ReviewAfter < due[j].ReviewAfter
		}
		return due[i].ID < due[j].ID
	})
	return due
}
