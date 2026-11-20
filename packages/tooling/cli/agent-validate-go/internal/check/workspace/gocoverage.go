package workspace

import (
	"fmt"
	"path"
	"sort"
	"strings"
)

// goCoverageTask is the task whose environment holds the coverage floors that
// .moon/scripts/check-go-package-coverage.sh enforces.
const goCoverageTask = "go-test-coverage"

const (
	moduleFloorVariable   = "GO_COVERAGE_MIN"
	packageFloorsVariable = "GO_PACKAGE_FLOORS"
	exemptFloor           = "exempt"
)

type goCoverageProject struct {
	Language string `yaml:"language"`
	Tasks    map[string]struct {
		Env map[string]any `yaml:"env"`
	} `yaml:"tasks"`
}

// goPackages are the package directories of a module, named as the floor list
// names them: "." for the module root and "./<dir>" below it.
type goPackages struct {
	// existing holds every directory with a Go file, test files included.
	existing map[string]bool
	// code holds every directory with a Go file that is not a test file.
	code map[string]bool
}

// goCoverageFloors reports a Go module whose coverage task declares no module
// floor or no package floors, a floor the coverage script refuses, a package
// the floors name that the module does not hold, and a package with Go code
// that the floors miss. The coverage script fails on a missed package only
// when it runs, and it never notices an entry that names nothing, so a stale
// entry would outlive the package it was written for.
func (r *run) goCoverageFloors() error {
	for _, p := range r.projects {
		if !r.ctx.Tree.IsFile(path.Join(p.dir, "go.mod")) {
			continue
		}
		var parsed goCoverageProject
		if !r.parseYAML("workspace.go-coverage", p.file, p.file, p.text, &parsed) || parsed.Language != "go" {
			continue
		}
		env := parsed.Tasks[goCoverageTask].Env
		minimum, floors := envText(env[moduleFloorVariable]), envText(env[packageFloorsVariable])
		var missing []string
		if minimum == "" {
			missing = append(missing, moduleFloorVariable)
		}
		if floors == "" {
			missing = append(missing, packageFloorsVariable)
		}
		if len(missing) > 0 {
			r.report.Fail("workspace.go-coverage-floor", p.file, 1,
				"%s task '%s' sets no %s: a Go module declares the floor of the module and of every package it builds",
				p.file, goCoverageTask, strings.Join(missing, ", "))
		}
		if minimum != "" && !isFloor(minimum) {
			r.report.Fail("workspace.go-coverage-invalid", p.file, 1,
				"%s %s is %q: the module floor must be a number", p.file, moduleFloorVariable, minimum)
		}
		if floors == "" {
			continue
		}
		packages, err := r.goPackages(p.dir)
		if err != nil {
			return err
		}
		r.checkPackageFloors(p.file, floors, packages)
	}
	return nil
}

// checkPackageFloors reports each entry of the floor list that the coverage
// script refuses or that names no package, and each package with Go code that
// no entry names. An exempt package counts as named.
func (r *run) checkPackageFloors(file, floors string, packages goPackages) {
	named := map[string]bool{}
	var unknown []string
	for _, entry := range strings.Fields(floors) {
		pkg, floor, ok := strings.Cut(entry, "=")
		if !ok || pkg == "" || (pkg != "." && !strings.HasPrefix(pkg, "./")) || (floor != exemptFloor && !isFloor(floor)) {
			r.report.Fail("workspace.go-coverage-invalid", file, 1,
				"%s %s entry %q is not ./PACKAGE=MINIMUM or ./PACKAGE=%s", file, packageFloorsVariable, entry, exemptFloor)
			continue
		}
		named[pkg] = true
		if !packages.existing[pkg] {
			unknown = append(unknown, pkg)
		}
	}
	if len(unknown) > 0 {
		r.report.Fail("workspace.go-coverage-unknown", file, 1,
			"%s %s names %s, which the module does not hold: drop the entry or correct its path",
			file, packageFloorsVariable, strings.Join(unknown, ", "))
	}
	var undeclared []string
	for pkg := range packages.code {
		if !named[pkg] {
			undeclared = append(undeclared, pkg)
		}
	}
	sort.Strings(undeclared)
	if len(undeclared) > 0 {
		r.report.Fail("workspace.go-coverage-undeclared", file, 1,
			"%s %s misses %s, which hold Go code: add PACKAGE=MINIMUM or PACKAGE=%s for each",
			file, packageFloorsVariable, strings.Join(undeclared, ", "), exemptFloor)
	}
}

// goPackages finds the package directories of the module at dir. It skips what
// the go command skips in a ./... pattern: directories whose name starts with
// "." or "_", testdata, vendor and nested modules, and it skips node_modules,
// which holds no package of the module.
func (r *run) goPackages(dir string) (goPackages, error) {
	packages := goPackages{existing: map[string]bool{}, code: map[string]bool{}}
	var visit func(rel string) error
	visit = func(rel string) error {
		entries, err := r.ctx.Tree.List(path.Join(dir, rel))
		if err != nil {
			return err
		}
		name := "."
		if rel != "" {
			name = "./" + rel
		}
		for _, e := range entries {
			switch {
			case e.IsDir && !skipGoDir(e.Name):
				child := path.Join(rel, e.Name)
				if r.ctx.Tree.IsFile(path.Join(dir, child, "go.mod")) {
					continue
				}
				if err := visit(child); err != nil {
					return err
				}
			case !e.IsDir && strings.HasSuffix(e.Name, ".go") && !skipGoName(e.Name):
				packages.existing[name] = true
				if !strings.HasSuffix(e.Name, "_test.go") {
					packages.code[name] = true
				}
			}
		}
		return nil
	}
	return packages, visit("")
}

func skipGoDir(name string) bool {
	return skipGoName(name) || name == "testdata" || name == "vendor" || name == "node_modules"
}

// skipGoName reports a file or directory name the go command ignores.
func skipGoName(name string) bool {
	return strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_")
}

// isFloor reports whether value is a floor as the coverage script accepts it:
// digits and dots only.
func isFloor(value string) bool {
	return value != "" && strings.Trim(value, "0123456789.") == ""
}

// envText returns a task environment value as the shell sees it: a YAML number
// is its decimal text, and a missing value is empty.
func envText(value any) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprint(value))
}
