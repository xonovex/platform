package testutil

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
)

// AgentRunOption configures an AgentRun.
type AgentRunOption func(*agentv1alpha1.AgentRun)

// WithAgent sets the agent type.
func WithAgent(agent agentv1alpha1.AgentType) AgentRunOption {
	return func(r *agentv1alpha1.AgentRun) {
		r.Spec.Agent = agent
	}
}

// WithRepository sets the repository URL.
func WithRepository(url string) AgentRunOption {
	return func(r *agentv1alpha1.AgentRun) {
		r.Spec.Repository = agentv1alpha1.RepositorySpec{URL: url}
	}
}

// WithProviderRef sets the provider reference.
func WithProviderRef(ref string) AgentRunOption {
	return func(r *agentv1alpha1.AgentRun) {
		r.Spec.ProviderRef = ref
	}
}

// WithImage sets the container image.
func WithImage(image string) AgentRunOption {
	return func(r *agentv1alpha1.AgentRun) {
		r.Spec.Image = image
	}
}

// WithTimeout sets the run timeout.
func WithTimeout(d metav1.Duration) AgentRunOption {
	return func(r *agentv1alpha1.AgentRun) {
		r.Spec.Timeout = &d
	}
}

// WithPrompt sets the prompt.
func WithPrompt(prompt string) AgentRunOption {
	return func(r *agentv1alpha1.AgentRun) {
		r.Spec.Prompt = prompt
	}
}

// WithPhase sets an initial status phase (for testing terminal-phase skipping).
func WithPhase(phase agentv1alpha1.AgentRunPhase) AgentRunOption {
	return func(r *agentv1alpha1.AgentRun) {
		r.Status.Phase = phase
	}
}

// NewAgentRun creates an AgentRun with defaults and applies options.
func NewAgentRun(namespace, name string, opts ...AgentRunOption) *agentv1alpha1.AgentRun {
	run := &agentv1alpha1.AgentRun{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: agentv1alpha1.AgentRunSpec{
			Agent:      agentv1alpha1.AgentTypeClaude,
			Repository: agentv1alpha1.RepositorySpec{URL: "https://github.com/example/repo"},
		},
	}
	for _, opt := range opts {
		opt(run)
	}
	return run
}

// AgentProviderOption configures an AgentProvider.
type AgentProviderOption func(*agentv1alpha1.AgentProvider)

// WithAgentTypes sets the supported agent types.
func WithAgentTypes(types ...agentv1alpha1.AgentType) AgentProviderOption {
	return func(p *agentv1alpha1.AgentProvider) {
		p.Spec.AgentTypes = types
	}
}

// WithAuthTokenSecretRef sets the auth token secret reference.
func WithAuthTokenSecretRef(name, key string) AgentProviderOption {
	return func(p *agentv1alpha1.AgentProvider) {
		p.Spec.AuthTokenSecretRef = &agentv1alpha1.SecretKeyRef{
			Name: name,
			Key:  key,
		}
	}
}

// WithEnvironment sets the provider environment variables.
func WithEnvironment(env map[string]string) AgentProviderOption {
	return func(p *agentv1alpha1.AgentProvider) {
		p.Spec.Environment = env
	}
}

// NewAgentProvider creates an AgentProvider with defaults and applies options.
func NewAgentProvider(namespace, name string, opts ...AgentProviderOption) *agentv1alpha1.AgentProvider {
	provider := &agentv1alpha1.AgentProvider{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: agentv1alpha1.AgentProviderSpec{
			AgentTypes: []agentv1alpha1.AgentType{agentv1alpha1.AgentTypeClaude},
		},
	}
	for _, opt := range opts {
		opt(provider)
	}
	return provider
}

// AgentConfigOption configures an AgentConfig.
type AgentConfigOption func(*agentv1alpha1.AgentConfig)

// WithDefaultAgent sets the default agent type.
func WithDefaultAgent(agent agentv1alpha1.AgentType) AgentConfigOption {
	return func(c *agentv1alpha1.AgentConfig) {
		c.Spec.DefaultAgent = agent
	}
}

// WithStorageSize sets the storage size for workspace PVCs.
func WithStorageSize(size string) AgentConfigOption {
	return func(c *agentv1alpha1.AgentConfig) {
		c.Spec.StorageSize = size
	}
}

// WithStorageClass sets the storage class for workspace PVCs.
func WithStorageClass(class string) AgentConfigOption {
	return func(c *agentv1alpha1.AgentConfig) {
		c.Spec.StorageClass = class
	}
}

// WithDefaultImage sets the default container image.
func WithDefaultImage(image string) AgentConfigOption {
	return func(c *agentv1alpha1.AgentConfig) {
		c.Spec.DefaultImage = image
	}
}

// WithDefaultProviders sets the default providers map.
func WithDefaultProviders(providers map[agentv1alpha1.AgentType]string) AgentConfigOption {
	return func(c *agentv1alpha1.AgentConfig) {
		c.Spec.DefaultProviders = providers
	}
}

// NewAgentConfig creates an AgentConfig with defaults and applies options.
func NewAgentConfig(namespace, name string, opts ...AgentConfigOption) *agentv1alpha1.AgentConfig {
	config := &agentv1alpha1.AgentConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	for _, opt := range opts {
		opt(config)
	}
	return config
}

// NewSecret creates a Secret with the given data.
func NewSecret(namespace, name string, data map[string][]byte) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Data: data,
	}
}
