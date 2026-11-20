# Explorer roles and briefs

## Contents

- Deriving the roles
- Roles for each kind of change
- Writing a brief
- Fan-out limits
- Reading what returns

## Deriving the roles

One role stands for one system, repository or concern that must be understood separately. Ask of each candidate: does answering the question about it need a different set of files, a different vocabulary or a different owner than the others? If yes, it is a role. If two candidates would read the same files, merge them.

Name each role by what it is in this request, such as `order-publisher`, `broker-service` or `shop-app-android`, never by a fixed category. The `roles` key of the research file lists the names used.

## Roles for each kind of change

These are starting points; the request decides.

| Kind | Likely roles |
| --- | --- |
| migration | The system that does the work today, the system that will do it, and each client that depends on either |
| feature | The module the feature extends, the neighbors it calls or is called by, the existing pattern to follow |
| quality | One role for each module or layer with the problem, and one for the measurements or tooling |
| investigation | The place the symptom appears, the path that leads to it, and the history of recent changes |
| project | The users and their needs, the constraints, the neighboring systems, comparable solutions |

A change that moves one system's traffic or behavior onto another across several repositories and client apps has its own roles, briefs and report sections in [migration.md](migration.md); read it for that kind.

## Writing a brief

Use [../assets/templates/explorer-brief.md](../assets/templates/explorer-brief.md), one brief for each role. A brief that is missing a part is the usual cause of duplicated or off-target work:

- **The question and the reason:** what to find out, and which decision it informs.
- **What to read:** the repositories (with the commit to read at), the documents, the web pages. Say what is out of bounds.
- **The required output:** an `Observed` heading for facts with an anchor, an `Inference` heading for conclusions drawn from facts, and an `Unknowns` section. One anchor per line, as the repository root plus the path and the line.
- **The boundary:** read-only, no file written, no command that changes state, no fix proposed unless asked.

## Fan-out limits

Run the explorers in parallel, but keep the number bounded: enough that independent roles do not wait for each other, few enough that the lead can read every report. Give each explorer the lowest model and effort that does its part reliably, and the most capable only for a design question or a hard fault, never for rework.

An explorer that returns without anchors, or with its facts and inferences mixed, is briefed again with the missing part named. Do not repair its output by guessing.

## Reading what returns

As each report returns, relay the highlights in a few lines: what is new, what contradicts a document, what contradicts another explorer. Say nothing when nothing is new. Collect the pointers to material outside the default branch for the follow-up read.
