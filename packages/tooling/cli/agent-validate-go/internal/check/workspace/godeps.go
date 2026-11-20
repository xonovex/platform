package workspace

import (
	"path"
	"strings"
)

// goModule is what the rule reads from a go.mod: the module path and every
// module it requires, direct or indirect.
type goModule struct {
	path     string
	requires []string
}

// parseGoMod reads the module and require directives. The rule needs no other
// directive, and a go.mod that the go tool accepts always parses.
func parseGoMod(text string) goModule {
	var module goModule
	block := false
	for _, line := range strings.Split(text, "\n") {
		line, _, _ = strings.Cut(line, "//")
		fields := strings.Fields(line)
		switch {
		case len(fields) == 0:
		case block && fields[0] == ")":
			block = false
		case block:
			module.requires = append(module.requires, fields[0])
		case fields[0] == "module" && len(fields) > 1:
			module.path = fields[1]
		case fields[0] == "require" && len(fields) > 1 && fields[1] == "(":
			block = true
		case fields[0] == "require" && len(fields) > 1:
			module.requires = append(module.requires, fields[1])
		}
	}
	return module
}

// goDependencies reports a Go project whose Moon dependsOn differs from the
// requires of its go.mod that name another project of the workspace. A missing
// edge lets Moon build or test the project before the library it compiles
// against; an extra edge orders tasks behind a library the project never
// imports.
func (r *run) goDependencies() error {
	modules := map[string]string{}
	parsed := map[string]goModule{}
	for _, p := range r.projects {
		text, found, err := r.ctx.Tree.Optional(path.Join(p.dir, "go.mod"))
		if err != nil {
			return err
		}
		if !found {
			continue
		}
		module := parseGoMod(text)
		modules[module.path] = p.id
		parsed[p.id] = module
	}
	for _, p := range r.projects {
		module, ok := parsed[p.id]
		if !ok {
			continue
		}
		var sp sourceProject
		if !r.parseYAML("workspace.go-dependencies", p.file, p.id, p.text, &sp) {
			continue
		}
		required := map[string]bool{}
		for _, requirement := range module.requires {
			if id, internal := modules[requirement]; internal {
				required[id] = true
			}
		}
		declared := map[string]bool{}
		for _, d := range sp.DependsOn {
			if _, isGo := parsed[d.id]; isGo {
				declared[d.id] = true
			}
		}
		if missing := difference(required, declared); len(missing) > 0 {
			r.report.Fail("workspace.go-dependencies-missing", p.file, 1,
				"%s go.mod requires %s, which dependsOn does not name: Moon may build or test %s before the library it compiles against",
				p.id, strings.Join(missing, ", "), p.id)
		}
		if extra := difference(declared, required); len(extra) > 0 {
			r.report.Fail("workspace.go-dependencies-extra", p.file, 1,
				"%s dependsOn names %s, which go.mod does not require: drop the edge or require the module",
				p.id, strings.Join(extra, ", "))
		}
	}
	return nil
}
