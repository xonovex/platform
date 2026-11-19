---
type: Workflow Research
schema: workflow/research/1
title: Offline facts
description: The mobile and web clients handle a lost connection differently
generated: {by: researching-change/0.1.0, at: 2026-10-20T09:00:00+02:00}
subject: The mobile and web clients handle a lost connection differently
questions:
  - id: offline-sync-window
    text: How long may the mobile client show its last synced list?
sources:
  - id: m
    resource: shop-mobile@1a2b3c4:src/list.ts
    title: Mobile list store
  - id: w
    resource: shop-web@5d6e7f8:src/list.ts
    title: Web list view
---

# Bottom line

The mobile client shows its last list; the web client shows an error.[^m]

# Observed

- Mobile caches the list.[^m]
- The web client shows an error state.[^w]
- The sync endpoint accepts a shared token that is printed in the service logs.[^m]

# Inference

None.

# Unknowns

None.

# Open decisions

| Id                    | Question                               | Options          | Recommendation |
| --------------------- | -------------------------------------- | ---------------- | -------------- |
| `offline-sync-window` | How long may the mobile list be shown? | 1 hour; 24 hours | None           |

# Not checked

Nothing ran.

[^m]: Mobile list store, lines 1 to 60

[^w]: Web list view, lines 1 to 40
