# Skills Composability Roadmap

Plan to bring the rest of the Xonovex skill catalog onto the composable-split pattern, and to improve the meta `skill` guide so the pattern is self-sustaining.

## Context

We restructured one cluster (`c99` / `c99-opinionated` / `c99-game-opinionated`, `data-oriented-design`, `memory-management`, `lock-free`, `gpu-rendering`, `gpu-rendering-vulkan`, `data-model`) into a clean composable split and codified the rules in `packages/skill/AGENTS.md` + the `skill` guide's Core Principles. The rest of the catalog still has duplication (notably `hono`↔`hono-opinionated`, `lua`↔`lua-opinionated`), missing cross-links, and some language/project coupling. This plan applies the same pattern everywhere and hardens the `skill` guide so future skills follow it by default.

This is informed by Anthropic's Agent Skills guidance (see Sources): skills load by **progressive disclosure** (metadata ~100 tokens → `SKILL.md` <5k tokens → `references/` on demand), and are **composable by design** — "your skill should work well alongside others, not assume it's the only capability available." Composability therefore means: small single-concern skills, one owner per concept, and cross-references instead of duplication.

## Pattern (the target, already codified)

- **One concept → one owner skill.** Never duplicate; cross-reference the owner _by name_ ("see zod-guide"), since cross-package file links don't resolve.
- **Tiered model:** general (paradigm/principle) → language → framework/opinionated. Specific tiers link _up_ to the general tier for the "why"; the general tier never depends on a specific one. Reference model: `general-fp`/`general-oop` → `typescript`/`python`/`lua`/`c99` → `hono`/`react`/`c99-opinionated`.
- **Generalize** anything not inherently language/API-specific into a general skill.
- **Sources only in `SOURCES.md`** — never name authors/companies/talks inside `SKILL.md` or `references/`.
- **Progressive disclosure:** keep `SKILL.md` an index (essentials + gotchas + one example + progressive-disclosure list); push anything needed <~20% of the time into a `references/` file.
- Register every skill in `.claude-plugin/marketplace.json` (compact, alphabetical); prettier-clean; validate links.

---

## Part A — Improve the `skill` guide (make the pattern self-sustaining)

The `skill` guide already authors single skills well and now states **Composable split** + **Sources in SOURCES.md** principles. Add the catalog-level method so an agent can both _author_ composably and _audit/refactor_ an existing catalog.

