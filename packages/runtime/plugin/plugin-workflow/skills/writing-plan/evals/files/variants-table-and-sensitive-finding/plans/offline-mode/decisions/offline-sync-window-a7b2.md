---
type: Workflow Decision
schema: workflow/decision/1
title: Offline sync window
description: How long the last synced list may be shown.
generated: {by: settling-questions/0.1.0, at: 2026-10-20T09:00:00+02:00}
verified: {by: human:jdoe, at: 2026-10-21T10:00:00+02:00}
state: open
owner: Product
---

# Decision

When the connection stays lost, the system shall stop showing the list after a window still to be decided.

# Reason

The window depends on how fast prices change.

# Options considered

| Option | Gains  | Costs  |
| ------ | ------ | ------ |
| A      | Gain a | Cost a |
| B      | Gain b | Cost b |
