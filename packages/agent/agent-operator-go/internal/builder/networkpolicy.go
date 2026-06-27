package builder

import (
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
)

// privateCIDRs are denied for Network=proxy: link-local + metadata
// (169.254.169.254), the RFC1918 ranges, and loopback. Egress is otherwise
// allowed to the public internet, narrowed to the allowlist by an FQDN-aware
// upgrade (Cilium toFQDNs or a Squid proxy) layered on top.
var privateCIDRs = []string{
	"169.254.0.0/16",
	"10.0.0.0/8",
	"172.16.0.0/12",
	"192.168.0.0/16",
	"127.0.0.0/8",
}

// BuildNetworkPolicy creates a per-AgentRun NetworkPolicy. Ingress is always
// denied. Egress is ALWAYS default-deny unless rules open it:
//   - explicit np.Egress rules take precedence (backward compatible);
//   - otherwise run.Spec.Network maps to rules — host = allow all (does NOT
//     satisfy egress-restricted); none = DNS only; proxy = DNS + public egress
//     except metadata/RFC1918/loopback.
func BuildNetworkPolicy(run *agentv1alpha1.AgentRun, np *agentv1alpha1.AgentNetworkPolicy) *networkingv1.NetworkPolicy {
	var egress []networkingv1.NetworkPolicyEgressRule
	switch {
	case np != nil && len(np.Egress) > 0:
		egress = np.Egress
	default:
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
	}
}

// egressForNetwork maps the egress axis to NetworkPolicy rules. The default
// (empty/none) is DNS-only — never implicit open egress.
func egressForNetwork(network string) []networkingv1.NetworkPolicyEgressRule {
	switch network {
	case "host":
		// A single empty rule allows all egress (explicit opt-in).
		return []networkingv1.NetworkPolicyEgressRule{{}}
	case "proxy":
		return []networkingv1.NetworkPolicyEgressRule{dnsEgressRule(), publicExceptPrivateRule()}
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

// publicExceptPrivateRule allows public egress while blocking metadata, RFC1918,
// link-local, and loopback.
func publicExceptPrivateRule() networkingv1.NetworkPolicyEgressRule {
	return networkingv1.NetworkPolicyEgressRule{
		To: []networkingv1.NetworkPolicyPeer{
			{IPBlock: &networkingv1.IPBlock{CIDR: "0.0.0.0/0", Except: privateCIDRs}},
		},
	}
}
