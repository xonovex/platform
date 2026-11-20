// Package egress is the pure model of an egress allowlist: which hosts and ports
// a managed run may reach. It denies by default, so an empty allowlist reaches
// nothing, and it does no I/O.
package egress

import (
	"errors"
	"fmt"
	"net/netip"
	"sort"
	"strconv"
	"strings"
)

// ErrEmpty reports an allowlist with no rule, which would allow nothing.
var ErrEmpty = errors.New("the egress allowlist has no host")

// DefaultPort is the port a rule without a port allows: the HTTPS port.
const DefaultPort = 443

// Rule allows one host, or every subdomain of a domain, on one port.
type Rule struct {
	// Host is the lower-case host name or address, without the leading "*." of
	// a wildcard rule.
	Host string
	// Subdomains reports a "*.domain" rule: it allows the names below Host, and
	// not Host itself.
	Subdomains bool
	Port       int
}

// String renders the rule in the form ParseRule reads.
func (r Rule) String() string {
	host := r.Host
	if r.Subdomains {
		host = "*." + host
	}
	return host + ":" + strconv.Itoa(r.Port)
}

// ParseRule reads "host", "host:port" or "*.domain[:port]". A bare host allows
// DefaultPort. It rejects a lone "*", a wildcard over a single label such as
// "*.com", a wildcard on an address, and a host with a path, scheme or space,
// because each would allow far more than the caller can have meant.
func ParseRule(text string) (Rule, error) {
	original := text
	text = strings.TrimSpace(text)
	if text == "" || strings.ContainsAny(text, "/@ \t\r\n") || strings.Contains(text, "://") {
		return Rule{}, fmt.Errorf("egress rule %q is not a host with an optional port", original)
	}
	host, port := text, DefaultPort
	if i := strings.LastIndex(text, ":"); i >= 0 && !strings.Contains(text[i+1:], "]") && strings.Count(text, ":") == 1 {
		host = text[:i]
		parsed, err := strconv.Atoi(text[i+1:])
		if err != nil || parsed < 1 || parsed > 65535 {
			return Rule{}, fmt.Errorf("egress rule %q has an invalid port", original)
		}
		port = parsed
	}
	host = strings.ToLower(strings.TrimSuffix(host, "."))
	rule := Rule{Port: port}
	if rest, ok := strings.CutPrefix(host, "*."); ok {
		rule.Subdomains = true
		host = rest
		if !strings.Contains(host, ".") {
			return Rule{}, fmt.Errorf("egress rule %q is a wildcard over a top-level domain", original)
		}
		if _, err := netip.ParseAddr(host); err == nil {
			return Rule{}, fmt.Errorf("egress rule %q is a wildcard over an address", original)
		}
	}
	if host == "" || strings.Contains(host, "*") {
		return Rule{}, fmt.Errorf("egress rule %q is not a host with an optional port", original)
	}
	rule.Host = host
	return rule, nil
}

// Allowlist is an ordered, de-duplicated set of rules.
type Allowlist struct {
	rules []Rule
}

// NewAllowlist parses texts into an allowlist. It fails on an invalid rule and
// on an empty list, because a managed run with nothing allowed is a mistake in
// the configuration and never what the user asked for.
func NewAllowlist(texts []string) (Allowlist, error) {
	seen := map[Rule]bool{}
	var rules []Rule
	for _, text := range texts {
		rule, err := ParseRule(text)
		if err != nil {
			return Allowlist{}, err
		}
		if !seen[rule] {
			seen[rule] = true
			rules = append(rules, rule)
		}
	}
	if len(rules) == 0 {
		return Allowlist{}, ErrEmpty
	}
	sort.Slice(rules, func(i, j int) bool { return rules[i].String() < rules[j].String() })
	return Allowlist{rules: rules}, nil
}

// Rules returns a copy of the rules in a stable order.
func (a Allowlist) Rules() []Rule { return append([]Rule(nil), a.rules...) }

// Allows reports whether the allowlist lets a connection to host and port
// through. It matches on the lower-case name without a trailing dot.
func (a Allowlist) Allows(host string, port int) bool {
	host = strings.ToLower(strings.TrimSuffix(host, "."))
	for _, rule := range a.rules {
		if rule.Port != port {
			continue
		}
		if rule.Subdomains {
			if strings.HasSuffix(host, "."+rule.Host) {
				return true
			}
			continue
		}
		if host == rule.Host {
			return true
		}
	}
	return false
}
