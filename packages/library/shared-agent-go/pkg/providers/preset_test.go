package providers

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/schemas"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/wiring"
)

func TestParsePresetFailsClosed(t *testing.T) {
	cases := map[string]string{
		"syntax":            `{`,
		"unknown member":    `{"name":"p","agent":"claude","mdl":{}}`,
		"two documents":     `{"name":"p","agent":"claude"} {"name":"q","agent":"claude"}`,
		"no name":           `{"agent":"claude"}`,
		"bad name":          `{"name":"Bad Name","agent":"claude"}`,
		"no agent":          `{"name":"p"}`,
		"bad agent":         `{"name":"p","agent":"Claude"}`,
		"bad extends":       `{"name":"p","agent":"claude","extends":"Up"}`,
		"half credential":   `{"name":"p","agent":"claude","credential":{"sourceEnv":"A"}}`,
		"bad credential":    `{"name":"p","agent":"claude","credential":{"sourceEnv":"A B","targetEnv":"C"}}`,
		"bad slot name":     `{"name":"p","agent":"claude","models":{"Big":{"id":"x"}}}`,
		"empty slot id":     `{"name":"p","agent":"claude","models":{"big":{"id":""}}}`,
		"space in slot id":  `{"name":"p","agent":"claude","models":{"big":{"id":"a b"}}}`,
		"control in the id": `{"name":"p","agent":"claude","models":{"big":{"id":"a\u0001b"}}}`,
		"bad effort":        `{"name":"p","agent":"claude","models":{"big":{"id":"x","effort":"High!"}}}`,
		"bad env key":       `{"name":"p","agent":"claude","environment":{"A-B":"x"}}`,
	}
	for name, document := range cases {
		if _, err := ParsePreset([]byte(document)); err == nil {
			t.Errorf("%s: ParsePreset() error = nil", name)
		}
	}
}

func TestParsePresetReadsEveryMember(t *testing.T) {
	parsed, err := ParsePreset([]byte(`{"$schema":"https://x.test/s.json","name":"p","agent":"codex","extends":"q","models":{"default":{"id":"m","effort":"low"}},"args":["a"]}`))
	if err != nil {
		t.Fatalf("ParsePreset() error = %v", err)
	}
	if parsed.Name != "p" || parsed.Agent != "codex" || parsed.Extends != "q" || parsed.Models["default"].Effort != "low" || parsed.Args[0] != "a" {
		t.Errorf("parsed = %+v", parsed)
	}
}

func TestExpandOnlyTouchesModelPlaceholders(t *testing.T) {
	slots := map[string]Slot{"a": {ID: "id-a", Effort: "high"}, "b": {ID: "id-b"}}
	cases := map[string]string{
		"plain":                   "plain",
		"{model:a}":               "id-a",
		"{model:a:effort}":        "id-a(high)",
		"{model:b:effort}":        "id-b",
		"x={model:a},y={model:b}": "x=id-a,y=id-b",
		"{other:a}":               "{other:a}",
	}
	for input, want := range cases {
		got, err := expand(input, slots)
		if err != nil || got != want {
			t.Errorf("expand(%q) = %q, %v, want %q", input, got, err, want)
		}
	}
	if _, err := expand("{model:}", slots); err == nil {
		t.Error("expand() accepted an empty slot")
	}
	if _, err := expand("{model:a:other}", slots); err == nil {
		t.Error("expand() accepted an unknown placeholder suffix")
	}
}

func TestOverlayKeepsBothMapsAndAllocatesNothingForTwoEmptyOnes(t *testing.T) {
	if got := overlay[string](nil, nil); got != nil {
		t.Errorf("overlay(nil, nil) = %v", got)
	}
	got := overlay(map[string]string{"a": "1", "b": "2"}, map[string]string{"b": "3"})
	if !reflect.DeepEqual(got, map[string]string{"a": "1", "b": "3"}) {
		t.Errorf("overlay() = %v", got)
	}
}

func TestSchemaListsEveryPresetMember(t *testing.T) {
	documents, err := schemas.Documents()
	if err != nil {
		t.Fatal(err)
	}
	var parsed struct {
		Defs struct {
			Route struct {
				Properties map[string]json.RawMessage `json:"properties"`
			} `json:"route"`
		} `json:"$defs"`
	}
	if err := json.Unmarshal(documents[schemas.Provider], &parsed); err != nil {
		t.Fatalf("the schema is not JSON: %v", err)
	}
	properties := parsed.Defs.Route.Properties
	var members []string
	typeOf := reflect.TypeOf(Preset{})
	for index := range typeOf.NumField() {
		members = append(members, strings.Split(typeOf.Field(index).Tag.Get("json"), ",")[0])
	}
	for _, member := range members {
		if _, ok := properties[member]; !ok {
			t.Errorf("the schema has no property %q", member)
		}
	}
	if len(properties) != len(members) {
		t.Errorf("the schema holds %d properties, the Preset type %d", len(properties), len(members))
	}
}

func TestParsePresetAgreesWithTheSchema(t *testing.T) {
	validator, err := wiring.NewValidator()
	if err != nil {
		t.Fatal(err)
	}
	for _, document := range []string{
		`{"name":"p","agent":"claude"}`,
		`{"$schema":"x","name":"p","agent":"codex","extends":"q","portable":true,"credential":{"sourceEnv":"A","targetEnv":"B"},"models":{"main":{"id":"m","effort":"high"}},"environment":{"X":"1"},"args":["-c"]}`,
		`{"agent":"claude"}`,
		`{"name":"p","agent":"claude","mdl":{}}`,
		`{"name":"p","agent":"claude","credential":{"sourceEnv":"A"}}`,
	} {
		violations, err := validator.Validate(schemas.Provider, []byte(document))
		if err != nil {
			t.Fatal(err)
		}
		_, parseErr := ParsePreset([]byte(document))
		if (len(violations) == 0) != (parseErr == nil) {
			t.Errorf("%s: schema violations %v, ParsePreset error %v", document, violations, parseErr)
		}
	}
}

func TestValidateRejectsAndAcceptsDirectly(t *testing.T) {
	if err := (Preset{Name: "p", Agent: "claude"}).Validate(); err != nil {
		t.Errorf("Validate() error = %v", err)
	}
}

func TestHostsNamesTheHostsOfTheRoute(t *testing.T) {
	preset, err := ParsePreset([]byte(`{"name":"proj","agent":"claude","environment":{"BASE_URL":"https://models.example.test:8443/v1","OTHER":"http://127.0.0.1:8317/v1","NOTE":"no url"},"args":["-c","base_url=\"https://args.example.test/x\""]}`))
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"127.0.0.1:8317", "args.example.test", "models.example.test:8443"}
	if got := preset.Hosts(); !reflect.DeepEqual(got, want) {
		t.Errorf("Hosts() = %q, want %q", got, want)
	}
}
