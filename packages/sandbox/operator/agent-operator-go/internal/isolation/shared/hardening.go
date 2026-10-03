package shared

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// AgentServiceAccountName is the dedicated zero-RBAC ServiceAccount for agent
	// pods. The agent runs untrusted code and never calls the Kubernetes API.
	AgentServiceAccountName = "agent-runner"
	// agentHome matches the uid-1000 HOME baked into the nix agent image.
	agentHome      = "/home/agent"
	homeVolumeName = "home"
	agentFSGroup   = int64(1000)
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
//   - resource bounds on every init and main container (node-DoS containment);
//   - a writable HOME emptyDir + fsGroup so read-only-root containers can write
//     config, cache, and state without weakening the root filesystem.
//
// The caller must resolve a sandboxed runtimeClassName before building the pod;
// this function never substitutes default runc.
func applyPodHardening(spec *corev1.PodSpec, resources corev1.ResourceRequirements) {
	autoMount := false
	spec.ServiceAccountName = AgentServiceAccountName
	spec.AutomountServiceAccountToken = &autoMount

	if len(resources.Requests) == 0 && len(resources.Limits) == 0 {
		resources = defaultAgentResources()
	}
	if spec.SecurityContext == nil {
		spec.SecurityContext = &corev1.PodSecurityContext{}
	}
	if spec.SecurityContext.FSGroup == nil {
		gid := agentFSGroup
		spec.SecurityContext.FSGroup = &gid
	}
	if !hasVolume(spec.Volumes, homeVolumeName) {
		spec.Volumes = append(spec.Volumes, corev1.Volume{
			Name:         homeVolumeName,
			VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}},
		})
	}
	for index := range spec.InitContainers {
		hardenContainer(&spec.InitContainers[index], resources)
	}
	for index := range spec.Containers {
		hardenContainer(&spec.Containers[index], resources)
	}
}

func hardenContainer(container *corev1.Container, resources corev1.ResourceRequirements) {
	container.Resources = *resources.DeepCopy()
	ensureVolumeMount(container, corev1.VolumeMount{Name: homeVolumeName, MountPath: agentHome})
	ensureVolumeMount(container, corev1.VolumeMount{Name: "tmp", MountPath: "/tmp"})
	setEnvironment(container, "HOME", agentHome)
	setEnvironment(container, "XDG_CONFIG_HOME", agentHome+"/.config")
	setEnvironment(container, "XDG_CACHE_HOME", agentHome+"/.cache")
	setEnvironment(container, "XDG_STATE_HOME", agentHome+"/.local/state")
}

func ensureVolumeMount(container *corev1.Container, mount corev1.VolumeMount) {
	for index := range container.VolumeMounts {
		if container.VolumeMounts[index].Name == mount.Name {
			container.VolumeMounts[index] = mount
			return
		}
	}
	container.VolumeMounts = append(container.VolumeMounts, mount)
}

func setEnvironment(container *corev1.Container, name, value string) {
	for index := range container.Env {
		if container.Env[index].Name == name {
			container.Env[index] = corev1.EnvVar{Name: name, Value: value}
			return
		}
	}
	container.Env = append(container.Env, corev1.EnvVar{Name: name, Value: value})
}

func hasVolume(volumes []corev1.Volume, name string) bool {
	for _, volume := range volumes {
		if volume.Name == name {
			return true
		}
	}
	return false
}
