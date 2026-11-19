# Sources

## How we built our multi-agent research system

- **URL:** https://www.anthropic.com/engineering/multi-agent-research-system
- **Last reviewed:** 2026-10-27
- **Used for:**
  - `references/delegation.md` → Delegating a task
  - `references/tasks.md` → Ordering the tasks
  - `assets/templates/task-brief.md` → the brief
- **Aspects extracted:**
  - A worker brief states the task, the files it reads and may change, the rules that apply, the checks and the required return, because vague briefs make parallel workers duplicate or miss work → delegation, task-brief
  - A lead agent starts workers in parallel and scales their number to the work, within a bound → tasks, delegation

## Building effective agents

- **URL:** https://www.anthropic.com/engineering/building-effective-agents
- **Last reviewed:** 2026-10-27
- **Used for:**
  - `SKILL.md` → Essentials
  - `references/tasks.md` → Ordering the tasks
  - `references/delegation.md` → Delegating a task
  - `references/verification.md` → What to run
- **Aspects extracted:**
  - Sectioning runs independent subtasks at the same time → tasks
  - An orchestrator splits a task, delegates the subtasks to workers and synthesizes their results instead of doing them itself → delegation
  - Progress is judged on ground truth from the environment, such as test runs, not on an agent's own report → verification, Essentials

## git-worktree documentation

- **URL:** https://git-scm.com/docs/git-worktree
- **Last reviewed:** 2026-10-27
- **Used for:**
  - `references/worktree-and-branch.md` → Finding or creating the worktree
- **Aspects extracted:**
  - A repository can have several branches checked out at once, each in its own linked worktree, so one work item's changes stay apart from the main clone and from other items → worktree-and-branch

## Xonovex work-item implementation conventions

- **Provenance:** Repository-original procedure for building one ready work item, built on the contract of **recording-handoff** and the worktree procedure of **using-git**: the preconditions, the worktree and its branch, the split into tasks, delegation and its fallback, the self-check and the implementation notes
- **Last reviewed:** 2026-11-19
- **References:** references/preconditions.md, references/worktree-and-branch.md, references/tasks.md, references/delegation.md, references/verification.md, references/notes-file.md, references/closing-the-run.md
- **Aspects extracted:** The run executes the entry step of **recording-handoff** first and loads knowledge skills from the skills map of the `implementing-work-item` phase; a dependency counts as done only with a passing verdict or the invoker's word; the notes and the log entry go to the main checkout's bundle and never to the worktree's copy of `plans/`; a decision note names one `path:line`; everything the run concludes is a self-check; nothing is committed, staged or pushed.
