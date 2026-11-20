// Package observe is the port for reading the current state of a source. A
// Transport reads one kind of address, such as a file, a git checkout or a web
// page, and reports what it found as an Observation. The rules package judges
// the observation, so a transport never decides whether a source has drifted.
package observe

import (
	"context"
	"crypto/sha256"
	"encoding/hex"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/sources"
)

// Status is the outcome of reading one document.
type Status string

// Statuses of a document.
const (
	// OK means the document was read.
	OK Status = "ok"
	// Missing means the address does not exist.
	Missing Status = "missing"
	// Restricted means the provider refused to answer, so the content is not
	// known, but the address may still be fine.
	Restricted Status = "restricted"
	// Failed means the provider could not be reached or answered with an error.
	Failed Status = "error"
)

// Request asks a transport to read one source.
type Request struct {
	Source sources.Source
	// Dir is the directory of the catalog. Relative addresses of the source
	// resolve against it.
	Dir string
	// Pull lets a transport that keeps a local copy refresh it first.
	Pull bool
}

// Document is the outcome of reading one address of a source.
type Document struct {
	Locator string `json:"locator"`
	Status  Status `json:"status"`
	Detail  string `json:"detail,omitempty"`
	// SHA256 is the hex digest of the body, empty when no body was read.
	SHA256 string `json:"sha256,omitempty"`
	// Body is the content, kept for the anchor check and left out of reports.
	Body []byte `json:"-"`
}

// Git is what a transport found in the local checkout of a source.
type Git struct {
	Checkout string `json:"checkout"`
	// Resolved is false when the checkout is missing or holds no repository.
	Resolved bool `json:"resolved"`
	// PullFailed is true when refreshing the tags failed.
	PullFailed bool `json:"pullFailed,omitempty"`
	// NoRef is true when the source pins neither a commit nor a version.
	NoRef bool `json:"noRef,omitempty"`
	// RefMissing is true when the pinned ref is not in the checkout.
	RefMissing bool `json:"refMissing,omitempty"`
	// Ref is the commit or version the diff starts from.
	Ref           string `json:"ref,omitempty"`
	PinnedVersion string `json:"pinnedVersion,omitempty"`
	LatestVersion string `json:"latestVersion,omitempty"`
	// Behind is true when the latest release is newer than the pinned version.
	Behind bool `json:"behind,omitempty"`
	// CommitCount is the number of commits on the watched paths since Ref.
	CommitCount int `json:"commitCount"`
	// Commits lists the first of those commits, one line each.
	Commits      []string `json:"commits,omitempty"`
	ChangedFiles []string `json:"changedFiles,omitempty"`
	// ReviewFeeds lists the files that draw on the changed watched paths.
	ReviewFeeds []string `json:"reviewFeeds,omitempty"`
}

// Observation is what one transport found for one source.
type Observation struct {
	Transport string     `json:"transport"`
	Documents []Document `json:"documents,omitempty"`
	Git       *Git       `json:"git,omitempty"`
}

// Transport reads the sources whose address it understands.
type Transport interface {
	// Name identifies the transport in reports.
	Name() string
	// Remote reports whether reading uses the network. A remote transport runs
	// only when the audit is asked to fetch.
	Remote() bool
	// Applies reports whether the transport can read the source.
	Applies(source sources.Source) bool
	// Observe reads the source. A source that is missing or refuses to answer is
	// part of the Observation. The error is for a transport that cannot work at
	// all, such as a program that is not installed.
	Observe(ctx context.Context, request Request) (Observation, error)
}

// Hash returns the hex SHA-256 digest of a body.
func Hash(body []byte) string {
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:])
}

// DigestPrefix starts every digest the source catalog records.
const DigestPrefix = "sha256:"

// Aggregate returns the digest that a source records for its documents: the
// digest of each locator followed by the digest of its body, in order. It
// returns false when a document was not read, because the digest would then
// stand for less than the whole source.
func Aggregate(documents []Document) (string, bool) {
	if len(documents) == 0 {
		return "", false
	}
	sum := sha256.New()
	for _, d := range documents {
		if d.Status != OK || d.SHA256 == "" {
			return "", false
		}
		sum.Write([]byte(d.Locator))
		sum.Write([]byte{0})
		sum.Write([]byte(d.SHA256))
		sum.Write([]byte{'\n'})
	}
	return DigestPrefix + hex.EncodeToString(sum.Sum(nil)), true
}
