package marketplace_test

import (
	"fmt"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/findings"

	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/check/marketplace"
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/checktest"
)

const dir = "packages/runtime/plugin/plugin-git"

func plugin(name, version, desc string) string {
	return fmt.Sprintf(`{"name":%q,"version":%q,"description":%q,"skills":"./skills/"}`, name, version, desc)
}

func base() map[string]string {
	market := `{"metadata":{"version":"1.2.3"},"plugins":[{"name":"xonovex-git","source":"./` + dir + `","description":"d"}]}`
	codex := `{"plugins":[{"name":"xonovex-git","source":{"path":"./` + dir + `"},"description":"d"}]}`
	return map[string]string{
		".claude-plugin/marketplace.json":   market,
		".agents/plugins/marketplace.json":  codex,
		"package-lock.json":                 `{"packages":{"` + dir + `":{"version":"1.2.3"}}}`,
		dir + "/package.json":               `{"name":"@xonovex/plugin-git","version":"1.2.3","description":"d","dependencies":{"@xonovex/plugin-base":"1.2.3","left-pad":"1.0.0"}}`,
		dir + "/plugin.json":                plugin("xonovex-git", "1.2.3", "d"),
		dir + "/.claude-plugin/plugin.json": plugin("xonovex-git", "1.2.3", "d"),
		dir + "/.codex-plugin/plugin.json":  plugin("xonovex-git", "1.2.3", "d"),
	}
}

func run(t *testing.T, files map[string]string, options ...any) []findings.Finding {
	t.Helper()
	ctx := checktest.Context(t, files, "packages/runtime/plugin")
	if len(options) > 0 {
		ctx = checktest.WithOptions(t, ctx, options[0])
	}
	return checktest.Run(t, marketplace.Check{}, ctx)
}

func with(changes map[string]string) map[string]string {
	files := base()
	for k, v := range changes {
		if v == "\x00" {
			delete(files, k)
		} else {
			files[k] = v
		}
	}
	return files
}

func TestIdentityAndCleanRun(t *testing.T) {
	c := marketplace.Check{}
	if c.Name() != "marketplace" || c.Summary() == "" || c.DefaultPaths()[0] != "packages/runtime/plugin" {
		t.Fatal("identity")
	}
	if got := run(t, base()); len(got) != 0 {
		t.Fatalf("findings = %v", got)
	}
	ctx := checktest.WithOptions(t, checktest.Context(t, nil, "x"), map[string]any{"x": 1})
	if _, err := c.Run(ctx); err == nil {
		t.Fatal("unknown option")
	}
}

func TestMarketplaceFiles(t *testing.T) {
	cases := map[string]struct {
		changes map[string]string
		code    string
	}{
		"missing claude": {map[string]string{".claude-plugin/marketplace.json": "\x00"}, "marketplace.missing"},
		"broken codex":   {map[string]string{".agents/plugins/marketplace.json": "{"}, "marketplace.invalid"},
		"bad version":    {map[string]string{".claude-plugin/marketplace.json": `{"metadata":{"version":"1.2"},"plugins":[]}`}, "marketplace.version"},
	}
	for name, tc := range cases {
		if got := run(t, with(tc.changes)); !checktest.Has(got, tc.code) {
			t.Errorf("%s: codes = %v", name, checktest.Codes(got))
		}
	}
}

func TestPackageRules(t *testing.T) {
	cases := map[string]struct {
		changes map[string]string
		code    string
	}{
		"neutral":             {map[string]string{dir + "/plugin.json": plugin("xonovex-git", "1.2.3", "other")}, "marketplace.neutral"},
		"package version":     {map[string]string{dir + "/package.json": `{"name":"p","version":"0.0.1","description":"d"}`}, "marketplace.version-package"},
		"claude version":      {map[string]string{dir + "/.claude-plugin/plugin.json": plugin("xonovex-git", "9.9.9", "d")}, "marketplace.version-claude"},
		"codex version":       {map[string]string{dir + "/.codex-plugin/plugin.json": plugin("xonovex-git", "9.9.9", "d")}, "marketplace.version-codex"},
		"claude source":       {map[string]string{".claude-plugin/marketplace.json": `{"metadata":{"version":"1.2.3"},"plugins":[{"name":"xonovex-git","source":"./elsewhere","description":"d"}]}`}, "marketplace.source-claude"},
		"codex source":        {map[string]string{".agents/plugins/marketplace.json": `{"plugins":[{"name":"xonovex-git","source":5,"description":"d"}]}`}, "marketplace.source-codex"},
		"names":               {map[string]string{dir + "/.codex-plugin/plugin.json": plugin("xonovex-other", "1.2.3", "d")}, "marketplace.names"},
		"description missing": {map[string]string{dir + "/package.json": `{"name":"p","version":"1.2.3"}`}, "marketplace.description-missing"},
		"description claude":  {map[string]string{dir + "/.claude-plugin/plugin.json": plugin("xonovex-git", "1.2.3", "x")}, "marketplace.description"},
		"dependency version":  {map[string]string{dir + "/package.json": `{"name":"p","version":"1.2.3","description":"d","dependencies":{"@xonovex/plugin-base":"1.0.0"}}`}, "marketplace.dependency-version"},
		"missing package":     {map[string]string{dir + "/package.json": "\x00"}, "marketplace.package-missing"},
		"broken manifest":     {map[string]string{dir + "/plugin.json": "{"}, "marketplace.package-invalid"},
	}
	for name, tc := range cases {
		if got := run(t, with(tc.changes)); !checktest.Has(got, tc.code) {
			t.Errorf("%s: codes = %v", name, checktest.Codes(got))
		}
	}
}

func TestInventory(t *testing.T) {
	got := run(t, with(map[string]string{
		".claude-plugin/marketplace.json": `{"metadata":{"version":"1.2.3"},"plugins":[{"name":"xonovex-ghost","source":"./x","description":"d"},{"name":"xonovex-ghost","source":"./x","description":"d"}]}`,
	}))
	for _, code := range []string{"marketplace.missing-entry", "marketplace.unexpected-entry", "marketplace.duplicate-entry"} {
		if !checktest.Has(got, code) {
			t.Errorf("missing %s in %v", code, checktest.Codes(got))
		}
	}
}

func TestLockfile(t *testing.T) {
	cases := map[string]struct {
		changes map[string]string
		code    string
	}{
		"missing": {map[string]string{"package-lock.json": "\x00"}, "marketplace.lockfile-missing"},
		"invalid": {map[string]string{"package-lock.json": "{"}, "marketplace.lockfile-invalid"},
		"stale":   {map[string]string{"package-lock.json": `{"packages":{}}`}, "marketplace.lockfile"},
	}
	for name, tc := range cases {
		if got := run(t, with(tc.changes)); !checktest.Has(got, tc.code) {
			t.Errorf("%s: codes = %v", name, checktest.Codes(got))
		}
	}
	if got := run(t, with(map[string]string{"package-lock.json": "\x00"}), map[string]any{"lockfile": ""}); len(got) != 0 {
		t.Fatalf("the lockfile rule is optional: %v", got)
	}
}

func TestManifestDependencyList(t *testing.T) {
	list := `{"name":"xonovex-git","version":"1.2.3","description":"d","skills":"./skills/","dependencies":["xonovex-base"]}`
	got := run(t, with(map[string]string{dir + "/.claude-plugin/plugin.json": list}))
	if checktest.Has(got, "marketplace.package-invalid") {
		t.Fatalf("a dependency array is a plugin manifest field: %v", got)
	}
}
