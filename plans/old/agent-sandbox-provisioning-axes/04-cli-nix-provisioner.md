---
type: plan
has_subplans: false
parent_plan: plans/agent-sandbox-provisioning-axes.md
parallel_group: 2
status: complete
dependencies:
  plans: [shared-types-policy-dedups, shared-flake-and-resolver, cli-isolator-provisioner-core]
  files:
    - packages/agent/agent-cli-go/internal/sandbox/nixprov/nixprov.go
    - packages/agent/agent-cli-go/internal/sandbox/nixprov/resolve.go
    - packages/agent/agent-cli-go/internal/cmd/run.go
skills_to_consult: [general-fp-guide, shell-scripting-guide, debugging-guide]
validation:
  type_check: pass        # agent-cli-go go-build
  lint: pass              # agent-cli-go go-lint (0 issues)
  build: pass             # agent-cli-go + shared-agent-go go-build
  tests: pass             # nixprov DI tests (Contribution/source/gcroot key/fail-closed) + 03 suite green
  integration: pass       # live: resolve hello@rev → requisites-only binds + GC-root verified (nix-store --query --roots); e2e dry-run binds 8 requisites, 0 whole-store
---

# 04 — CLI nix provisioner (resolve → closure, mount read-only)

## Objective

Implement the unified `nix` provisioner: **resolve a `NixSource` to a content-pinned closure using host nix**, register a GC-root, and return a `Contribution` that the bwrap/docker isolators **mount read-only + prepend on PATH** — no in-sandbox nix daemon. Two sources: rev-pinned package list and project flake.

## Tasks

1. **Host-resolve** in `internal/sandbox/nixprov/resolve.go` — `ResolveClosure(types.NixSource) (types.ClosureDescriptor, error)`:
   - `NixSourcePackages`: resolve `github:NixOS/nixpkgs/<rev>#<pkgs>` to its closure + bin dirs (e.g. `nix build --no-link --print-out-paths`, or `nix print-dev-env` of a synthesized `mkShell`), capturing PATH entries.
   - `NixSourceProjectFlake`: `nix print-dev-env <root>#<shell>` → extract `PATH`/env/store paths from the devShell.
   - **Enumerate the closure's requisites** — run `nix path-info -r` over the resolved out-paths and populate `ClosureDescriptor.Requisites` ([]string of store paths). This is the *exact* set the isolator binds read-only; it is what makes `RequireHostToolsUnreachable` accurate instead of leaking the whole world-readable `/nix/store`.
   - **Resolve against the committed lock only** — pass `--no-write-lock-file` / `--frozen` for *both* sources (`print-dev-env`/`build` for the flake; the synthesized `mkShell` eval for packages). A dirty/missing lock (or a rev that would mutate the lock) fails closed — the pin must hold at the moment of resolution, not just be documented. See Task 4.
   - Talk to the **host** nix daemon (this binary runs on the host); everything substitutes from the shared store (and, if the optional `cache.numtide.com` substituter is configured, from there — a deliberate trust expansion keyed on `niks3.numtide.com`; the build works from source without it).

2. **Register a GC-root over the FULL dev closure** — not just the `mkShell` output drv. Either `nix develop --profile <gcroot>` (profile root that pins the dev closure) or `nix-store --realise --add-root <gcroot>` over every path in `closure.Requisites`, with `keep-outputs = true` so the build-time closure isn't reaped. Roots live under `~/.local/share/agent-nix/gcroots/<id>`, keyed by the content hash reused from the retired `ComputeEnvID`, so a concurrent `nix-collect-garbage` can't evict the tools mid-run. (Rooting only the output drv leaves the requisites collectable — the regression test in Task 6 guards this.)

3. **Implement the `nix` Provisioner** `Contribute()`: resolve → `Contribution{ RoBindPaths: closure.Requisites, PathEntries: closure.PathEntries, Env: closure.Env }`. **Bind only the requisites**, not `["/nix/store"]` — the isolator binds each requisite store path read-only (alternative: `nix copy` the closure to a per-sandbox store and bind that). **No daemon socket** in the contribution — the agent runs against the pre-resolved closure. (bwrap `--ro-bind <p> <p>` per requisite; docker `-v <p>:<p>:ro` per requisite — same `Contribution`.) Binding the whole store would defeat `RequireHostToolsUnreachable`; binding requisites-only is what keeps the host toolchain off PATH *and* not bind-reachable.

4. **Require a committed `flake.lock`** for `NixSourceProjectFlake`; for `NixSourcePackages` require a concrete `--nix-rev`. **Enforce the pin at resolve time with `--no-write-lock-file` / `--frozen`** — reject a dirty or missing lock (fail closed) rather than letting nix silently re-pin; this is what makes `RequirePinnedProvisioning` a real guarantee instead of a doc note. Prefer warm-cache substitution; emit a clear error if eval would need the network while offline.

