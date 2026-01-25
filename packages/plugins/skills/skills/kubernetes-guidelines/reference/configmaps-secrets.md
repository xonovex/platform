# configmaps-secrets: Configuration and Secret Management

**Guideline:** Use ConfigMaps for non-sensitive configuration and Secrets for sensitive data. Encrypt secrets at rest with SOPS or External Secrets Operator.

**Rationale:** Separating configuration from images enables environment-specific settings without rebuilding. Encrypting secrets prevents credential exposure in Git.

**Example:**

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

**Techniques:**
- ConfigMaps: Store non-sensitive configuration data for reusability
- Secrets: Store sensitive data like passwords and API keys
- stringData: Use for automatic base64 encoding in YAML
- Encryption: Encrypt secrets with SOPS before committing to Git
- Environment refs: Reference via `envFrom` in pod specs for clean injection
