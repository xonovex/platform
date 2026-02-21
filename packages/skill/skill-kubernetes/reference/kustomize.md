# kustomize: Multi-Environment Management

**Guideline:** Organize manifests with base resources and environment-specific overlays using strategic merge patches.

**Rationale:** Kustomize enables DRY configuration management across environments. Base resources define common configuration while overlays customize for each environment.

**Example:**

```yaml
# base/kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: default
commonLabels:
  app.kubernetes.io/name: api
resources:
  - deployment.yaml
  - service.yaml
images:
  - name: ghcr.io/org/api
    newTag: 1.2.3

# overlays/production/kustomization.yaml
bases:
  - ../../base
namespace: production
patchesStrategicMerge:
  - replica-count.yaml
  - resources-patch.yaml
configMapGenerator:
  - name: api-config
    literals: [LOG_LEVEL=info, RATE_LIMIT=1000]
```

**Techniques:**

- Base resources: Create base/kustomization.yaml with common manifests
- Environment overlays: Create overlays/<env>/ for environment-specific customization
- Bases reference: Use bases field to inherit and extend base resources
- Strategic patches: Apply patchesStrategicMerge for targeted environment differences
- Config generation: Use configMapGenerator for environment-specific configuration
- Image management: Manage image tags in kustomization.yaml for consistency
