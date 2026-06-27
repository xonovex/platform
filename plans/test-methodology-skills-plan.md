# Plan / Prompt — TDD · BDD-ATDD · DDD · FDD · Testing · User-Stories skills (and distilling `plan-tdd-create`)

> **How to use this file.** Paste it (or point) into a **fresh session with ultracode on**. It is a self-contained brief: goal, the precedent to copy, the catalog conventions, the proposed skill set + ownership, all sources (some embedded verbatim), the `plan-tdd-create` distillation, and a validation checklist. Execute it with **Workflow-orchestrated** research → author → wire → validate, exactly as the `code-quality` skill was built.

---

## 1. Goal

Create **six** new general-tier guideline skills and **distill `plan-tdd-create`** into `plan-create` + the new `tdd` skill (so the `plan-tdd-create` command and its `plan-guide` reference are removed), mirroring how `plan-research-code-{align,harden,simplify}` were folded into `plan-research` + the new `code-quality` skill.

New skills (recommended names; see §8 for naming/scope decisions):

1. **`tdd`** — test-first workflow (red → green → refactor); classical vs mockist.
2. **`bdd-atdd`** — behaviour/acceptance from concrete examples; Gherkin Given-When-Then; three amigos; example mapping; "test the team's *understanding*, not the software."
3. **`ddd`** — ubiquitous language, bounded context, aggregates, strategic/tactical design.
4. **`fdd`** — Feature-Driven Development: feature lists, the five processes, client-valued features (confirm scope — possibly a thin skill).
5. **`testing`** — how to write good tests (AAA / Four-Phase, FIRST, naming, what to mock) **and** the canonical **test-double taxonomy** (Dummy / Stub / Spy / Mock / Fake) reconciling the "same name for different things" problem.
6. **`user-stories`** — writing / optimising / splitting stories (INVEST, 3Cs, vertical slicing, SPIDR, acceptance criteria).

---

## 2. The precedent to copy exactly (the harden/simplify / `code-quality` build)

This was just done in this repo; **replicate the pattern**:

- **Two-axis insight.** A command like `plan-tdd-create` fuses a *procedure* (owned by `plan-guide`: research → plan-document lifecycle) with *criteria/method* (the TDD discipline). Separate them: the **procedure stays in `plan-guide`**; the **method becomes a skill** (`tdd`) that `plan-create` composes with when the request is test-first.
- **Lean skill that routes, one cited owner per concept.** `code-quality` ended up = `SKILL.md` (audit method) + `robustness.md` (its one owned dimension) + `smell-catalog.md` (a cited routing table that points each smell to its **one owner**: OO-design → `oop-guide`, coupling → `connascence-guide`). Do the same here — don't duplicate a concept across skills; the owner defines it, everyone else cross-references **by name**.
- **Cite authoritatively, never folklore.** The smell work proved this: Wikipedia's "application/class/method-level" grouping was `[citation needed]` folklore; the real taxonomy was Mäntylä & Lassenius. **Verify every claim against the authoritative source** (book/paper/canonical site); put citations **only in `SOURCES.md`** (never author/book names in `SKILL.md` or `references/*`).
- **Project-agnostic.** Per `skill-skill/skill-guide/references/guideline-skills.md` → "Neutral Examples for General / Pattern Skills": use neutral example domains (orders, payments, a taxi dispatcher, a shopping cart), never this repo's code.
- **Directional, acyclic, declared dependencies.** Hard deps point **upward** (general tier never depends on a specific one), declared as a **bare plugin-name string in BOTH** `.claude-plugin/plugin.json` and `.codex-plugin/plugin.json`. No cycles.
- **Register + reconcile.** Add each skill to **both** marketplaces — `.claude-plugin/marketplace.json` (compact one-line, alphabetical) **and** `.agents/plugins/marketplace.json` (object form: `name` + `source{source:"local",path}` + `policy{installation:"AVAILABLE",authentication:"ON_INSTALL"}` + `category:"Productivity"`, alphabetical). Version `3.2.0` (lockstep; do **not** bump everything — the release PR does that). Run `npm install`, then **delete any stale `"extraneous": true` workspace entries** it leaves in `package-lock.json`. `npx prettier --write` each package.

---

## 3. Catalog conventions (must follow)

Read `packages/skill/AGENTS.md` and the `skill-skill` guide (`packages/skill/skill-skill/skill-guide/SKILL.md` + `references/{composability,guideline-skills,writing-descriptions}.md`) first. Key rules:

