package webhook

import (
	"context"
	"testing"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
)

func TestAgentToolchainWebhook_ValidatesPinnedNixToolchain(t *testing.T) {
	testCases := []struct {
		name    string
		nix     *agentv1alpha1.NixSpec
		wantErr bool
	}{
		{
			name: "accepts pinned package toolchain",
			nix: &agentv1alpha1.NixSpec{
				NixpkgsRev: testNixRevision,
				Packages:   []string{"ripgrep"},
				Image:      "ghcr.io/example/agent@sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			},
		},
		{
			name: "rejects moving image tag",
			nix: &agentv1alpha1.NixSpec{
				NixpkgsRev: testNixRevision,
				Packages:   []string{"ripgrep"},
				Image:      "ghcr.io/example/agent:latest",
			},
			wantErr: true,
		},
		{
			name: "rejects missing source revision",
			nix: &agentv1alpha1.NixSpec{
				Packages: []string{"ripgrep"},
				Image:    "ghcr.io/example/agent@sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			},
			wantErr: true,
		},
		{
			name: "rejects shell metacharacters",
			nix: &agentv1alpha1.NixSpec{
				NixpkgsRev: testNixRevision,
				Packages:   []string{"ripgrep;whoami"},
				Image:      "ghcr.io/example/agent@sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			},
			wantErr: true,
		},
	}
	webhook := &AgentToolchainWebhook{}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			toolchain := &agentv1alpha1.AgentToolchain{Spec: agentv1alpha1.ToolchainSpec{
				Type: agentv1alpha1.ToolchainTypeNix,
				Nix:  testCase.nix,
			}}

			_, err := webhook.ValidateCreate(context.Background(), toolchain)

			if (err != nil) != testCase.wantErr {
				t.Errorf("ValidateCreate() error = %v, wantErr %v", err, testCase.wantErr)
			}
		})
	}
}
