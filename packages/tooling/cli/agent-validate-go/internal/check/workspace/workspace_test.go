package workspace_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/findings"

	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/check/workspace"
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/checktest"
)

func run(t *testing.T, files map[string]string, repo *checktest.Repo, options ...any) []findings.Finding {
	t.Helper()
	ctx := checktest.Context(t, files, "packages")
	if repo != nil {
		ctx.Repository = *repo
	}
	if len(options) > 0 {
		ctx = checktest.WithOptions(t, ctx, options[0])
	}
	return checktest.Run(t, workspace.Check{}, ctx)
}

func runWithoutRepository(t *testing.T, files map[string]string, options any) []findings.Finding {
	t.Helper()
	ctx := checktest.Context(t, files, "packages")
	ctx.Repository = nil
	return checktest.Run(t, workspace.Check{}, checktest.WithOptions(t, ctx, options))
}

func rules(names ...string) map[string]any { return map[string]any{"rules": names} }

func TestIdentity(t *testing.T) {
	c := workspace.Check{}
	if c.Name() != "workspace" || c.Summary() == "" || c.DefaultPaths()[0] != "packages" {
		t.Fatal("identity")
	}
	ctx := checktest.WithOptions(t, checktest.Context(t, nil, "packages"), map[string]any{"rules": []string{"nope"}})
	if _, err := c.Run(ctx); err == nil || !strings.Contains(err.Error(), "known rules") {
		t.Fatalf("err = %v", err)
	}
	got := run(t, nil, nil)
	if !checktest.Has(got, "workspace.tag-files-missing") || !checktest.Has(got, "workspace.project-files-missing") {
		t.Fatalf("an empty workspace has nothing to check: %v", got)
	}
	complete := map[string]string{".moon/tasks/all.yml": "tasks: {}\n", "packages/a/moon.yml": "tags: [go]\n"}
	if got := run(t, complete, nil); len(got) != 0 {
		t.Fatalf("a workspace with tag files and project files is clean: %v", got)
	}
	noTags := map[string]string{"packages/a/moon.yml": "tags: [go]\n"}
	if got := run(t, noTags, nil, rules("task-files")); !checktest.Has(got, "workspace.tag-files-missing") || checktest.Has(got, "workspace.project-files-missing") {
		t.Fatalf("codes = %v", checktest.Codes(got))
	}
}

func TestTaskInheritance(t *testing.T) {
	files := map[string]string{
		".moon/tasks/all.yml":      "tasks:\n  ci-check:\n    deps: [a, b]\n",
		".moon/tasks/tag-ts.yml":   "extends: all.yml\ntasks:\n  ci-check:\n    deps: [a, c]\n  other:\n    deps: [x]\n",
		".moon/tasks/tag-ok.yml":   "extends: all.yml\ntasks:\n  ci-check:\n    deps:\n      - target: a\n      - b\n",
		".moon/tasks/tag-mg.yml":   "extends: tag-ts.yml\ntasks:\n  ci-check:\n    deps: [z]\n    options:\n      mergeDeps: replace\n",
		".moon/tasks/tag-lost.yml": "extends: missing.yml\n",
		".moon/tasks/readme.txt":   "ignored",
		".moon/tasks/dir/x.yml":    "x",
	}
	got := run(t, files, nil, rules("task-inheritance"))
	var dropped, extends int
	for _, f := range got {
		switch f.Code {
		case "workspace.tasks-dropped":
			dropped++
			if !strings.Contains(f.Message, "drops inherited b") {
				t.Fatalf("message = %s", f.Message)
			}
		case "workspace.tasks-extends":
			extends++
		}
	}
	if dropped != 1 || extends != 1 {
		t.Fatalf("dropped=%d extends=%d: %v", dropped, extends, got)
	}
	broken := map[string]string{".moon/tasks/a.yml": "tasks: [", ".moon/tasks/b.yml": "tasks: 3\n"}
	got = run(t, broken, nil, rules("task-inheritance"))
	if !checktest.Has(got, "workspace.tasks-yaml") || !checktest.Has(got, "workspace.tasks-shape") {
		t.Fatalf("codes = %v", checktest.Codes(got))
	}
	cycle := map[string]string{".moon/tasks/a.yml": "extends: b.yml\ntasks:\n  t:\n    deps: [x]\n", ".moon/tasks/b.yml": "extends: a.yml\ntasks:\n  t:\n    deps: [x, y]\n"}
	if got := run(t, cycle, nil, rules("task-inheritance")); len(got) != 1 {
		t.Fatalf("a cycle stops: %v", got)
	}
}

