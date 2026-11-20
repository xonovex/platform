# Sources

## Connascence

- **Title:** "What Every Programmer Should Know About Object-Oriented Design" (Meilir Page-Jones, 1995); "Comparing Techniques by Means of Encapsulation and Connascence" (1992)
- **URL:** https://en.wikipedia.org/wiki/Connascence
- **Last reviewed:** 2026-06-27
- **Used for:**
  - `SKILL.md` → Essentials, Gotchas
- **Aspects extracted:**
  - Static/dynamic forms; strength/locality/degree; rules of degree and locality; connascence of position → `references/connascence.md`

## Coupling & cohesion taxonomy

- **Title:** "Structured Design" (Stevens, Myers & Constantine, IBM Systems Journal, 1974)
- **URLs:**
  - https://en.wikipedia.org/wiki/Coupling_(computer_programming)
  - https://en.wikipedia.org/wiki/Cohesion_(computer_science)
- **Last reviewed:** 2026-06-27
- **Used for:**
  - `SKILL.md` → Essentials
- **Aspects extracted:**
  - Content/common/control/stamp/data coupling ladder → `references/coupling-ladder.md`
  - Coincidental→functional cohesion ladder; temporal cohesion → `references/cohesion-ladder.md`

## Law of Demeter

- **URL:** https://en.wikipedia.org/wiki/Law_of_Demeter
- **Last reviewed:** 2026-06-27
- **Used for:**
  - `references/law-of-demeter.md`
- **Aspects extracted:**
  - Principle of least knowledge; permitted call targets; train-wreck detection

## Bad smells in code

- **Title:** "Refactoring: Improving the Design of Existing Code" (Martin Fowler & Kent Beck), Ch. 3 "Bad Smells in Code"
- **URL:** https://martinfowler.com/books/refactoring.html
- **Last reviewed:** 2026-06-27
- **Used for:**
  - `references/connascence.md` → Named smells that are connascence
- **Aspects extracted:**
  - Data Clump, Feature Envy, Shotgun Surgery as named coupling smells mapped onto connascence forms

## Refresh Workflow

1. Re-read the upstream source(s) above
2. Diff against the prior pull
3. Update the corresponding `references/<topic>.md`
4. Bump **Last reviewed**
