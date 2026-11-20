# The delivery files

## Contents

- Where they live
- The commit message
- Limits
- Rerunning

## Where they live

Under `deliveries/<work-id>/` in the change's plan directory, where `<work-id>` is the work file's name without `.md`. The files describe the content of a delivery and apply nothing: the commit goes to **using-git**, and a push, a pull request and inline comments go to **using-github** or **using-gitlab**, each only when the user asks. [handoff.md](handoff.md) names what each of them takes, and [pull-request-and-comments.md](pull-request-and-comments.md) has the rules for the description and the comments.

| File | Type | Schema | Written when |
| --- | --- | --- | --- |
| `commit-message.md` | `Workflow Commit Message` | `workflow/commit-message/1` | Every delivery mode |
| `pull-request.md` | `Workflow Pull Request Description` | `workflow/pull-request/1` | The delivery mode is pull-request |
| `comments/<suffix>.md` | `Workflow Inline Comment` | `workflow/inline-comment/1` | Optional, in pull-request mode |

Each schema extends the shared header of **recording-handoff** and is in `assets/schemas/` of this skill, beside the templates in `assets/templates/`. `generated.by` is `preparing-delivery/<version>`, and the files carry no `verified` entry because the skill wrote them. `<version>` is the `version` field of the `plugin.json` of the plugin that ships this skill (read it, do not guess). Every `generated.at` comes from the system clock: run `date -u +%Y-%m-%dT%H:%M:%SZ` and use its output; never invent, round or copy a time from another file. A string value that contains a colon followed by a space, a quote or a leading special character is written in single quotes in the frontmatter; a pull request `title` such as `feat: add retry limit` needs them. Every directory except `diagrams/` has a generated index, so `deliveries/`, `deliveries/<work-id>/` and `comments/` each get one: regenerate them with the **recording-handoff** script after writing.

## The commit message

Keys beyond the header: `work` (concept ID of the work item) and `branch` (the worktree's branch, from the implementation notes).

The body of the file is the message. Subject line first, in the conventional commit form of **using-git** or the convention the project sets, at most 72 characters. A blank line, then plain sentences that state the durable reasoning: why the change exists, the choices that shaped it, each with its decision ID and a one-line claim, and the plan link. Test output, task lists, progress and the acceptance criteria stay out; they change or belong to the pull request.

A code comment is suggested only where the code cannot show the reason, as a last line `Suggested code comment (<path>:<line>): ...`. It is a proposal for the implementer. The skill edits no code.

The plan link is a URL only when the repository's `resource` is a browsable URL for files, that is an `https://` URL of a hosted repository. Otherwise write the repository name and the path `plans/<change>/plan.md`, for example `project: plans/greeting-locale/plan.md`. Never write a file URL or a bare repository URL.

## Limits

- Commit subject: at most 72 characters.
- Pull request description body: under about 3,500 characters. Hosts reject long descriptions, and the margin leaves room for what the host adds. Count with `wc -m`, and shorten by cutting detail from the sections, never by dropping the criteria.
- Inline comment: three sentences.

## Rerunning

The delivery files describe the current tree, so a rerun replaces them in place, and git keeps the history. A comment is matched by its `marker`: update the file that has it, set `line` and `path` again, and never write a second file with one marker. A comment whose marker no longer applies gets `status: deprecated` and stays, so the report can tell the user to retire it on the host. Preserve unknown frontmatter keys and update `generated.at` on every edit.
