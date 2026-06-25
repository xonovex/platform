---
name: pull-request-guide
description: "Use when authoring a pull or merge request - writing the description, sizing and splitting the change, documenting how it was tested, surfacing tradeoffs, and getting it review-ready before assigning reviewers. Triggers on opening a PR/MR, PR descriptions or templates, large or atomic PRs, what to put in a PR, draft vs ready, stacked PRs, or self-review, even when the user doesn't say 'pull request'."
---

# Pull Request Authoring Guidelines

How to write a pull request that gets reviewed fast and correctly. This skill owns the author's craft. Host mechanics (creating the PR, linking work items) live in your git host's own tooling, the reviewer's feedback side lives in `code-review-guide`, and branch/commit work lives in `git-guide`.

## Essentials

- **One concern, kept small** - one task per PR, aim 50-200 changed lines, see [references/size-and-atomicity.md](references/size-and-atomicity.md)
- **Lead with what / why / how** - scope, the goal, the approach, up front, see [references/description.md](references/description.md)
- **Show how you tested it** - environment, edge cases, how to reproduce, see [references/testing-evidence.md](references/testing-evidence.md)
- **Surface tradeoffs early** - state limitations and decisions before review, see [references/tradeoffs.md](references/tradeoffs.md)
- **Self-review before assigning** - read your own diff, confirm CI is green, see [references/self-review.md](references/self-review.md)
- **Use a lean template** - 4-6 prompts, link related PRs and work items, see [references/templates.md](references/templates.md)

## Gotchas

- A clear description is not optional polish - it is the first signal a reviewer reads, and a missing one forces them to reverse-engineer intent from the diff
- "Tested locally" tells the reviewer nothing - name the scenario, the environment, and how they can re-run it
- Large diffs get rubber-stamped, not reviewed - one 600-line PR gets a worse review than three 200-line PRs
- Line count measures volume, not risk - a 20-line auth change needs more scrutiny than a 400-line generated-code or scaffolding diff, so call those out
- A refactor bundled into a feature PR hides the real change - split it into its own PR

## Example

A filled description for a small, focused change:

```markdown
## What

Cache the resolved feature-flag set per request instead of re-reading it on every check.

## Why

PROJ-1234: the flag store was hit ~40x per request, adding ~15ms p95. Memoize it for the request scope.

## Changes

- Add a request-scoped FlagCache, resolve once in the entry middleware.
- Replace direct store reads in isEnabled() with the cache.

## Testing

- Unit: new FlagCache hit/miss/expiry tests.
- Manual: ran the planner flow locally, p95 dropped 210ms -> 195ms (logs attached).

## Tradeoffs / risks

- Flags changed mid-request are not picked up until the next request. Acceptable, flags do not change within a request.

## Related PRs

Part of the flag-perf set: !123, !124.
```

## Progressive Disclosure

- Read [references/description.md](references/description.md) - Load when writing or reviewing the PR description text (what/why/how, length, links, screenshots)
- Read [references/size-and-atomicity.md](references/size-and-atomicity.md) - Load when a PR feels large, mixes concerns, or needs splitting
- Read [references/testing-evidence.md](references/testing-evidence.md) - Load when documenting how the change was verified
- Read [references/tradeoffs.md](references/tradeoffs.md) - Load when the change has limitations, risks, or non-obvious decisions to disclose
- Read [references/self-review.md](references/self-review.md) - Load when finishing a PR before requesting review
- Read [references/templates.md](references/templates.md) - Load when setting up a PR template or structuring a description
