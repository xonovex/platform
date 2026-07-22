# Sources

## Conventional Comments (standard)

- **URL:** https://gitlab.com/conventionalcomments/conventionalcomments.gitlab.io
- **Last reviewed:** 2026-07-22
- **Used for:** `references/conventional-comments.md` — label list, the `label (decoration): subject` format, the blocking / non-blocking / if-minor decorations, and the machine-parseable shape

## Conventional comments — practitioner guides

- **URL:** https://graphite.com/guides/conventional-comments
- **Last reviewed:** 2026-06-24
- **Used for:** corroborating the decoration semantics and the "label removes ambiguity" rationale
- **References:** references/conventional-comments.md

## Internal review practice (applied)

- **Provenance:** Repository-original review conventions distilled from maintained team practice
- **Last reviewed:** 2026-06-24
- **Used for:** the summary-plus-inline structure, cross-linking instead of ordinal references, verify-before-assert (read the branch, confirm component / API signatures via the relevant design-system or library skill before suggesting a fix), and the house writing style (no em-dash / semicolon / ellipsis in review prose) mined from a real Bitbucket PR review
- **Also used for:** the platform-independent findings pipeline — `references/findings-schema.md` (canonical findings shape, new-file hunk anchoring), `references/review-analyze.md` (diff → findings, effort dial, prior-findings comparison), `references/review-refine.md` (per-finding operations, label discipline) — distilled from the in-repo `pr-review-analyze` / `pr-review-refine` / `pr-review-post` command bodies; host delivery of these findings lives in the host skills (`github-guide` / `gitlab-guide`)

## Guide-level synthesis

- **Provenance:** Repository-original integration of the source blocks above; these references combine multiple inputs or maintained conventions rather than one exclusive upstream
- **References:** references/review-structure.md
- **Last reviewed:** 2026-06-24

## Refresh Workflow

1. Re-check the Conventional Comments GitLab repository for any change to the label set or decorations
2. Fold recurring review-craft lessons (from `reflect-guide` retrospectives) into the references
3. Bump **Last reviewed** above
