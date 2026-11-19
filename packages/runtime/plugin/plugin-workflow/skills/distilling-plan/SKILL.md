---
name: distilling-plan
description: "Use when turning a completed plan's implementation (its branch, plan record, and execution logs) into a replayable suite of skills. Triggers on a replay playbook for finished work, mining git history and session logs for what the record omits, an umbrella map skill with granular skills, a temporary scaffolding inventory, and reproducing a migration on a clean base branch, even when the user doesn't say 'distill' and asks to make last year's upgrade repeatable."
---

# Distill Completed Work into a Skill Suite

Distill a finished body of work (a branch with commits, its plan record, its execution logs) into an umbrella map skill plus granular skills that make the work reproducible as a controlled, incremental replay onto a clean base branch. The original is an oracle, never something to merge wholesale. Unlike the read-only planning skills, authoring the suite writes skill files; that is the output artifact of this skill.

## Inputs

- **Source** (optional): the completed plan document or branch to distill; detected from the git configuration when absent
- **Skills directory** (optional): the target directory for the suite; defaults to the skills directory of the project

## Core Workflow

1. **Scope the sources, cheap first.** Establish the branch, merge base, and base drift; read the plan record (a follow-up record from **closing-plan**, when one exists, is exactly this) personally and fully before delegating; locate the execution logs of whichever harness ran the work plus CI runs and PR threads; identify docs the work produced (`git cat-file -e <base>:<path>` decides whether each is a worked example or an output to create). Logs are huge: establish the format with a `head`, then sample with jq/grep/python, never read wholesale. Per-source recipes and traps: read [references/sources.md](references/sources.md) when mining any source.
2. **Reconstruct by parallel fan-out.** One agent per work tier (foundation, the repeated per-module pattern, cleanup and late fixes) plus a log miner hunting what the written record omits: failed and superseded approaches, reverts, mid-run plan changes, operator steering, tooling pain. Require from every agent, per piece: WHAT (paths, signatures, 10-30 line snippets), WHY (intent, invariants, rejected alternatives), ORDER (including disproven dependencies), TEMPORARY vs FINAL (with removal criteria), FAILURE MODES (mechanisms, not symptoms), VALIDATION (exact commands). When sources disagree, investigate: git wins on WHAT, logs on WHY and mechanism, the record on intent and decisions.
3. **Synthesize the map before authoring.** Produce, as the umbrella's references: the phase map with dependency graph and team split; invariants plus the temporary-scaffolding inventory with removal criteria; a safety model classifying every change by what protects unmigrated code (reversible via a flag, live-on-merge with tests as the only protection, or a deliberate recorded behaviour change); pitfalls as symptom + mechanism + fix; the open-decisions register (gates vs decisions vs residue); process lessons; and a replay plan whose ordering improves on the original where evidence supports it (late fixes into the foundation, disproven dependencies dropped, mid-run hardening from day one).
4. **Author the suite.** Umbrella named `<topic>-map` (a router, not a runbook); granular skills with verb-led names, boundaries from the reconstructed work, split by who applies them and by decision vs recipe. Each skill: routing description, prerequisites, tolerant detect-current-state greps, the recipe as intent with essential code inline, failure modes with mechanisms, temporary-vs-final, deterministic validation; read [references/templates.md](references/templates.md) when writing the suite's files. Hygiene: self-containment (the suite works with the source branch deleted and the record archived); provenance quarantine (branch names, commit ids, and run narrative live only in per-skill provenance files; working text is timeless and normative, with recorded counts as grep-backed expectations); preserve ungeneralizable exceptions by name and carry unresolved designs as explicit do-not-improvise entries.
5. **Verify mechanically.** Frontmatter names match directories; commit ids and branch names grep clean outside provenance files; every cross-reference resolves; every validation command is runnable; a cold reader can answer, from the umbrella alone, which phase the tree is in and what to do first. Plant one violation to prove each sweep can fail before trusting its green.

## Boundaries

The plan record that this skill mines is richest when **closing-plan** wrote it. Lessons from one session, in place of a whole body of work, belong to **reflecting-on-sessions**.

## Gotchas

- Content that exists only at a commit (an exemplar later folded away) is invisible in the final tree; capture it inline or a squash-merge loses it.
- Docs the work created are outputs with one owning skill ("create if absent"), never ambiguous inputs.
- A "missing infrastructure" blocker in the record is a hypothesis until someone checks what the library already ships.
- When a skill-authoring skill is installed, apply its spec constraints and validation during authoring; route to it by installed name.

## Progressive Disclosure

- Read [references/sources.md](references/sources.md) - Load when mining a distillation source: git history, harness session logs, or CI and review records
- Read [references/templates.md](references/templates.md) - Load when writing the files of a distilled suite (layout, granular skill skeleton, provenance skeleton)
