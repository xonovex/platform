---
type: Workflow Commit Message
schema: workflow/commit-message/1
title: Commit message for the export model
description: The commit adds the export model for the orders of one period.
generated: {by: preparing-delivery/5.3.0, at: 2026-11-18T15:20:00Z}
work: work/export-model-q3va
branch: export-model-q3va
---

feat(export): build the export of the orders of a period

The export model selects the orders of a closed period and builds one row for each order, so the export API can start an export without reading the order store itself.

Plan: shop-api: plans/order-export/plan.md
