---
name: defining-subplans
description: "Use when expanding a parent plan into focused child plans with their execution order. Triggers on subplans, breaking a plan into independently completable parts of five to seven tasks, parallel groups and sequencing from dependencies and overlapping files, a split by phase, and the size of a subplan, even when the user doesn't say 'subplan' and asks to break the plan down into workable pieces."
---

# Define Subplans from a Parent Plan

Expand any explicit parent plan, inline or provider-native, into focused child plans. An approval or status field is not a precondition. Stop after returning the child plans and execution ordering.

## Inputs

- **Parent plan** (required): the parent plan, inline or as a provider-native reference with an optional native revision, such as `plans/auth.md`
- **By phase** (optional): split at the phase markers of the plan in place of a logical grouping
- **Preview** (optional): show the child plans without writing a file

## Core Workflow

1. Resolve the explicit parent plan, optional native revision, active canonical context, supporting references, and selected source-provider capabilities.
2. Extract the objective, components, constraints, risks, dependencies, success criteria, applicable context, and `skills_to_consult` capabilities.
3. Create focused child plans with roughly five to seven tasks, affected resources or packages, applicable decisions, validation, success criteria, dependencies, and any descriptive status metadata the provider needs.
4. Derive execution groups from explicit dependencies and overlapping mutable files, packages, environments, or provider resources. Non-overlapping children may run concurrently; conflicting children are sequenced.
5. Return inline child plans and ordering, including the parent, child, and dependency relationships a later Publish operation should preserve.
6. Stop before implementation.

## Boundaries

The parent plan comes from **writing-plan**. One child at a time is implemented through **implementing-plan**, and a supervised run over all children belongs to **delegating-plan**.

## Gotchas

- Do not reject an explicit parent because it lacks approval metadata or uses an unfamiliar status.
- A parent status may describe progress but never grants permission for expansion.
- Keep children independently completable; overlapping ownership without sequencing creates merge risk.
- Source-provider relationships are native resources, not a universal wrapper or required file layout.
- Subplans with more than seven tasks risk silent drops: target five to seven focused tasks

## Progressive Disclosure

- Read [references/example.md](references/example.md) - Load when laying out a child plan and its execution group
