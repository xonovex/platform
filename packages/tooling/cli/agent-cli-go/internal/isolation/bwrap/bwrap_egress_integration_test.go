//go:build integration

package bwrap

import (
	"os/exec"
	"testing"
)

// TestBwrap_NetworkNoneBlocksEgress verifies the mechanism the regression guard
// relies on: --unshare-net actually isolates the network. It spawns bwrap and
// bash, so it carries the integration build tag and stays out of ci-check; it is
// gated on both binaries and is robust offline (an unshared netns has no route,
// so the connect fails regardless of host connectivity).
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
