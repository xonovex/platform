# deployments: Deployment Resources

Set resource `requests`/`limits` and `livenessProbe`/`readinessProbe` on every Deployment.

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: api
spec:
  replicas: 3
  template:
    spec:
      containers:
        - name: api
          image: ghcr.io/org/api:1.2.3
          resources:
            requests: {cpu: 100m, memory: 128Mi}
            limits: {cpu: 500m, memory: 256Mi}
          livenessProbe:
            httpGet: {path: /healthz, port: 3000}
            initialDelaySeconds: 10
            periodSeconds: 10
          readinessProbe:
            httpGet: {path: /readyz, port: 3000}
            initialDelaySeconds: 5
            periodSeconds: 5
```
