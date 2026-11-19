---
name: researching-changes
description: "Use when researching a codebase and external sources to ground a future implementation plan, without writing the plan or changing code. Triggers on research before a plan exists, mapping the code paths a change touches, checking current library versions or API behaviour before a plan, comparing options with evidence, listing constraints and open questions for a change, and a read-only report for a later plan, even when the user doesn't say 'research' and only asks what a change involves."
---

# Research a Change for a Future Plan

Produce a read-only, planning-oriented report from an inline subject or an opaque provider-native reference plus an optional revision. Let a selected provider resolve native inputs and return the report inline.

## Inputs

- **Requirements** (required): a description of what to research, inline or as a provider-native reference with an optional native revision
- **Interactive** (optional): ask clarifying questions before the research starts
- **Save to** (optional): a file in which the caller wants the report; the report is also returned inline

## Core Workflow

1. **Resolve the subject**: use the supplied inline request or ask the selected provider to resolve the opaque reference and optional native revision. Clarify only material ambiguity.
2. **Inspect the codebase**: find architecture, integration points, similar implementations, dependency versions, validation tasks, and applicable implementation skills. Parallelize independent searches where available.
3. **Research external facts**: when current versions, APIs, standards, or recommendations matter, consult authoritative sources and record provenance and freshness.
4. **Separate evidence from synthesis**: distinguish observed facts, inferred implications, constraints, uncertainty, and unresolved questions.
5. **Return one report**: include current stack, relevant locations, options or recommendation, risks, skills to consult, and sources. Use a separate Publish operation if the result must be persisted.

If the request is a general hardening, simplification, alignment, duplication, comment, barrel, or TODO audit rather than research for an explicit future plan, hand it to **auditing-code-quality**. Planning research may cite an existing quality-audit result as evidence without redefining its detectors.

## Boundaries

Writing the plan from the report belongs to **writing-plan**. Settling an open question that the report lists belongs to **settling-questions**.

## Gotchas

- Research is not plan authoring or implementation; stop after the report.
- Read project-level task runners as well as language manifests before naming validation commands.
- Do not treat conversation memory as a durable source when the caller supplied a provider reference.
- An explicitly selected unavailable provider or source fails visibly instead of falling back to another storage model.

## Progressive Disclosure

- Read [references/example.md](references/example.md) - Load when shaping the research report or checking what a complete report holds
