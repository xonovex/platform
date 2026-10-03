# Run an agent in Kubernetes

Create a policy-governed namespace, configure a provider, and submit an AgentRun to start one agent Job. The cluster must already support the selected sandbox runtime and a digest-pinned agent image.

Install the operator only after the cluster has a digest-pinned agent image and a sandboxed RuntimeClass such as gVisor or Kata. Set these variables to values available in the cluster.

```bash
export XONOVEX_AGENT_IMAGE='ghcr.io/your-org/xonovex-agent@sha256:<64-hex-digest>'
export XONOVEX_RUNTIME_CLASS='gvisor'

# Requires cert-manager v1.16+ with its CA injector enabled
# Install CRDs and deploy the operator
kubectl apply -k https://github.com/xonovex/platform//packages/sandbox/operator/agent-operator-go/config/crd
kubectl apply -k https://github.com/xonovex/platform//packages/sandbox/operator/agent-operator-go/config/default

# Create one policy-governed namespace and provider credential
kubectl create namespace ai-agents --dry-run=client -o yaml | kubectl apply -f -
kubectl -n ai-agents create secret generic anthropic-credentials \
  --from-literal=api-key='your-key' \
  --dry-run=client -o yaml | kubectl apply -f -

kubectl apply -f - <<EOF
apiVersion: agent.xonovex.com/v1alpha1
kind: AgentPolicy
metadata:
  name: sandbox-policy
  namespace: ai-agents
spec:
  enforced:
    runtimeClassName: ${XONOVEX_RUNTIME_CLASS}
    requireSecurityContext: true
    requireNetworkPolicy: true
    maxTimeout: 1h0m0s
    maxResources:
      cpu: "2"
      memory: 4Gi
    allowedImages:
      - ${XONOVEX_AGENT_IMAGE}
    allowedRuntimeClassNames:
      - ${XONOVEX_RUNTIME_CLASS}
    allowedSecretNames:
      - anthropic-credentials
  defaults:
    image: ${XONOVEX_AGENT_IMAGE}
    runtimeClassName: ${XONOVEX_RUNTIME_CLASS}
    timeout: 30m0s
---
apiVersion: agent.xonovex.com/v1alpha1
kind: AgentProvider
metadata:
  name: anthropic-provider
  namespace: ai-agents
spec:
  displayName: Anthropic Claude
  authTokenSecretRef:
    name: anthropic-credentials
    key: api-key
  authTokenEnv: ANTHROPIC_API_KEY
  environment:
    ANTHROPIC_BASE_URL: https://api.anthropic.com
---
apiVersion: agent.xonovex.com/v1alpha1
kind: AgentRun
metadata:
  name: review-code
  namespace: ai-agents
spec:
  harness:
    type: claude
  providerRef: anthropic-provider
  workspace:
    type: git
    repository:
      url: https://github.com/xonovex/platform.git
      branch: main
  prompt: "Review the codebase and suggest improvements"
  network: host
  resources:
    requests:
      cpu: 500m
      memory: 512Mi
    limits:
      cpu: "2"
      memory: 2Gi
EOF
```

The `network: host` value permits unrestricted egress so this example can reach the public model API. Production namespaces should use an enforceable cluster-level egress proxy or a fully qualified domain name aware policy.

## Observe the run

Watch the AgentRun status and use the [operator monitoring commands](../README.md#monitoring-runs) to inspect its Job and logs.

```bash
kubectl -n ai-agents get agentruns -w
```

See the [resource reference](../README.md#custom-resources) for reusable harnesses, providers, workspaces, toolchains, and policy fields.