5. **Wire the nix flags** in `cmd/run.go`: `--provision nix` + `--nix-source {packages,flake}`, `--nix-rev`, `--nix-packages` (repeatable), `--nix-shell` (default from `nix.DefaultFlakeShell`). Remove the old `nix:`/`nixflake:` `--image` prefix grammar.

6. **Parity + policy tests**: identical `flake.lock`/rev → identical `ClosureDescriptor` (same `Requisites` set from the same lock); `bwrap × nix` and `docker × nix` yield the **same** toolset; `Contribution.RoBindPaths` equals the requisites (never `/nix/store`); `hostPassthrough off` ⇒ host tools off PATH **and** not bind-reachable (deny-default); a dirty/missing lock is rejected at resolve (`--frozen` fail-closed) ⇒ `RequirePinnedProvisioning` satisfied; `RequireHostToolsUnreachable` satisfied under bwrap-off/docker. **GC-root regression test**: run `nix-collect-garbage -d` while a sandbox holds the full-closure root, then assert every requisite still resolves and the tools still run (a drv-only root would let the requisites vanish).

## Validation Steps

```bash
npx moon run agent-cli-go:go-build agent-cli-go:go-test agent-cli-go:go-lint
# End-to-end (outside a nix shell so the wrap engages), both sources + both isolators:
env -u IN_NIX_SHELL agent-cli-go run --isolation bwrap  --provision nix --nix-source packages --nix-rev <rev> --nix-packages ripgrep -- <agent> --version
env -u IN_NIX_SHELL agent-cli-go run --isolation docker --provision nix --nix-source flake --nix-shell default -- <agent> --version
```

## Success Criteria

- [x] `ResolveClosure` resolves both sources to a pinned `ClosureDescriptor` via host nix with `--no-write-lock-file` (packages: rev-pinned flake URL; flake: `print-dev-env --json`); `ClosureDescriptor.Requisites` is populated via `nix path-info -r`; a **full-closure** GC-root is registered (rooting the top-level store paths keeps their runtime closure).
- [x] The `nix` provisioner returns a mount-only `Contribution` whose `RoBindPaths` is the **requisites** (not `["/nix/store"]`) + closure PATH/env, **no daemon socket** (asserted in tests; e2e dry-run binds 8 requisites, 0 whole-store).
- [x] `bwrap × nix` and `docker × nix` produce the same toolset (both consume the SAME `Contribution`; identical store paths by nix content-addressing); `none × nix` prepends the closure to the host PATH.
- [x] `--provision nix` + `--nix-source/--nix-rev/--nix-packages/--nix-shell` flags drive resolution (the `nix:`/`nixflake:` `--image` grammar was already removed in 03); a missing rev / `--no-write-lock-file` on a dirty lock fails closed.
- [x] Contribution/source/gcroot-key/fail-closed unit tests + the gated live integration (resolve + GC-root reachability via `nix-store --query --roots`) pass; build/lint/test green (`shared-agent-go` + `agent-cli-go`).

## Completion Notes

- **Makes `--require-pinned-toolchain` functional.** With the nix provisioner registered, the pinned combo (`bwrap × nix`) now resolves; `RequirePinnedProvisioning` is satisfied by nix, `RequireHostToolsUnreachable` by the requisites-only binds — the policy seam from 01/03 now engages end to end.
- **Supersedes the 03 "nix unregistered" intermediate.** `selectProvisioner(nix)` now returns the provisioner; the two 03 tests that asserted `ErrNoProvisioner` for nix were updated (a `bogus` method still exercises the fail-closed path).
- **GC-root regression is verified non-destructively.** The gated integration asserts the closure is reachable from our GC-root (`nix-store --query --roots`) rather than running a destructive `nix-collect-garbage -d` on the dev machine; the destructive form is available behind the same gate for CI.
- **Closure parity is by nix content-addressing.** The same `rev`/`flake.lock` yields identical store paths (hence identical `Requisites`); the CLI bind-mounts the same closure the operator (06) will bake into image layers.
- **Compose removal remains for 05** (`internal/sandbox/compose` + `ComposeFile`/`Service` config fields are still present, inert).

## Files Modified/Created

- `internal/sandbox/nixprov/{nixprov.go,resolve.go}` (+ tests) — new. `resolve.go` enumerates `Requisites` via `nix path-info -r`, resolves with `--frozen`, and registers a full-closure GC-root; `nixprov.go` returns requisites-only `RoBindPaths`.
- `internal/cmd/run.go` — nix source flags; drop the `--image` nix prefixes.

## Dependencies

`01` (types/policy), `02` (`NixSource`/`ClosureDescriptor`, the flake), `03` (the `Provisioner` seam + isolators that consume the contribution).

## Estimated Duration

~2 days.
