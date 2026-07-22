---
name: orthogonal-pattern-guide
description: "Use when deciding how to decompose a system into modules or packages along independent variation axes, and where each concern's boundary belongs. Triggers on shared-core plus per-variant leaves, package-by-feature vs package-by-layer, sibling naming symmetry, parallel consumers that should mirror the same concepts, cross-axis constraints, cross-cutting concerns that are not axes, false modularity, or a module tree that feels tangled — even when the user doesn't say 'orthogonal'."
---

# Orthogonal Axes

Decompose around independent decisions likely to change. An axis names the question; each variant is one answer. A configuration chooses one variant per axis.

## Essentials

- **Find independent decisions** — derive axes from what can vary separately, not execution steps, see [references/finding-axes.md](references/finding-axes.md)
- **Separate commonality from variability** — keep only truly shared rules at the axis root and put each answer in a leaf, see [references/commonality-variability.md](references/commonality-variability.md)
- **Make the tree reveal the model** — use one directory per axis and bare, symmetric variant names, see [references/applying-the-layout.md](references/applying-the-layout.md), [references/naming-symmetry.md](references/naming-symmetry.md)
- **Localize cross-axis constraints** — place exceptional two-axis glue in the dependent variant, see [references/variation-point-bridges.md](references/variation-point-bridges.md)
- **Distinguish cross-cutting concerns** — apply concerns that touch every selection once around the composed operation, see [references/cross-cutting-concerns.md](references/cross-cutting-concerns.md)
- **Align ownership and change** — a change to one concept should stay inside one axis and one owner, see [references/boundary-alignment.md](references/boundary-alignment.md)

## Handoffs

- Use **hexagonal-pattern-guide** to design the ports and adapters that enforce a chosen axis.
- Use **microkernel-pattern-guide** to implement registration, discovery, capability gating, and binding time.
- Use **connascence-guide** to grade and loosen coupling across a proposed seam.
- Use **ddd-guide** to discover bounded contexts and domain vocabulary before drawing domain boundaries.

## Gotchas

- Two choices that always change together are one axis; splitting them creates permanent synchronization work.
- A concern present in every configuration is cross-cutting, not another selectable variant.
- Do not extract an axis for one hypothetical implementation. Wait for independent change pressure or a real second variant.
- A tidy tree is false modularity when leaves still import siblings or share hidden mutable state; enforce the dependency rules owned by **hexagonal-pattern-guide** and assess violations with **connascence-guide**.

## Progressive Disclosure

- Read [references/finding-axes.md](references/finding-axes.md) - Load when identifying variation points or deciding whether two concerns are one axis or two
- Read [references/commonality-variability.md](references/commonality-variability.md) - Load when splitting an axis root from its per-variant leaves
- Read [references/applying-the-layout.md](references/applying-the-layout.md) - Load when mapping axes onto a new tree or incrementally reorganizing a tangled one
- Read [references/naming-symmetry.md](references/naming-symmetry.md) - Load when naming axes, sibling variants, or parallel consumer structures
- Read [references/variation-point-bridges.md](references/variation-point-bridges.md) - Load when one variant has a genuine constraint involving another axis
- Read [references/cross-cutting-concerns.md](references/cross-cutting-concerns.md) - Load when logging, policy, telemetry, or another concern touches every variant
- Read [references/boundary-alignment.md](references/boundary-alignment.md) - Load when checking whether an axis follows one change concept and one owner
