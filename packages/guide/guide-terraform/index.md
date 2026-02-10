---
name: terraform-guidelines
description: Trigger on `.tf` files, Terraform modules, `*.tfvars`. Use when writing Terraform 1.12+ for infrastructure as code. Apply for module design, environment isolation, state management. Keywords: Terraform, modules, variables, outputs, remote state, for_each, dynamic blocks, locals, environment isolation.
---

# Terraform Coding Guidelines

## Requirements

- Terraform ≥ 1.12; remote state; pinned providers; fmt/validate.

## Essentials

- **Module design** - One responsibility per module, clear inputs/outputs, typed variables, see [reference/module-structure.md](reference/module-structure.md), [reference/module-definition.md](reference/module-definition.md), [reference/typed-variables.md](reference/typed-variables.md)
- **Composition** - Thin root modules, compose child modules, use locals for computed values, see [reference/root-module.md](reference/root-module.md), [reference/locals.md](reference/locals.md)
- **Environment isolation** - Separate envs (dirs + tfvars), isolate state per env, see [reference/root-module.md](reference/root-module.md), [reference/remote-state.md](reference/remote-state.md)
- **Advanced patterns** - Dynamic blocks, for_each, count, see [reference/advanced-patterns.md](reference/advanced-patterns.md)

## Progressive disclosure

- Read [reference/module-structure.md](reference/module-structure.md) - When designing reusable Terraform modules
- Read [reference/module-definition.md](reference/module-definition.md) - When defining module inputs, outputs, or resources
- Read [reference/typed-variables.md](reference/typed-variables.md) - When adding variable validation or complex types
- Read [reference/locals.md](reference/locals.md) - When computing intermediate values or reducing duplication
- Read [reference/root-module.md](reference/root-module.md) - When organizing environment-specific configurations
- Read [reference/remote-state.md](reference/remote-state.md) - When configuring state backend or state isolation
- Read [reference/advanced-patterns.md](reference/advanced-patterns.md) - When using dynamic blocks, for_each, or count
