package builder

import (
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
)

// BuildNetworkPolicy creates a NetworkPolicy for an AgentRun's pods.
// It targets pods with the label app.kubernetes.io/instance=<run.Name>.
// Ingress is always denied. Egress uses the rules from the spec; empty rules = deny all.
func BuildNetworkPolicy(run *agentv1alpha1.AgentRun, np *agentv1alpha1.AgentNetworkPolicy) *networkingv1.NetworkPolicy {
	var egress []networkingv1.NetworkPolicyEgressRule
	if np != nil {
		egress = np.Egress
	}

	return &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      run.Name + "-netpol",
			Namespace: run.Namespace,
			Labels: map[string]string{
				"app.kubernetes.io/name":      "agent-operator",
				"app.kubernetes.io/instance":  run.Name,
				"app.kubernetes.io/component": "agent-network-policy",
			},
		},
		Spec: networkingv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app.kubernetes.io/instance": run.Name,
				},
			},
			PolicyTypes: []networkingv1.PolicyType{
				networkingv1.PolicyTypeIngress,
				networkingv1.PolicyTypeEgress,
			},
			Ingress: []networkingv1.NetworkPolicyIngressRule{},
			Egress:  egress,
		},
	}
}
