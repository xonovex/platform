---
type: plan
has_subplans: false
parent_plan: plans/agent-security-hardening.md
parallel_group: 4
status: pending
dependencies:
  plans:
    - 02-cli-egress-fail-closed
  files:
    - packages/shared/shared-agent-go/pkg/isolation/types.go
    - packages/agent/agent-cli-go/internal/isolation/**
    - packages/agent/agent-cli-go/internal/provision/nix/**
    - packages/agent/agent-cli-go/internal/cmd/run.go
    - packages/agent/agent-cli-go/internal/terminal/tmux/**
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

# 06 — CLI: Honest Host-Reach, Real Nix Pinning, Secrets Off Argv

## Objective

Make `RequireHostToolsUnreachable` honest (credential/tool mounts become
opt-in and the capability is computed from the actual mount set), fix the
worktree double-bind that breaks both isolators, make the nix pin actually
fail closed, surface swallowed credential errors, and get secrets out of argv
and tmux command strings.

## Context (read this first — no other context is assumed)

File:line references are anchors as of `main` @ `a6f765e2` (2026-07-02) and
will drift as earlier subplans land — always read the cited file before
editing; if a line reference doesn't match, locate the named construct
instead.

`packages/agent/agent-cli-go` (module
`github.com/xonovex/platform/packages/cli/agent-cli-go`) sandboxes an
untrusted AI agent via bwrap or docker. Shared types live in
`packages/shared/shared-agent-go`. The policy guarantee
`RequireHostToolsUnreachable` is documented as "host tools off PATH AND not
bind-reachable" (`internal/sandbox/policy.go:54-56`). The agent is launched
with `--permission-mode bypassPermissions`
(`internal/agents/claude.go:20`), so anything mounted into the sandbox is
fully readable/executable by untrusted code.

Current defects, verified against source (worktree failures reproduced live):

1. **Credential and tool mounts contradict `HidesHost=true`.**
   `shared-agent-go/pkg/isolation/types.go:27-41` (`UserConfigPaths`) lists
   `~/.ssh`, `.npmrc`, `.config` (gh tokens), `.claude.json` (OAuth creds),
   `.cargo`, `.local`, `.npm`, `.npm-global`. `bwrap.go:97-102` and
   `docker.go:123-127` mount them all; `~/.local/bin` and `~/.cargo/bin`
   contain executables, so host tools ARE bind-reachable. Yet `bwrap.go:43`
   and `docker.go:58` return `HidesHost=true` unconditionally, granting
   `RequireHostToolsUnreachable` falsely.
2. **Worktree double-bind.** `internal/cmd/run.go:239` adds `sourceRepoDir`
   to read-write `BindPaths` while `RunConfig.RepoDir` is separately bound
   read-only. Docker (`docker.go:118-119` + `138-144`) emits two `-v` flags
   for the same destination ⇒ daemon rejects with `Duplicate mount point`
   (reproduced). Bwrap (`bwrap.go:106-108` + `127-133`) mounts the later rw
   bind over the earlier ro bind ⇒ source repo silently writable (reproduced),
   defeating the documented read-only worktree source.
3. **Nix pin not fail-closed.** `internal/provision/nix/resolve.go:44,76` uses
   `--no-write-lock-file`, which only suppresses WRITING the lock — nix warns
   and proceeds with an in-memory re-resolution when the lock is dirty or
   missing. The flag that errors instead is `--no-update-lock-file`. The
   comment at `resolve.go:19-21` claims erroring behavior the flag doesn't
   provide, and `Provisioner.Pinned()` returns true unconditionally
   (`nix.go:74`), so `RequirePinnedProvision` can be satisfied by an unpinned
   resolve.
4. **Provider-credential errors swallowed.** `bwrap.go:180` and
   `docker.go:189` use `if providerEnv, err := ...; err == nil` (error
   dropped); `none.go:67` uses `providerEnv, _ =`. Requesting
   `--provider glm` without its token env launches anyway, silently falling
   back to whatever credentials are mounted from `~/.claude`.
5. **Secrets in argv.** Tokens are passed as `--setenv K V` to bwrap
   (`bwrap.go:146`), `-e K=V` to docker (`docker.go:155`), and baked into the
   tmux command string as `export K="V"; ...` (`terminal/tmux/env.go:47-70`,
   `tmux.go:56-57` — which also serializes the ENTIRE host environ for
   bwrap/docker runs). All are world-readable via `/proc/<pid>/cmdline` and
   `docker inspect` for the run's lifetime.

## Tasks

1. **Split `UserConfigPaths`; credentials opt-in; compute `HidesHost`.**
   - In `shared-agent-go/pkg/isolation/types.go`, split the single list:
     non-credential tool caches vs credentials/config (`.ssh`, `.npmrc`,
     `.config`, `.claude.json`). Parent-plan decision 4: credentials are
     mounted ONLY via an explicit repeatable flag (e.g.
     `--mount-user-config <group|path>`); default mounts NONE of them.
   - Any mount that puts host executables in reach (`.local`, `.cargo`,
     `.npm-global` bins) must flip the computed capability:
     `HidesHost` becomes a function of the resolved bind set (+
     `hostPassthrough`), not a constant — change the `Isolator` capability
     derivation in `bwrap.go`/`docker.go` accordingly, so requesting
     `RequireHostToolsUnreachable` together with such mounts REFUSES with an
     error naming the offending mount.
   - Note the behavior break: flows that silently depended on mounted
     `~/.claude.json` will now fail — task 4's error surfacing turns that
     into an actionable message. Update `packages/agent/AGENTS.md` if the
     axes description changes.
2. **Fix the worktree double-bind.** `run.go:239`: in worktree mode, do NOT
   add `sourceRepoDir` to rw `BindPaths`; the ro `RepoDir` bind is the single
   mount. Integration tests (`-tags=integration`): docker worktree run
   launches successfully; bwrap worktree run cannot write to the source repo
   (attempt a write, assert failure).
3. **Real nix pinning.** `resolve.go:44,76`: replace `--no-write-lock-file`
   with `--no-update-lock-file`; fix the comment at `resolve.go:19-21` to
   state actual behavior. Make `Pinned()` (`nix.go:74`) honest: true only when
   the resolve ran against an existing lock in the error-on-drift mode. Test:
   a fixture flake with a deliberately stale/missing lock must make resolve
   FAIL (guard with `-tags=integration` if it needs a real nix binary).
4. **Surface credential errors.** `bwrap.go:180`, `docker.go:189`,
   `none.go:67`: propagate the `BuildProviderEnv` error; a requested provider
   whose token is missing aborts the launch with the underlying message.
5. **Secrets off argv.**
   - docker: use `-e KEY` (no value — docker reads it from the launcher
     process env); set the values in the launcher's env just before exec.
   - bwrap: audit how env reaches the sandbox (`--setenv` vs inherited env vs
     `--clearenv`); secrets must arrive via inherited process env or a file
     descriptor, never `--setenv K V` argv. Preserve the existing env-hygiene
     semantics (if `--clearenv` is in use, keep it — only the secret transport
     changes).
   - tmux (`env.go:47-70`, `tmux.go:56-57`): write exports to a 0600 tempfile,
     `source` + `rm` it in the launched shell, never inline values in the
     command string; stop serializing the full host environ — pass only the
     computed sandbox env.
   - Test: after launch, `/proc/<pid>/cmdline` of the isolator process and
     `docker inspect` output contain no token value (integration test can use
     a fake token like `test-secret-sentinel`).

## Validation Steps

Prerequisites: Go toolchain is nix-managed — if `go` is missing, run inside
`nix develop`. The integration tests require docker, `bwrap` (bubblewrap),
and `nix` on the host. `shared-agent-go` changes: also run that project's
tasks.

```bash
npx moon run shared-agent-go:go-build shared-agent-go:go-test
npx moon run agent-cli-go:go-build
npx moon run agent-cli-go:go-lint
npx moon run agent-cli-go:go-test
cd packages/agent/agent-cli-go && go test -tags=integration ./test/integration/
```

## Success Criteria

- [ ] Default run mounts no credential paths; `--mount-user-config` opts in
      per group; `RequireHostToolsUnreachable` + host-executable mounts ⇒
      refusal naming the mount.
- [ ] `HidesHost` is computed from the actual bind set in both isolators.
- [ ] Worktree mode: docker launches; bwrap source repo is read-only (both
      integration-tested).
- [ ] Stale/missing flake lock makes nix resolve fail; `Pinned()` reflects
      reality.
- [ ] Missing provider token aborts launch with the underlying error.
- [ ] No token value in `/proc/*/cmdline`, `docker inspect`, or tmux command
      strings (sentinel test).
- [ ] All validation commands pass.

## Files Modified/Created

- Modified: `shared-agent-go/pkg/isolation/types.go`,
  `agent-cli-go/internal/isolation/bwrap/bwrap.go`,
  `agent-cli-go/internal/isolation/docker/docker.go`,
  `agent-cli-go/internal/isolation/none/none.go`,
  `agent-cli-go/internal/provision/nix/{resolve.go,nix.go}`,
  `agent-cli-go/internal/cmd/run.go`,
  `agent-cli-go/internal/terminal/tmux/{env.go,tmux.go}`,
  `packages/agent/AGENTS.md`, tests alongside.

## Dependencies

Depends on `02-cli-egress-fail-closed` (same isolation plugin files —
serialize). Independent of the operator subplans.

## Estimated Duration

1.5 days.
