---
type: Workflow Decision
schema: workflow/decision/1
title: retry-backoff-m3ab
description: When a retry fails, the system shall wait before the next attempt.
generated: {by: settling-questions/0.1.0, at: 2026-10-22T10:00:00+02:00}
state: settled
owner: platform
---

# Decision

When a retry fails, the system shall wait before the next attempt.

# Reason

The team agreed it.

# Options considered

| Option    | Gains    | Costs     |
| --------- | -------- | --------- |
| A. Chosen | Simple   | A limit   |
| B. Other  | Flexible | More code |
