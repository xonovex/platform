package tekuri_test

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/jsonschema"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/jsonschema/tekuri"
)

const (
	itemID  = "https://schemas.test/item.json"
	shelfID = "https://schemas.test/shelf.json"
)

func documents() map[string][]byte {
	return map[string][]byte{
		itemID: []byte(`{
			"$schema": "https://json-schema.org/draft/2020-12/schema",
			"$id": "https://schemas.test/item.json",
			"type": "object",
			"required": ["name", "day"],
			"properties": {
				"name": {"type": "string", "minLength": 1},
				"day": {"type": "string", "format": "date"}
			},
			"additionalProperties": false
		}`),
		shelfID: []byte(`{
			"$schema": "https://json-schema.org/draft/2020-12/schema",
			"$id": "https://schemas.test/shelf.json",
			"type": "object",
			"required": ["items"],
			"properties": {"items": {"type": "array", "items": {"$ref": "item.json"}}}
		}`),
	}
}

func newValidator(t *testing.T) *tekuri.Validator {
	t.Helper()
	v, err := tekuri.New(documents())
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func TestValidateValid(t *testing.T) {
	got, err := newValidator(t).Validate(itemID, []byte(`{"name":"a","day":"2026-02-28"}`))
	if err != nil || len(got) != 0 {
		t.Fatalf("Validate = %+v, %v", got, err)
	}
}

func TestValidateViolations(t *testing.T) {
	cases := []struct {
		name         string
		document     string
		wantPointers []string
	}{
		{"missing field", `{"name":"a"}`, []string{""}},
		{"wrong type", `{"name":3,"day":"2026-02-28"}`, []string{"/name"}},
		{"extra property", `{"name":"a","day":"2026-02-28","extra":1}`, []string{""}},
		{"format asserted", `{"name":"a","day":"2026-02-30"}`, []string{"/day"}},
		{"two problems sorted", `{"name":"","day":"nope"}`, []string{"/day", "/name"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := newValidator(t).Validate(itemID, []byte(tc.document))
			if err != nil {
				t.Fatal(err)
			}
			var pointers []string
			for _, violation := range got {
				if violation.Message == "" {
					t.Errorf("violation without a message: %+v", violation)
				}
				pointers = append(pointers, violation.Pointer)
			}
			if !reflect.DeepEqual(pointers, tc.wantPointers) {
				t.Fatalf("pointers = %v; want %v (%+v)", pointers, tc.wantPointers, got)
			}
		})
	}
}

func TestValidateFollowsReferenceBetweenDocuments(t *testing.T) {
	v := newValidator(t)
	good, err := v.Validate(shelfID, []byte(`{"items":[{"name":"a","day":"2026-01-01"}]}`))
	if err != nil || len(good) != 0 {
		t.Fatalf("good shelf = %+v, %v", good, err)
	}
	bad, err := v.Validate(shelfID, []byte(`{"items":[{"name":"a"}]}`))
	if err != nil || len(bad) == 0 {
		t.Fatalf("bad shelf = %+v, %v", bad, err)
	}
	if bad[0].Pointer != "/items/0" {
		t.Fatalf("pointer = %q", bad[0].Pointer)
	}
}

func TestValidateMalformedJSONIsAViolation(t *testing.T) {
	for _, document := range []string{`{`, ``, `{} {}`} {
		got, err := newValidator(t).Validate(itemID, []byte(document))
		if err != nil || len(got) != 1 || !strings.HasPrefix(got[0].Message, "invalid JSON") {
			t.Fatalf("Validate(%q) = %+v, %v", document, got, err)
		}
	}
}

func TestValidateUnknownSchema(t *testing.T) {
	_, err := newValidator(t).Validate("https://schemas.test/none.json", []byte(`{}`))
	if !errors.Is(err, jsonschema.ErrUnknownSchema) {
		t.Fatalf("err = %v", err)
	}
}

func TestIDsAreSorted(t *testing.T) {
	want := []string{itemID, shelfID}
	if got := newValidator(t).IDs(); !reflect.DeepEqual(got, want) {
		t.Fatalf("IDs = %v; want %v", got, want)
	}
}

func TestNewRejectsBrokenInput(t *testing.T) {
	cases := map[string]map[string][]byte{
		"not JSON": {"https://schemas.test/a.json": []byte(`{`)},
		"bad keyword value": {"https://schemas.test/a.json": []byte(
			`{"$schema":"https://json-schema.org/draft/2020-12/schema","type":12}`)},
		"reference outside the given documents": {"https://schemas.test/a.json": []byte(
			`{"$schema":"https://json-schema.org/draft/2020-12/schema","$ref":"https://elsewhere.test/x.json"}`)},
	}
	for name, docs := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := tekuri.New(docs); err == nil {
				t.Fatal("New must fail")
			}
		})
	}
}

func TestOfflineReferenceErrorIsExplicit(t *testing.T) {
	_, err := tekuri.New(map[string][]byte{"https://schemas.test/a.json": []byte(
		`{"$schema":"https://json-schema.org/draft/2020-12/schema","$ref":"https://elsewhere.test/x.json"}`)})
	if err == nil || !strings.Contains(err.Error(), "not among the given schemas") {
		t.Fatalf("err = %v", err)
	}
}

func TestValidateRepeatedRunsGiveTheSameViolations(t *testing.T) {
	v := newValidator(t)
	first, _ := v.Validate(itemID, []byte(`{"name":3,"day":"nope","extra":1}`))
	for i := 0; i < 20; i++ {
		again, _ := v.Validate(itemID, []byte(`{"name":3,"day":"nope","extra":1}`))
		if !reflect.DeepEqual(first, again) {
			t.Fatalf("run %d differs: %+v vs %+v", i, again, first)
		}
	}
}
