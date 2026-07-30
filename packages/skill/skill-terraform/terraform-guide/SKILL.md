---
name: terraform-guide
description: "Use when editing Terraform 1.12+ infrastructure code. Triggers on `.tf` files, `*.tfvars`, modules, and prompts about module design, environment isolation, remote state, for_each, dynamic blocks, locals, or providers, even when the user doesn't say 'Terraform'."
---

# Terraform Coding Guidelines

## Requirements

- Terraform ≥ 1.12; remote state; pinned providers; fmt/validate.

## Essentials

- **Module design** - One responsibility per module, clear inputs/outputs, typed variables, see [references/module-structure.md](references/module-structure.md), [references/module-definition.md](references/module-definition.md), [references/typed-variables.md](references/typed-variables.md)
- **Composition** - Thin root modules, compose child modules, use locals for computed values, see [references/root-module.md](references/root-module.md), [references/locals.md](references/locals.md)
- **Environment isolation** - Separate envs (dirs + tfvars), isolate state per env, see [references/root-module.md](references/root-module.md), [references/remote-state.md](references/remote-state.md)
- **Advanced patterns** - Data sources, dynamic blocks, see [references/advanced-patterns.md](references/advanced-patterns.md)

## Gotchas

- State file (`terraform.tfstate`) contains secrets in plaintext: store it remotely with encryption and backend-native locking (S3 `use_lockfile = true`; Azure Blob lease locking)
- `lifecycle { prevent_destroy = true }` blocks `terraform destroy` for that resource: useful for prod, painful in CI/test envs
- `terraform refresh` updates state from real infra but doesn't show what changed: use `terraform plan -refresh-only` to preview drift
- Provider version pinning lives in `required_providers`, not the resource block: unpinned providers break on minor upgrades

## Progressive disclosure

- Read [references/module-structure.md](references/module-structure.md) - Load when designing reusable Terraform modules
- Read [references/module-definition.md](references/module-definition.md) - Load when defining module inputs, outputs, or resources
- Read [references/typed-variables.md](references/typed-variables.md) - Load when adding variable validation or complex types
- Read [references/locals.md](references/locals.md) - Load when computing intermediate values or reducing duplication
- Read [references/root-module.md](references/root-module.md) - Load when organizing environment-specific configurations
- Read [references/remote-state.md](references/remote-state.md) - Load when configuring remote state, encryption, locking, or per-environment state keys
- Read [references/advanced-patterns.md](references/advanced-patterns.md) - Load when using data sources or dynamic blocks
