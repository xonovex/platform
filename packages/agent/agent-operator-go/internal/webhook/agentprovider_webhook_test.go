package webhook

import (
	"context"
	"testing"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
)

func TestAgentProviderWebhook_Validate_Valid(t *testing.T) {
	w := &AgentProviderWebhook{}
	provider := &agentv1alpha1.AgentProvider{
		Spec: agentv1alpha1.AgentProviderSpec{
			Type: "anthropic",
		},
	}

	_, err := w.ValidateCreate(context.Background(), provider)
	if err != nil {
		t.Errorf("ValidateCreate() error = %v", err)
	}
}

func TestAgentProviderWebhook_Validate_Empty(t *testing.T) {
	w := &AgentProviderWebhook{}
	provider := &agentv1alpha1.AgentProvider{
		Spec: agentv1alpha1.AgentProviderSpec{},
	}

	_, err := w.ValidateCreate(context.Background(), provider)
	if err != nil {
		t.Errorf("ValidateCreate() error = %v", err)
	}
}

func TestAgentProviderWebhook_ValidateUpdate(t *testing.T) {
	w := &AgentProviderWebhook{}
	old := &agentv1alpha1.AgentProvider{
		Spec: agentv1alpha1.AgentProviderSpec{
			Type: "anthropic",
		},
	}
	new := &agentv1alpha1.AgentProvider{
		Spec: agentv1alpha1.AgentProviderSpec{
			Type: "openai",
		},
	}

	_, err := w.ValidateUpdate(context.Background(), old, new)
	if err != nil {
		t.Errorf("ValidateUpdate() error = %v", err)
	}
}

func TestAgentProviderWebhook_ValidateDelete(t *testing.T) {
	w := &AgentProviderWebhook{}
	_, err := w.ValidateDelete(context.Background(), &agentv1alpha1.AgentProvider{})
	if err != nil {
		t.Errorf("ValidateDelete() error = %v", err)
	}
}
