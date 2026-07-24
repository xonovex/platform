# create: Create a Thin User-Invocable Command

Create a stable harness-specific argument contract and delegate its reusable procedure
to one owning skill.

## Goal

- Keep the user-facing command short and predictable.
- Put procedure, output format, error handling, and gotchas in one owner skill.
- Separate the skill's guide name from its distribution plugin.
- Keep exact installation dependencies distinct from optional description-based selection.
- Preview safely and never overwrite an existing command by accident.

## Arguments

- `description` (required) — what the caller wants the command to accomplish.
- `--name` (optional) — kebab-case command name.
- `--owner-skill` (optional) — exact guide that owns the procedure.
- `--owner-plugin` (optional) — distribution plugin for the owner skill when the
  harness supports declared command dependencies.
- `--operation` (optional) — operation/reference name within the owner skill.
- `--interactive` (optional) — ask about arguments and delegation choices.
- `--dry-run` (optional) — print every proposed artifact without writing.
- `--force` (optional) — replace one exact existing command only after preview.

## Core workflow

- [ ] Identify the public arguments, defaults, repeatability, and effect boundary.
- [ ] Select the target harness and its supported command format.
- [ ] Resolve one exact owner skill and operation. If none owns the procedure, create
      or extend that skill before the command.
- [ ] Keep only frontmatter, Arguments, and Delegation in a harness that can load
      skills. Put procedures and errors in the owner skill reference.
- [ ] Describe any optional supporting capability briefly in Delegation and let the
      owner operation select it from installed skill names and descriptions.
- [ ] Wire the owner plugin as a hard installation dependency where the harness
      supports it. Installation still requires explicit runtime loading.
- [ ] Validate metadata, argument-hint/body parity, delegation, dependency wiring,
      name, location, and line budget.
- [ ] Preview all files. Refuse an existing target unless `--force` explicitly names
      that target, then write and revalidate.

For a harness without command-to-skill delegation, keep the smallest supported prompt
surface and distribute or invoke the skill directly when possible.

## Output

Report:

- command path and harness;
- public argument contract;
- owner guide, owner plugin, and operation;
- optional supporting-capability descriptions;
- manifest changes;
- validation result;
- whether the run previewed, created, or replaced a file.

## Error handling

- Vague description or ambiguous arguments → ask before defining the public contract.
- Existing name without `--force` → stop and show the collision.
- Missing owner skill → route to skill authoring rather than embedding a fat workflow.
- Owner skill and plugin mismatch → stop before changing a manifest.
- Unsupported command surface → publish/invoke the skill directly or use the nearest
  harness-native prompt mechanism.

## Gotchas

- A plugin name is not a guide name; keep both explicit.
- One owner does not prevent several runtime-selected supporting skills.
- A command dependency installs a skill but does not load it into context.
- `--dry-run` must include the command, skill reference, and manifest changes.
