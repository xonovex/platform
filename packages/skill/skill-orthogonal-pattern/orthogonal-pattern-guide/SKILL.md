---
name: orthogonal-pattern-guide
description: "Use when deciding how to DECOMPOSE a system into modules/packages along orthogonal axes (variation points), and where each concern's boundary goes. Triggers on factoring concerns into independent axes and variants, shared-core plus per-variant leaves, package-by-feature vs package-by-layer, mirroring directory structure to the domain, naming parallel modules/siblings symmetrically, keeping two consumers (e.g. a CLI export command and a scheduled export service) symmetric, cross-cutting concerns (logging/policy/telemetry that aren't axes), aligning boundaries to domain seams, false modularity, or a tree that feels like spaghetti — even when the user doesn't say 'orthogonal'."
---

# Orthogonal Axes: Modular, Pluggable Decomposition

Decompose a system into **variation points** — orthogonal **axes**, each a design decision likely to change — then realize each as interchangeable **variants** (leaves) behind a port, wired by a registry: the **microkernel (plug-in) architecture**. The directory tree and the code are two views of one model. Factoring into axes serves separation of concerns — but not every concern is an axis; some cut across all of them.

## Essentials

- **Factor into variation points (axes)** — each is a design decision likely to change; derive them from what varies, not from processing steps; a configuration is one variant per axis, a point in their cross-product, see [references/finding-axes.md](references/finding-axes.md)
- **Commonality vs variability** — at every node split the shared core (commonality) from per-variant leaves (variability); "pick one variant per axis" is Strategy, see [references/commonality-variability.md](references/commonality-variability.md)
- **Structure mirrors the model** — abstract axes are ports, concrete variants are adapters, see [references/structure-isomorphism.md](references/structure-isomorphism.md)
- **Sibling names, sibling grammar** — parallel concepts get parallel names in the same part of speech, see [references/naming-symmetry.md](references/naming-symmetry.md)
- **Localize cross-axis coupling** — put the inevitable glue at a variation-point bridge inside the variant that needs it, never hoisted into a new top-level sibling, see [references/variation-point-bridges.md](references/variation-point-bridges.md)

## Code

- **Realize variants as plug-ins behind a port** — the axis interface is a port and each variant is an adapter; for the port substrate see **hexagonal-pattern-guide**, for the open registry and extension machinery see **microkernel-pattern-guide**
- **Modular code, graded coupling** — narrow ports, neutral data across seams, explicit state, see [references/modular-code.md](references/modular-code.md); grade each seam with **connascence-guide**

## Layout

- **Fan out, share at each node** — `shared-core → shared-lib → {consumer A, consumer B}`, each consumer adding only its realization; keep two consumers symmetric where they share function, see [references/applying-the-layout.md](references/applying-the-layout.md)

## Gotchas

- Orthogonal axes are rarely _perfectly_ independent — coupling is real, so localize it at a bridge; the goal is not "no coupling" but "no leaked coupling."
- A concern that touches every variant is not an axis — logging, the fail-closed encryption-policy gate, and telemetry are cross-cutting concerns; weave them once at the composition root, never as a sibling leaf (forcing them in causes scattering or tangling), see [references/cross-cutting-concerns.md](references/cross-cutting-concerns.md).
- Over-factoring is the opposite failure — wait for a real second variant or a genuine shared core; a premature shared core defended with flags and per-variant branches (the wrong abstraction) is costlier to unwind than duplication, so tolerate duplication until the rule of three.
- A fan-out tree proves nothing if leaves still import siblings or share mutable state — that's false modularity (a distributed monolith); enforce boundaries with import/architecture tests, not directory layout.
- A heavily-depended-on shared module that holds concretes sits in the Zone of Pain — stable yet rigid; keep the most-depended-on node abstract (a port).
- Realizing the variants — ports, registry, capability gating, binding time — and grading their seams are their own concerns; see **hexagonal-pattern-guide**, **microkernel-pattern-guide**, and **connascence-guide**. This skill is about finding and placing the axes.

## Progressive Disclosure

- Read [references/finding-axes.md](references/finding-axes.md) - Load when identifying variation points, deciding whether two concerns are one axis or two, or when NOT to add an axis
- Read [references/commonality-variability.md](references/commonality-variability.md) - Load when splitting shared-core from per-variant leaves (CCP/CRP/REP, deep-vs-shallow)
- Read [references/modular-code.md](references/modular-code.md) - Load when shaping the code itself — narrow ports, neutral data contracts, false-modularity enforcement
- Read [references/structure-isomorphism.md](references/structure-isomorphism.md) - Load when mapping the model onto a tree (ports/adapters, instability/Zone-of-Pain, feature models)
- Read [references/naming-symmetry.md](references/naming-symmetry.md) - Load when naming sibling variants/axes or making two consumers structurally symmetric
- Read [references/variation-point-bridges.md](references/variation-point-bridges.md) - Load when a variant needs another axis's specifics, or distinguishing a cross-tree constraint from a cross-cutting concern
- Read [references/cross-cutting-concerns.md](references/cross-cutting-concerns.md) - Load when a concern (logging, policy, telemetry) touches every variant and resists being modeled as an axis
- Read [references/applying-the-layout.md](references/applying-the-layout.md) - Load when refactoring a "spaghetti" layout into axes, reversing a bad split, or laying out a fresh one
- Read [references/boundary-alignment.md](references/boundary-alignment.md) - Load when choosing where axis boundaries go (domain seams, Conway's law, enforcing boundaries in code)
