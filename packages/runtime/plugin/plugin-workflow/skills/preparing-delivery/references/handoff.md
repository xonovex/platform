# Handing off the delivery

The skill applies nothing. Its report hands the delivery files to the skills that apply them, each only when the user asks: the commit to **using-git**, and a push, a pull request or merge request and inline comments to **using-github** or **using-gitlab**, chosen by the `host` of the repository in the project settings.

## Contents

- The tree hash condition
- The commit command
- What the host skill takes
- Rules the report states

## The tree hash condition

The report gives the commit command only when the worktree's current tree hash equals the `worktree_hash` of the verdict, as [verdict-matching.md](verdict-matching.md) computes and compares them. A commit made then holds exactly the content the verdict checked.

When the hashes differ or no verdict exists, the report gives no commit command. It says that the worktree changed after the verdict, or that no verdict exists, and offers a fresh **validating-work-item** run. Never adjust a hash or reuse a command from an earlier run to get past it.

## The commit command

Give the commands for the work item's worktree, with the tree hash of this run in place of `<tree-hash>`:

1. Compute the tree hash again just before the commit. When it differs from `<tree-hash>`, the worktree changed after this run: commit nothing and run this skill again.
2. Extract the message body without its frontmatter into a file outside the repository:

   ```bash
   awk 'n>=2{print; next} /^---$/{n++}' <bundle>/deliveries/<work-id>/commit-message.md > <tmp>/message.txt
   ```

3. Commit with the project's hooks running; never pass `--no-verify`:

   ```bash
   git add -A && git commit -F <tmp>/message.txt
   ```

   `-F` keeps lines that start with `#`, so a heading in the message survives.

4. Check that `git rev-parse 'HEAD^{tree}'` prints `<tree-hash>`. When a hook changed the content, they differ: push nothing and report it.

The commit happens only in the work item's worktree and never includes the plan files.

## What the host skill takes

Name these files of `deliveries/<work-id>/` in the report, and only in pull-request mode:

| File | What the host skill takes from it |
| --- | --- |
| `pull-request.md` | The `title`, `base` and `head` from the frontmatter, and the body without its frontmatter as the description of the pull request or merge request. |
| `comments/<suffix>.md` | One inline comment each: the body without its frontmatter as the text, anchored to `path`, `line` and `side`, with its `marker` so a later run updates the comment and never posts a second one. |
| `comments/<suffix>.md` with `status: deprecated` | An update of the comment with that marker to one line saying the note no longer applies. The comment is never deleted. |

In trunk and branch mode the report names only `commit-message.md`.

## Rules the report states

- A push is a plain fast-forward push: no `--force`, no `--force-with-lease` and no `+` in the refspec. When the remote branch has commits the worktree lacks, stop; never rebase, merge or force to get past it. A rejection stops the push and is reported.
- A push to the default branch happens only when the user names it, in trunk mode too.
- `head` must equal the work item's branch and `base` the target that `pull-request.md` names.
- When `verdict_matches` is `true`, the verdict's `worktree_hash` must still equal the committed tree. When it is `false`, the description carries no met or missing marks.
- An open pull request for the same head and base is updated, never opened twice.
- A comment that does not attach to its line is reported, not posted again.
