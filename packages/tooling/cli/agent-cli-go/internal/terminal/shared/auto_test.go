package shared

import (
	"errors"
	"strings"
	"testing"
)

func TestSelectAutoPicksTheOnlySessionTheCLIRunsIn(t *testing.T) {
	got, err := SelectAuto([]TerminalType{"core"})
	if err != nil || got != "core" {
		t.Errorf("SelectAuto() = (%q, %v), want core", got, err)
	}
}

func TestSelectAutoFailsClosedOutsideEverySession(t *testing.T) {
	got, err := SelectAuto(nil)
	if !errors.Is(err, ErrNotInside) || got != TerminalNone {
		t.Errorf("SelectAuto() = (%q, %v), want the refusal", got, err)
	}
}

func TestSelectAutoAsksForANameInsideTwoSessions(t *testing.T) {
	got, err := SelectAuto([]TerminalType{"core", "tui"})
	if err == nil || got != TerminalNone {
		t.Fatalf("SelectAuto() = (%q, %v), want a refusal", got, err)
	}
	if !strings.Contains(err.Error(), "core and tui") || !strings.Contains(err.Error(), "--terminal core or --terminal tui") {
		t.Errorf("error = %v, want both names and the flags to choose", err)
	}
}
