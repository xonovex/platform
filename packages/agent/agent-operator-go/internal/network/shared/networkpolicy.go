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
	agentvalidation "github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/validation"
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

// BuildWorkspaceInitNetworkPolicy creates the deny-ingress policy for a
// workspace clone pod. Egress is limited to DNS and the repository transport's
// TCP port; the clone job has no need for any other network traffic.
func BuildWorkspaceInitNetworkPolicy(workspace *agentv1alpha1.AgentWorkspace) (*networkingv1.NetworkPolicy, error) {
	repository, err := agentvalidation.ParseRepositoryURL(workspace.Spec.Repository.URL)
	if err != nil {
		return nil, err
	}
	tcp := corev1.ProtocolTCP
	portValue := intstr.FromInt32(repository.Port)
	return &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      workspace.Name + "-init-netpol",
			Namespace: workspace.Namespace,
			Labels: map[string]string{
				"app.kubernetes.io/name":      "agent-operator",
				"app.kubernetes.io/instance":  workspace.Name,
				"app.kubernetes.io/component": "workspace-init-network-policy",
			},
		},
		Spec: networkingv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{MatchLabels: map[string]string{
				"app.kubernetes.io/instance":  workspace.Name,
				"app.kubernetes.io/component": "workspace-init",
			}},
			PolicyTypes: []networkingv1.PolicyType{
				networkingv1.PolicyTypeIngress,
				networkingv1.PolicyTypeEgress,
			},
			Ingress: []networkingv1.NetworkPolicyIngressRule{},
			Egress: []networkingv1.NetworkPolicyEgressRule{
				dnsEgressRule(),
				{
					To:    publicNetworkPeers(),
					Ports: []networkingv1.NetworkPolicyPort{{Protocol: &tcp, Port: &portValue}},
				},
			},
		},
	}, nil
}

func publicNetworkPeers() []networkingv1.NetworkPolicyPeer {
	return []networkingv1.NetworkPolicyPeer{
		{IPBlock: &networkingv1.IPBlock{
			CIDR: "0.0.0.0/0",
			Except: []string{
				"0.0.0.0/8", "10.0.0.0/8", "100.64.0.0/10", "127.0.0.0/8",
				"169.254.0.0/16", "172.16.0.0/12", "192.0.0.0/24", "192.0.2.0/24",
				"192.88.99.0/24", "192.168.0.0/16", "198.18.0.0/15", "198.51.100.0/24",
				"203.0.113.0/24", "224.0.0.0/4", "240.0.0.0/4",
			},
		}},
		{IPBlock: &networkingv1.IPBlock{
			CIDR: "::/0",
			Except: []string{
				"::/128", "::1/128", "64:ff9b:1::/48", "100::/64",
				"2001:db8::/32", "fc00::/7", "fe80::/10", "ff00::/8",
			},
		}},
	}
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
