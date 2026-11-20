package openshell

import (
	"encoding/json"
	"fmt"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/egress"
	netshared "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/network/shared"
)

// The values that decide how strongly the gateway holds a sandbox. The gateway
// defaults an endpoint to audit, which allows a request and only logs it, and
// Landlock to best effort, which runs the sandbox unconfined where the kernel
// lacks the feature, so the policy names the strict value of each.
const (
	enforcement    = "enforce"
	landlockStrict = "hard_requirement"
	policyName     = "egress"
	anyBinary      = "/**"
	nullDevice     = "/dev/null"
)

// systemPaths are the paths a gateway adds to a policy that has a network rule,
// because the proxy it starts for one needs them. A policy with no network rule
// gets none of them, and a sandbox without them cannot start its entrypoint, so
// the document lists them for that policy.
func systemPaths() (readOnly, readWrite []string) {
	return []string{"/dev/urandom", "/etc", "/lib", "/proc", "/usr", "/var/log"}, []string{"/tmp"}
}

type policyDocument struct {
	Version         int                      `json:"version"`
	Filesystem      filesystemPolicy         `json:"filesystem_policy"`
	Landlock        landlockPolicy           `json:"landlock"`
	NetworkPolicies map[string]networkPolicy `json:"network_policies"`
}

type filesystemPolicy struct {
	IncludeWorkdir bool     `json:"include_workdir"`
	ReadOnly       []string `json:"read_only,omitempty"`
	ReadWrite      []string `json:"read_write,omitempty"`
}

type landlockPolicy struct {
	Compatibility string `json:"compatibility"`
}

type networkPolicy struct {
	Name      string     `json:"name"`
	Endpoints []endpoint `json:"endpoints"`
	Binaries  []binary   `json:"binaries"`
}

type binary struct {
	Path string `json:"path"`
}

type endpoint struct {
	Host        string `json:"host"`
	Port        int    `json:"port"`
	Protocol    string `json:"protocol"`
	Enforcement string `json:"enforcement"`
	Access      string `json:"access"`
}

// renderPolicy returns the sandbox policy document of a run. A none run has an
// empty network policy, which the gateway reads as deny all; a managed run lists
// each allowlist rule as an endpoint that any binary may reach.
func renderPolicy(mode netshared.Mode, allow egress.Allowlist) ([]byte, error) {
	document := policyDocument{
		Version:         1,
		Filesystem:      filesystemPolicy{IncludeWorkdir: true, ReadWrite: []string{nullDevice}},
		Landlock:        landlockPolicy{Compatibility: landlockStrict},
		NetworkPolicies: map[string]networkPolicy{},
	}
	switch mode {
	case netshared.ModeNone:
		readOnly, readWrite := systemPaths()
		document.Filesystem.ReadOnly = readOnly
		document.Filesystem.ReadWrite = append(document.Filesystem.ReadWrite, readWrite...)
	case netshared.ModeManaged:
		rules := allow.Rules()
		if len(rules) == 0 {
			return nil, fmt.Errorf("a managed policy needs an allowlist")
		}
		endpoints := make([]endpoint, 0, len(rules))
		for _, rule := range rules {
			host := rule.Host
			if rule.Subdomains {
				host = "*." + host
			}
			endpoints = append(endpoints, endpoint{Host: host, Port: rule.Port, Protocol: "rest", Enforcement: enforcement, Access: "full"})
		}
		document.NetworkPolicies[policyName] = networkPolicy{Name: policyName, Endpoints: endpoints, Binaries: []binary{{Path: anyBinary}}}
	default:
		return nil, fmt.Errorf("%w: a sandbox policy cannot express network mode %q; use none or managed", ErrUnsupported, mode)
	}
	return json.MarshalIndent(document, "", "  ")
}
