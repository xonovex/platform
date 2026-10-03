package v1alpha1

import (
	"reflect"
	"testing"

	"k8s.io/apimachinery/pkg/runtime"
)

func TestAddToSchemeRegistersRetainedTypes(t *testing.T) {
	testCases := []struct {
		kind     string
		wantType runtime.Object
	}{
		{kind: "AgentHarness", wantType: &AgentHarness{}},
		{kind: "AgentHarnessList", wantType: &AgentHarnessList{}},
		{kind: "AgentPolicy", wantType: &AgentPolicy{}},
		{kind: "AgentPolicyList", wantType: &AgentPolicyList{}},
		{kind: "AgentProvider", wantType: &AgentProvider{}},
		{kind: "AgentProviderList", wantType: &AgentProviderList{}},
		{kind: "AgentRun", wantType: &AgentRun{}},
		{kind: "AgentRunList", wantType: &AgentRunList{}},
		{kind: "AgentToolchain", wantType: &AgentToolchain{}},
		{kind: "AgentToolchainList", wantType: &AgentToolchainList{}},
		{kind: "AgentWorkspace", wantType: &AgentWorkspace{}},
		{kind: "AgentWorkspaceList", wantType: &AgentWorkspaceList{}},
	}
	scheme := runtime.NewScheme()
	if err := AddToScheme(scheme); err != nil {
		t.Fatalf("add retained API types to scheme: %v", err)
	}

	for _, testCase := range testCases {
		t.Run(testCase.kind, func(t *testing.T) {
			object, err := scheme.New(GroupVersion.WithKind(testCase.kind))

			if err != nil {
				t.Fatalf("create %s from scheme: %v", testCase.kind, err)
			}
			if gotType, wantType := reflect.TypeOf(object), reflect.TypeOf(testCase.wantType); gotType != wantType {
				t.Fatalf("registered type = %v, want %v", gotType, wantType)
			}
		})
	}
}

func TestRetainedTypesSupportRuntimeDeepCopy(t *testing.T) {
	objects := []runtime.Object{
		&AgentHarness{},
		&AgentHarnessList{},
		&AgentPolicy{},
		&AgentPolicyList{},
		&AgentProvider{},
		&AgentProviderList{},
		&AgentRun{},
		&AgentRunList{},
		&AgentToolchain{},
		&AgentToolchainList{},
		&AgentWorkspace{},
		&AgentWorkspaceList{},
	}

	for _, object := range objects {
		objectType := reflect.TypeOf(object)
		t.Run(objectType.Elem().Name(), func(t *testing.T) {
			copied := object.DeepCopyObject()

			if reflect.TypeOf(copied) != objectType {
				t.Fatalf("deep-copy type = %T, want %v", copied, objectType)
			}
			if copied == object {
				t.Fatal("deep copy returned the source object")
			}
		})
	}
}
