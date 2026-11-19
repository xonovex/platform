---
type: Workflow Decision
schema: workflow/decision/1
title: Transport
description: Reminders travel over a message queue.
generated: {by: settling-questions/0.1.0, at: 2026-10-20T10:00:00+02:00}
verified: {by: human:ana, at: 2026-10-20T10:30:00+02:00}
state: settled
---

# Decision

When a booking changes, the booking service shall publish an event to the message queue and the sender shall consume it.

# Reason

The queue already carries booking events.

# Options considered

| Option        | Gains                   | Costs                |
| ------------- | ----------------------- | -------------------- |
| Message queue | Existing infrastructure | Needs a consumer     |
| Direct call   | No new consumer         | Couples the services |
