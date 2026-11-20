package workspace

import (
	"encoding/json"
	"path"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

func jsonName(text string) string {
	var p struct {
		Name string `json:"name"`
	}
	_ = json.Unmarshal([]byte(text), &p)
	return p.Name
}

// dep is a task dependency written as a target string or as a mapping.
type dep struct{ target string }

func (d *dep) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind == yaml.ScalarNode {
		return node.Decode(&d.target)
	}
	var m struct {
		Target string `yaml:"target"`
	}
	if err := node.Decode(&m); err != nil {
		return err
	}
	d.target = m.Target
	return nil
}

type task struct {
	Deps    *[]dep `yaml:"deps"`
	Options struct {
		MergeDeps string `yaml:"mergeDeps"`
	} `yaml:"options"`
}

type tagFile struct {
	Extends string          `yaml:"extends"`
	Tasks   map[string]task `yaml:"tasks"`
}

func targets(deps []dep) []string {
	out := make([]string, len(deps))
	for i, d := range deps {
		out[i] = d.target
	}
	return out
}

// inherited returns the deps each task carries once extends is applied. A task
// that the child redefines with its own deps supersedes the parent's entry.
func inherited(name string, files map[string]tagFile, visiting map[string]bool) map[string][]string {
	file, ok := files[name]
	deps := map[string][]string{}
	if !ok || visiting[name] {
		return deps
	}
	if file.Extends != "" {
		next := map[string]bool{name: true}
		for k := range visiting {
			next[k] = true
		}
		deps = inherited(path.Base(file.Extends), files, next)
	}
	for taskName, t := range file.Tasks {
		if t.Deps != nil {
			deps[taskName] = targets(*t.Deps)
		}
	}
	return deps
}

// A workspace with no tag task files or no project files has nothing for the
// other rules to read, so it fails here rather than passing them vacuously.
func (r *run) taskFiles() error {
	entries, err := r.ctx.Tree.List(r.options.TasksDir)
	if err != nil {
		return err
	}
	tagFiles := 0
	for _, e := range entries {
		if !e.IsDir && strings.HasSuffix(e.Name, ".yml") {
			tagFiles++
		}
	}
	if tagFiles == 0 {
		r.report.Fail("workspace.tag-files-missing", r.options.TasksDir, 1, "%s must define tag task files", r.options.TasksDir)
	}
	projects := 0
	for _, p := range r.projects {
		if p.moon {
			projects++
		}
	}
	if projects == 0 {
		r.report.Fail("workspace.project-files-missing", strings.Join(r.ctx.Paths, ","), 1, "packages must define project task files")
	}
	return nil
}

// A tag file that redefines an inherited task with fewer deps than its parent
// declares drops part of the gate: Moon replaces the parent's deps across
// extends, and mergeDeps does not change that.
func (r *run) taskInheritance() error {
	entries, err := r.ctx.Tree.List(r.options.TasksDir)
	if err != nil {
		return err
	}
	files := map[string]tagFile{}
	var names []string
	for _, e := range entries {
		if e.IsDir || !strings.HasSuffix(e.Name, ".yml") {
			continue
		}
		file := path.Join(r.options.TasksDir, e.Name)
		text, err := r.ctx.Tree.Read(file)
		if err != nil {
			return err
		}
		var parsed tagFile
		if r.parseYAML("workspace.tasks", file, e.Name, text, &parsed) {
			files[e.Name] = parsed
			names = append(names, e.Name)
		}
	}
	sort.Strings(names)
	for _, name := range names {
		file := files[name]
		if file.Extends == "" {
			continue
		}
		parent := path.Base(file.Extends)
		if _, ok := files[parent]; !ok {
			r.report.Fail("workspace.tasks-extends", path.Join(r.options.TasksDir, name), 1, "%s extends missing %s", name, file.Extends)
			continue
		}
		parentDeps := inherited(parent, files, map[string]bool{name: true})
		taskNames := make([]string, 0, len(file.Tasks))
		for t := range file.Tasks {
			taskNames = append(taskNames, t)
		}
		sort.Strings(taskNames)
		for _, taskName := range taskNames {
			t := file.Tasks[taskName]
			if t.Deps == nil || t.Options.MergeDeps == "replace" {
				continue
			}
			inheritedDeps, ok := parentDeps[taskName]
			if !ok {
				continue
			}
			declared := map[string]bool{}
			for _, d := range targets(*t.Deps) {
				declared[d] = true
			}
			var dropped []string
			for _, d := range inheritedDeps {
				if !declared[d] {
					dropped = append(dropped, d)
				}
			}
			if len(dropped) > 0 {
				r.report.Fail("workspace.tasks-dropped", path.Join(r.options.TasksDir, name), 1,
					"%s task '%s' drops inherited %s: a task redefined across extends replaces the parent's deps, so restate them or declare mergeDeps: replace",
					name, taskName, strings.Join(dropped, ", "))
			}
		}
	}
	return nil
}

var typescriptFloorTags = map[string]bool{"typescript": true, "typescript-script": true}

var requiredFloors = []string{"TS_COVERAGE_MIN_LINES", "TS_COVERAGE_MIN_FUNCTIONS", "TS_COVERAGE_MIN_BRANCHES", "TS_COVERAGE_MIN_STATEMENTS"}

type floorProject struct {
	Tags  []string `yaml:"tags"`
	Tasks map[string]struct {
		Env map[string]any `yaml:"env"`
	} `yaml:"tasks"`
}

// A TypeScript project declares its own coverage floors. The tag template's
// default is a permissive placeholder, so a project that does not override it
// coasts there.
func (r *run) coverageFloors() error {
	for _, p := range r.projects {
		var parsed floorProject
		if !r.parseYAML("workspace.coverage", p.file, p.file, p.text, &parsed) {
			continue
		}
		typescript := false
		for _, tag := range parsed.Tags {
			typescript = typescript || typescriptFloorTags[tag]
		}
		if !typescript {
			continue
		}
		env := parsed.Tasks["ts-coverage"].Env
		var missing []string
		for _, floor := range requiredFloors {
			if value, ok := env[floor]; !ok || value == nil || value == "" {
				missing = append(missing, floor)
			}
		}
		if len(missing) > 0 {
			r.report.Fail("workspace.coverage-floor", p.file, 1,
				"%s task 'ts-coverage' sets no %s: a TypeScript project must declare its own coverage floor rather than inherit the tag template's permissive default",
				p.file, strings.Join(missing, ", "))
		}
	}
	return nil
}
