# The pull request description and inline comments

Both files are written only in pull-request mode, under `deliveries/<work-id>/`. The header rules, the limits and the rerun rules are in [delivery-files.md](delivery-files.md).

## Contents

- The description
- Inline comments

## The description

Keys beyond the header: `work`, `title`, `base`, `head`, `verdict` and `verdict_matches`. `verdict` is the verdict's concept ID or `none`. `verdict_matches` is true only when the verdict's `worktree_hash` equals the current tree hash; see [verdict-matching.md](verdict-matching.md). The schema rejects `verdict: none` together with `verdict_matches: true`.

`base` is the branch the invoker names, else the work repository's default branch read with `git -C <worktree> symbolic-ref --short refs/remotes/origin/HEAD` with the `origin/` prefix removed. When neither answers, ask. `head` is the work item's branch.

The body is short and plain, under about 3,500 characters, in the description style of **authoring-pull-requests**, with these sections in order:

1. **Work item**: the work item's concept ID.
2. **Acceptance criteria**: the criteria from the work file, text unchanged. With a matching verdict each carries `(met)` or `(missing)`; a self-check verdict gives `(met, self-check)`. Without a matching verdict the criteria carry no mark and a closing line says no verdict matches this tree.
3. **Plan**: the plan link, by the rule in [delivery-files.md](delivery-files.md).
4. **Decisions**: each decision the work item cites, as its concept ID and a one-line claim taken from the decision's `# Decision` sentence. Omit the section when none is cited.
5. **How it was tested**: the matching independent verdict with the first 12 characters of its tree hash, and the checks from the implementation notes, each labelled self-check. A check the implementer ran is never described as validated.

## Inline comments

Keys beyond the header: `work`, `path`, `line`, `side` (`left` for removed code, `right` for added or changed code), `marker` (a stable kebab-case key) and `label` (`decision` or `note`). The file name is a random four-character suffix that uses only the characters `a-z` and `2-7`, so `0`, `1`, `8` and `9` are never used.

A comment exists only where a decision shaped code that a reviewer cannot understand from the code alone. The source is a `decision_notes` entry of the implementation notes, whose `location` gives the path and line. Check the line in the current tree before using it. The body gives the reason in one to three sentences and cites the decision ID for a `decision` label. A comment never restates what the code does and never says the code is correct, safe or complete.
