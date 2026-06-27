# Robustness

Audit existing code for how it behaves at boundaries and under bad input — type safety, validation, error handling, and logging — then grade each finding by severity; this is a read-only pass, you flag and grade, you do not edit.

Read the project's own `AGENTS.md` / guidelines / linter + type-checker config **first**: they decide what counts as a violation. A pattern the project bans is high severity even if benign in general; a pattern it explicitly allows is not a finding.

## What to look for

Group findings by category, then grade each by severity.

| Category           | Signals to flag                                                                                                                                                                                             |
| ------------------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Type safety**    | `any` / dynamic / untyped escapes, implicit-any params, unchecked type assertions/casts, non-null assertions on untrusted values, `unknown` never narrowed, stringly-typed enums                            |
| **Validation**     | unvalidated external input, no schema at the boundary, missing guards/range/null checks, validating after use, trusting parsed shape without parsing                                                        |
| **Error handling** | swallowed/empty catch, caught-and-logged-then-continue, errors mapped to a sentinel (`null`/`-1`/`""`), bare re-throw losing context, unhandled rejection/async error, broad catch hiding distinct failures |
| **Logging**        | error path with no log, log missing context (id/operation/cause), inconsistent levels (errors at `info`, noise at `error`), secrets/PII in logs, log-and-rethrow double-reporting                           |
| **Code smells**    | long function (>30 lines), deep nesting (>3 levels), high cyclomatic complexity, boolean/flag params steering branches, primitive obsession at boundaries, duplicated guard logic                           |

## Grading by severity

- **High** — a trust-boundary input reaches logic unvalidated; an error is silently swallowed so a failure looks like success; a type escape (`any`/unchecked cast) defeats checking on untrusted data; secrets logged. Wrong-answer or security-relevant.
- **Medium** — internal type escape, lost error context (cause dropped, generic message), error path with no log, broad catch collapsing distinct failures, long/complex function on a hot path.
- **Low** — naming/level inconsistency, defensive check just inside a trusted call, mild nesting, stylistic primitive obsession. Note it, don't block on it.

Grade by **blast radius and likelihood**, not line count: an unvalidated request body outranks a 60-line pure helper.

## Recognize it fast

- **Boundary map** — where does external data enter (request, env, file, IPC, third-party response)? Each entry without a parse/validate step is a finding.
- **Follow the error** — for each `throw`/reject, is it caught, and does the catch preserve cause + log with context, or quietly degrade?
- **Trace one `any`** — where did the untyped value originate? Usually a single missing schema, not N call sites.

## BAD -> GOOD

Parse at the boundary, don't trust then check later:

```
// BAD — trust the shape, cast it in
const cfg = JSON.parse(raw) as Config;
connect(cfg.host, cfg.port);

// GOOD — parse-don't-validate: the type is earned, not asserted
const cfg = ConfigSchema.parse(JSON.parse(raw)); // throws with field-level detail
connect(cfg.host, cfg.port);
```

Preserve cause and context instead of swallowing:

```
// BAD — failure looks like success
try { return await fetchUser(id); }
catch { return null; }

// GOOD — fail-fast, keep the cause, log once with context
try { return await fetchUser(id); }
catch (cause) {
  log.error("fetchUser failed", { id, cause });
  throw new UserFetchError(id, { cause });
}
```

Narrow `unknown`, don't assert through it:

```
// BAD
const n = (input as { count: number }).count;

// GOOD
if (!isCountPayload(input)) throw new InvalidInput("count");
const n = input.count;
```

## Gotchas (false positives)

- **A pile of `any` is one missing schema** — fix the boundary, not every call site; don't file N type-safety findings for what is a single parse gap.
- **Validate at trust boundaries, trust internal callers** — a guard re-checking an invariant a caller already guaranteed is noise, not robustness; flag the boundary, not the interior.
- **Log at boundaries and on error paths, not everywhere** — "add logging here" inside a pure helper is wrong; missing logs only matter where failures cross a boundary or get handled.
- **`any` an external type definition forces on you** is the library's gap, not the project's — note it, grade it low unless it touches untrusted data.
- **A swallowed error may be intentional** (optional cache read, best-effort cleanup) — confirm there is no silent data loss before grading high; an empty catch with a comment stating why is acceptable.
- **A long function that is a flat declarative sequence** (config, a switch, a builder) is not automatically a smell — complexity is branching and nesting, not raw lines.
- **Defensive null checks the type system already rules out** are dead, not robust — flag them under simplify, not harden.

## See also

- [SKILL.md](../SKILL.md) — the audit method and how to grade
- [smell-catalog.md](smell-catalog.md) — every smell mapped to its detector signal and owner
- An unchecked downcast / cast that breaks the abstraction is the symptom of Refused Bequest / a Liskov violation — see **oop-guide**
- For the coupling vocabulary behind boolean/flag params and primitive obsession, see **connascence-guide**
