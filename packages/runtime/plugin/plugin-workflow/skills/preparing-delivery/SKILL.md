---
name: preparing-delivery
description: "Use when the implementation of one work item is done and the words that deliver it must be written: the commit message, the pull request description and the inline comments where a decision shaped non-obvious code. Reads the work file, the implementation notes, the worktree diff, the verdict, the plan and the cited decisions, marks acceptance criteria met or missing only when the verdict's tree hash equals the worktree's current hash, and saves the files under the deliveries directory of the work item without applying anything. Triggers on 'write the commit message', 'describe this change for review', 'prepare the pull request', 'add comments for the reviewer', even when the user doesn't say 'delivery'. Not for committing, pushing or opening the pull request, validating the work, or implementing it."
---

# Prepare a Delivery

Purpose: turn a finished work item into the text a reviewer and the history need, with the claims bound to what was actually checked. The output is files in the change's plan directory. The skill applies nothing: the commit goes to **using-git**, and a push, a pull request and inline comments go to **using-github** or **using-gitlab**, each only when the user asks.

**Load first:** the **recording-handoff** skill, and run its entry step before anything below.

## Essentials

- One work item per run. Files go to `deliveries/<work-id>/`: `commit-message.md` always, `pull-request.md` only in pull-request mode, `comments/<suffix>.md` optionally in that mode.
- A criterion is marked met or missing only when the verdict's tree hash equals the worktree's current tree hash. Otherwise the criteria appear unmarked, with a line saying no verdict matches. The implementer's own checks are labelled `self-check` and never called validated.
- The report gives the commit command only when the worktree's tree hash equals the tree hash the verdict records.
- The words follow **authoring-pull-requests** for the description and the conventional commit rules of **using-git** for the message.
- A comment is written only where a decision shaped code the reader cannot understand from the code. It never restates the code and never says the code is correct. Each carries a stable `marker`, so a rerun updates it.
- The pull request description stays under about 3,500 characters.
- Every file is scanned for secrets and for anything the plan marks sensitive before the run reports done.
- Never commit, stage, push or open anything. Never start another step.

## Procedure

1. **Run the entry step** from **recording-handoff**. The delivery files go into the plan directory's repository, so confirm that branch. When the plan directory and the work repository are the same repository, confirm once for the main checkout. The work repository is only read and needs no confirmation. The preconditions: the work file, its implementation notes (they name the worktree and the branch), and the project's `delivery` setting. A missing one is reported and the run stops; a missing `delivery` setting is a blocker. A run that stops without changing anything writes no log entry. Never run **implementing-work-item** or **validating-work-item** to fill the gap.
2. **Load the skills that apply** from the skills map that the settings reference of **recording-handoff** defines, for the phase **preparing-delivery**: the `skills_to_consult` of the plan and the work item, a custom workflow skill the project provides, the **preparing-delivery** entry of the project's `skills` map, and the default skills of the phase when the project's map does not name it. The commit convention comes from **using-git** and the project's instructions, and the description conventions from **authoring-pull-requests**.
3. **Read the inputs.** The work file (criteria, decisions, repository), the plan, each decision the work item cites, the implementation notes (branch, worktree, `decision_notes`, `checks`) and the verdict file if one exists. Read the delivery mode from [references/delivery-modes.md](references/delivery-modes.md).
4. **Fix the tree.** Compute the current tree hash of the worktree and the diff it stands for. An empty diff is reported and the run stops. See [references/verdict-matching.md](references/verdict-matching.md) for the verified command.
5. **Match the verdict.** Compare the verdict's `worktree_hash` with the current hash and set `verdict` and `verdict_matches`. Marks come only from a matching verdict.
6. **Write the files** from the templates: [assets/templates/commit-message.md](assets/templates/commit-message.md), then, in pull-request mode, [assets/templates/pull-request.md](assets/templates/pull-request.md) and, for each decision that shaped non-obvious code, [assets/templates/inline-comment.md](assets/templates/inline-comment.md). Announce each write in one line. On a rerun, replace in place and match comments by `marker`. See [references/delivery-files.md](references/delivery-files.md) and [references/pull-request-and-comments.md](references/pull-request-and-comments.md).
7. **Scan and check.** Scan all written content and the diff for secrets and for what the plan marks sensitive, count the description's characters, then validate the bundle with the **recording-handoff** script. See [references/content-scan.md](references/content-scan.md).
8. **Record and report.** Write one `Workflow Log Entry` (`skill: preparing-delivery`; `action: Creation` when the run created a delivery file, `Update` when it only changed existing ones; the delivery files in `touched`, never the log entry itself), regenerate the indexes and the log, and report the files, the tree hash, whether the verdict matched, and the scan result. Hand off as [references/handoff.md](references/handoff.md) shows: the commit command for **using-git** only when the hashes are equal, and the files that **using-github** or **using-gitlab** take as the text of the pull request and of each inline comment. Offer a fresh **validating-work-item** run when no verdict matches. Never start any of them.

