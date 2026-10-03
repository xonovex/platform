# Platform Plugin

Configure Moon tasks and build and operate Docker images, Kubernetes workloads, and Terraform infrastructure.

<!-- xonovex:installation:start -->

## Install

Add the Xonovex marketplace once, then install `xonovex-platform` in the selected harness. The bundle version is `5.3.0`.

### Claude Code

Install this bundle in Claude Code. The harness discovers its bundled skills and commands.

```bash
claude plugin marketplace add xonovex/platform
claude plugin install xonovex-platform@xonovex-marketplace
```

### Codex

Install this bundle in Codex. Codex loads the skills declared in its plugin manifest; Claude Code slash commands are separate components.

```bash
codex plugin marketplace add xonovex/platform
codex plugin add xonovex-platform@xonovex-marketplace
```

<!-- xonovex:installation:end -->

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
