---
type: Workflow Research
schema: workflow/research/1
title: Cache facts
description: The product pages read the database on every request
generated: {by: researching-change/0.1.0, at: 2026-10-20T09:00:00+02:00}
subject: The product pages read the database on every request
questions: []
sources:
  - id: page
    resource: shop-web@9d8e7f1:src/page.ts
    title: Page loader
---

# Bottom line

Every product page reads the database.[^page]

# Observed

- Pages query the database directly.[^page]

# Inference

None.

# Unknowns

None.

# Open decisions

None.

# Not checked

Nothing ran.

[^page]: Page loader, lines 1 to 30
