// Package approval is the trust gate of content a repository or a source
// supplies: the user approves the digest of exactly the content that was shown,
// and any change to the content needs a new approval. The package holds the pure
// check and the ledger of approved digests; the prompt and the files belong to
// the caller.
package approval

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/artifact"
)

// Check fails unless the user approved exactly the digest of the content that
// was shown. An empty approval is the preview step: the error names the digest
// to approve with the flag.
func Check(digest, approved, flag string) error {
	switch {
	case approved == "":
		return fmt.Errorf("approval required: review the content, then repeat the command with %s %s", flag, digest)
	case !artifact.ValidDigest(approved):
		return fmt.Errorf("the approved digest %q is not a sha256 digest", approved)
	case approved != digest:
		return fmt.Errorf("the approved digest %s differs from the digest of the content %s", approved, digest)
	}
	return nil
}

// LedgerVersion is the version of the ledger file format.
const LedgerVersion = 1

// Approved records one digest the user approved. Subject names what the digest
// covers, for a person who reads the ledger.
type Approved struct {
	Kind    string `json:"kind"`
	Subject string `json:"subject"`
	Digest  string `json:"digest"`
}

// Ledger is the set of approved digests, sorted by kind and digest.
type Ledger struct {
	Version  int        `json:"version"`
	Approved []Approved `json:"approved"`
}

// EmptyLedger returns a ledger that approves nothing.
func EmptyLedger() Ledger { return Ledger{Version: LedgerVersion} }

// Has reports whether the ledger approves the digest for the kind.
func (l Ledger) Has(kind, digest string) bool {
	return slices.ContainsFunc(l.Approved, func(a Approved) bool { return a.Kind == kind && a.Digest == digest })
}

// With returns a ledger that also approves the record. An approval of the same
// kind and digest keeps its place and takes the new subject.
func (l Ledger) With(record Approved) Ledger {
	entries := slices.DeleteFunc(slices.Clone(l.Approved), func(a Approved) bool {
		return a.Kind == record.Kind && a.Digest == record.Digest
	})
	entries = append(entries, record)
	slices.SortFunc(entries, func(a, b Approved) int {
		if c := strings.Compare(a.Kind, b.Kind); c != 0 {
			return c
		}
		return strings.Compare(a.Digest, b.Digest)
	})
	return Ledger{Version: LedgerVersion, Approved: entries}
}

// EncodeLedger returns the ledger file bytes: indented JSON with a final
// newline. It refuses a record that holds no kind or no valid digest.
func EncodeLedger(ledger Ledger) ([]byte, error) {
	if err := validate(ledger); err != nil {
		return nil, err
	}
	if ledger.Approved == nil {
		ledger.Approved = []Approved{}
	}
	// The values are plain strings and numbers, which cannot fail to encode.
	data, _ := json.MarshalIndent(ledger, "", "  ")
	return append(data, '\n'), nil
}

// DecodeLedger reads ledger file bytes and fails closed on an unknown key, a
// version other than the one this code writes, an invalid record or trailing
// data.
func DecodeLedger(data []byte) (Ledger, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var ledger Ledger
	if err := decoder.Decode(&ledger); err != nil {
		return Ledger{}, fmt.Errorf("approval ledger: %w", err)
	}
	if decoder.More() {
		return Ledger{}, errors.New("approval ledger: trailing data after the JSON document")
	}
	if err := validate(ledger); err != nil {
		return Ledger{}, err
	}
	return ledger, nil
}

func validate(ledger Ledger) error {
	if ledger.Version != LedgerVersion {
		return fmt.Errorf("approval ledger: version %d is not supported; this code writes version %d", ledger.Version, LedgerVersion)
	}
	seen := map[[2]string]bool{}
	for _, record := range ledger.Approved {
		if record.Kind == "" {
			return errors.New("approval ledger: a record has no kind")
		}
		if !artifact.ValidDigest(record.Digest) {
			return fmt.Errorf("approval ledger: the digest %q of a %s record is not a sha256 digest", record.Digest, record.Kind)
		}
		id := [2]string{record.Kind, record.Digest}
		if seen[id] {
			return fmt.Errorf("approval ledger: the %s digest %s repeats", record.Kind, record.Digest)
		}
		seen[id] = true
	}
	return nil
}
