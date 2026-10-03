package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	// GroupVersion is group version used to register these objects
	GroupVersion = schema.GroupVersion{Group: "agent.xonovex.com", Version: "v1alpha1"}

	// SchemeBuilder is used to add go types to the GroupVersionResource scheme
	SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)

	// AddToScheme adds the types in this group-version to the given scheme.
	AddToScheme = SchemeBuilder.AddToScheme

	// knownTypes holds the go types registered via registerTypes.
	knownTypes []runtime.Object
)

// registerTypes records types to be added to the scheme by AddToScheme.
// Each types file calls this from its init function.
func registerTypes(objects ...runtime.Object) {
	knownTypes = append(knownTypes, objects...)
}

func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(GroupVersion, knownTypes...)
	metav1.AddToGroupVersion(scheme, GroupVersion)
	return nil
}