func TestCoverageFloors(t *testing.T) {
	floors := "TS_COVERAGE_MIN_LINES: 80\n      TS_COVERAGE_MIN_FUNCTIONS: 80\n      TS_COVERAGE_MIN_BRANCHES: 80\n      TS_COVERAGE_MIN_STATEMENTS: 80\n"
	files := map[string]string{
		"packages/ok/moon.yml":      "tags: [typescript]\ntasks:\n  ts-coverage:\n    env:\n      " + floors,
		"packages/partial/moon.yml": "tags: [typescript-script]\ntasks:\n  ts-coverage:\n    env:\n      TS_COVERAGE_MIN_LINES: 80\n      TS_COVERAGE_MIN_FUNCTIONS: ''\n",
		"packages/none/moon.yml":    "tags: [typescript]\n",
		"packages/go/moon.yml":      "tags: [go]\n",
		"packages/empty/moon.yml":   "tags:\ntasks:\n",
		"packages/bad/moon.yml":     "tags: 3\n",
		"packages/syntax/moon.yml":  "tags: [",
	}
	got := run(t, files, nil, rules("coverage-floors"))
	byPath := map[string]string{}
	for _, f := range got {
		byPath[f.Path] = f.Code
	}
	want := map[string]string{
		"packages/partial/moon.yml": "workspace.coverage-floor",
		"packages/none/moon.yml":    "workspace.coverage-floor",
		"packages/bad/moon.yml":     "workspace.coverage-shape",
		"packages/syntax/moon.yml":  "workspace.coverage-yaml",
	}
	if len(byPath) != len(want) {
		t.Fatalf("findings = %v", got)
	}
	for p, code := range want {
		if byPath[p] != code {
			t.Errorf("%s: %s, want %s", p, byPath[p], code)
		}
	}
}

func TestDependencySources(t *testing.T) {
	files := map[string]string{
		"packages/lib/moon.yml":         "tags: [typescript]\n",
		"packages/lib/package.json":     `{"name":"@x/lib"}`,
		"packages/cfg/moon.yml":         "tags: [typescript-config]\n",
		"packages/cfg/package.json":     `{"name":"@x/cfg"}`,
		"packages/gopkg/moon.yml":       "tags: [go]\n",
		"packages/app/moon.yml":         "tags: [typescript]\ndependsOn: [lib, {id: gopkg}, unknown]\nfileGroups:\n  dependencySources:\n    - /packages/lib/src/**/*\n    - src/**/*\n",
		"packages/app/eslint.config.ts": "import c from '@x/cfg';\n",
		"packages/app/package.json":     `{"name":"@x/app"}`,
		"packages/ok/moon.yml":          "tags: [typescript]\ndependsOn: [lib]\nfileGroups:\n  dependencySources:\n    - /packages/lib/src/**/*\n",
		"packages/extra/moon.yml":       "tags: [typescript]\ndependsOn: [lib]\nfileGroups:\n  dependencySources:\n    - /packages/lib/src/**/*\n    - /packages/ghost/src/**/*\n",
		"packages/alone/moon.yml":       "tags: [typescript]\n",
		"packages/chain/moon.yml":       "tags: [typescript]\ndependsOn: [ok]\nfileGroups:\n  dependencySources:\n    - /packages/lib/src/**/*\n    - /packages/ok/src/**/*\n",
		"packages/bad/moon.yml":         "dependsOn: 3\n",
		"packages/syntax/moon.yml":      "x: [",
	}
	got := run(t, files, nil, rules("dependency-sources"))
	codes := map[string][]string{}
	for _, f := range got {
		codes[f.Path] = append(codes[f.Path], f.Code)
	}
	if len(codes["packages/app/moon.yml"]) != 1 || codes["packages/app/moon.yml"][0] != "workspace.sources-missing" {
		t.Fatalf("app: %v (%v)", codes["packages/app/moon.yml"], got)
	}
	if f, _ := checktest.Find(got, "workspace.sources-missing"); !strings.Contains(f.Message, "/packages/cfg/src/**/*") {
		t.Fatalf("the config package is a source: %s", f.Message)
	}
	if len(codes["packages/extra/moon.yml"]) != 1 || codes["packages/extra/moon.yml"][0] != "workspace.sources-extra" {
		t.Fatalf("extra: %v", codes["packages/extra/moon.yml"])
	}
	for _, clean := range []string{"packages/ok/moon.yml", "packages/chain/moon.yml", "packages/alone/moon.yml", "packages/lib/moon.yml"} {
		if len(codes[clean]) != 0 {
			t.Errorf("%s: %v", clean, codes[clean])
		}
	}
	if !checktest.Has(got, "workspace.sources-shape") || !checktest.Has(got, "workspace.sources-yaml") {
		t.Fatalf("codes = %v", checktest.Codes(got))
	}
}

