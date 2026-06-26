---
type: plan
has_subplans: false
parent_plan: plans/nix-toolchain-hardening.md
parallel_group: 3
status: pending
dependencies:
  plans: [sandbox-flake-devshell]
  files:
    - packages/shared/shared-agent-go/pkg/types/sandbox.go
    - packages/agent/agent-cli-go/internal/sandbox/registry.go
    - packages/agent/agent-cli-go/internal/cmd/run.go
    - packages/agent/agent-cli-go/internal/sandbox/bwrap/bwrap.go
    - packages/agent/agent-cli-go/internal/sandbox/none/none.go
    - packages/agent/agent-cli-go/internal/sandbox/nix/nix.go
    - packages/agent/agent-cli-go/internal/sandbox/registry_test.go
skills_to_consult: [general-fp-guide, code-review-guide, debugging-guide]
validation:
  type_check: pending
  lint: pending
  build: pending
  tests: pending
  integration: pending
---

# Sandbox Deny-Default: Mandate the Pinned Nix(flake) Tier

## Objective

Make host tools physically unreachable for pinned agent runs by adding a `RequirePinnedToolchain` policy that the sandbox selection seam (`internal/sandbox/registry.go`) enforces. When set, selection forces the nix(flake) tier from subplan 05 and returns an error for the leaky `bwrap`/`none` tiers and image-less `docker`/`compose` — because `bwrap` ro-binds host `/usr,/lib,/bin` and appends the host PATH (`bwrap.go:86-91`, `156-159`) and `none` execs directly on the host (`none.go:21-34`), neither can satisfy the guarantee.

## Tasks

1. **Add a `SandboxPolicy` value and a `Policy` field to `SandboxConfig`** — `packages/shared/shared-agent-go/pkg/types/sandbox.go` (struct at `14-35`, after the `SandboxMethod` consts `6-12`). The policy is a small pure value carried through selection; `SandboxNixFlake` is the new method introduced by subplan 05 (`sandbox-flake-devshell`).

   ```go
   const (
   	SandboxNone     SandboxMethod = "none"
   	SandboxBwrap    SandboxMethod = "bwrap"
   	SandboxDocker   SandboxMethod = "docker"
   	SandboxCompose  SandboxMethod = "compose"
   	SandboxNix      SandboxMethod = "nix"
   	SandboxNixFlake SandboxMethod = "nixflake" // defined in subplan 05
   )

   // SandboxPolicy expresses the isolation guarantees the caller demands of the
   // selected tier, independent of which method was requested.
   type SandboxPolicy struct {
   	// RequirePinnedToolchain mandates a tier whose toolchain comes entirely
   	// from a pinned source (a flake.lock devShell or a nix-built closure) with
   	// no host /usr,/lib,/bin bound and no host PATH appended. Tiers that leak
   	// host system directories (none, bwrap) are rejected at selection time.
   	RequirePinnedToolchain bool
   }
   ```

   Add `Policy SandboxPolicy` to the `SandboxConfig` struct so executors can assert it downstream. Note: keep `SandboxNixFlake` here only if subplan 05 has not already added it — coordinate to avoid a duplicate const.

