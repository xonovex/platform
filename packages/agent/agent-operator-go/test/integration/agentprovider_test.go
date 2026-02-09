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

func TestAgentProvider_BecomesReadyWithValidSecret(t *testing.T) {
	ns := createNamespace(t, "provider-ready")

	secret := testutil.NewSecret(ns, "auth-secret", map[string][]byte{
		"token": []byte("test-token"),
	})
	if err := k8sClient.Create(ctx, secret); err != nil {
		t.Fatalf("failed to create Secret: %v", err)
	}

	provider := testutil.NewAgentProvider(ns, "test-provider",
		testutil.WithAgentTypes(agentv1alpha1.AgentTypeClaude),
		testutil.WithAuthTokenSecretRef("auth-secret", "token"),
	)
	if err := k8sClient.Create(ctx, provider); err != nil {
		t.Fatalf("failed to create AgentProvider: %v", err)
	}

	testutil.WaitForCondition(t, ctx, 30*time.Second, func() bool {
		var p agentv1alpha1.AgentProvider
		if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(provider), &p); err != nil {
			return false
		}
		return p.Status.Ready
	})

	var updated agentv1alpha1.AgentProvider
	if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(provider), &updated); err != nil {
		t.Fatalf("failed to get AgentProvider: %v", err)
	}

	readyCond := findCondition(updated.Status.Conditions, "Ready")
	if readyCond == nil {
		t.Fatal("Ready condition not found")
	}
	if readyCond.Status != metav1.ConditionTrue {
		t.Errorf("Ready condition status = %q, want True", readyCond.Status)
	}
}

func TestAgentProvider_NotReadyWhenSecretMissing(t *testing.T) {
	ns := createNamespace(t, "provider-not-ready")

	provider := testutil.NewAgentProvider(ns, "test-provider",
		testutil.WithAgentTypes(agentv1alpha1.AgentTypeClaude),
		testutil.WithAuthTokenSecretRef("nonexistent-secret", "token"),
	)
	if err := k8sClient.Create(ctx, provider); err != nil {
		t.Fatalf("failed to create AgentProvider: %v", err)
	}

	// Wait for reconciliation to set conditions
	testutil.WaitForCondition(t, ctx, 30*time.Second, func() bool {
		var p agentv1alpha1.AgentProvider
		if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(provider), &p); err != nil {
			return false
		}
		return len(p.Status.Conditions) > 0
	})

	var updated agentv1alpha1.AgentProvider
	if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(provider), &updated); err != nil {
		t.Fatalf("failed to get AgentProvider: %v", err)
	}

	if updated.Status.Ready {
		t.Error("expected provider to be not ready with missing secret")
	}

	readyCond := findCondition(updated.Status.Conditions, "Ready")
	if readyCond == nil {
		t.Fatal("Ready condition not found")
	}
	if readyCond.Status != metav1.ConditionFalse {
		t.Errorf("Ready condition status = %q, want False", readyCond.Status)
	}
}

func TestAgentProvider_ReadyWithoutSecretRef(t *testing.T) {
	ns := createNamespace(t, "provider-no-secret")

	provider := testutil.NewAgentProvider(ns, "test-provider",
		testutil.WithAgentTypes(agentv1alpha1.AgentTypeClaude, agentv1alpha1.AgentTypeOpencode),
	)
	if err := k8sClient.Create(ctx, provider); err != nil {
		t.Fatalf("failed to create AgentProvider: %v", err)
	}

	testutil.WaitForCondition(t, ctx, 30*time.Second, func() bool {
		var p agentv1alpha1.AgentProvider
		if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(provider), &p); err != nil {
			return false
		}
		return p.Status.Ready
	})
}

func findCondition(conditions []metav1.Condition, condType string) *metav1.Condition {
	for i := range conditions {
		if conditions[i].Type == condType {
			return &conditions[i]
		}
	}
	return nil
}
