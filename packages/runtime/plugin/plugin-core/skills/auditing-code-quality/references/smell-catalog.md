# Smell Catalog: What to Audit For, and Who Owns Each

Every code smell, the signal that detects it, and the **one** skill that owns its definition and fix. Routes; does not redefine. Grouped by design-problem **family** (the peer-reviewed grouping), not by an application/class/method scope tier.

Owner key: `robustness.md` = this skill's robustness dimension · **applying-oop** = OO-design smells · **analyzing-connascence** = coupling/cohesion smells · _here_ = this skill owns the detector (duplication / dead code / over-abstraction / redundant-comment cleanup).

## Bloaters: grown too large

| Smell | Detector signal | Owner |
| --- | --- | --- |
| Long Method | function over ~30 lines; high cyclomatic / cognitive complexity | `robustness.md` (code smells) |
| Primitive Obsession | a domain concept carried as a raw string/number; magic literals | `robustness.md` / **analyzing-connascence** |
| Long Parameter List | more than ~4 params, especially boolean flag params | `robustness.md` + **analyzing-connascence** (connascence of position) |
| Data Clumps | the same group of values always travels together | **analyzing-connascence** |

## Object-Orientation Abusers

| Smell | Detector signal | Owner |
| --- | --- | --- |
| Switch on a type code | the same type-switch duplicated in several places | **applying-oop** (polymorphism) |
| Refused Bequest | subclass ignores or throws on inherited members (LSP violation) | **applying-oop** (LSP) |
| Downcasting | a cast that breaks the abstraction model | **applying-oop** (LSP): _supplementary; cited in SOURCES_ |
| Temporary Field | a field set/used only in some circumstances, empty otherwise | **applying-oop** |
| Alternative Classes, Different Interfaces | two classes do the same job with unswappable APIs | **applying-oop** |

## Change Preventers

| Smell | Detector signal | Owner |
| --- | --- | --- |
| Divergent Change | one class edited for many unrelated reasons | **applying-oop** (SRP) |
| Shotgun Surgery | one conceptual change smeared across many modules | **analyzing-connascence** (locality at a distance) |
| Parallel Inheritance Hierarchies | every new subclass forces a mirrored subclass | **applying-oop** |

## Dispensables: remove for free (this skill owns the detector)

| Smell | Detector signal | Owner |
| --- | --- | --- |
| Speculative Generality | abstraction / hook with no current user | _here_ (OVER-ENGINEERING) |
| Comments (noise) | a comment restating a well-named declaration, narrating a plan / provenance, or denser than the surrounding code | _here_ (REDUNDANT COMMENT): rename to self-document, or delete |
| Lazy Class / Data Class | a class too thin to justify itself, or only fields | **applying-oop** |
| Magic Numbers / Literals | unnamed constants begging for a name | `robustness.md`: _supplementary; cited in SOURCES_ |

## Couplers

| Smell | Detector signal | Owner |
| --- | --- | --- |
| Feature Envy | a method more interested in another object's data than its own | **analyzing-connascence** |
| Inappropriate Intimacy | two classes reach into each other's internals | **analyzing-connascence** |
| Message Chains | `a.b().c().d()` train-wreck navigation | **analyzing-connascence** (Law of Demeter) |
| Middle Man | a class that mostly just forwards to one collaborator | **analyzing-connascence** |

## Grade and report

Grade each finding by severity: blast radius × likelihood × remediation effort: group by family, and report read-only. See [SKILL.md](../SKILL.md) for the method and [robustness.md](robustness.md) for the robustness signals.
