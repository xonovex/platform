# The Brief for an Implementation Agent

The agent starts with no context, so the brief carries all of it:

- [ ] **Required reading**, by path: the subplan, its parent, completed predecessors, the `AGENTS.md` of each package touched, the source files built on
- [ ] **Carried facts**: predecessors' downstream-impact notes (renamed types, new APIs, changed capacities, moved conventions), so the agent inherits rather than rediscovers them
- [ ] **Skills**: every entry in `skills_to_consult`, loaded before implementation
- [ ] **Hard constraints**: governing decisions restated concretely (units a value carries, versions pinned, which module sees which type, what a hot path may not do), never cited by number
- [ ] **Deliverables and tests**: the subplan's tasks and success criteria, verbatim
- [ ] **Validation commands**: the fast checks of the touched package, strictest presets and format checks; the slow and dependent packages' suites stay with the supervisor, because an agent waiting minutes on a build loses its prompt cache and reprocesses its whole context
- [ ] **Output discipline**: filter long command output to its failures and summary lines, and read files by range after a search, because every line an agent reads stays in the context of all its later calls
- [ ] **Working agreement**: the current branch; do not commit, do not push
- [ ] **Document update**: set `status` and `validation` frontmatter, tick only criteria met, add a Results section in the siblings' factual style, leave `in_progress` with reasons if one is unmet
- [ ] **Required return**: status; files touched; each criterion met or not-met with its evidencing test; validation results; deviations with rationale; downstream impact
