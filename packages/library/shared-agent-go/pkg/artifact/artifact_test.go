package artifact

import (
	"strings"
	"testing"
)

func TestValidName(t *testing.T) {
	for name, want := range map[string]bool{
		"a": true, "using-git": true, "writing-c99": true, "a1-b2": true,
		"": false, "Using": false, "a--b": false, "-a": false, "a-": false, "a_b": false,
		strings.Repeat("a", 64): true, strings.Repeat("a", 65): false,
	} {
		if got := ValidName(name); got != want {
			t.Errorf("ValidName(%q) = %v, want %v", name, got, want)
		}
	}
}

func TestConfinedPath(t *testing.T) {
	for _, ok := range []string{"a", "a/b.md", ".hidden/x", "a b/c"} {
		if err := ConfinedPath(ok); err != nil {
			t.Errorf("ConfinedPath(%q) = %v", ok, err)
		}
	}
	for _, bad := range []string{"", "/a", "a//b", "a/", "./a", "a/./b", "../a", "a/../b", "a\\b", "a\x00b", "a\nb", "a\x7fb"} {
		if err := ConfinedPath(bad); err == nil {
			t.Errorf("ConfinedPath(%q) passed", bad)
		}
	}
}

func TestSetInstructionTextJoinsInSourceOrder(t *testing.T) {
	var set Set
	if !set.Empty() || set.InstructionText() != "" {
		t.Fatal("an empty set is empty and has no instruction text")
	}
	set.Instructions = []Instruction{{Source: "a", Content: "first\n\n"}, {Source: "b", Content: "second"}}
	if set.Empty() {
		t.Error("a set with an instruction is not empty")
	}
	if got, want := set.InstructionText(), "first\n\nsecond\n"; got != want {
		t.Errorf("InstructionText() = %q, want %q", got, want)
	}
}
