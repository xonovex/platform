# kustomize: Multi-Environment Management

Structure manifests as a `base/` plus per-environment `overlays/<env>/`; overlays inherit the base via `resources` and customize with `patches` (auto-detects strategic-merge vs JSON6902). Manage image tags with `images` and env config with `configMapGenerator`. On Kustomize ≥ 5 prefer `resources`/`patches`/`labels` over the deprecated `bases`/`patchesStrategicMerge`/`commonLabels`.

```yaml
# base/kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: default
labels:
  - pairs:
      app.kubernetes.io/name: api
    includeSelectors: true
resources:
  - deployment.yaml
  - service.yaml
images:
  - name: ghcr.io/org/api
    newTag: 1.2.3

# overlays/production/kustomization.yaml
resources:
  - ../../base
namespace: production
patches:
  - path: replica-count.yaml
  - path: resources-patch.yaml
configMapGenerator:
  - name: api-config
    literals: [LOG_LEVEL=info, RATE_LIMIT=1000]
```
