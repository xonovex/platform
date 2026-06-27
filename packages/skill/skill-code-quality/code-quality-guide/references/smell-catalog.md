# Smell Catalog — What to Audit For, and Who Owns Each

The audit checklist: every code smell, the signal that detects it, and the **one** skill that owns its definition and fix. This file routes; it does not redefine. Organized by design-problem **family** (the peer-reviewed grouping) — _not_ by an application/class/method scope tier, which is an informal, weakly-sourced label, not a citable taxonomy.

Owner key: `robustness.md` = this skill's robustness dimension · **oop-guide** = OO-design smells · **connascence-guide** = coupling/cohesion smells · _here_ = this skill owns the detector (duplication / dead code / over-abstraction cleanup).

## Bloaters — grown too large

| Smell                    | Detector signal                                                 | Owner                                                             |
| ------------------------ | --------------------------------------------------------------- | ----------------------------------------------------------------- |
| Long Method              | function over ~30 lines; high cyclomatic / cognitive complexity | `robustness.md` (code smells)                                     |
| Large Class / God Object | many fields/methods, low cohesion, `Manager`/`Util` grab-bag    | **oop-guide** (SRP)                                               |
| Primitive Obsession      | a domain concept carried as a raw string/number; magic literals | `robustness.md` / **connascence-guide**                           |
| Long Parameter List      | more than ~4 params, especially boolean flag params             | `robustness.md` + **connascence-guide** (connascence of position) |
| Data Clumps              | the same group of values always travels together                | **connascence-guide**                                             |

## Object-Orientation Abusers

| Smell                                     | Detector signal                                                 | Owner                                                   |
| ----------------------------------------- | --------------------------------------------------------------- | ------------------------------------------------------- |
| Switch on a type code                     | the same type-switch duplicated in several places               | **oop-guide** (polymorphism)                            |
| Refused Bequest                           | subclass ignores or throws on inherited members (LSP violation) | **oop-guide** (LSP)                                     |
| Downcasting                               | a cast that breaks the abstraction model                        | **oop-guide** (LSP) — _supplementary; cited in SOURCES_ |
| Temporary Field                           | a field set/used only in some circumstances, empty otherwise    | **oop-guide**                                           |
| Alternative Classes, Different Interfaces | two classes do the same job with unswappable APIs               | **oop-guide**                                           |

## Change Preventers

| Smell                            | Detector signal                                   | Owner                                          |
| -------------------------------- | ------------------------------------------------- | ---------------------------------------------- |
| Divergent Change                 | one class edited for many unrelated reasons       | **oop-guide** (SRP)                            |
| Shotgun Surgery                  | one conceptual change smeared across many modules | **connascence-guide** (locality at a distance) |
| Parallel Inheritance Hierarchies | every new subclass forces a mirrored subclass     | **oop-guide**                                  |

## Dispensables — remove for free (this skill owns the detector)

| Smell                    | Detector signal                                    | Owner                                               |
| ------------------------ | -------------------------------------------------- | --------------------------------------------------- |
| Duplicated Code          | identical / near-identical logic in 2+ places      | _here_ (DUPLICATES)                                 |
| Dead Code                | unreachable or never-called code, exports, params  | _here_ (DEAD CODE)                                  |
| Speculative Generality   | abstraction / hook with no current user            | _here_ (OVER-ENGINEERING)                           |
| Lazy Class / Data Class  | a class too thin to justify itself, or only fields | **oop-guide**                                       |
| Magic Numbers / Literals | unnamed constants begging for a name               | `robustness.md` — _supplementary; cited in SOURCES_ |

## Couplers

| Smell                  | Detector signal                                                | Owner                                  |
| ---------------------- | -------------------------------------------------------------- | -------------------------------------- |
| Feature Envy           | a method more interested in another object's data than its own | **connascence-guide**                  |
| Inappropriate Intimacy | two classes reach into each other's internals                  | **connascence-guide**                  |
| Message Chains         | `a.b().c().d()` train-wreck navigation                         | **connascence-guide** (Law of Demeter) |
| Middle Man             | a class that mostly just forwards to one collaborator          | **connascence-guide**                  |

## Grade and report

- Grade each finding by severity — blast radius × likelihood × remediation effort — group by family or bucket, and report read-only (no edits). See [SKILL.md](../SKILL.md) for the method and [robustness.md](robustness.md) for the robustness signals.
- The membership and family grouping above follow the peer-reviewed smell taxonomy (cited in SOURCES); the "application / class / method" scope labels are an editorial convenience only.