1. **New `references/composability.md`** (the depth behind the Core Principle):
   - one concept = one owner; design to coexist with other skills (don't assume sole capability);
   - the tiered general→language→framework model with the c99/lua/hono exemplars;
   - generalize-or-link decision: "is this concept inherently tied to this language/API? if not, it belongs in a general skill the specific one links to";
   - cross-reference by name (not cross-package file links); the general tier must not depend on specific tiers.
2. **New `references/catalog-audit.md`** — the repeatable audit method used here: inventory each skill (description + reference filenames) → find any concept in 2+ skills → name the rightful owner → extract/cross-link/generalize/split/merge → verify no duplication, links resolve, no `SOURCES` leakage. Include the "duplicated reference filenames across two skills = a tiered-split smell" heuristic.
3. **SKILL.md edits:**
   - Broaden the description to include catalog-level work (auditing/splitting/de-duplicating a set of skills), not only single-skill authoring.
   - Add a **progressive-disclosure budget** note to Spec Constraints: discovery is name+description only (~100 tokens) so descriptions decide routing; `SKILL.md` ≤ ~500 lines / 5k tokens; put anything used <~20% of the time in `references/`.
   - Add two Core Principles already reflected in `packages/skill/AGENTS.md`: **Design to coexist** (works alongside other skills) and **Routing-first descriptions** (the description is the router; tune trigger words; debug with "which skill did you use?").
   - Add progressive-disclosure links for `composability.md` and `catalog-audit.md`.
4. Keep the `skill` guide **project-independent** (no repo paths inside it); the repo-specific packaging/registration lives in `packages/skill/AGENTS.md`.

---

## Part B — Remaining-skills refactor roadmap (prioritized)

Apply the per-refactor execution pattern (below) to each. Do one cluster per commit.

### High priority — clear duplication, mirror the c99 tiered model

1. **Hono tier split** — `hono` becomes sole owner of the generic Hono references (application-structure, context-storage, cookie-handling, error-handling, middleware-patterns, middleware-combine, platform-runtime, security-middleware, validation-type-safety, websocket-support, ~10 files currently duplicated into `hono-opinionated`). `hono-opinionated` keeps only its opinionated overlay (controllers, the `openapi-*` set, router-selection, body-limit) and links to `hono-guide` for everything else. Update both SKILL.md bodies + descriptions.
2. **Lua tier split** — `lua` becomes sole owner of the 8 shared fundamentals (module-pattern, local-variables, metatables, coroutines, input-validation, error-handling, string-concatenation, idiomatic-patterns). `lua-opinionated` keeps only `jit-friendly-tables` + `cache-lookups` and links to `lua-guide`. Mirrors c99→c99-opinionated exactly.
3. **Validation + testing cross-links** — `expressjs` and `hono` reference **zod-guide** for schema patterns (they each re-show Zod inline) and **vitest-guide** for test setup; keep only the framework-specific glue locally. Low effort, high clarity.

### Medium priority — tiering + bridges

4. **TSTL → Lua link** — `typescript-to-lua` adds "Skip pure Lua (use lua-guide / lua-opinionated-guide)" and links to `lua-guide` for Lua idioms; keeps only transpiler-specific topics (namespaces-vs-classes, multi-return, decorators, stable-tables, avoiding-heavy-features, tsconfig, interop).
5. **Paradigm bridges** — `typescript`/`python`/`lua`/`c99` add progressive-disclosure pointers: "functional style → general-fp-guide; class design → general-oop-guide." (Reverse links; `general-fp`/`general-oop` already point down.)
6. **Language↔framework Skip reciprocity** — `typescript` description routes framework work out (expressjs/hono/react/astro → their guides); `expressjs`↔`hono` already cross-skip — verify and complete.

### Low priority — polish / generalization

7. **`content` generalization** — references and entry points are domain-coupled (news/travel/port/ship → CruiseReviews); rename to domain-neutral and state reusability, or move tenant-specific guides to a tenant skill. (Confirm intent — content may be deliberately product-scoped.)
8. **`lua-opinionated` scope wording** — clarify it's performance tuning that _especially_ benefits LuaJIT, with principles applying to vanilla Lua 5.4 too.
9. **Document intentional overlaps** — note in the audit reference that parallel `merge`/`create`/`simplify` ops across `instruction`/`prompt`/`skill` are _correct_ (one concept per domain), not duplication to collapse.

### Explicitly NOT changing (already correct)

- React/visual cluster (`react`/`motion-react`/`remotion`/`threejs`) — boundaries already explicit and cross-linked.
- Meta/tooling parallel ops (`insights`/`instruction`/`prompt`/`plan`/`skill`/`llmstxt`) — parallel-by-domain, not shared.
- Infra (`cmake`/`moon`/`docker`/`kubernetes`/`terraform`/`shell-scripting`) — independent single-tool skills; no forced "devops meta" skill.
- The already-restructured systems cluster (the reference model).

---

## Per-refactor execution pattern

For each move (mirrors the c99/DOD/memory work already done):

1. **Pick the owner** skill for the concept.
2. **Move/keep** the canonical reference in the owner; **delete** the duplicate from the other skill (`git rm`).
3. In the de-duplicated skill, **replace** the removed bullet/reference with a by-name cross-reference ("see hono-guide") in the SKILL.md bullet + the relevant `references/` Related lines; update the `description` Skip clause to route the concept to the owner.
4. **Fix links:** repoint any Related/`references/` links that pointed at a removed file; confirm no dangling links.
5. **Sources:** move any provenance to the owner's `SOURCES.md`; ensure no author/company names leak into prose.
6. **Validate:** JSON valid; every `SKILL.md`→`references/` link resolves; `npx prettier --write` the package (leave `marketplace.json` compact); spot-check `eval-queries.json` still routes (positives for the skill, negatives to the now-cross-linked sibling).
7. **Register** any new skill in `marketplace.json` (none expected in Part B; all moves are within existing skills).

## Verification & rollout

- One cluster per commit (`refactor(skill-hono): tier split…`, etc.) so each is reviewable and revertible.
- After each: run prettier on the touched packages, validate JSON, and run a link-integrity scan (`SKILL.md` + Related → existing files). Re-scan for inline source attributions.
- Optional: use each skill's `eval-queries.json` as the routing check; add a negative case pointing at the new cross-linked owner.

## Out of scope / deferred

- New general skills for "HTTP API contract design" or "build/devops orchestration" — only extract if recurring cross-tool work appears; not justified now.
- Renaming skills to Anthropic's verb-ing+noun convention — the repo's `<topic>-guide` convention is established; not worth a catalog-wide rename.

## Sources

- [Equipping agents for the real world with Agent Skills (Anthropic)](https://www.anthropic.com/engineering/equipping-agents-for-the-real-world-with-agent-skills)
- [The Complete Guide to Building Skills for Claude (Anthropic)](https://resources.anthropic.com/hubfs/The-Complete-Guide-to-Building-Skill-for-Claude.pdf)
- [Agent Skills: Progressive Disclosure as a System Design Pattern (SwirlAI)](https://www.newsletter.swirlai.com/p/agent-skills-progressive-disclosure)
- [Progressive Discovery: A Better Mental Model for Agent Skills (DEV)](https://dev.to/phil-whittaker/progressive-discovery-a-better-mental-model-for-agent-skills-51bd)
- [Claude Agent Skills: A First Principles Deep Dive](https://leehanchung.github.io/blogs/2025/10/26/claude-skills-deep-dive/)
- [Introduction to agent skills (Anthropic Courses)](https://anthropic.skilljar.com/introduction-to-agent-skills)
