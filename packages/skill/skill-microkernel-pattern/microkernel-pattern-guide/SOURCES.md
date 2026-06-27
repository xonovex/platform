# Sources

## Microkernel / plug-in architecture pattern

- **Title:** "Software Architecture Patterns" — microkernel architecture (Mark Richards, O'Reilly)
- **URL:** https://www.oreilly.com/library/view/software-architecture-patterns/9781098134280/ch04.html
- **Last reviewed:** 2026-06-27
- **Used for:**
  - `SKILL.md` → Essentials, Gotchas
- **Aspects extracted:**
  - Minimal core + plug-in components + registry (name, data contract, connection); plug-ins isolated via standard contracts; adapter for non-conforming plug-ins → `references/core-plugins-registry.md`, `references/wiring.md`

## Secure-by-default

- **Title:** "The Protection of Information in Computer Systems" (Saltzer & Schroeder, 1975); capability-based security
- **URLs:**
  - https://www.cs.virginia.edu/~evans/cs551/saltzer/
  - https://en.wikipedia.org/wiki/Capability-based_security
- **Last reviewed:** 2026-06-27
- **Used for:**
  - `references/capabilities-fail-closed.md`
- **Aspects extracted:**
  - Fail-safe defaults, principle of least authority, complete mediation, ambient authority

## Binding time (software product lines)

- **Title:** Feature-binding analysis / SPLE (Lee & Kang; Pohl, Böckle & van der Linden)
- **URL:** https://link.springer.com/book/10.1007/3-540-28901-1
- **Last reviewed:** 2026-06-27
- **Used for:**
  - `references/binding-time.md`
- **Aspects extracted:**
  - Compile/load/run-time binding spectrum; binding unit/technique/time; compositional vs annotative variability

## Wiring anti-patterns

- **Title:** "Service Locator is an Anti-Pattern" (Mark Seemann)
- **URL:** https://blog.ploeh.dk/2010/02/03/ServiceLocatorisanAnti-Pattern/
- **Last reviewed:** 2026-06-27
- **Used for:**
  - `references/wiring.md`
- **Aspects extracted:**
  - Dependency injection vs service locator; resolve at the composition root, inject the dependency

## Metapatterns lens (one-author synthesis, non-canonical)

- **Title:** "Architectural Metapatterns" / "Introduction to Software Architecture With Actors" (Denys Poltorak)
- **URLs:**
  - https://github.com/denyspoltorak/metapatterns
  - https://metapatterns.io/
- **Last reviewed:** 2026-06-27
- **Used for:**
  - `references/metapatterns-lens.md`
- **Aspects extracted:**
  - The "microkernel regarded as multiple ports-and-adapters cores sharing a common middleware", "middleware-gateway", and "distributed microkernel relates to mesh" framings — recorded explicitly as one author's non-canonical synthesis, with caveats (OS-vs-pattern conflation; gateway is a sub-type of middleware; a service mesh is a networking data plane and does not virtualize system resources). Single-author, self-published CC catalogue, not standard CS doctrine.

## Refresh Workflow

1. Re-read the upstream source(s) above
2. Diff against the prior pull
3. Update the corresponding `references/<topic>.md`
4. Bump **Last reviewed**
