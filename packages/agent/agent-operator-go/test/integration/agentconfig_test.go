//go:build integration

package integration

import (
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
	"github.com/xonovex/platform/packages/agent/agent-operator-go/test/testutil"
)

func TestAgentConfig_BecomesReady(t *testing.T) {
	ns := createNamespace(t, "config-ready")

	config := testutil.NewAgentConfig(ns, "agent-config",
		testutil.WithDefaultAgent(agentv1alpha1.AgentTypeClaude),
		testutil.WithStorageSize("10Gi"),
	)
	if err := k8sClient.Create(ctx, config); err != nil {
		t.Fatalf("failed to create AgentConfig: %v", err)
	}

	testutil.WaitForCondition(t, ctx, 30*time.Second, func() bool {
		var c agentv1alpha1.AgentConfig
		if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(config), &c); err != nil {
			return false
		}
		for _, cond := range c.Status.Conditions {
			if cond.Type == "Ready" && cond.Status == metav1.ConditionTrue {
				return true
			}
		}
		return false
	})

	var updated agentv1alpha1.AgentConfig
	if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(config), &updated); err != nil {
		t.Fatalf("failed to get AgentConfig: %v", err)
	}

	readyCond := findCondition(updated.Status.Conditions, "Ready")
	if readyCond == nil {
		t.Fatal("Ready condition not found")
	}
	if readyCond.Reason != "Validated" {
		t.Errorf("Ready reason = %q, want Validated", readyCond.Reason)
	}
}
