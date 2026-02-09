package builder

import (
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
)

// BuildJob creates a Kubernetes Job for an AgentRun
func BuildJob(run *agentv1alpha1.AgentRun, providerEnv map[string]string, pvcName string, defaultImage string, defaultTimeout time.Duration) *batchv1.Job {
	timeout := defaultTimeout
	if run.Spec.Timeout != nil {
		timeout = run.Spec.Timeout.Duration
	}
	activeDeadlineSeconds := int64(timeout.Seconds())

	image := defaultImage
	if run.Spec.Image != "" {
		image = run.Spec.Image
	}

	backoffLimit := int32(0)

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      run.Name,
			Namespace: run.Namespace,
			Labels: map[string]string{
				"app.kubernetes.io/name":       "agent-operator",
				"app.kubernetes.io/instance":   run.Name,
				"app.kubernetes.io/component":  "agent-run",
				"agent.xonovex.com/agent-type": string(run.Spec.Agent),
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
						"agent.xonovex.com/agent-type": string(run.Spec.Agent),
					},
				},
				Spec: corev1.PodSpec{
					RestartPolicy:  corev1.RestartPolicyNever,
					InitContainers: BuildInitContainers(run, image),
					Containers:     BuildMainContainers(run, providerEnv, image),
					Volumes: []corev1.Volume{
						{
							Name: "workspace",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: pvcName,
								},
							},
						},
					},
					NodeSelector: run.Spec.NodeSelector,
					Tolerations:  run.Spec.Tolerations,
				},
			},
		},
	}

	if len(run.Spec.Resources.Requests) > 0 || len(run.Spec.Resources.Limits) > 0 {
		job.Spec.Template.Spec.Containers[0].Resources = run.Spec.Resources
	}

	return job
}
