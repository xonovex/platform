# simplify: Simplify Instruction Files

**Guideline:** Reduce AGENTS.md/CLAUDE.md verbosity by 40-50% while preserving structure, workflows, and technology names.

**Rationale:** Instruction files grow verbose over time. Simplification improves scannability and readability while preserving essential information, commands, and workflows needed for development.

**Example:**

```markdown
# BEFORE (verbose, 145 lines)

## Setup and Configuration

The platform requires initialization through a multi-step process. First, you must install all necessary dependencies using the npm package manager. After installation, you should pull all large files from Git LFS to ensure you have the complete codebase available locally.

The setup process involves:

1. Running npm install to fetch all dependencies
2. Running git lfs pull to download large binary files

Once setup is complete, you can begin development work.

## Running Tasks with Moon

The project uses Moon as a task orchestration system. To run tasks, use the `npx moon run` command with the format `<project>:<task>`. You can also run tasks by tag using the `#<tag>:<task>` syntax...

# AFTER (simplified, 80% reduction, 29 lines)

## Setup

- Dependencies → `npm install`
- Large files → `git lfs pull`

## Running Tasks

- Project task → `npx moon run <project>:<task>`
- Tag-based → `npx moon run #<tag>:<task>`
- Tenants → `npx moon run #tenant-<name>:<task>`
```

**Techniques:**

- Measure baseline line count to track reduction
- Remove verbose prose and redundant descriptions
- Condense multi-line bullets to single lines with inline details
- Convert code blocks to inline arrow notation (`→`)
- Preserve section hierarchy, technology names, and integration points
- Keep command examples with actual tool names
