# Sources

## Test-first cycle, the two rules, the test list, green-bar strategies

- **Title:** Kent Beck — "Test-Driven Development by Example" (Addison-Wesley, 2002)
- **URL:** https://www2.cs.uh.edu/~rsingh/documents/software_design/TDD.pdf
- **Last reviewed:** 2026-06-27
- **Used for:**
  - `references/red-green-refactor.md`
  - `references/green-bar-strategies.md`
  - `references/test-list-and-design.md`
  - `SKILL.md` → Essentials / Gotchas / Example
- **Aspects extracted:**
  - The red-green-refactor mantra and the two rules (write production code only to pass a failing test; eliminate duplication); the "clean code that works" goal
  - Green-bar strategies — Fake It, Obvious Implementation, Triangulation — and refactoring driven by removing the duplication created to go green
  - The test list / to-do list worked one item at a time

## Triangulation as a fallback, not the default

- **Title:** "Notes on Test-Driven Development by Example" (study notes on Beck)
- **URL:** https://stanislaw.github.io/2016-01-25-notes-on-test-driven-development-by-example-by-kent-beck.html
- **Last reviewed:** 2026-06-27
- **Used for:**
  - `references/green-bar-strategies.md`
  - `SKILL.md` → Gotchas
- **Aspects extracted:**
  - Triangulation is reserved for when the generalization is not yet visible; otherwise write the general solution directly

## Classical vs mockist TDD as a design style

- **Title:** Martin Fowler — "Mocks Aren't Stubs"
- **URL:** https://martinfowler.com/articles/mocksArentStubs.html
- **Last reviewed:** 2026-06-27
- **Used for:**
  - `references/test-list-and-design.md` (design-pressure section)
- **Aspects extracted:**
  - Classical uses real collaborators where practical; mockist mocks any collaborator with interesting behaviour
  - Mockist tests couple to implementation (expected calls asserted) while classical tests focus on the result; the style choice drives different designs
  - The verification mechanism (state vs behaviour, what to mock, the double taxonomy) is owned by **testing-guide**, which cites this same source for that angle

## Arrange-Act-Assert (the shape of the test the cycle produces)

- **Title:** Bill Wake — "3A: Arrange, Act, Assert" (xp123)
- **URL:** https://xp123.com/articles/3a-arrange-act-assert/
- **Last reviewed:** 2026-06-27
- **Used for:**
  - `references/test-list-and-design.md`
- **Aspects extracted:**
  - Origin and naming of the AAA structure of the single-action test each red step produces; the AAA structure itself is owned by **testing-guide**

## Refresh Workflow

1. Re-read the upstream source(s) above
2. Diff against the prior pull (mantra wording, the two rules, the three green-bar strategies, Triangulation's status, the classical/mockist coupling tradeoff)
3. Update the corresponding `references/<topic>.md`
4. Bump **Last reviewed** date above
