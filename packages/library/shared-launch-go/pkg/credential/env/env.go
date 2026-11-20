// Package env is the credential store env: a credential held in an environment
// variable of this process.
package env

import (
	"fmt"
	"regexp"

	"github.com/xonovex/platform/packages/library/shared-launch-go/pkg/credential"
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
func (s Store) Lookup(locator string) (credential.Secret, error) {
	if !variable.MatchString(locator) {
		return credential.Secret{}, fmt.Errorf("%q is not an environment variable name", locator)
	}
	value := s.host.Getenv(locator)
	if value == "" {
		return credential.Secret{}, credential.ErrNotFound
	}
	return credential.NewSecret(value), nil
}
