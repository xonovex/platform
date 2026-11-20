// Package keychain is the credential store keychain: a generic password in the
// macOS keychain, read through the security command.
package keychain

import (
	"fmt"
	"strings"

	"github.com/xonovex/platform/packages/library/shared-launch-go/pkg/credential"
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
func (s Store) Lookup(locator string) (credential.Secret, error) {
	if err := credential.ValidateLocator(locator); err != nil {
		return credential.Secret{}, err
	}
	service, account, hasAccount := strings.Cut(locator, "/")
	args := []string{"find-generic-password", "-s", service}
	if hasAccount {
		if account == "" {
			return credential.Secret{}, fmt.Errorf("keychain locator %q has an empty account", locator)
		}
		args = append(args, "-a", account)
	}
	args = append(args, "-w")
	value, err := s.host.Capture(program, args, "")
	if err != nil {
		return credential.Secret{}, fmt.Errorf("%w: %w", credential.ErrNotFound, err)
	}
	if value == "" {
		return credential.Secret{}, credential.ErrNotFound
	}
	return credential.NewSecret(value), nil
}
