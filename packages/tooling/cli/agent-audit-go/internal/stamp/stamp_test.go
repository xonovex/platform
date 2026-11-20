package stamp_test

import (
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/fs"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/fs/memfs"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/sources"
	"github.com/xonovex/platform/packages/tooling/cli/agent-audit-go/internal/stamp"
)

var now = time.Date(2026, 11, 20, 23, 30, 0, 0, time.UTC)

const root = `{"version":1,"sources":[
  {"id":"alpha","kind":"original","title":"Alpha guide","provenance":"p","reviewedAt":"2026-01-01"},
  {"id":"beta","kind":"original","title":"Beta guide","provenance":"p","reviewedAt":"2026-11-20"}]}`

const shard = `{"version":1,"sources":[{"id":"gamma","kind":"original","title":"Gamma","provenance":"p","reviewedAt":"2026-02-02","claims":[{"id":"c","statement":"s"}]}]}`

func load(t *testing.T, fsys fs.FileSystem) sources.Catalog {
	t.Helper()
	var all sources.Catalog
	for _, file := range []string{"d/sources.json", "d/sources.d/g.json"} {
		data, err := fsys.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		doc, err := sources.Decode(data)
		if err != nil {
			t.Fatal(err)
		}
		for _, s := range doc.Sources {
			s.File = file
			all.Sources = append(all.Sources, s)
		}
	}
	return all
}

func files() map[string]string {
	return map[string]string{"d/sources.json": root, "d/sources.d/g.json": shard}
}

func TestReviewedStampsEverySourceByDefault(t *testing.T) {
	fsys := memfs.New(files())
	result, err := stamp.Reviewed(fsys, load(t, fsys), "", now)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(result.Edited, []string{"alpha", "gamma"}) || !reflect.DeepEqual(result.Files, []string{"d/sources.d/g.json", "d/sources.json"}) {
		t.Fatalf("result = %+v", result)
	}
	got := load(t, fsys)
	for _, s := range got.Sources {
		if s.ReviewedAt != "2026-11-20" {
			t.Errorf("%s: reviewedAt = %s", s.ID, s.ReviewedAt)
		}
	}
	data, _ := fsys.ReadFile("d/sources.d/g.json")
	if !strings.Contains(string(data), `"claims"`) || !strings.HasSuffix(string(data), "}\n") {
		t.Errorf("the rewrite lost content:\n%s", data)
	}
}

func TestReviewedFilterMatchesTheTitleIgnoringCase(t *testing.T) {
	fsys := memfs.New(files())
	result, err := stamp.Reviewed(fsys, load(t, fsys), "ALPHA", now)
	if err != nil || !reflect.DeepEqual(result.Edited, []string{"alpha"}) || !reflect.DeepEqual(result.Files, []string{"d/sources.json"}) {
		t.Fatalf("result = %+v, err = %v", result, err)
	}
	if data, _ := fsys.ReadFile("d/sources.d/g.json"); string(data) != shard {
		t.Error("a file with no match must not be rewritten")
	}
}

func TestReviewedLeavesACurrentCatalogAlone(t *testing.T) {
	fsys := memfs.New(map[string]string{"d/sources.json": `{"version":1,"sources":[{"id":"b","kind":"original","title":"B","provenance":"p","reviewedAt":"2026-11-20"}]}`})
	catalog := sources.Catalog{Sources: []sources.Source{{ID: "b", Title: "B", File: "d/sources.json"}}}
	result, err := stamp.Reviewed(fsys, catalog, "", now)
	if err != nil || len(result.Edited) != 0 || len(result.Files) != 0 {
		t.Fatalf("result = %+v, err = %v", result, err)
	}
}

type failing struct {
	fs.FileSystem
	write error
}

func (f failing) WriteFile(string, []byte) error { return f.write }

func TestReviewedReportsFaults(t *testing.T) {
	fsys := memfs.New(files())
	catalog := load(t, fsys)

	missing := sources.Catalog{Sources: []sources.Source{{ID: "x", Title: "X", File: "d/none.json"}}}
	if _, err := stamp.Reviewed(fsys, missing, "", now); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("read: %v", err)
	}

	bad := memfs.New(map[string]string{"d/sources.json": `{"version":1,"sources":[],"surprise":1}`})
	if _, err := stamp.Reviewed(bad, sources.Catalog{Sources: []sources.Source{{ID: "x", Title: "X", File: "d/sources.json"}}}, "", now); err == nil {
		t.Error("an undecodable file must be an error")
	}

	boom := errors.New("disk full")
	if _, err := stamp.Reviewed(failing{FileSystem: fsys, write: boom}, catalog, "", now); !errors.Is(err, boom) {
		t.Errorf("write: %v", err)
	}
}
