package validation

import (
	"strings"
	"testing"
)

func TestParseRepositoryURLAcceptsPublicRepositories(t *testing.T) {
	tests := []struct {
		raw         string
		wantHost    string
		wantPort    int32
		wantDisplay string
	}{
		{raw: "https://github.com/example/repo.git", wantHost: "github.com", wantPort: 443, wantDisplay: "github.com/example/repo.git"},
		{raw: "http://gitlab.com:8080/org/project.git", wantHost: "gitlab.com", wantPort: 8080, wantDisplay: "gitlab.com/org/project.git"},
		{raw: "git@github.com:example/repo.git", wantHost: "github.com", wantPort: 22, wantDisplay: "github.com/example/repo.git"},
		{raw: "https://8.8.8.8/example/repo", wantHost: "8.8.8.8", wantPort: 443, wantDisplay: "8.8.8.8/example/repo"},
	}
	for _, test := range tests {
		t.Run(test.raw, func(t *testing.T) {
			repository, err := ParseRepositoryURL(test.raw)

			if err != nil {
				t.Fatalf("ParseRepositoryURL() error = %v", err)
			}
			if repository.Host != test.wantHost || repository.Port != test.wantPort || repository.Display() != test.wantDisplay {
				t.Fatalf("ParseRepositoryURL() = %+v, display %q", repository, repository.Display())
			}
		})
	}
}

func TestValidateRepositoryURLRejectsCredentialAndInternalTargets(t *testing.T) {
	invalid := []string{
		"",
		"https://token@example.com/repo.git",
		"https://example.com/repo.git?token=secret",
		"https://example.com/repo.git#fragment",
		"https://localhost/repo.git",
		"https://git.internal/repo.git",
		"https://127.0.0.1/repo.git",
		"https://10.0.0.1/repo.git",
		"https://169.254.169.254/latest/meta-data",
		"https://EXAMPLE.com/repo.git",
		"git@localhost:example/repo.git",
		"git@github.com:example/repo",
		"ftp://example.com/repo.git",
		"https://example.com/repo.git; rm -rf /",
	}
	for _, raw := range invalid {
		t.Run(raw, func(t *testing.T) {
			if err := ValidateRepositoryURL(raw); err == nil {
				t.Fatalf("ValidateRepositoryURL(%q) error = nil, want error", raw)
			}
		})
	}
}

func TestValidateRepositoryURLDoesNotRepeatCredentialBearingInput(t *testing.T) {
	raw := "https://secret-token@example.com/repo.git"

	err := ValidateRepositoryURL(raw)

	if err == nil {
		t.Fatal("ValidateRepositoryURL() error = nil, want error")
	}
	if errorText := err.Error(); errorText == raw || strings.Contains(errorText, "secret-token") {
		t.Fatalf("ValidateRepositoryURL() error = %q, want credential-free message", errorText)
	}
}

func TestValidateBranchAcceptsGitRefs(t *testing.T) {
	for _, branch := range []string{"", "main", "develop", "feature/my-feature", "release-1.0", "v1.2.3", "refs/heads/main"} {
		if err := ValidateBranch(branch); err != nil {
			t.Errorf("ValidateBranch(%q) error = %v", branch, err)
		}
	}
}

func TestValidateBranchRejectsInvalidGitRefs(t *testing.T) {
	invalid := []string{
		"../main",
		"main..next",
		"/main",
		"main/",
		"feature//name",
		".hidden/main",
		"feature/main.lock",
		"main.",
		"-main",
		"main@{1}",
		"main; rm -rf /",
		"branch name",
	}
	for _, branch := range invalid {
		if err := ValidateBranch(branch); err == nil {
			t.Errorf("ValidateBranch(%q) error = nil, want error", branch)
		}
	}
}

func TestValidateCommit(t *testing.T) {
	for _, commit := range []string{"", "abc1234", "deadbeef", "abc1234567890abc1234567890abc1234567890a"} {
		if err := ValidateCommit(commit); err != nil {
			t.Errorf("ValidateCommit(%q) error = %v", commit, err)
		}
	}
	for _, commit := range []string{"abc123", "not-a-sha", "abc1234; rm -rf /", "ZZZZZZZZZZ", "abc1234567890abc1234567890abc1234567890ab"} {
		if err := ValidateCommit(commit); err == nil {
			t.Errorf("ValidateCommit(%q) error = nil, want error", commit)
		}
	}
}
