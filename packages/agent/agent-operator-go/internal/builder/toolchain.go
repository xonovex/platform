package builder

import (
	corev1 "k8s.io/api/core/v1"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
)

// ToolchainContributor provides containers, volumes, mounts, and env vars for a toolchain
type ToolchainContributor interface {
	InitContainer() *corev1.Container
	Volumes() []corev1.Volume
	VolumeMounts() []corev1.VolumeMount
	EnvVars() []corev1.EnvVar
}

// Toolchains returns the active toolchain contributors for the given toolchain spec
func Toolchains(tc *agentv1alpha1.ToolchainSpec) []ToolchainContributor {
	if tc == nil || tc.Type != agentv1alpha1.ToolchainTypeNix || tc.Nix == nil || len(tc.Nix.Packages) == 0 {
		return nil
	}
	return []ToolchainContributor{NewNixToolchain(tc.Nix)}
}
