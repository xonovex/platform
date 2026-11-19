# Delivery modes

The project's `delivery.mode` setting decides which files the skill writes. It has no default: a missing or unknown value is a blocker, because the mode decides what a delivery consists of. Report the missing setting and stop; never guess it from a remote or from a hosting tool the invoker mentioned.

| Mode | What delivery means | Files written |
| --- | --- | --- |
| `trunk` | The commit goes straight to the default branch. | `commit-message.md` only |
| `branch` | The branch is pushed and stays for a person to integrate. | `commit-message.md` only |
| `pull-request` | The branch is pushed and a pull request is opened for review. | `commit-message.md`, `pull-request.md`, optional `comments/*.md` |

A pull request description and inline comments are anchored to a pull request, so the two modes without one write neither, even when the implementation notes hold decision notes. In those modes the decisions go into the commit message body.

The mode changes only what is written. The commit goes to **using-git**, and a push or a pull request goes to **using-github** or **using-gitlab**, only when the user asks. A push to the default branch must be named by the user even in trunk mode.

A delivery that changes mode between runs is rewritten for the new mode. A file the new mode does not call for is left alone and reported, because deleting a delivery file is the invoker's decision.
