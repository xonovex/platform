package docker

import (
	"path/filepath"
	"strings"
	"testing"

	netshared "github.com/xonovex/platform/packages/agent/agent-cli-go/internal/network/shared"
	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/provision"
	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/types"
)

const digestPinnedImage = "ghcr.io/xonovex/agent@sha256:" +
	"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

func envValue(env []string, key string) (string, bool) {
	prefix := key + "="
	for _, entry := range env {
		if strings.HasPrefix(entry, prefix) {
			return strings.TrimPrefix(entry, prefix), true
		}
	}
	return "", false
}

// processEnv is the environment the docker client is launched with, so it is the
// host environment rather than the container's allowlist.
func TestProcessEnv_CarriesTheHostEnvironment(t *testing.T) {
	t.Setenv("XONOVEX_PROCESS_ENV_PROBE", "host-value")
	cfg := dockerCfg(t, netshared.ModeNone, t.TempDir())

	env, err := NewIsolator().processEnv(cfg, provision.Contribution{})

	if err != nil {
		t.Fatalf("processEnv() error = %v", err)
	}
	if got, ok := envValue(env, "XONOVEX_PROCESS_ENV_PROBE"); !ok || got != "host-value" {
		t.Errorf("host variable = %q (present %v), want it carried through", got, ok)
	}
}

// The container HOME is fixed, so the environment fails on an unresolvable
// provider credential rather than on directories.
func TestProcessEnv_ReportsAnUnresolvableProviderToken(t *testing.T) {
	cfg := dockerCfg(t, netshared.ModeNone, t.TempDir())
	cfg.Provider = &types.ModelProvider{
		CredentialSourceEnv: "MISSING_DOCKER_PROCESS_ENV_TOKEN",
		CredentialTargetEnv: "AGENT_PROVIDER_TOKEN",
	}

	if _, err := NewIsolator().processEnv(cfg, provision.Contribution{}); err == nil {
		t.Fatal("processEnv() must fail when the provider token cannot be resolved")
	}
}

func TestProcessEnv_ReportsMalformedCustomEnvironment(t *testing.T) {
	cfg := dockerCfg(t, netshared.ModeNone, t.TempDir())
	cfg.CustomEnv = []string{"not-a-key-value-pair"}

	if _, err := NewIsolator().processEnv(cfg, provision.Contribution{}); err == nil {
		t.Fatal("processEnv() must fail on a malformed custom environment entry")
	}
}

func TestTerminalCommand_ReturnsTheCommandAndItsEnvironment(t *testing.T) {
	cfg := dockerCfg(t, netshared.ModeNone, t.TempDir())

	command, env, err := NewIsolator().TerminalCommand(cfg, provision.Contribution{})

	if err != nil {
		t.Fatalf("TerminalCommand() error = %v", err)
	}
	if len(command) == 0 || command[0] != "docker" {
		t.Fatalf("command = %v, want it to start with docker", command)
	}
	expected := dockerCommand(t, cfg, provision.Contribution{})
	if strings.Join(command, " ") != strings.Join(expected, " ") {
		t.Errorf("command = %v, want the same command Command() returns", command)
	}
	if len(env) == 0 {
		t.Error("TerminalCommand() returned no environment")
	}
}

func TestTerminalCommand_ReportsAnUnresolvableDirectory(t *testing.T) {
	cfg := dockerCfg(t, netshared.ModeNone, filepath.Join(t.TempDir(), "absent"))

	if _, _, err := NewIsolator().TerminalCommand(
		cfg, provision.Contribution{},
	); err == nil {
		t.Fatal("TerminalCommand() must fail when the work directory cannot be resolved")
	}
}

// The image is a tool source of its own, so an unpinned image forfeits the
// guarantee no matter how the provisioner is configured.
func TestPinnedProvision_RequiresADigestPinnedImage(t *testing.T) {
	tests := []struct {
		name              string
		method            provision.ProvisionMethod
		provisionerPinned bool
		image             string
		want              bool
	}{
		{
			name:   "pinned image is the only tool source",
			method: provision.ProvisionNone,
			image:  digestPinnedImage,
			want:   true,
		},
		{
			name:              "pinned image and pinned provisioner",
			method:            provision.ProvisionNix,
			provisionerPinned: true,
			image:             digestPinnedImage,
			want:              true,
		},
		{
			name:              "pinned image but unpinned provisioner",
			method:            provision.ProvisionNix,
			provisionerPinned: false,
			image:             digestPinnedImage,
			want:              false,
		},
		{
			name:              "tag-only image",
			method:            provision.ProvisionNone,
			provisionerPinned: true,
			image:             "ghcr.io/xonovex/agent:latest",
			want:              false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := NewIsolator().PinnedProvision(
				test.method, test.provisionerPinned, test.image,
			)
			if got != test.want {
				t.Errorf("PinnedProvision() = %v, want %v", got, test.want)
			}
		})
	}
}

// runc reduces attack surface but is not a kernel trust boundary; only gVisor
// counts.
func TestKernelIsolated_OnlyAcceptsGvisorRuntimes(t *testing.T) {
	tests := map[string]bool{
		"runsc":  true,
		"gvisor": true,
		"runc":   false,
		"":       false,
		"kata":   false,
	}

	for runtime, want := range tests {
		if got := NewIsolator().KernelIsolated(runtime); got != want {
			t.Errorf("KernelIsolated(%q) = %v, want %v", runtime, got, want)
		}
	}
}

func TestHidesHost_RequiresADigestPinnedImage(t *testing.T) {
	isolator := NewIsolator()

	if !isolator.HidesHost(false, digestPinnedImage) {
		t.Error("HidesHost() = false for a digest-pinned image, want true")
	}
	if isolator.HidesHost(false, "ghcr.io/xonovex/agent:latest") {
		t.Error("HidesHost() = true for a tag-only image, want false")
	}
}
