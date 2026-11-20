package sources_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/findings"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/fs"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/fs/memfs"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/jsonschema"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/schemas"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/sources"
)

// scripted is a Validator test double: a document that contains the key of a
// script entry gets that entry's violations.
type scripted struct {
	byMarker map[string][]jsonschema.Violation
	err      error
	asked    []string
}

func (s *scripted) Validate(schemaID string, document []byte) ([]jsonschema.Violation, error) {
	s.asked = append(s.asked, schemaID)
	if s.err != nil {
		return nil, s.err
	}
	for marker, violations := range s.byMarker {
		if strings.Contains(string(document), marker) {
			return violations, nil
		}
	}
	return nil, nil
}

func file(ids ...string) string {
	var items []string
	for _, id := range ids {
		items = append(items, `{"id":"`+id+`","kind":"web","title":"t","urls":["https://e.test"],"reviewedAt":"2026-01-01"}`)
	}
	return `{"version":1,"sources":[` + strings.Join(items, ",") + `]}`
}

func TestLoadMergesRootAndShardsInOrder(t *testing.T) {
	validator := &scripted{}
	loader := sources.Loader{
		FS: memfs.New(map[string]string{
			"d/sources.json":            file("root"),
			"d/sources.d/b.json":        file("b"),
			"d/sources.d/a.json":        `{"$schema":"x","version":1,"sources":[{"id":"a","kind":"git","title":"t","checkout":"c","urls":["https://e.test"],"ref":"v1","reviewedAt":"2026-01-01","feeds":["f.md"],"claims":[{"id":"c","statement":"s","anchor":"h","evidence":"e"}]}]}`,
			"d/sources.d/notes.txt":     "ignored",
			"d/sources.d/nested/x.json": file("ignored"),
		}),
		Validator: validator,
	}
	catalog, got, err := loader.Load("d")
	if err != nil {
		t.Fatal(err)
	}
	var order []string
	for _, s := range catalog.Sources {
		order = append(order, s.ID+"@"+s.File)
	}
	want := []string{"root@d/sources.json", "a@d/sources.d/a.json", "b@d/sources.d/b.json"}
	if strings.Join(order, ",") != strings.Join(want, ",") {
		t.Fatalf("sources = %v; want %v", order, want)
	}
	if len(got) != 1 || got[0].Code != sources.CodeOK {
		t.Fatalf("findings = %v", got)
	}
	for _, id := range validator.asked {
		if id != schemas.Sources {
			t.Fatalf("validated against %q", id)
		}
	}
	if len(validator.asked) != 3 {
		t.Fatalf("validated %d files", len(validator.asked))
	}
	if a, _ := catalog.ByID("a"); a.Claims[0].Evidence != "e" || a.Feeds[0] != "f.md" || a.Ref != "v1" {
		t.Fatalf("decoded source = %+v", a)
	}
}

func TestLoadShardsOnly(t *testing.T) {
	loader := sources.Loader{FS: memfs.New(map[string]string{"d/sources.d/a.json": file("a")}), Validator: &scripted{}}
	catalog, got, err := loader.Load("d")
	if err != nil || len(catalog.Sources) != 1 || len(got) != 1 || got[0].Level != findings.Pass {
		t.Fatalf("Load = %+v, %v, %v", catalog, got, err)
	}
}

func TestLoadRootOnly(t *testing.T) {
	loader := sources.Loader{FS: memfs.New(map[string]string{"d/sources.json": file("a")}), Validator: &scripted{}}
	if catalog, _, err := loader.Load("d"); err != nil || len(catalog.Sources) != 1 {
		t.Fatalf("Load = %+v, %v", catalog, err)
	}
}

func TestLoadMissingCatalog(t *testing.T) {
	loader := sources.Loader{FS: memfs.New(map[string]string{"d/other.txt": ""}), Validator: &scripted{}}
	catalog, got, err := loader.Load("d")
	if err != nil || len(catalog.Sources) != 0 {
		t.Fatalf("Load = %+v, %v", catalog, err)
	}
	if len(got) != 1 || got[0].Code != sources.CodeMissing || got[0].Level != findings.Fail || got[0].Path != "d" {
		t.Fatalf("findings = %v", got)
	}
}

