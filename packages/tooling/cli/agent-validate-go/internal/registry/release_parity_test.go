package registry_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/findings"

	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/check"
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/checktest"
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/registry"
)

// The release contracts of the repository were once enforced by a separate
// validator. These tests keep each of its rules as one mutation of a complete
// release fixture: the fixture passes every release check, and each mutation
// breaks exactly one rule, which a named check must report.

const (
	pluginDir = "packages/runtime/plugin/plugin-git"
	toolDir   = "packages/tooling/tool"
	libDir    = "packages/tooling/lib"
)

// releaseChecks are the checks that carry the release contracts.
var releaseChecks = []string{"marketplace", "plugin", "repository", "workspace", "links"}

const releaseWorkflow = `name: Release
on:
  pull_request:
  workflow_dispatch:
jobs:
  release:
    steps:
      - uses: actions/checkout@v4
      - name: Publish
        if: github.event_name == 'pull_request'
        run: publish
      - name: Publish (dry run)
        if: github.event_name == 'workflow_dispatch'
        run: dry
      - uses: other
`

const coverageFloors = "      TS_COVERAGE_MIN_LINES: 80\n      TS_COVERAGE_MIN_FUNCTIONS: 80\n      TS_COVERAGE_MIN_BRANCHES: 80\n      TS_COVERAGE_MIN_STATEMENTS: 80\n"

func pluginManifest(version, description string) string {
	return `{"name":"xonovex-git","version":"` + version + `","description":"` + description + `","skills":"./skills/"}`
}

func releaseFixture() (map[string]string, checktest.Repo) {
	files := map[string]string{
		".claude-plugin/marketplace.json":  `{"metadata":{"version":"1.2.3"},"plugins":[{"name":"xonovex-git","source":"./` + pluginDir + `","description":"Git skills"}]}`,
		".agents/plugins/marketplace.json": `{"plugins":[{"name":"xonovex-git","source":{"path":"./` + pluginDir + `"},"description":"Git skills"}]}`,
		"package-lock.json":                `{"packages":{"` + pluginDir + `":{"version":"1.2.3"}}}`,

		pluginDir + "/package.json":               `{"name":"@xonovex/plugin-git","version":"1.2.3","description":"Git skills","dependencies":{"left-pad":"1.0.0"}}`,
		pluginDir + "/plugin.json":                pluginManifest("1.2.3", "Git skills"),
		pluginDir + "/.claude-plugin/plugin.json": pluginManifest("1.2.3", "Git skills"),
		pluginDir + "/.codex-plugin/plugin.json":  pluginManifest("1.2.3", "Git skills"),
		pluginDir + "/README.md":                  "# Git\n\n## Skills\n\n| Skill | Use |\n| --- | --- |\n| [using-git](skills/using-git/SKILL.md) | x |\n",
		pluginDir + "/skills/using-git/SKILL.md":  "---\nname: using-git\n---\n# Git\n",

		"README.md":       "# Platform\n\nSee the [plugin](" + pluginDir + "/README.md).\n",
		"AGENTS.md":       "# Root\n",
		"CLAUDE.md":       "See @AGENTS.md for complete documentation.",
		".moon/AGENTS.md": "# Moon\n",
		".moon/CLAUDE.md": "See @AGENTS.md for complete documentation.",

		".github/workflows/release.yml": releaseWorkflow,

		".moon/tasks/all.yml":            "tasks:\n  ci-check:\n    deps: [build, test]\n",
		".moon/tasks/tag-typescript.yml": "extends: all.yml\ntasks:\n  ci-check:\n    deps: [build, test, lint]\n",
		".moon/workspace.yml":            "hasher:\n  walkStrategy: glob\n  ignorePatterns:\n    - '**/node_modules/**/*'\n    - '**/coverage/**/*'\n",

		toolDir + "/package.json": `{"name":"@x/tool","private":true,"bin":{"tool":"./bin.js"}}`,
		toolDir + "/moon.yml":     "tags: [typescript]\ndependsOn: [lib]\nfileGroups:\n  dependencySources:\n    - /" + libDir + "/src/**/*\ntasks:\n  ts-coverage:\n    inputs: [bin.js]\n    env:\n" + coverageFloors,
		libDir + "/package.json":  `{"name":"@x/lib"}`,
		libDir + "/moon.yml":      "tags: [typescript]\ntasks:\n  ts-coverage:\n    env:\n" + coverageFloors,
	}
	for _, group := range []string{"packages", "packages/runtime", "packages/runtime/plugin", "packages/tooling"} {
		files[group+"/AGENTS.md"] = "# Group\n"
		files[group+"/CLAUDE.md"] = "See @AGENTS.md for complete documentation."
	}
	repo := checktest.Repo{
		Tracked: []string{toolDir + "/bin.js"},
		Ignored: []string{"node_modules/", "coverage/", "dist/"},
	}
	return files, repo
}

