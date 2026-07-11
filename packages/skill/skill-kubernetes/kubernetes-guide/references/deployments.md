# deployments: Secure Deployments

Harden Deployments: pin image tags (never `latest`), set both pod- and container-level `securityContext` (runAsNonRoot, runAsUser/fsGroup, readOnlyRootFilesystem, allowPrivilegeEscalation false, drop ALL capabilities, seccompProfile RuntimeDefault), define resource requests/limits, and add liveness/readiness probes. With a read-only root fs, mount an `emptyDir` on writable paths (e.g. `/tmp`).

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: api
spec:
  replicas: 3
  template:
    spec:
      securityContext:
        runAsNonRoot: true
        runAsUser: 1000
        fsGroup: 1000
        seccompProfile: {type: RuntimeDefault}
      containers:
        - name: api
          image: ghcr.io/org/api:1.2.3
          securityContext:
            readOnlyRootFilesystem: true
            allowPrivilegeEscalation: false
            capabilities: {drop: [ALL]}
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
          volumeMounts:
            - name: tmp
              mountPath: /tmp
      volumes:
        - name: tmp
          emptyDir: {}
```
