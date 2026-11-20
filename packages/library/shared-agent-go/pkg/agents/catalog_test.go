package agents

import (
	"errors"
	"reflect"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/jsonschema"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/schemas"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/wiring"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/types"
)

func validator(t *testing.T) jsonschema.Validator {
	t.Helper()
	v, err := wiring.NewValidator()
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func TestEveryEmbeddedFileMatchesTheHarnessSchema(t *testing.T) {
	v := validator(t)
	entries, err := embedded.ReadDir(harnessDir)
	if err != nil || len(entries) == 0 {
		t.Fatalf("ReadDir() = %v, %v", entries, err)
	}
	for _, entry := range entries {
		data, err := embedded.ReadFile(harnessDir + "/" + entry.Name())
		if err != nil {
			t.Fatal(err)
		}
		violations, err := v.Validate(schemas.Harness, data)
		if err != nil || len(violations) > 0 {
			t.Errorf("%s: %v, %v", entry.Name(), violations, err)
		}
	}
}

func TestLoadFailsClosed(t *testing.T) {
	valid := `{"name":"h","displayName":"H","binary":"h","nixPackage":"h","access":"provider","versionArgs":["-v"]}`
	cases := map[string]fstest.MapFS{
		"no directory":   {},
		"unreadable":     {"d/h.json": {Mode: 0o20000000000}},
		"not JSON":       {"d/h.json": {Data: []byte(`{`)}},
		"unknown member": {"d/h.json": {Data: []byte(`{"name":"h","argv":[]}`)}},
		"other name":     {"d/g.json": {Data: []byte(valid)}},
		"escaping path":  {"d/h.json": {Data: []byte(`{"name":"h","userConfigPaths":["../x"]}`)}},
		"absolute home":  {"d/h.json": {Data: []byte(`{"name":"h","configHome":{"env":"H","dir":"/h"}}`)}},
		"backslash path": {"d/h.json": {Data: []byte(`{"name":"h","instructionFile":"a\\b"}`)}},
	}
	for name, fsys := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := load(fsys, "d"); err == nil {
				t.Fatal("load() must fail")
			}
		})
	}
	catalog, err := load(fstest.MapFS{"d/h.json": {Data: []byte(valid)}, "d/README.md": {Data: []byte("x")}}, "d")
	if err != nil || !reflect.DeepEqual(catalog.Types(), []types.AgentType{"h"}) {
		t.Fatalf("load() = %v, %v", catalog.Types(), err)
	}
}

func TestMustPanicsOnABrokenBuild(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("must() did not panic")
		}
	}()
	must(Catalog{}, errors.New("broken"))
}

func TestGetCopiesEveryNestedValue(t *testing.T) {
	first := agentOf(t, types.AgentCodex)
	first.ConfigHome.Dir = "changed"
	first.Headless.Options["model"][0] = "changed"
	first.Headless.Output["stream"][0] = "changed"
	first.Headless.Args[0] = "changed"
	second := agentOf(t, types.AgentCodex)
	if second.ConfigHome.Dir == "changed" || second.Headless.Options["model"][0] == "changed" ||
		second.Headless.Output["stream"][0] == "changed" || second.Headless.Args[0] == "changed" {
		t.Fatal("Get() returned shared nested facts")
	}
}

func TestPluginLayout(t *testing.T) {
	if got := Embedded().PluginLayout(types.AgentCodex); got.Manifest != ".codex-plugin/plugin.json" || got.Marketplace != ".agents/plugins/marketplace.json" {
		t.Errorf("PluginLayout(codex) = %+v", got)
	}
	if got := Embedded().PluginLayout("other"); got != (types.PluginLayout{}) {
		t.Errorf("PluginLayout(other) = %+v, want the zero layout", got)
	}
}

func TestShippedHeadlessFacts(t *testing.T) {
	claude := agentOf(t, types.AgentClaude)
	if claude.InstructionFile != "CLAUDE.md" || claude.Plugin.Manifest != ".claude-plugin/plugin.json" || claude.Plugin.SkillsDir != "skills" {
		t.Errorf("claude facts = %+v", claude)
	}
	codex := agentOf(t, types.AgentCodex)
	if codex.InstructionFile != "AGENTS.md" || codex.SkillsDir != ".agents/skills" || *codex.ConfigHome != (types.ConfigHome{Env: "CODEX_HOME", Dir: ".codex"}) {
		t.Errorf("codex facts = %+v", codex)
	}
	if got := agentOf(t, types.AgentClaude).UserConfigPaths; !reflect.DeepEqual(got, []string{".claude", ".claude.json"}) {
		t.Errorf("claude user configuration = %v", got)
	}
	if got := agentOf(t, types.AgentOpencode).UserConfigPaths; !reflect.DeepEqual(got, []string{".config/opencode"}) {
		t.Errorf("opencode user configuration = %v", got)
	}
}

