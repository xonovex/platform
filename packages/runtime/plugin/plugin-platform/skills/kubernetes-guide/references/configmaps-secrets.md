# configmaps-secrets: Configuration and Secret Management

ConfigMaps for non-sensitive config, Secrets for sensitive data; inject both via `envFrom` in pod specs. Use `stringData` so values are base64-encoded automatically. Encrypt Secrets with SOPS or External Secrets Operator before committing to Git: plain Secrets are base64, not encrypted.

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: api-config
data:
  LOG_LEVEL: "info"
  RATE_LIMIT: "100"

---
apiVersion: v1
kind: Secret
metadata:
  name: api-secrets
type: Opaque
stringData:
  DATABASE_URL: "postgresql://user:pass@db:5432/prod"
  API_KEY: "secret-key"
```
