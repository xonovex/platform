---
name: kubernetes-guide
description: "Use when editing Kubernetes manifests in GitOps repos. Triggers on `.yaml`/`.yml` files in `clusters/`, `apps/`, or k8s manifest directories, and on prompts about Deployments, Services, ConfigMaps, Secrets (SOPS), Kustomize overlays, Flux/ArgoCD, labels, namespaces, or multi-environment configs, even when the user doesn't say 'Kubernetes'. Skip Helm chart authoring, Dockerfile work (use docker-guide), and cloud-provider-managed services."
---

# Kubernetes Coding Guidelines

## Requirements

- Kubernetes ≥ 1.28, Kustomize ≥ 5, GitOps (Flux).

## Essentials

- **Organization** - Use namespaces, labels, annotations consistently, see [references/deployments.md](references/deployments.md)
- **Container images** - No `latest` tags, set requests/limits and probes, see [references/deployments.md](references/deployments.md)
- **Security** - Run as non-root, read-only FS, drop capabilities, see [references/deployments.md](references/deployments.md)
- **Configuration** - Use ConfigMaps/Secrets, SOPS/External Secrets for secrets, see [references/configmaps-secrets.md](references/configmaps-secrets.md)
- **Multi-environment** - Manage with Kustomize bases/overlays, validate with `--dry-run`, see [references/kustomize.md](references/kustomize.md), [references/validation.md](references/validation.md)

## Gotchas

- `Secrets` are base64-encoded, not encrypted — encryption-at-rest requires enabling KMS provider on the cluster
- `resources.requests` is what the scheduler considers; `.limits` is what the kubelet enforces — without requests, pods compete unbounded
- Label selectors are immutable once a Service/Deployment is created — changing them requires recreate, not patch
- Namespace scope: `kubectl` defaults to `default` namespace; cross-namespace traffic needs `<svc>.<ns>.svc.cluster.local` or `NetworkPolicy`
- `Deployment` rolling updates require `readinessProbe` to be honest — a probe that returns 200 too early routes traffic to a not-ready pod

## Progressive disclosure

- Read [references/deployments.md](references/deployments.md) - Load when creating or updating Deployment resources
- Read [references/services.md](references/services.md) - Load when exposing applications or configuring load balancing
- Read [references/configmaps-secrets.md](references/configmaps-secrets.md) - Load when externalizing configuration or managing secrets
- Read [references/kustomize.md](references/kustomize.md) - Load when managing multiple environments with overlays
- Read [references/network-policies.md](references/network-policies.md) - Load when implementing network isolation between pods
- Read [references/validation.md](references/validation.md) - Load when validating manifests before applying to cluster
