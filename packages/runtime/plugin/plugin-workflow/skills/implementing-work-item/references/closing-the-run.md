# Closing the run

The run ends by writing the notes file of [notes-file.md](notes-file.md), its log entry and the indexes, and then it reports and stops.

## Contents

- Writing and closing the run
- A rerun

## Writing and closing the run

1. Write or update the notes file in the main checkout's bundle and announce the write in one line: the path, the number of tasks and the number of decision notes. Never write it into the worktree's tracked copy of `plans/`. Set `generated.by` to `implementing-work-item/<version>`, where `<version>` is the `version` field of the `plugin.json` of the plugin that ships this skill (read it, do not guess), and `generated.at` to the output of `date -u +%Y-%m-%dT%H:%M:%SZ`. Create the `implementation/<work-id>/` directory when it is missing; its generated index is regenerated with the script.
2. Write one `Workflow Log Entry` from the template of **recording-handoff**, in the same bundle, when the run changed the bundle (a run that stopped at a precondition wrote nothing and writes no entry). Set `skill: implementing-work-item`, `action: Creation` when the run created the notes file and `Update` on a rerun, and `touched` to the notes concept ID, `implementation/<work-id>/notes`, never the log entry itself. The entry name is `<date>-implementing-work-item-<suffix>` with a random four-character suffix from `a-z` and `2-7`, and `<date>` from the clock. A string value that contains a colon followed by a space, a quote or a leading special character is written in single quotes. The authorization key is only for a run that applied something outward, and this one does not, except to record the invoker's word that a blocker is done.
3. Regenerate the indexes and the log with the **recording-handoff** script, then validate the bundle and report any finding.
4. Do not commit or stage anything, in the worktree or in the plan directory's repository.
5. Report inline, outcome first: what was implemented, the worktree and branch, the criteria by state, the open points. Offer the next steps, **validating-work-item** and then **preparing-delivery**, and stop.

## A rerun

The notes file is edited in place, because git keeps its history. Keep the `tasks` of earlier runs that still describe the code, add the new ones, and rewrite `checks` to the checks of the latest run. Remove a decision note whose location no longer holds. Update `generated.at`. Preserve unknown frontmatter keys.
