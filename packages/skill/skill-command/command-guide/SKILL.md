---
name: command-guide
description: "Use when authoring, reviewing, merging, simplifying, or distilling reusable user-invocable prompt files (a.k.a. slash commands — files an agent harness exposes as `/command` invocations). Triggers on edits to prompt / command files, on prompts to make a slash command, save a workflow as a reusable command, merge across commands, condense a verbose command, or distill a fat command into a thin command that delegates to a skill — even when the user doesn't say 'prompt' or 'slash command'."
---

# Command Guidelines

Author, merge, simplify, and distill reusable user-invocable prompt files (sometimes called slash commands, custom commands, or user prompts depending on the agent harness).

## Core Principles

- **Generic by Default** — strip project-specific paths, domain terms, and tech names at author time so prompts reuse across projects
- **Match Style and Voice** — preserve the target's structure, voice, formatting when merging
- **Structure Integrity** — frontmatter, Arguments, and Delegation are the essential
  thin-command sections; the owning skill carries the procedure, output, and errors
- **Safe Modifications** — preview changes (`--dry-run`) before applying
- **Bound the Body** — target 15-40 lines for a delegating command and <150 lines for
  a harness that cannot delegate
- **Delegate, Don't Duplicate** — a command owns its argument contract and delegates the procedure to a guideline skill via the `Skill` tool; the skill is the single source of truth, the command a thin, stable interface
- **Depend on Skills Two Ways** — a command depends on a skill either **hard** (name the exact skill and, when the command harness installs declared dependencies, add it to the command plugin's `dependencies`) or **soft** (describe the capability needed and let the agent select the best-fitting installed skill at run time, declaring nothing and degrading gracefully when none matches). Use hard when one specific skill is always required, soft when several interchangeable skills could satisfy it. If the target harness distributes skills but not commands, publish and invoke the skill directly instead of inventing a command plugin surface. Either way, explicitly load the selected skill at run time (install ≠ in-context), see [references/distill.md](references/distill.md)

## Gotchas

- A prompt that hardcodes one repo's paths/domain terms isn't reusable — generalize at author time, not later
- Argument shape is the prompt's public contract — merging different argument styles silently breaks callers
- "Auto-generated name" + an existing file is a silent overwrite risk — always check before write
- Generalizing too aggressively destroys the example's instructional value — keep enough specificity to learn from
- A command-capable plugin that delegates to a skill but omits a supported dependency declaration can't guarantee the skill at run time — wire the dependency and load via the `Skill` tool (install ≠ in-context)

## Operations

- **Create** a thin command contract that delegates to one owning skill — see [references/create.md](references/create.md)
- **Merge** elements from one prompt into another — see [references/merge.md](references/merge.md)
- **Simplify** a verbose prompt — see [references/simplify.md](references/simplify.md)
- **Distill** a fat command into a thin skill-delegating command — see [references/distill.md](references/distill.md)

## Progressive Disclosure

- Read [references/create.md](references/create.md) - Load when creating a thin
  user-invocable command, selecting its owner skill/plugin, or defining its arguments
- Read [references/merge.md](references/merge.md) - Load when porting elements from one prompt into another
- Read [references/simplify.md](references/simplify.md) - Load when condensing a verbose prompt
- Read [references/distill.md](references/distill.md) - Load when refactoring a self-contained command into a thin delegator that loads its guideline skill at run time
- Read [references/harness-formats.md](references/harness-formats.md) - Load when authoring for a specific agent harness, migrating a prompt between harnesses, or deciding which format to target (Claude Code, Codex, Cursor, GitHub Copilot, Gemini CLI, Continue, Aider, Cline, Roo Code, OpenCode, Pi)
