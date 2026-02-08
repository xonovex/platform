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
			AgentTypes: []agentv1alpha1.AgentType{agentv1alpha1.AgentTypeClaude},
		},
	}

	_, err := w.ValidateCreate(context.Background(), provider)
	if err != nil {
		t.Errorf("ValidateCreate() error = %v", err)
	}
}

func TestAgentProviderWebhook_Validate_MultipleAgentTypes(t *testing.T) {
	w := &AgentProviderWebhook{}
	provider := &agentv1alpha1.AgentProvider{
		Spec: agentv1alpha1.AgentProviderSpec{
			AgentTypes: []agentv1alpha1.AgentType{
				agentv1alpha1.AgentTypeClaude,
				agentv1alpha1.AgentTypeOpencode,
			},
		},
	}

	_, err := w.ValidateCreate(context.Background(), provider)
	if err != nil {
		t.Errorf("ValidateCreate() error = %v", err)
	}
}

func TestAgentProviderWebhook_Validate_EmptyAgentTypes(t *testing.T) {
	w := &AgentProviderWebhook{}
	provider := &agentv1alpha1.AgentProvider{
		Spec: agentv1alpha1.AgentProviderSpec{
			AgentTypes: []agentv1alpha1.AgentType{},
		},
	}

	_, err := w.ValidateCreate(context.Background(), provider)
	if err == nil {
		t.Error("ValidateCreate() expected error for empty agent types")
	}
}

func TestAgentProviderWebhook_Validate_InvalidAgentType(t *testing.T) {
	w := &AgentProviderWebhook{}
	provider := &agentv1alpha1.AgentProvider{
		Spec: agentv1alpha1.AgentProviderSpec{
			AgentTypes: []agentv1alpha1.AgentType{"invalid"},
		},
	}

	_, err := w.ValidateCreate(context.Background(), provider)
	if err == nil {
		t.Error("ValidateCreate() expected error for invalid agent type")
	}
}

func TestAgentProviderWebhook_ValidateUpdate(t *testing.T) {
	w := &AgentProviderWebhook{}
	old := &agentv1alpha1.AgentProvider{
		Spec: agentv1alpha1.AgentProviderSpec{
			AgentTypes: []agentv1alpha1.AgentType{agentv1alpha1.AgentTypeClaude},
		},
	}
	new := &agentv1alpha1.AgentProvider{
		Spec: agentv1alpha1.AgentProviderSpec{
			AgentTypes: []agentv1alpha1.AgentType{agentv1alpha1.AgentTypeClaude, agentv1alpha1.AgentTypeOpencode},
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
