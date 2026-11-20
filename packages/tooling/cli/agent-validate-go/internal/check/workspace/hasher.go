package workspace

import (
	"path"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/glob"
)

// hasher.walkStrategy glob walks the file system and does not read .gitignore,
// so hasher.ignorePatterns restates the exclusions git makes. The two lists
// drift silently: a newly ignored directory starts being hashed, and a pattern
// added to the list starts excluding a declared input. Both directions are
// reported.

type workspaceFile struct {
	Hasher struct {
		WalkStrategy   string   `yaml:"walkStrategy"`
		IgnorePatterns []string `yaml:"ignorePatterns"`
	} `yaml:"hasher"`
}

type inputProject struct {
	Tasks map[string]struct {
		Inputs []string `yaml:"inputs"`
	} `yaml:"tasks"`
}

type compiled struct {
	pattern string
	regex   *regexp.Regexp
}

func matchesAny(p string, all []compiled) bool {
	for _, c := range all {
		if c.regex.MatchString(p) {
			return true
		}
	}
	return false
}

// A directory that git reports as dist/ has to match a pattern written for the
// files beneath it, so the pattern is compared against probe paths inside it.
func probes(ignored string) []string {
	if strings.HasSuffix(ignored, "/") {
		return []string{ignored + "probe", ignored + "nested/probe"}
	}
	return []string{ignored}
}

func isLiteral(input string) bool { return !strings.ContainsAny(input, "*?[]{}") }

type declared struct{ source, path string }

// declaredInputs resolves the inputs that a project declares as plain paths to
// workspace-relative form. A glob is skipped: comparing it with an ignore glob
// would need both resolved against the tree.
func (r *run) declaredInputs() []declared {
	var out []declared
	for _, p := range r.projects {
		var parsed inputProject
		if yaml.Unmarshal([]byte(p.text), &parsed) != nil {
			continue
		}
		for _, t := range parsed.Tasks {
			for _, entry := range t.Inputs {
				if !isLiteral(entry) || strings.HasPrefix(entry, "$") {
					continue
				}
				resolved := path.Join(p.dir, entry)
				if strings.HasPrefix(entry, "/") {
					resolved = strings.TrimPrefix(entry, "/")
				}
				out = append(out, declared{source: p.file, path: resolved})
			}
		}
	}
	return out
}

func (r *run) hasher() error {
	text, found, err := r.ctx.Tree.Optional(r.options.Workspace)
	if err != nil || !found {
		return err
	}
	var ws workspaceFile
	if !r.parseYAML("workspace.hasher", r.options.Workspace, r.options.Workspace+" hasher section", text, &ws) {
		return nil
	}
	if ws.Hasher.WalkStrategy != "glob" {
		return nil
	}
	if r.ctx.Repository == nil {
		r.report.Fail("workspace.hasher-git", r.options.Workspace, 1, "git could not list ignored paths for the hasher check")
		return nil
	}
	ignored, err := r.ctx.Repository.IgnoredPaths()
	if err != nil {
		r.report.Fail("workspace.hasher-git", r.options.Workspace, 1, "git could not list ignored paths for the hasher check: %v", err)
		return nil
	}
	patterns := make([]compiled, 0, len(ws.Hasher.IgnorePatterns))
	for _, pattern := range ws.Hasher.IgnorePatterns {
		regex, err := glob.Compile(pattern)
		if err != nil {
			r.report.Fail("workspace.hasher-pattern", r.options.Workspace, 1, "hasher.ignorePatterns entry %q is not a valid glob: %v", pattern, err)
			continue
		}
		patterns = append(patterns, compiled{pattern: pattern, regex: regex})
	}
	// A dist directory stays hashable: the bin-permissions tasks declare
	// dist/src/index.js as an input, and an ignore pattern drops declared
	// inputs as well as walked ones.
	var exempt []string
	for _, p := range ignored {
		if strings.HasSuffix(p, "dist/") {
			exempt = append(exempt, p)
		}
	}
	for _, p := range ignored {
		if hasPrefixAny(p, exempt) || allMatch(probes(p), patterns) {
			continue
		}
		r.report.Fail("workspace.hasher-uncovered", r.options.Workspace, 1,
			"%s is ignored by git but no hasher.ignorePatterns entry covers it: the glob walk will hash it whenever a task input glob reaches it", p)
	}
	for _, d := range r.declaredInputs() {
		var matched []string
		for _, c := range patterns {
			if c.regex.MatchString(d.path) {
				matched = append(matched, c.pattern)
			}
		}
		if len(matched) > 0 {
			r.report.Fail("workspace.hasher-input", d.source, 1,
				"%s declares input %s, which hasher.ignorePatterns excludes via %s: an ignore pattern drops declared inputs as well as walked ones, so the task stops re-running when that file changes",
				d.source, d.path, strings.Join(matched, ", "))
		}
	}
	return nil
}

func hasPrefixAny(value string, prefixes []string) bool {
	for _, p := range prefixes {
		if strings.HasPrefix(value, p) {
			return true
		}
	}
	return false
}

func allMatch(paths []string, all []compiled) bool {
	for _, p := range paths {
		if !matchesAny(p, all) {
			return false
		}
	}
	return true
}
