// Package secretservice is the credential store secret-service: a secret in a
// Linux secret service such as GNOME Keyring or KWallet, read through secret-tool.
package secretservice

import (
	"fmt"
	"strings"

	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/credential/shared"
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
func (s Store) Lookup(locator string) (shared.Secret, error) {
	if err := shared.ValidateLocator(locator); err != nil {
		return shared.Secret{}, err
	}
	service, account, hasAccount := strings.Cut(locator, "/")
	args := []string{"lookup", "service", service}
	if hasAccount {
		if account == "" {
			return shared.Secret{}, fmt.Errorf("secret-service locator %q has an empty account", locator)
		}
		args = append(args, "account", account)
	}
	value, err := s.host.Capture(program, args, "")
	if err != nil {
		return shared.Secret{}, fmt.Errorf("%w: %w", shared.ErrNotFound, err)
	}
	if value == "" {
		return shared.Secret{}, shared.ErrNotFound
	}
	return shared.NewSecret(value), nil
}
