// Package file is the credential store file: a credential held in a file that
// only its owner can read.
package file

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/credential/shared"
)

// Host is the port the store reads the file through.
type Host interface {
	Stat(path string) (fs.FileInfo, error)
	ReadFile(path string) ([]byte, error)
}

// Store reads a credential from a file.
type Store struct{ host Host }

// New returns the store over host.
func New(host Host) Store { return Store{host: host} }

// Available is always true.
func (Store) Available() bool { return true }

// Lookup reads the file at the absolute path locator. It refuses a file that is
// not a regular file, and a file that group or others can read, because a
// credential another account can read is already leaked. It strips the trailing
// line break an editor adds.
func (s Store) Lookup(locator string) (shared.Secret, error) {
	if !filepath.IsAbs(locator) {
		return shared.Secret{}, fmt.Errorf("credential file %q must be an absolute path", locator)
	}
	info, err := s.host.Stat(locator)
	if err != nil {
		return shared.Secret{}, fmt.Errorf("inspect credential file: %w", err)
	}
	if !info.Mode().IsRegular() {
		return shared.Secret{}, fmt.Errorf("credential file %q is not a regular file", locator)
	}
	if info.Mode().Perm()&0o077 != 0 {
		return shared.Secret{}, fmt.Errorf("credential file %q is readable by group or others; run chmod 600", locator)
	}
	data, err := s.host.ReadFile(locator)
	if err != nil {
		return shared.Secret{}, fmt.Errorf("read credential file: %w", err)
	}
	value := strings.TrimRight(string(data), "\r\n")
	if value == "" {
		return shared.Secret{}, shared.ErrNotFound
	}
	return shared.NewSecret(value), nil
}
