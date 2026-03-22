package builder

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
)

func newTestAgentRun(name, namespace string) *agentv1alpha1.AgentRun {
	return &agentv1alpha1.AgentRun{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
}

func TestBuildNetworkPolicy_DenyAll(t *testing.T) {
	run := newTestAgentRun("my-run", "default")
	np := BuildNetworkPolicy(run, nil)

	if np.Name != "my-run-netpol" {
		t.Errorf("name = %q, want %q", np.Name, "my-run-netpol")
	}
	if np.Namespace != "default" {
		t.Errorf("namespace = %q, want %q", np.Namespace, "default")
	}
	if np.Spec.Egress != nil {
		t.Errorf("egress = %v, want nil (deny all)", np.Spec.Egress)
	}
}

func TestBuildNetworkPolicy_WithEgressRules(t *testing.T) {
	run := newTestAgentRun("my-run", "default")
	port := intstr.FromInt32(443)
	protocol := networkingv1.NetworkPolicyEgressRule{
		To: []networkingv1.NetworkPolicyPeer{
			{
				IPBlock: &networkingv1.IPBlock{
					CIDR: "0.0.0.0/0",
				},
			},
		},
		Ports: []networkingv1.NetworkPolicyPort{
			{
				Port:     &port,
				Protocol: protocolPtr(corev1.ProtocolTCP),
			},
		},
	}

	np := BuildNetworkPolicy(run, &agentv1alpha1.AgentNetworkPolicy{
		Egress: []networkingv1.NetworkPolicyEgressRule{protocol},
	})

	if len(np.Spec.Egress) != 1 {
		t.Fatalf("egress rules = %d, want 1", len(np.Spec.Egress))
	}
	if np.Spec.Egress[0].To[0].IPBlock.CIDR != "0.0.0.0/0" {
		t.Errorf("CIDR = %q, want %q", np.Spec.Egress[0].To[0].IPBlock.CIDR, "0.0.0.0/0")
	}
}

func TestBuildNetworkPolicy_Labels(t *testing.T) {
	run := newTestAgentRun("test-run", "test-ns")
	np := BuildNetworkPolicy(run, nil)

	expectedLabels := map[string]string{
		"app.kubernetes.io/name":      "agent-operator",
		"app.kubernetes.io/instance":  "test-run",
		"app.kubernetes.io/component": "agent-network-policy",
	}
	for k, v := range expectedLabels {
		if np.Labels[k] != v {
			t.Errorf("label %q = %q, want %q", k, np.Labels[k], v)
		}
	}

	selector := np.Spec.PodSelector.MatchLabels
	if selector["app.kubernetes.io/instance"] != "test-run" {
		t.Errorf("pod selector instance = %q, want %q", selector["app.kubernetes.io/instance"], "test-run")
	}
}

func TestBuildNetworkPolicy_PolicyTypes(t *testing.T) {
	run := newTestAgentRun("my-run", "default")
	np := BuildNetworkPolicy(run, nil)

	if len(np.Spec.PolicyTypes) != 2 {
		t.Fatalf("policyTypes = %d, want 2", len(np.Spec.PolicyTypes))
	}
	if np.Spec.PolicyTypes[0] != networkingv1.PolicyTypeIngress {
		t.Errorf("policyTypes[0] = %q, want Ingress", np.Spec.PolicyTypes[0])
	}
	if np.Spec.PolicyTypes[1] != networkingv1.PolicyTypeEgress {
		t.Errorf("policyTypes[1] = %q, want Egress", np.Spec.PolicyTypes[1])
	}
}

func protocolPtr(p corev1.Protocol) *corev1.Protocol {
	return &p
}
