package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// GroupVersion is the group version of the objects in this package.
var GroupVersion = schema.GroupVersion{Group: "agent.xonovex.com", Version: "v1alpha1"}

// AddToScheme adds every object of this group version to scheme. The list is
// explicit, so registering the types changes no package state.
func AddToScheme(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(GroupVersion,
		&AgentHarness{}, &AgentHarnessList{},
		&AgentPolicy{}, &AgentPolicyList{},
		&AgentProvider{}, &AgentProviderList{},
		&AgentRun{}, &AgentRunList{},
		&AgentToolchain{}, &AgentToolchainList{},
		&AgentWorkspace{}, &AgentWorkspaceList{},
	)
	metav1.AddToGroupVersion(scheme, GroupVersion)
	return nil
}
