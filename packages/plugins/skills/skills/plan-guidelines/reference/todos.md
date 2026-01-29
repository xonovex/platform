# todos: Scan and Group TODO Comments

**Guideline:** Scan for TODO comments, group by intent, and generate research report.

**Rationale:** TODO comments accumulate and become difficult to track. Scanning and grouping by intent helps prioritize work, identify systemic issues, and convert notes into actionable plans.

**Example:**

```
Scan Results: 47 TODO comments found

Group 1: Auth improvements (12 occurrences, 3 files)
- "TODO: Add 2FA support"
- Files: services/auth, packages/web, packages/admin
- Priority: HIGH (blocks security audit)
- Action: Convert to plan-create task

Group 2: Performance optimizations (18 occurrences, 7 files)
- "TODO: Add caching layer"
- "TODO: Optimize database queries"
- Files: across services
- Priority: MEDIUM
- Action: Create code-simplify analysis

Group 3: Deprecations (17 occurrences, 5 files)
- "TODO: Remove legacy API"
- Files: api routes
- Priority: LOW
- Action: Schedule for v2.0 release
```

**Techniques:**

- Scan directory recursively for TODO patterns (`TODO:`, `FIXME:`, `NOTE:`)
- Extract unique TODO messages, normalized for whitespace consistency
- Group TODOs by similarity: identical text, conceptual intent, file patterns
- Map each group to source files, count occurrences and affected files
- Infer applicable skills from file extensions and framework indicators
- Categorize by priority: blocking work, technical debt, nice-to-have
- Generate report with groups, file counts, and action recommendations
- Convert high-priority TODOs into tasks for tracking and execution
