package builder

import (
	corev1 "k8s.io/api/core/v1"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
)

// AKS node label for confidential compute nodes.
const aksConfidentialNodeLabel = "kubernetes.azure.com/confidential-computing"

// Default runtimeClassNames per TEE type on AKS.
var teeRuntimeClassNames = map[agentv1alpha1.TEEType]string{
	agentv1alpha1.TEETypeAMDSEVSNP: "kata-cc",
	agentv1alpha1.TEETypeIntelTDX:  "kata-tdx",
}

// TEERuntimeClassName returns the runtimeClassName for a TEE configuration.
func TEERuntimeClassName(cc *agentv1alpha1.ConfidentialComputingSpec) *string {
	if cc == nil {
		return nil
	}
	if cc.OverrideRuntimeClassName != nil {
		return cc.OverrideRuntimeClassName
	}
	name := teeRuntimeClassNames[cc.TEE]
	if name == "" {
		return nil
	}
	return &name
}

// TEENodeAffinity returns the node affinity for a TEE configuration.
// Returns nil if DisableNodeAffinity is true or cc is nil.
func TEENodeAffinity(cc *agentv1alpha1.ConfidentialComputingSpec) *corev1.NodeAffinity {
	if cc == nil || cc.DisableNodeAffinity {
		return nil
	}

	return &corev1.NodeAffinity{
		RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
			NodeSelectorTerms: []corev1.NodeSelectorTerm{{
				MatchExpressions: []corev1.NodeSelectorRequirement{{
					Key:      aksConfidentialNodeLabel,
					Operator: corev1.NodeSelectorOpIn,
					Values:   []string{"true"},
				}},
			}},
		},
	}
}
