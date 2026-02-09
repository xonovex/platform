package testutil

import (
	"context"
	"testing"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/client"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
)

// WaitForCondition polls fn every 250ms until it returns true or the timeout expires.
func WaitForCondition(t *testing.T, ctx context.Context, timeout time.Duration, fn func() bool) {
	t.Helper()

	deadline := time.After(timeout)
	tick := time.NewTicker(250 * time.Millisecond)
	defer tick.Stop()

	for {
		select {
		case <-deadline:
			t.Fatal("timed out waiting for condition")
		case <-ctx.Done():
			t.Fatal("context cancelled waiting for condition")
		case <-tick.C:
			if fn() {
				return
			}
		}
	}
}

// WaitForAgentRunPhase waits for an AgentRun to reach the given phase.
func WaitForAgentRunPhase(t *testing.T, ctx context.Context, c client.Client, key client.ObjectKey, phase agentv1alpha1.AgentRunPhase, timeout time.Duration) {
	t.Helper()

	WaitForCondition(t, ctx, timeout, func() bool {
		var run agentv1alpha1.AgentRun
		if err := c.Get(ctx, key, &run); err != nil {
			return false
		}
		return run.Status.Phase == phase
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
