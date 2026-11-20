// Package secretservice is the credential store secret-service: a secret in a
// Linux secret service such as GNOME Keyring or KWallet, read through secret-tool.
package secretservice

import (
	"fmt"
	"strings"

	"github.com/xonovex/platform/packages/library/shared-launch-go/pkg/credential"
)

// program is the libsecret command line tool.
const program = "secret-tool"

// Host is the port the store runs the tool through.
type Host interface {
	LookPath(file string) (string, error)
	Capture(name string, args []string, dir string) (string, error)
}

// Store reads a secret from a secret service.
type Store struct{ host Host }

// New returns the store over host.
func New(host Host) Store { return Store{host: host} }

// Available reports whether secret-tool is installed.
func (s Store) Available() bool {
	_, err := s.host.LookPath(program)
	return err == nil
}

// Lookup reads the secret whose "service" attribute is the first part of locator
// and whose "account" attribute is the second, when there is one.
func (s Store) Lookup(locator string) (credential.Secret, error) {
	if err := credential.ValidateLocator(locator); err != nil {
		return credential.Secret{}, err
	}
	service, account, hasAccount := strings.Cut(locator, "/")
	args := []string{"lookup", "service", service}
	if hasAccount {
		if account == "" {
			return credential.Secret{}, fmt.Errorf("secret-service locator %q has an empty account", locator)
		}
		args = append(args, "account", account)
	}
	value, err := s.host.Capture(program, args, "")
	if err != nil {
		return credential.Secret{}, fmt.Errorf("%w: %w", credential.ErrNotFound, err)
	}
	if value == "" {
		return credential.Secret{}, credential.ErrNotFound
	}
	return credential.NewSecret(value), nil
}
