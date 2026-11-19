---
type: Workflow Plan
schema: workflow/plan/1
title: Shop retry plan
description: Shop retry plan.
generated: {by: writing-plan/0.1.0, at: 2026-10-22T10:00:00+02:00}
---

# Objective

Retries in the shop app stop at a limit that operators set.

# Scope

In: the retry helper and the settings loader. Out: backoff.

# Success criteria

- Retries stop after the configured limit.
