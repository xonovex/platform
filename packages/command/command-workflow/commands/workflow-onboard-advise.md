---
description: Recommend profile-compatible lifecycle methods, skills, providers, executors, policy evidence, and environment modules without applying changes
allowed-tools:
  - Bash
  - Read
  - Glob
  - Grep
  - Skill
argument-hint: "[scope] [--profile <reference>] [--provider <selection>] [--module <name>]"
---

# /xonovex-workflow:workflow-onboard-advise — Advise Workflow and Governance Onboarding

## Arguments

- `scope` (optional): Environment, repository, organization, harness, CI, provider, or workflow scope; defaults to current environment.
- `--profile` (optional): Desired workflow/governance profile reference; onboard-advise recommends a compatible shipped reference profile from the per-plane `assets/profiles/` libraries plus the methods, providers, executors, and modules to satisfy it.
- `--provider` (optional): Candidate result provider to assess.
- `--module` (optional): Candidate skill, adapter, policy, harness, CI, or environment module.

## Delegation

Load `plan-guide` and perform **lifecycle-onboard-advise**, `workflow-guide` for profile,
provider, and result semantics, and `agent-governance-guide` for module trust, permissions,
data flows, enforcement guarantees, verification, rollback, and drift. Recommend compatible
methods, skills, providers, executors, and modules; stop after discovery, assessment,
recommendation, and exact preview. Never authorize or apply configuration.
