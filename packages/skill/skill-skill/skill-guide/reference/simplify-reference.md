# simplify-reference: Condense Verbose Reference Files

**Guideline:** Reduce reference/\*.md files by merging overlapping sections, trimming redundant examples, and removing prose that restates headings.

**Rationale:** Reference files are the detail layer of progressive disclosure. When they grow verbose with overlapping sections or redundant examples, they defeat their purpose as scannable technical references.

**Example:**

```markdown
# BEFORE (task-inheritance.md, 139 lines, 3 overlapping merge sections)

## Deep Merging (Moon 2.0)

- Configs merge sequentially: global → extends → local
- fileGroups combine instead of replace
- deps arrays merge from inherited + project tasks
- command arrays merge (use script: to fully override)

## Preventing Merge Issues

- Use script: instead of command: to fully replace inherited commands
- Project-level script: completely overrides template task
  [code example]

## Merge Strategy Overrides

- Merging applies to args, deps, env, inputs, outputs, toolchains...
  [code example]

# AFTER (task-inheritance.md, 110 lines, single Merging section)

## Merging

Configs merge sequentially: global → extends → local.

- fileGroups combine instead of replace
- command arrays merge (use script: to fully replace)
- args, deps, env, inputs, outputs, toolchains merge via configurable strategies
  [single unified code example covering command/script/mergeDeps]

# Reduction: 3 sections → 1, redundant examples removed, all content preserved
```

**Techniques:**

- Read the full reference file and identify overlapping sections
- Merge sections that cover the same concept from different angles
- Consolidate multiple code examples into fewer examples that demonstrate more
- Remove prose that restates what the heading already says
- Remove redundant code examples already shown in top-level example block
- Keep the structure: `# title`, `**Guideline:**`, `**Rationale:**`, `**Example:**`, `**Techniques:**`
- Preserve all distinct technical content — merge, don't delete
- Target 20-40% reduction (reference files are already the detail layer)
- Do not move content to sub-files — reference is the bottom of the disclosure hierarchy
