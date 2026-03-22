package validator

import (
	"fmt"
	"regexp"
)

// repoURLPattern allows http/https and git+ssh URLs.
// The https character class includes RFC 3986 characters minus shell metacharacters (;|&$`).
var repoURLPattern = regexp.MustCompile(
	`^(https?://[a-zA-Z0-9._~:/?#@!'+,=%\[\]()-]+|git@[a-zA-Z0-9._-]+:[a-zA-Z0-9/_.-]+\.git)$`,
)

// refPattern allows letters, digits, ., -, _, / — no shell metacharacters
var refPattern = regexp.MustCompile(`^[a-zA-Z0-9._/\-]+$`)

// commitPattern allows hex SHA, 7-40 chars
var commitPattern = regexp.MustCompile(`^[0-9a-fA-F]{7,40}$`)

func ValidateRepositoryURL(url string) error {
	if url == "" {
		return fmt.Errorf("repository URL is required")
	}
	if !repoURLPattern.MatchString(url) {
		return fmt.Errorf("repository URL %q contains invalid characters or unsupported scheme", url)
	}
	return nil
}

func ValidateBranch(branch string) error {
	if branch == "" {
		return nil
	}
	if !refPattern.MatchString(branch) {
		return fmt.Errorf("branch %q contains invalid characters", branch)
	}
	return nil
}

func ValidateCommit(commit string) error {
	if commit == "" {
		return nil
	}
	if !commitPattern.MatchString(commit) {
		return fmt.Errorf("commit %q must be a 7-40 character hex SHA", commit)
	}
	return nil
}