func TestLoadSchemaViolationsBecomeFindings(t *testing.T) {
	validator := &scripted{byMarker: map[string][]jsonschema.Violation{
		"BROKEN": {{Pointer: "/sources/0/kind", Message: "bad kind"}, {Message: "bad root"}},
	}}
	loader := sources.Loader{
		FS: memfs.New(map[string]string{
			"d/sources.json":     file("good"),
			"d/sources.d/x.json": `{"BROKEN":true}`,
		}),
		Validator: validator,
	}
	catalog, got, err := loader.Load("d")
	if err != nil {
		t.Fatal(err)
	}
	if len(catalog.Sources) != 1 || catalog.Sources[0].ID != "good" {
		t.Fatalf("a broken file must add no sources: %+v", catalog.Sources)
	}
	if len(got) != 2 {
		t.Fatalf("findings = %v", got)
	}
	if got[0].Code != sources.CodeSchema || got[0].Path != "d/sources.d/x.json" || got[0].Message != "/sources/0/kind: bad kind" {
		t.Fatalf("first = %+v", got[0])
	}
	if got[1].Message != "/: bad root" {
		t.Fatalf("root violation = %q", got[1].Message)
	}
	if findings.ExitCode(got, nil, findings.DefaultPolicy()) != findings.ExitFindings {
		t.Fatal("schema violations must fail the run")
	}
}

func TestLoadAppliesCrossSourceChecks(t *testing.T) {
	loader := sources.Loader{
		FS:        memfs.New(map[string]string{"d/sources.json": file("a"), "d/sources.d/b.json": file("a")}),
		Validator: &scripted{},
	}
	_, got, err := loader.Load("d")
	if err != nil || len(got) != 1 || got[0].Code != sources.CodeDuplicateID {
		t.Fatalf("Load = %v, %v", got, err)
	}
}

func TestLoadNeverReportsAPassBesideASchemaFailure(t *testing.T) {
	validator := &scripted{byMarker: map[string][]jsonschema.Violation{"BROKEN": {{Message: "x"}}}}
	loader := sources.Loader{FS: memfs.New(map[string]string{"d/sources.json": `{"BROKEN":1}`}), Validator: validator}
	_, got, _ := loader.Load("d")
	for _, f := range got {
		if f.Level == findings.Pass {
			t.Fatalf("unexpected pass: %v", got)
		}
	}
}

func TestLoadInfrastructureErrors(t *testing.T) {
	boom := errors.New("boom")
	t.Run("validator", func(t *testing.T) {
		loader := sources.Loader{FS: memfs.New(map[string]string{"d/sources.json": file("a")}), Validator: &scripted{err: boom}}
		if _, _, err := loader.Load("d"); !errors.Is(err, boom) {
			t.Fatalf("err = %v", err)
		}
	})
	t.Run("stat", func(t *testing.T) {
		loader := sources.Loader{FS: faulty{FileSystem: memfs.New(nil), stat: boom}, Validator: &scripted{}}
		if _, _, err := loader.Load("d"); !errors.Is(err, boom) {
			t.Fatalf("err = %v", err)
		}
	})
	t.Run("list shards", func(t *testing.T) {
		loader := sources.Loader{FS: faulty{FileSystem: memfs.New(nil), readDir: boom}, Validator: &scripted{}}
		if _, _, err := loader.Load("d"); !errors.Is(err, boom) {
			t.Fatalf("err = %v", err)
		}
	})
	t.Run("read", func(t *testing.T) {
		loader := sources.Loader{FS: faulty{FileSystem: memfs.New(map[string]string{"d/sources.json": file("a")}), read: boom}, Validator: &scripted{}}
		if _, _, err := loader.Load("d"); !errors.Is(err, boom) {
			t.Fatalf("err = %v", err)
		}
	})
	t.Run("read of a later file", func(t *testing.T) {
		loader := sources.Loader{
			FS:        faulty{FileSystem: memfs.New(map[string]string{"d/sources.json": file("a"), "d/sources.d/b.json": file("b")}), read: boom, readOnly: "d/sources.d/b.json"},
			Validator: &scripted{},
		}
		if _, _, err := loader.Load("d"); !errors.Is(err, boom) {
			t.Fatalf("err = %v", err)
		}
	})
	t.Run("the schema accepted a file the model cannot decode", func(t *testing.T) {
		loader := sources.Loader{FS: memfs.New(map[string]string{"d/sources.json": `{"sources":"no"}`}), Validator: &scripted{}}
		_, _, err := loader.Load("d")
		if err == nil || !strings.Contains(err.Error(), "decode d/sources.json") {
			t.Fatalf("err = %v", err)
		}
	})
}

// faulty wraps a file system and fails the calls it is told to.
type faulty struct {
	fs.FileSystem
	stat, readDir, read error
	readOnly            string
}

func (f faulty) Stat(p string) (fs.Info, error) {
	if f.stat != nil {
		return fs.Info{}, f.stat
	}
	return f.FileSystem.Stat(p)
}

func (f faulty) ReadDir(p string) ([]fs.Entry, error) {
	if f.readDir != nil {
		return nil, f.readDir
	}
	return f.FileSystem.ReadDir(p)
}

func (f faulty) ReadFile(p string) ([]byte, error) {
	if f.read != nil && (f.readOnly == "" || f.readOnly == p) {
		return nil, f.read
	}
	return f.FileSystem.ReadFile(p)
}
