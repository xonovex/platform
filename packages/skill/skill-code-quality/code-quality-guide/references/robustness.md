# Robustness

Read-only audit of how code behaves at boundaries and under bad input — type safety, validation, error handling, logging — flagging and grading each finding, never editing.

Read the project's own `AGENTS.md` / guidelines / linter + type-checker config **first**: they decide what counts as a violation. A pattern the project bans is a finding even if benign in general; a pattern it explicitly allows is not.

## What to flag

| Category           | Signals to flag                                                                                                                                                                   |
| ------------------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Type safety**    | `any` / dynamic / untyped escapes, implicit-any params, unchecked type assertions/casts, non-null assertions on untrusted values, `unknown` never narrowed, stringly-typed enums  |
| **Validation**     | unvalidated external input, no schema at the boundary, missing guards/range/null checks, validating after use, trusting parsed shape without parsing                              |
| **Error handling** | swallowed/empty catch, caught-and-logged-then-continue, bare re-throw losing context, unhandled rejection/async error, broad catch hiding distinct failures                       |
| **Logging**        | error path with no log, log missing context (id/operation/cause), inconsistent levels (errors at `info`, noise at `error`), secrets/PII in logs, log-and-rethrow double-reporting |
| **Code smells**    | long function (>30 lines), deep nesting (>3 levels), high cyclomatic complexity, boolean/flag params steering branches, primitive obsession at boundaries, duplicated guard logic |

## Grading by severity

- **Medium** — internal type escape, lost error context (cause dropped, generic message), error path with no log, broad catch collapsing distinct failures, long/complex function on a hot path.
- **Low** — naming/level inconsistency, defensive check just inside a trusted call, mild nesting, stylistic primitive obsession. Note it, don't block on it.

## Recognize it fast

- **Boundary map** — where does external data enter (request, env, file, IPC, third-party response)? Each entry without a parse/validate step is a finding.
- **Follow the error** — for each `throw`/reject, is it caught, and does the catch preserve cause + log with context, or quietly degrade?

## Gotchas (false positives)

- **Validate at trust boundaries, trust internal callers** — a guard re-checking an invariant a caller already guaranteed is noise, not robustness; flag the boundary, not the interior.
- **Log at boundaries and on error paths, not everywhere** — "add logging here" inside a pure helper is wrong; missing logs only matter where failures cross a boundary or get handled.
- **`any` an external type definition forces on you** is the library's gap, not the project's — note it, grade it low unless it touches untrusted data.
- **A long function that is a flat declarative sequence** (config, a switch, a builder) is not automatically a smell — complexity is branching and nesting, not raw lines.
- **Defensive null checks the type system already rules out** are dead, not robust — flag them under simplify, not harden.

## See also

- [SKILL.md](../SKILL.md) — the audit method and how to grade
- [smell-catalog.md](smell-catalog.md) — every smell mapped to its detector signal and owner
- An unchecked downcast / cast that breaks the abstraction is the symptom of Refused Bequest / a Liskov violation — see **oop-guide**
- For the coupling vocabulary behind boolean/flag params and primitive obsession, see **connascence-guide**
