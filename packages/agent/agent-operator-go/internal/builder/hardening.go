package builder

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
)

// BuildAgentServiceAccount returns the dedicated zero-RBAC ServiceAccount agent
// pods bind to. It has no RoleBindings (the agent never calls the Kubernetes API)
// and disables token automounting at the account level — defense in depth on top
// of the pod-level AutomountServiceAccountToken=false.
func BuildAgentServiceAccount(namespace string) *corev1.ServiceAccount {
	automount := false
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      AgentServiceAccountName,
			Namespace: namespace,
			Labels: map[string]string{
				"app.kubernetes.io/name":      "agent-operator",
				"app.kubernetes.io/component": "agent-serviceaccount",
			},
		},
		AutomountServiceAccountToken: &automount,
	}
}

const (
	// AgentServiceAccountName is the dedicated zero-RBAC ServiceAccount for agent
	// pods. The agent runs untrusted code and never calls the Kubernetes API.
	AgentServiceAccountName = "agent-runner"
	// agentHome matches the uid-1000 HOME baked into the nix agent image.
	agentHome      = "/home/agent"
	homeVolumeName = "home"
	agentFSGroup   = int64(1000)
)

// defaultAgentResources bounds node-DoS for an untrusted workload. They are
// deliberately generous (containment, not a quota) and overridden by any
// requests/limits the run specifies.
func defaultAgentResources() corev1.ResourceRequirements {
	return corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("250m"),
			corev1.ResourceMemory: resource.MustParse("512Mi"),
		},
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("2"),
			corev1.ResourceMemory: resource.MustParse("4Gi"),
		},
	}
}

// applyPodHardening applies the deny-default pod-path hardening shared by the
// standalone and workspace Jobs:
//   - a dedicated zero-RBAC ServiceAccount with no mounted token (the agent never
//     reaches the Kubernetes API);
//   - default resource bounds on the agent container when the run did not request
//     them (node-DoS containment);
//   - for the nix image (non-root uid 1000 with a read-only rootfs), a writable
//     HOME emptyDir + fsGroup so the agent can write config/cache/state. The image
//     bakes only empty XDG dirs under HOME, so the mount shadows nothing.
//
// The container runtimeClassName is left to run.Spec.RuntimeClassName (defaulted
// from the harness DefaultRuntimeClassName): RequireKernelIsolation is satisfied
// only by a sandboxed runtimeClass, never by default runc.
func applyPodHardening(spec *corev1.PodSpec, run *agentv1alpha1.AgentRun, tc *agentv1alpha1.ToolchainSpec) {
	autoMount := false
	spec.ServiceAccountName = AgentServiceAccountName
	spec.AutomountServiceAccountToken = &autoMount

	if len(spec.Containers) > 0 {
		if len(run.Spec.Resources.Requests) > 0 || len(run.Spec.Resources.Limits) > 0 {
			spec.Containers[0].Resources = run.Spec.Resources
		} else {
			spec.Containers[0].Resources = defaultAgentResources()
		}
	}

	// An image-based toolchain runs non-root at a read-only rootfs; reconcile with
	// a writable HOME emptyDir + fsGroup.
	if tcl := ResolveToolchain(tc); tcl != nil && tcl.Image() != "" {
		if spec.SecurityContext == nil {
			spec.SecurityContext = &corev1.PodSecurityContext{}
		}
		if spec.SecurityContext.FSGroup == nil {
			gid := agentFSGroup
			spec.SecurityContext.FSGroup = &gid
		}
		spec.Volumes = append(spec.Volumes, corev1.Volume{
			Name:         homeVolumeName,
			VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}},
		})
		if len(spec.Containers) > 0 {
			spec.Containers[0].VolumeMounts = append(spec.Containers[0].VolumeMounts, corev1.VolumeMount{
				Name:      homeVolumeName,
				MountPath: agentHome,
			})
		}
	}
}
