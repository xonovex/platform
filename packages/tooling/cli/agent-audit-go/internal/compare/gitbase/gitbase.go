// Package gitbase is the adapter that reads the base of a comparison from a git
// ref, through the process port.
package gitbase

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/findings"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/fs"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/sources"
	"github.com/xonovex/platform/packages/tooling/cli/agent-audit-go/internal/gitfs"
	"github.com/xonovex/platform/packages/tooling/cli/agent-audit-go/internal/process"
)

// ErrUnknownRef reports a base ref that the repository does not have.
var ErrUnknownRef = errors.New("unknown git ref")

// Base reads the catalog and the changed files at Ref.
type Base struct {
	Runner process.Runner
	Ref    string
	// LoaderFor returns the catalog loader that reads through the file system.
	LoaderFor func(fs.FileSystem) (sources.Loader, error)
}

func (b Base) git(dir string, args ...string) (process.Result, error) {
	return b.Runner.Run(context.Background(), process.Command{Dir: dir, Name: "git", Args: args})
}

// verify fails when the ref is not a commit of the repository, so a mistyped
// ref cannot read as an empty catalog.
func (b Base) verify(dir string) error {
	result, err := b.git(dir, "rev-parse", "--verify", "--quiet", b.Ref+"^{commit}")
	if err != nil {
		return err
	}
	if !result.Succeeded() {
		return fmt.Errorf("%w %q", ErrUnknownRef, b.Ref)
	}
	return nil
}

// Catalog loads the catalog in dir as it was at the ref. A catalog that was not
// there yet is empty.
func (b Base) Catalog(dir string) (sources.Catalog, error) {
	if err := b.verify(dir); err != nil {
		return sources.Catalog{}, err
	}
	loader, err := b.LoaderFor(gitfs.FS{Runner: b.Runner, Dir: dir, Ref: b.Ref})
	if err != nil {
		return sources.Catalog{}, err
	}
	catalog, loaded, err := loader.Load(".")
	if err != nil {
		return sources.Catalog{}, err
	}
	for _, f := range loaded {
		if f.Code == sources.CodeMissing && f.Level == findings.Fail {
			return sources.Catalog{}, nil
		}
	}
	return catalog, nil
}

// ChangedFiles lists the files below dir that differ between the ref and the
// working tree.
func (b Base) ChangedFiles(dir string) ([]string, error) {
	if err := b.verify(dir); err != nil {
		return nil, err
	}
	result, err := b.git(dir, "diff", "--name-only", "--relative", b.Ref, "--", ".")
	if err != nil {
		return nil, err
	}
	if !result.Succeeded() {
		return nil, fmt.Errorf("git diff against %q failed with status %d", b.Ref, result.ExitCode)
	}
	var files []string
	for _, line := range strings.Split(string(result.Stdout), "\n") {
		if line = strings.TrimSpace(line); line != "" {
			files = append(files, line)
		}
	}
	return files, nil
}
