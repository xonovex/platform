// Package trust is the user's approval gate for content a source or a repository
// supplies. One component serves every kind of content: it asks the person at a
// terminal, takes an explicit digest from a flag, fails closed without either,
// and keeps a ledger of the digests the user approved. It reaches the machine
// through ports, so a unit test runs on in-memory doubles.
package trust

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path"
	"strings"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/approval"
)

// ErrDeclined reports an answer other than yes to the approval question.
var ErrDeclined = errors.New("approval declined: the digest was not approved and nothing changed")

// Prompter is the terminal a person answers on.
type Prompter interface {
	// IsTerminal reports whether a person can answer a question on the input and
	// the output.
	IsTerminal() bool
	// ReadLine reads one line of the input without its line ending.
	ReadLine() (string, error)
}

// Decide returns the digest the user approved. A digest given by flag is returned
// unchecked, so the caller's approval check compares it with the content. Without
// a flag, a person at a terminal is asked and yes approves digest, and any other
// answer returns ErrDeclined. Without a flag and without a terminal it returns the
// empty approval, which the check turns into the instruction that names the
// digest.
func Decide(host Prompter, out io.Writer, digest, flagged string) (string, error) {
	if flagged != "" || !host.IsTerminal() {
		return flagged, nil
	}
	yes, err := ask(host, out)
	if err != nil {
		return "", err
	}
	if !yes {
		return "", ErrDeclined
	}
	return digest, nil
}

// ask asks the user to approve the digest just shown. Only y and yes approve; an
// empty answer, any other answer or the end of the input decline.
func ask(host Prompter, out io.Writer) (bool, error) {
	if _, err := fmt.Fprint(out, "Approve this digest? [y/N] "); err != nil {
		return false, err
	}
	answer, err := host.ReadLine()
	if err != nil {
		// The question stays unanswered on a closed input, so nothing is approved.
		_, _ = fmt.Fprintln(out)
		if errors.Is(err, io.EOF) {
			return false, nil
		}
		return false, fmt.Errorf("read the answer: %w", err)
	}
	switch strings.ToLower(strings.TrimSpace(answer)) {
	case "y", "yes":
		return true, nil
	default:
		return false, nil
	}
}

// Files is the port the ledger store reads and writes through.
type Files interface {
	ReadFile(path string) ([]byte, error)
	WriteFile(path string, data []byte, perm fs.FileMode) error
	MkdirAll(path string, perm fs.FileMode) error
	Rename(oldPath, newPath string) error
}

// DefaultFile returns the ledger file: agent-cli/approvals.json below the XDG
// data directory when it is absolute, else below ~/.local/share.
func DefaultFile(getenv func(string) string, home string) string {
	data := getenv("XDG_DATA_HOME")
	if !path.IsAbs(data) {
		data = path.Join(home, ".local", "share")
	}
	return path.Join(data, "agent-cli", "approvals.json")
}

// Store is the ledger file on disk.
type Store struct {
	files Files
	file  string
}

// NewStore returns the store of the ledger file, which must be a clean absolute
// path.
func NewStore(files Files, file string) (Store, error) {
	if !path.IsAbs(file) || path.Clean(file) != file {
		return Store{}, fmt.Errorf("the approval ledger %q is not a clean absolute path", file)
	}
	return Store{files: files, file: file}, nil
}

// File returns the path of the ledger file.
func (s Store) File() string { return s.file }

// Ledger reads the ledger. A store without a file approves nothing, and a corrupt
// file fails closed.
func (s Store) Ledger() (approval.Ledger, error) {
	data, err := s.files.ReadFile(s.file)
	if errors.Is(err, fs.ErrNotExist) {
		return approval.EmptyLedger(), nil
	}
	if err != nil {
		return approval.Ledger{}, err
	}
	return approval.DecodeLedger(data)
}

// Approve records the approval of a digest. The file is replaced through a
// rename, so a reader sees the old ledger or the new one and never a part of it.
func (s Store) Approve(record approval.Approved) error {
	ledger, err := s.Ledger()
	if err != nil {
		return err
	}
	data, err := approval.EncodeLedger(ledger.With(record))
	if err != nil {
		return err
	}
	if err := s.files.MkdirAll(path.Dir(s.file), 0o700); err != nil {
		return err
	}
	staged := s.file + ".tmp"
	if err := s.files.WriteFile(staged, data, 0o600); err != nil {
		return err
	}
	return s.files.Rename(staged, s.file)
}

// Approved reports whether the ledger approves the digest for the kind.
func (s Store) Approved(kind, digest string) (bool, error) {
	ledger, err := s.Ledger()
	if err != nil {
		return false, err
	}
	return ledger.Has(kind, digest), nil
}
