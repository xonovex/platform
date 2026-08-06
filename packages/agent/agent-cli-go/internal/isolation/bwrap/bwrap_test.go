package bwrap

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	isoshared "github.com/xonovex/platform/packages/agent/agent-cli-go/internal/isolation/shared"
	netshared "github.com/xonovex/platform/packages/agent/agent-cli-go/internal/network/shared"
	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/provision"
	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/types"
)

// TestBwrap_NetworkNoneBlocksEgress verifies the mechanism the regression guard
// relies on: --unshare-net actually isolates the network. It is gated on bwrap +
// bash and is robust offline (an unshared netns has no route, so the connect
// fails regardless of host connectivity).
func TestBwrap_NetworkNoneBlocksEgress(t *testing.T) {
	for _, bin := range []string{"bwrap", "bash"} {
		if _, err := exec.LookPath(bin); err != nil {
			t.Skipf("%s not available", bin)
		}
	}
	probe := exec.Command("bwrap", "--unshare-net", "--ro-bind", "/", "/", "--dev", "/dev", "--proc", "/proc",
		"--", "bash", "-c", "exec 3<>/dev/tcp/1.1.1.1/53")
	if err := probe.Run(); err == nil {
		t.Error("--unshare-net should block egress, but the TCP connect succeeded")
	}
}

// claudeCfg builds a run config rooted in an empty temporary home directory, so
// the arguments depend on the config paths a test creates rather than on the
// host user's configuration.
func claudeCfg(t *testing.T, net netshared.Mode, passthrough bool, workDir string) isoshared.RunConfig {
	t.Helper()
	return isoshared.RunConfig{
		Agent:           &types.AgentConfig{Type: types.AgentClaude, Binary: "claude"},
		WorkDir:         workDir,
		HomeDir:         t.TempDir(),
		Network:         net,
		HostPassthrough: passthrough,
	}
}

// canonicalPath resolves the symlinks a bind source is reported through, which
// is how the isolator names a path it mounts.
func canonicalPath(t *testing.T, path string) string {
	t.Helper()
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		t.Fatalf("EvalSymlinks(%q) error = %v", path, err)
	}
	return resolved
}

func bwrapCommand(t *testing.T, cfg isoshared.RunConfig, c provision.Contribution) []string {
	t.Helper()
	command, err := NewIsolator().Command(cfg, c)
	if err != nil {
		t.Fatalf("Command() error = %v", err)
	}
	return command
}

// argHas reports whether args contains s.
func argHas(args []string, s string) bool {
	for _, a := range args {
		if a == s {
			return true
		}
	}
	return false
}

// argHasPair reports whether a is immediately followed by b.
func argHasPair(args []string, a, b string) bool {
	for i := 0; i+1 < len(args); i++ {
		if args[i] == a && args[i+1] == b {
			return true
		}
	}
	return false
}

func TestBwrap_DenyDefaultHardening(t *testing.T) {
	t.Setenv("BWRAP_UNRELATED_HOST_VALUE", "must-not-enter-sandbox")
	work := t.TempDir()
	cfg := claudeCfg(t, netshared.ModeNone, false, work)
	args := bwrapCommand(t, cfg, provision.Contribution{})

	if !argHasPair(args, "--dev", "/dev") {
		t.Error("missing minimal --dev /dev")
	}
	if argHas(args, "--dev-bind") {
		t.Error("must not use --dev-bind (exposes /dev/sda, /dev/mem)")
	}
	if !argHasPair(args, "--unsetenv", "BWRAP_UNRELATED_HOST_VALUE") {
		t.Error("unrelated host environment variable must be explicitly unset")
	}
	homeDir := canonicalPath(t, cfg.HomeDir)
	if !argHasPair(args, "--tmpfs", homeDir) {
		t.Error("HOME must be a sandbox-local tmpfs")
	}
	// No host-$HOME bind in deny-default mode.
	if argHasTriple(args, "--bind", homeDir, homeDir) {
		t.Error("must not bind host-$HOME in deny-default mode")
	}
	// No host system dirs bound.
	if argHasPair(args, "--ro-bind", "/usr") {
		t.Error("deny-default must not ro-bind host /usr")
	}
}

// The sandbox-local HOME is a tmpfs, so the curated config paths are the only
// home content bound back into it, each read-only at its original path.
func TestBwrap_BindsExistingUserConfigPathsReadOnly(t *testing.T) {
	cfg := claudeCfg(t, netshared.ModeNone, false, t.TempDir())
	configDir := filepath.Join(cfg.HomeDir, ".claude")
	if err := os.Mkdir(configDir, 0o755); err != nil {
		t.Fatalf("create %q: %v", configDir, err)
	}
	configFile := filepath.Join(cfg.HomeDir, ".claude.json")
	if err := os.WriteFile(configFile, []byte("{}\n"), 0o600); err != nil {
		t.Fatalf("write %q: %v", configFile, err)
	}

	args := bwrapCommand(t, cfg, provision.Contribution{})

	homeDir := canonicalPath(t, cfg.HomeDir)
	for _, configPath := range []string{".claude", ".claude.json"} {
		source := canonicalPath(t, filepath.Join(cfg.HomeDir, configPath))
		if !argHasTriple(args, "--ro-bind", source, filepath.Join(homeDir, configPath)) {
			t.Errorf("missing read-only user config bind for %q", configPath)
		}
	}
}

