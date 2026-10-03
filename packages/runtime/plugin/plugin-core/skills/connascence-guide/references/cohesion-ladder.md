# The Cohesion Ladder

Cohesion is the dual of coupling: how well one module's own parts belong together. Grade a module's rung, then re-cut it upward. Strength of co-change is the other half, see [connascence](connascence.md).

## The ladder (worst to best)

- **Coincidental (worst)**: parts share a module for no reason; a `utils`/`helpers` grab-bag.
- **Logical**: parts do the same _kind_ of thing, selected by a flag (an `IO` module reading files _or_ sockets by mode). Control coupling turned inward.
- **Temporal**: parts grouped because they run in the same phase; a `setup/` blob. The trap rung.
- **Procedural**: parts run in a fixed order but on unrelated data.
- **Communicational**: parts act on the same data.
- **Sequential**: one part's output feeds the next's input; a pipeline with a single subject.
- **Functional (best)**: every part serves one job; remove any one and the job breaks. One reason to exist, one reason to change.

**Test the cut by deletion**: if you can drop a function and the module still does its stated job, that function never belonged.

## Temporal cohesion is the trap

Grouping by _when_ code runs (the package-by-layer mistake) forces several jobs into one file. Split by what changes together, not when it runs.

```go
// BAD, temporal: setup() groups three concerns because they run in one phase.
func setup(app *App) {
    app.Config = parseConfig(path)         // config concern
    app.Cache = newCache(app.Config.Size)  // cache concern
    app.Log = newLogger(app.Config.Level)  // logging concern
}
// GOOD, functional: each concern owns the part that changes with it.
func parseConfig(path string) Config { /* ... */ }
func newCache(size int) Cache        { /* ... */ }
func newLogger(level string) Logger  { /* ... */ }
```

Back to the overview: [SKILL.md](../SKILL.md).
