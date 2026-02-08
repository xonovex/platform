package builder

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
)

func TestBuildPVC_Basic(t *testing.T) {
	owner := &agentv1alpha1.AgentRun{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-run",
			UID:  types.UID("test-uid-123"),
		},
	}

	pvc := BuildPVC("test-workspace", "default", "", "10Gi", owner)

	if pvc.Name != "test-workspace" {
		t.Errorf("PVC name = %q, want %q", pvc.Name, "test-workspace")
	}
	if pvc.Namespace != "default" {
		t.Errorf("PVC namespace = %q, want %q", pvc.Namespace, "default")
	}
	if pvc.Labels["app.kubernetes.io/component"] != "workspace" {
		t.Errorf("component label = %q, want %q", pvc.Labels["app.kubernetes.io/component"], "workspace")
	}
	if pvc.Labels["app.kubernetes.io/instance"] != "test-run" {
		t.Errorf("instance label = %q, want %q", pvc.Labels["app.kubernetes.io/instance"], "test-run")
	}

	if len(pvc.OwnerReferences) != 1 {
		t.Fatalf("len(OwnerReferences) = %d, want 1", len(pvc.OwnerReferences))
	}
	ownerRef := pvc.OwnerReferences[0]
	if ownerRef.Name != "test-run" {
		t.Errorf("owner ref name = %q, want %q", ownerRef.Name, "test-run")
	}
	if ownerRef.UID != "test-uid-123" {
		t.Errorf("owner ref UID = %q, want %q", ownerRef.UID, "test-uid-123")
	}
	if ownerRef.Kind != "AgentRun" {
		t.Errorf("owner ref kind = %q, want %q", ownerRef.Kind, "AgentRun")
	}

	if len(pvc.Spec.AccessModes) != 1 || pvc.Spec.AccessModes[0] != corev1.ReadWriteOnce {
		t.Errorf("AccessModes = %v, want [ReadWriteOnce]", pvc.Spec.AccessModes)
	}

	storageReq := pvc.Spec.Resources.Requests[corev1.ResourceStorage]
	if storageReq.String() != "10Gi" {
		t.Errorf("storage request = %q, want %q", storageReq.String(), "10Gi")
	}

	if pvc.Spec.StorageClassName != nil {
		t.Errorf("StorageClassName = %v, want nil", pvc.Spec.StorageClassName)
	}
}

func TestBuildPVC_WithStorageClass(t *testing.T) {
	owner := &agentv1alpha1.AgentRun{
		ObjectMeta: metav1.ObjectMeta{Name: "test-run", UID: "uid"},
	}

	pvc := BuildPVC("ws", "ns", "fast-ssd", "50Gi", owner)

	if pvc.Spec.StorageClassName == nil {
		t.Fatal("StorageClassName is nil, want non-nil")
	}
	if *pvc.Spec.StorageClassName != "fast-ssd" {
		t.Errorf("StorageClassName = %q, want %q", *pvc.Spec.StorageClassName, "fast-ssd")
	}

	storageReq := pvc.Spec.Resources.Requests[corev1.ResourceStorage]
	if storageReq.String() != "50Gi" {
		t.Errorf("storage request = %q, want %q", storageReq.String(), "50Gi")
	}
}
