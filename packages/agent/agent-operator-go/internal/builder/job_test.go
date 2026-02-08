package builder

import (
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
)

func TestBuildJob_Basic(t *testing.T) {
	run := &agentv1alpha1.AgentRun{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-run",
			Namespace: "default",
		},
		Spec: agentv1alpha1.AgentRunSpec{
			Agent: agentv1alpha1.AgentTypeClaude,
			Repository: agentv1alpha1.RepositorySpec{
				URL: "https://github.com/example/repo.git",
			},
		},
	}

	job := BuildJob(run, nil, "test-pvc", "node:trixie-slim", time.Hour)

	if job.Name != "test-run" {
		t.Errorf("job name = %q, want %q", job.Name, "test-run")
	}
	if job.Namespace != "default" {
		t.Errorf("job namespace = %q, want %q", job.Namespace, "default")
	}
	if job.Labels["agent.xonovex.com/agent-type"] != "claude" {
		t.Errorf("agent-type label = %q, want %q", job.Labels["agent.xonovex.com/agent-type"], "claude")
	}

	expectedTimeout := int64(3600)
	if *job.Spec.ActiveDeadlineSeconds != expectedTimeout {
		t.Errorf("ActiveDeadlineSeconds = %d, want %d", *job.Spec.ActiveDeadlineSeconds, expectedTimeout)
	}
	if *job.Spec.BackoffLimit != 0 {
		t.Errorf("BackoffLimit = %d, want 0", *job.Spec.BackoffLimit)
	}

	podSpec := job.Spec.Template.Spec
	if podSpec.RestartPolicy != corev1.RestartPolicyNever {
		t.Errorf("RestartPolicy = %q, want %q", podSpec.RestartPolicy, corev1.RestartPolicyNever)
	}
	if len(podSpec.InitContainers) != 1 {
		t.Fatalf("len(InitContainers) = %d, want 1", len(podSpec.InitContainers))
	}
	if len(podSpec.Containers) != 1 {
		t.Fatalf("len(Containers) = %d, want 1", len(podSpec.Containers))
	}
	if podSpec.Volumes[0].PersistentVolumeClaim.ClaimName != "test-pvc" {
		t.Errorf("PVC claim = %q, want %q", podSpec.Volumes[0].PersistentVolumeClaim.ClaimName, "test-pvc")
	}
}

func TestBuildJob_CustomTimeout(t *testing.T) {
	customTimeout := metav1.Duration{Duration: 30 * time.Minute}
	run := &agentv1alpha1.AgentRun{
		ObjectMeta: metav1.ObjectMeta{Name: "test-run", Namespace: "default"},
		Spec: agentv1alpha1.AgentRunSpec{
			Agent:      agentv1alpha1.AgentTypeClaude,
			Repository: agentv1alpha1.RepositorySpec{URL: "https://example.com/repo.git"},
			Timeout:    &customTimeout,
		},
	}

	job := BuildJob(run, nil, "pvc", "image", time.Hour)

	expected := int64(1800)
	if *job.Spec.ActiveDeadlineSeconds != expected {
		t.Errorf("ActiveDeadlineSeconds = %d, want %d", *job.Spec.ActiveDeadlineSeconds, expected)
	}
}

func TestBuildJob_CustomImage(t *testing.T) {
	run := &agentv1alpha1.AgentRun{
		ObjectMeta: metav1.ObjectMeta{Name: "test-run", Namespace: "default"},
		Spec: agentv1alpha1.AgentRunSpec{
			Agent:      agentv1alpha1.AgentTypeClaude,
			Repository: agentv1alpha1.RepositorySpec{URL: "https://example.com/repo.git"},
			Image:      "custom-image:latest",
		},
	}

	job := BuildJob(run, nil, "pvc", "default-image", time.Hour)

	initImage := job.Spec.Template.Spec.InitContainers[0].Image
	if initImage != "custom-image:latest" {
		t.Errorf("init container image = %q, want %q", initImage, "custom-image:latest")
	}
	mainImage := job.Spec.Template.Spec.Containers[0].Image
	if mainImage != "custom-image:latest" {
		t.Errorf("main container image = %q, want %q", mainImage, "custom-image:latest")
	}
}