## When not to use

- Committing: **using-git**. Pushing, opening the pull request or merge request, or commenting on the host: **using-github** or **using-gitlab**, when the user asks.
- Judging whether the work meets its criteria: **validating-work-item** writes the verdict this skill reads.
- Writing the code or the notes that feed this skill: **implementing-work-item**.
- A change with no work item file: write the message with **using-git** and the description with **authoring-pull-requests**.

## Gotchas

- Setting the temporary index variable for `git add` only. In `GIT_INDEX_FILE=x git add -A && git write-tree` the second command reads the real index and prints a hash that is not the worktree's. Set it for both commands.
- Marking criteria from a verdict that checked earlier content. Compare hashes as strings and mark nothing when they differ.
- Giving the commit command when the verdict's hash differs from the worktree's. The commit would carry content no verdict checked.
- Showing the implementer's checks as validated. They are `self-check`, in the description and in the notes.
- Copying the criteria into the commit message, or the commit's reasoning into the description. The message keeps the lasting reasoning, the description keeps the review view.
- Inventing the reason behind a decision. Quote the decision's own claim and cite its ID.
- Comments that narrate the code or vouch for it. Only the reason that the code cannot show belongs in a comment.
- Two comment files with one `marker`, or a changed marker on a rerun. Both duplicate the comment on the host.
- Writing `pull-request.md` in trunk or branch mode, or dropping decision notes in those modes. The decisions go into the commit message body instead.
- A description that grows past the limit. Shorten the detail, never the criteria.
- Printing a secret to point at it. Report the path and line only.
- A plan link that is a file URL or a bare repository URL. Follow the plan link rule in [references/delivery-files.md](references/delivery-files.md).
- Inventing a time, a version or a file suffix. The clock, the plugin's manifest and the alphabet `a-z` and `2-7` supply them.

## Done when

- `commit-message.md` exists with a subject line, the decision IDs and the plan link, and the mode's other files exist only where the mode calls for them.
- `verdict_matches` agrees with a fresh hash comparison, and no criterion carries a mark unless it is true.
- The description is under about 3,500 characters, and every comment file has a `marker`, a `path`, a `line` and a `side`.
- The scan found nothing, or the run reported the finding and did not report done.
- The report holds the commit command only when the hashes are equal, and names the files the host skill takes.
- Every timestamp comes from `date -u +%Y-%m-%dT%H:%M:%SZ`, and `generated.by` carries the version of the plugin that ships this skill, read from the plugin's manifest; the full rule is in the frontmatter reference of **recording-handoff**.
- The bundle validates with no error and no schema-not-found warning for the three types, and the log entry exists.

## Progressive Disclosure

- Read [references/delivery-files.md](references/delivery-files.md) - Load when writing the commit message, the description or a comment, or when checking the limits and the rerun rules.
- Read [references/pull-request-and-comments.md](references/pull-request-and-comments.md) - Load when writing the pull request description or an inline comment: its keys, sections, marks and sources.
- Read [references/delivery-modes.md](references/delivery-modes.md) - Load when reading the `delivery` setting or deciding which files the mode calls for.
- Read [references/verdict-matching.md](references/verdict-matching.md) - Load when computing the tree hash, comparing it with a verdict, or labelling a self-check.
- Read [references/content-scan.md](references/content-scan.md) - Load when scanning the written content and the diff for secrets and sensitive terms.
- Read [references/handoff.md](references/handoff.md) - Load when writing the report: the commit command, its tree hash condition, and the files the host skill takes.
- Read [assets/templates/commit-message.md](assets/templates/commit-message.md) - Load when writing the commit message.
- Read [assets/templates/pull-request.md](assets/templates/pull-request.md) - Load when writing the pull request description.
- Read [assets/templates/inline-comment.md](assets/templates/inline-comment.md) - Load when writing an inline comment.
- Read [assets/schemas/](assets/schemas/) - Load when the validator reports a schema finding on a delivery file.
