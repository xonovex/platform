---
type: plan
has_subplans: false
parent_plan: plans/agent-sandbox-provisioning-axes.md
parallel_group: 1
status: complete
dependencies:
  plans: [shared-types-policy-dedups]
  files:
    - flake.nix
    - flake.lock
    - nix/agent-env.nix
    - nix/mkAgentImage.nix
    - packages/shared/shared-agent-go/pkg/nix/nix.go
    - packages/shared/shared-agent-go/pkg/nix/source.go
    - packages/agent/agent-cli-go/internal/nixenv/
skills_to_consult: [general-fp-guide, docker-guide, shell-scripting-guide, debugging-guide]
validation:
  type_check: pass        # shared-agent-go go-build
  lint: pass              # shared-agent-go go-lint (0 issues)
  build: pass             # shared-agent-go go-build + nix flake check + agentImageTest/agentEnvTest realize
  tests: pass             # shared-agent-go go-test (incl. pkg/nix source_test.go)
  integration: pass       # nix flake check, rebuild-digest-equality identical, devShell/image closure parity
---

# 02 — The pinned flake (with llm-agents.nix) + the shared closure/source model

## Objective

Establish the **one `flake.lock`-pinned declarative core** both surfaces resolve from: add `numtide/llm-agents.nix` as a pinned input (packaging only; isolation out of scope), expose a parameterized **devShell/package-set** (for the CLI's runtime resolve) and a single **`streamLayeredImage` target** (for the operator) built via a **vendored `nix/mkAgentImage.nix`**, and add the shared Go `NixSource`/`ClosureDescriptor` types — the latter carrying the closure's `Requisites` so isolators bind ONLY the closure, never the whole `/nix/store`. Retire the channel-pinned `buildEnv` renderer.

Both surfaces provision the SAME content-addressed store-path closure (identical store-path hashes from the same `flake.lock`) — the CLI bind-mounts store directories, the operator bakes them into TAR layers; verify by comparing `nix path-info -r`, NOT layer bytes. The committed-lock resolve discipline (`--no-write-lock-file`/`--frozen`) lives in the CLI host-resolve (04); this subplan only ships the pinned `flake.lock` + pure types.

## Tasks

1. **Add `llm-agents.nix` as a pinned flake input** in the workspace `flake.nix`:
   ```nix
   inputs.llm-agents.url = "github:numtide/llm-agents.nix";  # pinned by flake.lock to a rev
   ```
   `llm-agents.nix` is **packaging only; isolation is out of scope** — consumers pin the agent binaries via **`flake.lock` ALONE** (the upstream `hashes.json`/`update.py` is upstream bookkeeping, NOT the consumer pin surface). Optionally add `extra-substituters = https://cache.numtide.com` to the nix config used by CLI/CI/operator builds; the trusted public key NAME is **`niks3.numtide.com`**, not `cache.numtide.com`:
   ```
   extra-trusted-public-keys = niks3.numtide.com-1:DTx8wZduET09hRmMtKdQDxNNthLQETkc/yaX7M4qK0g=
   ```
   Adding `cache.numtide.com` is a **TRUST EXPANSION** (numtide's CI/signing key then builds the agent binaries on your behalf) — call it out in Risk Assessment; the build works from source without it. Do **NOT** set `inputs.nixpkgs.follows` against `llm-agents` (breaks its binary-cache hits); `flake.lock` still pins the exact rev.

2. **Create `nix/agent-env.nix`** — a function `{ agent, packages ? defaultBasePackages, extraPackages ? [] }` producing **two** outputs from the same closure:
   - `devShell` (a `mkShell` with `[agent] ++ packages ++ extraPackages` on PATH) — the CLI resolves this.
   - `image` via the vendored `nix/mkAgentImage.nix` (task 2a) wrapping `pkgs.dockerTools.streamLayeredImage` — the operator builds this. Pre-create XDG dirs (`.config/.cache/.local/.local/state`) under HOME (per agent-images).
   Source `agent` from `llm-agents.packages.<system>.<name>`. Expose both under workspace flake outputs (`devShells`/`packages`).

2a. **Vendor `nix/mkAgentImage.nix` (ADAPT, don't reuse).** Fork `nothingnesses/agent-images` `lib/mkAgentImage.nix` into the repo rather than taking it as a live flake input — it's **bus-factor 1** and uses `buildLayeredImage` + a NAMED user. Port its layout into a **`streamLayeredImage`** build (`streamLayeredImage` never realizes a store tarball):
   - hand-written `/etc/passwd` + `/etc/group` for **NUMERIC uid 1000** (don't rely on a named user);
   - `mkdir`+`chown` of `HOME/.config`, `.cache`, `.local`, `.local/state`, `workspace`;
   - `config = { User = "1000:1000"; WorkingDir = "/workspace"; }`.
   ```nix
   pkgs.dockerTools.streamLayeredImage {
     name = "agent"; maxLayers = 100;          # deliberate: 128 sits at overlay2's modern ceiling (zero headroom)
     contents = [agent] ++ packages ++ extraPackages ++ [ passwdLayer ];
     config = { User = "1000:1000"; WorkingDir = "/workspace"; };
     # NEVER set created = "now" — dockerTools defaults to the epoch; a fixed time is the reproducibility invariant.
   }
   ```
   Commit to **ONE** image builder: `streamLayeredImage`. (`nix2container` only if you push on every small change — note its `maxLayers` defaults to 1; out of scope for the committed design.)
   **Reproducibility invariant:** never set `created = now`; add a **rebuild-digest-equality** smoke (build twice, assert identical image digest).

3. **Add the shared source/closure types** to `shared-agent-go/pkg/nix/source.go`:
   ```go
   type NixSourceKind string
   const ( NixSourcePackages NixSourceKind = "packages"; NixSourceProjectFlake = "project-flake" )
   type NixSource struct {
       Kind     NixSourceKind
       Rev      string   // pinned nixpkgs rev (packages source)
       Packages []string // package set (packages source)
       FlakeRef string   // <projectRoot> (project-flake source)
       Shell    string   // devShell name (project-flake source)
   }
   // ClosureDescriptor is the resolved, mount-ready result a realizer produces.
   type ClosureDescriptor struct {
       StorePaths  []string          // top-level realized paths (devShell/agent outputs)
       Requisites  []string          // transitive closure store paths (`nix path-info -r`) — isolators bind ONLY these, never the whole /nix/store
       PathEntries []string
       Env         map[string]string
   }
   ```
   `Requisites` is what makes `RequireHostToolsUnreachable` accurate: the bwrap/docker isolators bind exactly the resolved closure's requisites (read-only), not the world-readable store. The host-resolve in 04 **populates `Requisites`** by running `nix path-info -r` over the descriptor; this subplan only declares the field.
   Add `ValidateSource(NixSource) error` (reuse `nix.ValidatePin`/`ExpandPackageSets`). These are pure types/validators — **no** host nix call (the host-resolve impl lives in 04; the image build in 06).

4. **Retire the channel-pinned renderer.** Delete `agent-cli-go/internal/nixenv/render.go` (the `fetchTarball`/`buildEnv` renderer, `GetNixpkgsTarballURL`) and the now-dead `resolve.go`/`build.go` paths that only served it. Keep any pure helper that the 04 host-resolve will reuse (e.g. `ComputeEnvID` for the GC-root/cache key) — move it next to `source.go` in shared.

5. **Pin a concrete `flake.lock`** and add a `nix flake check` smoke (the workspace flake evaluates; `agent-env.nix` builds a tiny test image + devShell against a pinned rev).

6. **Document the source→output mapping** in `nix/agent-env.nix` header: `NixSourcePackages` → rev-pinned devShell/image; `NixSourceProjectFlake` → the project's own flake (CLI only — the operator always builds the image from the synthesized env).

## Validation Steps

```bash
npx moon run shared-agent-go:go-build shared-agent-go:go-test shared-agent-go:go-lint
nix flake check                                  # workspace flake + agent-env evaluate
nix build .#agentImageTest && nix develop .#agentEnvTest --command true   # both outputs realize
# rebuild-digest-equality smoke (reproducibility invariant): identical digest across two builds
nix build .#agentImageTest --rebuild && cmp <(./result | sha256sum) <(./result | sha256sum)
```

## Success Criteria

- [x] `llm-agents.nix` is a `flake.lock`-pinned input (rev `05f2ea60…`; consumers pin via `flake.lock` alone); no `follows` against it; `cache.numtide.com` is NOT added (agents build from source) — the `niks3.numtide.com-1:…` key + trust-expansion note are recorded in the `flake.nix` input comment.
- [x] `nix/mkAgentImage.nix` is vendored (adapted from `nothingnesses/agent-images`): numeric uid-1000 via hand-written `/etc/passwd`+`/etc/group`, XDG/`workspace` dirs owned 1000:1000, `streamLayeredImage`, `maxLayers = 100`, no `created = now` (verified: config `User=1000:1000`/`WorkingDir=/workspace`, dirs `1000/1000`, fixed `1980-01-01` epoch).
- [x] `nix/agent-env.nix` produces a devShell **and** a single `streamLayeredImage` from one parameterized closure (uid-1000/`/workspace`/XDG); both surfaces resolve the **same** store paths (verified: `print-dev-env` of `agentEnvTest` and the image both reference `…-hello-2.12.3` / `…-coreutils-9.11`).
- [x] `shared-agent-go/pkg/nix` exposes `NixSource`/`ClosureDescriptor` (with `Requisites`)/`ValidateSource` (pure) + the moved `ComputeEnvID`; the channel-pinned `buildEnv` renderer (`render.go`/`build.go`) is deleted.
- [x] `nix flake check` passes; the test image + devShell build from a pinned rev; the rebuild-digest-equality smoke is identical (`5b59835f…` across `--rebuild`).

## Completion Notes

- Scope per the wave design: 02 is the shared foundation (flake + pure Go types) and retires the channel renderer. Its validation is deliberately scoped to `shared-agent-go` + the flake, **not** `agent-cli-go:go-build`.
- Deleting `render.go`/`build.go` (the channel `buildEnv` path) leaves exactly one intended, contained dangling reference — `internal/sandbox/nix/nix.go:70` (`nixenv.BuildEnv`/`BuildOptions`), the old fused `nix` tier executor. **Subplan 03** removes that executor when it collapses the tiers into `Select` + isolators (and `04` lands the `nixprov` provisioner). The `nixenv` package itself and `internal/sandbox/nixflake` still compile (they use the surviving `GetAgentNixDir`/`GetAgentsDir`/`ExpandPackageSets` + `FlakeSandboxConfig`/`DefaultFlakeShell`). `ComputeEnvID` now lives in shared for 04's GC-root key.
- Consequence: `agent-cli-go` does not fully build until 03 lands; 02 should land together with the group-2 CLI wave (per the parent's "shared foundation must merge before the CLI/operator waves resolve their imports"), not released standalone.

## Files Modified/Created

- `flake.nix`, `flake.lock` — `llm-agents` input + outputs + (optional) `niks3.numtide.com` cache key.
- `nix/agent-env.nix` — new (devShell + image from one closure).
- `nix/mkAgentImage.nix` — new (vendored/adapted from `nothingnesses/agent-images`; `streamLayeredImage`, numeric uid 1000, `maxLayers = 100`, fixed `created`).
- `shared-agent-go/pkg/nix/source.go` (+ test) — new (`NixSource`/`ClosureDescriptor` with `Requisites`/`ValidateSource`).
- `agent-cli-go/internal/nixenv/` — delete the channel-pinned renderer/dead paths.

## Dependencies

`01-shared-types-policy-dedups` (consumes the package conventions; lands second in group 1).

## Estimated Duration

~1.5 days.
