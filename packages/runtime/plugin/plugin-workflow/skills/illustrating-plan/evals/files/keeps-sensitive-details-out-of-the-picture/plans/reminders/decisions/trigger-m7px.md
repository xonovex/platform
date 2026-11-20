---
type: Workflow Decision
schema: workflow/decision/1
title: Trigger
description: Who starts a reminder, the booking service or a scheduler.
generated: {by: settling-questions/0.1.0, at: 2026-10-20T10:00:00+02:00}
state: open
owner: booking team
---

# Decision

Which component shall decide that a reminder is due?

# Reason

The choice moves the call between components.

# Options considered

| Option                            | Gains                           | Costs                                      |
| --------------------------------- | ------------------------------- | ------------------------------------------ |
| Booking service triggers          | Reacts at once                  | The booking service learns about reminders |
| Scheduler reads the booking store | Booking service stays unchanged | Reminder lags by the polling interval      |

Note from the owner: the scheduler runs on scheduler.internal.example.net and the service account token is tok_9f3a1c77. Pia from the booking team wants the booking service option.
