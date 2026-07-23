# merge: Port Elements from One Command into Another

Merge compatible public-contract elements while preserving one procedure owner.

## Arguments

- `target` (required) — target command file.
- `source` (required) — source command file.
- `--aspects <aspects>` (optional) — arguments, requirements, metadata, or
  owner-skill procedure.
- `--percentage <percent>` (optional) — intensity 10-100; defaults to 50.
- `--interactive` (optional) — ask about incompatible contracts.
- `--dry-run` (optional) — preview without modifying.

## Core workflow

- [ ] Resolve both command formats and whether the target harness can delegate.
- [ ] Compare namespaces, arguments, effect boundaries, requirements, owner skills,
      and operations.
- [ ] Preserve the target's metadata, voice, formatting, and existing public
      arguments.
- [ ] Merge compatible argument or requirement additions into the command contract.
- [ ] For a delegating harness, merge procedure, validation, errors, and examples
      into the target owner skill operation; keep the command thin.
- [ ] For a non-delegating harness, merge only the minimum inline procedure the
      harness needs.
- [ ] Preview the command, owner-skill, and manifest changes, then validate all
      affected artifacts.

## Compatibility rules

- Different owner skills for the same procedure are an ownership conflict; select
  one owner before merging.
- An argument collision with different meaning, defaults, or effect scope is a
  public-contract conflict.
- A source requirement may merge only when its semantic contract and strength remain
  intact.
- A source procedure for a different operation belongs in a separate command.
- `--percentage` controls optional examples and explanatory depth; it never weakens
  safety, validation, requirements, or argument semantics.

## Output

Report the preserved and added arguments, requirements, owner operation, affected
manifests, changed command/skill files, and validation outcome.

## Error handling

- Missing file or unsupported format → stop with the unresolved path or harness.
- Incompatible argument or effect contract → stop and describe the collision.
- Multiple possible owners → ask for one owner rather than combining procedures.
- No new compatible content → return unchanged.

## Gotchas

- Style compatibility does not make two procedures the same operation.
- A delegating command that absorbs procedure content becomes a second source of
  truth.
- A metadata merge must not silently widen tools, effects, or model selection.
