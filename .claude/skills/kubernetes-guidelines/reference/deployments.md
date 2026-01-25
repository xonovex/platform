# deployments: Secure Deployments

**Guideline:** Configure secure, production-ready deployments with security contexts, resource management, and health probes.

**Rationale:** Deployments need security hardening, resource limits, and health checks to prevent privilege escalation, resource exhaustion, and availability issues.

**Example:**

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

**Techniques:**
- Security context: Set runAsNonRoot and fsGroup for pod-level security
- Container hardening: Use readOnlyRootFilesystem and drop all capabilities
- Resource limits: Define requests and limits to prevent resource exhaustion
- Health probes: Add liveness, readiness, startup probes for reliability
- Image tagging: Use specific tags, never `latest`, for reproducibility
- Writable paths: Use emptyDir volumes for writable paths with read-only fs
