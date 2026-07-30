package validation

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
)

// RepositoryURL is the validated network location of a Git repository.
type RepositoryURL struct {
	Scheme string
	Host   string
	Path   string
	Port   int32
}

// Display returns a credential-free repository description for logs and events.
func (repository RepositoryURL) Display() string {
	host := repository.Host
	if net.ParseIP(host) != nil && strings.Contains(host, ":") {
		host = "[" + host + "]"
	}
	return host + repository.Path
}

// ParseRepositoryURL validates and decomposes an HTTP(S) or git+SSH URL.
func ParseRepositoryURL(raw string) (RepositoryURL, error) {
	if raw == "" {
		return RepositoryURL{}, fmt.Errorf("repository URL is required")
	}
	if strings.HasPrefix(raw, "git@") {
		return parseSSHRepositoryURL(raw)
	}

	parsed, err := url.Parse(raw)
	if err != nil {
		return RepositoryURL{}, fmt.Errorf("repository URL is malformed")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return RepositoryURL{}, fmt.Errorf("repository URL must use http, https, or git+SSH")
	}
	if parsed.User != nil {
		return RepositoryURL{}, fmt.Errorf("repository URL must not contain credentials; use credentialsSecretRef")
	}
	if parsed.Hostname() == "" {
		return RepositoryURL{}, fmt.Errorf("repository URL host is required")
	}
	if parsed.RawQuery != "" || parsed.ForceQuery || parsed.Fragment != "" {
		return RepositoryURL{}, fmt.Errorf("repository URL must not contain a query or fragment")
	}
	if err := validateRepositoryHost(parsed.Hostname()); err != nil {
		return RepositoryURL{}, err
	}
	if err := validateRepositoryPath(parsed.Path, false); err != nil {
		return RepositoryURL{}, err
	}

	port, err := repositoryPort(parsed.Scheme, parsed.Port())
	if err != nil {
		return RepositoryURL{}, err
	}
	return RepositoryURL{
		Scheme: parsed.Scheme,
		Host:   strings.ToLower(parsed.Hostname()),
		Path:   parsed.EscapedPath(),
		Port:   port,
	}, nil
}

func parseSSHRepositoryURL(raw string) (RepositoryURL, error) {
	host, repositoryPath, found := strings.Cut(strings.TrimPrefix(raw, "git@"), ":")
	if !found || host == "" || repositoryPath == "" || strings.Contains(repositoryPath, ":") {
		return RepositoryURL{}, fmt.Errorf("git+SSH repository URL must use git@host:owner/repository.git")
	}
	if err := validateRepositoryHost(host); err != nil {
		return RepositoryURL{}, err
	}
	if err := validateRepositoryPath(repositoryPath, true); err != nil {
		return RepositoryURL{}, err
	}
	return RepositoryURL{
		Scheme: "ssh",
		Host:   strings.ToLower(host),
		Path:   "/" + repositoryPath,
		Port:   22,
	}, nil
}

func repositoryPort(scheme, rawPort string) (int32, error) {
	if rawPort == "" {
		if scheme == "http" {
			return 80, nil
		}
		return 443, nil
	}
	port, err := strconv.ParseInt(rawPort, 10, 32)
	if err != nil || port < 1 || port > 65535 {
		return 0, fmt.Errorf("repository URL port is invalid")
	}
	return int32(port), nil
}

func validateRepositoryHost(host string) error {
	normalized := strings.ToLower(strings.TrimSuffix(host, "."))
	if normalized == "" || normalized != host {
		return fmt.Errorf("repository URL host must use canonical lowercase DNS syntax")
	}
	if address := net.ParseIP(normalized); address != nil {
		if !address.IsGlobalUnicast() || address.IsPrivate() || address.IsLoopback() || address.IsLinkLocalUnicast() {
			return fmt.Errorf("repository URL host must be publicly routable")
		}
		return nil
	}
	if !strings.Contains(normalized, ".") || strings.HasSuffix(normalized, ".localhost") ||
		strings.HasSuffix(normalized, ".local") || strings.HasSuffix(normalized, ".internal") ||
		strings.HasSuffix(normalized, ".home.arpa") {
		return fmt.Errorf("repository URL host must be a public DNS name")
	}
	if len(normalized) > 253 {
		return fmt.Errorf("repository URL host is too long")
	}
	for _, label := range strings.Split(normalized, ".") {
		if !isDNSLabel(label) {
			return fmt.Errorf("repository URL host must use canonical DNS syntax")
		}
	}
	return nil
}

func isDNSLabel(label string) bool {
	if label == "" || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
		return false
	}
	for index := range len(label) {
		character := label[index]
		if character != '-' && (character < 'a' || character > 'z') && (character < '0' || character > '9') {
			return false
		}
	}
	return true
}

func validateRepositoryPath(repositoryPath string, requireGitSuffix bool) error {
	trimmed := strings.Trim(repositoryPath, "/")
	if trimmed == "" || strings.ContainsAny(repositoryPath, "\x00\r\n\t \\;|&$`\"'<>(){}") {
		return fmt.Errorf("repository URL path is invalid")
	}
	for _, component := range strings.Split(trimmed, "/") {
		if component == "" || component == "." || component == ".." {
			return fmt.Errorf("repository URL path must be canonical")
		}
	}
	if requireGitSuffix && !strings.HasSuffix(trimmed, ".git") {
		return fmt.Errorf("git+SSH repository URL path must end in .git")
	}
	return nil
}

func ValidateRepositoryURL(raw string) error {
	_, err := ParseRepositoryURL(raw)
	return err
}

func ValidateBranch(branch string) error {
	if branch == "" {
		return nil
	}
	if strings.HasPrefix(branch, "-") || strings.HasPrefix(branch, "/") || strings.HasSuffix(branch, "/") ||
		strings.HasSuffix(branch, ".") || strings.Contains(branch, "//") || strings.Contains(branch, "..") ||
		strings.Contains(branch, "@{") || branch == "@" || hasInvalidRefCharacter(branch) {
		return fmt.Errorf("branch is not a valid Git ref")
	}
	for _, component := range strings.Split(branch, "/") {
		if strings.HasPrefix(component, ".") || strings.HasSuffix(component, ".lock") {
			return fmt.Errorf("branch is not a valid Git ref")
		}
	}
	return nil
}

func hasInvalidRefCharacter(ref string) bool {
	for _, character := range ref {
		if character != '.' && character != '-' && character != '_' && character != '/' &&
			(character < 'a' || character > 'z') && (character < 'A' || character > 'Z') &&
			(character < '0' || character > '9') {
			return true
		}
	}
	return false
}

func ValidateCommit(commit string) error {
	if commit == "" {
		return nil
	}
	if len(commit) < 7 || len(commit) > 40 {
		return fmt.Errorf("commit must be a 7-40 character hex SHA")
	}
	for _, character := range commit {
		if character < '0' || character > '9' {
			lower := character | 0x20
			if lower < 'a' || lower > 'f' {
				return fmt.Errorf("commit must be a 7-40 character hex SHA")
			}
		}
	}
	return nil
}