func TestGoDependencies(t *testing.T) {
	const prefix = "github.com/x/"
	files := map[string]string{
		"packages/core/moon.yml":              "tags: [go]\n",
		"packages/core/go.mod":                "module " + prefix + "core\n\ngo 1.27.0\n",
		"packages/contract/moon.yml":          "tags: [go]\n",
		"packages/contract/go.mod":            "module " + prefix + "contract\n\nrequire github.com/other/mod v1.0.0\n",
		"packages/config/moon.yml":            "tags: [go]\ndependsOn: [contract]\n",
		"packages/config/go.mod":              "module " + prefix + "config\n\nrequire " + prefix + "contract v0.0.0 // local\n",
		"packages/tool/moon.yml":              "tags: [go]\ndependsOn: [config, {id: core}, eslint-config]\n",
		"packages/tool/go.mod":                "module " + prefix + "tool\n\nrequire (\n\t" + prefix + "config v0.0.0\n\t" + prefix + "contract v0.0.0 // indirect\n\n\tgithub.com/other/mod v1.0.0\n)\n",
		"packages/eslint-config/package.json": `{"name":"@x/eslint-config"}`,
		"packages/broken/moon.yml":            "dependsOn: 3\n",
		"packages/broken/go.mod":              "module " + prefix + "broken\n",
		"packages/ts/moon.yml":                "tags: [typescript]\ndependsOn: [core]\n",
	}
	got := run(t, files, nil, rules("go-dependencies"))
	codes := map[string][]string{}
	for _, f := range got {
		codes[f.Path] = append(codes[f.Path], f.Code)
	}
	if want := []string{"workspace.go-dependencies-missing", "workspace.go-dependencies-extra"}; strings.Join(codes["packages/tool/moon.yml"], ",") != strings.Join(want, ",") {
		t.Fatalf("tool: %v (%v)", codes["packages/tool/moon.yml"], got)
	}
	missing, _ := checktest.Find(got, "workspace.go-dependencies-missing")
	extra, _ := checktest.Find(got, "workspace.go-dependencies-extra")
	if !strings.Contains(missing.Message, "requires contract,") || !strings.Contains(extra.Message, "names core,") {
		t.Fatalf("messages = %q, %q", missing.Message, extra.Message)
	}
	for _, clean := range []string{"packages/core/moon.yml", "packages/contract/moon.yml", "packages/config/moon.yml", "packages/ts/moon.yml"} {
		if len(codes[clean]) != 0 {
			t.Errorf("%s: %v", clean, codes[clean])
		}
	}
	if !checktest.Has(got, "workspace.go-dependencies-shape") {
		t.Fatalf("codes = %v", checktest.Codes(got))
	}
}

const workspaceFile = "hasher:\n  walkStrategy: glob\n  ignorePatterns:\n    - '**/node_modules/**/*'\n    - '**/coverage/**/*'\n    - '**/*.log'\n    - packages/*/skip.txt\n"

func TestHasher(t *testing.T) {
	files := map[string]string{
		".moon/workspace.yml": workspaceFile,
		"packages/a/moon.yml": "tasks:\n  t:\n    inputs:\n      - app.log\n      - /packages/a/skip.txt\n      - src/**/*\n      - $projectRoot/x\n      - kept.txt\n",
	}
	repo := &checktest.Repo{Ignored: []string{"node_modules/", "packages/a/coverage/", "tmp/", "build.log", "packages/a/dist/", "other"}}
	got := run(t, files, repo, rules("hasher"))
	uncovered, input := 0, 0
	for _, f := range got {
		switch f.Code {
		case "workspace.hasher-uncovered":
			uncovered++
		case "workspace.hasher-input":
			input++
		}
	}
	if uncovered != 2 || input != 2 {
		t.Fatalf("uncovered=%d input=%d: %v", uncovered, input, got)
	}
}