2. **Add pure tier classification + a policy-enforcing selector** — `packages/agent/agent-cli-go/internal/sandbox/registry.go` (alongside `GetExecutor` `14-30`). Keep `GetExecutor` as the dumb factory; add a pure `tierIsolation` classifier and a `SelectExecutor` that applies the policy before constructing.

   ```go
   // Isolation classifies a tier's host-tool reachability guarantee.
   type Isolation int

   const (
   	// IsolationHostToolsLeaked: the tier ro-binds host /usr,/lib,/bin and/or
   	// appends the host PATH, so host binaries stay reachable (none, bwrap).
   	IsolationHostToolsLeaked Isolation = iota
   	// IsolationContainerPinned: host tools are unreachable only when a concrete
   	// --image/pin is supplied; an empty image resolves host-equivalent tools.
   	IsolationContainerPinned
   	// IsolationHostToolsUnreachable: the tier exposes only a nix-built closure
   	// (/nix/store + /env, PATH=/env/bin) or a flake.lock devShell — no host
   	// /usr,/lib,/bin is bound (nix, nixflake).
   	IsolationHostToolsUnreachable
   )

   // tierIsolation maps a method to its guarantee. Pure: depends only on method.
   func tierIsolation(method types.SandboxMethod) Isolation {
   	switch method {
   	case types.SandboxNix, types.SandboxNixFlake:
   		return IsolationHostToolsUnreachable
   	case types.SandboxDocker, types.SandboxCompose:
   		return IsolationContainerPinned
   	default: // SandboxNone, SandboxBwrap
   		return IsolationHostToolsLeaked
   	}
   }

   // SelectExecutor resolves an executor for method under policy. When the policy
   // requires a pinned toolchain, leaky tiers are rejected and image-less
   // container tiers are rejected; the nix(flake) tier is mandated.
   func SelectExecutor(method types.SandboxMethod, image string, policy types.SandboxPolicy) (types.SandboxExecutor, types.SandboxMethod, error) {
   	if policy.RequirePinnedToolchain {
   		if err := enforcePinnedToolchain(method, image); err != nil {
   			return nil, "", err
   		}
   	}
   	exec, err := GetExecutor(method)
   	if err != nil {
   		return nil, "", err
   	}
   	return exec, method, nil
   }

   // enforcePinnedToolchain rejects any method that cannot guarantee unreachable
   // host tools under a pinned-toolchain policy.
   func enforcePinnedToolchain(method types.SandboxMethod, image string) error {
   	switch tierIsolation(method) {
   	case IsolationHostToolsUnreachable:
   		return nil
   	case IsolationContainerPinned:
   		if image == "" {
   			return fmt.Errorf("sandbox %q needs a pinned --image under require-pinned-toolchain: an image-less container resolves host-equivalent tools", method)
   		}
   		return nil
   	default:
   		return fmt.Errorf("sandbox %q is rejected by require-pinned-toolchain: it ro-binds host /usr,/lib,/bin and appends host PATH (internal/sandbox/bwrap/bwrap.go:86-91,156-159), so host tools stay reachable — use the nix or nixflake tier", method)
   	}
   }
   ```

   Note: `tierIsolation` is the single source of truth for the per-tier guarantee, so adding a future tier forces a classification decision.

3. **Wire the `--require-pinned-toolchain` flag and route selection through `SelectExecutor`** — `packages/agent/agent-cli-go/internal/cmd/run.go` (flag vars `32-52`, `init` flags `58-79`, selection `163-177`). Default the unspecified method to the pinned tier; reject an explicit leaky tier via the enforcer.

   ```go
   // var block (~32-52)
   	flagRequirePinnedToolchain bool

   // init() flags (~58-79)
   	runCmd.Flags().BoolVar(&flagRequirePinnedToolchain, "require-pinned-toolchain", false,
   		"Mandate the nix(flake) tier; reject leaky bwrap/none and image-less docker/compose")

   // replace lines 163-168 (sandboxMethod / GetExecutor)
   	sandboxMethod := types.SandboxMethod(flagSandbox)
   	if flagRequirePinnedToolchain && !cmd.Flags().Changed("sandbox") {
   		// Mandate the pinned tier when the caller did not pick one explicitly.
   		sandboxMethod = types.SandboxNixFlake
   	}
   	policy := types.SandboxPolicy{RequirePinnedToolchain: flagRequirePinnedToolchain}
   	sandboxExecutor, sandboxMethod, err := sandbox.SelectExecutor(sandboxMethod, flagImage, policy)
   	if err != nil {
   		return err
   	}
   ```

   Also set `Policy: policy` on the `&types.SandboxConfig{...}` literal (`180-195`). The early `if sandboxMethod == types.SandboxNone` direct-exec branch (`207-229`) is now unreachable under the policy because `none` errors in `enforcePinnedToolchain`; leave it for the non-policy path.

4. **Document each tier's guarantee at the executor source** — add a doc comment above each `Executor` so the guarantee lives next to the code that provides (or breaks) it. `bwrap.go:13` and `none.go:8` document the leak; `nix.go:17` documents the unreachable guarantee.

   ```go
   // internal/sandbox/bwrap/bwrap.go (above `type Executor struct{}`, line 13)
   // Executor implements the bubblewrap tier.
   //
   // Isolation: IsolationHostToolsLeaked. buildBwrapArgs ro-binds host
   // /usr,/lib,/lib64,/bin,/etc (bwrap.go:86-91) and getSandboxEnvironment
   // appends the host PATH (bwrap.go:156-159), so host binaries stay reachable.
   // It cannot satisfy RequirePinnedToolchain and is rejected at selection.

   // internal/sandbox/none/none.go (above `type Executor struct{}`, line 8)
   // Executor implements direct host execution (no sandbox).
   //
   // Isolation: IsolationHostToolsLeaked. Execute runs the agent on the host with
   // the host PATH (none.go:21-34); every host tool is reachable. Rejected under
   // RequirePinnedToolchain.

   // internal/sandbox/nix/nix.go (above `type Executor struct{}`, line 17)
   // Executor implements the nix tier.
   //
   // Isolation: IsolationHostToolsUnreachable. buildBwrapArgs binds only
   // /nix/store + the env closure to /env with PATH=/env/bin (nix.go:252-262,
   // 336-343); no host /usr,/lib,/bin is bound, so host tools are unreachable.
   ```

   Note: the nixflake executor from subplan 05 carries the same `IsolationHostToolsUnreachable` doc comment; add it there too when 05 lands.

