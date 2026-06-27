---
type: plan
has_subplans: false
parent_plan: plans/agent-orthogonal-axis-reorg.md
parallel_group: 4
status: pending
dependencies:
  plans: [02-cli-isolation-provision-network.md, 03-cli-workspace-terminal.md, 04-cli-cmd-flags-wiring.md]
  files:
    - packages/agent/agent-cli-go/internal/architecture_test.go
    - packages/agent/agent-cli-go/moon.yml
skills_to_consult: [orthogonal-pattern-guide, hexagonal-pattern-guide, microkernel-pattern-guide, moon-guide]
validation:
  type_check: pending
  lint: pending
  build: pending
  tests: pending
  integration: pending
---

# CLI Architecture / Import Fitness Tests

## Objective

Lock the orthogonal-axis layout in place with executable fitness tests that fail the build if a leaf imports the composition root, a sibling leaf, or a cross-axis package in the wrong direction. The tests encode the hexagonal (ports in `shared/`, sole composition root), microkernel (lazy factory registry, method-agnostic policy gate), and orthogonal-pattern (per-axis `shared/` + bare leaves, glue in bridge files) invariants so the structure cannot silently regress. They run in CI via the existing moon `test` task for `agent-cli-go`.

## Tasks

1. **Add the import-direction test (leaves never reach the composition root).**
   File: `packages/agent/agent-cli-go/internal/architecture_test.go` (new). Load every package under `internal/isolation/<type>`, `internal/provision/<type>`, `internal/network/<type>` with `golang.org/x/tools/go/packages` (already an indirect dep) and assert none import `internal/sandbox`; only `internal/sandbox/plugins` is permitted to import the concrete leaves. Use a module-level helper, no globals.
   ```go
   func loadPkgs(t *testing.T, patterns ...string) []*packages.Package {
       cfg := &packages.Config{Mode: packages.NeedName | packages.NeedImports | packages.NeedModule}
       pkgs, err := packages.Load(cfg, patterns...)
       if err != nil { t.Fatalf("load: %v", err) }
       return pkgs
   }

   func TestLeavesDoNotImportCompositionRoot(t *testing.T) {
       const root = modPath + "/internal/sandbox"
       for _, p := range loadPkgs(t, modPath+"/internal/isolation/...", modPath+"/internal/provision/...", modPath+"/internal/network/...") {
           for imp := range p.Imports {
               if imp == root || (strings.HasPrefix(imp, root+"/") && imp != root+"/plugins") {
                   t.Errorf("%s imports composition root %s", p.PkgPath, imp)
               }
           }
       }
   }
   ```

2. **Add the no-sibling-reach test (no leaf imports a sibling leaf in the same axis).**
   File: same `internal/architecture_test.go`. For each axis assert `isolation/bwrap` does not import `isolation/docker` (and vice-versa); same for `provision/*` and `workspace/*` leaves. Cross-leaf knowledge must live in `shared/` or a bridge file, not leaf-to-leaf.
   ```go
   func TestLeavesDoNotImportSiblings(t *testing.T) {
       axes := []string{"isolation", "provision", "workspace", "network"}
       for _, axis := range axes {
           base := modPath + "/internal/" + axis + "/"
           for _, p := range loadPkgs(t, base+"...") {
               leaf := strings.TrimPrefix(p.PkgPath, base) // e.g. "bwrap"
               for imp := range p.Imports {
                   if sib, ok := strings.CutPrefix(imp, base); ok {
                       if sib != "shared" && leaf != "shared" && !strings.HasPrefix(sib, leaf) {
                           t.Errorf("%s reaches sibling leaf %s", p.PkgPath, imp)
                       }
                   }
               }
           }
       }
   }
   ```

3. **Add the cross-axis bridge-direction test (glue flows one way into `shared/`).**
   File: same. Assert the bridge files — `isolation/<type>/network.go` and `provision/nix` — may import `network/shared` and `provision/shared`, but `network/shared` and `provision/shared` import no isolation leaf. The dependent leaf owns the bridge; the shared core stays free of any concrete sibling-axis variant.
   ```go
   func TestSharedCoresImportNoLeaf(t *testing.T) {
       for _, p := range loadPkgs(t, modPath+"/internal/network/shared", modPath+"/internal/provision/shared", modPath+"/internal/isolation/shared") {
           for imp := range p.Imports {
               for _, axis := range []string{"isolation", "provision", "network", "workspace"} {
                   base := modPath + "/internal/" + axis + "/"
                   if leaf, ok := strings.CutPrefix(imp, base); ok && leaf != "shared" {
                       t.Errorf("shared core %s imports leaf %s", p.PkgPath, imp)
                   }
               }
           }
       }
   }
   ```

