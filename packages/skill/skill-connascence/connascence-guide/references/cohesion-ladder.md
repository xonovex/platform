# The Cohesion Ladder

Cohesion is the dual of coupling: how well one module's own parts belong together — grade a module by its rung, then re-cut it upward.

## The ladder (worst to best)

Push every module up this ladder. Each rung names a stronger reason its contents sit together.

- **Coincidental (worst)** — parts share a module for no reason at all; a `utils` / `helpers` grab-bag. No principle predicts what lands here.
- **Logical** — parts do the same _kind_ of thing, selected by a flag; an `IO` module that reads files _or_ sockets depending on a mode argument. This is control coupling turned inward.
- **Temporal** — parts are grouped because they run in the same phase; a `setup/` blob of unrelated init. The trap rung — see below.
- **Procedural** — parts run in a fixed order but on unrelated data; "do step A, then B, then C" with no shared subject.
- **Communicational** — parts act on the same data; several functions all reading and writing one record.
- **Sequential** — one part's output feeds the next's input; a clean pipeline with a single subject flowing through.
- **Functional (best)** — every part serves one job; remove any one and the job breaks. The module has one reason to exist and one reason to change.

## Aim for functional

- **A well-cut module does one job** — and everything in it serves that job. One pricing rule, one cart total, one config parser: the file name predicts every line.
- **Test the cut by deletion** — if you can drop a function and the module still does its stated job, that function never belonged; it inflated the rung downward.
- **Cohesion and coupling move together** — a functionally cohesive module exposes one narrow surface, so callers couple to less. Low cohesion leaks: a grab-bag forces callers to know which half they meant.

```go
// Functionally cohesive: the file's one job is the cart total.
// Every function serves "compute the price of this cart".
func (c Cart) Total() Money              { /* ... */ }
func (c Cart) Subtotal() Money           { /* ... */ }
func (c Cart) lineTotal(item Item) Money { /* ... */ }
```

## Temporal cohesion is the trap

- **Grouping by _when_ code runs is the package-by-layer / flowchart mistake** — a `setup/` package holding unrelated config, logging, and cache init groups three jobs because they all happen "at startup". That phase is not a job.
- **Split by what changes together, not by when it runs** — config parsing changes with the config schema, cache wiring with the cache, logging with the log format. A shared phase forces three reasons to change into one file; an edit to one risks the other two.
- **The tell** — the only sentence that describes the module uses a time word ("first", "during init", "on teardown"). If you cannot name the module's job without naming a phase, the cohesion is temporal.

```go
// BAD — temporal: setup() groups three concerns because they run in one phase.
// Editing the cache touches the same file as config parsing; nothing shares a subject.
func setup(app *App) {
    app.Config = parseConfig(path)         // config concern
    app.Cache = newCache(app.Config.Size)  // cache concern
    app.Log = newLogger(app.Config.Level)  // logging concern
}

// GOOD — functional: each concern owns the part that changes with it.
// Each function has one job and one reason to change.
func parseConfig(path string) Config { /* parse the file */ }
func newCache(size int) Cache        { /* size and evict */ }
func newLogger(level string) Logger  { /* format and sink */ }
```

## How to use it

- **Grade, then re-cut** — name a module's rung. If it is below sequential, find the seam: which parts share a subject, which share only a phase. Split off the phase-grouped parts into the modules whose job they actually serve.
- **Functional cohesion is the goal, temporal the warning** — when a new file's name wants to be a phase (`setup`, `init`, `teardown`), stop and ask which job each line belongs to.
- **Pair this with the coupling grade** — strength of co-change is the other half. For the unit of coupling and the rules that drive a re-cut, see [connascence](connascence.md).

Back to the overview: [SKILL.md](../SKILL.md).
