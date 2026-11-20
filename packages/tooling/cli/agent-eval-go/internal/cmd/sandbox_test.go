package cmd_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/agents"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/harness"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/model"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/predicate"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/registry"
)

const userFile = "/home/u/.config/agentic/agentic.json"

// sandboxOf runs a trigger eval and returns the sandbox the engine was built
// with.
func sandboxOf(t *testing.T, files map[string]string, args ...string) (model.Sandbox, error) {
	t.Helper()
	f := newFixture(t, files, selects(good))
	var got model.Sandbox
	f.deps.Engine = func(_ agents.Catalog, sandbox model.Sandbox) (harness.Executor, *predicate.Registry, error) {
		got = sandbox
		return f.exec, registry.Predicates(f.exec), nil
	}
	_, err := f.run(append([]string{"trigger", "alpha", "--runs", "1"}, args...)...)
	return got, err
}

func TestTheHarnessRunsUnconfinedByDefault(t *testing.T) {
	got, err := sandboxOf(t, baseFiles())
	if err != nil || !got.Unconfined() {
		t.Fatalf("%v %+v", err, got)
	}
}

func TestTheUserChoosesTheSandbox(t *testing.T) {
	files := baseFiles()
	files[userFile] = `{"version":1,"eval":{"sandbox":{"isolation":"docker","image":"registry.example/eval:1","customEnv":["A=1"]}}}`
	got, err := sandboxOf(t, files)
	if err != nil || got.Isolation != "docker" || got.Image != "registry.example/eval:1" {
		t.Fatalf("the user file chooses the sandbox: %v %+v", err, got)
	}
}

func TestAProjectFileMayOnlyDemandAGuarantee(t *testing.T) {
	files := baseFiles()
	files[projectFile] = `{"version":1,"eval":{"defaultProfile":"ci","profiles":{"ci":{"harness":"claude","model":"m"}},"pluginDirs":["plugins"],"sandbox":{"requireEgressRestricted":true}}}`
	got, err := sandboxOf(t, files)
	if err != nil || !got.RequireEgressRestricted || got.Isolation != "" {
		t.Fatalf("a project demands a guarantee: %v %+v", err, got)
	}
	for _, member := range []string{`"isolation":"none"`, `"network":"host"`, `"allowHosts":["evil.example"]`, `"credentials":["K=env:K"]`, `"hostPassthrough":true`} {
		files[projectFile] = `{"version":1,"eval":{"profiles":{},"sandbox":{` + member + `}}}`
		if _, err := sandboxOf(t, files); err == nil || !strings.Contains(err.Error(), "eval.sandbox.") {
			t.Errorf("%s: a project file may not choose the confinement: %v", member, err)
		}
	}
}

func TestConfigShowNamesTheSourceOfTheSandbox(t *testing.T) {
	files := baseFiles()
	files[userFile] = `{"version":1,"eval":{"sandbox":{"isolation":"docker"}}}`
	f := newFixture(t, files, selects(good))
	if code, err := f.run("config", "show", "--sources"); err != nil || code != 0 ||
		!strings.Contains(f.out.String(), `eval.sandbox.isolation = "docker"  (user `+userFile+`)`) {
		t.Fatalf("%d %v\n%s", code, err, f.out)
	}
}

type bridge struct {
	socket, listen string
	command        []string
	code           int
	err            error
}

func (b *bridge) Run(socket, listen string, command []string) (int, error) {
	b.socket, b.listen, b.command = socket, listen, command
	return b.code, b.err
}

func TestTheEgressBridgeRunsTheHarnessAndPassesItsStatusOn(t *testing.T) {
	f := newFixture(t, baseFiles(), selects(good))
	b := &bridge{code: 3}
	f.deps.Bridge = b
	code, err := f.run("egress-bridge", "--socket", "/run/p.sock", "--listen", "127.0.0.1:3128", "--", "claude", "-p", "hi")
	if err != nil || code != 3 || b.socket != "/run/p.sock" || b.listen != "127.0.0.1:3128" || strings.Join(b.command, " ") != "claude -p hi" {
		t.Fatalf("%d %v %+v", code, err, b)
	}
	b.code, b.err = 0, errors.New("listen refused")
	if _, err := f.run("egress-bridge", "--socket", "/run/p.sock", "--listen", "127.0.0.1:3128", "--", "claude"); err == nil || !strings.Contains(err.Error(), "listen refused") {
		t.Fatalf("err = %v", err)
	}
	b.err = nil
	if code, err := f.run("egress-bridge", "--socket", "/run/p.sock", "--listen", "127.0.0.1:3128", "--", "claude"); err != nil || code != 0 {
		t.Fatalf("%d %v", code, err)
	}
}
