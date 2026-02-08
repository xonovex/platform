package webhook

import (
	"context"
	"testing"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
)

func TestAgentConfigWebhook_Validate_Empty(t *testing.T) {
	w := &AgentConfigWebhook{}
	config := &agentv1alpha1.AgentConfig{
		Spec: agentv1alpha1.AgentConfigSpec{},
	}

	_, err := w.ValidateCreate(context.Background(), config)
	if err != nil {
		t.Errorf("ValidateCreate() error = %v", err)
	}
}

func TestAgentConfigWebhook_Validate_ValidDefaultAgent(t *testing.T) {
	w := &AgentConfigWebhook{}

	for _, agent := range []agentv1alpha1.AgentType{agentv1alpha1.AgentTypeClaude, agentv1alpha1.AgentTypeOpencode} {
		config := &agentv1alpha1.AgentConfig{
			Spec: agentv1alpha1.AgentConfigSpec{
				DefaultAgent: agent,
			},
		}

		_, err := w.ValidateCreate(context.Background(), config)
		if err != nil {
			t.Errorf("ValidateCreate() with agent %q error = %v", agent, err)
		}
	}
}

func TestAgentConfigWebhook_Validate_InvalidDefaultAgent(t *testing.T) {
	w := &AgentConfigWebhook{}
	config := &agentv1alpha1.AgentConfig{
		Spec: agentv1alpha1.AgentConfigSpec{
			DefaultAgent: "invalid-agent",
		},
	}

	_, err := w.ValidateCreate(context.Background(), config)
	if err == nil {
		t.Error("ValidateCreate() expected error for invalid default agent")
	}
}

func TestAgentConfigWebhook_ValidateUpdate(t *testing.T) {
	w := &AgentConfigWebhook{}
	old := &agentv1alpha1.AgentConfig{
		Spec: agentv1alpha1.AgentConfigSpec{DefaultAgent: agentv1alpha1.AgentTypeClaude},
	}
	new := &agentv1alpha1.AgentConfig{
		Spec: agentv1alpha1.AgentConfigSpec{DefaultAgent: agentv1alpha1.AgentTypeOpencode},
	}

	_, err := w.ValidateUpdate(context.Background(), old, new)
	if err != nil {
		t.Errorf("ValidateUpdate() error = %v", err)
	}
}

func TestAgentConfigWebhook_ValidateDelete(t *testing.T) {
	w := &AgentConfigWebhook{}
	_, err := w.ValidateDelete(context.Background(), &agentv1alpha1.AgentConfig{})
	if err != nil {
		t.Errorf("ValidateDelete() error = %v", err)
	}
}
