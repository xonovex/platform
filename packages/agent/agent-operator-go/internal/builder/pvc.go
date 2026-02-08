package builder

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
)

// BuildPVC creates a PersistentVolumeClaim for an AgentRun workspace
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
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse(storageSize),
				},
			},
		},
	}

	if storageClass != "" {
		pvc.Spec.StorageClassName = &storageClass
	}

	return pvc
}
