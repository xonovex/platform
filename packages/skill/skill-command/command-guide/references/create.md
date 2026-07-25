# create: Create a Prompt from a Completed Task

Generate a new reusable prompt file (a.k.a. slash command, user-invocable command) from a recently completed task. Extracts essential steps, makes them generic and reusable.

## Goal

- Convert completed tasks into reusable prompts
- Extract essential steps and make them generic
- Follow a minimal, language-agnostic structure
- Validate against prompt-authoring best practices

## Arguments

- `description` (required) — brief description of what the task accomplished
- `--name` (optional) — prompt name (auto-generated from description if not provided)
- `--interactive` (optional) — ask clarifying questions about arguments, validation, output, error handling

## Core Workflow

1. **Analyze Task** — parse description to identify core goal, required inputs, key steps, tools used, validation, output
2. **Generate Name** — create kebab-case name from description (e.g. "Remove comments" → `code-comments-remove`)
3. **Make Generic** — replace project coordinates with neutral equivalents that preserve shape: keep a path's depth and role (`packages/billing-service/src/` → `packages/example/src/`, not `src/`), swap a domain cluster for a neutral one (`users`/`orders`/`payments` → `items`), drop vendor and service names
4. **Pick Target Format** — determine the target agent harness and its file format (see [harness-formats.md](harness-formats.md) for the per-harness matrix)
5. **Structure Prompt** — thin shape by default (skeleton in [distill.md](distill.md)); self-contained only when no owning skill exists: metadata block + Goal (3-5 bullets) + Arguments + Core Workflow (4-8 steps) + Implementation Details + Error Handling
6. **Validate Structure** — required sections present, metadata block parses, generic examples, no project-specific content, file length <150 lines
7. **Write File** — save to the harness-specific location and extension (see [harness-formats.md](harness-formats.md))

## Sections for a self-contained command (no owning skill)

- Metadata block: at minimum a 1-sentence `description`; add other fields (tools, permissions, argument hint, activation scope) appropriate for the target harness — see [harness-formats.md](harness-formats.md)
- Goal (3-5 bullets), Arguments (required/optional, defaults), Core Workflow (4-8 steps), Implementation Details, Error Handling
- Let the argument contract carry usage instead of example blocks: enumerate allowed values inline (`--tone formal|casual|technical`), state defaults, and mark required arguments as required in the `argument-hint`

## Error Handling

- Description too vague → ask for more details
- Name already exists → suggest alternatives; never overwrite
- Invalid name → must be kebab-case, alphanumeric
- Missing required sections → validate before writing

## Safety

Preview before writing, check for existing files, validate the metadata block parses in the target format, ensure no sensitive/project-specific data.
