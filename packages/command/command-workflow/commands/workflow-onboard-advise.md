---
description: Recommend a minimal workflow composition without installing or enabling anything
allowed-tools:
  - Bash
  - Read
  - Glob
  - Grep
  - Skill
argument-hint: "[scope] [--trigger <kind>] [--executor <plugin>] [--host <kind>] [--control <plugin:mode>...] [--evidence <plugin:failure>...]"
---

# /xonovex-workflow:workflow-onboard-advise — Advise a Composition

## Arguments

- `scope` (optional): Repository, project, command, or external workflow to assess.
- `--trigger`, `--executor`, `--host` (optional): Independent execution selections.
- `--control` (repeatable): Optional control plus `observe` or `enforce` mode.
- `--evidence` (repeatable): Optional sink plus `ignore` or `fail` behavior.

## Delegation

Discover only facts needed to describe the requested composition, report every unselected
dimension as unselected, and recommend the smallest set of adapters and plugins that
satisfies explicit requirements. Treat maturity as an optional separate assessment. Do
not install, mutate, activate, or imply controls.
