// Package shared is the operator's network axis core: the per-AgentRun
// NetworkPolicy builder. Network is a closed set mapped to egress rules; there is
// no per-type leaf (the operator does not vary the realizer per network mode).
package shared

import (
	"errors"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
)

// ErrProxyEnforcementUnavailable reports that proxy mode has no backend capable
// of preventing direct sockets from bypassing an HTTP(S) allowlist.
var ErrProxyEnforcementUnavailable = errors.New("network=proxy has no enforceable FQDN-aware backend")

// BuildNetworkPolicy creates a per-AgentRun NetworkPolicy. Ingress is always
// denied. Egress is ALWAYS default-deny unless rules open it:
//   - an explicit NetworkPolicy takes precedence, including an empty deny-all
//     egress list;
//   - otherwise host allows all and none/unset allows DNS only;
//   - proxy fails closed until an FQDN-aware backend is available.
func BuildNetworkPolicy(run *agentv1alpha1.AgentRun, np *agentv1alpha1.AgentNetworkPolicy) (*networkingv1.NetworkPolicy, error) {
	if run.Spec.Network == agentv1alpha1.NetworkModeProxy {
		return nil, fmt.Errorf("build network policy: %w", ErrProxyEnforcementUnavailable)
	}

	var egress []networkingv1.NetworkPolicyEgressRule
	if np != nil {
		egress = np.Egress
	} else {
		egress = egressForNetwork(run.Spec.Network)
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
	}, nil
}

// egressForNetwork maps the egress axis to NetworkPolicy rules. The default
// (empty/none) is DNS-only — never implicit open egress.
func egressForNetwork(network agentv1alpha1.NetworkMode) []networkingv1.NetworkPolicyEgressRule {
	switch network {
	case agentv1alpha1.NetworkModeHost:
		// A single empty rule allows all egress (explicit opt-in).
		return []networkingv1.NetworkPolicyEgressRule{{}}
	default: // "none" and unset: DNS only so the pod can resolve, nothing else.
		return []networkingv1.NetworkPolicyEgressRule{dnsEgressRule()}
	}
}

// dnsEgressRule allows DNS to kube-system so pods can resolve names.
func dnsEgressRule() networkingv1.NetworkPolicyEgressRule {
	udp := corev1.ProtocolUDP
	tcp := corev1.ProtocolTCP
	port53 := intstr.FromInt32(53)
	return networkingv1.NetworkPolicyEgressRule{
		To: []networkingv1.NetworkPolicyPeer{
			{NamespaceSelector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"kubernetes.io/metadata.name": "kube-system"},
			}},
		},
		Ports: []networkingv1.NetworkPolicyPort{
			{Protocol: &udp, Port: &port53},
			{Protocol: &tcp, Port: &port53},
		},
	}
}