func releaseCheck(t *testing.T, name string) check.Check {
	t.Helper()
	for _, c := range registry.All() {
		if c.Name() == name {
			return c
		}
	}
	t.Fatalf("no check named %s", name)
	return nil
}

func runRelease(t *testing.T, name string, files map[string]string, repo checktest.Repo) []findings.Finding {
	t.Helper()
	c := releaseCheck(t, name)
	ctx := checktest.Context(t, files, c.DefaultPaths()...)
	ctx.Repository = repo
	return checktest.Run(t, c, ctx)
}

func TestReleaseFixtureIsClean(t *testing.T) {
	files, repo := releaseFixture()
	for _, name := range releaseChecks {
		if got := runRelease(t, name, files, repo); len(checktest.Codes(got)) != 0 {
			t.Errorf("%s: %v", name, got)
		}
	}
}

// mutation breaks one rule of the release contract.
type mutation struct {
	check string
	code  string
	apply func(files map[string]string, repo *checktest.Repo)
}

func replace(path, old, replacement string) func(map[string]string, *checktest.Repo) {
	return func(files map[string]string, _ *checktest.Repo) {
		if !strings.Contains(files[path], old) {
			panic("fixture drift: " + path + " lacks " + old)
		}
		files[path] = strings.Replace(files[path], old, replacement, 1)
	}
}

func set(path, content string) func(map[string]string, *checktest.Repo) {
	return func(files map[string]string, _ *checktest.Repo) { files[path] = content }
}

func remove(path string) func(map[string]string, *checktest.Repo) {
	return func(files map[string]string, _ *checktest.Repo) {
		if _, ok := files[path]; !ok {
			panic("fixture drift: no " + path)
		}
		delete(files, path)
	}
}

func repoFault(files map[string]string, repo *checktest.Repo) { repo.Err = errors.New("no git") }

