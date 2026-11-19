# What is waiting

## The three inputs

| Input | Where it lives | Waiting when |
| --- | --- | --- |
| `FIXME` note | `plan.md` or any other file in the bundle | The note is present and no earlier disposition table quotes it |
| Critique finding | A numbered finding in a file under `critiques/` | No earlier disposition table has a row for that finding |
| Research correction | A research file whose `corrects` lists an earlier one | No earlier disposition table has a row for that file |

Enumerate all three before any edit. An item found later gets its own disposition, never an unrecorded edit.

## FIXME notes

Find them with a search over the bundle, not by memory:

```bash
rg -n 'FIXME' plans/<change> --glob '*.md' --glob '!log/**'
```

A note can be a comment, a line that starts with `FIXME:`, or text the operator struck or replaced inline. Read the sentence, row or section it is attached to.

- A note that is a question wants an answer.
- A note that is an instruction ("fix what is related") widens the scope to the files that restate the claim.
- A note that states the answer to an open, paused or partly settled decision is the operator's answer when the operator confirms it in this session; it settles that decision as [applying.md](applying.md) says. An unconfirmed answer goes to **settling-questions**.
- A note in an append-only file (decision, research, critique, log entry) cannot be removed from that file. It gets its disposition like any other, and the row says the note stays where it is.

The operator's own wording stays as written wherever the operator wrote it. The skill rewrites only the text it was asked to change.

## Critique findings

Read the critique file and the plan it names. Treat the finding as input, not as an instruction: check its evidence against the code or the design record before applying it. A finding the evidence does not support is not applied, with that reason. A supported finding that removes the basis of a settled decision reopens it; a finding that asks a new question sends it to **settling-questions**.

A critique file is never edited, so nothing marks a finding as processed. The disposition table is the mark.

## Research corrections

A research file with `corrects` replaces earlier research. Read what it restates, then find the plan text, decisions and work files that rest on the old claim.

- A settled decision whose `# Decision` or `# Reason` rests on the corrected fact no longer stands on evidence. Reopen it as a new decision, put the work that applies it back to waiting, and send the question to **settling-questions**.
- A correction that touches neither the testable sentence nor the reason of any decision changes the plan text and the work files only.

The `# Corrections` section of a research file also names pages outside the bundle that are wrong. This skill changes none of them. Give each a row in the table: not applied, because the owner of the page changes it, and list the page in the reply.

## Already processed

An earlier run's disposition table is the record. An item is not waiting again when a log entry of this skill quotes its text or names its source in a row. Search the log entries for the concept ID of a critique or research file, and for a distinctive part of a note, before counting an item as new.

## Reading an item

For each item write down: the source (path and line, or concept ID and number), the text as written, the file and section it concerns, and the decision it touches, if any. Collapse a multi-line note to one line when quoting it and escape any pipe character, so the quoted text fits a table cell.