// An absent config path contributes no bind, so a home without agent
// configuration leaves the sandbox HOME an empty tmpfs.
func TestBwrap_SkipsAbsentUserConfigPaths(t *testing.T) {
	cfg := claudeCfg(t, netshared.ModeNone, false, t.TempDir())

	args := bwrapCommand(t, cfg, provision.Contribution{})

	homeDir := canonicalPath(t, cfg.HomeDir)
	for _, configPath := range []string{".claude", ".claude.json"} {
		if argHas(args, filepath.Join(homeDir, configPath)) {
			t.Errorf("absent user config path %q must not be bound", configPath)
		}
	}
}

func TestBwrap_NetworkExplicit(t *testing.T) {
	work := t.TempDir()
	args := bwrapCommand(t, claudeCfg(t, netshared.ModeNone, false, work), provision.Contribution{})
	if !argHas(args, "--unshare-net") || argHas(args, "--share-net") {
		t.Error("network none must --unshare-net and not --share-net")
	}
	args = bwrapCommand(t, claudeCfg(t, netshared.ModeHost, false, work), provision.Contribution{})
	if !argHas(args, "--share-net") || argHas(args, "--unshare-net") {
		t.Error("network host must --share-net and not --unshare-net")
	}
	if _, err := NewIsolator().Command(claudeCfg(t, netshared.ModeProxy, false, work), provision.Contribution{}); !errors.Is(err, netshared.ErrProxyEnforcementUnavailable) {
		t.Errorf("network proxy error = %v, want %v", err, netshared.ErrProxyEnforcementUnavailable)
	}
}

func TestBwrap_AppliesContribution(t *testing.T) {
	work := t.TempDir()
	closure := t.TempDir()
	c := provision.Contribution{
		RoBindPaths:  []string{closure},
		PathEntries:  []string{"/nix/store/abc/bin"},
		Env:          map[string]string{"FOO": "bar"},
		InitCommands: []string{"echo hi"},
	}
	args := bwrapCommand(t, claudeCfg(t, netshared.ModeNone, false, work), c)
	env, err := NewIsolator().sandboxEnv(claudeCfg(t, netshared.ModeNone, false, work), c, work)
	if err != nil {
		t.Fatalf("sandboxEnv() error = %v", err)
	}

	if !argHasTriple(args, "--ro-bind", closure, closure) {
		t.Error("contribution RoBindPaths must be ro-bound")
	}
	if path := env["PATH"]; !strings.HasPrefix(path, "/nix/store/abc/bin:") {
		t.Errorf("PATH = %q, want contribution entry prepended", path)
	}
	if env["FOO"] != "bar" {
		t.Error("contribution env not applied")
	}
	// Init command wraps the agent: the trailing command is sh -c '... echo hi ...'.
	if !argHas(args, "--") {
		t.Fatal("missing command separator")
	}
	tail := strings.Join(args, " ")
	if !strings.Contains(tail, "echo hi") || !strings.Contains(tail, "sh -c") {
		t.Errorf("init command not wrapped: %q", tail)
	}
}

func TestBwrap_HostPassthrough(t *testing.T) {
	work := t.TempDir()
	args := bwrapCommand(t, claudeCfg(t, netshared.ModeHost, true, work), provision.Contribution{})
	if !argHasTriple(args, "--ro-bind", "/usr", "/usr") {
		t.Error("HostPassthrough must ro-bind host /usr")
	}
}

func TestBwrap_CommandRejectsMissingBindAndProviderToken(t *testing.T) {
	work := t.TempDir()
	cfg := claudeCfg(t, netshared.ModeNone, false, work)
	cfg.RoBindPaths = []string{work + "/missing"}
	if _, err := NewIsolator().Command(cfg, provision.Contribution{}); err == nil {
		t.Error("Command() error = nil, want missing bind error")
	}

	cfg.RoBindPaths = nil
	cfg.Provider = &types.ModelProvider{
		CredentialSourceEnv: "MISSING_BWRAP_TEST_TOKEN",
		CredentialTargetEnv: "AGENT_PROVIDER_TOKEN",
	}
	if _, err := NewIsolator().Command(cfg, provision.Contribution{}); err == nil {
		t.Error("Command() error = nil, want missing provider token error")
	}
}

func TestBwrap_CommandKeepsProviderSecretOutOfArguments(t *testing.T) {
	const secret = "bwrap-secret-must-not-appear"
	t.Setenv("BWRAP_TEST_TOKEN", secret)
	cfg := claudeCfg(t, netshared.ModeNone, false, t.TempDir())
	cfg.Provider = &types.ModelProvider{
		CredentialSourceEnv: "BWRAP_TEST_TOKEN",
		CredentialTargetEnv: "ANTHROPIC_AUTH_TOKEN",
		Environment:         map[string]string{"ANTHROPIC_BASE_URL": "http://proxy.example"},
	}

	command := bwrapCommand(t, cfg, provision.Contribution{})

	if strings.Contains(strings.Join(command, " "), secret) {
		t.Fatal("bubblewrap command arguments contain the provider secret")
	}
}

// argHasTriple reports whether a,b,c appear consecutively.
func argHasTriple(args []string, a, b, c string) bool {
	for i := 0; i+2 < len(args); i++ {
		if args[i] == a && args[i+1] == b && args[i+2] == c {
			return true
		}
	}
	return false
}

func TestBwrap_Capabilities(t *testing.T) {
	i := NewIsolator()
	if !i.HidesHost(false, "") {
		t.Error("bwrap deny-default must hide host tools")
	}
	if i.HidesHost(true, "") {
		t.Error("bwrap with HostPassthrough must not hide host tools")
	}
	if i.KernelIsolated("runsc") {
		t.Error("bwrap is never a kernel boundary")
	}
}
