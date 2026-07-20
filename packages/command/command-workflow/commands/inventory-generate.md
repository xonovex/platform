---
description: Deterministically generate an exact-revision SBOM, AI/ML/cryptographic/service, or agent-environment inventory
allowed-tools:
  - Read
  - Write
  - Bash
  - Glob
  - Grep
  - AskUserQuestion
  - Skill
argument-hint: "<subject-reference> [--revision <native-revision>] --kind <sbom|ai-sbom|ml-bom|cbom|service|agent-environment> [--format <selection>] [--provider <selection>] [--enrich-descriptions]"
---

# /xonovex-workflow:inventory-generate — Generate an Inventory

## Arguments

- `subject-reference` (required): Opaque native subject reference.
- `--revision` (optional): Exact native subject revision; required for revision-sensitive publication.
- `--kind` (required): Inventory specialization.
- `--format` (optional): Pinned interoperable specification/schema or provider-native representation.
- `--provider` (optional): Result/BOM provider selection.
- `--enrich-descriptions` (optional): Allow bounded non-authoritative prose enrichment only.

## Delegation

Perform **inventory-generate** with deterministic ecosystem generators and the selected
provider skill. Never use model output to invent or modify component identities,
versions, digests, relationships, provenance, or completeness.
