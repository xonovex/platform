# from-a-session: Fold the Lessons of a Session into AGENTS.md

Capture what a session taught and write each project lesson as a rule in the nearest AGENTS.md. Reusable domain guidance goes to its owning skill through **authoring-skills**.

## Inputs

- **Category** (optional): one lesson category, such as testing; the default takes every lesson of the session
- **From reflections** (optional): read stored `reflections/*.md` files in place of the session
- **Persist** (optional): also store each lesson as `reflections/<topic>.md` for an audit trail; off by default
- **Instruction file** (optional): the target AGENTS.md; detected from `applies_to` when absent
- **Preview** (optional): show the result without a write

## Extract

Keep only actual mistakes, discoveries and corrections of the session that a future agent must remember and that apply beyond the current task; never a hypothetical, and never a step that was correct in its context. Give each lesson a `category` (the technology or domain), a `topic` slug and an `applies_to` list of specific routing keys: directory, package or domain names. Store a lesson only with the persist input; direct integration is the default.

```markdown
---
category: testing
topic: json-response-type-safety
applies_to:
  - API testing
created: "2026-01-07"
applied: false
---

# Testing: untyped JSON response in a route test

- **MISTAKE**: What went wrong
- **DISCOVERY**: How the mistake was found
- **FIX**: How to avoid it
```

## Route to one owner

- A small project-specific lesson goes to the nearest AGENTS.md, by the steps below
- Reusable domain guidance goes to the existing skill that owns the domain, through **authoring-skills**
- Organization policy goes to its policy owner, and executable behavior to its module owner

A direct write changes the working tree only. It does not publish, install, enable, enforce or promote the result for an organization.

## Integrate into AGENTS.md

1. Group the lessons by topic, and find the target: the named file, else the nearest AGENTS.md whose directory or package matches `applies_to`; ask when two match.
2. Write each lesson as one or two rules in the style of the file, merged into a related group or appended as a new one.
3. Remove each duplicate of an existing rule and each rule that a reader can infer from the code; never remove an unrelated rule.
4. After the write, set `applied: true` on each stored lesson.

## Govern a managed or executable target

When the target is managed, organization-wide, executable, enforcing, configuration-changing or privileged, gather the candidates also from lifecycle results, policy denials, incidents, exceptions, drift, rollbacks and support cases, with the exact source, scope, observed outcome, confidence and expiry. Check other explanations, data quality and recurrence before a causal lesson. Then:

1. Preview the exact content, scope, permissions, conflicts, failure behavior and rollback.
2. Get the review and the exact-scope authorization of the target owner.
3. Version the target and keep the provenance from candidate to decision to version.
4. Validate, and canary representative contexts where the target is executable.
5. Measure the intended outcome, false positives, bypasses, user friction and privacy.
6. Promote in stages, verify the effective state, or roll back.

Never promote one denial, incident, exception, drift event, metric or reflection automatically into global instructions or enforcement. Record the candidate, source, owner, version, scope, reviewer, decision, validation, canary, measures, rollback path and expiry; a rejected or rolled-back candidate stays evidence and does not return to the queue silently.

## Gotchas

- `applies_to: ["general"]` routes to the root AGENTS.md or nowhere: give a specific key
- A lesson that restates the code is filler
- A lesson marked `applied: true` before its write is lost
