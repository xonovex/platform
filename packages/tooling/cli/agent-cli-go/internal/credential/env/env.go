// Package env is the credential store env: a credential held in an environment
// variable of the CLI.
package env

import (
	"fmt"
	"regexp"

	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/credential/shared"
)

// Host is the port the store reads the environment through.
type Host interface {
	Getenv(key string) string
}

var variable = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// Store reads a credential from an environment variable.
type Store struct{ host Host }

// New returns the store over host.
func New(host Host) Store { return Store{host: host} }

// Available is always true: every process has an environment.
func (Store) Available() bool { return true }

// Lookup returns the value of the variable named by locator.
func (s Store) Lookup(locator string) (shared.Secret, error) {
	if !variable.MatchString(locator) {
		return shared.Secret{}, fmt.Errorf("%q is not an environment variable name", locator)
	}
	value := s.host.Getenv(locator)
	if value == "" {
		return shared.Secret{}, shared.ErrNotFound
	}
	return shared.NewSecret(value), nil
}
