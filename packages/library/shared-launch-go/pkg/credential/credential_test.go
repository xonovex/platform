package credential

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestSecretNeverPrints(t *testing.T) {
	secret := NewSecret("hunter2")
	for _, verb := range []string{"%s", "%v", "%+v", "%#v", "%q"} {
		if out := fmt.Sprintf(verb, secret); strings.Contains(out, "hunter2") {
			t.Errorf("%s printed the value: %q", verb, out)
		}
	}
	if fmt.Sprint(map[string]Secret{"K": secret}) != "map[K:[redacted]]" {
		t.Errorf("a map of secrets printed %q", fmt.Sprint(map[string]Secret{"K": secret}))
	}
	if secret.Reveal() != "hunter2" || secret.GoString() != Redacted {
		t.Error("Reveal must return the value and GoString the placeholder")
	}
}

func TestParseBinding(t *testing.T) {
	got, err := ParseBinding("API_KEY=keychain:service/account")
	if err != nil || got != (Binding{Name: "API_KEY", Scheme: "keychain", Locator: "service/account"}) {
		t.Fatalf("ParseBinding = %+v, %v", got, err)
	}
	if got.String() != "API_KEY=keychain:service/account" {
		t.Errorf("String() = %q", got.String())
	}
	for _, bad := range []string{"", "API_KEY", "API_KEY=env", "API_KEY=:x", "API_KEY=env:", "1BAD=env:X", "A-B=env:X", "=env:X"} {
		if _, err := ParseBinding(bad); err == nil {
			t.Errorf("ParseBinding(%q) succeeded", bad)
		}
	}
}

func TestParseBindings(t *testing.T) {
	bindings, err := ParseBindings([]string{"A=env:X", "B=env:Y"})
	if err != nil || len(bindings) != 2 {
		t.Fatalf("ParseBindings = %v, %v", bindings, err)
	}
	if _, err := ParseBindings([]string{"A=env:X", "A=env:Y"}); err == nil {
		t.Error("a variable bound twice must fail")
	}
	if _, err := ParseBindings([]string{"nope"}); err == nil {
		t.Error("an invalid binding must fail")
	}
	if got, err := ParseBindings(nil); err != nil || got != nil {
		t.Errorf("ParseBindings(nil) = %v, %v", got, err)
	}
}

func TestValidateLocator(t *testing.T) {
	for _, good := range []string{"a", "service/account", "a b"} {
		if err := ValidateLocator(good); err != nil {
			t.Errorf("ValidateLocator(%q) = %v", good, err)
		}
	}
	for _, bad := range []string{"", "-flag", "a\nb", "a\x00b", "a\x7fb"} {
		if err := ValidateLocator(bad); err == nil {
			t.Errorf("ValidateLocator(%q) succeeded", bad)
		}
	}
}

type fakeStore struct {
	available bool
	value     string
	err       error
}

func (f fakeStore) Available() bool { return f.available }
func (f fakeStore) Lookup(string) (Secret, error) {
	if f.err != nil {
		return Secret{}, f.err
	}
	return NewSecret(f.value), nil
}

func TestRegistryAndResolve(t *testing.T) {
	registry := NewRegistry().
		Register("good", func() Store { return fakeStore{available: true, value: "v"} }).
		Register("off", func() Store { return fakeStore{} }).
		Register("empty", func() Store { return fakeStore{available: true, err: ErrNotFound} })
	if got := strings.Join(registry.Schemes(), ","); got != "empty,good,off" {
		t.Errorf("Schemes() = %q", got)
	}
	secrets, err := Resolve(registry, []Binding{{Name: "A", Scheme: "good", Locator: "x"}})
	if err != nil || secrets["A"].Reveal() != "v" {
		t.Fatalf("Resolve = %v, %v", secrets, err)
	}
	plain := Reveal(secrets)
	if plain["A"] != "v" || Reveal(nil) != nil {
		t.Errorf("Reveal = %v", plain)
	}
	cases := map[string]error{
		"unknown": ErrUnknownScheme,
		"off":     ErrUnavailable,
		"empty":   ErrNotFound,
	}
	for scheme, want := range cases {
		_, err := Resolve(registry, []Binding{{Name: "A", Scheme: scheme, Locator: "x"}})
		if !errors.Is(err, want) {
			t.Errorf("Resolve(%s) = %v, want %v", scheme, err, want)
		}
	}
}
