package cmd

import (
	"errors"
	"slices"
	"testing"

	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/launch"
)

func TestEgressBridgeCommandIsHiddenAndPassesItsArguments(t *testing.T) {
	r := newTestRoot()
	if err := r.execute("egress-bridge", "--socket", "/run/p.sock", "--listen", "127.0.0.1:3128", "--", "agent", "--flag"); err != nil {
		t.Fatal(err)
	}
	if r.bridge.socket != "/run/p.sock" || r.bridge.listen != "127.0.0.1:3128" || !slices.Equal(r.bridge.command, []string{"agent", "--flag"}) {
		t.Errorf("bridge = %+v", r.bridge)
	}
	root := NewRootCommand(r.deps, r.ports, RecordPorts{}, r.bridge)
	found, _, err := root.Find([]string{"egress-bridge"})
	if err != nil || !found.Hidden {
		t.Errorf("egress-bridge must be a hidden command: %v, %v", found, err)
	}
}

func TestEgressBridgePassesOnTheExitCodeOfTheCommand(t *testing.T) {
	r := newTestRoot()
	r.bridge.code = 6
	var exit launch.ExitError
	if err := r.execute("egress-bridge", "--socket", "/s", "--listen", "127.0.0.1:1", "--", "agent"); !errors.As(err, &exit) || exit.Code != 6 {
		t.Errorf("error = %v, want exit code 6", err)
	}
}

func TestEgressBridgeReportsAFailureAndRequiresItsInputs(t *testing.T) {
	r := newTestRoot()
	r.bridge.err = errors.New("no proxy")
	if err := r.execute("egress-bridge", "--socket", "/s", "--listen", "127.0.0.1:1", "--", "agent"); err == nil {
		t.Error("a bridge failure must fail the command")
	}
	for _, args := range [][]string{
		{"egress-bridge", "--listen", "127.0.0.1:1", "--", "agent"},
		{"egress-bridge", "--socket", "/s", "--", "agent"},
		{"egress-bridge", "--socket", "/s", "--listen", "127.0.0.1:1"},
	} {
		if err := newTestRoot().execute(args...); err == nil {
			t.Errorf("%v succeeded", args)
		}
	}
}