5. **Unit-test the policy matrix** — new file `packages/agent/agent-cli-go/internal/sandbox/registry_test.go` (table-driven, pure, no nix required). Covers: `denyHostTools + bwrap -> error`, `+ none -> error`, `+ docker (no image) -> error`, `+ docker (image) -> ok`, `+ nix -> ok`, `+ nixflake -> ok`, and `policy off -> bwrap selectable`.

   ```go
   package sandbox

   import (
   	"testing"

   	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/types"
   )

   func TestSelectExecutor_RequirePinnedToolchain(t *testing.T) {
   	deny := types.SandboxPolicy{RequirePinnedToolchain: true}
   	cases := []struct {
   		name    string
   		method  types.SandboxMethod
   		image   string
   		policy  types.SandboxPolicy
   		wantErr bool
   	}{
   		{"deny+bwrap rejects leak", types.SandboxBwrap, "", deny, true},
   		{"deny+none rejects host exec", types.SandboxNone, "", deny, true},
   		{"deny+docker without image", types.SandboxDocker, "", deny, true},
   		{"deny+docker with image", types.SandboxDocker, "alpine:3.20", deny, false},
   		{"deny+nix ok", types.SandboxNix, "", deny, false},
   		{"deny+nixflake ok", types.SandboxNixFlake, "", deny, false},
   		{"policy off keeps bwrap", types.SandboxBwrap, "", types.SandboxPolicy{}, false},
   	}
   	for _, tc := range cases {
   		t.Run(tc.name, func(t *testing.T) {
   			_, _, err := SelectExecutor(tc.method, tc.image, tc.policy)
   			if (err != nil) != tc.wantErr {
   				t.Fatalf("SelectExecutor(%s) err=%v, wantErr=%v", tc.method, err, tc.wantErr)
   			}
   		})
   	}
   }

   func TestTierIsolation_Classification(t *testing.T) {
   	want := map[types.SandboxMethod]Isolation{
   		types.SandboxNone:     IsolationHostToolsLeaked,
   		types.SandboxBwrap:    IsolationHostToolsLeaked,
   		types.SandboxDocker:   IsolationContainerPinned,
   		types.SandboxCompose:  IsolationContainerPinned,
   		types.SandboxNix:      IsolationHostToolsUnreachable,
   		types.SandboxNixFlake: IsolationHostToolsUnreachable,
   	}
   	for m, w := range want {
   		if got := tierIsolation(m); got != w {
   			t.Errorf("tierIsolation(%s)=%d, want %d", m, got, w)
   		}
   	}
   }
   ```

6. **Integration-test that the pinned tier hides host `/usr/bin`** — add `TestRequirePinnedToolchain_NixflakeHostToolsUnreachable` (skipped when `nix-build`/`bwrap`/`/nix/store` are absent, mirroring `nix.IsAvailable` `nix.go:26-46`). Select under the policy, then run a probe inside the sandbox asserting a host-only binary is unreachable and host `/usr/bin` is empty inside the namespace.

   ```go
   func TestRequirePinnedToolchain_NixflakeHostToolsUnreachable(t *testing.T) {
   	for _, bin := range []string{"nix-build", "bwrap"} {
   		if _, err := exec.LookPath(bin); err != nil {
   			t.Skipf("%s not available", bin)
   		}
   	}
   	if _, err := os.Stat("/nix/store"); os.IsNotExist(err) {
   		t.Skip("/nix/store not present")
   	}
   	executor, method, err := SelectExecutor("", "", types.SandboxPolicy{RequirePinnedToolchain: true})
   	if err != nil {
   		t.Fatalf("select: %v", err)
   	}
   	if method != types.SandboxNixFlake {
   		t.Fatalf("policy forced %s, want nixflake", method)
   	}
   	// Probe: list /usr/bin inside the sandbox; under the pinned tier the host
   	// /usr is never bound, so the directory is empty/absent and host-only tools
   	// do not resolve. The nixflake executor + flake fixture come from subplan 05.
   	cfg := &types.SandboxConfig{
   		Method:  method,
   		WorkDir: t.TempDir(),
   		Policy:  types.SandboxPolicy{RequirePinnedToolchain: true},
   		AgentArgs: []string{"-c",
   			`test -z "$(ls -A /usr/bin 2>/dev/null)" && ! command -v fdisk`},
   	}
   	if code, err := executor.Execute(cfg); err != nil || code != 0 {
   		t.Fatalf("host tools reachable under pinned tier: code=%d err=%v", code, err)
   	}
   }
   ```

   Note: the probe shape depends on the nixflake executor's command wiring from subplan 05; adjust `AgentArgs`/agent binary to that executor's contract. Keep it a real namespace check (not a string scan of args) so a regression that re-binds `/usr` actually fails.