func TestHasherConditions(t *testing.T) {
	vcs := map[string]string{".moon/workspace.yml": "hasher:\n  walkStrategy: vcs\n"}
	if got := run(t, vcs, nil, rules("hasher")); len(got) != 0 {
		t.Fatalf("the vcs walk applies .gitignore: %v", got)
	}
	if got := run(t, nil, nil, rules("hasher")); len(got) != 0 {
		t.Fatalf("no workspace file: %v", got)
	}
	glob := map[string]string{".moon/workspace.yml": workspaceFile}
	if got := runWithoutRepository(t, glob, rules("hasher")); !checktest.Has(got, "workspace.hasher-git") {
		t.Fatalf("no repository: %v", got)
	}
	if got := run(t, glob, &checktest.Repo{Err: errors.New("boom")}, rules("hasher")); !checktest.Has(got, "workspace.hasher-git") {
		t.Fatalf("repository error: %v", got)
	}
	broken := map[string]string{".moon/workspace.yml": "hasher: 3\n"}
	if got := run(t, broken, nil, rules("hasher")); !checktest.Has(got, "workspace.hasher-shape") {
		t.Fatalf("codes = %v", checktest.Codes(got))
	}
	badGlob := map[string]string{".moon/workspace.yml": "hasher:\n  walkStrategy: glob\n  ignorePatterns:\n    - '[x'\n"}
	got := run(t, badGlob, &checktest.Repo{}, rules("hasher"))
	if len(got) != 0 {
		t.Fatalf("an escaped bracket is a literal: %v", got)
	}
}

func TestBinTargets(t *testing.T) {
	files := map[string]string{
		"packages/priv/package.json":        `{"name":"@x/priv","private":true,"bin":{"priv":"./dist/bin.js"}}`,
		"packages/priv2/package.json":       `{"name":"@x/priv2","private":true,"bin":"bin.js"}`,
		"packages/pub/package.json":         `{"name":"@x/pub","files":["dist/"],"bin":{"pub":"dist/cli.js"}}`,
		"packages/pub2/package.json":        `{"name":"@x/pub2","bin":{"pub2":"dist/cli.js"}}`,
		"packages/none/package.json":        `{"name":"@x/none"}`,
		"packages/bad/package.json":         `{`,
		"packages/skip/package.json":        `{"name":"@x/skip","bin":{"x":"x.js"}}`,
		"packages/skip/nested/package.json": `{"name":"nested","bin":{"n":"n.js"}}`,
		"packages/arr/package.json":         `{"name":"arr","bin":["x"]}`,
	}
	repo := &checktest.Repo{Tracked: []string{"packages/priv2/bin.js", "packages/skip/x.js"}}
	got := run(t, files, repo, rules("bin-targets"))
	messages := []string{}
	for _, f := range got {
		messages = append(messages, f.Code+" "+f.Path)
	}
	want := []string{
		"workspace.bin-manifest packages/bad/package.json",
		"workspace.bin-target packages/priv/package.json",
		"workspace.bin-target packages/pub2/package.json",
	}
	if len(got) != len(want) {
		t.Fatalf("findings = %v", messages)
	}
	for i, f := range got {
		if f.Code+" "+f.Path != want[i] {
			t.Errorf("finding %d = %s, want %s", i, f.Code+" "+f.Path, want[i])
		}
	}
	if !strings.Contains(got[1].Message, "cold install") || !strings.Contains(got[2].Message, "files allowlist") {
		t.Fatalf("messages = %v", got)
	}
	if got := runWithoutRepository(t, files, rules("bin-targets")); !checktest.Has(got, "workspace.bin-git") {
		t.Fatalf("no repository: %v", got)
	}
	if got := run(t, files, &checktest.Repo{Err: errors.New("boom")}, rules("bin-targets")); !checktest.Has(got, "workspace.bin-git") {
		t.Fatalf("repository error: %v", got)
	}
}

func TestProjectsInPluginSkillDirectories(t *testing.T) {
	files := map[string]string{
		"packages/plugin/package.json":              `{"name":"p"}`,
		"packages/plugin/moon.yml":                  "tags: [go]\n",
		"packages/plugin/skills/using-x/SKILL.md":   "x",
		"packages/plugin/skills/using-x/moon.yml":   "tags: [typescript]\n",
		"packages/plugin/skills/notaskill/moon.yml": "tags: [typescript]\n",
	}
	got := run(t, files, nil, rules("coverage-floors"))
	if len(got) != 1 || got[0].Path != "packages/plugin/skills/using-x/moon.yml" {
		t.Fatalf("findings = %v", got)
	}
}
