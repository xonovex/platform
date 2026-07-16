# size-and-atomicity: Keep PRs Small and Single-Concern

One PR does one thing - one bug, one feature slice, or one refactor. Aim for 50-200 changed lines, treat ~400 as a soft cap and split above it; the real target is review time under ~1 hour. Split a refactor out of a feature or fix PR, or stack dependent PRs. Limit files touched, not just lines.

Exception: a single atomic change that cannot be split (a schema migration plus the code that depends on it) is fine when large - flag it and say why it is one unit. For legitimately big diffs (scaffolding, generated code, dependency bumps, large renames), say so up front.

```text
// Bad: one PR: new endpoint + framework upgrade + reformat of 30 files.
// Good: PR 1 reformat (tooling only). PR 2 framework upgrade. PR 3 the new endpoint.
```

## Related

[self-review.md](./self-review.md)