- **Package layout:** `skill-<topic>/` → `package.json` (`@xonovex/skill-<topic>`), `moon.yml`, `prettier.config.ts`, `.claude-plugin/plugin.json` + `.codex-plugin/plugin.json` (`xonovex-skill-<topic>`), and `<topic>-guide/` containing `SKILL.md`, `references/*.md`, `SOURCES.md`, `eval-queries.json`. Frontmatter `name:` **must equal** the guide dir name. Codex `skills` is a string path; Claude `skills` is an array.
- **Composability:** one concept → one owner skill; cross-reference others **by bold name** (`**testing-guide**`), never by cross-package file link; general tier stays self-contained. Soft dep = describe a capability, declare nothing; hard dep = name the exact skill in `dependencies` (both manifests).
- **Descriptions are the router** — `Use when … Triggers on … Skip … (see other-guide)`; ≤1024 chars; re-check after edits.
- **SKILL.md** = 3-7 essentials + Gotchas + one example + Progressive Disclosure (each reference with a load-when trigger). Reference files: statement / rationale / how-to / example / counter-example; no frontmatter.
- **eval-queries.json**: ~8-16 queries, mix should_trigger true/false with near-miss negatives pointing at the right owner.

---

## 4. Proposed skill set, ownership & relationships (recommended — refine with research)