func TestBuildJob_WithResources(t *testing.T) {
	run := &agentv1alpha1.AgentRun{
		ObjectMeta: metav1.ObjectMeta{Name: "test-run", Namespace: "default"},
		Spec: agentv1alpha1.AgentRunSpec{
			Agent:      agentv1alpha1.AgentTypeClaude,
			Repository: agentv1alpha1.RepositorySpec{URL: "https://example.com/repo.git"},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("500m"),
					corev1.ResourceMemory: resource.MustParse("1Gi"),
				},
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("2"),
					corev1.ResourceMemory: resource.MustParse("4Gi"),
				},
			},
		},
	}

	job := BuildJob(run, nil, "pvc", "image", time.Hour)

	resources := job.Spec.Template.Spec.Containers[0].Resources
	if resources.Requests.Cpu().String() != "500m" {
		t.Errorf("CPU request = %q, want %q", resources.Requests.Cpu().String(), "500m")
	}
	if resources.Limits.Memory().String() != "4Gi" {
		t.Errorf("memory limit = %q, want %q", resources.Limits.Memory().String(), "4Gi")
	}
}

func TestBuildJob_NodeSelectorAndTolerations(t *testing.T) {
	run := &agentv1alpha1.AgentRun{
		ObjectMeta: metav1.ObjectMeta{Name: "test-run", Namespace: "default"},
		Spec: agentv1alpha1.AgentRunSpec{
			Agent:        agentv1alpha1.AgentTypeClaude,
			Repository:   agentv1alpha1.RepositorySpec{URL: "https://example.com/repo.git"},
			NodeSelector: map[string]string{"gpu": "true"},
			Tolerations: []corev1.Toleration{
				{Key: "gpu", Operator: corev1.TolerationOpExists, Effect: corev1.TaintEffectNoSchedule},
			},
		},
	}

	job := BuildJob(run, nil, "pvc", "image", time.Hour)

	podSpec := job.Spec.Template.Spec
	if podSpec.NodeSelector["gpu"] != "true" {
		t.Errorf("NodeSelector[gpu] = %q, want %q", podSpec.NodeSelector["gpu"], "true")
	}
	if len(podSpec.Tolerations) != 1 {
		t.Fatalf("len(Tolerations) = %d, want 1", len(podSpec.Tolerations))
	}
	if podSpec.Tolerations[0].Key != "gpu" {
		t.Errorf("Toleration key = %q, want %q", podSpec.Tolerations[0].Key, "gpu")
	}
}

func TestBuildJob_Labels(t *testing.T) {
	run := &agentv1alpha1.AgentRun{
		ObjectMeta: metav1.ObjectMeta{Name: "my-run", Namespace: "ns"},
		Spec: agentv1alpha1.AgentRunSpec{
			Agent:      agentv1alpha1.AgentTypeOpencode,
			Repository: agentv1alpha1.RepositorySpec{URL: "https://example.com/repo.git"},
		},
	}

	job := BuildJob(run, nil, "pvc", "image", time.Hour)

	expectedLabels := map[string]string{
		"app.kubernetes.io/name":      "agent-operator",
		"app.kubernetes.io/instance":  "my-run",
		"app.kubernetes.io/component": "agent-run",
		"agent.xonovex.com/agent-type": "opencode",
	}
	for k, want := range expectedLabels {
		if got := job.Labels[k]; got != want {
			t.Errorf("job label %q = %q, want %q", k, got, want)
		}
		if got := job.Spec.Template.Labels[k]; got != want {
			t.Errorf("pod template label %q = %q, want %q", k, got, want)
		}
	}
}
