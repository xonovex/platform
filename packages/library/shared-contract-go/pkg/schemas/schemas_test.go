package schemas_test

import (
	"encoding/json"
	"errors"
	iofs "io/fs"
	"reflect"
	"sort"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/schemas"
)

func TestDocumentsMatchIDs(t *testing.T) {
	documents, err := schemas.Documents()
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for id := range documents {
		got = append(got, id)
	}
	sort.Strings(got)
	if want := schemas.IDs(); !reflect.DeepEqual(got, want) {
		t.Fatalf("embedded ids = %v; constants = %v", got, want)
	}
}

func TestEverySchemaIsDraft2020UnderTheNamespace(t *testing.T) {
	documents, err := schemas.Documents()
	if err != nil {
		t.Fatal(err)
	}
	for id, data := range documents {
		if !strings.HasPrefix(id, schemas.Namespace) {
			t.Errorf("%s is outside the namespace", id)
		}
		var header struct {
			Schema string `json:"$schema"`
			Title  string `json:"title"`
		}
		if err := json.Unmarshal(data, &header); err != nil {
			t.Fatalf("%s: %v", id, err)
		}
		if header.Schema != "https://json-schema.org/draft/2020-12/schema" {
			t.Errorf("%s declares %q", id, header.Schema)
		}
		if header.Title == "" {
			t.Errorf("%s has no title", id)
		}
	}
}

func TestReadRejectsBadDocuments(t *testing.T) {
	cases := map[string]fstest.MapFS{
		"not JSON":    {"a.schema.json": {Data: []byte(`{`)}},
		"no id":       {"a.schema.json": {Data: []byte(`{}`)}},
		"repeated id": {"a.schema.json": {Data: []byte(`{"$id":"x"}`)}, "b.schema.json": {Data: []byte(`{"$id":"x"}`)}},
	}
	for name, fsys := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := schemas.Read(fsys); err == nil {
				t.Fatal("Read must fail")
			}
		})
	}
	if documents, err := schemas.Read(fstest.MapFS{"a.schema.json": {Data: []byte(`{"$id":"x"}`)}}); err != nil || len(documents) != 1 {
		t.Fatalf("Read = %v, %v", documents, err)
	}
}

type brokenFS struct{}

func (brokenFS) Open(string) (iofs.File, error) { return nil, errors.New("closed") }

func TestReadReportsAnUnreadableTree(t *testing.T) {
	if _, err := schemas.Read(brokenFS{}); err == nil {
		t.Fatal("Read must fail")
	}
}

// forbidden names keywords whose errors the validator adapter locates wrongly.
var forbidden = []string{"propertyNames", "contentSchema"}

func TestNoSchemaUsesAKeywordTheAdapterMislocates(t *testing.T) {
	documents, err := schemas.Documents()
	if err != nil {
		t.Fatal(err)
	}
	var walk func(id string, node any)
	walk = func(id string, node any) {
		switch value := node.(type) {
		case map[string]any:
			for key, child := range value {
				for _, word := range forbidden {
					if key == word {
						t.Errorf("%s uses %s, which the validator reports at the wrong pointer", id, word)
					}
				}
				walk(id, child)
			}
		case []any:
			for _, child := range value {
				walk(id, child)
			}
		}
	}
	for id, data := range documents {
		var root any
		if err := json.Unmarshal(data, &root); err != nil {
			t.Fatal(err)
		}
		walk(id, root)
	}
}
