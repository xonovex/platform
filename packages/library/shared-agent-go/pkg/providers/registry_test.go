package providers

import (
	"errors"
	"maps"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/types"
)

func preset(t *testing.T, document string) Preset {
	t.Helper()
	parsed, err := ParsePreset([]byte(document))
	if err != nil {
		t.Fatalf("ParsePreset(%s) error = %v", document, err)
	}
	return parsed
}

func doc(t *testing.T, layer, origin, document string) Document {
	t.Helper()
	return Document{Preset: preset(t, document), Layer: layer, Origin: origin, Raw: []byte(document)}
}

const routeDocument = `{
  "name": "route",
  "displayName": "Example route",
  "agent": "claude",
  "portable": true,
  "credential": {"sourceEnv": "HOST_KEY", "targetEnv": "AGENT_KEY"},
  "models": {
    "big": {"id": "big-id", "effort": "high"},
    "small": {"id": "small-id"}
  },
  "environment": {
    "BASE_URL": "https://models.example.test/api",
    "BIG_MODEL": "{model:big:effort}",
    "SMALL_MODEL": "{model:small:effort}",
    "PLAIN": "{model:big}"
  },
  "args": ["--model", "{model:small}"]
}`

func TestResolveExpandsPlaceholders(t *testing.T) {
	catalog := NewCatalog(doc(t, "user", "/u/route.json", routeDocument))

	provider, err := catalog.Resolve("route", types.AgentClaude)

	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	wantEnv := map[string]string{
		"BASE_URL": "https://models.example.test/api", "BIG_MODEL": "big-id(high)", "SMALL_MODEL": "small-id", "PLAIN": "big-id",
	}
	if !reflect.DeepEqual(provider.Environment, wantEnv) {
		t.Errorf("Environment = %v, want %v", provider.Environment, wantEnv)
	}
	if !reflect.DeepEqual(provider.CliArgs, []string{"--model", "small-id"}) {
		t.Errorf("CliArgs = %v", provider.CliArgs)
	}
	if provider.DisplayName != "Example route" || !provider.Portable || provider.CredentialSourceEnv != "HOST_KEY" || provider.CredentialTargetEnv != "AGENT_KEY" {
		t.Errorf("provider = %+v", provider)
	}
	wantSlots := []types.ModelSlot{
		{Name: "big", ID: "big-id", Effort: "high", Source: "/u/route.json"},
		{Name: "small", ID: "small-id", Source: "/u/route.json"},
	}
	if !reflect.DeepEqual(provider.Slots, wantSlots) {
		t.Errorf("Slots = %+v, want %+v", provider.Slots, wantSlots)
	}
}

func TestResolveNamesTheProviderThatIsNotDefined(t *testing.T) {
	_, err := NewCatalog().Resolve("missing", types.AgentClaude)
	if !errors.Is(err, ErrNotDefined) || !strings.Contains(err.Error(), `"missing"`) {
		t.Errorf("Resolve() error = %v, want ErrNotDefined naming the provider", err)
	}
	var zero Catalog
	if _, err := zero.Resolve("x", types.AgentCodex); !errors.Is(err, ErrNotDefined) {
		t.Errorf("the zero catalog error = %v, want ErrNotDefined", err)
	}
	catalog := NewCatalog(doc(t, "user", "f", `{"name":"route","agent":"codex"}`))
	if _, err := catalog.Resolve("route", types.AgentClaude); !errors.Is(err, ErrNotDefined) {
		t.Errorf("a provider of another agent error = %v, want ErrNotDefined", err)
	}
}

func TestResolveWithoutDisplayNameUsesTheName(t *testing.T) {
	provider, err := NewCatalog(doc(t, "user", "f", `{"name":"route","agent":"codex"}`)).Resolve("route", types.AgentCodex)
	if err != nil || provider.DisplayName != "route" || provider.Portable || provider.Environment != nil || provider.CliArgs != nil {
		t.Errorf("Resolve() = %+v, %v", provider, err)
	}
}

func TestResolveAppliesOverridesAndRecordsTheirSource(t *testing.T) {
	catalog := NewCatalog(doc(t, "user", "/u/route.json", routeDocument))

	provider, err := catalog.Resolve("route", types.AgentClaude, Override{Slot: "big", ID: "other-id", Source: "command line"})

	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if provider.Environment["BIG_MODEL"] != "other-id(high)" {
		t.Errorf("BIG_MODEL = %q, want the override with the kept effort", provider.Environment["BIG_MODEL"])
	}
	if provider.Slots[0].Source != "command line" || provider.Slots[1].Source != "/u/route.json" {
		t.Errorf("Slots = %+v", provider.Slots)
	}
}

