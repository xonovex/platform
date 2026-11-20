package egress

import (
	"errors"
	"testing"
)

func TestParseRule(t *testing.T) {
	good := map[string]Rule{
		"api.example.com":        {Host: "api.example.com", Port: 443},
		"API.Example.com.":       {Host: "api.example.com", Port: 443},
		"api.example.com:8443":   {Host: "api.example.com", Port: 8443},
		"*.example.com":          {Host: "example.com", Subdomains: true, Port: 443},
		"*.example.com:80":       {Host: "example.com", Subdomains: true, Port: 80},
		"192.0.2.10:9000":        {Host: "192.0.2.10", Port: 9000},
		" registry.example.org ": {Host: "registry.example.org", Port: 443},
	}
	for text, want := range good {
		got, err := ParseRule(text)
		if err != nil || got != want {
			t.Errorf("ParseRule(%q) = %+v, %v; want %+v", text, got, err, want)
		}
	}
	bad := []string{"", "  ", "*", "*.com", "*.192.0.2.10", "https://example.com", "example.com/path", "user@example.com", "a b", "example.com:0", "example.com:70000", "example.com:http", "ex*mple.com", "*."}
	for _, text := range bad {
		if _, err := ParseRule(text); err == nil {
			t.Errorf("ParseRule(%q) succeeded, want an error", text)
		}
	}
}

func TestRuleString(t *testing.T) {
	if got := (Rule{Host: "example.com", Subdomains: true, Port: 443}).String(); got != "*.example.com:443" {
		t.Errorf("String() = %q", got)
	}
	if got := (Rule{Host: "example.com", Port: 80}).String(); got != "example.com:80" {
		t.Errorf("String() = %q", got)
	}
}

func TestNewAllowlist(t *testing.T) {
	if _, err := NewAllowlist(nil); !errors.Is(err, ErrEmpty) {
		t.Errorf("empty allowlist error = %v, want %v", err, ErrEmpty)
	}
	if _, err := NewAllowlist([]string{"example.com", "*"}); err == nil {
		t.Error("an invalid rule must fail the allowlist")
	}
	list, err := NewAllowlist([]string{"b.example.com", "a.example.com", "b.example.com"})
	if err != nil {
		t.Fatal(err)
	}
	rules := list.Rules()
	if len(rules) != 2 || rules[0].Host != "a.example.com" || rules[1].Host != "b.example.com" {
		t.Errorf("Rules() = %+v, want two sorted rules without the duplicate", rules)
	}
	rules[0].Host = "changed"
	if list.Rules()[0].Host != "a.example.com" {
		t.Error("Rules() must return a copy")
	}
}

func TestAllows(t *testing.T) {
	list, err := NewAllowlist([]string{"api.example.com", "*.cdn.example.org", "files.example.net:8080"})
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		host string
		port int
		want bool
	}{
		{"api.example.com", 443, true},
		{"API.EXAMPLE.COM.", 443, true},
		{"api.example.com", 80, false},
		{"other.example.com", 443, false},
		{"x.api.example.com", 443, false},
		{"a.cdn.example.org", 443, true},
		{"a.b.cdn.example.org", 443, true},
		{"cdn.example.org", 443, false},
		{"evilcdn.example.org", 443, false},
		{"files.example.net", 8080, true},
		{"files.example.net", 443, false},
	}
	for _, c := range cases {
		if got := list.Allows(c.host, c.port); got != c.want {
			t.Errorf("Allows(%q, %d) = %t, want %t", c.host, c.port, got, c.want)
		}
	}
	var none Allowlist
	if none.Allows("api.example.com", 443) {
		t.Error("the zero allowlist must deny everything")
	}
}
