package runconfig

import (
	"encoding/json"

	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/merge"
	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/section/run"
	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/spec"
)

// Key names one key of the run section, as the member name of agentic.json.
// The run section of shared-config-go declares the rule of each key and the
// layers that may set it.
type Key string

// The keys of the run section.
const (
	KeyAgent     Key = "agent"
	KeyProvider  Key = "provider"
	KeyModels    Key = "models"
	KeyIsolation Key = "isolation"
	KeyProvision Key = "provision"
	KeyNetwork   Key = "network"
	KeyHomeDir   Key = "homeDir"

	KeyBindPaths   Key = "bindPaths"
	KeyRoBindPaths Key = "roBindPaths"
	KeyCustomEnv   Key = "customEnv"
	KeyAllowHosts  Key = "allowHosts"
	KeyCredentials Key = "credentials"

	KeyRequirePinnedProvision      Key = "requirePinnedProvision"
	KeyRequireHostToolsUnreachable Key = "requireHostToolsUnreachable"
	KeyRequireEgressRestricted     Key = "requireEgressRestricted"
	KeyRequireKernelIsolation      Key = "requireKernelIsolation"
)

// pointer is the JSON pointer of the key inside the run section.
func (k Key) pointer() string { return "/" + string(k) }

// keyOf returns the key a run section declaration names.
func keyOf(declared spec.Key) Key { return Key(declared.Pointer[1:]) }

// Keys is a set of keys.
type Keys map[Key]struct{}

// KeysOf returns the set of keys.
func KeysOf(keys ...Key) Keys {
	set := Keys{}
	for _, key := range keys {
		set[key] = struct{}{}
	}
	return set
}

// Has reports whether key is in the set.
func (k Keys) Has(key Key) bool { _, ok := k[key]; return ok }

// Text is the command line value of a Replace key.
func Text(key Key, value string) merge.Flag { return flag(key, value) }

// List is the command line items of an Append key.
func List(key Key, items []string) merge.Flag { return flag(key, items) }

// Demand is the command line demand of a Require key.
func Demand(key Key) merge.Flag { return flag(key, true) }

func flag(key Key, value any) merge.Flag {
	// A string, a string slice and a boolean always encode.
	raw, _ := json.Marshal(value)
	return merge.Flag{Section: run.Name, Pointer: key.pointer(), Value: raw}
}
