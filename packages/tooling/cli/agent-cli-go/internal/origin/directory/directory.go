// Package directory is the origin=directory leaf: a source that already is a
// directory on this machine.
package directory

import (
	"fmt"
	"io/fs"
	"path"

	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/origin/shared"
)

// Host is the machine a directory origin consults.
type Host interface {
	Getwd() (string, error)
	Stat(path string) (fs.FileInfo, error)
}

// Origin serves a source from a directory.
type Origin struct{ host Host }

// New returns the directory origin over host.
func New(host Host) *Origin { return &Origin{host: host} }

// Accepts reports whether spec names an existing directory.
func (o *Origin) Accepts(spec string) bool {
	dir, err := o.absolute(spec)
	if err != nil {
		return false
	}
	info, err := o.host.Stat(dir)
	return err == nil && info.IsDir()
}

// Fetch returns the directory itself. Nothing is created, so Release does
// nothing.
func (o *Origin) Fetch(spec string) (shared.Checkout, error) {
	dir, err := o.absolute(spec)
	if err != nil {
		return shared.Checkout{}, err
	}
	info, err := o.host.Stat(dir)
	if err != nil {
		return shared.Checkout{}, err
	}
	if !info.IsDir() {
		return shared.Checkout{}, fmt.Errorf("%s is not a directory", dir)
	}
	return shared.Checkout{Dir: dir, Source: dir, Release: func() {}}, nil
}

func (o *Origin) absolute(spec string) (string, error) {
	if spec == "" {
		return "", fmt.Errorf("the source is empty")
	}
	if path.IsAbs(spec) {
		return path.Clean(spec), nil
	}
	cwd, err := o.host.Getwd()
	if err != nil {
		return "", fmt.Errorf("resolve current working directory: %w", err)
	}
	return path.Join(cwd, spec), nil
}