## Validation Steps

Run from the repo root (`/home/mvierssen/Projects/xonovex/xonovex-platform`); all must pass.

```bash
# Type-check / vet (Go)
npx moon run agent-cli-go:go-typecheck
npx moon run shared-agent-go:go-typecheck

# Lint
npx moon run agent-cli-go:go-lint
npx moon run shared-agent-go:go-lint

# Build
npx moon run shared-agent-go:go-build
npx moon run agent-cli-go:go-build

# Unit tests (policy matrix + classification — no nix needed)
npx moon run agent-cli-go:go-test
# focused:
( cd packages/agent/agent-cli-go && go test ./internal/sandbox/... -run 'SelectExecutor|TierIsolation' -v )

# Integration (skips automatically without nix-build/bwrap/_nix store)
( cd packages/agent/agent-cli-go && go test ./internal/sandbox/... -run NixflakeHostToolsUnreachable -v )
```

## Success Criteria

- [ ] `types.SandboxPolicy{RequirePinnedToolchain}` and `SandboxConfig.Policy` exist; `SandboxNixFlake` is defined once (here or by subplan 05, not duplicated).
- [ ] `SelectExecutor` + pure `tierIsolation`/`enforcePinnedToolchain` live in `registry.go`; `GetExecutor` stays a plain factory.
- [ ] `--require-pinned-toolchain` defaults an unspecified method to `nixflake` and rejects explicit `bwrap`/`none` and image-less `docker`/`compose` with a clear error naming the leak.
- [ ] Each tier's guarantee is documented at its executor (`bwrap.go`, `none.go`, `nix.go`, and nixflake in subplan 05).
- [ ] `denyHostTools + bwrap -> error` and `denyHostTools + nixflake -> ok` are covered by the unit matrix; classification test pins every tier.
- [ ] Integration test confirms host `/usr/bin` is empty and a host-only binary is unreachable under the pinned nixflake tier (or skips cleanly without nix).
- [ ] typecheck, lint, build, unit tests green; integration green where nix is present.

## Files Modified/Created

- Modify: `packages/shared/shared-agent-go/pkg/types/sandbox.go` — `SandboxPolicy`, `SandboxConfig.Policy`, (coordinate) `SandboxNixFlake` const.
- Modify: `packages/agent/agent-cli-go/internal/sandbox/registry.go` — `Isolation`, `tierIsolation`, `SelectExecutor`, `enforcePinnedToolchain`.
- Modify: `packages/agent/agent-cli-go/internal/cmd/run.go` — `--require-pinned-toolchain` flag, selection via `SelectExecutor`, `Policy` on config.
- Modify: `packages/agent/agent-cli-go/internal/sandbox/bwrap/bwrap.go` — tier-guarantee doc comment.
- Modify: `packages/agent/agent-cli-go/internal/sandbox/none/none.go` — tier-guarantee doc comment.
- Modify: `packages/agent/agent-cli-go/internal/sandbox/nix/nix.go` — tier-guarantee doc comment.
- Create: `packages/agent/agent-cli-go/internal/sandbox/registry_test.go` — policy matrix, classification, and gated integration test.

## Dependencies

- **`sandbox-flake-devshell` (subplan 05) must land first.** It introduces the `nixflake` tier (`internal/sandbox/nixflake/`, the `types.SandboxNixFlake` method, and registration in `GetExecutor`). This subplan's policy forces selection to that tier and the integration test runs inside it — without 05 there is no pinned flake.lock devShell to mandate. Both are group 3 and serialize: flake-devShell, then deny-default.

## Estimated Duration

~0.5-1 day: ~2-3 h for the policy seam (types + registry + run.go wiring), ~1 h for tier doc comments, ~2-3 h for the unit matrix and the gated nix integration test.
