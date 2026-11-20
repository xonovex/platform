package workspace

import (
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// dependsOn entries are a project id or a mapping that names one.
type dependsOn struct{ id string }

func (d *dependsOn) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind == yaml.ScalarNode {
		return node.Decode(&d.id)
	}
	var m struct {
		ID string `yaml:"id"`
	}
	if err := node.Decode(&m); err != nil {
		return err
	}
	d.id = m.ID
	return nil
}

type sourceProject struct {
	Tags       []string            `yaml:"tags"`
	DependsOn  []dependsOn         `yaml:"dependsOn"`
	FileGroups map[string][]string `yaml:"fileGroups"`
}

var packageReference = regexp.MustCompile(`@[a-z0-9-]+/[a-z0-9-]+`)

// A project carrying any of these inherits the TypeScript task templates, whose
// tasks read the source of the packages the project depends on.
var typescriptTags = map[string]bool{
	"typescript": true, "typescript-acceptance": true, "typescript-config": true,
	"typescript-integration": true, "typescript-script": true,
}

// Another TypeScript package is read through its exports map and a tsconfig
// package through the tsconfig that extends it. A Go dependency is read by the
// Go tasks, never by these.
var readableTags = map[string]bool{
	"typescript": true, "typescript-acceptance": true, "typescript-config": true,
	"typescript-integration": true, "typescript-script": true, "tsconfig": true,
}

const dependencySources = "dependencySources"

// Every internal dependency publishes its types from src, so a project's tasks
// read the source of the packages it depends on. Moon does not fold a
// dependency task's hash into its dependent, so a project that leaves one out
// serves a cached pass over code its dependency has broken. This reports a file
// group that has drifted, in either direction, from the transitive dependsOn
// closure together with the config packages the project names.
func (r *run) dependencySources() error {
	parsed := map[string]sourceProject{}
	byID := map[string]project{}
	idByPackage := map[string]string{}
	for _, p := range r.projects {
		if p.packageName != "" {
			idByPackage[p.packageName] = p.id
		}
	}
	for _, p := range r.projects {
		var sp sourceProject
		if !r.parseYAML("workspace.sources", p.file, p.id, p.text, &sp) {
			continue
		}
		parsed[p.id] = sp
		byID[p.id] = p
	}
	var closure func(id string, seen map[string]bool)
	closure = func(id string, seen map[string]bool) {
		for _, d := range parsed[id].DependsOn {
			if seen[d.id] {
				continue
			}
			if _, ok := parsed[d.id]; !ok {
				continue
			}
			seen[d.id] = true
			closure(d.id, seen)
		}
	}
	configClosure := func(id string) map[string]bool {
		referenced := map[string]bool{}
		for _, text := range byID[id].configTexts {
			for _, ref := range packageReference.FindAllString(text, -1) {
				if target, ok := idByPackage[ref]; ok && target != id {
					referenced[target] = true
				}
			}
		}
		for target := range copyOf(referenced) {
			closure(target, referenced)
		}
		delete(referenced, id)
		return referenced
	}
	ids := make([]string, 0, len(parsed))
	for id := range parsed {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		project := parsed[id]
		typescript := false
		for _, tag := range project.Tags {
			typescript = typescript || typescriptTags[tag]
		}
		if !typescript {
			continue
		}
		reachable := map[string]bool{}
		closure(id, reachable)
		for target := range configClosure(id) {
			reachable[target] = true
		}
		expected := map[string]bool{}
		for dependency := range reachable {
			for _, tag := range parsed[dependency].Tags {
				if readableTags[tag] {
					expected["/"+byID[dependency].dir+"/src/**/*"] = true
					break
				}
			}
		}
		if len(expected) == 0 {
			continue
		}
		declared := map[string]bool{}
		for _, glob := range project.FileGroups[dependencySources] {
			declared[glob] = true
		}
		file := byID[id].file
		if missing := difference(expected, declared); len(missing) > 0 {
			r.report.Fail("workspace.sources-missing", file, 1,
				"%s file group '%s' is missing %s: a dependency whose source is not an input leaves this project's hash unchanged when that dependency changes, and moon serves a cached pass",
				id, dependencySources, strings.Join(missing, ", "))
		}
		var extra []string
		for _, glob := range difference(declared, expected) {
			if glob != "src/**/*" {
				extra = append(extra, glob)
			}
		}
		if len(extra) > 0 {
			r.report.Fail("workspace.sources-extra", file, 1,
				"%s file group '%s' declares %s, which neither dependsOn nor a config file reaches: drop it or add the dependency",
				id, dependencySources, strings.Join(extra, ", "))
		}
	}
	return nil
}

func copyOf(set map[string]bool) map[string]bool {
	out := make(map[string]bool, len(set))
	for k := range set {
		out[k] = true
	}
	return out
}

func difference(a, b map[string]bool) []string {
	var out []string
	for k := range a {
		if !b[k] {
			out = append(out, k)
		}
	}
	sort.Strings(out)
	return out
}
