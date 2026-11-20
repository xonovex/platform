package webhook

import (
	"context"
	"fmt"
	"maps"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	agentv1alpha1 "github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/api/v1alpha1"
	"github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/internal/resolver"
)

// applyReferencedDefaults snapshots harness, provider, and toolchain references
// during admission so reconciliation consumes the exact execution inputs policy
// approved. Toolchain images take precedence because the controller executes
// that image.
func (w *AgentRunWebhook) applyReferencedDefaults(ctx context.Context, run *agentv1alpha1.AgentRun) error {
	harness, err := resolver.ResolveHarness(ctx, w.references, run.Namespace, run.Spec.HarnessRef, run.Spec.Harness)
	if err != nil {
		return fmt.Errorf("resolve harness defaults: %w", err)
	}
	if harness != nil {
		resolver.ApplyHarnessDefaults(run, harness)
		if run.Spec.ProviderRef == "" && run.Spec.Provider == nil {
			run.Spec.ProviderRef = harness.Spec.DefaultProvider
		}
		if run.Spec.HarnessRef != "" {
			run.Spec.Harness = harness.Spec.DeepCopy()
			run.Spec.HarnessRef = ""
		}
	}

	if run.Spec.ProviderRef != "" {
		referencedProvider, err := w.references.GetAgentProvider(ctx, types.NamespacedName{Name: run.Spec.ProviderRef, Namespace: run.Namespace})
		if err != nil {
			return fmt.Errorf("resolve provider defaults: %w", err)
		}
		if err := validateProviderConfig(
			referencedProvider.Spec.PresetRef,
			referencedProvider.Spec.AgentType,
			referencedProvider.Spec.AuthTokenSecretRef,
			referencedProvider.Spec.AuthTokenEnv,
			referencedProvider.Spec.Environment,
			referencedProvider.Spec.CliArgs,
		); err != nil {
			return fmt.Errorf("invalid referenced provider: %w", err)
		}
		run.Spec.Provider = &agentv1alpha1.ProviderSpec{
			PresetRef:     referencedProvider.Spec.PresetRef,
			AgentType:     referencedProvider.Spec.AgentType,
			AuthSecretRef: referencedProvider.Spec.AuthTokenSecretRef.DeepCopy(),
			AuthTokenEnv:  referencedProvider.Spec.AuthTokenEnv,
			Environment:   maps.Clone(referencedProvider.Spec.Environment),
			CliArgs:       append([]string{}, referencedProvider.Spec.CliArgs...),
		}
		run.Spec.ProviderRef = ""
	}

	toolchain, err := resolver.ResolveToolchain(ctx, w.references, run.Namespace, run.Spec.ToolchainRef, run.Spec.Toolchain)
	if err != nil {
		return fmt.Errorf("resolve toolchain defaults: %w", err)
	}
	if toolchain != nil && toolchain.Nix != nil {
		if err := validateNixSpec(toolchain.Nix); err != nil {
			return err
		}
		run.Spec.Toolchain = toolchain.DeepCopy()
		run.Spec.ToolchainRef = ""
		run.Spec.Image = toolchain.Nix.Image
	}
	return nil
}

func applyBuiltInDefaults(run *agentv1alpha1.AgentRun) {
	if run.Spec.Timeout == nil {
		defaultTimeout := metav1.Duration{Duration: resolver.DefaultTimeout}
		run.Spec.Timeout = &defaultTimeout
	}
}

// applyPolicyDefaults fills the image, timeout and runtimeClassName a run leaves
// empty from the namespace policy defaults; a nil policy changes nothing.
func applyPolicyDefaults(run *agentv1alpha1.AgentRun, policy *agentv1alpha1.AgentPolicy) {
	if policy == nil {
		return
	}
	defaults := policy.Spec.Defaults
	if run.Spec.Image == "" && defaults.Image != "" {
		run.Spec.Image = defaults.Image
	}
	if run.Spec.Timeout == nil && defaults.Timeout != nil {
		run.Spec.Timeout = defaults.Timeout.DeepCopy()
	}
	if run.Spec.RuntimeClassName == nil && defaults.RuntimeClassName != nil {
		runtimeClassName := *defaults.RuntimeClassName
		run.Spec.RuntimeClassName = &runtimeClassName
	}
}
