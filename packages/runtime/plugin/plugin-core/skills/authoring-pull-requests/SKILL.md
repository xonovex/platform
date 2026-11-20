---
name: authoring-pull-requests
description: "Use when an author writes or prepares their own pull or merge request for review: its title and description, its size and split, how it was tested, and a self-review before reviewers are assigned. Triggers on opening a PR or MR, PR descriptions or templates, large, atomic or stacked PRs, draft vs ready for review, or 'is my PR ready', even when the user doesn't say 'pull request'."
---

# Pull Request Authoring Guidelines

How to write a pull request that gets reviewed fast and correctly. This skill owns the author's craft and returns the title and description inline during Create. Host creation, updates, and work-item links require a separate Publish operation with the git host's tooling; reviewing a change for defects and writing reviewer feedback live in **reviewing-code**, and branch or commit work lives in **using-git**.

## Essentials

- **One concern, kept small** - one task per PR, aim 50-200 changed lines, see [references/size-and-atomicity.md](references/size-and-atomicity.md)
- **Lead with what / why / how** - scope, the goal, the approach, up front, see [references/description.md](references/description.md)
- **Show how you tested it** - environment, edge cases, how to reproduce, see [references/testing-evidence.md](references/testing-evidence.md)
- **Surface tradeoffs early** - state limitations and decisions before review, see [references/tradeoffs.md](references/tradeoffs.md)
- **Self-review before assigning** - read your own diff, confirm CI is green, see [references/self-review.md](references/self-review.md)
- **Use a lean template** - 4-6 prompts, link related PRs and work items, see [references/templates.md](references/templates.md)

## Gotchas

- "Tested locally" tells the reviewer nothing - name the scenario, the environment, and how they can re-run it
- Line count measures volume, not risk - a 20-line auth change needs more scrutiny than a 400-line generated-code or scaffolding diff, so call those out
- A refactor bundled into a feature PR hides the real change - split it into its own PR
- A PR merges only on the required approving review(s) plus green required status checks (branch protection) - never bypass the gate
- Drafting a PR is not authorization to create or update one on a host

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
