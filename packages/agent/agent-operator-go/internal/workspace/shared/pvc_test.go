package shared

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
)

func TestBuildPVC_Basic(t *testing.T) {
	owner := &agentv1alpha1.AgentRun{
		ObjectMeta: metav1.ObjectMeta{Name: "test-run", UID: types.UID("test-uid-123")},
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
	if ownerRef.Name != "test-run" || ownerRef.UID != "test-uid-123" || ownerRef.Kind != "AgentRun" {
		t.Errorf("owner ref = %+v, want AgentRun/test-run/test-uid-123", ownerRef)
	}
	if len(pvc.Spec.AccessModes) != 1 || pvc.Spec.AccessModes[0] != corev1.ReadWriteOnce {
		t.Errorf("AccessModes = %v, want [ReadWriteOnce]", pvc.Spec.AccessModes)
	}
	if storageReq := pvc.Spec.Resources.Requests[corev1.ResourceStorage]; storageReq.String() != "10Gi" {
		t.Errorf("storage request = %q, want %q", storageReq.String(), "10Gi")
	}
	if pvc.Spec.StorageClassName != nil {
		t.Errorf("StorageClassName = %v, want nil", pvc.Spec.StorageClassName)
	}
}

func TestBuildPVC_WithStorageClass(t *testing.T) {
	owner := &agentv1alpha1.AgentRun{ObjectMeta: metav1.ObjectMeta{Name: "test-run", UID: "uid"}}
	pvc := BuildPVC("ws", "ns", "fast-ssd", "50Gi", owner)

	if pvc.Spec.StorageClassName == nil || *pvc.Spec.StorageClassName != "fast-ssd" {
		t.Fatal("StorageClassName should be fast-ssd")
	}
	if storageReq := pvc.Spec.Resources.Requests[corev1.ResourceStorage]; storageReq.String() != "50Gi" {
		t.Errorf("storage request = %q, want %q", storageReq.String(), "50Gi")
	}
}

func testWorkspace() *agentv1alpha1.AgentWorkspace {
	return &agentv1alpha1.AgentWorkspace{
		ObjectMeta: metav1.ObjectMeta{Name: "my-workspace", Namespace: "default", UID: types.UID("test-uid")},
	}
}

func TestBuildWorkspacePVC_Basic(t *testing.T) {
	ws := testWorkspace()
	ws.Spec.StorageSize = "10Gi"
	pvc := BuildWorkspacePVC("my-workspace-ws", ws)

	if pvc.Name != "my-workspace-ws" || pvc.Namespace != "default" {
		t.Errorf("name/ns = %s/%s", pvc.Name, pvc.Namespace)
	}
	if pvc.Labels["app.kubernetes.io/component"] != "workspace" {
		t.Errorf("component = %s", pvc.Labels["app.kubernetes.io/component"])
	}
	if pvc.Labels["app.kubernetes.io/instance"] != "my-workspace" {
		t.Errorf("instance = %s", pvc.Labels["app.kubernetes.io/instance"])
	}
	if len(pvc.Spec.AccessModes) != 1 || pvc.Spec.AccessModes[0] != corev1.ReadWriteMany {
		t.Errorf("access mode = %v, want RWX", pvc.Spec.AccessModes)
	}
	if len(pvc.OwnerReferences) != 1 || pvc.OwnerReferences[0].Kind != "AgentWorkspace" || pvc.OwnerReferences[0].Name != "my-workspace" {
		t.Errorf("owner ref = %+v", pvc.OwnerReferences)
	}
	if got := pvc.Spec.Resources.Requests[corev1.ResourceStorage]; !got.Equal(resource.MustParse("10Gi")) {
		t.Errorf("size = %s, want 10Gi", got.String())
	}
}

func TestBuildWorkspacePVC_DefaultStorageSize(t *testing.T) {
	pvc := BuildWorkspacePVC("my-workspace-ws", testWorkspace())
	if got := pvc.Spec.Resources.Requests[corev1.ResourceStorage]; !got.Equal(resource.MustParse("10Gi")) {
		t.Errorf("default size = %s, want 10Gi", got.String())
	}
}

func TestBuildWorkspacePVC_WithStorageClass(t *testing.T) {
	ws := testWorkspace()
	ws.Spec.StorageClass = "nfs-csi"
	ws.Spec.StorageSize = "20Gi"
	pvc := BuildWorkspacePVC("my-workspace-ws", ws)
	if pvc.Spec.StorageClassName == nil || *pvc.Spec.StorageClassName != "nfs-csi" {
		t.Error("expected storage class nfs-csi")
	}
}

func TestBuildSharedVolumePVC_Basic(t *testing.T) {
	vol := agentv1alpha1.SharedVolumeSpec{Name: "claude-config", MountPath: "/root/.claude", StorageSize: "2Gi"}
	pvc := BuildSharedVolumePVC("my-workspace-claude-config", testWorkspace(), vol)

	if pvc.Name != "my-workspace-claude-config" {
		t.Errorf("name = %s", pvc.Name)
	}
	if pvc.Labels["app.kubernetes.io/component"] != "shared-volume" {
		t.Errorf("component = %s, want shared-volume", pvc.Labels["app.kubernetes.io/component"])
	}
	if len(pvc.Spec.AccessModes) != 1 || pvc.Spec.AccessModes[0] != corev1.ReadWriteMany {
		t.Error("expected RWX access mode")
	}
	if pvc.OwnerReferences[0].Kind != "AgentWorkspace" {
		t.Errorf("owner kind = %s", pvc.OwnerReferences[0].Kind)
	}
	if got := pvc.Spec.Resources.Requests[corev1.ResourceStorage]; !got.Equal(resource.MustParse("2Gi")) {
		t.Errorf("size = %s, want 2Gi", got.String())
	}
}

func TestBuildSharedVolumePVC_DefaultStorageSize(t *testing.T) {
	vol := agentv1alpha1.SharedVolumeSpec{Name: "claude-config", MountPath: "/root/.claude"}
	pvc := BuildSharedVolumePVC("my-workspace-claude-config", testWorkspace(), vol)
	if got := pvc.Spec.Resources.Requests[corev1.ResourceStorage]; !got.Equal(resource.MustParse("1Gi")) {
		t.Errorf("default size = %s, want 1Gi", got.String())
	}
}
