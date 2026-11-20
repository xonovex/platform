// Package keychain is the credential store keychain: a generic password in the
// macOS keychain, read through the security command.
package keychain

import (
	"fmt"
	"strings"

	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/credential/shared"
)

// program is the macOS keychain command line tool.
const program = "security"

// Host is the port the store runs the tool through.
type Host interface {
	LookPath(file string) (string, error)
	Capture(name string, args []string, dir string) (string, error)
}

// Store reads a credential from the macOS keychain.
type Store struct{ host Host }

// New returns the store over host.
func New(host Host) Store { return Store{host: host} }

// Available reports whether the security tool is installed.
func (s Store) Available() bool {
	_, err := s.host.LookPath(program)
	return err == nil
}

// Lookup reads the generic password for locator, which is "service" or
// "service/account". The tool prints the value on standard output and nothing
// else, so an error never holds it.
func (s Store) Lookup(locator string) (shared.Secret, error) {
	if err := shared.ValidateLocator(locator); err != nil {
		return shared.Secret{}, err
	}
	service, account, hasAccount := strings.Cut(locator, "/")
	args := []string{"find-generic-password", "-s", service}
	if hasAccount {
		if account == "" {
			return shared.Secret{}, fmt.Errorf("keychain locator %q has an empty account", locator)
		}
		args = append(args, "-a", account)
	}
	args = append(args, "-w")
	value, err := s.host.Capture(program, args, "")
	if err != nil {
		return shared.Secret{}, fmt.Errorf("%w: %w", shared.ErrNotFound, err)
	}
	if value == "" {
		return shared.Secret{}, shared.ErrNotFound
	}
	return shared.NewSecret(value), nil
}
