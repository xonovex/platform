// Package credential is the credential axis core: the Store port that a credential
// store implements, the Secret value that never prints, the Binding of an
// environment variable to a stored credential, and the Registry that maps a
// scheme to a store. It names no concrete store; the per-store packages are the
// leaves.
package credential

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// Secret is a credential value. It prints as a placeholder through every fmt verb,
// so a log line, an error or a dry run cannot leak it by accident; the value
// leaves only through Reveal, which the executor calls to build the environment of
// the agent.
type Secret struct{ value string }

// NewSecret wraps a credential value.
func NewSecret(value string) Secret { return Secret{value: value} }

// Redacted is what a Secret prints as.
const Redacted = "[redacted]"

// String returns Redacted.
func (Secret) String() string { return Redacted }

// GoString returns Redacted, for the %#v verb.
func (Secret) GoString() string { return Redacted }

// Reveal returns the credential value.
func (s Secret) Reveal() string { return s.value }

var (
	// ErrNotFound reports a credential the store does not hold, or holds empty.
	ErrNotFound = errors.New("credential not found")
	// ErrUnavailable reports a store that cannot run on this host.
	ErrUnavailable = errors.New("credential store unavailable")
	// ErrUnknownScheme reports a binding that names no registered store.
	ErrUnknownScheme = errors.New("unknown credential store")
)

// Store reads one credential from a credential store.
type Store interface {
	// Available reports whether the store can run on this host.
	Available() bool
	// Lookup returns the credential at locator. It never returns the value in an
	// error.
	Lookup(locator string) (Secret, error)
}

// Factory builds a store.
type Factory func() Store

// Registry maps a scheme to a store factory. It is built once at the composition
// root and passed explicitly, so there is no global state, and a test builds a
// registry of its own.
type Registry struct {
	factories map[string]Factory
}

// NewRegistry returns an empty registry.
func NewRegistry() *Registry { return &Registry{factories: map[string]Factory{}} }

// Register adds or replaces the store for scheme; chainable.
func (r *Registry) Register(scheme string, factory Factory) *Registry {
	r.factories[scheme] = factory
	return r
}

// Schemes returns the registered schemes in sorted order.
func (r *Registry) Schemes() []string {
	schemes := make([]string, 0, len(r.factories))
	for scheme := range r.factories {
		schemes = append(schemes, scheme)
	}
	sort.Strings(schemes)
	return schemes
}

// Store builds the store for scheme.
func (r *Registry) Store(scheme string) (Store, error) {
	factory, ok := r.factories[scheme]
	if !ok {
		return nil, fmt.Errorf("%w %q; available: %s", ErrUnknownScheme, scheme, strings.Join(r.Schemes(), ", "))
	}
	return factory(), nil
}

// Binding names the environment variable that receives a credential and where
// the credential is stored.
type Binding struct {
	Name    string
	Scheme  string
	Locator string
}

// String renders the binding in the form ParseBinding reads. It holds no secret.
func (b Binding) String() string { return b.Name + "=" + b.Scheme + ":" + b.Locator }

var variableName = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// ParseBinding reads NAME=scheme:locator.
func ParseBinding(text string) (Binding, error) {
	name, source, ok := strings.Cut(text, "=")
	if !ok {
		return Binding{}, fmt.Errorf("credential %q is not NAME=scheme:locator", text)
	}
	scheme, locator, ok := strings.Cut(source, ":")
	if !ok || scheme == "" || locator == "" {
		return Binding{}, fmt.Errorf("credential %q is not NAME=scheme:locator", text)
	}
	if !variableName.MatchString(name) {
		return Binding{}, fmt.Errorf("credential %q: %q is not an environment variable name", text, name)
	}
	return Binding{Name: name, Scheme: scheme, Locator: locator}, nil
}

// ParseBindings reads each text and refuses two bindings for one variable.
func ParseBindings(texts []string) ([]Binding, error) {
	seen := map[string]bool{}
	var bindings []Binding
	for _, text := range texts {
		binding, err := ParseBinding(text)
		if err != nil {
			return nil, err
		}
		if seen[binding.Name] {
			return nil, fmt.Errorf("credential variable %s is bound twice", binding.Name)
		}
		seen[binding.Name] = true
		bindings = append(bindings, binding)
	}
	return bindings, nil
}

// ValidateLocator rejects a locator that a command line store could read as an
// option or that carries a control character.
func ValidateLocator(locator string) error {
	if locator == "" || strings.HasPrefix(locator, "-") {
		return fmt.Errorf("credential locator %q is empty or starts with a dash", locator)
	}
	for _, r := range locator {
		if r < 0x20 || r == 0x7f {
			return fmt.Errorf("credential locator %q has a control character", locator)
		}
	}
	return nil
}

// Resolve reads every binding through its store and returns the secrets by
// variable name. It fails closed: one store that is unknown, unavailable or
// missing the credential fails the whole run, so the agent never starts without a
// credential the user asked for.
func Resolve(registry *Registry, bindings []Binding) (map[string]Secret, error) {
	secrets := make(map[string]Secret, len(bindings))
	for _, binding := range bindings {
		store, err := registry.Store(binding.Scheme)
		if err != nil {
			return nil, fmt.Errorf("credential %s: %w", binding.Name, err)
		}
		if !store.Available() {
			return nil, fmt.Errorf("credential %s: %w: %s", binding.Name, ErrUnavailable, binding.Scheme)
		}
		secret, err := store.Lookup(binding.Locator)
		if err != nil {
			return nil, fmt.Errorf("credential %s from %s:%s: %w", binding.Name, binding.Scheme, binding.Locator, err)
		}
		secrets[binding.Name] = secret
	}
	return secrets, nil
}

// Reveal converts secrets to the plain map the isolators put in the environment.
func Reveal(secrets map[string]Secret) map[string]string {
	if len(secrets) == 0 {
		return nil
	}
	plain := make(map[string]string, len(secrets))
	for name, secret := range secrets {
		plain[name] = secret.Reveal()
	}
	return plain
}