func TestEveryFormerReleaseRuleStillFails(t *testing.T) {
	claude := ".claude-plugin/marketplace.json"
	codex := ".agents/plugins/marketplace.json"
	mutations := map[string]mutation{
		// Both marketplaces and the lockfile.
		"claude marketplace is not JSON":    {"marketplace", "marketplace.invalid", set(claude, "{")},
		"codex marketplace is absent":       {"marketplace", "marketplace.missing", remove(codex)},
		"release version is not semantic":   {"marketplace", "marketplace.version", replace(claude, `"1.2.3"`, `"1.2"`)},
		"lockfile records another version":  {"marketplace", "marketplace.lockfile", set("package-lock.json", `{"packages":{"`+pluginDir+`":{"version":"1.0.0"}}}`)},
		"lockfile is not JSON":              {"marketplace", "marketplace.lockfile-invalid", set("package-lock.json", "{")},
		"lockfile is absent":                {"marketplace", "marketplace.lockfile-missing", remove("package-lock.json")},
		"marketplace repeats a plugin":      {"marketplace", "marketplace.duplicate-entry", replace(claude, `"plugins":[`, `"plugins":[{"name":"xonovex-git","source":"./`+pluginDir+`","description":"Git skills"},`)},
		"codex marketplace lacks a plugin":  {"marketplace", "marketplace.missing-entry", set(codex, `{"plugins":[]}`)},
		"codex marketplace has an extra":    {"marketplace", "marketplace.unexpected-entry", replace(codex, `"plugins":[`, `"plugins":[{"name":"xonovex-ghost","source":"./x","description":"d"},`)},
		"claude marketplace source differs": {"marketplace", "marketplace.source-claude", replace(claude, `"source":"./`+pluginDir+`"`, `"source":"./elsewhere"`)},
		"codex marketplace source differs":  {"marketplace", "marketplace.source-codex", replace(codex, `"path":"./`+pluginDir+`"`, `"path":"./elsewhere"`)},

		// One release version and one description per plugin package.
		"package version drifts":          {"marketplace", "marketplace.version-package", replace(pluginDir+"/package.json", `"1.2.3"`, `"1.2.2"`)},
		"claude manifest version drifts":  {"marketplace", "marketplace.version-claude", set(pluginDir+"/.claude-plugin/plugin.json", pluginManifest("1.2.2", "Git skills"))},
		"codex manifest version drifts":   {"marketplace", "marketplace.version-codex", set(pluginDir+"/.codex-plugin/plugin.json", pluginManifest("1.2.2", "Git skills"))},
		"neutral manifest version drifts": {"marketplace", "marketplace.neutral", set(pluginDir+"/plugin.json", pluginManifest("1.2.2", "Git skills"))},
		"manifest names differ": {"marketplace", "marketplace.names", set(pluginDir+"/.codex-plugin/plugin.json",
			`{"name":"xonovex-other","version":"1.2.3","description":"Git skills","skills":"./skills/"}`)},
		"package has no description":                    {"marketplace", "marketplace.description-missing", replace(pluginDir+"/package.json", `"description":"Git skills",`, ``)},
		"claude manifest description drifts":            {"marketplace", "marketplace.description", set(pluginDir+"/.claude-plugin/plugin.json", pluginManifest("1.2.3", "Other"))},
		"codex manifest description drifts":             {"marketplace", "marketplace.description", set(pluginDir+"/.codex-plugin/plugin.json", pluginManifest("1.2.3", "Other"))},
		"claude marketplace description drifts":         {"marketplace", "marketplace.description", replace(claude, `"description":"Git skills"`, `"description":"Other"`)},
		"codex marketplace description drifts":          {"marketplace", "marketplace.description", replace(codex, `"description":"Git skills"`, `"description":"Other"`)},
		"plugin dependency is pinned elsewhere":         {"marketplace", "marketplace.dependency-version", replace(pluginDir+"/package.json", `"left-pad":"1.0.0"`, `"@xonovex/plugin-base":"1.0.0"`)},
		"plugin devDependency is pinned elsewhere":      {"marketplace", "marketplace.dependency-version", replace(pluginDir+"/package.json", `"dependencies":{"left-pad":"1.0.0"}`, `"devDependencies":{"@xonovex/plugin-base":"1.0.0"}`)},
		"plugin peerDependency is pinned elsewhere":     {"marketplace", "marketplace.dependency-version", replace(pluginDir+"/package.json", `"dependencies":{"left-pad":"1.0.0"}`, `"peerDependencies":{"@xonovex/plugin-base":"1.0.0"}`)},
		"plugin optionalDependency is pinned elsewhere": {"marketplace", "marketplace.dependency-version", replace(pluginDir+"/package.json", `"dependencies":{"left-pad":"1.0.0"}`, `"optionalDependencies":{"@xonovex/plugin-base":"1.0.0"}`)},
		"package.json is absent":                        {"marketplace", "marketplace.package-missing", remove(pluginDir + "/package.json")},
		"neutral manifest is absent":                    {"marketplace", "marketplace.package-missing", remove(pluginDir + "/plugin.json")},
		"claude manifest is absent":                     {"marketplace", "marketplace.package-missing", remove(pluginDir + "/.claude-plugin/plugin.json")},
		"codex manifest is absent":                      {"marketplace", "marketplace.package-missing", remove(pluginDir + "/.codex-plugin/plugin.json")},

		// Skill packaging.
		"plugin has no skills":              {"plugin", "plugin.no-skills", remove(pluginDir + "/skills/using-git/SKILL.md")},
		"claude manifest lists skills":      {"plugin", "plugin.skills-form", set(pluginDir+"/.claude-plugin/plugin.json", `{"name":"xonovex-git","version":"1.2.3","description":"Git skills","skills":["./skills/"]}`)},
		"codex manifest lists skills":       {"plugin", "plugin.skills-form", set(pluginDir+"/.codex-plugin/plugin.json", `{"name":"xonovex-git","version":"1.2.3","description":"Git skills","skills":["./skills/"]}`)},
		"manifest skill path is not direct": {"plugin", "plugin.skills-path", set(pluginDir+"/.claude-plugin/plugin.json", `{"name":"xonovex-git","version":"1.2.3","description":"Git skills","skills":"./other/"}`)},

		// The root README and the release workflow.
		"README link does not resolve": {"links", "links.missing-target", replace("README.md", pluginDir+"/README.md", "gone/README.md")},
		"README overclaims":            {"repository", "repository.readme-claim", set("README.md", "# P\n\nAll harnesses support every rule.\n")},
		"workflow publishes from tags": {"repository", "repository.workflow-tags", replace(".github/workflows/release.yml", "on:\n", "on:\n  push:\n")},
		"workflow publishes by input":  {"repository", "repository.workflow-dispatch-input", replace(".github/workflows/release.yml", "run: publish", "run: publish\n        env:\n          X: ${{ inputs.dry_run != true }}")},
		"publish step is not PR only":  {"repository", "repository.workflow-publish", replace(".github/workflows/release.yml", "        if: github.event_name == 'pull_request'\n", "")},
		"dry run is not dispatch only": {"repository", "repository.workflow-dry-run", replace(".github/workflows/release.yml", "        if: github.event_name == 'workflow_dispatch'\n", "")},
		"publish borrows a later job":  {"repository", "repository.workflow-publish", set(".github/workflows/release.yml", "name: R\njobs:\n  a:\n    steps:\n      - name: Publish\n        run: publish\n  b:\n    steps:\n      - name: x\n        if: github.event_name == 'pull_request'\n        run: y\n")},

		// Instruction files.
		"AGENTS.md has no CLAUDE.md":         {"repository", "repository.claude-missing", remove("packages/tooling/CLAUDE.md")},
		"CLAUDE.md does not point at AGENTS": {"repository", "repository.claude-pointer", set("CLAUDE.md", "Nothing")},
		"moon directory is unpaired":         {"repository", "repository.claude-missing", remove(".moon/CLAUDE.md")},
		"package group is undocumented": {"repository", "repository.group-undocumented", func(files map[string]string, _ *checktest.Repo) {
			delete(files, "packages/tooling/AGENTS.md")
			delete(files, "packages/tooling/CLAUDE.md")
		}},

		// Moon task files.
		"no tag task files": {"workspace", "workspace.tag-files-missing", func(files map[string]string, _ *checktest.Repo) {
			delete(files, ".moon/tasks/all.yml")
			delete(files, ".moon/tasks/tag-typescript.yml")
		}},
		"no project files": {"workspace", "workspace.project-files-missing", func(files map[string]string, _ *checktest.Repo) {
			delete(files, toolDir+"/moon.yml")
			delete(files, libDir+"/moon.yml")
		}},
		"tag file drops inherited deps":               {"workspace", "workspace.tasks-dropped", replace(".moon/tasks/tag-typescript.yml", "[build, test, lint]", "[lint]")},
		"tag file extends a missing one":              {"workspace", "workspace.tasks-extends", replace(".moon/tasks/tag-typescript.yml", "all.yml", "gone.yml")},
		"typescript project keeps the template floor": {"workspace", "workspace.coverage-floor", replace(libDir+"/moon.yml", "      TS_COVERAGE_MIN_LINES: 80\n", "")},
		"dependency source is not an input":           {"workspace", "workspace.sources-missing", replace(toolDir+"/moon.yml", "    - /"+libDir+"/src/**/*\n", "    - src/**/*\n")},
		"dependency source nothing reaches":           {"workspace", "workspace.sources-extra", replace(toolDir+"/moon.yml", "    - /"+libDir+"/src/**/*\n", "    - /"+libDir+"/src/**/*\n    - /packages/ghost/src/**/*\n")},

		// The hasher and the bins.
		"ignored directory no pattern covers": {"workspace", "workspace.hasher-uncovered", func(_ map[string]string, repo *checktest.Repo) {
			repo.Ignored = append(repo.Ignored, "tmp/")
		}},
		"ignore pattern drops a declared input": {"workspace", "workspace.hasher-input", replace(".moon/workspace.yml", "ignorePatterns:\n", "ignorePatterns:\n    - '"+toolDir+"/bin.js'\n")},
		"git cannot list ignored paths":         {"workspace", "workspace.hasher-git", repoFault},
		"private bin target is not committed":   {"workspace", "workspace.bin-target", func(_ map[string]string, repo *checktest.Repo) { repo.Tracked = nil }},
		"published bin target is not packed": {"workspace", "workspace.bin-target", func(files map[string]string, repo *checktest.Repo) {
			files[toolDir+"/package.json"] = `{"name":"@x/tool","files":["lib/"],"bin":{"tool":"./bin.js"}}`
			repo.Tracked = nil
		}},
		"git cannot list committed files": {"workspace", "workspace.bin-git", repoFault},
	}
	for name, m := range mutations {
		t.Run(name, func(t *testing.T) {
			files, repo := releaseFixture()
			m.apply(files, &repo)
			if got := runRelease(t, m.check, files, repo); !checktest.Has(got, m.code) {
				t.Errorf("%s did not report %s: %v", m.check, m.code, checktest.Codes(got))
			}
		})
	}
}
