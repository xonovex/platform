package config

import (
	"reflect"
	"strings"
	"testing"
)

func input(layer Layer, origin string, entries Entries) Input {
	return Input{Layer: layer, Origin: origin, Entries: entries}
}

func text(value string) Value { return Value{Text: value} }

func TestMergeReplaceFollowsTheLayerPrecedence(t *testing.T) {
	user := input(LayerUser, "/home/u/config", Entries{KeyAgent: text("codex"), KeyProvider: text("glm")})
	project := input(LayerProject, "/repo/.agent-cli.yaml", Entries{KeyAgent: text("opencode")})

	resolved, err := Merge([]Input{user, project})
	if err != nil {
		t.Fatalf("Merge() error = %v", err)
	}
	if got := resolved.Text(KeyAgent); got != "opencode" {
		t.Errorf("agent = %q, want the project value", got)
	}
	if got := resolved.Text(KeyProvider); got != "glm" {
		t.Errorf("provider = %q, want the user value", got)
	}
	if got := resolved.Layers(); !reflect.DeepEqual(got[KeyAgent], []Layer{LayerProject}) || !reflect.DeepEqual(got[KeyProvider], []Layer{LayerUser}) {
		t.Errorf("layers = %v, want the deciding layer of each replace key", got)
	}

	flags := input(LayerFlags, "command line", Entries{KeyAgent: text("pi")})
	resolved, err = Merge([]Input{user, project, flags})
	if err != nil || resolved.Text(KeyAgent) != "pi" {
		t.Errorf("agent = %q (%v), want the command line to win over the files", resolved.Text(KeyAgent), err)
	}
	if resolved.Has(KeyIsolation) || resolved.Text(KeyIsolation) != "" {
		t.Error("an unset key must be absent")
	}
}

func TestMergeManagedWinsAndRefusesAContradictingFlag(t *testing.T) {
	managed := input(LayerManaged, "/etc/agent-cli/managed.yaml", Entries{KeyIsolation: text("docker")})
	user := input(LayerUser, "/home/u/config", Entries{KeyIsolation: text("none")})

	resolved, err := Merge([]Input{managed, user})
	if err != nil || resolved.Text(KeyIsolation) != "docker" {
		t.Fatalf("isolation = %q (%v), want the managed value over the user file", resolved.Text(KeyIsolation), err)
	}

	same := input(LayerFlags, "command line", Entries{KeyIsolation: text("docker")})
	if _, err := Merge([]Input{managed, same}); err != nil {
		t.Errorf("a flag equal to the managed value must pass, got %v", err)
	}

	other := input(LayerFlags, "command line", Entries{KeyIsolation: text("none")})
	_, err = Merge([]Input{managed, other})
	if err == nil || !strings.Contains(err.Error(), "managed value cannot be changed") || !strings.Contains(err.Error(), "/etc/agent-cli/managed.yaml") {
		t.Errorf("Merge() error = %v, want the contradiction named with its managed source", err)
	}
}

func TestMergeAppendKeepsEveryLayerWithTheManagedLast(t *testing.T) {
	user := input(LayerUser, "u", Entries{KeyCustomEnv: {Items: []string{"A=user", "B=user"}}})
	flags := input(LayerFlags, "command line", Entries{KeyCustomEnv: {Items: []string{"A=flag", "B=user"}}})
	managed := input(LayerManaged, "m", Entries{KeyCustomEnv: {Items: []string{"A=managed"}}})

	resolved, err := Merge([]Input{managed, flags, user})
	if err != nil {
		t.Fatalf("Merge() error = %v", err)
	}
	want := []string{"A=user", "B=user", "A=flag", "A=managed"}
	if got := resolved.Items(KeyCustomEnv); !reflect.DeepEqual(got, want) {
		t.Errorf("customEnv = %v, want %v (duplicates removed, managed last)", got, want)
	}
	if got := resolved.Items(KeyBindPaths); got != nil {
		t.Errorf("an unset list = %v, want nil", got)
	}
	layers := resolved.Layers()[KeyCustomEnv]
	if !reflect.DeepEqual(layers, []Layer{LayerUser, LayerFlags, LayerManaged}) {
		t.Errorf("layers = %v, want user, flags, managed", layers)
	}
}

func TestMergeRequireCannotBeRelaxed(t *testing.T) {
	managed := input(LayerManaged, "m", Entries{KeyRequireKernelIsolation: {On: true}})
	flags := input(LayerFlags, "command line", Entries{KeyRequireKernelIsolation: {On: false}})
	user := input(LayerUser, "u", Entries{KeyRequireEgressRestricted: {On: true}})

	resolved, err := Merge([]Input{managed, flags, user})
	if err != nil {
		t.Fatalf("Merge() error = %v", err)
	}
	if !resolved.On(KeyRequireKernelIsolation) || !resolved.On(KeyRequireEgressRestricted) {
		t.Error("a guarantee a layer demands must stay demanded")
	}
	if resolved.On(KeyRequirePinnedProvision) {
		t.Error("a guarantee no layer demands must stay off")
	}
	layers := resolved.Layers()
	if !reflect.DeepEqual(layers[KeyRequireKernelIsolation], []Layer{LayerManaged}) || !reflect.DeepEqual(layers[KeyRequireEgressRestricted], []Layer{LayerUser}) {
		t.Errorf("layers = %v, want the layer that demanded each guarantee", layers)
	}
}

