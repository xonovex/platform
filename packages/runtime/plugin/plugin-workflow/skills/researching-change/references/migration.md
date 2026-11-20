# Researching a migration

A migration moves one system's traffic or behavior onto another across several repositories and client apps: replacing an API, moving a publisher to a new delivery service, retiring a send path, switching a provider. This reference fills in what the procedure of the skill leaves open for that kind: which roles to derive, what each brief asks, and which sections the report carries. It adds no document type and no schema: the output is the `Workflow Research` file with `kind: migration`.

## Contents

- Essentials
- Prerequisites
- When a migration is not the kind
- Steps
- Gotchas

## Essentials

- Follow the procedure of the skill: the entry step, the intent check, the exploration, the write rules and the stop. This reference does not repeat it.
- A migration has three kinds of role: the current producer, the target and each consumer app. Each gets one read-only explorer and one brief from the migration templates in `assets/templates/`. [migration-role-briefs.md](migration-role-briefs.md) says what each role reads and why.
- The research file sets `kind: migration` and lists the roles used in `roles`. Every gap that needs a decision becomes an entry in `questions` and a row in the `# Open decisions` table.
- Facts and inferences stay apart in the file. The migration sections go under `# Observed` and `# Inference` as `##` headings, as [migration-report-sections.md](migration-report-sections.md) lays out.

## Prerequisites

- The subject names a current producer, a target and at least one consumer. If the target does not exist yet, research its contract or its design record instead, and say in the file that no running target was read.
- Read access to each repository. A repository that is not cloned is a gap to report under `# Unknowns`, never a reason to guess.

## When a migration is not the kind

- A change inside one repository, or a new feature or a new project with no existing system to move: use the general roles of [explorer-roles.md](explorer-roles.md).
- Settling the decisions the report raises: use **settling-questions**. Writing the plan: use **writing-plan**.

## Steps

1. Check the prerequisites against the request as part of the intent check. If the subject lacks a producer, a target or a consumer, ask the one question that names the missing part before reading anything.
2. Derive the roles: one for the current producer, one for the target, and one for each client app that receives from or calls the system. Name each role for the system it covers, such as `order-publisher`, `delivery-service`, `shop-app-mobile` and `shop-app-web`. A web client and a mobile client are two roles because they release differently.
3. Fill in one brief for each role from [../assets/templates/producer-brief.md](../assets/templates/producer-brief.md), [../assets/templates/target-brief.md](../assets/templates/target-brief.md) and [../assets/templates/consumer-brief.md](../assets/templates/consumer-brief.md). Name the repositories, the question and the reason in every brief.
4. Fan the explorers out, relay highlights, follow pointers outside the default branch and re-verify the claims that would change a decision, as the procedure of the skill says.
5. Write the sections of [migration-report-sections.md](migration-report-sections.md) into the research file. Paste the gap table and the producer changes list from [../assets/templates/gap-table.md](../assets/templates/gap-table.md).
6. Compare the report with [migration-failure-modes.md](migration-failure-modes.md) before saving: each mode names a symptom you can search for.

## Gotchas

- A page stamped with a commit the repository has moved past is a correction to that page, never a fact about the system.
- A design document that claims behavior the code lacks (persistence, deduplication) is reported as the code shows it, and the document goes under `# Corrections`.
- One consumer role for two client apps hides the difference in their versions and release cadence. Use one role for each app.
- Computing the supported versions without the release cadence. The versions section needs both before any window is stated.
