# services: Kubernetes Service Configuration

Expose pods with a Service whose `selector` exactly matches the Deployment's pod labels. Use named ports (`targetPort: http`) so container ports can change without updating references.

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
