---
type: plan
has_subplans: false
parent_plan: plans/agent-security-hardening.md
parallel_group: 1
status: pending
dependencies:
  plans: []
  files:
    - packages/agent/agent-cli-go/internal/isolation/docker/**
    - packages/agent/agent-cli-go/internal/network/**
    - packages/agent/agent-cli-go/internal/cmd/run.go
skills_to_consult:
  - docker-guide
  - microkernel-pattern-guide
  - testing-guide
  - code-quality-guide
validation:
  type_check: pending
  lint: pending
  build: pending
  tests: pending
  integration: pending
---

# 02 — CLI: Docker Proxy Egress Fails Closed

## Objective

Stop the docker isolator from claiming egress restriction it does not enforce.
Phase A: `--isolation docker --network proxy` refuses to run with a clear
error. Phase B: implement real enforcement with a docker `--internal` network
plus a proxy container. Also wire the two policy guarantees that currently
have no CLI flag.

## Context (read this first — no other context is assumed)

File:line references are anchors as of `main` @ `a6f765e2` (2026-07-02) and
will drift as earlier subplans land — always read the cited file before
editing; if a line reference doesn't match, locate the named construct
instead.

`packages/agent/agent-cli-go` is a Go CLI (module path
`github.com/xonovex/platform/packages/cli/agent-cli-go`) that runs an untrusted
AI agent inside a sandbox. The sandbox has three orthogonal axes — isolation
{none,bwrap,docker}, provision {none,nix,command}, network {host,none,proxy} —
plus a policy engine (`internal/sandbox/policy.go`) with four fail-closed
guarantees gated on capability booleans. Threat model: the agent runs
untrusted, model-generated code; guarantees must never be granted unless the
runtime actually enforces them.

Current defects, verified against source:

1. `internal/isolation/docker/network.go:18`: network mode `proxy` maps to
   `--network bridge` — i.e. normal NAT'd egress. The only "restriction" is
   advisory `HTTP_PROXY`/`HTTPS_PROXY` env vars set on the container
   (`docker.go` around line 200). Untrusted code can ignore those env vars and
   open direct TCP connections. Nothing blocks the metadata endpoint
   (169.254.169.254), RFC1918, or loopback.
2. Despite that, `network.EgressIsRestricted(proxy)` returns true (see
   `internal/network/shared/` and `internal/network/types.go`), so
   `RequireEgressRestricted` is satisfied by an unrestricted bridge. The bwrap
   isolator is fine (`--unshare-net` for none; verify how it handles proxy).
3. `internal/cmd/run.go` `proxyEnv()` (around line 390): when the
   `AGENT_SANDBOX_PROXY` env/proxy URL is unset, it returns nil with NO error —
   proxy mode silently degrades to a plain bridge.
4. `run.go:132-137`: only one flag (`--require-pinned-toolchain`) exists,
   setting `RequirePinnedProvision` + `RequireHostToolsUnreachable`. There is
   no flag for `RequireEgressRestricted` or `RequireKernelIsolation` — half the
   policy engine is unreachable from the CLI.

## Tasks

1. **Phase A — refuse docker+proxy.**
   - Read `internal/isolation/docker/docker.go` and find where the isolator
     validates/prepares a run and where its `Capabilities` are computed
     (`HidesHost`, `KernelIsolated` are around lines 58-64).
   - Until Phase B is complete, docker + `Network=proxy` must return an
     explicit error BEFORE launching, e.g.:
     `docker isolation does not enforce proxy egress yet; use --network none, bwrap isolation, or --network host (host egress is unrestricted and does not satisfy --require-egress-restricted)`.
   - The capability path must agree: egress-restricted must not be derivable
     for docker+proxy. Do NOT special-case the policy engine on method names —
     it is method-agnostic by design (checks capability booleans only); the
     isolator itself must report/refuse correctly.

2. **Missing proxy URL is an error, all isolators.**
   - `run.go` `proxyEnv()`: change the signature to return
     `(map[string]string, error)`; `Network=proxy` with no proxy URL configured
     ⇒ error, never nil-and-continue. Update callers.
   - Check the bwrap path handles `proxy` coherently (read
     `internal/isolation/bwrap/bwrap.go` network handling): proxy mode with no
     reachable proxy must also refuse.

3. **Phase B — real docker proxy enforcement.**
   - Design (from parent plan decision 1): per run, create a docker network
     with `--internal` (no external routing); attach the sandbox container ONLY
     to that network; run a proxy container (the allowlist proxy the CLI
     already configures via `internal/network/proxy` — read that package)
     attached to BOTH the internal network and the default bridge; point
     `HTTP_PROXY`/`HTTPS_PROXY` at the proxy container's internal-network
     address. Direct egress is then impossible at the network layer; only
     proxied, allowlisted HTTP(S) leaves.
   - Teardown: remove the proxy container and the network on exit (including
     error paths — use the same cleanup mechanism the docker isolator already
     uses for the sandbox container).
   - Once enforced, docker+proxy stops refusing and reports egress-restricted
     truthfully.
   - If the existing proxy implementation is host-side only and cannot run as
     a container, containerize the invocation (the proxy binary is part of
     this CLI; check `internal/network/proxy` before assuming).

4. **Wire the missing policy flags.**
   - `run.go:132-137`: add `--require-egress-restricted` and
     `--require-kernel-isolation`, each setting exactly its policy field.
     Follow the existing flag naming/help style. Update command help text so
     all four guarantees are documented.

5. **Tests.**
   - Unit: docker isolator refuses proxy (Phase A) / computes capabilities
     correctly (Phase B); `proxyEnv` errors without URL; flag wiring maps to
     policy fields.
   - Integration (`test/integration/`, `-tags=integration`, requires docker):
     Phase B — container on the internal network cannot reach an external IP
     directly but can via the proxy; metadata IP blocked.

## Validation Steps

Prerequisites: Go toolchain is nix-managed — if `go` is missing, run inside
`nix develop`. The integration tests require a running docker daemon on the
host.

```bash
npx moon run agent-cli-go:go-build
npx moon run agent-cli-go:go-lint
npx moon run agent-cli-go:go-test
cd packages/agent/agent-cli-go && go test -tags=integration ./test/integration/
```

## Success Criteria

- [ ] `--isolation docker --network proxy` either refuses (Phase A) or
      provably blocks direct egress at the network layer (Phase B) — at no
      point does it run with advisory-env-only "restriction".
- [ ] `Network=proxy` with no proxy URL is an error on every isolator.
- [ ] All four `Require*` guarantees are settable via CLI flags.
- [ ] Policy engine remains method-agnostic (no isolator names in
      `internal/sandbox/`); `architecture_test.go` still passes.
- [ ] All validation commands pass.

## Files Modified/Created

- Modified: `internal/isolation/docker/network.go`,
  `internal/isolation/docker/docker.go`, `internal/cmd/run.go`,
  `internal/network/shared/*` (only if capability derivation lives there),
  bwrap network handling if needed, tests alongside each.
- Created: proxy-container/network lifecycle code inside the docker plugin
  package; integration tests.

## Dependencies

None. Serializes BEFORE 06-cli-host-reach-and-pinning (same plugin files).

## Estimated Duration

Phase A + flags: 0.5 day. Phase B + integration tests: 1–1.5 days.
