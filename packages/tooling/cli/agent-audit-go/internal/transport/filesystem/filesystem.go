// Package filesystem is the transport that reads a source from a path on disk.
package filesystem

import (
	"context"
	"errors"
	"path"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/fs"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/sources"
	"github.com/xonovex/platform/packages/tooling/cli/agent-audit-go/internal/observe"
)

// Name is the name of the transport in reports.
const Name = "filesystem"

// Transport reads the file a source names with its path.
type Transport struct {
	FS fs.FileSystem
}

// Name returns the name of the transport.
func (Transport) Name() string { return Name }

// Remote returns false: reading a path uses no network.
func (Transport) Remote() bool { return false }

// Applies reports whether the source names a path.
func (Transport) Applies(source sources.Source) bool { return source.Path != "" }

// Observe reads the file at the path, resolved against the catalog directory. A
// directory is present but has no content to compare.
func (t Transport) Observe(_ context.Context, request observe.Request) (observe.Observation, error) {
	locator := request.Source.Path
	document := observe.Document{Locator: locator}
	full := path.Join(request.Dir, locator)

	info, err := t.FS.Stat(full)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		document.Status = observe.Missing
		document.Detail = "no such path"
	case err != nil:
		document.Status = observe.Failed
		document.Detail = err.Error()
	case info.IsDir:
		document.Status = observe.OK
		document.Detail = "directory"
	default:
		t.read(full, &document)
	}
	return observe.Observation{Transport: Name, Documents: []observe.Document{document}}, nil
}

func (t Transport) read(full string, document *observe.Document) {
	body, err := t.FS.ReadFile(full)
	if err != nil {
		document.Status = observe.Failed
		document.Detail = err.Error()
		return
	}
	document.Status = observe.OK
	document.SHA256 = observe.Hash(body)
	document.Body = body
}
