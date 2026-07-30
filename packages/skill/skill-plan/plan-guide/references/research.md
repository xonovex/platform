# Research: Codebase and External Evidence for a Plan

Produce a read-only, planning-oriented report from an inline subject or an opaque provider-native reference plus an optional revision. Let a selected provider resolve native inputs and return the report inline.

## Core Workflow

1. **Resolve the subject**: use the supplied inline request or ask the selected provider to resolve the opaque reference and optional native revision. Clarify only material ambiguity.
2. **Inspect the codebase**: find architecture, integration points, similar implementations, dependency versions, validation tasks, and applicable implementation skills. Parallelize independent searches where available.
3. **Research external facts**: when current versions, APIs, standards, or recommendations matter, consult authoritative sources and record provenance and freshness.
4. **Separate evidence from synthesis**: distinguish observed facts, inferred implications, constraints, uncertainty, and unresolved questions.
5. **Return one report**: include current stack, relevant locations, options or recommendation, risks, skills to consult, and sources. Use a separate Publish operation if the result must be persisted.

If the request is a general hardening, simplification, alignment, duplication, comment,
barrel, or TODO audit rather than research for an explicit future plan, hand it to
**code-quality-guide**. Planning research may cite an existing quality-audit result as
evidence without redefining its detectors.

## Gotchas

- Research is not plan authoring or implementation; stop after the report.
- Read project-level task runners as well as language manifests before naming validation commands.
- Do not treat conversation memory as a durable source when the caller supplied a provider reference.
- An explicitly selected unavailable provider or source fails visibly instead of falling back to another storage model.
