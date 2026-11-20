package registry_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/agents"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/provision"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/types"
	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/logging"
	"github.com/xonovex/platform/packages/library/shared-launch-go/pkg/oshost/oshosttest"

	confinelaunch "github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/confine/launch"
	confinenone "github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/confine/none"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/harness"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/model"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/predicate"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/registry"
)

type exec struct{}

func (exec) Execute(context.Context, harness.Request) (harness.Observation, error) {
	return harness.Observation{Structured: `{"assertion_results":[{"text":"c","passed":true,"evidence":"e"}]}`}, nil
}

func TestEveryLeafIsRegistered(t *testing.T) {
	harnesses, err := registry.Harnesses(agents.Embedded().Get)
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(harnesses.Names(), ","); got != "claude,codex" {
		t.Fatal(got)
	}
	names := registry.Protocols().Names()
	if len(names) != len(model.Protocols()) {
		t.Fatalf("every protocol has a leaf: %v", names)
	}
	for _, p := range model.Protocols() {
		if _, err := registry.Protocols().Get(p); err != nil {
			t.Error(err)
		}
	}
}

func TestHarnessesFailOnFactsALeafCannotUse(t *testing.T) {
	missing := func(types.AgentType) (*types.AgentConfig, error) { return nil, errors.New("no facts") }
	if _, err := registry.Harnesses(missing); err == nil || err.Error() != "no facts" {
		t.Errorf("Harnesses() error = %v, want the lookup fault", err)
	}
	bare := func(agentType types.AgentType) (*types.AgentConfig, error) {
		return &types.AgentConfig{Type: agentType}, nil
	}
	if _, err := registry.Harnesses(bare); err == nil {
		t.Error("Harnesses() over facts without a layout must fail")
	}
}

func TestPredicateRegistryEvaluatesEveryKind(t *testing.T) {
	preds := []model.Predicate{
		{Kind: model.KindFilesystem, Path: "a", Check: "absent"},
		{Kind: model.KindSession, Event: "command", Match: "x", Max: new(int)},
		{Kind: model.KindLLM, Criterion: "c"},
	}
	got, _, err := registry.Predicates(exec{}).Evaluate(context.Background(), preds, predicate.Subject{})
	if err != nil || !got[0].Passed || !got[1].Passed || !got[2].Passed {
		t.Fatalf("%v %+v", err, got)
	}
}

func TestTheSandboxSelectsTheConfinement(t *testing.T) {
	host := oshosttest.New()
	log := logging.New(&strings.Builder{}, &strings.Builder{}, false, false)
	unconfined, err := registry.Confiner(model.Sandbox{Isolation: "none", Network: "host"}, agents.Embedded(), host, log)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := unconfined.(confinenone.Confiner); !ok {
		t.Errorf("%T, want the confinement that confines nothing", unconfined)
	}
	confined, err := registry.Confiner(model.Sandbox{Isolation: "docker"}, agents.Embedded(), host, log)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := confined.(*confinelaunch.Confiner); !ok {
		t.Errorf("%T, want the shared isolator", confined)
	}
	host.HomeErr = errors.New("no home")
	if _, err := registry.Confiner(model.Sandbox{Isolation: "docker"}, agents.Embedded(), host, log); err == nil {
		t.Error("a sandbox needs the home directory it hides")
	}
}

func TestOnlyTheIsolationsOfOneCommandAreOffered(t *testing.T) {
	sandboxes := registry.Sandboxes(oshosttest.New(), logging.New(&strings.Builder{}, &strings.Builder{}, false, false))
	var names []string
	for _, method := range sandboxes.IsolationMethods() {
		names = append(names, string(method))
		if _, err := sandboxes.Isolator(method); err != nil {
			t.Errorf("%s: %v", method, err)
		}
	}
	if got := strings.Join(names, ","); got != "bwrap,docker,none,seatbelt" {
		t.Errorf("isolations = %s", got)
	}
	for _, method := range []provision.ProvisionMethod{provision.ProvisionNone, provision.ProvisionNix, provision.ProvisionCommand} {
		if _, err := sandboxes.Provisioner(method); err != nil {
			t.Errorf("%s: %v", method, err)
		}
	}
	credentials := registry.Credentials(oshosttest.New())
	if got := strings.Join(credentials.Schemes(), ","); got != "env,file,keychain,pass,secret-service" {
		t.Errorf("credential schemes = %s", got)
	}
	for _, scheme := range credentials.Schemes() {
		if _, err := credentials.Store(scheme); err != nil {
			t.Errorf("%s: %v", scheme, err)
		}
	}
}

func TestTheBridgeRunsOverTheHost(t *testing.T) {
	host := oshosttest.New()
	if bridge := registry.Bridge(host); bridge.Host != host {
		t.Errorf("bridge = %+v", bridge)
	}
}
