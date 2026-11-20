---
type: Workflow Critique
schema: workflow/critique/1
title: Critique of the retry policy plan
description: A blind review of the retry policy plan; the most serious finding is retries may charge twice.
generated: {by: critiquing-plan/0.1.0, at: 2026-10-21T09:00:00+02:00}
plan: plan
lenses:
  - red-team
blind: true
---

# First pass, blind

1. Retries may charge twice
   - Severity: high
   - Confidence: likely
   - Section: `# Success criteria`
   - Failure mode: A retried payment that reached the gateway the first time is charged twice, because the plan never says whether the retry reuses the idempotency key.
   - Evidence: The plan and the decisions say nothing about idempotency keys.
   - Direction: Decide whether a retry reuses the idempotency key of the first attempt.
