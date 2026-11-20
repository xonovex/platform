package wiring_test

import (
	"errors"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/findings"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/fs/memfs"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/sources"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/wiring"
)

func TestSourcesLoaderEndToEnd(t *testing.T) {
	fsys := memfs.New(map[string]string{
		"skill/sources.json": `{"version":1,"sources":[
			{"id":"spec","kind":"http","title":"Spec","url":"https://e.test/spec","reviewedAt":"2026-01-01","reviewAfter":"2026-02-01","claims":[{"id":"c","statement":"s","anchor":"intro"}]}]}`,
		"skill/sources.d/extra.json": `{"version":1,"sources":[
			{"id":"repo","kind":"repository","title":"Repo","path":"docs","reviewedAt":"2026-01-01","supersededBy":"spec"}]}`,
	})
	loader, err := wiring.NewSourcesLoader(fsys)
	if err != nil {
		t.Fatal(err)
	}
	catalog, got, err := loader.Load("skill")
	if err != nil {
		t.Fatal(err)
	}
	if len(catalog.Sources) != 2 || len(got) != 1 || got[0].Code != sources.CodeOK {
		t.Fatalf("Load = %+v, %v", catalog, got)
	}
}

func TestSourcesLoaderReportsRealSchemaViolations(t *testing.T) {
	fsys := memfs.New(map[string]string{
		"skill/sources.json": `{"version":1,"sources":[{"id":"spec","kind":"git","title":"Spec","reviewedAt":"2026-01-01"}]}`,
	})
	loader, err := wiring.NewSourcesLoader(fsys)
	if err != nil {
		t.Fatal(err)
	}
	_, got, err := loader.Load("skill")
	if err != nil || len(got) == 0 {
		t.Fatalf("Load = %v, %v", got, err)
	}
	for _, f := range got {
		if f.Code != sources.CodeSchema || f.Level != findings.Fail || f.Path != "skill/sources.json" {
			t.Fatalf("finding = %+v", f)
		}
	}
}

func TestValidatorFromReportsBrokenSchemaSets(t *testing.T) {
	broken := map[string]func() (map[string][]byte, error){
		"source fails": func() (map[string][]byte, error) { return nil, errors.New("embed broke") },
		"schema fails": func() (map[string][]byte, error) { return map[string][]byte{"https://x.test/a": []byte(`{`)}, nil },
	}
	for name, source := range broken {
		t.Run(name, func(t *testing.T) {
			if _, err := wiring.ValidatorFrom(source); err == nil {
				t.Fatal("ValidatorFrom must fail")
			}
		})
	}
}