func TestHeadlessArgs(t *testing.T) {
	agent := &types.AgentConfig{Type: "h", Headless: types.Headless{
		Args: []string{"-p"}, HermeticArgs: []string{"--clean"}, Output: map[string][]string{"stream": {"--json"}},
	}}
	got, err := HeadlessArgs(agent, "stream")
	if want := []string{"-p", "--json", "--clean"}; err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("HeadlessArgs() = %v, %v, want %v", got, err, want)
	}
	if _, err := HeadlessArgs(agent, "json"); err == nil || !strings.Contains(err.Error(), "no json output") {
		t.Fatalf("HeadlessArgs(json) error = %v", err)
	}
}

func TestPromptArgs(t *testing.T) {
	agent := &types.AgentConfig{Type: "h", Headless: types.Headless{Args: []string{"-p"}, HermeticArgs: []string{"--clean"}}}
	got, err := PromptArgs(agent, "fix the tests")
	if want := []string{"-p", "fix the tests"}; err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("PromptArgs() = %v, %v, want %v", got, err, want)
	}
	got[0] = "changed"
	if agent.Headless.Args[0] != "-p" {
		t.Error("PromptArgs() shares the arguments of the facts")
	}
	if _, err := PromptArgs(agent, "--help"); err == nil || !strings.Contains(err.Error(), "dash") {
		t.Errorf("PromptArgs(--help) error = %v, want a dash refused", err)
	}
	interactive := &types.AgentConfig{Type: "i"}
	if _, err := PromptArgs(interactive, "fix"); err == nil || !strings.Contains(err.Error(), "harness i declares no non-interactive run") {
		t.Errorf("PromptArgs(interactive) error = %v", err)
	}
}

func TestOption(t *testing.T) {
	agent := &types.AgentConfig{Type: "h", Headless: types.Headless{Options: map[string][]string{
		"effort": {"-c", `effort="{value}"`}, "readOnly": {"--sandbox", "read-only"},
	}}}
	if got, err := Option(agent, "effort", "high"); err != nil || !reflect.DeepEqual(got, []string{"-c", `effort="high"`}) {
		t.Errorf("Option(effort) = %v, %v", got, err)
	}
	if got, err := Option(agent, "readOnly", ""); err != nil || !reflect.DeepEqual(got, []string{"--sandbox", "read-only"}) {
		t.Errorf("Option(readOnly) = %v, %v", got, err)
	}
	if _, err := Option(agent, "model", "m"); err == nil || !strings.Contains(err.Error(), "no model option") {
		t.Errorf("Option(model) error = %v", err)
	}
}

func TestOptions(t *testing.T) {
	agent := &types.AgentConfig{Type: "h", Headless: types.Headless{Options: map[string][]string{
		"model": {"--model", "{value}"}, "turns": {"--turns={value}"},
	}}}
	got, err := Options(agent, Setting{Name: "model", Value: "m"}, Setting{Name: "turns", Value: "2"})
	if want := []string{"--model", "m", "--turns=2"}; err != nil || !reflect.DeepEqual(got, want) {
		t.Errorf("Options() = %v, %v, want %v", got, err, want)
	}
	if _, err := Options(agent, Setting{Name: "model", Value: "m"}, Setting{Name: "effort", Value: "high"}); err == nil {
		t.Error("Options() with an undeclared option must fail")
	}
}

func TestSkillPrompt(t *testing.T) {
	agent := &types.AgentConfig{Type: "h", Headless: types.Headless{SkillPrompt: "/{skill} {prompt}"}}
	if got, err := SkillPrompt(agent, "review", "check {skill}"); err != nil || got != "/review check {skill}" {
		t.Errorf("SkillPrompt() = %q, %v", got, err)
	}
	if _, err := SkillPrompt(&types.AgentConfig{Type: "h"}, "s", "p"); err == nil {
		t.Error("SkillPrompt() without a template must fail")
	}
}

