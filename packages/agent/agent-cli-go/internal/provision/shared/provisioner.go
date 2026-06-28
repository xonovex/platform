// Package shared is the provision axis core: the Provisioner port plus the
// neutral Input the composition root hands every provisioner. It names no
// concrete provisioner (none/command/nix). Input carries the shared-module nix
// source type, so this core depends on no CLI leaf.
package shared

import (
	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/provision"
	sharednix "github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/provision/nix"
)

// Input is the neutral per-run carrier handed to Contribute. Each provisioner
// reads only the field it needs (none reads nothing); it holds no method-specific
// behavior flag. NixSource is the shared-module value type, keeping this core free
// of any CLI provision leaf.
type Input struct {
	InitCommands []string
	NixSource    sharednix.NixSource
}

// Provisioner produces a Contribution (tools/binds/env/init) that an Isolator
// applies, and declares its own guarantees so the policy engine never names it.
type Provisioner interface {
	Contribute(in Input) (provision.Contribution, error)
	// Pinned reports whether provisioning comes from a pinned source (so
	// RequirePinnedProvision can be honored).
	Pinned() bool
}
