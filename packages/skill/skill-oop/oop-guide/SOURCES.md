# Sources

## Bad smells in code

- **Title:** "Refactoring: Improving the Design of Existing Code" (Martin Fowler & Kent Beck), Ch. 3 "Bad Smells in Code"
- **URL:** https://martinfowler.com/books/refactoring.html
- **Last reviewed:** 2026-06-27
- **Used for:**
  - `SKILL.md` → Design smells
- **Aspects extracted:**
  - God Object / Large Class, Refused Bequest, Divergent Change, Parallel Inheritance Hierarchies, Alternative Classes with Different Interfaces, Temporary Field

## Liskov Substitution Principle

- **Title:** Liskov & Wing — "A Behavioral Notion of Subtyping" (ACM TOPLAS, 1994); Barbara Liskov, OOPSLA 1987 keynote
- **URL:** https://dl.acm.org/doi/10.1145/197320.197383
- **Last reviewed:** 2026-06-27
- **Used for:**
  - `SKILL.md` → Design smells (Refused Bequest, Downcasting), Gotchas (Liskov)
- **Aspects extracted:**
  - Behavioral subtyping; substitutability as the contract refused bequest and downcasting violate

## SOLID principles

- **URL:** https://en.wikipedia.org/wiki/SOLID
- **Last reviewed:** 2026-06-27
- **Used for:**
  - `SKILL.md` → Best practices, Design smells
- **Aspects extracted:**
  - Single responsibility (god object, divergent change), open/closed, Liskov, interface segregation, dependency inversion

## Refresh Workflow

1. Re-read the upstream source(s) above
2. Diff against the prior pull
3. Update the corresponding section
4. Bump **Last reviewed**
