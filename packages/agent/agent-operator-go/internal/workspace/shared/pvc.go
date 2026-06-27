package shared

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
)

// BuildPVC creates a ReadWriteOnce PersistentVolumeClaim for an AgentRun workspace.
func BuildPVC(name, namespace, storageClass, storageSize string, owner *agentv1alpha1.AgentRun) *corev1.PersistentVolumeClaim {
	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"app.kubernetes.io/name":      "agent-operator",
				"app.kubernetes.io/instance":  owner.Name,
				"app.kubernetes.io/component": "workspace",
			},
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: agentv1alpha1.GroupVersion.String(),
					Kind:       "AgentRun",
					Name:       owner.Name,
					UID:        owner.UID,
				},
			},
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{corev1.ResourceStorage: resource.MustParse(storageSize)},
			},
		},
	}
	if storageClass != "" {
		pvc.Spec.StorageClassName = &storageClass
	}
	return pvc
}

// BuildWorkspacePVC creates a ReadWriteMany PVC for an AgentWorkspace.
func BuildWorkspacePVC(name string, ws *agentv1alpha1.AgentWorkspace) *corev1.PersistentVolumeClaim {
	storageSize := ws.Spec.StorageSize
	if storageSize == "" {
		storageSize = "10Gi"
	}
	pvc := workspacePVC(name, ws, "workspace", storageSize)
	if ws.Spec.StorageClass != "" {
		pvc.Spec.StorageClassName = &ws.Spec.StorageClass
	}
	return pvc
}

// BuildSharedVolumePVC creates a ReadWriteMany PVC for a shared volume.
func BuildSharedVolumePVC(name string, ws *agentv1alpha1.AgentWorkspace, vol agentv1alpha1.SharedVolumeSpec) *corev1.PersistentVolumeClaim {
	storageSize := vol.StorageSize
	if storageSize == "" {
		storageSize = "1Gi"
	}
	pvc := workspacePVC(name, ws, "shared-volume", storageSize)
	if ws.Spec.StorageClass != "" {
		pvc.Spec.StorageClassName = &ws.Spec.StorageClass
	}
	return pvc
}

// workspacePVC builds the common ReadWriteMany PVC skeleton owned by an
// AgentWorkspace.
func workspacePVC(name string, ws *agentv1alpha1.AgentWorkspace, component, storageSize string) *corev1.PersistentVolumeClaim {
	return &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ws.Namespace,
			Labels: map[string]string{
				"app.kubernetes.io/name":      "agent-operator",
				"app.kubernetes.io/instance":  ws.Name,
				"app.kubernetes.io/component": component,
			},
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: agentv1alpha1.GroupVersion.String(),
					Kind:       "AgentWorkspace",
					Name:       ws.Name,
					UID:        ws.UID,
				},
			},
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteMany},
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{corev1.ResourceStorage: resource.MustParse(storageSize)},
			},
		},
	}
}
