# network-policies: Network Segmentation with NetworkPolicy

Control pod ingress/egress with NetworkPolicy. A policy selecting a pod is implicit default-deny for the listed `policyTypes` — anything not explicitly allowed is dropped. Always allow DNS egress (UDP 53 to kube-system), or in-cluster name resolution breaks.

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: api-network-policy
  namespace: production
spec:
  podSelector:
    matchLabels:
      app: api
  policyTypes:
    - Ingress
    - Egress
  ingress:
    - from:
        - namespaceSelector:
            matchLabels:
              name: ingress-nginx
      ports:
        - protocol: TCP
          port: 3000
  egress:
    - to:
        - namespaceSelector:
            matchLabels:
              name: database
      ports:
        - protocol: TCP
          port: 5432
    - to: # Allow DNS
        - namespaceSelector:
            matchLabels:
              name: kube-system
      ports:
        - protocol: UDP
          port: 53
```
