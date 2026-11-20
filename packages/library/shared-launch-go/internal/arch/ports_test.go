//go:build arch

package arch

import (
	"go/ast"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-launch-go/pkg/credential"
	"github.com/xonovex/platform/packages/library/shared-launch-go/pkg/isolation"
	"github.com/xonovex/platform/packages/library/shared-launch-go/pkg/sandbox"
)

const agentLibrary = "github.com/xonovex/platform/packages/library/shared-agent-go/pkg/"

// selectors calls fn for every package-qualified selector of the file, with the
// import path the qualifier names.
func selectors(f sourceFile, fn func(path, name string)) {
	alias := map[string]string{}
	for _, spec := range f.parsed.Imports {
		path, _ := strconv.Unquote(spec.Path.Value)
		name := path[strings.LastIndexByte(path, '/')+1:]
		if spec.Name != nil {
			name = spec.Name.Name
		}
		alias[name] = path
	}
	ast.Inspect(f.parsed, func(node ast.Node) bool {
		if selector, ok := node.(*ast.SelectorExpr); ok {
			if pkg, ok := selector.X.(*ast.Ident); ok && alias[pkg.Name] != "" {
				fn(alias[pkg.Name], selector.Sel.Name)
			}
		}
		return true
	})
}

// TestSelectionNamesNoVariant keeps the selection and policy gate on the
// capabilities a variant declares, so a new variant is one registration.
func TestSelectionNamesNoVariant(t *testing.T) {
	variants := map[string][]string{
		agentLibrary + "isolation": {"IsolationNone", "IsolationBwrap", "IsolationDocker", "IsolationSeatbelt", "IsolationDockerSandboxes", "IsolationOpenShell"},
		agentLibrary + "provision": {"ProvisionNone", "ProvisionNix", "ProvisionCommand"},
		agentLibrary + "network":   {"ModeHost", "ModeNone", "ModeManaged"},
	}
	for _, f := range nonTest(moduleFiles(t)) {
		if f.dir != "pkg/sandbox" {
			continue
		}
		selectors(f, func(path, name string) {
			if slices.Contains(variants[path], name) {
				t.Errorf("%s names the variant %s; gate on the declared capabilities", f.path, name)
			}
		})
	}
}

// agentSurface are the names of the agent types package that name no harness.
var agentSurface = map[string]bool{"AgentType": true, "AgentConfig": true, "AgentExecOptions": true}

// TestNoHarnessIsNamed keeps the core on the facts a harness declares.
func TestNoHarnessIsNamed(t *testing.T) {
	for _, f := range nonTest(moduleFiles(t)) {
		selectors(f, func(path, name string) {
			if path == agentLibrary+"types" && strings.HasPrefix(name, "Agent") && !agentSurface[name] {
				t.Errorf("%s names the harness %s; compare the declared facts instead", f.path, name)
			}
		})
	}
}

// TestRegistryFactoriesAreLazy keeps a registry from building a variant that a
// run does not select.
func TestRegistryFactoriesAreLazy(t *testing.T) {
	kinds := map[string]reflect.Type{
		"sandbox.IsolatorFactory":    reflect.TypeFor[sandbox.IsolatorFactory[isolation.Confinement]](),
		"sandbox.ProvisionerFactory": reflect.TypeFor[sandbox.ProvisionerFactory](),
		"credential.Factory":         reflect.TypeFor[credential.Factory](),
	}
	for name, kind := range kinds {
		if kind.Kind() != reflect.Func {
			t.Errorf("%s is a %s; register a function that builds the variant", name, kind.Kind())
		}
	}
}

// TestCredentialValuesStayInsideSecret keeps a credential value out of every
// field and log: only the credential core reveals a Secret.
func TestCredentialValuesStayInsideSecret(t *testing.T) {
	secret := reflect.TypeFor[credential.Secret]()
	for i := range secret.NumField() {
		if secret.Field(i).IsExported() {
			t.Errorf("Secret exports the field %s", secret.Field(i).Name)
		}
	}
	for _, f := range nonTest(moduleFiles(t)) {
		if f.dir == "pkg/credential" {
			continue
		}
		ast.Inspect(f.parsed, func(node ast.Node) bool {
			if selector, ok := node.(*ast.SelectorExpr); ok && selector.Sel.Name == "Reveal" {
				t.Errorf("%s calls Reveal; hand the Secret on to the executor", f.path)
			}
			return true
		})
	}
}

// TestConfinementCannotStartAProcess keeps the declaration of a sandbox apart
// from the launcher that runs it on one machine.
func TestConfinementCannotStartAProcess(t *testing.T) {
	confinement := reflect.TypeFor[isolation.Confinement]()
	launcher := reflect.TypeFor[isolation.Launcher]()
	if launcher.NumMethod() == 0 {
		t.Fatal("Launcher declares no method")
	}
	for i := range launcher.NumMethod() {
		if _, found := confinement.MethodByName(launcher.Method(i).Name); found {
			t.Errorf("Confinement declares the launcher method %s", launcher.Method(i).Name)
		}
	}
	for _, name := range []string{"Run", "Command", "TerminalCommand", "CheckAgent", "Available"} {
		if _, found := confinement.MethodByName(name); found {
			t.Errorf("Confinement declares %s", name)
		}
	}
}

// specConcerns are the packages the planner and its Spec must not reach: the
// run record of an executor, and the workspace and configuration concerns of a
// consumer.
var specConcerns = []string{
	module + "/pkg/record",
	agentLibrary + "workspace",
	agentLibrary + "agentcmd",
	"github.com/xonovex/platform/packages/library/shared-config-go",
	"github.com/xonovex/platform/packages/tooling",
	"github.com/xonovex/platform/packages/sandbox",
	"github.com/spf13/cobra",
}

func TestSpecReachesNoForbiddenPackage(t *testing.T) {
	for _, f := range moduleFiles(t) {
		if f.dir != "pkg/launch" {
			continue
		}
		for _, imp := range f.imports {
			for _, concern := range specConcerns {
				if imp == concern || strings.HasPrefix(imp, concern+"/") {
					t.Errorf("%s imports %s", f.path, imp)
				}
			}
		}
	}
}

// TestIsolationBuildsNoAgentCommand keeps the agent command line and its
// provider environment in the planner: an isolation, and the helpers its leaves
// share in hostlaunch, confine the planned process.
func TestIsolationBuildsNoAgentCommand(t *testing.T) {
	builders := map[string][]string{
		agentLibrary + "agents":    {"BuildArgs", "BuildEnv"},
		agentLibrary + "providers": {"GetProviderCliArgs"},
		agentLibrary + "agentcmd":  {"PlanProviderEnv"},
	}
	for _, f := range nonTest(moduleFiles(t)) {
		if f.dir != "pkg/isolation" && f.dir != "pkg/hostlaunch" && !strings.HasPrefix(f.dir, "pkg/isolation/") {
			continue
		}
		for _, imp := range f.imports {
			if imp == agentLibrary+"agentcmd" {
				t.Errorf("%s imports the agent command builder %s", f.path, imp)
			}
		}
		selectors(f, func(path, name string) {
			if slices.Contains(builders[path], name) {
				t.Errorf("%s calls %s; confine the planned process instead", f.path, name)
			}
		})
	}
}
