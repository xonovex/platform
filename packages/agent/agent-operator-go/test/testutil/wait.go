package testutil

import (
	"context"
	"fmt"
	"testing"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
)

// WaitForCondition polls fn every 250ms until it returns true or the timeout expires.
// Logs progress every 30 seconds to help diagnose slow waits.
func WaitForCondition(t *testing.T, ctx context.Context, timeout time.Duration, fn func() bool) {
	t.Helper()

	start := time.Now()
	deadline := time.After(timeout)
	tick := time.NewTicker(250 * time.Millisecond)
	defer tick.Stop()

	lastLog := start

	for {
		select {
		case <-deadline:
			t.Fatalf("timed out waiting for condition after %v", timeout)
		case <-ctx.Done():
			t.Fatal("context cancelled waiting for condition")
		case <-tick.C:
			if fn() {
				return
			}
			if time.Since(lastLog) >= 30*time.Second {
				t.Logf("still waiting for condition (%v elapsed of %v timeout)", time.Since(start).Round(time.Second), timeout)
				lastLog = time.Now()
			}
		}
	}
}

// WaitForAgentRunPhase waits for an AgentRun to reach the given phase.
// Logs the current phase and related Job/Pod state every 30 seconds for diagnostics.
func WaitForAgentRunPhase(t *testing.T, ctx context.Context, c client.Client, key client.ObjectKey, phase agentv1alpha1.AgentRunPhase, timeout time.Duration) {
	t.Helper()

	start := time.Now()
	lastLog := start
	WaitForCondition(t, ctx, timeout, func() bool {
		var run agentv1alpha1.AgentRun
		if err := c.Get(ctx, key, &run); err != nil {
			return false
		}
		if run.Status.Phase == phase {
			return true
		}
		if time.Since(lastLog) >= 30*time.Second {
			msg := fmt.Sprintf("AgentRun %s: phase=%s (want %s)", key.Name, run.Status.Phase, phase)
			if run.Status.JobName != "" {
				var job batchv1.Job
				if err := c.Get(ctx, client.ObjectKey{Name: run.Status.JobName, Namespace: key.Namespace}, &job); err == nil {
					msg += fmt.Sprintf(", job active=%d succeeded=%d failed=%d", job.Status.Active, job.Status.Succeeded, job.Status.Failed)
					for _, cond := range job.Status.Conditions {
						msg += fmt.Sprintf(", jobCond=%s/%s", cond.Type, cond.Status)
					}
				}
				var podList corev1.PodList
				if err := c.List(ctx, &podList, client.InNamespace(key.Namespace), client.MatchingLabels{
					"app.kubernetes.io/instance": key.Name,
				}); err == nil && len(podList.Items) > 0 {
					pod := podList.Items[0]
					msg += fmt.Sprintf(", pod=%s", pod.Status.Phase)
					for _, cs := range pod.Status.InitContainerStatuses {
						if cs.State.Waiting != nil {
							msg += fmt.Sprintf(", init=%s(%s)", cs.Name, cs.State.Waiting.Reason)
						} else if cs.State.Terminated != nil {
							msg += fmt.Sprintf(", init=%s(exit=%d)", cs.Name, cs.State.Terminated.ExitCode)
						}
					}
					for _, cs := range pod.Status.ContainerStatuses {
						if cs.State.Waiting != nil {
							msg += fmt.Sprintf(", container=%s(%s)", cs.Name, cs.State.Waiting.Reason)
						} else if cs.State.Terminated != nil {
							msg += fmt.Sprintf(", container=%s(exit=%d)", cs.Name, cs.State.Terminated.ExitCode)
						}
					}
				}
			}
			t.Log(msg)
			lastLog = time.Now()
		}
		return false
	})
}

// WaitForWorkspacePhase waits for an AgentWorkspace to reach the given phase.
func WaitForWorkspacePhase(t *testing.T, ctx context.Context, c client.Client, key client.ObjectKey, phase agentv1alpha1.AgentWorkspacePhase, timeout time.Duration) {
	t.Helper()

	WaitForCondition(t, ctx, timeout, func() bool {
		var ws agentv1alpha1.AgentWorkspace
		if err := c.Get(ctx, key, &ws); err != nil {
			return false
		}
		return ws.Status.Phase == phase
	})
}
