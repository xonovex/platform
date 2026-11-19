# Report sections

A migration research file is a `Workflow Research` file with `kind: migration`. It keeps the headings the research file requires and places the migration sections under them as `##` headings. Write the bottom line first and keep a footnote on every fact.

## Contents

- Where each section goes
- Current behavior
- The target contract
- What each consumer does
- Consumer versions and release cadence
- Gap table
- Changes needed inside the producer
- Risks
- Unresolved items
- Corrections to existing pages
- Frontmatter for a migration

## Where each section goes

| Migration section | Heading in the research file |
| --- | --- |
| Current behavior | `## Current behavior` under `# Observed` |
| The target contract | `## Target contract` under `# Observed` |
| What each consumer does | `## Consumers` under `# Observed` |
| Consumer versions and release cadence | `## Versions and cadence` under `# Observed` |
| Gap table | `## Gap table` under `# Inference` |
| Changes needed inside the producer | `## Producer changes` under `# Inference` |
| Risks | `## Risks` under `# Inference` |
| Unresolved items | `# Unknowns` |
| Open decisions | `# Open decisions`, one row for each question |
| Corrections to existing pages | `# Corrections` |
| What was not verified at runtime | `# Not checked` |

## Current behavior

What the producer sends, through which path, what it stores and what it does with the answer. Cite the seam and every bypass. Say which repository commit each fact was read at.

## The target contract

What the target accepts, how it authenticates, what it does on bad input, and which documented behavior the code lacks. When the target does not exist yet, say that the section describes a design record and not running code.

## What each consumer does

One paragraph or list for each consumer app: the identifiers it holds, how it handles old and new message shapes, what a user action on a message does, and when it registers again. Do not merge two apps into one description.

## Consumer versions and release cadence

For each consumer, the released versions in the window that matters, the first version that handles the new shape, how often it releases and how long old versions stay in use. The supported versions are a computation over these facts, so label the computed window as `Inference` and keep the inputs as facts.

## Gap table

One row for each gap between the producer's behavior and the target's contract, built from the template in [../assets/templates/gap-table.md](../assets/templates/gap-table.md). Each row carries the gap, the evidence, the options, a recommendation and the question id when the row needs a decision. The recommendation is an inference and keeps its label. A row that needs a decision is also a `questions` entry in the frontmatter and a row of `# Open decisions`.

## Changes needed inside the producer

A list of the changes the producer needs for the target, each tied to a gap row and to the code it touches. It names changes and never schedules or splits them into work: that belongs to the plan.

## Risks

What can go wrong during the switch and after it, each with the fact it rests on. Mark a risk that rests only on inference.

## Unresolved items

What the explorers could not establish, and every repository that was not cloned. These go under `# Unknowns` with the reason each stays unknown.

## Corrections to existing pages

Each existing page that the research contradicts, named by page, with what it says, what the code shows and the commit it was checked at. A page whose stamped commit trails its repository goes here too, as a correction and not as a fact.

## Frontmatter for a migration

- `kind: migration`.
- `roles`: the roles used, named for the systems they cover, for example `order-publisher`, `delivery-service`, `shop-app-mobile` and `shop-app-web`.
- `questions`: one entry for each gap row that needs a decision, with a kebab-case id that the table repeats.
- `corrects`: include when the file replaces an earlier research file.
- `sources`: one entry for each cited file at a pinned commit, with an id that the footnotes repeat.
