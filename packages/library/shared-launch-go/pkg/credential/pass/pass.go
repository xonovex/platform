// Package pass is the credential store pass: an entry of the pass password
// manager, whose first line is the secret by convention.
package pass

import (
	"fmt"
	"strings"

	"github.com/xonovex/platform/packages/library/shared-launch-go/pkg/credential"
)

// program is the pass command line tool.
const program = "pass"

// Host is the port the store runs the tool through.
type Host interface {
	LookPath(file string) (string, error)
	Capture(name string, args []string, dir string) (string, error)
}

// Store reads an entry from pass.
type Store struct{ host Host }

// New returns the store over host.
func New(host Host) Store { return Store{host: host} }

// Available reports whether pass is installed.
func (s Store) Available() bool {
	_, err := s.host.LookPath(program)
	return err == nil
}

// Lookup reads the first line of the entry named by locator. It refuses an entry
// name that climbs out of the store.
func (s Store) Lookup(locator string) (credential.Secret, error) {
	if err := credential.ValidateLocator(locator); err != nil {
		return credential.Secret{}, err
	}
	for _, part := range strings.Split(locator, "/") {
		if part == ".." {
			return credential.Secret{}, fmt.Errorf("pass entry %q climbs out of the store", locator)
		}
	}
	output, err := s.host.Capture(program, []string{"show", locator}, "")
	if err != nil {
		return credential.Secret{}, fmt.Errorf("%w: %w", credential.ErrNotFound, err)
	}
	value, _, _ := strings.Cut(output, "\n")
	value = strings.TrimRight(value, "\r")
	if value == "" {
		return credential.Secret{}, credential.ErrNotFound
	}
	return credential.NewSecret(value), nil
}
