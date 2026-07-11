---
description: "Draft: formalize acceptance criteria as Given-When-Then scenarios from user stories and designs, via three-amigos discovery and example mapping"
allowed-tools:
  - Read
  - Write
  - Glob
  - Grep
  - Bash
  - Skill
argument-hint: "[stories-or-feature] [--out <dir>]"
---

# /xonovex-workflow:acceptance-formalize — Formalize acceptance criteria as Gherkin

## Arguments

- `[stories-or-feature]` (optional): Path/URL to refined stories, or a feature name (auto-detects `stories/<feature>/` or reads the prompt)
- `[--out <dir>]` (optional): Output directory for `.feature` files (default `acceptance/<feature>/`)

## Delegation

Load the `bdd-guide` skill (plugin `xonovex-skill-bdd`) — and `user-stories-guide` for
acceptance-criteria completeness — and follow them to turn the stories (and any provided
designs) into Given-When-Then scenarios: run three-amigos discovery (business /
development / testing perspectives), example-map each rule (yellow/green/red/blue cards),
and write the declarative scenarios each rule needs (often a couple, sometimes none) covering happy path, boundaries, errors, UI
states, and accessibility. Include non-functional acceptance criteria (performance,
security, load) alongside the functional ones where they matter — they feed the
acceptance sign-off. The skills are the source of truth for the method — do not restate
them.

Write `.feature` files to the output directory. If a work-item-tracker host skill is
installed (any `xonovex-skill-<host>` exposing issue/test-case sync), additionally attach
the scenarios to the tracker item; otherwise write to disk only and note which host skill
enables sync.
