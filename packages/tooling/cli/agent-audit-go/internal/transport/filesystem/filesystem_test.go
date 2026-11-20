package filesystem_test

import (
	"context"
	"errors"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/fs"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/fs/memfs"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/sources"
	"github.com/xonovex/platform/packages/tooling/cli/agent-audit-go/internal/observe"
	"github.com/xonovex/platform/packages/tooling/cli/agent-audit-go/internal/transport/filesystem"
)

func observeOne(t *testing.T, fsys fs.FileSystem, locator string) observe.Document {
	t.Helper()
	got, err := filesystem.Transport{FS: fsys}.Observe(context.Background(), observe.Request{
		Source: sources.Source{Kind: sources.Filesystem, Path: locator}, Dir: "cat",
	})
	if err != nil || got.Transport != filesystem.Name || len(got.Documents) != 1 {
		t.Fatalf("Observe = %+v, %v", got, err)
	}
	return got.Documents[0]
}

func TestObserve(t *testing.T) {
	fsys := memfs.New(map[string]string{"cat/docs/a.md": "# A\n"})

	doc := observeOne(t, fsys, "docs/a.md")
	if doc.Status != observe.OK || doc.SHA256 != observe.Hash([]byte("# A\n")) || string(doc.Body) != "# A\n" || doc.Locator != "docs/a.md" {
		t.Errorf("file = %+v", doc)
	}
	if doc := observeOne(t, fsys, "docs"); doc.Status != observe.OK || doc.SHA256 != "" || doc.Detail != "directory" {
		t.Errorf("directory = %+v", doc)
	}
	if doc := observeOne(t, fsys, "nope.md"); doc.Status != observe.Missing {
		t.Errorf("missing = %+v", doc)
	}
}

type broken struct {
	fs.FileSystem
	stat, read error
}

func (b broken) Stat(p string) (fs.Info, error) {
	if b.stat != nil {
		return fs.Info{}, b.stat
	}
	return b.FileSystem.Stat(p)
}

func (b broken) ReadFile(p string) ([]byte, error) {
	if b.read != nil {
		return nil, b.read
	}
	return b.FileSystem.ReadFile(p)
}

func TestObserveReportsAFaultAsAFailedDocument(t *testing.T) {
	boom := errors.New("boom")
	base := memfs.New(map[string]string{"cat/a.md": "x"})
	if doc := observeOne(t, broken{FileSystem: base, stat: boom}, "a.md"); doc.Status != observe.Failed || doc.Detail != "boom" {
		t.Errorf("stat fault = %+v", doc)
	}
	if doc := observeOne(t, broken{FileSystem: base, read: boom}, "a.md"); doc.Status != observe.Failed || doc.Detail != "boom" {
		t.Errorf("read fault = %+v", doc)
	}
}

func TestTransportFacts(t *testing.T) {
	tr := filesystem.Transport{}
	if tr.Name() != "filesystem" || tr.Remote() {
		t.Error("filesystem is a local transport")
	}
	if !tr.Applies(sources.Source{Path: "x"}) || tr.Applies(sources.Source{URLs: []string{"https://e.test"}}) {
		t.Error("Applies must follow the path")
	}
}
