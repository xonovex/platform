package shared

import (
	"fmt"
	"maps"
	"math"
	"strings"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	agentv1alpha1 "github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/api/v1alpha1"
	harnessshared "github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/internal/harness/shared"
	wsshared "github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/internal/workspace/shared"
)

// JobPolicy holds the operator settings of every Job the isolation core builds.
type JobPolicy struct {
	// DefaultTTLSecondsAfterFinished applies to a run Job whose run sets no TTL.
	DefaultTTLSecondsAfterFinished int32
	// WorkspaceInitTimeout bounds the Job that clones a shared workspace.
	WorkspaceInitTimeout time.Duration
}

// DefaultJobPolicy returns the JobPolicy the operator uses when no flag
// overrides it.
func DefaultJobPolicy() JobPolicy {
	return JobPolicy{
		DefaultTTLSecondsAfterFinished: 3600,
		WorkspaceInitTimeout:           10 * time.Minute,
	}
}

// resolveTTL returns the run TTL, or the policy default when the run sets none.
func resolveTTL(ttl *int32, policy JobPolicy) *int32 {
	if ttl != nil {
		return ttl
	}
	defaultTTL := policy.DefaultTTLSecondsAfterFinished
	return &defaultTTL
}

// RunJob is the resolved input of the agent Job of one AgentRun. Harness and VCS
// are the leaves the composition root resolved for the run's agent and
// workspace types.
type RunJob struct {
	ProviderEnv     []corev1.EnvVar
	ProviderCliArgs []string
	PVCName         string
	Image           string
	Timeout         time.Duration
	AgentType       agentv1alpha1.AgentType
	Harness         harnessshared.CommandBuilder
	VCS             wsshared.VCSStrategy
	TTL             *int32
	Workspace       *WorkspaceBinding
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

const repositoryCredentialsVolumeName = "repository-credentials"

func repositoryCredentialsVolume(ref *agentv1alpha1.SecretKeyRef) corev1.Volume {
	mode := int32(0440)
	return corev1.Volume{
		Name: repositoryCredentialsVolumeName,
		VolumeSource: corev1.VolumeSource{Secret: &corev1.SecretVolumeSource{
			SecretName: ref.Name,
			Items: []corev1.KeyToPath{{
				Key:  ref.Key,
				Path: "credentials",
				Mode: &mode,
			}},
		}},
	}
}

// BuildJob creates the Kubernetes Job for an AgentRun. It is the single pod
// realizer: a nil input.Workspace produces the standalone Job, a non-nil one
// produces the workspace-based Job. input.PVCName is the workspace PVC the pod
// mounts.
func BuildJob(run *agentv1alpha1.AgentRun, input RunJob, policy JobPolicy) (*batchv1.Job, error) {
	if input.Timeout <= 0 {
		return nil, fmt.Errorf("timeout must be greater than zero")
	}
	activeDeadlineSeconds := int64(math.Ceil(input.Timeout.Seconds()))
	backoffLimit := int32(0)

	labels := map[string]string{
		"app.kubernetes.io/name":       "agent-operator",
		"app.kubernetes.io/instance":   run.Name,
		"app.kubernetes.io/component":  "agent-run",
		"agent.xonovex.com/agent-type": string(input.AgentType),
	}

	volumes := []corev1.Volume{pvcVolume(wsshared.WorkspaceVolumeName, input.PVCName)}

	var initContainers, mainContainers []corev1.Container
	if ws := input.Workspace; ws != nil {
		labels["agent.xonovex.com/workspace"] = ws.WorkspaceRef
		for _, vol := range ws.SharedVolumes {
			if sharedPVC, ok := ws.SharedVolumePVCs[vol.Name]; ok {
				volumes = append(volumes, pvcVolume(vol.Name, sharedPVC))
			}
		}
		volumes = append(volumes, tmpVolume())
		initContainers = buildWorktreeInitContainers(run, input.Image, input.VCS, ws.WorktreeBranch, ws.SourceBranch, run.Spec.SecurityContext)
		var err error
		mainContainers, err = buildWorkspaceMainContainers(run, input.ProviderEnv, input.ProviderCliArgs, input.Image, input.Harness, ws.SharedVolumes, ws.SharedVolumePVCs, run.Spec.SecurityContext)
		if err != nil {
			return nil, err
		}
	} else {
		volumes = append(volumes, tmpVolume())
		if run.Spec.Workspace != nil && run.Spec.Workspace.Repository.CredentialsSecretRef != nil {
			volumes = append(volumes, repositoryCredentialsVolume(run.Spec.Workspace.Repository.CredentialsSecretRef))
		}
		initContainers = buildInitContainers(run, input.Image, input.VCS, run.Spec.SecurityContext)
		var err error
		mainContainers, err = buildMainContainers(run, input.ProviderEnv, input.ProviderCliArgs, input.Image, input.Harness, run.Spec.SecurityContext)
		if err != nil {
			return nil, err
		}
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
			TTLSecondsAfterFinished: resolveTTL(input.TTL, policy),
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

	applyPodHardening(&job.Spec.Template.Spec, run.Spec.Resources)

	return job, nil
}

// BuildWorkspaceInitJob creates the Job that clones the repository into the
// workspace PVC (run by the AgentWorkspace controller). strategy is the VCS leaf
// of the workspace type.
func BuildWorkspaceInitJob(ws *agentv1alpha1.AgentWorkspace, pvcName, image string, runtimeClassName *string, strategy wsshared.VCSStrategy, policy JobPolicy) (*batchv1.Job, error) {
	if runtimeClassName == nil || strings.TrimSpace(*runtimeClassName) == "" {
		return nil, fmt.Errorf("workspace initialization requires a sandboxed runtimeClassName")
	}
	if policy.WorkspaceInitTimeout <= 0 {
		return nil, fmt.Errorf("workspace initialization timeout must be greater than zero")
	}
	activeDeadlineSeconds := int64(math.Ceil(policy.WorkspaceInitTimeout.Seconds()))
	backoffLimit := int32(0)

	script := wsshared.CloneScript(ws.Spec.Repository, strategy)

	job := &batchv1.Job{
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
					Volumes: []corev1.Volume{
						pvcVolume(wsshared.WorkspaceVolumeName, pvcName),
						tmpVolume(),
					},
				},
			},
		},
	}
	if ref := ws.Spec.Repository.CredentialsSecretRef; ref != nil {
		job.Spec.Template.Spec.Volumes = append(job.Spec.Template.Spec.Volumes, repositoryCredentialsVolume(ref))
		job.Spec.Template.Spec.Containers[0].VolumeMounts = append(job.Spec.Template.Spec.Containers[0].VolumeMounts, corev1.VolumeMount{
			Name:      repositoryCredentialsVolumeName,
			MountPath: wsshared.RepositoryCredentialsMountPath,
			ReadOnly:  true,
		})
	}
	applyPodHardening(&job.Spec.Template.Spec, corev1.ResourceRequirements{})
	return job, nil
}
