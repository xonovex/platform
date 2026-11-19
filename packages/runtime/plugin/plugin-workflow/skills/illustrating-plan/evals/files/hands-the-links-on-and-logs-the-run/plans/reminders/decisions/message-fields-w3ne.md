---
type: Workflow Decision
schema: workflow/decision/1
title: Message fields
description: Which fields the reminder message carries.
generated: {by: settling-questions/0.1.0, at: 2026-10-20T10:00:00+02:00}
state: open
owner: sender team
---

# Decision

Which fields shall the reminder message carry?

# Reason

The fields decide how much the sender must look up.

# Options considered

| Option       | Gains         | Costs                           |
| ------------ | ------------- | ------------------------------- |
| Minimal ids  | Small message | The sender looks up the booking |
| Full booking | No lookup     | Larger message                  |
