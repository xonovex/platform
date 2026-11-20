// Package workspace checks the Moon workspace of a repository: that tag task
// files and project files exist, task inheritance across tag files, coverage
// floors, dependency source file groups, the hasher ignore patterns and the
// targets of package bins.
package workspace

import (
	"path"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/findings"

	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/check"
)

// Name is the name of the check.
const Name = "workspace"

// Options tune the check.
type Options struct {
	// TasksDir holds the tag task files.
	TasksDir string `json:"tasksDir"`
	// Workspace is the Moon workspace file.
	Workspace string `json:"workspace"`
	// Rules names the rules that run. An empty list runs all of them.
	Rules []string `json:"rules"`
}

func defaults() Options {
	return Options{TasksDir: ".moon/tasks", Workspace: ".moon/workspace.yml"}
}

var allRules = []string{"task-files", "task-inheritance", "coverage-floors", "dependency-sources", "hasher", "bin-targets"}

// Check is the workspace check.
type Check struct{}

// Name returns the name of the check.
func (Check) Name() string { return Name }

// Summary describes the check.
func (Check) Summary() string {
	return "Moon task files, task inheritance, coverage floors, dependency sources, hasher ignores and bin targets"
}

// DefaultPaths returns the package tree root.
func (Check) DefaultPaths() []string { return []string{"packages"} }

// project is one moon.yml with what the rules read beside it.
type project struct {
	id          string
	dir         string
	file        string
	text        string
	packageName string
	configTexts []string
	// moon is true when the directory holds a moon.yml.
	moon bool
}

type run struct {
	ctx      check.Context
	report   *check.Reporter
	options  Options
	projects []project
}

// configFiles are the files in which a project names a config package.
var configFiles = []string{
	"eslint.config.ts", "eslint.config.js", "prettier.config.ts", "prettier.config.js",
	"vitest.config.ts", "vite.config.ts", "tsconfig.json", "tsconfig.build.json", "tsconfig.test.json",
}

// Run checks the workspace.
func (Check) Run(ctx check.Context) ([]findings.Finding, error) {
	options := defaults()
	if err := check.DecodeOptions(ctx.Options, &options); err != nil {
		return nil, err
	}
	r := &run{ctx: ctx, report: check.NewReporter(Name), options: options}
	for _, root := range ctx.Paths {
		if err := r.loadProjects(root); err != nil {
			return nil, err
		}
	}
	rules := options.Rules
	if len(rules) == 0 {
		rules = allRules
	}
	steps := map[string]func() error{
		"task-files":         r.taskFiles,
		"task-inheritance":   r.taskInheritance,
		"coverage-floors":    r.coverageFloors,
		"dependency-sources": r.dependencySources,
		"hasher":             r.hasher,
		"bin-targets":        r.binTargets,
	}
	for _, rule := range rules {
		step, ok := steps[rule]
		if !ok {
			return nil, &unknownRuleError{rule}
		}
		if err := step(); err != nil {
			return nil, err
		}
	}
	return r.report.Findings(), nil
}

type unknownRuleError struct{ rule string }

func (e *unknownRuleError) Error() string {
	return "unknown workspace rule " + e.rule + "; known rules: " + strings.Join(allRules, ", ")
}

// loadProjects finds every moon.yml: a package is a directory with a
// package.json or a moon.yml and is not entered further, except that its
// skills/ directories may be projects too.
func (r *run) loadProjects(root string) error {
	var visit func(dir string) error
	visit = func(dir string) error {
		isPackage := r.ctx.Tree.IsFile(path.Join(dir, "package.json")) || r.ctx.Tree.IsFile(path.Join(dir, "moon.yml"))
		if isPackage {
			if err := r.addProject(dir); err != nil {
				return err
			}
			skills, err := r.ctx.Tree.List(path.Join(dir, "skills"))
			if err != nil {
				return err
			}
			for _, e := range skills {
				if e.IsDir && r.ctx.Tree.IsFile(path.Join(dir, "skills", e.Name, "SKILL.md")) {
					if err := r.addProject(path.Join(dir, "skills", e.Name)); err != nil {
						return err
					}
				}
			}
			return nil
		}
		entries, err := r.ctx.Tree.List(dir)
		if err != nil {
			return err
		}
		for _, e := range entries {
			if e.IsDir && e.Name != "node_modules" && !strings.HasPrefix(e.Name, ".") {
				if err := visit(path.Join(dir, e.Name)); err != nil {
					return err
				}
			}
		}
		return nil
	}
	if !r.ctx.Tree.IsDir(root) {
		return nil
	}
	if err := visit(root); err != nil {
		return err
	}
	sort.Slice(r.projects, func(i, j int) bool { return r.projects[i].file < r.projects[j].file })
	return nil
}

func (r *run) addProject(dir string) error {
	file := path.Join(dir, "moon.yml")
	text, hasMoon, err := r.ctx.Tree.Optional(file)
	if err != nil {
		return err
	}
	p := project{id: path.Base(dir), dir: dir, file: file, text: text, moon: hasMoon}
	if raw, found, err := r.ctx.Tree.Optional(path.Join(dir, "package.json")); err != nil {
		return err
	} else if found {
		p.packageName = jsonName(raw)
	}
	for _, name := range configFiles {
		if raw, found, err := r.ctx.Tree.Optional(path.Join(dir, name)); err != nil {
			return err
		} else if found {
			p.configTexts = append(p.configTexts, raw)
		}
	}
	r.projects = append(r.projects, p)
	return nil
}

// parseYAML reports a syntax error as a finding and a shape error as
// an unreadable definition, and decodes into out.
func (r *run) parseYAML(code, file, label, text string, out any) bool {
	var syntax any
	if err := yaml.Unmarshal([]byte(text), &syntax); err != nil {
		r.report.Fail(code+"-yaml", file, 1, "%s is not valid YAML: %v", label, err)
		return false
	}
	if err := yaml.Unmarshal([]byte(text), out); err != nil {
		r.report.Fail(code+"-shape", file, 1, "%s has an unreadable definition", label)
		return false
	}
	return true
}
