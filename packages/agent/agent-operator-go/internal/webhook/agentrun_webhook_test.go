package webhook

import (
	"context"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
)

func TestAgentRunWebhook_Default_SetsAgent(t *testing.T) {
	w := &AgentRunWebhook{}
	run := &agentv1alpha1.AgentRun{
		Spec: agentv1alpha1.AgentRunSpec{},
	}

	if err := w.Default(context.Background(), run); err != nil {
		t.Fatalf("Default() error = %v", err)
	}

	if run.Spec.Agent != agentv1alpha1.AgentTypeClaude {
		t.Errorf("Agent = %q, want %q", run.Spec.Agent, agentv1alpha1.AgentTypeClaude)
	}
}

func TestAgentRunWebhook_Default_SetsTimeout(t *testing.T) {
	w := &AgentRunWebhook{}
	run := &agentv1alpha1.AgentRun{
		Spec: agentv1alpha1.AgentRunSpec{
			Agent: agentv1alpha1.AgentTypeClaude,
		},
	}

	if err := w.Default(context.Background(), run); err != nil {
		t.Fatalf("Default() error = %v", err)
	}

	if run.Spec.Timeout == nil {
		t.Fatal("Timeout is nil, want non-nil")
	}
	if run.Spec.Timeout.Duration != time.Hour {
		t.Errorf("Timeout = %v, want %v", run.Spec.Timeout.Duration, time.Hour)
	}
}

func TestAgentRunWebhook_Default_PreservesExistingValues(t *testing.T) {
	w := &AgentRunWebhook{}
	customTimeout := metav1.Duration{Duration: 30 * time.Minute}
	run := &agentv1alpha1.AgentRun{
		Spec: agentv1alpha1.AgentRunSpec{
			Agent:   agentv1alpha1.AgentTypeOpencode,
			Timeout: &customTimeout,
		},
	}

	if err := w.Default(context.Background(), run); err != nil {
		t.Fatalf("Default() error = %v", err)
	}

	if run.Spec.Agent != agentv1alpha1.AgentTypeOpencode {
		t.Errorf("Agent = %q, want %q (should not override)", run.Spec.Agent, agentv1alpha1.AgentTypeOpencode)
	}
	if run.Spec.Timeout.Duration != 30*time.Minute {
		t.Errorf("Timeout = %v, want %v (should not override)", run.Spec.Timeout.Duration, 30*time.Minute)
	}
}

func TestAgentRunWebhook_Validate_ValidClaude(t *testing.T) {
	w := &AgentRunWebhook{}
	run := &agentv1alpha1.AgentRun{
		Spec: agentv1alpha1.AgentRunSpec{
			Agent: agentv1alpha1.AgentTypeClaude,
			Repository: agentv1alpha1.RepositorySpec{
				URL: "https://github.com/example/repo.git",
			},
		},
	}

	warnings, err := w.ValidateCreate(context.Background(), run)
	if err != nil {
		t.Errorf("ValidateCreate() error = %v", err)
	}
	if len(warnings) > 0 {
		t.Errorf("ValidateCreate() warnings = %v, want none", warnings)
	}
}

func TestAgentRunWebhook_Validate_ValidOpencode(t *testing.T) {
	w := &AgentRunWebhook{}
	run := &agentv1alpha1.AgentRun{
		Spec: agentv1alpha1.AgentRunSpec{
			Agent: agentv1alpha1.AgentTypeOpencode,
			Repository: agentv1alpha1.RepositorySpec{
				URL: "https://github.com/example/repo.git",
			},
		},
	}

	_, err := w.ValidateCreate(context.Background(), run)
	if err != nil {
		t.Errorf("ValidateCreate() error = %v", err)
	}
}

func TestAgentRunWebhook_Validate_InvalidAgentType(t *testing.T) {
	w := &AgentRunWebhook{}
	run := &agentv1alpha1.AgentRun{
		Spec: agentv1alpha1.AgentRunSpec{
			Agent: "invalid",
			Repository: agentv1alpha1.RepositorySpec{
				URL: "https://github.com/example/repo.git",
			},
		},
	}

	_, err := w.ValidateCreate(context.Background(), run)
	if err == nil {
		t.Error("ValidateCreate() expected error for invalid agent type")
	}
}

func TestAgentRunWebhook_Validate_MissingRepoURL(t *testing.T) {
	w := &AgentRunWebhook{}
	run := &agentv1alpha1.AgentRun{
		Spec: agentv1alpha1.AgentRunSpec{
			Agent:      agentv1alpha1.AgentTypeClaude,
			Repository: agentv1alpha1.RepositorySpec{},
		},
	}

	_, err := w.ValidateCreate(context.Background(), run)
	if err == nil {
		t.Error("ValidateCreate() expected error for missing repo URL")
	}
}

func TestAgentRunWebhook_Validate_BothProviderRefAndInline(t *testing.T) {
	w := &AgentRunWebhook{}
	run := &agentv1alpha1.AgentRun{
		Spec: agentv1alpha1.AgentRunSpec{
			Agent: agentv1alpha1.AgentTypeClaude,
			Repository: agentv1alpha1.RepositorySpec{
				URL: "https://github.com/example/repo.git",
			},
			ProviderRef: "my-provider",
			Provider: &agentv1alpha1.ProviderSpec{
				Name: "gemini",
			},
		},
	}

	_, err := w.ValidateCreate(context.Background(), run)
	if err == nil {
		t.Error("ValidateCreate() expected error for both providerRef and inline provider")
	}
}

func TestAgentRunWebhook_ValidateUpdate(t *testing.T) {
	w := &AgentRunWebhook{}
	oldRun := &agentv1alpha1.AgentRun{
		Spec: agentv1alpha1.AgentRunSpec{
			Agent:      agentv1alpha1.AgentTypeClaude,
			Repository: agentv1alpha1.RepositorySpec{URL: "https://example.com/repo.git"},
		},
	}
	newRun := &agentv1alpha1.AgentRun{
		Spec: agentv1alpha1.AgentRunSpec{
			Agent:      agentv1alpha1.AgentTypeClaude,
			Repository: agentv1alpha1.RepositorySpec{URL: "https://example.com/repo.git"},
		},
	}

	_, err := w.ValidateUpdate(context.Background(), oldRun, newRun)
	if err != nil {
		t.Errorf("ValidateUpdate() error = %v", err)
	}
}

func TestAgentRunWebhook_ValidateDelete(t *testing.T) {
	w := &AgentRunWebhook{}
	run := &agentv1alpha1.AgentRun{}

	_, err := w.ValidateDelete(context.Background(), run)
	if err != nil {
		t.Errorf("ValidateDelete() error = %v", err)
	}
}
