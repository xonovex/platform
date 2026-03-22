package testutil

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
)

// AgentRunOption configures an AgentRun.
type AgentRunOption func(*agentv1alpha1.AgentRun)

// WithHarnessRef sets the harness reference.
func WithHarnessRef(ref string) AgentRunOption {
	return func(r *agentv1alpha1.AgentRun) {
		r.Spec.HarnessRef = ref
	}
}

// WithHarness sets the inline harness.
func WithHarness(spec *agentv1alpha1.AgentSpec) AgentRunOption {
	return func(r *agentv1alpha1.AgentRun) {
		r.Spec.Harness = spec
	}
}

// WithWorkspace sets the inline workspace.
func WithWorkspace(spec *agentv1alpha1.WorkspaceSpec) AgentRunOption {
	return func(r *agentv1alpha1.AgentRun) {
		r.Spec.Workspace = spec
	}
}

// WithToolchain sets the inline toolchain.
func WithToolchain(spec *agentv1alpha1.ToolchainSpec) AgentRunOption {
	return func(r *agentv1alpha1.AgentRun) {
		r.Spec.Toolchain = spec
	}
}

// WithToolchainRef sets the toolchain reference.
func WithToolchainRef(ref string) AgentRunOption {
	return func(r *agentv1alpha1.AgentRun) {
		r.Spec.ToolchainRef = ref
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

// WithSecurityContext sets the container security context override.
func WithSecurityContext(sc *corev1.SecurityContext) AgentRunOption {
	return func(r *agentv1alpha1.AgentRun) {
		r.Spec.SecurityContext = sc
	}
}

// WithPodSecurityContext sets the pod-level security context override.
func WithPodSecurityContext(psc *corev1.PodSecurityContext) AgentRunOption {
	return func(r *agentv1alpha1.AgentRun) {
		r.Spec.PodSecurityContext = psc
	}
}

// WithRuntimeClassName sets the runtime class name.
func WithRuntimeClassName(name string) AgentRunOption {
	return func(r *agentv1alpha1.AgentRun) {
		r.Spec.RuntimeClassName = &name
	}
}

// WithWorkspaceRef sets the workspace reference.
func WithWorkspaceRef(ref string) AgentRunOption {
	return func(r *agentv1alpha1.AgentRun) {
		r.Spec.WorkspaceRef = ref
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
			Workspace: &agentv1alpha1.WorkspaceSpec{
				Repository: agentv1alpha1.RepositorySpec{URL: "https://github.com/example/repo"},
			},
		},
	}
	for _, opt := range opts {
		opt(run)
	}
	return run
}

// AgentProviderOption configures an AgentProvider.
type AgentProviderOption func(*agentv1alpha1.AgentProvider)

// WithProviderType sets the provider type.
func WithProviderType(t agentv1alpha1.ProviderType) AgentProviderOption {
	return func(p *agentv1alpha1.AgentProvider) {
		p.Spec.Type = t
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
		Spec: agentv1alpha1.AgentProviderSpec{},
	}
	for _, opt := range opts {
		opt(provider)
	}
	return provider
}

// AgentHarnessOption configures an AgentHarness.
type AgentHarnessOption func(*agentv1alpha1.AgentHarness)

// WithHarnessType sets the agent type.
func WithHarnessType(t agentv1alpha1.AgentType) AgentHarnessOption {
	return func(h *agentv1alpha1.AgentHarness) {
		h.Spec.Type = t
	}
}

// WithDefaultImage sets the default image on the harness.
func WithDefaultImage(image string) AgentHarnessOption {
	return func(h *agentv1alpha1.AgentHarness) {
		h.Spec.DefaultImage = image
	}
}

// WithDefaultRuntimeClassName sets the default runtime class name on the harness.
func WithDefaultRuntimeClassName(name string) AgentHarnessOption {
	return func(h *agentv1alpha1.AgentHarness) {
		h.Spec.DefaultRuntimeClassName = &name
	}
}

// WithDefaultProvider sets the default provider on the harness.
func WithDefaultProvider(provider string) AgentHarnessOption {
	return func(h *agentv1alpha1.AgentHarness) {
		h.Spec.DefaultProvider = provider
	}
}

// NewAgentHarness creates an AgentHarness with defaults and applies options.
func NewAgentHarness(namespace, name string, opts ...AgentHarnessOption) *agentv1alpha1.AgentHarness {
	harness := &agentv1alpha1.AgentHarness{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: agentv1alpha1.AgentSpec{
			Type: agentv1alpha1.AgentTypeClaude,
		},
	}
	for _, opt := range opts {
		opt(harness)
	}
	return harness
}

// AgentWorkspaceOption configures an AgentWorkspace.
type AgentWorkspaceOption func(*agentv1alpha1.AgentWorkspace)

// WithWorkspaceRepository sets the workspace repository URL.
func WithWorkspaceRepository(url string) AgentWorkspaceOption {
	return func(ws *agentv1alpha1.AgentWorkspace) {
		ws.Spec.Repository = agentv1alpha1.RepositorySpec{URL: url}
	}
}

// WithWorkspaceRepositoryBranch sets the workspace repository branch.
func WithWorkspaceRepositoryBranch(url, branch string) AgentWorkspaceOption {
	return func(ws *agentv1alpha1.AgentWorkspace) {
		ws.Spec.Repository = agentv1alpha1.RepositorySpec{URL: url, Branch: branch}
	}
}

// WithWorkspaceStorageClass sets the storage class.
func WithWorkspaceStorageClass(class string) AgentWorkspaceOption {
	return func(ws *agentv1alpha1.AgentWorkspace) {
		ws.Spec.StorageClass = class
	}
}

// WithWorkspaceStorageSize sets the storage size.
func WithWorkspaceStorageSize(size string) AgentWorkspaceOption {
	return func(ws *agentv1alpha1.AgentWorkspace) {
		ws.Spec.StorageSize = size
	}
}

// WithSharedVolumes sets the shared volumes.
func WithSharedVolumes(volumes ...agentv1alpha1.SharedVolumeSpec) AgentWorkspaceOption {
	return func(ws *agentv1alpha1.AgentWorkspace) {
		ws.Spec.SharedVolumes = volumes
	}
}

// WithWorkspaceType sets the workspace type.
func WithWorkspaceType(t agentv1alpha1.WorkspaceType) AgentWorkspaceOption {
	return func(ws *agentv1alpha1.AgentWorkspace) {
		ws.Spec.Type = t
	}
}

// WithWorkspacePhase sets an initial status phase.
func WithWorkspacePhase(phase agentv1alpha1.AgentWorkspacePhase) AgentWorkspaceOption {
	return func(ws *agentv1alpha1.AgentWorkspace) {
		ws.Status.Phase = phase
	}
}

// NewAgentWorkspace creates an AgentWorkspace with defaults and applies options.
func NewAgentWorkspace(namespace, name string, opts ...AgentWorkspaceOption) *agentv1alpha1.AgentWorkspace {
	ws := &agentv1alpha1.AgentWorkspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: agentv1alpha1.AgentWorkspaceSpec{
			Repository: agentv1alpha1.RepositorySpec{URL: "https://github.com/example/repo"},
		},
	}
	for _, opt := range opts {
		opt(ws)
	}
	return ws
}

// E2ESecurityOverrides returns AgentRunOptions that relax security defaults
// for e2e tests using images that run as root (e.g. busybox:1.37).
func E2ESecurityOverrides() []AgentRunOption {
	f := false
	return []AgentRunOption{
		WithPodSecurityContext(&corev1.PodSecurityContext{RunAsNonRoot: &f}),
		WithSecurityContext(&corev1.SecurityContext{
			RunAsNonRoot:           &f,
			ReadOnlyRootFilesystem: &f,
		}),
	}
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
