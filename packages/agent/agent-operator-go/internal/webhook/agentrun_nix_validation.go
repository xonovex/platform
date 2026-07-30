package webhook

import (
	"fmt"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
	"github.com/xonovex/platform/packages/agent/agent-operator-go/internal/validator"
	sharednix "github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/provision/nix"
)

// validateNixSpec validates the nix toolchain: a pinned rev, exactly one source
// (packages XOR project flake), and a pre-built pinned image. The provisioning is
// build-time, so the image must be supplied — fail closed otherwise.
func validateNixSpec(nix *agentv1alpha1.NixSpec) error {
	if nix == nil {
		return nil
	}
	if !sharednix.IsImmutableRevision(nix.NixpkgsRev) {
		return fmt.Errorf("nix toolchain requires nixpkgsRev as a complete 40- or 64-character hexadecimal revision")
	}
	hasPackages := len(nix.Packages) > 0
	hasFlake := nix.FlakeRef != ""
	if hasPackages && hasFlake {
		return fmt.Errorf("nix toolchain: packages and flakeRef are mutually exclusive")
	}
	if !hasPackages && !hasFlake {
		return fmt.Errorf("nix toolchain requires a source: packages or flakeRef")
	}
	if nix.Image == "" {
		return fmt.Errorf("nix toolchain requires a pre-built pinned image (build-time provisioning)")
	}
	if err := validator.ValidatePinnedImageReference(nix.Image); err != nil {
		return fmt.Errorf("nix toolchain image must use an immutable @sha256: digest")
	}
	return nil
}
