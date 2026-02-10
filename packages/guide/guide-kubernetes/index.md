---
name: kubernetes-guidelines
description: Trigger on `.yaml/.yml` files in clusters/ or k8s directories. Use when writing Kubernetes manifests for GitOps deployments. Apply for Deployments, Services, security, multi-environment config. Keywords: Kubernetes, Deployment, Service, ConfigMap, Secret, Kustomize, Flux, RLS, SOPS, labels, namespaces.
---

# Kubernetes Coding Guidelines

## Requirements

- Kubernetes ≥ 1.28, Kustomize ≥ 5, GitOps (Flux).

## Essentials

- **Organization** - Use namespaces, labels, annotations consistently, see [reference/deployments.md](reference/deployments.md)
- **Container images** - No `latest` tags, set requests/limits and probes, see [reference/deployments.md](reference/deployments.md)
- **Security** - Run as non-root, read-only FS, drop capabilities, see [reference/deployments.md](reference/deployments.md)
- **Configuration** - Use ConfigMaps/Secrets, SOPS/External Secrets for secrets, see [reference/configmaps-secrets.md](reference/configmaps-secrets.md)
- **Multi-environment** - Manage with Kustomize bases/overlays, validate with `--dry-run`, see [reference/kustomize.md](reference/kustomize.md), [reference/validation.md](reference/validation.md)

## Progressive disclosure

- Read [reference/deployments.md](reference/deployments.md) - When creating or updating Deployment resources
- Read [reference/services.md](reference/services.md) - When exposing applications or configuring load balancing
- Read [reference/configmaps-secrets.md](reference/configmaps-secrets.md) - When externalizing configuration or managing secrets
- Read [reference/kustomize.md](reference/kustomize.md) - When managing multiple environments with overlays
- Read [reference/network-policies.md](reference/network-policies.md) - When implementing network isolation between pods
- Read [reference/validation.md](reference/validation.md) - When validating manifests before applying to cluster
