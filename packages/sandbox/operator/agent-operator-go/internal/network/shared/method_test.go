package shared

import (
	"errors"
	"testing"

	networkingv1 "k8s.io/api/networking/v1"

	agentnetwork "github.com/xonovex/platform/packages/library/shared-agent-go/pkg/network"
	agentv1alpha1 "github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/api/v1alpha1"
)

func TestMethodMapsTheEgressOfARun(t *testing.T) {
	custom := &agentv1alpha1.AgentNetworkPolicy{Egress: []networkingv1.NetworkPolicyEgressRule{{}}}
	cases := []struct {
		name    string
		network agentv1alpha1.NetworkMode
		policy  *agentv1alpha1.AgentNetworkPolicy
		want    agentnetwork.NetworkMethod
	}{
		{name: "unset", want: agentnetwork.NetworkNone},
		{name: "none", network: agentv1alpha1.NetworkModeNone, want: agentnetwork.NetworkNone},
		{name: "deny all", network: agentv1alpha1.NetworkModeNone, policy: &agentv1alpha1.AgentNetworkPolicy{}, want: agentnetwork.NetworkNone},
		{name: "host", network: agentv1alpha1.NetworkModeHost, want: agentnetwork.NetworkHost},
		{name: "host with deny all", network: agentv1alpha1.NetworkModeHost, policy: &agentv1alpha1.AgentNetworkPolicy{}, want: agentnetwork.NetworkHost},
		{name: "disabled", network: agentv1alpha1.NetworkModeNone, policy: &agentv1alpha1.AgentNetworkPolicy{Disabled: true}, want: agentnetwork.NetworkHost},
		{name: "custom egress", policy: custom, want: agentnetwork.NetworkHost},
	}
	for _, tc := range cases {
		run := newTestAgentRun("run", "default")
		run.Spec.Network = tc.network
		run.Spec.NetworkPolicy = tc.policy
		got, err := Method(run)
		if err != nil || got != tc.want {
			t.Errorf("%s: Method() = %q, %v, want %q", tc.name, got, err, tc.want)
		}
	}
}

func TestMethodRefusesProxy(t *testing.T) {
	run := newTestAgentRun("run", "default")
	run.Spec.Network = agentv1alpha1.NetworkModeProxy

	if _, err := Method(run); !errors.Is(err, ErrProxyEnforcementUnavailable) {
		t.Fatalf("Method(proxy) err = %v, want ErrProxyEnforcementUnavailable", err)
	}
}
