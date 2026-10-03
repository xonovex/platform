# validation: Manifest Validation Commands

Validate before applying: `--dry-run=server` checks against the live cluster API (not just client-side schema), `kustomize build` verifies overlay composition, and `kubectl diff -k` previews the exact change.

```bash
kubectl apply --dry-run=server -f deployment.yaml
kustomize build k8s/overlays/production
kubectl apply -k k8s/overlays/production
kubectl apply -k k8s/overlays/production --dry-run=server
kubectl diff -k k8s/overlays/production
```
