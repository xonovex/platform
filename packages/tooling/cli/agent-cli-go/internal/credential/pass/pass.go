// Package pass is the credential store pass: an entry of the pass password
// manager, whose first line is the secret by convention.
package pass

import (
	"fmt"
	"strings"

	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/credential/shared"
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
func (s Store) Lookup(locator string) (shared.Secret, error) {
	if err := shared.ValidateLocator(locator); err != nil {
		return shared.Secret{}, err
	}
	for _, part := range strings.Split(locator, "/") {
		if part == ".." {
			return shared.Secret{}, fmt.Errorf("pass entry %q climbs out of the store", locator)
		}
	}
	output, err := s.host.Capture(program, []string{"show", locator}, "")
	if err != nil {
		return shared.Secret{}, fmt.Errorf("%w: %w", shared.ErrNotFound, err)
	}
	value, _, _ := strings.Cut(output, "\n")
	value = strings.TrimRight(value, "\r")
	if value == "" {
		return shared.Secret{}, shared.ErrNotFound
	}
	return shared.NewSecret(value), nil
}
