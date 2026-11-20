package sources_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/sources"
)

func TestEncodeDecodeRoundTrip(t *testing.T) {
	in := sources.Document{
		Schema:   "https://schemas.xonovex.com/agent/sources.schema.json",
		Version:  1,
		Title:    "Sources",
		Preamble: "p",
		Sources: []sources.Source{{
			ID: "a", Kind: sources.Git, Title: "A <b> & c", Checkout: "../up", URLs: []string{"https://e.test/?a=1&b=2"},
			Version: "1.2.3", Commit: "abcdef1", Watch: []sources.Watch{{Path: "src", Feeds: []string{"references/x.md"}}},
			ContentSHA256: "sha256:" + strings.Repeat("0", 64), ReviewedAt: "2026-01-01", CoversAllReferences: true,
			Usage: []string{"u"}, Claims: []sources.Claim{{ID: "c1", Statement: "s"}}, Remarks: "r",
		}},
		Notes: []sources.Note{{Title: "Refresh", Text: "t", After: "a"}},
	}
	data, err := sources.Encode(in)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(string(data), "}\n") || !strings.Contains(string(data), "A <b> & c") || !strings.Contains(string(data), "a=1&b=2") {
		t.Fatalf("encoded text is not plain indented JSON:\n%s", data)
	}
	out, err := sources.Decode(data)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(out, in) {
		t.Fatalf("round trip changed the document:\n%+v\n%+v", out, in)
	}
	again, err := sources.Encode(out)
	if err != nil || string(again) != string(data) {
		t.Fatalf("encoding is not stable: %v", err)
	}
}

func TestEncodeWritesAnEmptySourceList(t *testing.T) {
	data, err := sources.Encode(sources.Document{Version: 1})
	if err != nil || !strings.Contains(string(data), `"sources": []`) {
		t.Fatalf("data = %s, err = %v", data, err)
	}
}

func TestEncodeReportsAnUnencodableDocument(t *testing.T) {
	// A document always encodes, so the failure path is reached through the
	// seam that takes any value.
	if _, err := sources.EncodeJSON(make(chan int)); err == nil {
		t.Fatal("want an error")
	}
}

func TestDecodeRejectsAnUnknownField(t *testing.T) {
	if _, err := sources.Decode([]byte(`{"version":1,"sources":[],"extra":1}`)); err == nil {
		t.Fatal("want an error")
	}
}
