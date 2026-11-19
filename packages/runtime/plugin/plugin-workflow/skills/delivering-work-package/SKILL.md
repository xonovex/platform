---
name: delivering-work-package
description: "Use when the ready work items of one work package are to be carried from implementation to a local commit, one item at a time in dependency order: each item is implemented, validated independently and given its delivery text, each step in its own agent, and the supervisor reads every verdict and commits the item locally only when the verdict covers the tree. Pushes nothing and opens no pull request. Triggers on running the whole work package with agents, delivering every ready work item, supervising implementation agents, one local commit per validated item, or resuming a package run, even when the user doesn't say 'work package'. Not for one step of one work item, validating the package, pushing or opening a pull request."
---

# Deliver a Work Package by Delegation

Purpose: carry each ready work item of one work package through its implementation, its independent validation and its delivery text, and commit it locally so the next item builds on it. You are the supervisor: you select, brief, verify, commit and record. You never implement, validate or write delivery text yourself. The output is one local commit for each item whose verdict covers its tree, one log entry for each item the run reached, and a report. The run pushes nothing and opens no pull request.

**Load first:** the **recording-handoff** skill, and run its entry step before anything below.

## Essentials

- One item at a time, one step at a time, in dependency order: each step regenerates the shared indexes, and each item builds on the commit before it.
- Each step runs in its own new agent: **implementing-work-item**, **validating-work-item**, then **preparing-delivery**. Set the model and the effort in the call that starts the agent: the cheapest that does the step with no loss of quality, the most capable for a hard fault.
- Start only the steps an item needs. Its files tell its state; see [references/item-states.md](references/item-states.md).
- Verify each agent's work from the files and from Git, never from its reply. The reply is a lead, not proof.
- Commit an item only when the tree hash of its worktree equals the `worktree_hash` of a verdict whose results are all `pass`, and only with the message that **preparing-delivery** wrote.
- A failed verdict stops that item, not the package. An item that waits on a decision or on a stopped or skipped item is skipped and reported.
- Never push, open a pull request, fix a failure, edit a verdict or a delivery file, or start **validating-work-package**.

## Procedure

1. **Run the entry step** of **recording-handoff**. The run writes log entries into the plan directory's repository and commits into each work repository, so confirm the branch of the main checkout of each. The confirmed branch is the base of the first item in that repository. Preconditions: `plan.md`, at least one work item in the package, the `delivery` setting that **preparing-delivery** needs, a `repositories` entry for each repository the items name, and a harness that can start an agent with a model and an effort. Report a missing one and stop.
2. **Load the skills that apply:** **using-git** for the commit, and the **delivering-work-package** entry of the project's `skills` map. Each step agent loads the skills of its own phase.
3. **Resolve the package and the order.** Read the package, sort its items by `depends_on`, read the state of each item and mark each skip with its reason. Report the order and the skips before the first agent starts. See [references/item-states.md](references/item-states.md).
4. **Run each item.** For each step the item needs, write a brief from [assets/templates/step-brief.md](assets/templates/step-brief.md), start one agent with a model and an effort, wait for it, then verify its work. A finding goes back to the same agent. See [references/supervising-agents.md](references/supervising-agents.md).
5. **Read the verdict.** Every result `pass`, `source: independent`, one result for each criterion, and a `worktree_hash` equal to the tree hash you compute: the item goes on to **preparing-delivery**. Anything else stops the item and skips the items that depend on it.
6. **Commit and record.** Commit the item through **using-git** with the body of `deliveries/<work-id>/commit-message.md`, check the tree of the commit, write the log entry of the item, regenerate the indexes and the log, validate the bundle, and report the item in a few lines. See [references/commit-and-record.md](references/commit-and-record.md).
7. **Stop after the last item**, or at the stop point the invoker named. Report the package and offer **validating-work-package**, and a push or a pull request through **using-git**, **using-github** or **using-gitlab** when the invoker asks. Start none of them.

## When not to use

- One step of one work item: use **implementing-work-item**, **validating-work-item** or **preparing-delivery** directly.
- Checking the package against the plan's success criteria: use **validating-work-package**.
- Pushing a branch, opening a pull request or commenting on a host: use **using-git**, **using-github** or **using-gitlab**.
- Cutting or changing work items: use **defining-work-items** or **revising-plan**.

## Gotchas

- Committing because the verdict says `pass` while the tree has moved since. Compare the hashes as strings; a mismatch needs a new validation.
- Accepting a result that is close but short. A measured 212 ms against a criterion of 200 ms is a `fail`, whatever the agent's reply says.
- Giving the validation agent the implementer's reply, notes or reasoning. It reads the criteria and the diff only.
- Doing a step yourself because it is small, or fixing a failure while you are there.
- Choosing the most capable model to redo a thin result. Brief the same agent again with the missing part named.
- Committing with `--no-verify`, or amending a commit to get past a hook.
- Stopping the package on one failed item, or running an item whose dependency stopped.
- Running **implementing-work-item** again for an item whose verdict already covers its tree.
- Forwarding a branch confirmation that the invoker did not give.
- Pushing the last branch because every item is committed.
- Handling feedback from the invoker inside a running agent. A change to a work item goes through **revising-plan** or **defining-work-items**, and the item runs again.

## Done when

- Every item of the package has a result: committed, stopped with its reason, or skipped with its reason.
- The tree of each commit equals the `worktree_hash` of the verdict of its item, and its worktree is clean.
- Each item the run reached has one log entry, and the bundle validates with the **recording-handoff** script.
- Nothing was pushed, no pull request was opened, and **validating-work-package** was offered, not started.

## Progressive Disclosure

- Read [references/item-states.md](references/item-states.md) - Load when resolving the package, ordering its items, reading the state of an item, or deciding to skip one.
- Read [references/supervising-agents.md](references/supervising-agents.md) - Load when choosing a model and an effort, briefing a step agent, verifying its work, or handling an agent that failed.
- Read [references/commit-and-record.md](references/commit-and-record.md) - Load when comparing the tree hash, committing an item, writing its log entry, or reporting an item or the package.
- Read [assets/templates/step-brief.md](assets/templates/step-brief.md) - Load when writing the brief for a step agent.
