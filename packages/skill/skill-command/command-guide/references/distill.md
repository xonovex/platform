# distill: Distill a Fat Command into a Skill Delegator

Move a self-contained command's reusable procedure into one owner skill while keeping
its public argument contract stable.

## Goal

- Move procedure, output, errors, and gotchas into the owner skill.
- Keep frontmatter, Arguments, and an explicit Delegation block in the command.
- Distinguish guide name, plugin name, operation, and optional supporting capabilities.
- Preserve callers while reducing duplicated prompt content.

## Arguments

- `command-file` (required): exact command to distill.
- `--owner-skill` (optional): exact guide that owns the procedure.
- `--owner-plugin` (optional): distribution plugin for the owner guide.
- `--operation` (optional): operation/reference within the owner skill.
- `--dry-run` (optional): preview command, skill, and manifest changes.

## Content boundary

Keep in the command:

- description, allowed tools, and argument hint;
- argument names, defaults, repeatability, and effect boundary;
- exact owner skill, owner plugin, and operation;
- brief natural-language supporting needs that the owner operation should select from
  installed skill names and descriptions.

Move to the owner skill:

- goal and procedure;
- output contract;
- examples;
- error handling;
- gotchas.

## Core workflow

- [ ] Parse the existing public arguments before changing any content.
- [ ] Resolve one concept owner. Reuse an existing owner rather than copying.
- [ ] Resolve the owner plugin separately and verify that it distributes the guide.
- [ ] Move the reusable procedure to the owner's operation reference and register the
      operation in its progressive-disclosure index.
- [ ] Describe supporting interchangeable skills in Delegation. Use exact
      dependencies only when the procedure cannot complete without one named guide.
- [ ] Replace the command body with Arguments and Delegation while preserving its
      argument contract.
- [ ] Add the skill-loading capability and wire supported plugin dependencies.
- [ ] Preview or apply, then validate command syntax, argument parity, links,
      manifests, and the skill.

## Output

Report:

- before/after command line count;
- preserved argument contract;
- owner guide, plugin, and operation;
- added soft capability descriptions or exact hard dependencies;
- command, skill, and manifest files affected;
- validation result.

## Error handling

- Already thin → validate and report without rewriting.
- No concept owner → create or extend the owner skill first.
- Multiple possible owners → stop and ask; never merge ownership by prompt order.
- Plugin does not distribute the named guide → stop and report the mismatch.
- Public argument drift → restore the original contract or declare a separate
  migration.

## Gotchas

- `--owner-skill` names a guide; `--owner-plugin` names its distribution package.
- Supporting skills do not become co-owners of the command procedure.
- Installation is not loading; the delegation must load the owner at runtime.
- Copying instead of moving leaves two sources of truth.
