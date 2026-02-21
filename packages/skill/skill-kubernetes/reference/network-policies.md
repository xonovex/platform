# network-policies: Network Segmentation with NetworkPolicy

**Guideline:** Implement NetworkPolicies to control ingress and egress traffic between pods and namespaces.

**Rationale:** Provides network segmentation and least-privilege access control, preventing unauthorized communication between workloads.

**Example:**

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

**Techniques:**

- Pod selection: Use podSelector with matchLabels to target specific pods
- Policy types: Specify policyTypes for ingress, egress, or both rules
- Namespace selection: Use namespaceSelector to allow traffic from namespaces
- DNS allowance: Always allow DNS egress to kube-system for pod resolution
- Explicit rules: Define exact allowed ports and protocols for security
- Default deny: NetworkPolicy acts as implicit default-deny by default
