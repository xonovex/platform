package builder

import (
	"fmt"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
)

const (
	workspaceVolumeName = "workspace"
	worktreeBasePath    = "/workspace-wt"
)

// BuildWorkspacePVC creates a ReadWriteMany PVC for an AgentWorkspace
func BuildWorkspacePVC(name string, ws *agentv1alpha1.AgentWorkspace) *corev1.PersistentVolumeClaim {
	storageSize := ws.Spec.StorageSize
	if storageSize == "" {
		storageSize = "10Gi"
	}

	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ws.Namespace,
			Labels: map[string]string{
				"app.kubernetes.io/name":      "agent-operator",
				"app.kubernetes.io/instance":  ws.Name,
				"app.kubernetes.io/component": "workspace",
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
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteMany,
			},
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse(storageSize),
				},
			},
		},
	}

	if ws.Spec.StorageClass != "" {
		pvc.Spec.StorageClassName = &ws.Spec.StorageClass
	}

	return pvc
}

// BuildSharedVolumePVC creates a ReadWriteMany PVC for a shared volume
func BuildSharedVolumePVC(name string, ws *agentv1alpha1.AgentWorkspace, vol agentv1alpha1.SharedVolumeSpec) *corev1.PersistentVolumeClaim {
	storageSize := vol.StorageSize
	if storageSize == "" {
		storageSize = "1Gi"
	}

	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ws.Namespace,
			Labels: map[string]string{
				"app.kubernetes.io/name":      "agent-operator",
				"app.kubernetes.io/instance":  ws.Name,
				"app.kubernetes.io/component": "shared-volume",
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
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteMany,
			},
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse(storageSize),
				},
			},
		},
	}

	if ws.Spec.StorageClass != "" {
		pvc.Spec.StorageClassName = &ws.Spec.StorageClass
	}

	return pvc
}

// BuildWorkspaceInitJob creates a Job that clones the repository into the workspace PVC
func BuildWorkspaceInitJob(ws *agentv1alpha1.AgentWorkspace, pvcName, image string, runtimeClassName *string) *batchv1.Job {
	activeDeadlineSeconds := int64((10 * time.Minute).Seconds())
	backoffLimit := int32(0)

	script := buildWorkspaceCloneScript(&ws.Spec.Repository, ws.Spec.Type)

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
								{
									Name:      workspaceVolumeName,
									MountPath: workspaceMountPath,
								},
							},
							SecurityContext: DefaultContainerSecurityContext(nil),
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: workspaceVolumeName,
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: pvcName,
								},
							},
						},
					},
				},
			},
		},
	}
}

func buildWorkspaceCloneScript(repo *agentv1alpha1.RepositorySpec, wsType agentv1alpha1.WorkspaceType) string {
	script := "set -e\n"
	script += "cd " + workspaceMountPath + "\n"
	script += "git clone"
	if repo.Branch != "" {
		script += " --branch " + shellQuote(repo.Branch)
	}
	script += " --single-branch --depth 1"
	script += " -- " + shellQuote(repo.URL) + " .\n"

	if repo.Commit != "" {
		script += "git fetch origin " + shellQuote(repo.Commit) + "\n"
		script += "git checkout " + shellQuote(repo.Commit) + "\n"
	}

	if strategy, err := GetVCSStrategy(wsType); err == nil {
		script += strategy.PostCloneScript()
	}

	return script
}

// BuildWorktreeInitContainers builds init containers that create a git worktree (or jj workspace) for an AgentRun
func BuildWorktreeInitContainers(run *agentv1alpha1.AgentRun, image string, wsType agentv1alpha1.WorkspaceType, worktreeBranch, sourceBranch string, sc *corev1.SecurityContext) []corev1.Container {
	worktreePath := fmt.Sprintf("%s/%s", worktreeBasePath, run.Name)
	if sourceBranch == "" {
		sourceBranch = "HEAD"
	}

	script := "set -e\n"
	script += "cd " + workspaceMountPath + "\n"

	containerName := "git-worktree"
	if vcs, err := GetVCSStrategy(wsType); err == nil {
		containerName = vcs.InitContainerName()
		script += vcs.WorktreeScript(worktreePath, worktreeBranch, sourceBranch)
	}

	return []corev1.Container{
		{
			Name:    containerName,
			Image:   image,
			Command: []string{"sh"},
			Args:    []string{"-c", script},
			VolumeMounts: []corev1.VolumeMount{
				{
					Name:      workspaceVolumeName,
					MountPath: workspaceMountPath,
				},
			},
			SecurityContext: DefaultContainerSecurityContext(sc),
		},
	}
}

