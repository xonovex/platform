# Sources

## The Pragmatic Programmer — orthogonality

- **URL:** https://pragprog.com/titles/tpp20/the-pragmatic-programmer-20th-anniversary-edition/
- **Last reviewed:** 2026-06-27
- **Used for:**
  - `SKILL.md` → Essentials (factor into orthogonal axes)
- **Aspects extracted:**
  - Orthogonality, decoupling, the cross-product design space → `references/finding-axes.md`

## SOLID / Clean Architecture — single responsibility as an axis of change, screaming architecture

- **URL:** https://blog.cleancoder.com/uncle-bob/2011/09/30/Screaming-Architecture.html
- **Last reviewed:** 2026-06-27
- **Used for:**
  - `SKILL.md` → Essentials (structure mirrors the model), Gotchas (package by layer hides axes)
- **Aspects extracted:**
  - The "axis of change" reading of SRP → `references/finding-axes.md`
  - Package-by-axis/feature vs package-by-layer (screaming architecture) → `references/structure-isomorphism.md`

## Software product lines — commonality / variability analysis

- **URL:** https://www.sei.cmu.edu/our-work/software-product-lines/
- **Last reviewed:** 2026-06-27
- **Used for:**
  - `SKILL.md` → Essentials (shared-core + per-variant leaves)
- **Aspects extracted:**
  - Commonality vs variability, variation points, the shared kernel → `references/commonality-variability.md`, `references/variation-point-bridges.md`

## Design Patterns — Strategy and Bridge

- **URL:** https://en.wikipedia.org/wiki/Bridge_pattern
- **Last reviewed:** 2026-06-27
- **Used for:**
  - `SKILL.md` → Essentials (Strategy per axis, Bridge for cross-axis glue)
- **Aspects extracted:**
  - One interchangeable family per axis (Strategy) → `references/commonality-variability.md`
  - Decoupling an abstraction from its implementation so both vary independently (Bridge) → `references/variation-point-bridges.md`

## Information hiding & module decomposition

- **Title:** "On the Criteria To Be Used in Decomposing Systems into Modules" (D.L. Parnas, 1972)
- **URL:** https://dl.acm.org/doi/10.1145/361598.361623
- **Last reviewed:** 2026-06-27
- **Used for:**
  - `SKILL.md` → Essentials, Gotchas
- **Aspects extracted:**
  - Decompose around design decisions likely to change; information hiding; temporal decomposition → `references/finding-axes.md`, `references/variation-point-bridges.md`

## Module depth & complexity

- **Title:** "A Philosophy of Software Design" (John Ousterhout)
- **URL:** https://web.stanford.edu/~ouster/cgi-bin/aposd.php
- **Last reviewed:** 2026-06-27
- **Used for:**
  - `SKILL.md` → Gotchas (over-factoring)
- **Aspects extracted:**
  - Deep vs shallow modules (benefit = functionality, cost = interface); information leakage → `references/commonality-variability.md`, `references/finding-axes.md`

## Component (package) principles

- **Title:** "Clean Architecture" / component principles (Robert C. Martin)
- **URL:** https://en.wikipedia.org/wiki/Package_principles
- **Last reviewed:** 2026-06-27
- **Used for:**
  - `SKILL.md` → Gotchas (Zone of Pain)
- **Aspects extracted:**
  - REP/CCP/CRP component cohesion; ADP/SDP/SAP component coupling; instability metric and Main Sequence (Zone of Pain) → `references/commonality-variability.md`, `references/structure-isomorphism.md`, `references/variation-point-bridges.md`

## Software product lines & variability

- **Title:** "Software Product Line Engineering" + Orthogonal Variability Model (Pohl, Böckle & van der Linden); "Feature-Oriented Domain Analysis (FODA)" (Kang et al.)
- **URLs:**
  - https://link.springer.com/book/10.1007/3-540-28901-1
  - https://resources.sei.cmu.edu/library/asset-view.cfm?assetid=11231
- **Last reviewed:** 2026-06-27
- **Used for:**
  - `SKILL.md` → framing, Essentials
- **Aspects extracted:**
  - Commonality/variability analysis, variation point, variant → `references/finding-axes.md`, `references/commonality-variability.md`
  - Feature models, cross-tree constraints (requires/excludes), orthogonal variability model → `references/structure-isomorphism.md`, `references/variation-point-bridges.md`

## Separation of concerns & cross-cutting concerns

- **Title:** "On the role of scientific thought" (EWD447, Dijkstra); "Aspect-Oriented Programming" (Kiczales et al., 1997)
- **URLs:**
  - https://www.cs.utexas.edu/users/EWD/transcriptions/EWD04xx/EWD447.html
  - https://en.wikipedia.org/wiki/Aspect-oriented_programming
- **Last reviewed:** 2026-06-27
- **Used for:**
  - `SKILL.md` → framing, Gotchas
- **Aspects extracted:**
  - Separation of concerns; cross-cutting concern, tangling, scattering, weaving; homogeneous vs heterogeneous → `references/cross-cutting-concerns.md`

## Abstraction discipline

- **Title:** "The Wrong Abstraction" (Sandi Metz); "Yagni" (Martin Fowler); speculative generality (refactoring smell)
- **URLs:**
  - https://sandimetz.com/blog/2016/1/20/the-wrong-abstraction
  - https://martinfowler.com/bliki/Yagni.html
  - https://refactoring.guru/smells/speculative-generality
- **Last reviewed:** 2026-06-27
- **Used for:**
  - `SKILL.md` → Gotchas (over-factoring)
- **Aspects extracted:**
  - Duplication is cheaper than the wrong abstraction; rule of three; YAGNI; speculative generality → `references/finding-axes.md`, `references/commonality-variability.md`

## Boundary forces

- **Title:** "How Do Committees Invent?" (Conway's Law, Melvin Conway); "Domain-Driven Design" bounded context & ubiquitous language (Eric Evans)
- **URLs:**
  - https://www.melconway.com/Home/Committees_Paper.html
  - https://martinfowler.com/bliki/BoundedContext.html
- **Last reviewed:** 2026-06-27
- **Used for:**
  - `references/boundary-alignment.md`
- **Aspects extracted:**
  - Structure mirrors team communication; single owner per axis; bounded-context / domain-seam alignment; fitness functions / architecture tests → `references/structure-isomorphism.md`, `references/applying-the-layout.md`

## Refresh Workflow

1. Re-read the upstream source(s) above
2. Diff against the prior pull (or scan for newly added sections)
3. For each changed area, update the corresponding `references/<topic>.md`
4. Bump **Last reviewed** date above
