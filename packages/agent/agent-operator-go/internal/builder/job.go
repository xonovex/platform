package builder

import (
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
)

// BuildJob creates a Kubernetes Job for an AgentRun
func BuildJob(run *agentv1alpha1.AgentRun, providerEnv map[string]string, pvcName string, image string, timeout time.Duration, agentType agentv1alpha1.AgentType, wsType agentv1alpha1.WorkspaceType, tc *agentv1alpha1.ToolchainSpec) *batchv1.Job {
	activeDeadlineSeconds := int64(timeout.Seconds())

	backoffLimit := int32(0)

	volumes := []corev1.Volume{
		{
			Name: "workspace",
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: pvcName,
				},
			},
		},
		{
			Name:         "tmp",
			VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}},
		},
	}

	for _, t := range Toolchains(tc) {
		volumes = append(volumes, t.Volumes()...)
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
			},
		},
		Spec: batchv1.JobSpec{
			ActiveDeadlineSeconds: &activeDeadlineSeconds,
			BackoffLimit:          &backoffLimit,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app.kubernetes.io/name":       "agent-operator",
						"app.kubernetes.io/instance":   run.Name,
						"app.kubernetes.io/component":  "agent-run",
						"agent.xonovex.com/agent-type": string(agentType),
					},
				},
				Spec: corev1.PodSpec{
					RestartPolicy:    corev1.RestartPolicyNever,
					SecurityContext:  DefaultPodSecurityContext(run.Spec.PodSecurityContext),
					InitContainers:   BuildInitContainers(run, image, wsType, tc, run.Spec.SecurityContext),
					Containers:       BuildMainContainers(run, providerEnv, image, agentType, tc, run.Spec.SecurityContext),
					Volumes:          volumes,
					NodeSelector:     run.Spec.NodeSelector,
					Tolerations:      run.Spec.Tolerations,
					RuntimeClassName: run.Spec.RuntimeClassName,
				},
			},
		},
	}

	if len(run.Spec.Resources.Requests) > 0 || len(run.Spec.Resources.Limits) > 0 {
		job.Spec.Template.Spec.Containers[0].Resources = run.Spec.Resources
	}

	return job
}
