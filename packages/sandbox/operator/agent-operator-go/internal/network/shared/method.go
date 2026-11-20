package shared

import (
	"fmt"

	agentnetwork "github.com/xonovex/platform/packages/library/shared-agent-go/pkg/network"
	agentv1alpha1 "github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/api/v1alpha1"
)

// Method maps the egress of a run onto the network axis of the shared plan.
// The NetworkPolicy the operator creates for the run decides it: none and unset
// restrict egress to DNS, while host, a disabled NetworkPolicy and custom
// egress rules, which no check can prove restricted, leave egress open. proxy
// fails closed until a backend can enforce an FQDN allowlist.
func Method(run *agentv1alpha1.AgentRun) (agentnetwork.NetworkMethod, error) {
	switch run.Spec.Network {
	case agentv1alpha1.NetworkModeProxy:
		return "", fmt.Errorf("plan network: %w", ErrProxyEnforcementUnavailable)
	case agentv1alpha1.NetworkModeHost:
		return agentnetwork.NetworkHost, nil
	}
	if policy := run.Spec.NetworkPolicy; policy != nil && (policy.Disabled || len(policy.Egress) > 0) {
		return agentnetwork.NetworkHost, nil
	}
	return agentnetwork.NetworkNone, nil
}
