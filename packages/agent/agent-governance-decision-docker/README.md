# Governance Decision Service Image

Packages the TypeScript governance decision service as the pod-local sidecar used by the agent operator admission webhook.

```bash
npx moon run agent-governance-decision-docker:docker-build
```

The service listens on port `8787`, exposes `/healthz`, and stores idempotent verdict evidence in `/var/lib/xonovex/verdicts.jsonl`.
