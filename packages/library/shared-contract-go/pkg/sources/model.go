// Package sources is the typed model of the source catalog, the loader that
// reads sources.json and its sources.d shards, and the checks that hold across
// sources. Reading goes through the fs and jsonschema ports.
package sources

import (
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
)

// Claim is one statement a source backs.
type Claim struct {
	ID        string `json:"id"`
	Statement string `json:"statement"`
	Anchor    string `json:"anchor,omitempty"`
	Evidence  string `json:"evidence,omitempty"`
}

// Source is one entry of the catalog.
type Source struct {
	ID           string   `json:"id"`
	Kind         Kind     `json:"kind"`
	Title        string   `json:"title"`
	URL          string   `json:"url,omitempty"`
	Path         string   `json:"path,omitempty"`
	Ref          string   `json:"ref,omitempty"`
	Provenance   string   `json:"provenance,omitempty"`
	ReviewedAt   string   `json:"reviewedAt"`
	ReviewAfter  string   `json:"reviewAfter,omitempty"`
	SupersededBy string   `json:"supersededBy,omitempty"`
	Feeds        []string `json:"feeds,omitempty"`
	Claims       []Claim  `json:"claims,omitempty"`

	// File is the catalog file that declared the source. The loader sets it.
	File string `json:"-"`
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
