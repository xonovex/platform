# The plan sections

## Contents

- Order of the sections
- Objective
- Scope
- Current state
- Differences between variants
- Success criteria
- The generated block

## Order of the sections

1. `# Objective`
2. `# Scope`
3. `# Current state` (optional)
4. `# Differences between variants` (optional)
5. `# Success criteria`
6. The generated block between `<!-- generated:begin -->` and `<!-- generated:end -->`

Open each section with its main message. Leave an optional section out entirely when it does not apply, and leave no empty heading.

## Objective

A picture of the settled state only: what is true when the change is done, in two to four sentences, as a future-state paragraph. Name the actors, the behavior and the result. Leave out the route, the options that lost, the reasons and anything not yet decided. Where a decision shapes the picture, state its settled outcome in your own words and do not quote the decision file.

Example: "Finance receives each invoice as an event within one minute of its issue, and the shop portal shows the same invoices from the same feed. The nightly export file is no longer produced."

## Scope

Two lists, both always present.

- **In scope:** capabilities and areas the change covers, one line each.
- **Out of scope:** what the change leaves alone, with a short reason when a reader would expect it to be included, and the owner when someone else takes it up.

Scope names areas and capabilities, never tasks. A task belongs in a work file.

## Current state

Include when the reader needs today's behavior to understand the objective. Write observed facts grouped by system, each with a footnote keyed to a `sources` entry for the research file, and keep the labels the research gave: an inference stays labelled `Inference`. Summarize in a few lines; the research file holds the detail and the anchors. Name facts nobody established as unestablished. Leave out recommendations and proposals.

## Differences between variants

Include only when variants of the same behavior differ, such as platforms, clients, markets or tenants. A table with one row for each behavior, one column for each variant, and a last column `Decided by` that links the decision file that settles or will settle the row.

```markdown
| Behavior         | Mobile client              | Web client                | Decided by                                          |
| ---------------- | -------------------------- | ------------------------- | --------------------------------------------------- |
| Invoice list     | Pages of 20 from the feed  | Pages of 50 from the feed | [Page size](/decisions/page-size-<suffix>.md)       |
| Offline behavior | Shows the last synced list | Shows an error state      | [Offline mode](/decisions/offline-mode-<suffix>.md) |
```

Describe what each variant does or will do, not why. A row that rests on an open decision links it, and the cell states only what is known.

## Success criteria

One bullet for each result a reader can observe by a stated means: a metric and its threshold, a request path and its response, a behavior in a named flow, a check that passes. Say the means in the bullet. Never name an implementation detail such as a class, a library or a table.

- Good: "The invoice feed delivers 99% of invoices within 60 seconds of issue, measured on the delivery dashboard over one week."
- Good: "A request for `GET /invoices?since=...` on the portal returns the same ids as the feed for the same period."
- Bad: "The `InvoiceFeedPublisher` class replaces the export job." It names a mechanism, and no observation shows it.

## The generated block

The block links every decision that is not deprecated and every work file, one line each, with the file's `title` as the link text and a bundle-relative path. It repeats no decision text and no work definition. Rewrite the block as a whole and edit nothing inside it by hand.

1. Write the work files and make sure the decision files exist.
2. Run `index --write` for the bundle, so `decisions/index.md` and `work/index.md` hold the current order.
3. Copy the order of the two indexes into the block: decisions grouped as the decisions index groups them, work in the dependency order of the work index. Omit the display numbers.

```markdown
<!-- generated:begin -->

## Decisions

- [Delivery method](/decisions/delivery-method-<suffix>.md)
- [Retention of the nightly file](/decisions/nightly-retention-<suffix>.md)

## Work

- [Publish invoices as events](/work/publish-invoice-events-<suffix>.md)
- [Show the feed in the portal](/work/portal-invoice-feed-<suffix>.md)

<!-- generated:end -->
```

The block does not say which decisions are open or which work waits. The decisions and work indexes group them, and each work file states its own readiness.
