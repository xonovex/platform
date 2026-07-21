package testutil

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
)

// NewScheme creates a runtime.Scheme with all types required by the operator.
func NewScheme() *runtime.Scheme {
	s := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(s); err != nil {
		panic(fmt.Sprintf("register Kubernetes core API: %v", err))
	}
	if err := agentv1alpha1.AddToScheme(s); err != nil {
		panic(fmt.Sprintf("register agent API: %v", err))
	}
	return s
}
