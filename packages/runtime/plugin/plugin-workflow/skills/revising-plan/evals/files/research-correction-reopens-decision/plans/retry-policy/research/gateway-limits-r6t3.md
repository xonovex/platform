---
type: Workflow Research
schema: workflow/research/1
title: Gateway limits corrected
description: The gateway sheds load after the second request in a minute, not the fourth.
generated: {by: researching-change/0.1.0, at: 2026-10-22T09:00:00+02:00}
subject: How much load the payment gateway accepts before it sheds requests
kind: investigation
corrects:
  - research/gateway-limits-q4w2
---

# Bottom line

The gateway sheds load after the second request in a minute, not the fourth as the earlier research said.

# Observed

- The gateway sheds load after the second request in a minute, not the fourth as the earlier research said.
