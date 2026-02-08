package resolver

import (
	"context"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
)

func TestResolveConfig_NotFound(t *testing.T) {
	c := fake.NewClientBuilder().WithScheme(newScheme()).Build()

	config, err := ResolveConfig(context.Background(), c, "default")
	if err != nil {
		t.Fatalf("ResolveConfig() error = %v", err)
	}
	if config != nil {
		t.Errorf("config = %v, want nil", config)
	}
}

func TestResolveConfig_Found(t *testing.T) {
	agentConfig := &agentv1alpha1.AgentConfig{
		ObjectMeta: metav1.ObjectMeta{Name: "default-config", Namespace: "default"},
		Spec: agentv1alpha1.AgentConfigSpec{
			DefaultAgent: agentv1alpha1.AgentTypeClaude,
			StorageSize:  "20Gi",
		},
	}

	c := fake.NewClientBuilder().WithScheme(newScheme()).WithObjects(agentConfig).Build()

	config, err := ResolveConfig(context.Background(), c, "default")
	if err != nil {
		t.Fatalf("ResolveConfig() error = %v", err)
	}
	if config == nil {
		t.Fatal("config is nil, want non-nil")
	}
	if config.Spec.DefaultAgent != agentv1alpha1.AgentTypeClaude {
		t.Errorf("DefaultAgent = %q, want %q", config.Spec.DefaultAgent, agentv1alpha1.AgentTypeClaude)
	}
	if config.Spec.StorageSize != "20Gi" {
		t.Errorf("StorageSize = %q, want %q", config.Spec.StorageSize, "20Gi")
	}
}

func TestResolveConfig_WrongNamespace(t *testing.T) {
	agentConfig := &agentv1alpha1.AgentConfig{
		ObjectMeta: metav1.ObjectMeta{Name: "config", Namespace: "other-ns"},
		Spec: agentv1alpha1.AgentConfigSpec{
			DefaultAgent: agentv1alpha1.AgentTypeClaude,
		},
	}

	c := fake.NewClientBuilder().WithScheme(newScheme()).WithObjects(agentConfig).Build()

	config, err := ResolveConfig(context.Background(), c, "default")
	if err != nil {
		t.Fatalf("ResolveConfig() error = %v", err)
	}
	if config != nil {
		t.Errorf("config = %v, want nil (wrong namespace)", config)
	}
}
