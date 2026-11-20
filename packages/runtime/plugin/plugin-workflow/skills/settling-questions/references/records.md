# Decision records

## The four states

The operator's answer decides the state of the record. Offer the open state in every brief and never push for a settled one.

| State | Use when | `state` |
| --- | --- | --- |
| Settled | The operator picked an option, in their own words, with any nuance they added. | `settled` |
| Partly settled | One part is decided and another part goes to its owner. | `partly-settled` |
| Paused | The operator needs a named fact first. Name the fact and who gathers it. | `paused` |
| Left open | The operator wants the owner, often the engineers, to validate. Keep the operator's phasing. | `open` |

## The claim

Record the answer as one testable sentence: "When <trigger>, the system shall <response>". It must be unambiguous enough that a check can fail it. Keep the nuance the operator gave ("yes, but not in the app layer") in the same sentence or in the reason, never drop it. When the claim is about a process and not a system, name the actor in place of "the system".

For a question that is not settled, `# Decision` states the question to settle in one sentence, and `# Reason` says why it matters.

## The file

One `Workflow Decision` file for each recorded answer, at `decisions/<slug>-<suffix>.md`, from the decision template of **recording-handoff**. Name the slug after the question, such as `target-identifier`.

- `generated.by` is `settling-questions/<version>`.
- `verified` names the operator as `human:<id>` when their answer produced the record. Take the id from the invoker and ask once when it is unknown; never invent one. Leave the key out only for a record nobody answered.
- `owner` names the person or team that settles an open, paused or partly settled question.
- `depends_on` lists earlier decisions this one rests on; `blocks` lists existing decisions that wait on this one.
- `sources` cites the research file as `/research/<slug>-<suffix>.md` and any code or page the reason rests on, with footnotes keyed to the ids. Quote no sensitive finding.
- `# Options considered` lists every lettered option with gains and costs, including the operator's own and the ones excluded, with the reason. `# Holds either way` lists the rules that hold under every option, once.

## Sections by state

| State | `# Decision` | `# Reason` | Also |
| --- | --- | --- | --- |
| Settled | The claim | Why this option wins, and the correction it implies for a page or plan | `# Holds either way` when any |
| Partly settled | The claim for the settled part | Why it is settled, then "Left to the owner:" and the remaining question | `owner`; the candidates for the remaining part in `# Options considered` |
| Paused | The question to settle | Why it matters, then "Paused until:" the fact and who gathers it | `owner` for the person who gathers the fact; no option is marked chosen |
| Left open | The question to settle | Why the operator wants validation, and the phasing they gave | `owner`; the candidates, the excluded ones with reasons, `# Holds either way` |

A paused decision returns to the queue when the fact is known. Its answer is a new file; see below.

## Superseding a decision

An answer that changes an existing decision, including one that settles an open or paused file, is a new file that follows the write rules of **recording-handoff**: write the new file with `supersedes`, then set only `status: deprecated` and `superseded_by` on each old file, and list the new and the old concept IDs in `touched` of the log entry. The validator checks the pair.

## Examples

A settled record, invented:

```markdown
---
type: Workflow Decision
schema: workflow/decision/1
title: Address messages by device id
description: Messages are addressed by device id, never by account id.
generated: {by: settling-questions/0.1.0, at: 2026-03-02T10:15:00+01:00}
verified: {by: human:jdoe, at: 2026-03-02T10:14:00+01:00}
state: settled
owner: Notifications team
---

# Decision

When the publisher sends a visible message, it shall address the message by device id and store no account id.

# Reason

The delivery service accepts device targets only.

# Options considered

| Option        | Gains                    | Costs                          |
| ------------- | ------------------------ | ------------------------------ |
| A. Device id  | Matches the delivery API | Needs the id on every device   |
| B. Account id | Reaches every device     | The service has no such target |
```

A left-open record names its phasing and its rule:

```text
Left open for the engineers: how the publisher learns the device id.
Candidates: A (request header), B (body field), C (event), E (added by the operator: remap at send time).
Phasing: A first, C later is allowed.
Excluded: D, because the platform offers no such operation.
Holds whichever wins: storage and routing do not depend on where the id came from.
```
