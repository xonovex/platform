---
name: terraform-guide
description: "Use when editing Terraform 1.12+ infrastructure code. Triggers on `.tf` files, `*.tfvars`, modules, and prompts about module design, environment isolation, remote state, for_each, dynamic blocks, locals, or providers, even when the user doesn't say 'Terraform'. Skip Pulumi, OpenTofu-only features, CloudFormation, and ad-hoc cloud CLI work."
---

# Terraform Coding Guidelines

## Requirements

- Terraform ≥ 1.12; remote state; pinned providers; fmt/validate.

## Essentials

- **Module design** - One responsibility per module, clear inputs/outputs, typed variables, see [references/module-structure.md](references/module-structure.md), [references/module-definition.md](references/module-definition.md), [references/typed-variables.md](references/typed-variables.md)
- **Composition** - Thin root modules, compose child modules, use locals for computed values, see [references/root-module.md](references/root-module.md), [references/locals.md](references/locals.md)
- **Environment isolation** - Separate envs (dirs + tfvars), isolate state per env, see [references/root-module.md](references/root-module.md), [references/remote-state.md](references/remote-state.md)
- **Advanced patterns** - Dynamic blocks, for_each, count, see [references/advanced-patterns.md](references/advanced-patterns.md)

## Gotchas

- State file (`terraform.tfstate`) contains secrets in plaintext — store remotely with encryption (S3 + KMS) and lock with DynamoDB / `azurerm_storage_account` lease
- `count` and `for_each` differ on identity: `count`-indexed resources are positional (removing one shifts all subsequent), `for_each` is keyed and stable
- `lifecycle { prevent_destroy = true }` blocks `terraform destroy` for that resource — useful for prod, painful in CI/test envs
- `terraform refresh` updates state from real infra but doesn't show what changed — use `terraform plan -refresh-only` to preview drift
- Provider version pinning lives in `required_providers`, not the resource block — unpinned providers break on minor upgrades

## Progressive disclosure

- Read [references/module-structure.md](references/module-structure.md) - When designing reusable Terraform modules
- Read [references/module-definition.md](references/module-definition.md) - When defining module inputs, outputs, or resources
- Read [references/typed-variables.md](references/typed-variables.md) - When adding variable validation or complex types
- Read [references/locals.md](references/locals.md) - When computing intermediate values or reducing duplication
- Read [references/root-module.md](references/root-module.md) - When organizing environment-specific configurations
- Read [references/remote-state.md](references/remote-state.md) - When configuring state backend or state isolation
- Read [references/advanced-patterns.md](references/advanced-patterns.md) - When using dynamic blocks, for_each, or count
