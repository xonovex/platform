package seatbelt

import (
	"fmt"
	"net"
	"sort"
	"strconv"
	"strings"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/network"
)

// systemWritable are the places every process writes: the null device, the
// terminal, the descriptor directory, and the temporary directories of macOS.
var systemWritable = []string{"/dev/null", "/dev/tty", "/dev/dtracehelper", "/dev/fd", "/private/tmp", "/private/var/folders"}

// profileSpec is what a sandbox profile is rendered from. Every path is absolute
// and canonical, because the sandbox matches a rule against the resolved path of
// the file a process opens, so a rule for a symbolic link never applies.
type profileSpec struct {
	// Writable are the directories and files the agent may write besides the
	// system places: the work directory, the binds, and the agent configuration.
	Writable []string
	// Home is the home directory whose contents are unreadable except for
	// Readable.
	Home string
	// Readable are the places inside Home that the run names and the agent reads.
	Readable []string
	Network  network.NetworkMethod
	// ProxyAddress is the host:port of the loopback proxy, for a managed run.
	ProxyAddress string
}

// quote renders path as an SBPL string literal. It rejects a character that could
// end the string and start a rule, and never escapes one.
func quote(path string) (string, error) {
	if path == "" || !strings.HasPrefix(path, "/") {
		return "", fmt.Errorf("profile path %q is not absolute", path)
	}
	for _, r := range path {
		if r == '"' || r == '\\' || r < 0x20 || r == 0x7f {
			return "", fmt.Errorf("profile path %q has a character the profile syntax cannot carry", path)
		}
	}
	return `"` + path + `"`, nil
}

func rule(action, operation, kind string, paths []string) (string, error) {
	if len(paths) == 0 {
		return "", nil
	}
	var b strings.Builder
	b.WriteString("(" + action + " " + operation)
	for _, path := range paths {
		q, err := quote(path)
		if err != nil {
			return "", err
		}
		b.WriteString(" (" + kind + " " + q + ")")
	}
	b.WriteString(")\n")
	return b.String(), nil
}

// renderProfile returns the Seatbelt profile of a run. It starts from allowing
// everything and then takes away every file write and, for the none and managed
// modes, every network operation, and adds back only what the run names. The
// sandbox applies the last matching rule, so the order of the rules below is the
// policy.
func renderProfile(spec profileSpec) (string, error) {
	var b strings.Builder
	b.WriteString("(version 1)\n(allow default)\n(deny file-write*)\n")

	subpaths := append([]string(nil), spec.Writable...)
	sort.Strings(subpaths)
	var literals []string
	for _, path := range systemWritable {
		if strings.HasPrefix(path, "/dev/") && path != "/dev/fd" {
			literals = append(literals, path)
		} else {
			subpaths = append(subpaths, path)
		}
	}
	allowLiterals, err := rule("allow", "file-write*", "literal", literals)
	if err != nil {
		return "", err
	}
	allowSubpaths, err := rule("allow", "file-write*", "subpath", subpaths)
	if err != nil {
		return "", err
	}
	b.WriteString(allowLiterals + allowSubpaths)

	if spec.Home != "" {
		hidden, err := rule("deny", "file-read-data", "subpath", []string{spec.Home})
		if err != nil {
			return "", err
		}
		readable := append([]string(nil), spec.Readable...)
		sort.Strings(readable)
		back, err := rule("allow", "file-read-data", "subpath", readable)
		if err != nil {
			return "", err
		}
		b.WriteString(hidden + back)
	}

	network, err := networkRules(spec)
	if err != nil {
		return "", err
	}
	b.WriteString(network)
	return b.String(), nil
}

// networkRules returns the network section of the profile. A managed run may
// connect only to the loopback port of the proxy, so a direct socket to any other
// address is denied by the kernel.
func networkRules(spec profileSpec) (string, error) {
	switch spec.Network {
	case network.NetworkHost:
		return "", nil
	case network.NetworkNone:
		return "(deny network*)\n", nil
	case network.NetworkManaged:
		_, port, err := net.SplitHostPort(spec.ProxyAddress)
		if err != nil {
			return "", fmt.Errorf("managed egress needs the proxy address: %w", err)
		}
		if number, err := strconv.Atoi(port); err != nil || number < 1 || number > 65535 {
			return "", fmt.Errorf("proxy port %q is not a port", port)
		}
		return `(deny network*)` + "\n" + `(allow network-outbound (remote ip "localhost:` + port + `"))` + "\n", nil
	default:
		return "", fmt.Errorf("unknown network mode %q", spec.Network)
	}
}
