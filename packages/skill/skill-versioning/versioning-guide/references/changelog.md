# changelog: Generating Changelog Entries

Turn the commits behind a version bump into a changelog entry, prepended newest-first.

## Pick One Convention — and Be Faithful

- **Keep a Changelog** — a `# Changelog` title with `## [x.y.z] - YYYY-MM-DD` sections, an `## [Unreleased]` area, and six change groups: Added, Changed, Deprecated, Removed, Fixed, Security. Human-curated — do **not** paste raw git logs.
- **Changesets-style** — a `# <package-name>` title with bare `## x.y.z` sections and `### Major Changes` / `### Minor Changes` / `### Patch Changes` groups, each entry a commit/PR-linked bullet. A single release can carry all three groups because each change declares its own level.
- Choose the one the repo already uses; don't blend them.

## Deriving Entries From Commits

- Parse each commit header as a conventional commit `type(scope)!: description`. The `git-guide` skill owns the commit **format**; this is the read side.
- Bump mapping: `feat` → minor, `fix` → patch, a breaking change → major. A breaking change overrides the type (`fix!:` is major, not patch).
- **Detect breaking changes from both signals**: a `!` before the `:` in the header _and_ a `BREAKING CHANGE:` / `BREAKING-CHANGE:` footer in the body. A header-only match drops `feat!:` entirely — scan the body for the footer too.
- An include-set (e.g. `feat,fix,refactor,perf,docs`) is a fine curated filter for _which entries to list_, but check for breaking changes **before** filtering, or a breaking `chore!:` vanishes.

## Bullet Format (Changesets-style)

- Authentic form: ``- [#PR](…/pull/PR) [`hash`](…/commit/hash) Thanks [@login](https://github.com/login)! - <description>``.
- The author is the **PR-author login**, not the git commit author name — a display name like `Jane Doe` produces a 404 profile link. Resolve the handle, or omit the author link.
- Group dependency updates under the matching level as their own bullets, e.g. a `- Updated dependencies` line with nested `  - name@version` sub-bullets.

## Prepend, Don't Append

- Insert the new entry directly under the title so the newest version is first; create the file with the title if it does not exist.

## Gotchas

- Deriving one heading level from the version delta files every commit under a single group — a trivial `docs:` can land under `### Major Changes`. Bucket by each change's declared level when you can.
- A placeholder like `- Version bump` belongs to no convention — prefer an honest "no user-facing changes" note only when the entry is truly empty.
- "Don't dump git logs" is the whole point of a changelog — summarize intent, not every commit subject.
