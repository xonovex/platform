package condition_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/condition"
)

func TestParse(t *testing.T) {
	for _, c := range condition.All() {
		got, err := condition.Parse(string(c))
		if err != nil || got != c {
			t.Fatalf("Parse(%q) = %q, %v", c, got, err)
		}
		if !c.Valid() {
			t.Fatalf("%q must be valid", c)
		}
	}
	if _, err := condition.Parse("removed"); !errors.Is(err, condition.ErrUnknown) {
		t.Fatalf("err = %v", err)
	}
	if condition.Condition("removed").Valid() {
		t.Fatal("unknown condition must not be valid")
	}
}

func TestKeepsInstruction(t *testing.T) {
	want := map[condition.Condition]bool{condition.Present: true, condition.Ablated: true, condition.Absent: false}
	for c, keeps := range want {
		if got := c.KeepsInstruction(); got != keeps {
			t.Errorf("%q.KeepsInstruction() = %v", c, got)
		}
	}
}

func TestUnmarshalJSON(t *testing.T) {
	var holder struct {
		Condition condition.Condition `json:"condition"`
	}
	if err := json.Unmarshal([]byte(`{"condition":"ablated"}`), &holder); err != nil || holder.Condition != condition.Ablated {
		t.Fatalf("got %q, %v", holder.Condition, err)
	}
	if err := json.Unmarshal([]byte(`{"condition":"nope"}`), &holder); !errors.Is(err, condition.ErrUnknown) {
		t.Fatalf("unknown name err = %v", err)
	}
	if err := json.Unmarshal([]byte(`{"condition":3}`), &holder); err == nil {
		t.Fatal("a number must fail")
	}
}
