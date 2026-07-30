# Moon Task Scripts

- One package ships one binary, named `moon-<domain>-<action>-<subject>`. A package that ships none is shared code: `script-moon-common` for anything cross-cutting, `script-moon-skill-eval-common` for the eval harness, `script-moon-skill-catalog-common` for catalog file, budget, and vocabulary reading.
- No script package may depend on `shared-core`. `shared-core` is tagged `npm`, so its `npm-check`, publish, and version tasks depend on the script packages, and the reverse edge makes moon reject the project graph with `would_cycle`. Duplicate the few helpers a script needs into `script-moon-common` instead.
- Run these tools through `npx moon run <project>:<task>`, never `npx <bin>` directly. Each task declares a `~:build` dependency that refreshes `dist/`; `dist/` is gitignored and a moon task output, so invoking the linked bin directly executes whatever build happens to be on disk.
- Pass ad-hoc flags after `--`, for example `npx moon run skill-hono:skill-eval-triggers -- --split validation --runs 3`.
- Use `script:` for anything needing shell features (`case`, pipes, chaining); `command:` arrays are for simple executables and mangle multi-line arguments.
- Model-invoking evals are capped at `MAX_TRIGGER_MODEL_RUNS` per batch. Keep `queryCount * runs` within that cap by setting `--batch-size`, not by trimming queries.
- An eval infrastructure failure invalidates the whole run by design; partial results must never be reported as a pass.
