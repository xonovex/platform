package shell

import "testing"

func TestQuote(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"plain string", "hello", "'hello'"},
		{"string with spaces", "hello world", "'hello world'"},
		{"single quotes", "it's", "'it'\\''s'"},
		{"double quotes", `say "hi"`, `'say "hi"'`},
		{"semicolons", "a;b", "'a;b'"},
		{"empty string", "", "''"},
		{"path with slashes", "/usr/local/bin", "'/usr/local/bin'"},
		{"multiple single quotes", "a'b'c", "'a'\\''b'\\''c'"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Quote(tt.input); got != tt.expected {
				t.Errorf("Quote(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestContainsMetachars(t *testing.T) {
	clean := []struct {
		name  string
		input string
	}{
		{"plain text", "hello"},
		{"path", "/usr/local/bin"},
		{"url-safe", "https://example.com/repo.git"},
		{"digits", "1234567890"},
		{"spaces and tabs", "hello world\there"},
		{"empty", ""},
	}

	for _, tt := range clean {
		t.Run("clean/"+tt.name, func(t *testing.T) {
			if ContainsMetachars(tt.input) {
				t.Errorf("ContainsMetachars(%q) = true, want false", tt.input)
			}
		})
	}

	dangerous := []struct {
		name  string
		input string
	}{
		{"semicolon", "cmd;evil"},
		{"pipe", "cmd|evil"},
		{"ampersand", "cmd&evil"},
		{"dollar", "cmd$evil"},
		{"backtick", "cmd`evil`"},
		{"backslash", "cmd\\evil"},
		{"double quote", `cmd"evil`},
		{"single quote", "cmd'evil"},
		{"less than", "cmd<evil"},
		{"greater than", "cmd>evil"},
		{"open paren", "cmd(evil"},
		{"close paren", "cmd)evil"},
		{"open brace", "cmd{evil"},
		{"close brace", "cmd}evil"},
		{"exclamation", "cmd!evil"},
		{"hash", "cmd#evil"},
		{"tilde", "cmd~evil"},
		{"newline", "cmd\nevil"},
		{"carriage return", "cmd\revil"},
	}

	for _, tt := range dangerous {
		t.Run("dangerous/"+tt.name, func(t *testing.T) {
			if !ContainsMetachars(tt.input) {
				t.Errorf("ContainsMetachars(%q) = false, want true", tt.input)
			}
		})
	}
}