4. **Add the policy-gate-purity test (gate reads `Capabilities`, switches on no method literal).**
   File: same. Parse `shared-agent-go`'s `pkg/policy/policy.go` and the CLI call site with `go/parser`/`go/ast` and assert the AST contains no identifier reference to `IsolationMethod`, `ProvisionMethod`, or `NetworkMethod` enum literals in a `switch`/`if` — the gate consumes only the `Capabilities{Pinned, HostToolsUnreachable, EgressRestricted, KernelIsolated}` struct via `EnforcePolicy`.
   ```go
   func TestPolicyGateNamesNoMethod(t *testing.T) {
       banned := []string{"IsolationMethod", "ProvisionMethod", "NetworkMethod"}
       src := mustReadPolicySource(t) // CLI usage + the vendored pkg/policy path
       fset := token.NewFileSet()
       f, err := parser.ParseFile(fset, src.path, src.bytes, 0)
       if err != nil { t.Fatalf("parse: %v", err) }
       ast.Inspect(f, func(n ast.Node) bool {
           if id, ok := n.(*ast.Ident); ok {
               for _, b := range banned {
                   if id.Name == b { t.Errorf("policy gate references method enum %s", b) }
               }
           }
           return true
       })
   }
   ```

5. **Add the composition-root-only-concretes + registry-factory tests.**
   File: same. Assert the constructors `none.New`, `bwrap.New`, `docker.New`, `nix.New` are referenced ONLY from `internal/sandbox/plugins/plugins.go` (scan all `internal/...` non-test files), and assert the registry stores `func()`-typed factories rather than eager instances — confirm `Registry` is `map[Method]Factory` where `Factory` is a function type, so binding is lazy and fail-closed selection happens at `Select`.
   ```go
   func TestConcretesOnlyInPlugins(t *testing.T) {
       hits := grepConstructorRefs(t, modPath+"/internal", []string{"none.New", "bwrap.New", "docker.New", "nix.New"})
       for file := range hits {
           if !strings.HasSuffix(file, "internal/sandbox/plugins/plugins.go") {
               t.Errorf("concrete constructor referenced outside composition root: %s", file)
           }
       }
   }

   func TestRegistryFactoriesAreLazy(t *testing.T) {
       var reg sandbox.Registry           // map[Method]Factory
       var f sandbox.Factory = nil        // must be a func type
       _ = reg
       if reflect.TypeOf(f).Kind() != reflect.Func { t.Fatal("Factory must be func()-typed") }
   }
   ```

6. **Wire the tests into the moon `test` task and document each guard.**
   File: `packages/agent/agent-cli-go/moon.yml` (and the inherited `.moon/tasks/*.yml` if the Go `test` task already globs `*_test.go`, no change is needed beyond confirming the package is included). Confirm `architecture_test.go` lives in a package that the `agent-cli-go:test` task compiles, and add a short comment block at the top of the file mapping each test name to the principle it guards (import-direction → hexagonal composition root; no-sibling-reach → orthogonal bare leaves; bridge-direction → one-way cross-axis glue; policy-gate-purity → microkernel capability gate; concretes-only → sole composition root; registry-factory → microkernel lazy binding).

## Validation Steps

- `npx moon run agent-cli-go:typecheck`
- `npx moon run agent-cli-go:lint`
- `npx moon run agent-cli-go:build`
- `npx moon run agent-cli-go:test` (must include and pass the new `internal/architecture_test.go`)
- In `packages/agent/agent-cli-go`: `go build ./...` and `go test ./internal/...` pass.
- Integration: `go test ./test/integration/...` — confirm the fitness tests do not perturb the help-text tests touched in subplan 04 (no flag changes here, but the test task runs both).

## Success Criteria

- [ ] `internal/architecture_test.go` exists and is compiled by `agent-cli-go:test`.
- [ ] Import-direction test fails if any axis leaf imports `internal/sandbox` (except `plugins`).
- [ ] No-sibling-reach test fails on any leaf-to-leaf import within an axis.
- [ ] Bridge-direction test confirms `network/shared` and `provision/shared` import no isolation leaf.
- [ ] Policy-gate-purity test fails if `pkg/policy` or its CLI usage references any `*Method` enum in a branch.
- [ ] Composition-root test fails if `none/bwrap/docker/nix.New` is referenced outside `plugins.go`.
- [ ] Registry-factory test asserts `Factory` is a function type (lazy binding).
- [ ] Tests are mapped to their guarded principle in a header comment; all green in CI.

## Files Modified/Created

- Created: `packages/agent/agent-cli-go/internal/architecture_test.go`
- Modified (only if needed): `packages/agent/agent-cli-go/moon.yml`

## Dependencies

- `02-cli-isolation-provision-network.md` — creates the `internal/{isolation,provision,network}/{shared,leaves}` dirs and bridge files the tests assert against.
- `03-cli-workspace-terminal.md` — creates `internal/workspace/{shared,git}` exercised by the no-sibling-reach test.
- `04-cli-cmd-flags-wiring.md` — lands the `ResolvedAxes` struct and the final `plugins.go` composition root and policy call site the concretes/policy tests scan; tests must run against the settled layout.

## Estimated Duration

Half a day — one new test file plus a moon-task confirmation; effort is in getting the `packages.Load`/`go/ast` helpers precise and the path constants right, not in new product code.
