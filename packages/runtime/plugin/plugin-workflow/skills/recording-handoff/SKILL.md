---
name: recording-handoff
description: "Use when a workflow step reads or writes the directory of a change, plans/change or plans/family/change: the entry step that confirms the branch, validates inputs and checks preconditions; plan, decision, work item and log files; file names and IDs; frontmatter; generated indexes and logs; bundle validation; a new change; or the project's Workflow settings. Triggers on plan.md, decisions/, work/, log/, index.md, plan-bundle.py, Open Knowledge Format bundles, readiness or depends_on, even when the user doesn't say 'handoff'. Not for an architecture decision record, a commit, a tracker item or a README."
compatibility: "Python 3.12+ with uv runs the bundled validator and index generator; Git reports the branch of each repository."
allowed-tools: "Read Bash(uv:*) Bash(git:*) Bash(date:*)"
---

# Record and Hand Off a Change

Purpose: every workflow step leaves what it learned and decided in one directory of markdown files, so the next step, possibly another agent or person in another session, starts from files and never from memory. This skill owns the shared contract: the layout, the shared schemas and templates, the names, the write rules, the entry step, the project settings and the validator.

## Essentials

- A change is one bundle at `plans/<change>/`, or at `plans/<family>/<change>/` when a family directory with a `roadmap.md` groups related changes (the `plans/` name is a project setting). Each file is markdown with YAML frontmatter, and the directory is an Open Knowledge Format v0.2 bundle. A file's concept ID is its path without `.md`, and there is no `id` key.
- Run the entry step before any other work: confirm the branch of each repository the step writes to, validate the files it reads, read the generated indexes, check its own preconditions. The next section lists the four moves and [references/entry-step.md](references/entry-step.md) has the detail.
- Decisions, research, critiques and log entries are append-only: a change is a new file. Work files and `plan.md` are edited in place, and `index.md` and `log.md` only by the script.
- Names are `<slug>-<suffix>` with a random four-character suffix. A file is never moved or renamed, and a display number never appears in a name or in frontmatter.
- Readers derive trust from the `human:` actor prefix: write it only for content a person wrote or confirmed in this session, and never invent the id.
- Every edit keeps unknown frontmatter keys and updates `generated.at`. Every run that changes the bundle writes one log entry. Never commit or stage.

## The entry step

Run it first, in every workflow step, and report the outcome in a few lines.

1. For each repository the step writes to, name the branch and the uncommitted files, and proceed on the branch the invoker confirms. The invocation confirms it by naming the branch or by saying to work on the current one; authorization to write files does not. Until the answer arrives the step only reads, and an unattended run stops after the read-only work and reports the question. A repository the step only reads needs no confirmation.
2. Validate the files the step reads with `uv run scripts/plan-bundle.py validate plans/<change>`, or follow the schemas by hand when the script is not available, and report every finding.
3. Read the root `index.md`, then the index of each directory the step needs. Regenerate a stale index in the working tree.
4. Check the step's own preconditions: the files it needs exist. If one is missing, report it and stop. Never assume an earlier step ran and never start another step.

There is no drift check against repository heads, no revision counter and no stamp on copies.

## Start a change

When the invoker names a new change and no bundle exists:

1. Run the branch confirmation of the entry step.
2. Create `plans/<change>/`, or `plans/<family>/<change>/` when the invoker names a family, with a lowercase kebab-case name and write one log entry from [assets/templates/log-entry.md](assets/templates/log-entry.md) with `action: Initialization`.
3. Run `index --write` and then `log --write`, then `validate`, and report the result.

Starting a change creates no other file. The first step after it writes the plan or the research.

## Shared versus owned

- A schema and a template live here only when more than one skill writes that kind of document: plan, decision, work item and log entry. Any other type has its schema and template in its owner skill, and this skill only lists it.
- Only the owner skill of a type creates files of that type. Only **writing-plan** and **revising-plan** change `plan.md`.
- A skill that needs a new key on a shared type adds it here, because the key must be valid for every writer.

## Available scripts

- `scripts/plan-bundle.py validate <bundle-dir>`: three layers of checks (conformance, schemas, cross-file). Exit 0 clean, 1 findings with errors, 2 usage error.
- `scripts/plan-bundle.py index <bundle-dir> --check|--write` and `log <bundle-dir> --check|--write`: the only writers of `index.md` and `log.md`. The output is deterministic. Given a family directory, each command runs on every bundle of the family.
- Options: `--schemas <dir>` (repeatable) for extra schema directories, and `--format text|json`. Python 3.12 or later with the standard library only; run it with `uv run`. [references/validator.md](references/validator.md) has the output format and the fallback.

## Gotchas

- Editing a decision to change it. A change of mind is a new decision that supersedes the old one, and the old file gets only `status: deprecated` and `superseded_by`.
- Writing `human:<id>` for content nobody confirmed, or adding `verified` to content the skill wrote itself.
- A date or a datetime without a UTC offset. Every timestamp needs an explicit offset such as `+01:00` or `Z`.
- Inventing, rounding or copying a timestamp. Take every one from `date -u +%Y-%m-%dT%H:%M:%SZ`, and a name suffix only from `a-z` and `2-7`.
- Treating permission to write as a branch confirmation. Until the invoker names the branch or says to use the current one, do only read-only work.
- Editing `index.md` or `log.md` by hand, or leaving them stale after adding a file. Regenerate with the script.
- Copying decision text or work item definitions into `plan.md`. The plan links to them inside its generated block.
- Dropping a frontmatter key the schema does not name. Unknown keys are valid and must survive every edit.
- Treating a missing file as a sign the earlier step was skipped, then running that step. Report the gap and stop.
- A relation key that holds a file name, a title or a display number. Use the concept ID without `.md` and without a leading slash.
- Reading a missing schema warning as a pass. A Workflow document whose schema file cannot be found is only warned about, so check the type's owner skill.

## Progressive Disclosure

- Read [references/entry-step.md](references/entry-step.md) - Load when starting any step, to report the branch, the findings and a missing precondition correctly.
- Read [references/layout.md](references/layout.md) - Load when you need the directory layout, the family level, the owner and schema of each type, or who may edit which file.
- Read [references/frontmatter.md](references/frontmatter.md) - Load when writing or editing frontmatter: the common header, actors, timestamps, sources, relation keys, status.
- Read [references/ids-and-write-rules.md](references/ids-and-write-rules.md) - Load when naming a file, deriving a slug, deciding whether a file may be edited, or superseding a decision.
- Read [references/generated-files.md](references/generated-files.md) - Load when an index or the log looks stale, or when you need the grouping and ordering the generator applies.
- Read [references/validator.md](references/validator.md) - Load when running the script, reading its findings, or working without it.
- Read [references/settings.md](references/settings.md) - Load when reading or setting up the project's `Workflow` settings, or when a step needs the plans directory, repositories, delivery mode or phase skills.
- Read [assets/templates/](assets/templates/) - Load when creating a plan, decision, work item or log entry, one template for each type.
- Read [assets/schemas/](assets/schemas/) - Load when the validator reports a schema finding, or when adding a key to a shared type.
