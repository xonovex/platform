//go:build integration

// Package integration holds the tests that read the real repository. They carry
// the integration build tag and run in the ci-integration task, never in
// ci-check.
package integration

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/discovery"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/fs/osfs"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/schemas"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/sources"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/wiring"
)

// repositoryRoot returns the root of the repository that holds this module.
func repositoryRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	// file = <root>/packages/tooling/cli/agent-audit-go/test/integration/convert_test.go
	return filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(file)))))))
}

func TestEverySourcesFileOfThePluginsIsValidCatalogData(t *testing.T) {
	fsys := osfs.New()
	root := filepath.ToSlash(filepath.Join(repositoryRoot(t), "packages/runtime/plugin"))
	catalogs, err := discovery.Walk(fsys, root, discovery.Options{Skip: discovery.DefaultSkip(), Match: discovery.Named("sources.json")})
	if err != nil {
		t.Fatal(err)
	}
	if len(catalogs) < 80 {
		t.Fatalf("found %d sources.json files; the catalog has more than 80", len(catalogs))
	}
	legacyFiles, err := discovery.Walk(fsys, root, discovery.Options{Skip: discovery.DefaultSkip(), Match: discovery.Named("SOURCES.md")})
	if err != nil {
		t.Fatal(err)
	}
	if len(legacyFiles) != 0 {
		t.Fatalf("legacy SOURCES.md files remain: %v", legacyFiles)
	}
	validator, err := wiring.NewValidator()
	if err != nil {
		t.Fatal(err)
	}

	for _, file := range catalogs {
		t.Run(strings.TrimPrefix(file, root+"/"), func(t *testing.T) {
			data, err := fsys.ReadFile(file)
			if err != nil {
				t.Fatal(err)
			}
			violations, err := validator.Validate(schemas.Sources, data)
			if err != nil || len(violations) > 0 {
				t.Fatalf("violations = %v, err = %v", violations, err)
			}
			if _, err := sources.Decode(data); err != nil {
				t.Fatal(err)
			}
		})
	}
}
