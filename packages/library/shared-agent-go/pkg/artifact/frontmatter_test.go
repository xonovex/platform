package artifact

import "testing"

func TestParseDocument(t *testing.T) {
	document, err := ParseDocument("---\nname: x\ndescription: \"Use when: a, b\"\nmodel: 'o''k'\ntools:\n  - a\n  - b\nempty:\n# note\n---\nbody\n")
	if err != nil {
		t.Fatal(err)
	}
	if !document.HasFrontmatter || document.Body != "body\n" {
		t.Errorf("document = %+v", document)
	}
	want := map[string]string{"name": "x", "description": "Use when: a, b", "model": "o'k", "tools": "", "empty": ""}
	for key, value := range want {
		if got, ok := document.Fields[key]; !ok || got != value {
			t.Errorf("field %q = %q, %v; want %q", key, got, ok, value)
		}
	}
	if len(document.Fields) != len(want) {
		t.Errorf("fields = %v", document.Fields)
	}
}

func TestParseDocumentWithoutFrontmatterIsAllBody(t *testing.T) {
	document, err := ParseDocument("# Title\n---\nnot front matter\n")
	if err != nil || document.HasFrontmatter || document.Body != "# Title\n---\nnot front matter\n" {
		t.Errorf("document = %+v, %v", document, err)
	}
}

func TestParseDocumentHandlesCRLFAndOddValues(t *testing.T) {
	document, err := ParseDocument("---\r\nname: x\r\nbad: \"unterminated\\\"\r\nsingle: 'a'\r\nplain: a:b\r\n---\r\nbody")
	if err != nil {
		t.Fatal(err)
	}
	if document.Fields["name"] != "x" || document.Fields["single"] != "a" || document.Fields["plain"] != "a:b" {
		t.Errorf("fields = %v", document.Fields)
	}
	if got := document.Fields["bad"]; got != `unterminated\` {
		t.Errorf("bad = %q", got)
	}
	if document.Body != "body" {
		t.Errorf("body = %q", document.Body)
	}
}

func TestParseDocumentRejectsAnUnclosedFence(t *testing.T) {
	if _, err := ParseDocument("---\nname: x\n"); err == nil {
		t.Error("an unclosed front matter parsed")
	}
}

func TestParseDocumentIgnoresLinesThatAreNotKeys(t *testing.T) {
	document, err := ParseDocument("---\nno colon\n: nokey\nwith space: v\nok: v\n---\n")
	if err != nil {
		t.Fatal(err)
	}
	if len(document.Fields) != 1 || document.Fields["ok"] != "v" {
		t.Errorf("fields = %v", document.Fields)
	}
}

func TestDropFieldsRemovesKeysAndTheirContinuation(t *testing.T) {
	text := "---\nname: x\nmodel: sonnet\ntools:\n  - a\n  - b\ndescription: d\n---\nbody\n"
	got := DropFields(text, "tools", "model")
	if want := "---\nname: x\ndescription: d\n---\nbody\n"; got != want {
		t.Errorf("DropFields = %q, want %q", got, want)
	}
}

func TestDropFieldsLeavesUnchangedText(t *testing.T) {
	for name, text := range map[string]string{
		"no front matter": "body\n",
		"unclosed":        "---\nname: x\n",
	} {
		if got := DropFields(text, "name"); got != text {
			t.Errorf("%s: DropFields changed the text to %q", name, got)
		}
	}
	if got := DropFields("---\nname: x\n---\nb"); got != "---\nname: x\n---\nb" {
		t.Errorf("no keys changed the text to %q", got)
	}
}

func TestDropFieldsCanEmptyTheFrontMatter(t *testing.T) {
	if got, want := DropFields("---\nmodel: a\n---\nb", "model"), "---\n---\nb"; got != want {
		t.Errorf("DropFields = %q, want %q", got, want)
	}
}