func TestApplyOverridesOnlyTheNamedMembers(t *testing.T) {
	base := Embedded()
	next, err := base.Apply(validator(t),
		Override{Name: types.AgentClaude, Origin: "user", Data: []byte(`{"$schema":"x","binary":"/opt/claude","headless":{"options":{"maxTurns":["--turns","{value}"]}}}`)},
		Override{Name: types.AgentClaude, Origin: "managed", Data: []byte(`{"name":"claude","sandboxArgs":["--safe"]}`)},
	)
	if err != nil {
		t.Fatal(err)
	}
	got := agentOfCatalog(t, next, types.AgentClaude)
	if got.Binary != "/opt/claude" || !reflect.DeepEqual(got.SandboxArgs, []string{"--safe"}) {
		t.Errorf("overridden facts = %+v", got)
	}
	if !reflect.DeepEqual(got.Headless.Options["maxTurns"], []string{"--turns", "{value}"}) || !reflect.DeepEqual(got.Headless.Options["model"], []string{"--model", "{value}"}) {
		t.Errorf("options = %v, want maxTurns replaced and model kept", got.Headless.Options)
	}
	if got.NixPackage != "claude-code" || got.InstructionFile != "CLAUDE.md" {
		t.Errorf("unnamed members changed: %+v", got)
	}
	if original := agentOfCatalog(t, base, types.AgentClaude); original.Binary != "claude" {
		t.Errorf("Apply() changed the base catalog: %+v", original)
	}
}

func TestApplyFailsClosed(t *testing.T) {
	cases := map[string]struct {
		override Override
		want     string
	}{
		"unknown harness":   {Override{Name: "other", Origin: "o", Data: []byte(`{}`)}, "no harness"},
		"unknown member":    {Override{Name: types.AgentClaude, Origin: "o", Data: []byte(`{"argv":[]}`)}, "invalid harness facts"},
		"not JSON":          {Override{Name: types.AgentClaude, Origin: "o", Data: []byte(`{`)}, "invalid harness facts"},
		"other name":        {Override{Name: types.AgentClaude, Origin: "o", Data: []byte(`{"name":"codex"}`)}, "names codex"},
		"escaping path":     {Override{Name: types.AgentClaude, Origin: "o", Data: []byte(`{"userConfigPaths":["a/../../b"]}`)}, "leaves the directory"},
		"removes the probe": {Override{Name: types.AgentClaude, Origin: "o", Data: []byte(`{"versionArgs":[]}`)}, "invalid harness facts"},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := Embedded().Apply(validator(t), c.override)
			if err == nil || !strings.Contains(err.Error(), c.want) || !strings.Contains(err.Error(), "harness override o") {
				t.Fatalf("Apply() error = %v, want it to name the file and %q", err, c.want)
			}
		})
	}
}

type brokenValidator struct{}

func (brokenValidator) Validate(string, []byte) ([]jsonschema.Violation, error) {
	return nil, jsonschema.ErrUnknownSchema
}

type afterOverride struct{ v jsonschema.Validator }

// Validate accepts the override and rejects the facts it produces, which a
// complete schema does when an override empties a required member.
func (a afterOverride) Validate(id string, document []byte) ([]jsonschema.Violation, error) {
	if id == schemas.Harness {
		return []jsonschema.Violation{{Pointer: "/binary", Message: "rejected"}}, nil
	}
	return a.v.Validate(id, document)
}

func TestApplyReportsTheValidator(t *testing.T) {
	override := Override{Name: types.AgentClaude, Origin: "o", Data: []byte(`{}`)}
	if _, err := Embedded().Apply(brokenValidator{}, override); !errors.Is(err, jsonschema.ErrUnknownSchema) {
		t.Errorf("Apply() error = %v, want the validator fault", err)
	}
	if _, err := Embedded().Apply(afterOverride{validator(t)}, override); err == nil || !strings.Contains(err.Error(), "the facts it produces: invalid harness facts: /binary: rejected") {
		t.Errorf("Apply() error = %v, want the violation of the produced facts", err)
	}
}

func agentOfCatalog(t *testing.T, catalog Catalog, agentType types.AgentType) *types.AgentConfig {
	t.Helper()
	agent, err := catalog.Get(agentType)
	if err != nil {
		t.Fatal(err)
	}
	return agent
}
