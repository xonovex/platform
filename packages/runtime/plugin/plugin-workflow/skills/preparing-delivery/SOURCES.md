# Sources

## git-write-tree documentation

- **URL:** https://git-scm.com/docs/git-write-tree
- **Last reviewed:** 2026-10-27
- **Used for:**
  - `SKILL.md` → Gotchas
  - `references/verdict-matching.md` → The tree hash
  - `references/handoff.md` → The commit command
- **Aspects extracted:**
  - `git write-tree` creates a tree object from the current index and prints its name, so a filled temporary index gives the tree of the worktree → verdict matching, handoff

## Git Internals, Environment Variables

- **URL:** https://git-scm.com/book/en/v2/Git-Internals-Environment-Variables
- **Last reviewed:** 2026-10-27
- **Used for:**
  - `SKILL.md` → Gotchas
  - `references/verdict-matching.md` → The tree hash
- **Aspects extracted:**
  - `GIT_INDEX_FILE` sets the location of the index file, so a temporary index leaves the real one alone, and the variable must reach both `git add` and `git write-tree` → verdict matching, Gotchas

## Writing good change descriptions

- **URL:** https://google.github.io/eng-practices/review/developer/cl-descriptions.html
- **Last reviewed:** 2026-10-27
- **Used for:**
  - `references/delivery-files.md` → The commit message
  - `assets/templates/commit-message.md` → the body
- **Aspects extracted:**
  - A message opens with a short subject line, then a blank line, then a body that explains why the change is made rather than listing what the code does → delivery files, commit-message template

## Xonovex change-workflow conventions

- **Provenance:** Repository-original rules for the delivery text of one work item, built on the contract of **recording-handoff**: the three delivery files and their schemas, the delivery modes, verdict matching by tree hash, self-check labels, inline comment markers, the content scan, and the handoff of the commit to **using-git** and of the push, pull request and comments to **using-github** or **using-gitlab**
- **Last reviewed:** 2026-11-19
- **References:** references/delivery-files.md, references/pull-request-and-comments.md, references/delivery-modes.md, references/verdict-matching.md, references/content-scan.md, references/handoff.md
- **Aspects extracted:** The step runs the entry step of **recording-handoff** first; the delivery files carry the shared header with a skill actor and no `verified` entry and are replaced in place on a rerun; criteria are marked only from a verdict whose tree hash equals the worktree's; the report gives the commit command only under that same equality; the words follow **authoring-pull-requests** and the conventional commit rules of **using-git**; the step commits, pushes and opens nothing.