func TestResolveRejectsAnOverrideOfAnUndefinedSlotOrABadID(t *testing.T) {
	catalog := NewCatalog(doc(t, "user", "f", routeDocument), doc(t, "user", "g", `{"name":"bare","agent":"claude"}`))

	if _, err := catalog.Resolve("route", types.AgentClaude, Override{Slot: "huge", ID: "x"}); err == nil || !strings.Contains(err.Error(), `no model slot "huge"`) || !strings.Contains(err.Error(), "big") {
		t.Errorf("an undefined slot error = %v, want it to list the slots", err)
	}
	if _, err := catalog.Resolve("bare", types.AgentClaude, Override{Slot: "x", ID: "y"}); err == nil || !strings.Contains(err.Error(), "no slot") {
		t.Errorf("a provider without slots error = %v", err)
	}
	if _, err := catalog.Resolve("route", types.AgentClaude, Override{Slot: "big", ID: "has space"}); err == nil || !strings.Contains(err.Error(), "whitespace") {
		t.Errorf("a bad id error = %v", err)
	}
}

func TestResolveFailsOnAPlaceholderForAnUndefinedSlot(t *testing.T) {
	for name, document := range map[string]string{
		"environment": `{"name":"p","agent":"claude","environment":{"A":"{model:ghost}"}}`,
		"args":        `{"name":"p","agent":"claude","args":["{model:ghost}"]}`,
	} {
		_, err := NewCatalog(doc(t, "user", "f", document)).Resolve("p", types.AgentClaude)
		if err == nil || !strings.Contains(err.Error(), `"ghost"`) {
			t.Errorf("%s: error = %v, want the undefined slot named", name, err)
		}
	}
	_, err := NewCatalog(doc(t, "user", "f", `{"name":"p","agent":"claude","args":["{model:Bad}"]}`)).Resolve("p", types.AgentClaude)
	if err == nil || !strings.Contains(err.Error(), "malformed") {
		t.Errorf("a malformed placeholder error = %v", err)
	}
}

func TestExtendsRefinesTheBaseProvider(t *testing.T) {
	catalog := NewCatalog(
		doc(t, "managed", "/m/route.json", routeDocument),
		doc(t, "user", "/u/child.json", `{
			"name": "child", "agent": "claude", "extends": "route", "displayName": "Child", "portable": false,
			"credential": {"sourceEnv": "OTHER", "targetEnv": "AGENT_KEY"},
			"models": {"small": {"id": "child-small"}},
			"environment": {"EXTRA": "{model:small}"}
		}`),
	)

	provider, err := catalog.Resolve("child", types.AgentClaude)

	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if provider.Name != "child" || provider.DisplayName != "Child" || provider.Portable || provider.CredentialSourceEnv != "OTHER" {
		t.Errorf("provider = %+v", provider)
	}
	if provider.Environment["BIG_MODEL"] != "big-id(high)" || provider.Environment["SMALL_MODEL"] != "child-small" || provider.Environment["EXTRA"] != "child-small" {
		t.Errorf("Environment = %v", provider.Environment)
	}
	if !reflect.DeepEqual(provider.CliArgs, []string{"--model", "child-small"}) {
		t.Errorf("CliArgs = %v, want the inherited args over the refined slot", provider.CliArgs)
	}
	if provider.Slots[0].Source != "/m/route.json" || provider.Slots[1].Source != "/u/child.json" {
		t.Errorf("Slots = %+v, want each slot to name the file that decided it", provider.Slots)
	}
	if got := catalog.Layers("child", types.AgentClaude); !reflect.DeepEqual(got, []string{"user", "managed"}) {
		t.Errorf("Layers() = %v", got)
	}
}

func TestExtendsReplacesArgsWhenSet(t *testing.T) {
	catalog := NewCatalog(
		doc(t, "user", "a", routeDocument),
		doc(t, "user", "b", `{"name":"child","agent":"claude","extends":"route","args":[]}`),
	)
	provider, err := catalog.Resolve("child", types.AgentClaude)
	if err != nil || len(provider.CliArgs) != 0 {
		t.Errorf("Resolve() = %+v, %v, want the empty args to replace the inherited ones", provider, err)
	}
}

func TestExtendsFailsOnAMissingBaseAndOnACycle(t *testing.T) {
	missing := NewCatalog(doc(t, "user", "f", `{"name":"child","agent":"claude","extends":"nowhere"}`))
	if _, err := missing.Resolve("child", types.AgentClaude); err == nil || !strings.Contains(err.Error(), `"nowhere"`) {
		t.Errorf("a missing base error = %v", err)
	}
	cycle := NewCatalog(
		doc(t, "user", "a", `{"name":"one","agent":"claude","extends":"two"}`),
		doc(t, "user", "b", `{"name":"two","agent":"claude","extends":"one"}`),
	)
	if _, err := cycle.Resolve("one", types.AgentClaude); err == nil || !strings.Contains(err.Error(), "cycle") {
		t.Errorf("a cycle error = %v", err)
	}
	if got := cycle.Layers("one", types.AgentClaude); len(got) != 2 {
		t.Errorf("Layers() of a cycle = %v, want it to stop after each document once", got)
	}
	if got := missing.Layers("nowhere", types.AgentClaude); got != nil {
		t.Errorf("Layers() of an undefined provider = %v", got)
	}
}