// BuildWorkspaceMainContainers builds the main agent container for workspace-based runs
func BuildWorkspaceMainContainers(run *agentv1alpha1.AgentRun, providerEnv map[string]string, image string, agentType agentv1alpha1.AgentType, sharedVolumes []agentv1alpha1.SharedVolumeSpec, sharedVolumePVCs map[string]string, tc *agentv1alpha1.ToolchainSpec, sc *corev1.SecurityContext) []corev1.Container {
	env := BuildEnvVars(run, providerEnv)
	command, args := buildAgentCommand(run, agentType)
	worktreePath := fmt.Sprintf("%s/%s", worktreeBasePath, run.Name)

	volumeMounts := []corev1.VolumeMount{
		{
			Name:      workspaceVolumeName,
			MountPath: workspaceMountPath,
		},
	}

	for _, vol := range sharedVolumes {
		if _, ok := sharedVolumePVCs[vol.Name]; ok {
			volumeMounts = append(volumeMounts, corev1.VolumeMount{
				Name:      vol.Name,
				MountPath: vol.MountPath,
			})
		}
	}

	for _, t := range Toolchains(tc) {
		volumeMounts = append(volumeMounts, t.VolumeMounts()...)
		env = append(env, t.EnvVars()...)
	}

	volumeMounts = append(volumeMounts, corev1.VolumeMount{
		Name:      "tmp",
		MountPath: "/tmp",
	})

	return []corev1.Container{
		{
			Name:            "agent",
			Image:           image,
			Command:         command,
			Args:            args,
			Env:             env,
			WorkingDir:      worktreePath,
			VolumeMounts:    volumeMounts,
			SecurityContext: DefaultContainerSecurityContext(sc),
		},
	}
}

// BuildWorkspaceJob creates a Job for an AgentRun that uses a shared workspace
func BuildWorkspaceJob(run *agentv1alpha1.AgentRun, providerEnv map[string]string, workspacePVC string, sharedVolumes []agentv1alpha1.SharedVolumeSpec, sharedVolumePVCs map[string]string, image string, timeout time.Duration, agentType agentv1alpha1.AgentType, wsType agentv1alpha1.WorkspaceType, worktreeBranch, sourceBranch string, tc *agentv1alpha1.ToolchainSpec, ttl *int32) *batchv1.Job {
	activeDeadlineSeconds := int64(timeout.Seconds())

	backoffLimit := int32(0)

	volumes := []corev1.Volume{
		{
			Name: workspaceVolumeName,
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: workspacePVC,
				},
			},
		},
	}

	for _, vol := range sharedVolumes {
		if pvcName, ok := sharedVolumePVCs[vol.Name]; ok {
			volumes = append(volumes, corev1.Volume{
				Name: vol.Name,
				VolumeSource: corev1.VolumeSource{
					PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
						ClaimName: pvcName,
					},
				},
			})
		}
	}

	volumes = append(volumes, corev1.Volume{
		Name:         "tmp",
		VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}},
	})

	initContainers := BuildWorktreeInitContainers(run, image, wsType, worktreeBranch, sourceBranch, run.Spec.SecurityContext)

	for _, t := range Toolchains(tc) {
		if c := t.InitContainer(); c != nil {
			c.SecurityContext = DefaultContainerSecurityContext(run.Spec.SecurityContext)
			initContainers = append(initContainers, *c)
		}
		volumes = append(volumes, t.Volumes()...)
	}

	// TEE runtimeClassName takes precedence over direct RuntimeClassName
	runtimeClass := run.Spec.RuntimeClassName
	if teeRC := TEERuntimeClassName(run.Spec.ConfidentialComputing); teeRC != nil {
		runtimeClass = teeRC
	}

	// Build node affinity from TEE config
	var affinity *corev1.Affinity
	if nodeAffinity := TEENodeAffinity(run.Spec.ConfidentialComputing); nodeAffinity != nil {
		affinity = &corev1.Affinity{NodeAffinity: nodeAffinity}
	}

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      run.Name,
			Namespace: run.Namespace,
			Labels: map[string]string{
				"app.kubernetes.io/name":       "agent-operator",
				"app.kubernetes.io/instance":   run.Name,
				"app.kubernetes.io/component":  "agent-run",
				"agent.xonovex.com/agent-type": string(agentType),
				"agent.xonovex.com/workspace":  run.Spec.WorkspaceRef,
			},
		},
		Spec: batchv1.JobSpec{
			ActiveDeadlineSeconds:   &activeDeadlineSeconds,
			BackoffLimit:            &backoffLimit,
			TTLSecondsAfterFinished: resolveTTL(ttl),
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app.kubernetes.io/name":       "agent-operator",
						"app.kubernetes.io/instance":   run.Name,
						"app.kubernetes.io/component":  "agent-run",
						"agent.xonovex.com/agent-type": string(agentType),
						"agent.xonovex.com/workspace":  run.Spec.WorkspaceRef,
					},
				},
				Spec: corev1.PodSpec{
					RestartPolicy:    corev1.RestartPolicyNever,
					SecurityContext:  DefaultPodSecurityContext(run.Spec.PodSecurityContext),
					InitContainers:   initContainers,
					Containers:       BuildWorkspaceMainContainers(run, providerEnv, image, agentType, sharedVolumes, sharedVolumePVCs, tc, run.Spec.SecurityContext),
					Volumes:          volumes,
					NodeSelector:     run.Spec.NodeSelector,
					Tolerations:      run.Spec.Tolerations,
					RuntimeClassName: runtimeClass,
					Affinity:         affinity,
				},
			},
		},
	}

	if len(run.Spec.Resources.Requests) > 0 || len(run.Spec.Resources.Limits) > 0 {
		job.Spec.Template.Spec.Containers[0].Resources = run.Spec.Resources
	}

	return job
}
