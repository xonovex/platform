# Platform Plugin

Configure Moon tasks and build and operate Docker images, Kubernetes workloads, and Terraform infrastructure.

## Install

Add the [Xonovex marketplace](../../../../README.md#agent-plugins) once, then install this plugin in the selected harness.

```bash
# Claude Code
claude plugin install xonovex-platform@xonovex-marketplace

# Codex
codex plugin add xonovex-platform@xonovex-marketplace
```

## Skills

Choose the skill that owns the task. The harness loads its instructions when the request matches its routing description.

| Skill | Use when |
| --- | --- |
| [docker-guide](skills/docker-guide/SKILL.md) | Writing or editing Docker images and Compose files for production. |
| [kubernetes-guide](skills/kubernetes-guide/SKILL.md) | Editing Kubernetes manifests in GitOps repos. |
| [moon-guide](skills/moon-guide/SKILL.md) | Configuring moonrepo monorepo tasks. |
| [terraform-guide](skills/terraform-guide/SKILL.md) | Editing Terraform 1.12+ infrastructure code. |

## Validation

Run `npx moon run plugin-platform:ci-check --force` from the repository root to check the plugin manifests, formatting, and owned skills.
