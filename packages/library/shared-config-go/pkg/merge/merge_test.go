package merge_test

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/layer"
	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/merge"
	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/spec"
)

var section = spec.Section{Name: "audit", Keys: []spec.Key{
	{Section: "audit", Pointer: "/name", Rule: spec.Replace, Layers: spec.EveryFile},
	{Section: "audit", Pointer: "/home", Rule: spec.Replace, Layers: spec.TrustedFiles},
	{Section: "audit", Pointer: "/items", Rule: spec.Append, Layers: spec.EveryFile},
	{Section: "audit", Pointer: "/strict", Rule: spec.Require, Layers: spec.EveryFile},
	{Section: "audit", Pointer: "/limits/runs", Rule: spec.Cap, Layers: spec.EveryFile, Ceiling: 24},
	{Section: "audit", Pointer: "/entries", Rule: spec.Merge, Layers: spec.EveryFile},
}}

func doc(t *testing.T, l spec.Layer, body string) layer.Document {
	t.Helper()
	d, err := layer.Parse(l, "/"+string(l)+"/agentic.json", []byte(`{"version":1,"audit":`+body+`}`))
	if err != nil {
		t.Fatal(err)
	}
	return d
}

func flag(pointer, value string) merge.Flag {
	return merge.Flag{Section: "audit", Pointer: pointer, Value: json.RawMessage(value)}
}

