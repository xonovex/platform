package bwrap

import (
	"os/exec"
	"strings"
	"testing"

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

func claudeCfg(net types.NetworkMethod, passthrough bool, workDir string) *types.SandboxConfig {
	return &types.SandboxConfig{
		Agent:           &types.AgentConfig{Type: types.AgentClaude, Binary: "claude"},
		HomeDir:         "/home/testuser",
		WorkDir:         workDir,
		Network:         net,
		HostPassthrough: passthrough,
	}
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

// setenvValue returns the value of a --setenv KEY, or "" if absent.
func setenvValue(args []string, key string) string {
	for i := 0; i+2 < len(args); i++ {
		if args[i] == "--setenv" && args[i+1] == key {
			return args[i+2]
		}
	}
	return ""
}

func TestBwrap_DenyDefaultHardening(t *testing.T) {
	work := t.TempDir()
	args := NewIsolator().Command(claudeCfg(types.NetworkNone, false, work), types.Contribution{})

	if !argHasPair(args, "--dev", "/dev") {
		t.Error("missing minimal --dev /dev")
	}
	if argHas(args, "--dev-bind") {
		t.Error("must not use --dev-bind (exposes /dev/sda, /dev/mem)")
	}
	if !argHas(args, "--clearenv") {
		t.Error("missing --clearenv")
	}
	if !argHasPair(args, "--tmpfs", "/home/testuser") {
		t.Error("HOME must be a sandbox-local tmpfs")
	}
	// No host-$HOME bind in deny-default mode.
	if argHasPair(args, "--bind", "/home/testuser") {
		t.Error("must not bind host-$HOME in deny-default mode")
	}
	// No host system dirs bound.
	if argHasPair(args, "--ro-bind", "/usr") {
		t.Error("deny-default must not ro-bind host /usr")
	}
}

func TestBwrap_NetworkExplicit(t *testing.T) {
	work := t.TempDir()
	// none + proxy must emit --unshare-net (the regression guard); host shares.
	for _, net := range []types.NetworkMethod{types.NetworkNone, types.NetworkProxy} {
		args := NewIsolator().Command(claudeCfg(net, false, work), types.Contribution{})
		if !argHas(args, "--unshare-net") {
			t.Errorf("network %q must emit --unshare-net", net)
		}
		if argHas(args, "--share-net") {
			t.Errorf("network %q must not share the net", net)
		}
	}
	args := NewIsolator().Command(claudeCfg(types.NetworkHost, false, work), types.Contribution{})
	if !argHas(args, "--share-net") || argHas(args, "--unshare-net") {
		t.Error("network host must --share-net and not --unshare-net")
	}
}

func TestBwrap_AppliesContribution(t *testing.T) {
	work := t.TempDir()
	closure := t.TempDir()
	c := types.Contribution{
		RoBindPaths:  []string{closure},
		PathEntries:  []string{"/nix/store/abc/bin"},
		Env:          map[string]string{"FOO": "bar"},
		InitCommands: []string{"echo hi"},
	}
	args := NewIsolator().Command(claudeCfg(types.NetworkNone, false, work), c)

	if !argHasTriple(args, "--ro-bind", closure, closure) {
		t.Error("contribution RoBindPaths must be ro-bound")
	}
	if path := setenvValue(args, "PATH"); !strings.HasPrefix(path, "/nix/store/abc/bin:") {
		t.Errorf("PATH = %q, want contribution entry prepended", path)
	}
	if setenvValue(args, "FOO") != "bar" {
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
	args := NewIsolator().Command(claudeCfg(types.NetworkHost, true, work), types.Contribution{})
	if !argHasTriple(args, "--ro-bind", "/usr", "/usr") {
		t.Error("HostPassthrough must ro-bind host /usr")
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
