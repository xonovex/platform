# services: Kubernetes Service Configuration

**Guideline:** Create Services with consistent labels and named ports for service discovery.

**Rationale:** Services provide stable endpoints for pod communication. Named ports improve readability and allow port changes without updating all references.

**Example:**

```yaml
apiVersion: v1
kind: Service
metadata:
  name: api
  namespace: production
  labels:
    app: api
    component: backend
spec:
  type: ClusterIP
  selector:
    app: api
  ports:
    - name: http
      port: 80
      targetPort: http
      protocol: TCP
  sessionAffinity: None
```

**Techniques:**

- ClusterIP type: Use ClusterIP for internal service discovery within cluster
- Selector matching: Match service selector labels exactly with deployment labels
- Named ports: Use named ports matching container port names for clarity
- Label consistency: Apply consistent labeling for service organization
- Port naming: Named ports allow changing container ports without updating all references
- Discovery: Services provide stable DNS names for pod communication
