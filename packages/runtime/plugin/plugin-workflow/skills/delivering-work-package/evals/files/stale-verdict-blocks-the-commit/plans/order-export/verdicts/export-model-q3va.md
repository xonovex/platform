---
type: Workflow Verdict
schema: workflow/verdict/1
title: Verdict for the export model
description: 2 pass, 0 fail, 0 blocked.
generated: {by: validating-work-item/5.3.0, at: 2026-11-18T15:02:00Z}
work: work/export-model-q3va
worktree_hash: 3f9c2b7e41d05a8c6e2f9b1d7a4c08e5b63f2d91
source: independent
results:
  - criterion: c1
    result: pass
    evidence: "npm test -- export-model: 14 passed, including the test for orders on the period boundary"
  - criterion: c2
    result: pass
    evidence: "npm run bench -- export-model: p95 171 ms over 10000 orders"
---

# Result

2 pass, 0 fail, 0 blocked. Tree 3f9c2b7e41d0, independent.
