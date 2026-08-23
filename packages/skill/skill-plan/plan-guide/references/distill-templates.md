# Distill Templates: Suite File Skeletons

The layout follows the portable Agent Skills convention: a directory per skill, `SKILL.md` with
`name`/`description` frontmatter, supporting files alongside, placed in whatever skills directory
the consuming harness reads.

## Suite layout

```
<skills-dir>/
  <topic>-map/                      # umbrella: map + router, NOT a runbook
    SKILL.md
    references/
      migration-map.md              # phases, dependencies, team split, patterns vs exceptions
      replay-plan.md                # ordered replay + ordering-rationale section
      invariants.md                 # rules + scaffolding inventory + keep-forever list
      pitfalls.md                   # symptom + mechanism + fix
      open-decisions.md             # gates, decisions with owners + recommendations, residue
      reference-commits.md          # HISTORY: what came from where (branch, commit ids)
      process-lessons.md            # HISTORY: how the source run went
  <topic>-<responsibility>/         # one per distinct responsibility
    SKILL.md
    references/
      <recipe>.md                   # code, idiom tables, worked before/after examples
      provenance.md                 # HISTORY: this skill's commits + run notes
```

## Granular SKILL.md skeleton

```markdown
---
name: <topic>-<responsibility>
description:
  "Use when <trigger>. <What it does in one sentence>. Not for <adjacent thing>.
  <Hard prerequisite if any>."
---

# <Title>

Purpose: <the outcome and why it matters, 2-4 lines; the invariant this step establishes or
preserves.>

Safety zone <n> (see the umbrella): <what protects the not-yet-migrated code, in one sentence.>

## Prerequisites

<Which skills/phases must exist first; which assumed dependencies are NOT real.>

## When not to use

<Shapes that look applicable and are not, each with the skill or pattern that owns them.>

## Detect current state

<tolerant greps: already done? partially done? absent? use 2>/dev/null; interpret each outcome>

## The recipe

<Numbered steps stated as intent, not filenames. Inline the essential code shapes; push long
tables and worked examples to references/. Name output artifacts this skill creates ("create X;
this skill owns the page").>

## Known failure modes

<Each: symptom, mechanism, mitigation. Timeless phrasing; counts as grep-backed expectations.>

## Temporary vs final

<What this step adds as scaffolding, with its removal criterion; what is final architecture.>

## Validation

<deterministic commands: tests, builds, lint, gates; prove a gate or test CAN fail once before
trusting it>

(History: `references/provenance.md`.)
```

## provenance.md skeleton (the only place history lives)

```markdown
# Provenance (optional; the skill does not depend on this)

Distilled from <branch> (merge base <commit>; may no longer exist). Archaeology only, via
`git show <commit>`.

- `<commit>` - <what it contributed to this skill>
- `<commit>` - <...; flag content that exists ONLY at a commit because it was later folded away>

Run notes: <the anecdotes stripped from the working text: reverts, failed attempts, false
dependencies, counts at distillation time>. <Plan-record paths, marked excluded if so.>
```