func TestMergeRefusesAKeyThatALayerMayNotSet(t *testing.T) {
	cases := map[string]Input{
		"project bind":       input(LayerProject, "/repo/.agent-cli.yaml", Entries{KeyBindPaths: {Items: []string{"/home/u/.ssh"}}}),
		"project credential": input(LayerProject, "/repo/.agent-cli.yaml", Entries{KeyCredentials: {Items: []string{"T=pass:x"}}}),
		"project isolation":  input(LayerProject, "/repo/.agent-cli.yaml", Entries{KeyIsolation: text("none")}),
	}
	for name, in := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := Merge([]Input{in}); err == nil || !strings.Contains(err.Error(), "may not set") {
				t.Errorf("Merge() error = %v, want a refusal naming the layer", err)
			}
		})
	}
	if _, err := Merge([]Input{input(LayerUser, "u", Entries{"colour": text("red")})}); err == nil || !strings.Contains(err.Error(), "unknown key") {
		t.Errorf("Merge(unknown key) error = %v, want an unknown key error", err)
	}
	if _, err := Merge([]Input{input(LayerProject, "p", Entries{KeyRequirePinnedProvision: {On: true}, KeyAgent: text("codex")})}); err != nil {
		t.Errorf("a project may require a guarantee and select an agent, got %v", err)
	}
}

func TestProvenanceNamesTheSourceOfEveryValue(t *testing.T) {
	resolved, err := Merge([]Input{
		input(LayerManaged, "/etc/agent-cli/managed.yaml", Entries{KeyNetwork: text("managed"), KeyRequireEgressRestricted: {On: true}}),
		input(LayerUser, "/home/u/config", Entries{KeyAgent: text("codex"), KeyAllowHosts: {Items: []string{"example.com"}}}),
	})
	if err != nil {
		t.Fatalf("Merge() error = %v", err)
	}
	want := []string{
		"agent: codex (user /home/u/config)",
		"network: managed (managed /etc/agent-cli/managed.yaml)",
		"allowHosts: example.com (user /home/u/config)",
		"requireEgressRestricted: required (managed /etc/agent-cli/managed.yaml)",
	}
	if got := resolved.Provenance(); !reflect.DeepEqual(got, want) {
		t.Errorf("Provenance() = %v, want %v", got, want)
	}
	if got := (Source{Layer: LayerFlags}).String(); got != "flags" {
		t.Errorf("Source.String() = %q, want the bare layer", got)
	}
	wantSources := []Source{{LayerUser, "/home/u/config"}, {LayerManaged, "/etc/agent-cli/managed.yaml"}}
	if got := resolved.Sources(); !reflect.DeepEqual(got, wantSources) {
		t.Errorf("Sources() = %v, want %v", got, wantSources)
	}
}

func TestEveryKeyHasASpecAndAnEntry(t *testing.T) {
	file := FileConfig{
		Agent: "a", Provider: "p", Models: []string{"m=1"}, Isolation: "i", Provision: "v", Network: "n", HomeDir: "h",
		BindPaths: []string{"b"}, RoBindPaths: []string{"r"}, CustomEnv: []string{"E=1"}, AllowHosts: []string{"x"}, Credentials: []string{"C=env:X"},
		RequirePinnedProvision: true, RequireHostToolsUnreachable: true, RequireEgressRestricted: true, RequireKernelIsolation: true,
	}
	entries := file.Entries()
	if len(entries) != len(Specs) {
		t.Fatalf("FileConfig sets %d keys, Specs declares %d", len(entries), len(Specs))
	}
	for _, spec := range Specs {
		if _, ok := entries[spec.Key]; !ok {
			t.Errorf("FileConfig.Entries() lacks the key %s", spec.Key)
		}
	}
	if len((FileConfig{}).Entries()) != 0 {
		t.Error("an empty file must set no key")
	}
}

func TestKeysIsAnImmutableSet(t *testing.T) {
	base := KeysOf(KeyAgent)
	extended := base.With(KeyNetwork)

	if !base.Has(KeyAgent) || base.Has(KeyNetwork) {
		t.Errorf("With changed the set it was called on: %v", base)
	}
	if !extended.Has(KeyAgent) || !extended.Has(KeyNetwork) || extended.Has(KeyProvider) {
		t.Errorf("With = %v, want the old keys and the new one", extended)
	}
	var none Keys
	if none.Has(KeyAgent) || !none.With(KeyAgent).Has(KeyAgent) {
		t.Error("the zero set must be an empty set that can grow")
	}
}
