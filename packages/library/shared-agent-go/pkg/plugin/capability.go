package plugin

import (
	"fmt"
	"slices"
	"strings"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/artifact"
)

// Capability is a kind of active configuration a plugin can carry besides its
// skills, agents and commands. A plugin that carries one is rejected until the
// user grants it by name.
type Capability string

const (
	CapabilityHooks    Capability = "hooks"
	CapabilityMCP      Capability = "mcp"
	CapabilityLSP      Capability = "lsp"
	CapabilitySettings Capability = "settings"
)

// Capabilities lists every capability in order.
func Capabilities() []Capability {
	return []Capability{CapabilityHooks, CapabilityLSP, CapabilityMCP, CapabilitySettings}
}

// ParseCapability returns the capability that name denotes.
func ParseCapability(name string) (Capability, error) {
	for _, capability := range Capabilities() {
		if string(capability) == name {
			return capability, nil
		}
	}
	return "", fmt.Errorf("unknown capability %q; valid values: hooks, lsp, mcp, settings", name)
}

// Detect lists the capabilities the files and the manifest keys of a plugin
// carry, in order.
func Detect(tree artifact.Tree, manifest Manifest) []Capability {
	found := map[Capability]bool{}
	for _, file := range tree {
		switch {
		case strings.HasPrefix(file.Path, "hooks/"), strings.HasSuffix(file.Path, "hooks.json"):
			found[CapabilityHooks] = true
		case file.Path == ".mcp.json", file.Path == "mcp.json":
			found[CapabilityMCP] = true
		case file.Path == ".lsp.json":
			found[CapabilityLSP] = true
		case file.Path == "settings.json":
			found[CapabilitySettings] = true
		}
	}
	for _, key := range manifest.Keys {
		switch key {
		case "hooks":
			found[CapabilityHooks] = true
		case "mcpServers":
			found[CapabilityMCP] = true
		case "lspServers":
			found[CapabilityLSP] = true
		}
	}
	var detected []Capability
	for _, capability := range Capabilities() {
		if found[capability] {
			detected = append(detected, capability)
		}
	}
	return detected
}

// Authorize fails when the plugin carries a capability that the user did not
// grant. It names every capability that is missing a grant.
func Authorize(carried, granted []Capability) error {
	var missing []string
	for _, capability := range carried {
		if !slices.Contains(granted, capability) {
			missing = append(missing, string(capability))
		}
	}
	if len(missing) == 0 {
		return nil
	}
	return fmt.Errorf("the plugin carries %s; grant each one with --allow, or leave the plugin out", strings.Join(missing, ", "))
}

// CheckApproval fails unless the user approved exactly the digest of the content
// that was fetched. An empty approval is the preview step: it names the digest to
// approve.
func CheckApproval(digest, approved string) error {
	switch {
	case approved == "":
		return fmt.Errorf("approval required: review the plugin, then repeat the command with --approve-digest %s", digest)
	case !artifact.ValidDigest(approved):
		return fmt.Errorf("the approved digest %q is not a sha256 digest", approved)
	case approved != digest:
		return fmt.Errorf("the approved digest %s differs from the digest of the fetched content %s", approved, digest)
	}
	return nil
}
