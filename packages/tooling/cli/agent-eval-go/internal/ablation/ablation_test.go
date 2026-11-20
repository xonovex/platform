package ablation_test

import (
	"errors"
	"testing"

	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/ablation"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/model"
)

const doc = "# Project\n\nIntro.\n\n## Testing\n\nRun tests first.\nNever skip them.\n\n```md\n## Fake\n```\n\n### Detail\n\nDeep.\n\n## Style\n\nUse tabs.\n"

func TestPresentKeepsTheFile(t *testing.T) {
	got, err := ablation.Apply(doc, model.Anchor{}, "", model.Present)
	if err != nil || got != doc {
		t.Fatalf("%v", err)
	}
}

func TestAbsentRemovesTheSectionToTheNextHeadingOfTheSameLevel(t *testing.T) {
	got, err := ablation.Apply(doc, model.Anchor{Heading: "## Testing"}, "", model.Absent)
	want := "# Project\n\nIntro.\n\n## Style\n\nUse tabs.\n"
	if err != nil || got != want {
		t.Fatalf("%v\n%q", err, got)
	}
	// A bare heading text matches too, and the last section runs to the end.
	got, err = ablation.Apply(doc, model.Anchor{Heading: "Style"}, "", model.Absent)
	if err != nil || got[len(got)-len("Deep.\n\n"):] != "Deep.\n\n" {
		t.Fatalf("%v %q", err, got)
	}
}

func TestAbsentRemovesAnExactQuotation(t *testing.T) {
	got, err := ablation.Apply(doc, model.Anchor{Text: "Never skip them.\n"}, "", model.Absent)
	if err != nil || got == doc || len(got) != len(doc)-len("Never skip them.\n") {
		t.Fatalf("%v %q", err, got)
	}
}

func TestAblatedRemovesOnlyTheNamedPhrase(t *testing.T) {
	got, err := ablation.Apply(doc, model.Anchor{Heading: "Testing"}, "Never skip them.", model.Ablated)
	if err != nil || len(got) != len(doc)-len("Never skip them.") {
		t.Fatalf("%v %q", err, got)
	}
	got, err = ablation.Apply(doc, model.Anchor{Text: "Run tests first."}, "first", model.Ablated)
	if err != nil || len(got) != len(doc)-len("first") {
		t.Fatalf("%v %q", err, got)
	}
}

func TestFailures(t *testing.T) {
	cases := map[string]struct {
		anchor model.Anchor
		ablate string
		c      model.Condition
	}{
		"heading missing":    {model.Anchor{Heading: "Nope"}, "", model.Absent},
		"fenced is no head":  {model.Anchor{Heading: "Fake"}, "", model.Absent},
		"text missing":       {model.Anchor{Text: "zzz"}, "", model.Absent},
		"text twice":         {model.Anchor{Text: "."}, "", model.Absent},
		"no anchor":          {model.Anchor{}, "", model.Absent},
		"ablate empty":       {model.Anchor{Heading: "Testing"}, "", model.Ablated},
		"ablate outside":     {model.Anchor{Heading: "Testing"}, "Use tabs.", model.Ablated},
		"ablate twice":       {model.Anchor{Heading: "Testing"}, "t", model.Ablated},
		"unknown condition":  {model.Anchor{Heading: "Testing"}, "", model.Condition("odd")},
		"not a heading line": {model.Anchor{Heading: "Intro."}, "", model.Absent},
	}
	for name, c := range cases {
		if _, err := ablation.Apply(doc, c.anchor, c.ablate, c.c); err == nil {
			t.Errorf("%s: want an error", name)
		}
	}
	if _, err := ablation.Apply(doc, model.Anchor{Heading: "Nope"}, "", model.Absent); !errors.Is(err, ablation.ErrAnchor) {
		t.Fatal("an anchor failure wraps ErrAnchor")
	}
}
