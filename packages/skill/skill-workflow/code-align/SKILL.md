---
description: "Compare two similar implementations, surface differences, and suggest a unified design. Use when the user asks to align, reconcile, harmonize, or compare two pieces of similar code. Keywords: code alignment, reconcile, harmonize, similar implementations, dedupe, unify design."
---

# /xonovex-workflow:code-align – Research Code Alignment Between Similar Implementations

Analyzes two similar implementations to identify structural differences, inconsistencies, and opportunities for alignment. Generates a detailed research report. Does NOT create plans or make changes - run `/xonovex-workflow:plan-create` afterward to create an implementation plan.

## Core Workflow

**Use Task agents with subagent_type=Explore and model=haiku for codebase analysis. Do NOT use EnterPlanMode.**

1. **Discovery & Loading**
   - Resolve files from paths or glob patterns
   - Read both file contents
   - Detect code type (API route, processor, etc.)

2. **Structural Analysis** - Compares:
   - Imports & dependencies: matching imports, missing imports, import order
   - Interfaces & types: definitions, properties, type consistency
   - Functions & exports: signatures, helpers, parameters
   - Code structure: control flow, error handling, logging, naming
   - Logic & algorithms: processing logic, edge cases, bugs
   - Configuration: default values, constants

3. **Alignment Report** - Generates:
   - Summary: Alignment percentage
   - Critical Differences: Bug-causing issues
   - Structural Differences: Inconsistent patterns
   - Missing Features: Functionality in one but not other
   - Recommendations: Prioritized changes

Example output:

```
=== Alignment Report ===
Overall Alignment: 72% aligned

Critical Differences:
  [X] File2 missing error handling in process() function
  [X] File1 uses deprecated API vs File2 modern approach

Structural Differences:
  [WARN]  Different control flow (if/else vs switch)
  [WARN]  Different logging (console.log vs logger.info)
  [WARN]  Inconsistent defaults (maxItems: 100 vs 200)

Missing Features:
  [+] File1 has metrics tracking
  [+] File2 has deduplication logic

Recommendations:
  1. Add error handling to File2
  2. Align logging approach
  3. Standardize control flow
  4. Sync configuration values
```

4. **Interactive Mode** (if requested): Asks questions about reference implementation, control flow preference, default values, missing features

## Next Steps

After running this research command:

1. Review the alignment report for accuracy
2. Run `/xonovex-workflow:plan-create` to create an implementation plan from this research

## Error Handling

- File not found: Verify paths exist
- Invalid glob: Check pattern syntax
- Parse errors: Ensure valid TypeScript/JavaScript
- Validation failures: Review changes, fix test/lint errors
