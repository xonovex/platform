# Platform Plugin

Configure Moon tasks and build and operate Docker images, Kubernetes workloads, and Terraform infrastructure.

<!-- xonovex:installation:start -->

## Install

Add the Xonovex marketplace once, then install `xonovex-platform` in the selected harness. The bundle version is `6.0.0`.

### Claude Code

Install this bundle in Claude Code. The harness discovers its bundled skills.

```bash
claude plugin marketplace add xonovex/platform
claude plugin install xonovex-platform@xonovex-marketplace
```

### Codex

Install this bundle in Codex. Codex loads the skills declared in its plugin manifest.

```bash
codex plugin marketplace add xonovex/platform
codex plugin add xonovex-platform@xonovex-marketplace
```

<!-- xonovex:installation:end -->

## Skills

Choose the skill that owns the task. The harness loads its instructions when the request matches its routing description.

| Skill | Use when |
| --- | --- |
| [using-docker](skills/using-docker/SKILL.md) | Writing or editing a Dockerfile or Docker Compose file for production images. |
| [using-kubernetes](skills/using-kubernetes/SKILL.md) | Writing or editing Kubernetes manifests in a GitOps repository. |
| [using-moon](skills/using-moon/SKILL.md) | Adding or changing tasks in a moonrepo monorepo. |
| [using-terraform](skills/using-terraform/SKILL.md) | Writing or editing Terraform 1.12+ infrastructure code. |

## Validation

Run `npx moon run plugin-platform:ci-check --force` from the repository root to check the plugin manifests, formatting, and owned skills.
