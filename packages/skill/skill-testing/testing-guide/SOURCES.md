# Sources

## Test-double taxonomy, Four-Phase Test, and test smells

- **Title:** Gerard Meszaros — "xUnit Test Patterns: Refactoring Test Code" (Addison-Wesley, 2007) / xunitpatterns.com
- **URLs:**
  - http://xunitpatterns.com/Mocks,%20Fakes,%20Stubs%20and%20Dummies.html
  - https://www.informit.com/articles/article.aspx?p=1398624
  - http://xunitpatterns.com/Four%20Phase%20Test.html
  - http://xunitpatterns.com/Obscure%20Test.html
- **Last reviewed:** 2026-06-27
- **Used for:**
  - `references/test-double-taxonomy.md`
  - `references/state-vs-behaviour-and-what-to-mock.md`
  - `references/test-structure-and-first.md`
  - `references/test-smells-and-fixtures.md`
- **Aspects extracted:**
  - The Test Double umbrella term and the five kinds (Dummy / Stub / Spy / Mock / Fake) and the continuum caveat
  - Responder vs Saboteur stubs; indirect input (Stub) vs indirect output (Spy / Mock); Fake as an alternative implementation
  - The Four-Phase Test (Setup / Exercise / Verify / Teardown), its relationship to Arrange-Act-Assert, and teardown usually being unnecessary for unit tests
  - Test smells — Obscure Test, Mystery Guest, Eager Test, Fragile Test (Sensitive Equality, Indirect Testing), Erratic Test — and how smells cause other smells

## Definitions of the doubles and state vs behaviour verification

- **Title:** Martin Fowler — "Mocks Aren't Stubs"
- **URL:** https://martinfowler.com/articles/mocksArentStubs.html
- **Last reviewed:** 2026-06-27
- **Used for:**
  - `references/test-double-taxonomy.md`
  - `references/state-vs-behaviour-and-what-to-mock.md`
- **Aspects extracted:**
  - Concise definitions of Dummy / Stub / Spy / Mock / Fake
  - State vs behaviour verification; "only mocks insist on behaviour verification"; a stub answers queries while a mock verifies commands; the loose-usage problem with the word "mock"

## Arrange-Act-Assert

- **Title:** Bill Wake — "3A: Arrange, Act, Assert" (xp123)
- **URL:** https://xp123.com/articles/3a-arrange-act-assert/
- **Last reviewed:** 2026-06-27
- **Used for:**
  - `references/test-structure-and-first.md`
- **Aspects extracted:**
  - Origin and naming of Arrange-Act-Assert; the single-action-per-test framing

## The FIRST qualities

- **Title:** Tim Ottinger & Brett Schuchert — "FIRST" (in Robert C. Martin (ed.), "Clean Code", ch. 9, p. 132)
- **URL:** https://medium.com/pragmatic-programmers/unit-tests-are-first-fast-isolated-repeatable-self-verifying-and-timely-a83e8070698e
- **Last reviewed:** 2026-06-27
- **Used for:**
  - `references/test-structure-and-first.md`
- **Aspects extracted:**
  - The five FIRST qualities (Fast, Independent, Repeatable, Self-validating, Timely); the Independent-vs-Isolated "I" and Self-validating-vs-Self-verifying "S" ambiguity

## Fixture-creation patterns

- **Title:** Nat Pryce — "Test Data Builders: an alternative to the Object Mother pattern" (2007)
- **URL:** http://www.natpryce.com/articles/000714.html
- **Last reviewed:** 2026-06-27
- **Used for:**
  - `references/test-smells-and-fixtures.md`
- **Aspects extracted:**
  - The Test Data Builder pattern and why it beats the Object Mother (every variation needs a new method, violating single responsibility); fluent specify-only-what-matters fixtures

## Refresh Workflow

1. Re-read the upstream source(s) above
2. Diff against the prior pull captured in the reference files
3. Update the corresponding `references/<topic>.md`
4. Bump **Last reviewed** date above
