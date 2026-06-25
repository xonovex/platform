---
name: command-guide
description: "Use when authoring, reviewing, merging, or simplifying reusable user-invocable prompt files (a.k.a. slash commands — files an agent harness exposes as `/command` invocations). Triggers on edits to prompt / command files, on prompts to make a slash command, save a workflow as a reusable command, merge across commands, or condense a verbose command — even when the user doesn't say 'prompt' or 'slash command'."
---

# Command Guidelines

Author, merge, and simplify reusable user-invocable prompt files (sometimes called slash commands, custom commands, or user prompts depending on the agent harness).

## Core Principles

- **Generic by Default** — strip project-specific paths, domain terms, and tech names at author time so prompts reuse across projects
- **Match Style and Voice** — preserve the target's structure, voice, formatting when merging
- **Structure Integrity** — front matter, Goal, Arguments, Core Workflow, Implementation, Error Handling are the load-bearing sections
- **Safe Modifications** — preview changes (`--dry-run`) before applying
- **Bound the Body** — target <150 lines per prompt; anything longer usually wants to be two prompts

## Gotchas

- A prompt that hardcodes one repo's paths/domain terms isn't reusable — generalize at author time, not later
- Argument shape is the prompt's public contract — merging different argument styles silently breaks callers
- "Auto-generated name" + an existing file is a silent overwrite risk — always check before write
- Generalizing too aggressively destroys the example's instructional value — keep enough specificity to learn from

## Operations

- **Create** a new prompt from a completed task — see [references/create.md](references/create.md)
- **Merge** elements from one prompt into another — see [references/merge.md](references/merge.md)
- **Simplify** a verbose prompt — see [references/simplify.md](references/simplify.md)

## Progressive Disclosure

- Read [references/create.md](references/create.md) - Load when capturing a completed task or workflow as a reusable prompt
- Read [references/merge.md](references/merge.md) - Load when porting elements from one prompt into another
- Read [references/simplify.md](references/simplify.md) - Load when condensing a verbose prompt
- Read [references/harness-formats.md](references/harness-formats.md) - Load when authoring for a specific agent harness, migrating a prompt between harnesses, or deciding which format to target (Claude Code, Cursor, GitHub Copilot, Gemini CLI, Continue, Aider, Cline, Roo Code, OpenCode, Pi)
