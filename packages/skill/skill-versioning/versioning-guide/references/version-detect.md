# version-detect: Find Packages Whose Version Changed

List the workspace packages whose `version` differs between the working tree and a git ref (default the previous commit).

## Goal

- Identify which publishable packages changed version since a baseline ref, to drive the publish/release step that follows.

## Core Workflow

1. **Enumerate** the workspace packages (e.g. from the monorepo project list).
2. For each, **read** the working-tree `package.json`; skip it when it has no `name` or is `private`.
3. **Read the baseline** with `git show <ref>:<path>/package.json`; if the file did not exist at `<ref>`, the package is new — skip it (or include it per policy).
4. **Compare** the `version` fields; collect packages where they differ.
5. **Emit** the collected identifiers (e.g. a JSON array) for the caller.

Default `<ref>` is `HEAD~1`; accept an override (a `--ref` flag or positional argument).

## Output

A machine-readable list, e.g.:

```
["@scope/pkg-a","@scope/pkg-b"]
```

## Error Handling

- **Skip (not error)** — `git show` fails because the package or file is absent at `<ref>` → treat the package as new and skip it rather than aborting.

## Gotchas

- `git show <ref>:…` reads **committed** state only — an uncommitted working-tree bump shows as changed, and a staged-but-uncommitted change is still judged against `<ref>`.
- Keep the detect ref and any downstream "already bumped" check on the **same** baseline — detecting against `HEAD~1` while idempotency compares against `HEAD` can disagree.
- A package present now but absent at `<ref>` is **new**, not "changed" — decide explicitly whether new packages belong in the output.