func TestALaterDocumentReplacesAnEarlierOneOfTheSameName(t *testing.T) {
	catalog := NewCatalog(
		doc(t, "user", "u", `{"name":"p","agent":"claude","environment":{"FROM":"user"}}`),
		doc(t, "managed", "m", `{"name":"p","agent":"claude","environment":{"FROM":"managed"}}`),
	)
	provider, err := catalog.Resolve("p", types.AgentClaude)
	if err != nil || provider.Environment["FROM"] != "managed" {
		t.Errorf("Resolve() = %+v, %v, want the later layer to win", provider, err)
	}
}

func TestResolvePortableRefusesALocalRoute(t *testing.T) {
	catalog := NewCatalog(doc(t, "user", "f", routeDocument), doc(t, "user", "g", `{"name":"local","agent":"claude"}`))

	if _, err := catalog.ResolvePortable("route", types.AgentClaude); err != nil {
		t.Errorf("ResolvePortable(portable) error = %v", err)
	}
	if _, err := catalog.ResolvePortable("local", types.AgentClaude); !errors.Is(err, ErrLocalOnly) {
		t.Errorf("ResolvePortable(local) error = %v, want ErrLocalOnly", err)
	}
	if _, err := catalog.ResolvePortable("ghost", types.AgentClaude); !errors.Is(err, ErrNotDefined) {
		t.Errorf("ResolvePortable(ghost) error = %v, want ErrNotDefined", err)
	}
}

func TestNamesAndListAreSorted(t *testing.T) {
	catalog := NewCatalog(
		doc(t, "user", "/u/b.json", `{"name":"b","agent":"claude","portable":true}`),
		doc(t, "managed", "/m/a.json", `{"name":"a","agent":"claude","extends":"b"}`),
		doc(t, "user", "/u/c.json", `{"name":"c","agent":"codex","displayName":"C"}`),
	)

	if got := catalog.Names(types.AgentClaude); !slices.Equal(got, []string{"a", "b"}) {
		t.Errorf("Names(claude) = %v", got)
	}
	if got := NewCatalog().Names(types.AgentClaude); len(got) != 0 {
		t.Errorf("Names() of an empty catalog = %v", got)
	}
	want := []Entry{
		{Name: "a", Agent: "claude", Extends: "b", Layer: "managed", Origin: "/m/a.json"},
		{Name: "b", Agent: "claude", Portable: true, Layer: "user", Origin: "/u/b.json"},
		{Name: "c", DisplayName: "C", Agent: "codex", Layer: "user", Origin: "/u/c.json"},
	}
	if got := catalog.List(); !reflect.DeepEqual(got, want) {
		t.Errorf("List() = %+v, want %+v", got, want)
	}
}

func TestListOrdersTheSameAgentByNameBothWays(t *testing.T) {
	catalog := NewCatalog(
		doc(t, "user", "1", `{"name":"x","agent":"claude"}`),
		doc(t, "user", "2", `{"name":"y","agent":"claude"}`),
		doc(t, "user", "3", `{"name":"x","agent":"codex"}`),
	)
	names := make([]string, 0, 3)
	for _, entry := range catalog.List() {
		names = append(names, string(entry.Agent)+"/"+entry.Name)
	}
	if !slices.Equal(names, []string{"claude/x", "claude/y", "codex/x"}) {
		t.Errorf("List() order = %v", names)
	}
}

func TestGetProviderCliArgsReturnsACopy(t *testing.T) {
	provider := &types.ModelProvider{CliArgs: []string{"a", "b"}}
	args := GetProviderCliArgs(provider)
	args[0] = "changed"
	if provider.CliArgs[0] != "a" {
		t.Error("GetProviderCliArgs() shares the provider slice")
	}
}

func TestSlotNamesAreSorted(t *testing.T) {
	if got := slotNames(map[string]Slot{"b": {ID: "x"}, "a": {ID: "y"}}); got != `["a" "b"]` {
		t.Errorf("slotNames() = %s", got)
	}
	if got := slices.Sorted(maps.Keys(map[string]Slot{})); len(got) != 0 {
		t.Errorf("sorted keys of nothing = %v", got)
	}
}
