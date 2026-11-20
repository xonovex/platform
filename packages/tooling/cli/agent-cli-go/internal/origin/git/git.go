// Package git is the origin=git leaf: a source that is a git repository, cloned
// shallowly into a temporary directory for one command.
package git

import (
	"errors"
	"fmt"
	"io/fs"
	"path"
	"regexp"
	"strings"

	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/origin/shared"
)

// Host is the machine a git origin consults.
type Host interface {
	MkdirTemp(pattern string) (string, error)
	RemoveAll(path string) error
	Stat(path string) (fs.FileInfo, error)
	// Stream runs name in dir with its output on the output streams of the CLI.
	Stream(name string, args []string, dir string) error
	// Capture runs name in dir and returns its trimmed standard output.
	Capture(name string, args []string, dir string) (string, error)
}

// Origin clones a repository.
type Origin struct{ host Host }

// New returns the git origin over host.
func New(host Host) *Origin { return &Origin{host: host} }

var (
	schemePattern = regexp.MustCompile(`^(https|ssh|file)://[^\s#]+$`)
	scpPattern    = regexp.MustCompile(`^[A-Za-z0-9._-]+@[A-Za-z0-9.-]+:[^\s#:][^\s#]*$`)
	refPattern    = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._/-]*$`)
)

// Accepts reports whether spec is a repository URL, with an optional #ref, that
// names an https, ssh or file transport. A plain http URL is refused, and so is
// any transport that runs a program.
func (o *Origin) Accepts(spec string) bool {
	_, _, err := split(spec)
	return err == nil
}

// Fetch clones the repository into a new temporary directory and returns the
// checkout with the commit it resolved to. Release removes the directory.
func (o *Origin) Fetch(spec string) (shared.Checkout, error) {
	url, ref, err := split(spec)
	if err != nil {
		return shared.Checkout{}, err
	}
	tmp, err := o.host.MkdirTemp("agent-cli-origin-*")
	if err != nil {
		return shared.Checkout{}, fmt.Errorf("create a directory for the clone: %w", err)
	}
	release := func() { _ = o.host.RemoveAll(tmp) }
	dir := path.Join(tmp, "checkout")
	args := []string{"-c", "protocol.ext.allow=never", "clone", "--quiet", "--depth", "1", "--no-tags"}
	if ref != "" {
		args = append(args, "--branch", ref)
	}
	args = append(args, "--", url, dir)
	if err := o.host.Stream("git", args, tmp); err != nil {
		release()
		return shared.Checkout{}, fmt.Errorf("clone %s: %w", url, err)
	}
	commit, err := o.host.Capture("git", []string{"rev-parse", "HEAD"}, dir)
	if err != nil {
		release()
		return shared.Checkout{}, err
	}
	source := url
	if ref != "" {
		source += "#" + ref
	}
	return shared.Checkout{Dir: dir, Source: source, Resolved: commit, Release: release}, nil
}

// split separates the URL of spec from its reference and validates both.
func split(spec string) (url, ref string, err error) {
	url, ref, _ = strings.Cut(spec, "#")
	switch {
	case strings.HasPrefix(url, "-"):
		return "", "", errors.New("a repository URL may not start with a dash")
	case !schemePattern.MatchString(url) && !scpPattern.MatchString(url):
		return "", "", fmt.Errorf("%q is not a repository URL", url)
	case strings.Contains(spec, "#") && !refPattern.MatchString(ref):
		return "", "", fmt.Errorf("%q is not a branch or tag name", ref)
	}
	return url, ref, nil
}
