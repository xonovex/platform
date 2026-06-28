package shared

import (
	"fmt"
	"maps"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
	"github.com/xonovex/platform/packages/agent/agent-operator-go/internal/plugins"
	wsshared "github.com/xonovex/platform/packages/agent/agent-operator-go/internal/workspace/shared"
)

// resolveTTL returns the TTL pointer to use, defaulting to 3600 if nil.
func resolveTTL(ttl *int32) *int32 {
	if ttl != nil {
		return ttl
	}
	defaultTTL := int32(3600)
	return &defaultTTL
}

// WorkspaceBinding is the workspace diff fed into the unified Job builder. A nil
// binding produces the standalone Job (clone into the run's own PVC); a non-nil
// binding produces the workspace-based Job (worktree into a shared workspace PVC,
// shared volumes, and the workspace label).
type WorkspaceBinding struct {
	SharedVolumes    []agentv1alpha1.SharedVolumeSpec
	SharedVolumePVCs map[string]string
	WorktreeBranch   string
	SourceBranch     string
	WorkspaceRef     string
}

// pvcVolume returns a PVC-backed volume source for claimName.
func pvcVolume(name, claimName string) corev1.Volume {
	return corev1.Volume{
		Name: name,
		VolumeSource: corev1.VolumeSource{
			PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{ClaimName: claimName},
		},
	}
}

// tmpVolume returns the emptyDir /tmp volume.
func tmpVolume() corev1.Volume {
	return corev1.Volume{Name: "tmp", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}}
}

// BuildJob creates the Kubernetes Job for an AgentRun. It is the single pod
// realizer: a nil ws produces the standalone Job, a non-nil ws produces the
// workspace-based Job. pvcName is the workspace PVC the pod mounts.
func BuildJob(run *agentv1alpha1.AgentRun, providerEnv map[string]string, pvcName, image string, timeout time.Duration, agentType agentv1alpha1.AgentType, wsType agentv1alpha1.WorkspaceType, tc *agentv1alpha1.ToolchainSpec, ttl *int32, ws *WorkspaceBinding) *batchv1.Job {
	activeDeadlineSeconds := int64(timeout.Seconds())
	backoffLimit := int32(0)

	labels := map[string]string{
		"app.kubernetes.io/name":       "agent-operator",
		"app.kubernetes.io/instance":   run.Name,
		"app.kubernetes.io/component":  "agent-run",
		"agent.xonovex.com/agent-type": string(agentType),
	}

	volumes := []corev1.Volume{pvcVolume(wsshared.WorkspaceVolumeName, pvcName)}

	var initContainers, mainContainers []corev1.Container
	if ws != nil {
		labels["agent.xonovex.com/workspace"] = ws.WorkspaceRef
		for _, vol := range ws.SharedVolumes {
			if sharedPVC, ok := ws.SharedVolumePVCs[vol.Name]; ok {
				volumes = append(volumes, pvcVolume(vol.Name, sharedPVC))
			}
		}
		volumes = append(volumes, tmpVolume())
		initContainers = BuildWorktreeInitContainers(run, image, wsType, ws.WorktreeBranch, ws.SourceBranch, run.Spec.SecurityContext)
		mainContainers = BuildWorkspaceMainContainers(run, providerEnv, image, agentType, ws.SharedVolumes, ws.SharedVolumePVCs, run.Spec.SecurityContext)
	} else {
		volumes = append(volumes, tmpVolume())
		initContainers = BuildInitContainers(run, image, wsType, run.Spec.SecurityContext)
		mainContainers = BuildMainContainers(run, providerEnv, image, agentType, run.Spec.SecurityContext)
	}

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      run.Name,
			Namespace: run.Namespace,
			Labels:    labels,
		},
		Spec: batchv1.JobSpec{
			ActiveDeadlineSeconds:   &activeDeadlineSeconds,
			BackoffLimit:            &backoffLimit,
			TTLSecondsAfterFinished: resolveTTL(ttl),
			Template: corev1.PodTemplateSpec{
				// Clone so the pod-template labels are independent of the Job's; a
				// later mutation of one must not corrupt the other's identity/selector.
				ObjectMeta: metav1.ObjectMeta{Labels: maps.Clone(labels)},
				Spec: corev1.PodSpec{
					RestartPolicy:    corev1.RestartPolicyNever,
					SecurityContext:  DefaultPodSecurityContext(run.Spec.PodSecurityContext),
					InitContainers:   initContainers,
					Containers:       mainContainers,
					Volumes:          volumes,
					NodeSelector:     run.Spec.NodeSelector,
					Tolerations:      run.Spec.Tolerations,
					RuntimeClassName: run.Spec.RuntimeClassName,
				},
			},
		},
	}

	applyPodHardening(&job.Spec.Template.Spec, run, tc)

	return job
}

// BuildWorkspaceInitJob creates the Job that clones the repository into the
// workspace PVC (run by the AgentWorkspace controller).
func BuildWorkspaceInitJob(ws *agentv1alpha1.AgentWorkspace, pvcName, image string, runtimeClassName *string) *batchv1.Job {
	activeDeadlineSeconds := int64((10 * time.Minute).Seconds())
	backoffLimit := int32(0)

	strategy, _ := plugins.GetVCSStrategy(ws.Spec.Type)
	script := wsshared.CloneScript(ws.Spec.Repository, strategy)

	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-init", ws.Name),
			Namespace: ws.Namespace,
			Labels: map[string]string{
				"app.kubernetes.io/name":      "agent-operator",
				"app.kubernetes.io/instance":  ws.Name,
				"app.kubernetes.io/component": "workspace-init",
			},
		},
		Spec: batchv1.JobSpec{
			ActiveDeadlineSeconds: &activeDeadlineSeconds,
			BackoffLimit:          &backoffLimit,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app.kubernetes.io/name":      "agent-operator",
						"app.kubernetes.io/instance":  ws.Name,
						"app.kubernetes.io/component": "workspace-init",
					},
				},
				Spec: corev1.PodSpec{
					RestartPolicy:    corev1.RestartPolicyNever,
					SecurityContext:  DefaultPodSecurityContext(nil),
					RuntimeClassName: runtimeClassName,
					Containers: []corev1.Container{
						{
							Name:    "git-clone",
							Image:   image,
							Command: []string{"sh"},
							Args:    []string{"-c", script},
							VolumeMounts: []corev1.VolumeMount{
								{Name: wsshared.WorkspaceVolumeName, MountPath: wsshared.WorkspaceMountPath},
							},
							SecurityContext: DefaultContainerSecurityContext(nil),
						},
					},
					Volumes: []corev1.Volume{pvcVolume(wsshared.WorkspaceVolumeName, pvcName)},
				},
			},
		},
	}
}
