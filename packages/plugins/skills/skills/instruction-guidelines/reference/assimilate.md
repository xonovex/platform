# assimilate: Extract Patterns from Source Instructions

**Guideline:** Extract organizational patterns from source AGENTS.md/CLAUDE.md and integrate into target while preserving structure, style, technology names, and context.

**Rationale:** Share organizational patterns between projects without losing project-specific details, technology choices, or critical context that defines how the project operates.

**Example:**

```bash
# Source: source-project AGENTS.md
# Target: target-project AGENTS.md

# Extract workflow pattern from source:
# "Setup: npm install, git lfs pull"
# "Tasks: npx moon run <project>:<task>"

# Apply to target with its technology names:
# Target uses Gradle and Maven instead of npm
# Result: "Setup: gradle build, git lfs pull"
#         "Tasks: gradle run <project>:<task>"

# Extract section grouping (Subdirectories + Workflow + Integration)
# Reorder target to match, preserving custom sections

# Before: Services > Deployment > Infrastructure > Utilities
# After:  Services > Deployment > Infrastructure > Workflow > Integration Points > Utilities

# All paths, tech names, and project context unchanged
# Only organizational structure adopted from source
```

**Techniques:**

- Load both target and source files and analyze their structure
- Extract organizational patterns from source (not content or tech names)
- Preserve target's technology names, paths, and project-specific context
- Rewrite patterns using target's terminology and formatting
- Maintain section order and hierarchy integrity
- Commit before running and verify output