func resolve(t *testing.T, docs []layer.Document, flags ...merge.Flag) merge.Resolved {
	t.Helper()
	r, err := merge.Merge([]spec.Section{section}, docs, flags)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func value(t *testing.T, r merge.Resolved, pointer string) merge.Value {
	t.Helper()
	v, ok := r.Source("audit", pointer)
	if !ok {
		t.Fatalf("%s is not set", pointer)
	}
	return v
}

func TestReplacePrecedence(t *testing.T) {
	user := doc(t, spec.User, `{"name":"u"}`)
	project := doc(t, spec.Project, `{"name":"p"}`)
	managed := doc(t, spec.Managed, `{"name":"m"}`)
	cases := []struct {
		docs  []layer.Document
		flags []merge.Flag
		want  string
		from  spec.Layer
	}{
		{[]layer.Document{user}, nil, `"u"`, spec.User},
		{[]layer.Document{user, project}, nil, `"p"`, spec.Project},
		{[]layer.Document{user, project}, []merge.Flag{flag("/name", `"f"`)}, `"f"`, spec.Flags},
		{[]layer.Document{managed, user, project}, nil, `"m"`, spec.Managed},
		{[]layer.Document{managed}, []merge.Flag{flag("/name", ` "m" `)}, `"m"`, spec.Managed},
	}
	for i, c := range cases {
		v := value(t, resolve(t, c.docs, c.flags...), "/name")
		if string(v.Raw) != c.want || v.Origins[0].Layer != c.from {
			t.Errorf("case %d: %s from %v", i, v.Raw, v.Origins)
		}
	}
}

func TestAFlagThatContradictsTheManagedFileFails(t *testing.T) {
	_, err := merge.Merge([]spec.Section{section}, []layer.Document{doc(t, spec.Managed, `{"name":"m"}`)}, []merge.Flag{flag("/name", `"f"`)})
	if err == nil || !strings.Contains(err.Error(), "the command line sets audit.name") {
		t.Fatalf("err = %v", err)
	}
}

func TestAProjectCannotSetATrustedKey(t *testing.T) {
	_, err := merge.Merge([]spec.Section{section}, []layer.Document{doc(t, spec.Project, `{"home":"/h"}`)}, nil)
	if err == nil || !strings.Contains(err.Error(), "the project file sets audit.home, which only the managed and user files may set") {
		t.Fatalf("err = %v", err)
	}
	if v := value(t, resolve(t, []layer.Document{doc(t, spec.User, `{"home":"/h"}`)}), "/home"); string(v.Raw) != `"/h"` {
		t.Fatalf("user home = %s", v.Raw)
	}
}

func TestAppendPutsTheManagedItemsLast(t *testing.T) {
	r := resolve(t, []layer.Document{
		doc(t, spec.Managed, `{"items":["m"]}`), doc(t, spec.User, `{"items":["u"]}`), doc(t, spec.Project, `{"items":["p"]}`),
	}, flag("/items", `["f"]`))
	v := value(t, r, "/items")
	if string(v.Raw) != `["u","p","f","m"]` || len(v.Origins) != 4 {
		t.Fatalf("items = %s from %v", v.Raw, v.Origins)
	}
}

func TestRequireIsTrueWhenAnyLayerDemandsIt(t *testing.T) {
	v := value(t, resolve(t, []layer.Document{doc(t, spec.User, `{"strict":false}`), doc(t, spec.Project, `{"strict":true}`)}), "/strict")
	if string(v.Raw) != "true" || len(v.Origins) != 1 || v.Origins[0].Layer != spec.Project {
		t.Fatalf("strict = %s from %v", v.Raw, v.Origins)
	}
	v = value(t, resolve(t, []layer.Document{doc(t, spec.User, `{"strict":false}`)}), "/strict")
	if string(v.Raw) != "false" || len(v.Origins) != 0 {
		t.Fatalf("strict = %s from %v", v.Raw, v.Origins)
	}
}

func TestCap(t *testing.T) {
	runs := func(l spec.Layer, n string) layer.Document { return doc(t, l, `{"limits":{"runs":`+n+`}}`) }
	cases := map[string]struct {
		docs  []layer.Document
		flags []merge.Flag
		want  string
		from  spec.Layer
	}{
		"default":                     {nil, nil, "24", spec.Default},
		"project lowers":              {[]layer.Document{runs(spec.Project, "5")}, nil, "5", spec.Project},
		"project cannot raise":        {[]layer.Document{runs(spec.Project, "50")}, nil, "24", spec.Default},
		"user raises":                 {[]layer.Document{runs(spec.User, "50")}, nil, "50", spec.User},
		"project lowers the user":     {[]layer.Document{runs(spec.User, "50"), runs(spec.Project, "30")}, nil, "30", spec.Project},
		"project stays under user":    {[]layer.Document{runs(spec.User, "10"), runs(spec.Project, "30")}, nil, "10", spec.User},
		"managed raises the default":  {[]layer.Document{runs(spec.Managed, "40")}, nil, "40", spec.Managed},
		"managed bounds the user":     {[]layer.Document{runs(spec.Managed, "40"), runs(spec.User, "50")}, nil, "40", spec.Managed},
		"user under managed":          {[]layer.Document{runs(spec.Managed, "40"), runs(spec.User, "30")}, nil, "30", spec.User},
		"flag raises":                 {nil, []merge.Flag{flag("/limits/runs", "100")}, "100", spec.Flags},
		"flag beats a project":        {[]layer.Document{runs(spec.Project, "5")}, []merge.Flag{flag("/limits/runs", "8")}, "8", spec.Flags},
		"flag within the managed cap": {[]layer.Document{runs(spec.Managed, "40")}, []merge.Flag{flag("/limits/runs", "40")}, "40", spec.Flags},
	}
	for name, c := range cases {
		v := value(t, resolve(t, c.docs, c.flags...), "/limits/runs")
		if string(v.Raw) != c.want || v.Origins[0].Layer != c.from {
			t.Errorf("%s: %s from %v", name, v.Raw, v.Origins)
		}
	}
	_, err := merge.Merge([]spec.Section{section}, []layer.Document{runs(spec.Managed, "40")}, []merge.Flag{flag("/limits/runs", "41")})
	if err == nil || !strings.Contains(err.Error(), "above the ceiling 40") {
		t.Fatalf("flag over the managed cap: %v", err)
	}
}

func TestMergeTakesEachEntryWhole(t *testing.T) {
	r := resolve(t, []layer.Document{
		doc(t, spec.Managed, `{"entries":{"a":{"x":1}}}`),
		doc(t, spec.User, `{"entries":{"a":{"y":2},"b":{"y":2},"c":{}}}`),
		doc(t, spec.Project, `{"entries":{"b":{"z":3}}}`),
	})
	v := value(t, r, "/entries")
	if string(v.Raw) != `{"a":{"x":1},"b":{"z":3},"c":{}}` {
		t.Fatalf("entries = %s", v.Raw)
	}
	if v.Entries["a"].Layer != spec.Managed || v.Entries["b"].Layer != spec.Project || v.Entries["c"].Layer != spec.User {
		t.Fatalf("entry origins = %v", v.Entries)
	}
	layers := []spec.Layer{}
	for _, o := range v.Origins {
		layers = append(layers, o.Layer)
	}
	if !reflect.DeepEqual(layers, []spec.Layer{spec.Managed, spec.Project, spec.User}) {
		t.Fatalf("origins = %v", layers)
	}
}

func TestMergeRejectsAFlagEntryThatContradictsTheManagedFile(t *testing.T) {
	_, err := merge.Merge([]spec.Section{section}, []layer.Document{doc(t, spec.Managed, `{"entries":{"a":1}}`)}, []merge.Flag{flag("/entries", `{"a":2}`)})
	if err == nil || !strings.Contains(err.Error(), "audit.entries.a") {
		t.Fatalf("err = %v", err)
	}
	if _, err := merge.Merge([]spec.Section{section}, []layer.Document{doc(t, spec.Managed, `{"entries":{"a":1}}`)}, []merge.Flag{flag("/entries", `{"b":2}`)}); err != nil {
		t.Fatalf("another entry: %v", err)
	}
}

func TestSectionRendersTheMergedValues(t *testing.T) {
	r := resolve(t, []layer.Document{doc(t, spec.User, `{"name":"u","limits":{"runs":3},"other":1}`)})
	if got := string(r.Section("audit")); got != `{"limits":{"runs":3},"name":"u"}` {
		t.Fatalf("section = %s", got)
	}
	if got := string(r.Section("none")); got != `{}` {
		t.Fatalf("empty section = %s", got)
	}
	if _, ok := r.Source("audit", "/items"); ok {
		t.Fatal("an unset key has no value")
	}
	if _, ok := resolve(t, []layer.Document{doc(t, spec.User, `{"limits":{}}`)}).Source("audit", "/limits/runs"); !ok {
		t.Fatal("a cap key always has a value")
	}
	if len(r.Values()) != 2 {
		t.Fatalf("values = %v", r.Values())
	}
	whole := spec.Section{Name: "eval", Keys: []spec.Key{{Section: "eval", Rule: spec.Merge, Layers: spec.EveryFile}}}
	d, _ := layer.Parse(spec.User, "/u", []byte(`{"version":1,"eval":{"a":{}}}`))
	r, err := merge.Merge([]spec.Section{whole}, []layer.Document{d}, nil)
	if err != nil || string(r.Section("eval")) != `{"a":{}}` {
		t.Fatalf("whole section = %s, %v", r.Section("eval"), err)
	}
}

func TestMergeRejectsMalformedInput(t *testing.T) {
	cases := map[string]struct {
		body  string
		flags []merge.Flag
		want  string
	}{
		"unknown flag":       {`{}`, []merge.Flag{flag("/nope", "1")}, "not a key of the requested sections"},
		"parent not object":  {`{"limits":3}`, nil, "not an object"},
		"append not array":   {`{"items":"x"}`, nil, "not an array"},
		"require not bool":   {`{"strict":"yes"}`, nil, "not a boolean"},
		"cap not number":     {`{"limits":{"runs":"many"}}`, nil, "not a number"},
		"merge not object":   {`{"entries":[]}`, nil, "not an object"},
		"merge null":         {`{"entries":null}`, nil, "not an object"},
		"flag in no section": {`{}`, []merge.Flag{{Section: "t", Pointer: "/name"}}, "t/name"},
		"flag not json":      {`{}`, []merge.Flag{flag("/name", "x")}, "not JSON"},
	}
	for name, c := range cases {
		if _, err := merge.Merge([]spec.Section{section}, []layer.Document{doc(t, spec.User, c.body)}, c.flags); err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: err = %v", name, err)
		}
	}
	bogus := spec.Section{Name: "audit", Keys: []spec.Key{{Section: "audit", Pointer: "/name", Rule: "sum", Layers: spec.EveryFile}}}
	if _, err := merge.Merge([]spec.Section{bogus}, []layer.Document{doc(t, spec.User, `{"name":1}`)}, nil); err == nil || !strings.Contains(err.Error(), `unknown rule "sum"`) {
		t.Fatalf("unknown rule: %v", err)
	}
}

func TestAFileWithoutTheSectionGivesNothing(t *testing.T) {
	other, err := layer.Parse(spec.Project, "/p", []byte(`{"version":1,"eval":{}}`))
	if err != nil {
		t.Fatal(err)
	}
	v := value(t, resolve(t, []layer.Document{other, doc(t, spec.User, `{"items":["u"]}`)}), "/items")
	if string(v.Raw) != `["u"]` || len(v.Origins) != 1 {
		t.Fatalf("items = %s from %v", v.Raw, v.Origins)
	}
}
