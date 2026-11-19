---
type: Workflow Decision
schema: workflow/decision/1
title: Export file format
description: Should the export be CSV or JSON Lines?
generated: {by: settling-questions/5.3.0, at: 2026-11-10T11:00:00Z}
state: open
---

# Decision

Should the export be CSV or JSON Lines?

# Reason

Shop owners open exports in spreadsheets, and the reporting team reads them with scripts.

# Options considered

| Option     | Gains                  | Costs                         |
| ---------- | ---------------------- | ----------------------------- |
| CSV        | Opens in a spreadsheet | Nested fields need flattening |
| JSON Lines | Keeps nested fields    | Needs a tool to open          |
