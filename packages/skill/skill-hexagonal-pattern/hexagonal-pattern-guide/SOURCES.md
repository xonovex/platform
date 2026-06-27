# Sources

## Hexagonal architecture (ports and adapters)

- **Title:** "Hexagonal Architecture" (Alistair Cockburn)
- **URL:** https://alistair.cockburn.us/hexagonal-architecture/
- **Last reviewed:** 2026-06-27
- **Used for:**
  - `SKILL.md` → Essentials, Gotchas
- **Aspects extracted:**
  - Ports owned by the core; driving (primary) vs driven (secondary) adapters; the hexagon metaphor → `references/ports-and-adapters.md`

## Clean / Onion architecture and the dependency rule

- **Title:** "Clean Architecture" / "The Clean Architecture" (Robert C. Martin); Onion Architecture (Jeffrey Palermo)
- **URL:** https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html
- **Last reviewed:** 2026-06-27
- **Used for:**
  - `SKILL.md` → Essentials (dependency rule), composition root
- **Aspects extracted:**
  - Dependencies point inward toward stable abstractions; dependency inversion → `references/dependency-inversion.md`
  - The composition root as the one place that names concretes → `references/composition-root.md`

## Component principles (stability)

- **Title:** Component coupling principles (Robert C. Martin)
- **URL:** https://en.wikipedia.org/wiki/Package_principles
- **Last reviewed:** 2026-06-27
- **Used for:**
  - `references/dependency-inversion.md`
- **Aspects extracted:**
  - Stable-dependencies / stable-abstractions, instability, Zone of Pain → `references/dependency-inversion.md`

## Refresh Workflow

1. Re-read the upstream source(s) above
2. Diff against the prior pull
3. Update the corresponding `references/<topic>.md`
4. Bump **Last reviewed**
