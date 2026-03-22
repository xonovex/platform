//go:build e2e

package e2e

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/xonovex/platform/packages/agent/agent-operator-go/test/testutil"
)

func TestE2E_ImageDeployment(t *testing.T) {
	if useExistingCluster {
		t.Skip("image deployment test requires Kind (skipped with USE_EXISTING_CLUSTER=true)")
	}

	workspaceRoot := gitWorkspaceRoot(t)
	operatorDir := filepath.Join(workspaceRoot, "packages", "agent", "agent-operator-go")

	// Build image
	dockerfile := filepath.Join(operatorDir, "Dockerfile")
	buildCmd := exec.Command("docker", "build", "-f", dockerfile, "-t", "agent-operator:latest", ".")
	buildCmd.Dir = workspaceRoot
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("docker build failed: %v", err)
	}

	// Retry kind load — containerd may not be ready immediately after cluster creation
	var loadErr error
	for range 5 {
		loadErr = runCmd("kind", "load", "docker-image", "agent-operator:latest", "--name", clusterName)
		if loadErr == nil {
			break
		}
		time.Sleep(3 * time.Second)
	}
	if loadErr != nil {
		t.Fatalf("kind load failed after retries: %v", loadErr)
	}

	ns := createNamespace(t, "e2e-image")

	// Create ServiceAccount
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{Name: "agent-operator", Namespace: ns},
	}
	if err := k8sClient.Create(ctx, sa); err != nil {
		t.Fatalf("failed to create ServiceAccount: %v", err)
	}

	// Create ClusterRoleBinding (test-only: grant cluster-admin)
	crb := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{Name: "agent-operator-e2e-image-" + ns},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     "cluster-admin",
		},
		Subjects: []rbacv1.Subject{{
			Kind:      "ServiceAccount",
			Name:      "agent-operator",
			Namespace: ns,
		}},
	}
	if err := k8sClient.Create(ctx, crb); err != nil {
		t.Fatalf("failed to create ClusterRoleBinding: %v", err)
	}
	t.Cleanup(func() {
		_ = k8sClient.Delete(ctx, crb)
	})

	// Create Deployment matching config/manager/manager.yaml
	replicas := int32(1)
	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: "agent-operator", Namespace: ns},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "agent-operator"},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": "agent-operator"},
				},
				Spec: corev1.PodSpec{
					ServiceAccountName:            "agent-operator",
					TerminationGracePeriodSeconds: ptr(int64(10)),
					Containers: []corev1.Container{{
						Name:            "manager",
						Image:           "agent-operator:latest",
						ImagePullPolicy: corev1.PullNever,
						Command:         []string{"/operator"},
						Args:            []string{"--health-probe-bind-address=:8081"},
						Ports: []corev1.ContainerPort{{
							ContainerPort: 8081,
							Name:          "health",
							Protocol:      corev1.ProtocolTCP,
						}},
						LivenessProbe: &corev1.Probe{
							ProbeHandler: corev1.ProbeHandler{
								HTTPGet: &corev1.HTTPGetAction{
									Path: "/healthz",
									Port: intstr.FromInt32(8081),
								},
							},
							InitialDelaySeconds: 5,
							PeriodSeconds:       10,
						},
						ReadinessProbe: &corev1.Probe{
							ProbeHandler: corev1.ProbeHandler{
								HTTPGet: &corev1.HTTPGetAction{
									Path: "/readyz",
									Port: intstr.FromInt32(8081),
								},
							},
							InitialDelaySeconds: 3,
							PeriodSeconds:       5,
						},
						SecurityContext: &corev1.SecurityContext{
							AllowPrivilegeEscalation: ptr(false),
							Capabilities: &corev1.Capabilities{
								Drop: []corev1.Capability{"ALL"},
							},
						},
					}},
					SecurityContext: &corev1.PodSecurityContext{
						RunAsNonRoot: ptr(true),
					},
				},
			},
		},
	}
	if err := k8sClient.Create(ctx, dep); err != nil {
		t.Fatalf("failed to create Deployment: %v", err)
	}

	// Wait for pod to be running and ready (proves binary starts and health probes pass)
	lastLog := time.Now()
	testutil.WaitForCondition(t, ctx, 180*time.Second, func() bool {
		var d appsv1.Deployment
		if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(dep), &d); err != nil {
			return false
		}
		if d.Status.ReadyReplicas == 1 {
			return true
		}
		if time.Since(lastLog) >= 15*time.Second {
			var podList corev1.PodList
			if err := k8sClient.List(ctx, &podList, client.InNamespace(ns), client.MatchingLabels{"app": "agent-operator"}); err == nil {
				for _, pod := range podList.Items {
					msg := fmt.Sprintf("pod=%s phase=%s", pod.Name, pod.Status.Phase)
					for _, cs := range pod.Status.ContainerStatuses {
						if cs.State.Waiting != nil {
							msg += fmt.Sprintf(" container=%s(%s: %s)", cs.Name, cs.State.Waiting.Reason, cs.State.Waiting.Message)
						} else if cs.State.Terminated != nil {
							msg += fmt.Sprintf(" container=%s(exit=%d: %s)", cs.Name, cs.State.Terminated.ExitCode, cs.State.Terminated.Message)
						} else if cs.State.Running != nil {
							msg += fmt.Sprintf(" container=%s(running, ready=%v)", cs.Name, cs.Ready)
						}
					}
					for _, cond := range pod.Status.Conditions {
						if cond.Status != corev1.ConditionTrue {
							msg += fmt.Sprintf(" cond=%s(%s)", cond.Type, cond.Message)
						}
					}
					t.Log(msg)
				}
			}
			lastLog = time.Now()
		}
		return false
	})

	// Verify pod details
	var podList corev1.PodList
	if err := k8sClient.List(ctx, &podList, client.InNamespace(ns), client.MatchingLabels{
		"app": "agent-operator",
	}); err != nil {
		t.Fatalf("failed to list pods: %v", err)
	}
	if len(podList.Items) == 0 {
		t.Fatal("no operator pods found")
	}

	pod := podList.Items[0]
	if pod.Status.Phase != corev1.PodRunning {
		t.Errorf("pod phase = %q, want Running", pod.Status.Phase)
	}
	for _, cs := range pod.Status.ContainerStatuses {
		if !cs.Ready {
			t.Errorf("container %q not ready", cs.Name)
		}
	}
}

func gitWorkspaceRoot(t *testing.T) string {
	t.Helper()
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("failed to find workspace root: %v", err)
	}
	return filepath.Clean(string(out[:len(out)-1])) // trim trailing newline
}

func ptr[T any](v T) *T {
	return &v
}
