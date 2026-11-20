# Generated files

## Contents

- Rules for all generated files
- Root index
- Directory indexes
- Decisions index
- Work index
- Log

## Rules for all generated files

- `scripts/plan-bundle.py` is the only writer of `index.md` and `log.md`. `--check` reports a stale or missing file and `--write` rewrites it. Same bundle, same bytes: the output depends on file contents only, never on file system times, the clock or the platform.
- An `index.md` has no frontmatter. Only the root index holds `okf_version: "0.2"`.
- The body is one or more `# Heading` groups. Each entry is `* [Title](<relative-url>) - description`, using the `title` and `description` of the concept. A subdirectory entry is `* [name](<name>/) - description`. The file ends with one newline.
- The creation time of a file is its `generated.at`. Where two entries tie on every ordering rule below, the concept ID in ascending order decides.
- Every directory except `diagrams/` gets an index. A directory with no concept documents and no subdirectory has no index.

## Root index

Frontmatter holds only `okf_version: "0.2"`. A `# Documents` group lists `plan.md` and then `close-out.md` when it exists. A `# Directories` group lists the subdirectories that have an index, alphabetically, with the descriptions below.

## Directory indexes

A directory without special rules has one group of its concept files, ordered by `generated.at` newest first, and a `# Directories` group for subdirectories, alphabetically.

| Directory | Group heading | Description of the directory |
| --- | --- | --- |
| `research/` | `# Research` | Research behind the plan |
| `critiques/` | `# Critiques` | Adversarial reviews of the plan |
| `verdicts/` | `# Verdicts` | Validation verdicts for each work item |
| `log/` | `# Log entries` | One entry for each skill run |
| `implementation/` | `# Directories` | Implementation notes for each work item |
| `deliveries/` | `# Directories` | Delivery text for each work item |
| `decisions/` | see below | Decisions that shape the plan |
| `work/` | see below | Work packages and work items |

A directory named for a work item (`implementation/<work-id>/`, `deliveries/<work-id>/`) is described by the work item's `title`. Inside `deliveries/<work-id>/`, the commit message and pull request description sit under `# Delivery` and `comments/` under `# Directories`.

## Decisions index

Titles carry a derived display number: `* [1. Title](<file>.md) - description`. Numbers run from 1 across the whole directory in the order the groups and their entries are written, so a decision that moves to another group takes a new number and no file or frontmatter ever holds one.

Groups, in this order, by `state` and `status`:

1. `# Settled`: state `settled`, status not `deprecated`.
2. `# Open`: state `open`.
3. `# Paused`: state `paused`.
4. `# Partly settled`: state `partly-settled`.
5. `# Superseded`: status `deprecated`, whatever the state.

A group with no members is left out. Within a group, entries sort by this key, each part ascending:

1. `generated.at`.
2. The concept ID.

## Work index

Two groups in this order: `# Ready now` (readiness `ready`) and `# Waiting on a decision` (readiness `waiting`). Each group is in dependency order: a topological order by `depends_on`, so a work file follows the files it depends on, with ties by creation time. A `depends_on` cycle is a validation error and the generator reports it instead of writing. A dependency in the other group does not move an entry between groups.

## Log

`log.md` sits at the bundle root. It starts with `# Directory Update Log`, then one `## YYYY-MM-DD` heading for each date, newest first. The date is the date part of `generated.at` of the entry, as written.

Under a heading, each entry is one line. Entries of the same date keep the order of their `generated.at`, newest first, compared as instants, with the concept ID in ascending order as the tie-break:

```text
* **Creation**: One sentence from the entry body. [<date>-<skill>-<suffix>](/log/<date>-<skill>-<suffix>.md)
```

The bold word is the entry's `action`, the sentence is the first paragraph of its body, up to the first blank line and joined into one line, and the link text is the file name without `.md`. Entries come from the files in `log/`; the file `log/index.md` is not an entry.