| Skill | Owns (one-owner) | Hard deps (upward) | Cross-refs (by name) | Reconcile with existing |
|---|---|---|---|---|
| `tdd` | red-green-refactor, test-first discipline, classical vs mockist, "let tests drive design" | `testing` | `bdd-atdd` (acceptance-first), `connascence`/`oop` (the design the refactor step improves) | — |
| `bdd-atdd` | behaviour from examples, Gherkin Given-When-Then, three amigos, example mapping, specification-by-example, "test understanding not software" | `user-stories` | `ddd` (ubiquitous language), `testing` (step impls, stub external services) | — |
| `ddd` | ubiquitous language, bounded context, aggregates/entities/value-objects, strategic vs tactical | — (foundational) | — | **`orthogonal-pattern`/`boundary-alignment.md` already cites bounded context/DDD** → make `ddd` the owner and have boundary-alignment + `data-model-guide` cross-reference it; remove any duplicated definition |
| `fdd` | feature lists, the 5 FDD processes, "feature = `<action>` the `<result>` of a(n) `<object>`", client-valued function | `ddd` (step 1 builds the domain object model) | `user-stories` (features ≈ stories) | Confirm it deserves a full skill vs a short one (it's a niche process) |
| `testing` | AAA / Four-Phase test, FIRST, test naming, what-to-mock, **test-double taxonomy** (Dummy/Stub/Spy/Mock/Fake), state vs behaviour verification, test smells, Test Data Builder / Object Mother | — (foundational) | `connascence` (a test that needs many doubles = coupling smell) | Owner of test-double terminology; `code-quality` may later cross-ref it for "untested/over-mocked" signals |
| `user-stories` | INVEST, 3Cs (Card/Conversation/Confirmation), `As a … I want … so that …`, vertical slicing, story splitting (SPIDR), acceptance criteria, backlog refinement | — | `bdd-atdd` (express acceptance criteria as Given-When-Then examples) | — |

**Dependency direction (keep ACYCLIC):** `tdd → testing`; `bdd-atdd → {user-stories, testing, ddd}`; `fdd → {ddd, user-stories}`; `ddd`, `testing`, `user-stories` are leaves. **Decision to settle (see §8):** the `user-stories ↔ bdd-atdd` edge — recommend `bdd-atdd → user-stories` (examples flesh out a story), with `user-stories` only *soft*-mentioning Given-When-Then to avoid a cycle.

---

## 5. Distill `plan-tdd-create` (mirror the `plan-research-code-*` folding)

- **Delete** `packages/command/command-workflow/commands/plan-tdd-create.md`.
- **Delete** `packages/skill/skill-plan/plan-guide/references/plan-tdd-create.md` (its TDD-specific structure moves into the `tdd` skill).
- **`plan-create.md`** (plan-guide reference): add a short note — "For a test-first plan, apply **`tdd-guide`**'s red-green-refactor (or **`bdd-atdd-guide`** for acceptance-first): structure steps as failing-test → implement → refactor, and list the test doubles per **`testing-guide`**." (This is the same move as `plan-research.md` gaining its "Code-quality audits" section.)
- **`plan-guide/SKILL.md`**: remove `plan-tdd-create` from the **Plan Lifecycle**, **Plan Operations**, and **Progressive Disclosure** lists; note that `plan-create` covers TDD via `tdd-guide`. (Search `plan-guide` for every `plan-tdd-create` mention — e.g. other references say "`plan-create` (or `plan-tdd-create`)"; update them.)
- **`plan-guide` plugin.json (both)**: add hard deps `xonovex-skill-tdd` (and `xonovex-skill-bdd-atdd` if `plan-create` routes to it).
- **`command-workflow`**: remove the `plan-tdd-create` row from `README.md`'s command table; update the workflow ASCII + `packages/diagram/diagram-agent-workflow/workflow-diagram.dot` (it shows `plan-create\n(or plan-tdd-create)` → drop the `(or plan-tdd-create)` or note "TDD via tdd-guide") and **regenerate the PNG** (`dot -Tpng workflow-diagram.dot -o workflow-diagram.png`; graphviz is available). `command-workflow` already deps `xonovex-skill-plan`; no command-dep change needed unless `plan-create` now also needs `tdd` present — if so add `xonovex-skill-tdd` to `command-workflow` deps too.
- After: `grep -rn 'plan-tdd-create' packages/` must return **zero**.

---

## 6. Sources (research these online; verify, do not copy uncritically)

Run an ultracode **research workflow** (parallel web-research agents, then a synthesis that resolves ownership/overlap), exactly like the smell-taxonomy research. Per topic, the canonical sources:

- **Test doubles & test writing:**
  - Martin Fowler, "Mocks Aren't Stubs" — https://martinfowler.com/articles/mocksArentStubs.html (state vs behaviour verification; classical vs mockist TDD).
  - Gerard Meszaros, *xUnit Test Patterns* — http://xunitpatterns.com/ (the **canonical** taxonomy + patterns: Four-Phase Test, Test Data Builder, Object Mother; test smells).
  - "Mocks, Fakes, Stubs and Dummies" — http://xunitpatterns.com/Mocks,%20Fakes,%20Stubs%20and%20Dummies.html (**reconcile the "same name for different things"** — Meszaros vs Fowler vs common usage).
  - Also: Kent Beck "TDD by Example" (red-green-refactor); FIRST principles; AAA / Arrange-Act-Assert.
- **TDD:** Kent Beck, *Test-Driven Development by Example*; the red-green-refactor cycle; "test list".
- **BDD / ATDD / Gherkin:** Dan North "Introducing BDD"; Cucumber/Gherkin docs (Given-When-Then, scenarios, feature files); Matt Wynne "Example Mapping"; Gojko Adzic *Specification by Example*; the **three amigos**. **Embed the Cucumber-creator framing in Appendix A** — BDD is about testing *shared understanding of unwritten software*, discovery workshops, thumbs-up/down on whether a story is small enough, Gherkin to make examples concrete, drive **domain** logic (stub external services), avoid UI tests.
- **DDD:** Eric Evans, *Domain-Driven Design*; Vaughn Vernon, *Implementing DDD*; ubiquitous language, bounded context, context mapping, aggregates/entities/value objects, anti-corruption layer.
- **FDD:** Coad, Lefebvre & De Luca, *Java Modeling in Color with UML* / Stephen Palmer & Mac Felsing, *A Practical Guide to Feature-Driven Development*; the 5 processes (develop overall model → build feature list → plan by feature → design by feature → build by feature); feature template `<action> the <result> of a(n) <object>`.
- **User stories:** Mike Cohn, *User Stories Applied* + SPIDR splitting (https://www.mountaingoatsoftware.com); Bill Wake, **INVEST** (https://xp123.com/articles/invest-in-good-stories-and-smart-tasks/); Ron Jeffries, the **3Cs** (Card, Conversation, Confirmation); Humanizing Work "story splitting flowchart"; example mapping; vertical-slice / "walking skeleton". **Embed the user-story best-practices in Appendix B.**

The two appendices below are **primary framing material the requester supplied** — preserve their intent in `bdd-atdd` and `user-stories`.

---

## 7. Execution workflow (ultracode)

1. **Research** (Workflow): clusters = {tdd+testing+test-doubles}, {bdd-atdd+gherkin}, {ddd}, {fdd}, {user-stories}, + a **synthesis** that returns: per-skill scope, the one-owner ownership map, cross-ref/dep directions (acyclic), per-skill cited sources, and the overlap reconciliations (DDD↔boundary-alignment/data-model; testing↔code-quality; bdd-atdd↔user-stories). Mirror the smell-taxonomy workflow shape.
2. **Scaffold** the 6 packages (boilerplate; copy an existing skill package's files and rename).
3. **Author** each skill (Workflow fan-out, one agent per reference): SKILL.md (routing description), references, SOURCES.md (cited), eval-queries.json — **project-agnostic**, criteria-only where it routes, no author names in body.
4. **Wire** deps in both manifests (upward, acyclic) + cross-references by name; **reconcile** `boundary-alignment.md`/`data-model-guide` to point at `ddd`.
5. **Distill** `plan-tdd-create` per §5.
6. **Register** 6 skills in both marketplaces (alphabetical positions), `npm install`, clear extraneous lockfile entries, `prettier --write`.
7. **Validate** (§9). One cluster/commit if committing.

---

## 8. Open decisions (with recommendations)

- **`bdd-atdd` one skill vs separate `bdd` + `atdd`** → **one** `bdd-atdd` (ATDD is the acceptance-test framing of the same example-driven idea); confirm via research.
- **`fdd` full vs thin** → likely **thin** (niche process); confirm whether it earns a full skill or a short one that leans on `ddd` + `user-stories`.
- **DDD ownership** → `ddd` owns ubiquitous-language / bounded-context / aggregates; `orthogonal-pattern/boundary-alignment.md` and `data-model-guide` **cross-reference** it (remove duplicated definitions). Verify what each currently says.
- **`user-stories ↔ bdd-atdd` direction** → `bdd-atdd → user-stories` (acyclic); `user-stories` soft-mentions Given-When-Then.
- **Naming** → `tdd` / `bdd-atdd` / `ddd` / `fdd` / `testing` / `user-stories`. Alternatives to weigh: `test-doubles` or `unit-testing` vs `testing`; `acceptance-testing` vs `bdd-atdd`; `story-writing` vs `user-stories`.
- **`testing` vs `code-quality` overlap** → `testing` owns how-to-write-tests + doubles; `code-quality` may later add a "missing/over-mocked tests" detector signal that routes to `testing`. Keep one owner.

---

## 9. Validation checklist (copy from the `code-quality` build)

- `grep -rn 'plan-tdd-create' packages/` → zero.
- Every `**x-guide**` cross-reference names a **real** skill (no dangling).
- No author/company/book names in any `SKILL.md`/`references/*` (only in `SOURCES.md`); principle/pattern names (Given-When-Then, INVEST, AAA) are fine.
- All within-skill links resolve; descriptions ≤1024 chars; frontmatter `name` == guide dir.
- Dependency graph **acyclic**, every dep resolves to a marketplace plugin, declared in **both** manifests.
- Both marketplaces valid JSON, all 6 present, `.claude-plugin` fully alphabetical.
- `package-lock.json` valid, registers the 6 packages, **no `"extraneous"` entries**, `npm install` idempotent.
- `prettier --check` clean for every changed package; PNG regenerated.

---

## Appendix A — BDD framing (the Cucumber creator, HN 10194242) — preserve in `bdd-atdd`

> Cucumber is **not a tool for testing software. It is a tool for testing people's understanding of how software (yet to be written) should behave.** Most bugs/delays from rework arise from misunderstandings. Workflow: pop a story, run a ~20-min **Discovery Workshop** with business + IT, talk through **concrete examples** in plain language ("the one where the picture is too big"; "the one with five taxis in range"). Thumbs-up/down: if enough thumbs-down, the story goes back / gets split. Then a dev (with a tester) makes 2-5 examples concrete in **Gherkin (Given-When-Then)**, e.g.:
>
>     Scenario: Close taxis with higher rating win
>       Given taxi A with rating 0.8 is 1400m from the customer
>       And taxi B with 0.9 is 1500m from the customer
>       When the customer requests a taxi
>       Then taxi B should be assigned
>
> Business confirms it. Then the dev does **regular TDD** using the scenario to drive **core domain logic** — external services / queues / DBs are **stubbed out**, no UI/Selenium. Far **more unit tests** than Cucumber scenarios. **Cucumber = write the right code; unit tests = write the code right.** Don't drive Gherkin through the UI (slow, volatile, doesn't localise the bug). Output: executable **living documentation** accessible to the whole team that prevents defects by surfacing bad assumptions up-front.

## Appendix B — User-story best practices — preserve in `user-stories`

> A story is a **placeholder for a conversation about user value**, not an exhaustive spec.
> 1. **INVEST** — Independent, Negotiable, Valuable, Estimable, Small, Testable.
> 2. **"What" & "Why", not "How"** — e.g. ✗ "Add a DB column for cart items" vs ✓ "As an online shopper, I want to save items to my cart so that I can purchase them together later."
> 3. **Clear acceptance criteria** — happy path, boundaries, errors, UI changes; BDD Given/When/Then for automation.
> 4. **Slice vertically** — every story delivers a thin, usable slice (not "build DB", "build API", "build UI" in sequence).
> 5. **Continuously refine the backlog** — PM + Dev + QA in refinement; understand the problem before writing the story.
> Standard template: `As a [user], I want to [action] so that [value]`. Keep small enough to finish in a sprint.
