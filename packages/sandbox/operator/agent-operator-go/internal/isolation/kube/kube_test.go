package kube

import (
	"testing"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/network"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/provision"
)

const pinnedImage = "ghcr.io/xonovex/agent@sha256:" + "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

func TestPinnedProvisionNeedsADigestPinnedImage(t *testing.T) {
	confinement := New(nil)
	if !confinement.PinnedProvision(provision.ProvisionNone, false, pinnedImage) {
		t.Error("a digest-pinned image must be pinned provision")
	}
	for _, image := range []string{"", "node:trixie-slim", "ghcr.io/xonovex/agent@sha256:abc"} {
		if confinement.PinnedProvision(provision.ProvisionNix, true, image) {
			t.Errorf("image %q must not be pinned provision, whatever the provisioner declares", image)
		}
	}
}

func TestKernelIsolatedNeedsAnApprovedRuntimeClass(t *testing.T) {
	approved := []string{"gvisor", "kata"}
	confinement := New(approved)
	approved[0] = "runc"

	for runtime, want := range map[string]bool{"gvisor": true, "kata": true, "runc": false, "": false, " ": false} {
		if got := confinement.KernelIsolated(runtime); got != want {
			t.Errorf("KernelIsolated(%q) = %v, want %v", runtime, got, want)
		}
	}
	if New([]string{""}).KernelIsolated("") {
		t.Error("an empty runtime class is never kernel isolation")
	}
}

func TestTheFixedGuaranteesOfAPod(t *testing.T) {
	confinement := New(nil)
	if !confinement.HidesHost(true, "") {
		t.Error("a pod hides host tools")
	}
	if !confinement.ConfinesNetwork() {
		t.Error("a pod confines the network through its NetworkPolicy")
	}
	if confinement.EgressTransport() != network.TransportNone {
		t.Error("a pod declares no managed egress transport")
	}
	if !confinement.Bounded() {
		t.Error("a pod bounds the agent")
	}
	if confinement.AcceptsHostBinds() {
		t.Error("a pod accepts no host bind")
	}
	if !confinement.RequiresImage() {
		t.Error("a pod starts from its image")
	}
}
