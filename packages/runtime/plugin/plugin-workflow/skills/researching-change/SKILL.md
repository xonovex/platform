---
name: researching-change
description: "Use when a change needs checked facts before anyone decides or plans: a migration, a new feature in existing software, a code-quality improvement, an investigation of a problem or a new project. Reads the existing documents, fans out read-only explorers by role, keeps observed fact apart from inference, re-checks the claims that would change a decision, and saves one research file with the open questions and the corrections existing pages need; a migration also gets producer, target and consumer roles. Triggers on 'find out first', 'research before we decide', what a system does today, what a target accepts, or which pages are wrong, even when the user doesn't say 'research'. Not for a single-file lookup, writing the plan, settling the decisions, auditing code quality or debugging."
---

# Research Before Deciding

Purpose: establish what is true today, what the change touches and what is still unknown, so the decisions that follow rest on checked facts. The output is one research file in the change's plan directory. The research decides nothing and plans nothing.

**Load first:** **recording-handoff**, and run its entry step before anything below.

## Essentials

- The explorer roles come from the request, never from a fixed list. One role stands for one system, repository or concern that must be understood separately.
- Explorers are read-only, run in parallel and bounded in number, each with a stated question and reason. They return `Observed`, `Inference` and `Unknowns`, with one anchor per line.
- A fact carries an anchor, an inference keeps its label wherever it is repeated, and nothing unverified becomes a claim.
- A research file is frozen. A later correction is a new file that names the old one in `corrects`.
- Never commit or stage. Never start the next step.

## Procedure

1. **Run the entry step** of **recording-handoff**. The research file is written into the plan directory's repository, so confirm that branch. The repositories the explorers read need no confirmation. The step's own precondition is the change directory; when the invoker names a new change, **recording-handoff** starts it.
2. **Check the intent once.** The request must say what is changing, what it affects and where to look. If it does not, ask one question before reading anything, together with the branch question when both are due. A new project has no existing system: say so in the file and research the context, constraints and comparable solutions instead of inventing a current system. See [references/intent-check.md](references/intent-check.md).
3. **Load the skills that apply** from the sources that the settings reference of **recording-handoff** lists for the `researching-change` phase. This skill names none itself.
4. **Read the existing documents** on the subject yourself and note where they disagree; the first disagreement is the first question for the code. Check every repository you will read: it is cloned, and what commit it is at. Keep the commands you ran. A repository that is not cloned is a gap to report, never a reason to guess. See [references/evidence.md](references/evidence.md).
5. **Derive the roles and fan out.** Write one brief per role from [assets/templates/explorer-brief.md](assets/templates/explorer-brief.md) and run the explorers in parallel, within the fan-out limits. When the request moves one system's traffic or behavior onto another across several repositories and apps, follow [references/migration.md](references/migration.md) for its roles, briefs and report sections. Relay highlights as each report returns, in a few lines, and say nothing when nothing is new. See [references/explorer-roles.md](references/explorer-roles.md).
6. **Follow the pointers and re-verify.** Read any material the explorers point to outside the default branch yourself. Then check in the code the few claims that would change a decision, and say which side you checked.
7. **Write the report inline**, bottom line first, then save it as `research/<slug>-<suffix>.md` from [assets/templates/research.md](assets/templates/research.md), with a `Workflow Log Entry` file and regenerated indexes. Announce the write in one line. See [references/research-file.md](references/research-file.md).
8. **Stop** and offer the next steps: settling the open questions with **settling-questions**, and a correction to any existing page. Never start another step.

## When not to use

- A single-file lookup: read the file.
- Writing the plan: use **writing-plan**. Settling the questions the research raises: use **settling-questions**.
- Auditing code quality against a standard: use **auditing-code-quality**. Debugging a failure in native code: use **debugging-native-code**. This skill gathers facts before a change and fixes nothing.

## Gotchas

- Explorers that write files or fix what they find. The brief says read-only, and the lead writes the one file.
- A vague brief. Without an objective, an output format, the sources to read and a boundary, explorers duplicate each other's work.
- A design document that claims behavior the code lacks. Report the code, and list the document under corrections.
- Labels that weaken in synthesis: "Inference" becomes "probably" and then disappears. Keep the label wherever the claim is repeated.
- Wrapped paths in a sub-report cannot be spot-checked. Ask for one anchor per line.
- Collapsed commands cannot be reconstructed. List the commands that checked the repositories.
- Inventing a current system for a new project, or reading before the intent is clear.
- Editing a saved research file. Write a new file with `corrects` instead.
- Reading the explorers' summaries as proof. Re-check the claims that would change a decision yourself.

## Done when

- The file validates with the script of **recording-handoff** and has no schema-not-found warning.
- Every fact has a footnote anchor, every inference is labelled, and `# Not checked` says what was not verified at runtime.
- Each `questions` id appears as a row in `# Open decisions`.
- Every contradiction between documents, and between documents and code, is listed, with each page named under `# Corrections`.

## Progressive Disclosure

- Read [references/intent-check.md](references/intent-check.md) - Load when the request is vague, names no system, or describes a new project.
- Read [references/explorer-roles.md](references/explorer-roles.md) - Load when deriving roles from the request, writing a brief, or limiting the fan-out.
- Read [references/evidence.md](references/evidence.md) - Load when checking repositories, labelling facts, citing sources, or re-verifying a claim.
- Read [references/research-file.md](references/research-file.md) - Load when writing the file, its questions, corrections and log entry, or correcting an earlier file.
- Read [references/migration.md](references/migration.md) - Load when the research moves the traffic or behavior of one system onto another.
- Read [references/migration-role-briefs.md](references/migration-role-briefs.md) - Load when filling in a migration brief, to see what the producer, the target and each consumer must be asked and why.
- Read [references/migration-report-sections.md](references/migration-report-sections.md) - Load when writing a migration research file, to place each migration section under the headings the research file requires.
- Read [references/migration-failure-modes.md](references/migration-failure-modes.md) - Load when saving a migration research file, to check it against the known failure modes and their fixes.
- Read [assets/templates/research.md](assets/templates/research.md) - Load when writing the research file.
- Read [assets/templates/explorer-brief.md](assets/templates/explorer-brief.md) - Load when briefing an explorer.
- Read [assets/templates/producer-brief.md](assets/templates/producer-brief.md) - Load when briefing the explorer for the current producer of a migration.
- Read [assets/templates/target-brief.md](assets/templates/target-brief.md) - Load when briefing the explorer for the target of a migration.
- Read [assets/templates/consumer-brief.md](assets/templates/consumer-brief.md) - Load when briefing the explorer for a consumer app of a migration, once for each app.
- Read [assets/templates/gap-table.md](assets/templates/gap-table.md) - Load when writing the gap table and the producer changes list into a migration research file.
- Read [assets/schemas/research.schema.json](assets/schemas/research.schema.json) - Load when the validator reports a schema finding on a research file.
