---
type: Workflow Research
schema: workflow/research/1
title: Research for saved-searches
description: What exists today for saved-searches and the questions it leaves.
generated: {by: researching-change/0.1.0, at: 2026-10-21T10:00:00+02:00}
subject: saved-searches
questions:
  - id: storage-location
    text: Do saved searches live on the device or in the account?
  - id: notify-on-new-results
    text: Does the app notify when a saved search has new results?
  - id: search-limit
    text: How many saved searches may one user keep?
---

# Bottom line

Three questions remain.

# Open decisions

| Id                      | Question                                                 | Options                       | Recommendation              |
| ----------------------- | -------------------------------------------------------- | ----------------------------- | --------------------------- |
| `storage-location`      | Do saved searches live on the device or in the account?  | On the device; in the account | Inference: the first option |
| `notify-on-new-results` | Does the app notify when a saved search has new results? | Notify; do not notify         | Inference: the first option |
| `search-limit`          | How many saved searches may one user keep?               | A fixed cap; no cap           | Inference: the first option |
