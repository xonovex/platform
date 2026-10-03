# simplify: Simplify a User-Invocable Command

Reduce command surface without changing its public contract or creating a second
procedure owner.

## Arguments

- `prompt-file` (required): exact command file to simplify.
- `--dry-run` (optional): preview without writing.
- `--target-reduction` (optional): target reduction percentage; defaults to 50,
  range 30-70.

## Core workflow

- [ ] Resolve the command harness and parse metadata, arguments, and delegation.
- [ ] Record the original public argument and effect contract.
- [ ] If the harness can delegate, keep only metadata, Arguments, and Delegation;
      move reusable procedure into the registered owner operation.
- [ ] If the harness cannot delegate, retain the minimum inline procedure,
      validation, errors, and safety rules needed for execution.
- [ ] Remove duplicated explanations, redundant examples, and project-specific
      coordinates that are not part of the contract.
- [ ] Preview command, owner-skill, and manifest changes; validate argument parity,
      operation registration, dependencies, and line budget.

## Preservation rules

- Preserve command name, namespace, argument names, defaults, repeatability, and
  effect boundary.
- Preserve natural-language supporting-capability needs that affect runtime behavior.
- Preserve the exact owner skill, plugin, and operation unless the caller requests a
  migration.
- Move safety and error behavior to the owner operation for delegating harnesses;
  never delete it merely to meet a reduction target.
- Treat the target percentage as a preference, not a correctness criterion.

## Output

Report before/after line counts, preserved public contract, moved procedure sections,
owner changes, manifest changes, and validation outcome.

## Error handling

- Missing or unrecognized command → stop with the unresolved format.
- No registered owner for a delegating command → route to distill first.
- Required reduction would alter the public contract → keep the larger valid form.
- Existing thin command → validate and return unchanged.

## Gotchas

- A short command can still be fat when it owns procedure duplicated in a skill.
- An inline safety rule belongs in the owner operation before it leaves the command.
- Generic examples must retain enough structure to explain argument semantics.
