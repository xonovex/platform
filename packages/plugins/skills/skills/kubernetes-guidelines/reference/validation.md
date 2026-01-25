# validation: Manifest Validation Commands

**Guideline:** Always validate Kubernetes manifests before applying to clusters using dry-run and diff commands.

**Rationale:** Validation catches configuration errors, API version mismatches, and unintended changes before they impact running workloads.

**Example:**

```bash
# Dry-run validation
kubectl apply --dry-run=server -f deployment.yaml

# Kustomize build
kustomize build k8s/overlays/production

# Apply with Kustomize
kubectl apply -k k8s/overlays/production

# Validate with server-side dry-run
kubectl apply -k k8s/overlays/production --dry-run=server

# Diff before applying
kubectl diff -k k8s/overlays/production
```

**Techniques:**
- Server dry-run: Use `--dry-run=server` to validate against cluster API
- Kustomize build: Run `kustomize build` to verify overlay composition
- Kubectl diff: Use `kubectl diff` to preview exact changes before applying
- Manifest validation: Validate both raw manifests and kustomized output
- Pre-flight checks: Catch configuration errors before they impact workloads
- Dry-run safety: Server-side dry-run validates against current cluster state
