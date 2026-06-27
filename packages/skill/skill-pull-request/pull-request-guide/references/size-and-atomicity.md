# size-and-atomicity: Keep PRs Small and Single-Concern

## Guideline

Each PR should do one thing and stay small - aim for 50-200 changed lines, treat ~400 as a soft cap and split above it.

## Rationale

Small PRs are reviewed faster, more thoroughly, merged more often, and are easier to revert. Large ones get superficial "LGTM" reviews and hide the reasoning behind individual changes. Keeping review time under roughly an hour is the real target that line counts approximate.

## How to Apply

1. Scope to one concern - one bug, one feature slice, one refactor, not several.
2. Split a refactor out of a feature or fix PR into its own PR.
3. If a change is large, break it into atomic, independently reviewable slices, or stack dependent PRs.
4. Limit files touched, not just lines - a wide diff is hard to follow even when small.
5. For changes that are legitimately big (scaffolding, generated code, dependency bumps, large renames), say so up front so the reviewer calibrates and does not skim the parts that matter.

## Example

```text
// Bad
One PR: new endpoint + a framework upgrade + reformat of 30 files.

// Good
PR 1: reformat (tooling only, no behaviour). PR 2: framework upgrade. PR 3: the new endpoint.
```

## Counter-Example

A single atomic change that cannot be split (one schema migration plus the code that depends on it) is fine when large - flag it and explain why it is one unit.

## Related

[self-review.md](./self-review.md)
