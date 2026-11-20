package trust

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	iofs "io/fs"
	"path/filepath"
	"regexp"
	"sort"
)

// LedgerFile is the name of the ledger in the state directory.
const LedgerFile = "trust.json"

const ledgerVersion = 1

var sha256Digest = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)

// Store is the port the ledger is read and written through. ReadFile wraps
// io/fs.ErrNotExist for a missing file, and WriteFile replaces a file whole and
// creates its directory.
type Store interface {
	ReadFile(path string) ([]byte, error)
	WriteFile(path string, data []byte) error
}

// Ledger keeps the approvals in one JSON file in the state directory.
type Ledger struct {
	Store Store
	// StateDir finds the state directory when an approval is first looked up,
	// so a tool that reads no gated project file needs no state directory.
	StateDir func() (string, error)
}

func (l Ledger) path() (string, error) {
	dir, err := l.StateDir()
	if err != nil {
		return "", fmt.Errorf("the approval ledger: %w", err)
	}
	return filepath.Join(dir, LedgerFile), nil
}

type ledgerFile struct {
	Version   int        `json:"version"`
	Approvals []Approval `json:"approvals"`
}

// Approved reports whether the ledger holds the digest. A missing ledger holds
// nothing.
func (l Ledger) Approved(digest string) (bool, error) {
	path, err := l.path()
	if err != nil {
		return false, err
	}
	approvals, err := read(l.Store, path)
	if err != nil {
		return false, err
	}
	for _, approval := range approvals {
		if approval.Digest == digest {
			return true, nil
		}
	}
	return false, nil
}

// Approve adds the approval, once, and keeps the ledger sorted by digest.
func (l Ledger) Approve(approval Approval) error {
	if !sha256Digest.MatchString(approval.Digest) {
		return fmt.Errorf("approve %q: not a sha256 digest", approval.Digest)
	}
	path, err := l.path()
	if err != nil {
		return err
	}
	approvals, err := read(l.Store, path)
	if err != nil {
		return err
	}
	for _, existing := range approvals {
		if existing.Digest == approval.Digest {
			return nil
		}
	}
	approvals = append(approvals, approval)
	sort.Slice(approvals, func(i, j int) bool { return approvals[i].Digest < approvals[j].Digest })
	return write(l.Store, path, approvals)
}

// Revoke removes the approval of the digest and reports whether the ledger
// held it. A ledger that does not hold it is left as it is.
func (l Ledger) Revoke(digest string) (bool, error) {
	path, err := l.path()
	if err != nil {
		return false, err
	}
	approvals, err := read(l.Store, path)
	if err != nil {
		return false, err
	}
	kept := make([]Approval, 0, len(approvals))
	for _, approval := range approvals {
		if approval.Digest != digest {
			kept = append(kept, approval)
		}
	}
	if len(kept) == len(approvals) {
		return false, nil
	}
	return true, write(l.Store, path, kept)
}

func write(store Store, path string, approvals []Approval) error {
	// The ledger holds only strings, which always encode.
	data, _ := json.MarshalIndent(ledgerFile{Version: ledgerVersion, Approvals: approvals}, "", "  ")
	if err := store.WriteFile(path, append(data, '\n')); err != nil {
		return fmt.Errorf("write the approval ledger %s: %w", path, err)
	}
	return nil
}

func read(store Store, path string) ([]Approval, error) {
	data, err := store.ReadFile(path)
	if errors.Is(err, iofs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read the approval ledger %s: %w", path, err)
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var file ledgerFile
	if err := decoder.Decode(&file); err != nil {
		return nil, fmt.Errorf("approval ledger %s: %w", path, err)
	}
	if file.Version != ledgerVersion {
		return nil, fmt.Errorf("approval ledger %s: version %d, want %d", path, file.Version, ledgerVersion)
	}
	for _, approval := range file.Approvals {
		if !sha256Digest.MatchString(approval.Digest) {
			return nil, fmt.Errorf("approval ledger %s: %q is not a sha256 digest", path, approval.Digest)
		}
	}
	return file.Approvals, nil
}
