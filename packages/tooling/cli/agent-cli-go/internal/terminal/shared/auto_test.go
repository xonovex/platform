package shared

import (
	"errors"
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

func TestSelectAutoPicksTheFirstCandidateInsideTwoSessions(t *testing.T) {
	got, err := SelectAuto([]TerminalType{"core", "tui"})
	if err != nil || got != "core" {
		t.Errorf("SelectAuto() = (%q, %v), want core, the first candidate", got, err)
	}
}
