---
type: plan
has_subplans: false
parent_plan: ../runtime-enforcement-completion.md
parallel_group: 2
status: pending
dependencies:
  plans:
    - governance-decision-point-fail-closed
  files:
    - packages/skill/skill-claude-code/code-harness-guide/references/capabilities.md
    - packages/skill/skill-claude-code/code-harness-guide/**
    - packages/skill/skill-agent-governance/agent-governance-guide/assets/walking-skeleton/*
skills_to_consult:
  - plan-guide
  - shell-scripting-guide
  - skill-guide
  - testing-guide
  - connascence-guide
validation:
  type_check: not_run
  lint: not_run
  build: not_run
  tests: not_run
updated: "2026-07-17"
---

# One Claude Code Native Hook Proven to Block a Denied Action End-to-End

## Objective

Turn Claude Code's capability matrix from documentation into runtime conformance: a
semantic-intent → native-hook mapping is actually **registered** with the installed runtime and
observed to **block** a denied tool call end-to-end. The hook obtains its verdict from the Phase 1
decision point, so harness enforcement and admission enforcement share one decision logic. Only the
rows actually exercised are upgraded — the rest stay documentation-verified.

## Tasks

1. Author the real `PreToolUse` registration: produce the actual Claude Code settings entry and the
   registration procedure under `packages/skill/skill-claude-code/code-harness-guide/` (and its
   walking-skeleton assets), wiring the existing
   `agent-governance-guide/assets/walking-skeleton/guard.sh` (or the Phase 1 decision service) as the
   handler — a hook the harness genuinely loads, not a scratch `project-settings.json` the harness
   never reads (contrast `run-skeleton.sh:76-77`).
2. Map the semantic intent `before-tool-use` → native `PreToolUse` (per
   `code-harness-guide/references/capabilities.md:24`, "Tool before use") and drive a real session
   against the probed runtime `2.1.211`: a denied operation (write to a protected path) is **blocked by
   the harness** (tool call refused via exit `2`) and a permitted operation proceeds. Use
   `command`/decision-service enforcement — **not** the experimental `agent` handler
   (`capabilities.md:30`, "Cannot satisfy mandatory profile").
3. Have the hook call the **Phase 1 decision point** for its verdict so the block is decided by the
   same code that backs admission (connascence kept at the decision service, not duplicated in the
   hook); confirm a shared verdict evidence record ties the blocked harness action to the decision
   service.
4. Upgrade `code-harness-guide/references/capabilities.md`: change the `Tool before use` / `PreToolUse`
   row (line 24) and the matrix-identity `Evidence status` line (line 11) from documentation-verified to
   **runtime-verified**, recording the probe date and observed version, and rewrite the honest note
   (line 14, "native hook registration was not [exercised]") to reflect that registration was now
   exercised. Do **not** upgrade any row that was not actually exercised (e.g. `SessionStart`,
   `PostToolUse`, `PermissionRequest`, model/agent/http rows).
5. Add a repeatable **refresh probe script** bundled with the adapter under
   `code-harness-guide/` that re-registers the hook and reconfirms the block/allow pair on a clean
   checkout, per the matrix `Refresh trigger` (line 12).

## Acceptance criteria

- With the hook registered, a real Claude Code session is **denied** a write to a protected path (the
  write does not occur) and **allowed** a write to an unprotected path, captured in a reproducible
  transcript/log — not a simulation that invokes `guard.sh` directly (contrast `run-skeleton.sh:39`,
  `run_guard` → `bash "$GUARD"`).
- The blocked action produces a verdict evidence record emitted by the Phase 1 decision service, shared
  with the admission path (same decision logic, single source of truth).
- `capabilities.md` `Evidence status` (line 11) and the `Tool before use` row (line 24) accurately state
  runtime-verified registration with the probe date/version; a grep of `capabilities.md` for
  "documentation-verified" / "native hook registration was not" no longer describes the `PreToolUse`
  row, and no un-exercised row is upgraded.
- The refresh probe script reproduces the deny/allow pair on a clean checkout and exits non-zero if the
  block regresses.

## Dependencies

- **`governance-decision-point-fail-closed` (Phase 1, group 1) must land first.** The parent requires
  the hook to "call the **Phase 1 decision point** for its verdict, so the harness enforcement and the
  admission enforcement share one decision logic"; without the decision service there is no shared
  verdict to obtain or to record as evidence.

This subplan shares parallel group 2 with `a3-unattended-orchestration-runtime` (Phase 4): both depend
only on Phase 1 and touch disjoint files (this one edits skill/harness assets; Phase 4 edits operator
Go), so they run concurrently.
