package cmd

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/merge"
	evalsection "github.com/xonovex/platform/packages/library/shared-config-go/pkg/section/eval"
)

// guarantees maps the value of --sandbox-require to the key that demands it.
var guarantees = map[string]string{
	"pinned-provision":       "/sandbox/requirePinnedProvision",
	"host-tools-unreachable": "/sandbox/requireHostToolsUnreachable",
	"egress-restricted":      "/sandbox/requireEgressRestricted",
	"kernel-isolation":       "/sandbox/requireKernelIsolation",
}

// sandboxFlag is one command line flag of the sandbox and the key it sets.
type sandboxFlag struct {
	name, pointer, usage string
	list                 bool
}

var sandboxFlags = []sandboxFlag{
	{"sandbox", "/sandbox/isolation", "isolation of every harness run: none, bwrap, docker or seatbelt (default none)", false},
	{"sandbox-provision", "/sandbox/provision", "where the sandbox gets its tools: none, nix or command", false},
	{"sandbox-network", "/sandbox/network", "network egress of the sandbox: host, none or managed", false},
	{"sandbox-image", "/sandbox/image", "container image of a docker sandbox, which holds the harness program", false},
	{"sandbox-runtime", "/sandbox/runtime", "container runtime of a docker sandbox, such as runsc", false},
	{"sandbox-nix-rev", "/sandbox/nixRev", "pinned nixpkgs revision of a nix provision", false},
	{"sandbox-init-command", "/sandbox/initCommands", "command a command provision runs before the harness (repeatable)", true},
	{"sandbox-nix-package", "/sandbox/nixPackages", "package a nix provision adds (repeatable)", true},
	{"sandbox-allow-host", "/sandbox/allowHosts", "host a managed network may reach (repeatable)", true},
	{"sandbox-credential", "/sandbox/credentials", "credential binding of the harness, as NAME=scheme:locator (repeatable)", true},
	{"sandbox-env", "/sandbox/customEnv", "environment variable of the harness, as KEY=VALUE (repeatable)", true},
}

// sandboxOptions holds the sandbox flags of a protocol command.
type sandboxOptions struct {
	texts       map[string]*string
	lists       map[string]*[]string
	passthrough bool
	require     []string
}

func (s *sandboxOptions) bind(c *cobra.Command) {
	f := c.Flags()
	s.texts, s.lists = map[string]*string{}, map[string]*[]string{}
	for _, flag := range sandboxFlags {
		if flag.list {
			s.lists[flag.name] = f.StringArray(flag.name, nil, flag.usage)
			continue
		}
		s.texts[flag.name] = f.String(flag.name, "", flag.usage)
	}
	f.BoolVar(&s.passthrough, "sandbox-passthrough", false, "expose the host tools to a bwrap sandbox, which forfeits host-tools-unreachable")
	f.StringArrayVar(&s.require, "sandbox-require", nil, "guarantee the sandbox must give: "+strings.Join(guaranteeNames(), ", ")+" (repeatable)")
}

// layer returns the flags layer of the sandbox: a key for each flag the user
// set and no other, so an unset flag never hides a file.
func (s *sandboxOptions) layer(c *cobra.Command) ([]merge.Flag, error) {
	var flags []merge.Flag
	for _, flag := range sandboxFlags {
		if !c.Flags().Changed(flag.name) {
			continue
		}
		if flag.list {
			flags = append(flags, sandboxKey(flag.pointer, *s.lists[flag.name]))
			continue
		}
		flags = append(flags, sandboxKey(flag.pointer, *s.texts[flag.name]))
	}
	if c.Flags().Changed("sandbox-passthrough") {
		flags = append(flags, sandboxKey("/sandbox/hostPassthrough", s.passthrough))
	}
	for _, name := range s.require {
		pointer, ok := guarantees[name]
		if !ok {
			return nil, fmt.Errorf("--sandbox-require %s: valid guarantees: %s", name, strings.Join(guaranteeNames(), ", "))
		}
		flags = append(flags, sandboxKey(pointer, true))
	}
	return flags, nil
}

func sandboxKey(pointer string, value any) merge.Flag {
	// A string, a string slice and a boolean always encode.
	raw, _ := json.Marshal(value)
	return merge.Flag{Section: evalsection.Name, Pointer: pointer, Value: raw}
}

func guaranteeNames() []string {
	names := make([]string, 0, len(guarantees))
	for name := range guarantees {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
