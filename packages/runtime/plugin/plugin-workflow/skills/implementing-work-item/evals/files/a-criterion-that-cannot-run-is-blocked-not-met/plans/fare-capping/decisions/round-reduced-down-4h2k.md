---
type: Workflow Decision
schema: workflow/decision/1
title: Round reduced totals down
description: A reduced daily total is rounded down to whole cents.
generated: {by: settling-questions/0.1.0, at: 2026-10-21T10:00:00+02:00}
verified: {by: human:pricing-owner, at: 2026-10-21T10:30:00+02:00}
state: settled
---

# Decision

When a percentage reduction applies to a daily total, the system shall round the reduced total down to whole cents.

# Reason

Finance reports in whole cents and never wants a rider charged a fraction more than the rule states.

# Options considered

| Option     | Gains                   | Costs                      |
| ---------- | ----------------------- | -------------------------- |
| Round down | Never charges too much  | Loses up to one cent       |
| Round half | Matches common rounding | Can charge a cent too much |
