---
description: Create a new slash command from a completed task or goal
allowed-tools:
  - Read
  - Write
  - Glob
  - TodoWrite
  - AskUserQuestion
argument-hint: "[description] [--name <name>] [--interactive]"
---

# /xonovex-utility:slashcommand-create – Create Slash Command from Task

Automatically generates a new slash command file based on a recently completed task. Extracts essential steps, makes them generic and reusable.

## Goal

- Convert completed tasks into reusable slash commands
- Extract essential steps and make them generic
- Follow minimal, language-agnostic structure
- Validate against slash command best practices

## Usage

```bash
# Create from task description
/xonovex-utility:slashcommand-create "Remove comments from source files while preserving directives"

# Create with specific name and interactive mode
/xonovex-utility:slashcommand-create "Deploy application" --name deploy --interactive
```

## Arguments

- `description` (required): Brief description of what the task accomplished
- `--name` (optional): Command name (auto-generated from description if not provided)
- `--interactive` (optional): Ask clarifying questions about arguments, validation, output, and error handling

## Core Workflow

1. **Analyze Task**: Parse description to identify core goal, required inputs, key steps, tools used, validation, and output
2. **Generate Name**: Create kebab-case command name from description (e.g., "Remove comments" → "code-comments-remove")
3. **Make Generic**: Strip project-specific paths (`packages/myapp/` → `src/`), domain terms (`users` → `items`), and technologies
4. **Pick Target Format**: Determine the target agent harness and its file format (metadata block shape, file extension, install location)
5. **Structure Command**: Create file with metadata block, Goal (3-5 bullets), Usage (2-3 examples), Arguments, Core Workflow (5-8 steps), Implementation Details, and Error Handling
6. **Validate Structure**: Ensure all required sections present, metadata block parses in the target format, generic examples, no project-specific content, file length < 150 lines
7. **Write File**: Save to the agent harness's slash-command directory under the chosen name

## Implementation Details

**Name Generation**: Convert to lowercase, replace spaces with hyphens, remove special characters, prefix with category (e.g., `git-`, `code-`, `test-`)

**Required Sections**:

- Metadata block: at minimum a 1-sentence `description`; format is harness-specific (YAML frontmatter in Claude Code, with optional fields like `argument-hint`, `allowed-tools`, `model`; other harnesses use their own conventions)
- Goal: 3-5 bullet points
- Usage: 2-3 bash examples
- Arguments: List with required/optional, defaults, descriptions
- Core Workflow: 4-8 numbered steps with sub-bullets
- Implementation Details: Technical approach, key functions
- Error Handling: Common errors and solutions

## Error Handling

- Description too vague → ask for more details
- Command name already exists → suggest alternatives or --force flag
- Invalid command name → must be kebab-case, alphanumeric
- Missing required sections → validate before writing

## Gotchas

- A command that hardcodes one repo's paths/domain terms isn't reusable — generalize aggressively at generate time, not later
- A command longer than 150 lines usually means two commands hiding inside — split rather than bloat
- "Auto-generated name" + existing file is a silent overwrite risk — always check before write

## Safety

Preview before writing, check for existing files, validate the metadata block parses in the target format, ensure no sensitive/project-specific data
