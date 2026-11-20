package overlay

import (
	"errors"
	"io/fs"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/agents"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/artifact"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/overlay/shared"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/types"
	"github.com/xonovex/platform/packages/library/shared-launch-go/pkg/oshost/oshosttest"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/launch"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/origin/plugins"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/plugin"
)

const store = "/home/user/.local/share/agentic/plugins"

// host wraps the fake so a test can fail one operation.
type host struct {
	*oshosttest.Fake
	fail func(op, path string) error
}

func (h host) check(op, path string) error {
	if h.fail == nil {
		return nil
	}
	return h.fail(op, path)
}

func (h host) Stat(p string) (fs.FileInfo, error) {
	if err := h.check("stat", p); err != nil {
		return nil, err
	}
	return h.Fake.Stat(p)
}

func (h host) MkdirTemp(pattern string) (string, error) {
	if err := h.check("mkdirtemp", pattern); err != nil {
		return "", err
	}
	return h.Fake.MkdirTemp(pattern)
}

func (h host) MkdirAll(p string, perm fs.FileMode) error {
	if err := h.check("mkdir", p); err != nil {
		return err
	}
	return h.Fake.MkdirAll(p, perm)
}

func (h host) WriteFile(p string, data []byte, perm fs.FileMode) error {
	if err := h.check("write", p); err != nil {
		return err
	}
	return h.Fake.WriteFile(p, data, perm)
}

var errInjected = errors.New("injected")

func failOn(op string) func(string, string) error {
	return func(got, _ string) error {
		if got == op {
			return errInjected
		}
		return nil
	}
}

func machine() *oshosttest.Fake {
	fake := oshosttest.New()
	fake.AddDir("/work/repo").
		AddFile("/src/hello/.claude-plugin/plugin.json", `{"name":"hello","version":"1.0.0"}`, 0o644).
		AddFile("/src/hello/skills/greet/SKILL.md", "---\nname: greet\ndescription: Use when greeting.\n---\n# Greet\n", 0o644).
		AddFile("/src/hello/skills/greet/run.sh", "#!/bin/sh\n", 0o755).
		AddFile("/extras/AGENTS.md", "Extra rules.\n", 0o644).
		AddFile("/extras/commands/fix.md", "---\ndescription: Fix it.\n---\nFix the bug.\n", 0o644)
	return fake
}

func builder(h host) Builder {
	return Builder{Host: h, Origins: plugins.DefaultRegistry(h.Fake)}
}

func installHello(t *testing.T, b Builder) {
	t.Helper()
	dir, err := plugin.NewStore(b.Host, store)
	if err != nil {
		t.Fatal(err)
	}
	service := plugin.Service{Store: dir, Files: b.Host, Origins: b.Origins}
	fetched, err := service.Fetch("/src/hello")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.Commit(fetched, nil); err != nil {
		t.Fatal(err)
	}
}

func request(t *testing.T, agentType types.AgentType) launch.OverlayRequest {
	t.Helper()
	agent, err := agents.GetAgent(agentType)
	if err != nil {
		t.Fatal(err)
	}
	return launch.OverlayRequest{Agent: agent, WorkDir: "/work/repo", Plugins: []string{"hello"}, Directories: []string{"/extras"}}
}

func TestPrepareWritesTheOverlayOfAPluginAndADirectory(t *testing.T) {
	h := host{Fake: machine()}
	b := builder(h)
	installHello(t, b)

	result, err := b.Prepare(request(t, types.AgentClaude))
	if err != nil {
		t.Fatalf("Prepare() error = %v", err)
	}
	if result.Root != "/tmp/agent-cli-overlay-1" || !slices.Equal(result.Arguments, []string{"--add-dir", result.Root}) || !slices.Equal(result.RoBindPaths, []string{result.Root}) || len(result.BindPaths) != 0 {
		t.Errorf("result = %+v", result)
	}
	for _, path := range []string{"/.claude/skills/greet/SKILL.md", "/.claude/skills/greet/run.sh", "/.claude/commands/fix.md", "/CLAUDE.md"} {
		if !h.Exists(result.Root + path) {
			t.Errorf("%s is missing", path)
		}
	}
	if info, _ := h.Stat(result.Root + "/.claude/skills/greet/run.sh"); info.Mode().Perm() != 0o755 {
		t.Errorf("run.sh mode = %v", info.Mode())
	}
	if len(result.Sources) != 2 || result.Sources[0].Kind != launch.SourcePlugin || result.Sources[0].Name != "hello" ||
		result.Sources[1].Kind != launch.SourceDirectory || result.Sources[1].Name != "/extras" {
		t.Errorf("sources = %+v, want the plugin and the directory", result.Sources)
	}
	for _, source := range result.Sources {
		if !artifact.ValidDigest(source.Digest) {
			t.Errorf("source %s digest = %q, want a sha256 digest", source.Name, source.Digest)
		}
	}
	if len(result.Summary) != 1 || !strings.Contains(result.Summary[0], "placed 4 files") {
		t.Errorf("summary = %v", result.Summary)
	}
	result.Cleanup()
	if h.Exists(result.Root) {
		t.Error("Cleanup left the overlay")
	}
}

func TestPrepareYieldsToTheProjectInstructionFile(t *testing.T) {
	fake := machine().AddFile("/work/repo/CLAUDE.md", "Project.\n", 0o644)
	b := builder(host{Fake: fake})
	installHello(t, b)
	result, err := b.Prepare(request(t, types.AgentClaude))
	if err != nil {
		t.Fatal(err)
	}
	if fake.Exists(result.Root+"/CLAUDE.md") || !strings.Contains(strings.Join(result.Summary, "\n"), "skipped instruction") {
		t.Errorf("the project file did not win: %v", result.Summary)
	}
}

func TestADryRunWritesNothing(t *testing.T) {
	fake := machine()
	b := builder(host{Fake: fake})
	installHello(t, b)
	req := request(t, types.AgentClaude)
	req.DryRun = true
	result, err := b.Prepare(req)
	if err != nil {
		t.Fatal(err)
	}
	if result.Root != dryRunRoot || fake.Exists(dryRunRoot) || fake.Exists("/tmp/agent-cli-overlay-1") || !strings.Contains(result.Summary[0], "would place") {
		t.Errorf("result = %+v", result)
	}
	result.Cleanup()
}

func TestAWritableDialectGetsAWritableMount(t *testing.T) {
	b := builder(host{Fake: machine()})
	req := request(t, types.AgentOpencode)
	req.Plugins = nil
	result, err := b.Prepare(req)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(result.BindPaths, []string{result.Root}) || len(result.RoBindPaths) != 0 || result.Environment[0] != "OPENCODE_CONFIG_DIR="+result.Root {
		t.Errorf("result = %+v", result)
	}
}

func TestAnOverlayWithoutArtifactsPlacesNothingAndLeavesNoDirectory(t *testing.T) {
	fake := machine().AddDir("/empty")
	b := builder(host{Fake: fake})
	req := request(t, types.AgentClaude)
	req.Plugins, req.Directories = nil, []string{"/empty"}
	result, err := b.Prepare(req)
	if err != nil || result.Root != "" || len(result.Arguments) != 0 {
		t.Fatalf("result = %+v, %v", result, err)
	}
	if fake.Exists("/tmp/agent-cli-overlay-1") {
		t.Error("an empty overlay left its directory")
	}
	result.Cleanup()
}

func TestThePluginStoreFollowsTheRequestAndTheEnvironment(t *testing.T) {
	fake := machine()
	fake.Env["XDG_DATA_HOME"] = "/data"
	b := builder(host{Fake: fake})
	dir, _ := plugin.NewStore(fake, "/data/agentic/plugins")
	service := plugin.Service{Store: dir, Files: fake, Origins: b.Origins}
	fetched, _ := service.Fetch("/src/hello")
	if _, err := service.Commit(fetched, nil); err != nil {
		t.Fatal(err)
	}
	req := request(t, types.AgentClaude)
	req.Directories = nil
	if _, err := b.Prepare(req); err != nil {
		t.Errorf("the environment store: %v", err)
	}
	req.Store = "/data/agentic/plugins"
	if _, err := b.Prepare(req); err != nil {
		t.Errorf("the requested store: %v", err)
	}
	req.Store = "relative"
	if _, err := b.Prepare(req); err == nil {
		t.Error("a relative store passed")
	}
}

func TestPrepareFailsClosed(t *testing.T) {
	cases := map[string]func(t *testing.T) (Builder, launch.OverlayRequest){
		"no dialect": func(t *testing.T) (Builder, launch.OverlayRequest) {
			req := request(t, types.AgentClaude)
			req.Agent = &types.AgentConfig{Type: "nope"}
			return builder(host{Fake: machine()}), req
		},
		"plugin not installed": func(t *testing.T) (Builder, launch.OverlayRequest) {
			return builder(host{Fake: machine()}), request(t, types.AgentClaude)
		},
		"tampered plugin": func(t *testing.T) (Builder, launch.OverlayRequest) {
			h := host{Fake: machine()}
			b := builder(h)
			installHello(t, b)
			lock, err := h.ReadFile(store + "/lock.json")
			if err != nil {
				t.Fatalf("read the lock: %v", err)
			}
			digest := regexp.MustCompile(`sha256:([0-9a-f]{64})`).FindSubmatch(lock)
			if digest == nil {
				t.Fatalf("no digest in the lock %s", lock)
			}
			if err := h.WriteFile(store+"/content/"+string(digest[1])+"/skills/greet/SKILL.md", []byte("tampered"), 0o644); err != nil {
				t.Fatal(err)
			}
			return b, request(t, types.AgentClaude)
		},
		"no home": func(t *testing.T) (Builder, launch.OverlayRequest) {
			fake := machine()
			fake.HomeErr = errors.New("no home")
			return builder(host{Fake: fake}), request(t, types.AgentClaude)
		},
		"missing directory": func(t *testing.T) (Builder, launch.OverlayRequest) {
			req := request(t, types.AgentClaude)
			req.Plugins, req.Directories = nil, []string{"/missing"}
			return builder(host{Fake: machine()}), req
		},
		"colliding names": func(t *testing.T) (Builder, launch.OverlayRequest) {
			fake := machine().AddFile("/extras/skills/greet/SKILL.md", "---\nname: greet\ndescription: Use when clashing.\n---\n", 0o644)
			b := builder(host{Fake: fake})
			installHello(t, b)
			return b, request(t, types.AgentClaude)
		},
		"bad skill": func(t *testing.T) (Builder, launch.OverlayRequest) {
			fake := machine().AddFile("/extras/skills/Bad Name/SKILL.md", "x", 0o644)
			req := request(t, types.AgentClaude)
			req.Plugins = nil
			return builder(host{Fake: fake}), req
		},
		"no temporary directory": func(t *testing.T) (Builder, launch.OverlayRequest) {
			req := request(t, types.AgentClaude)
			req.Plugins = nil
			return builder(host{machine(), failOn("mkdirtemp")}), req
		},
		"unreadable project": func(t *testing.T) (Builder, launch.OverlayRequest) {
			req := request(t, types.AgentClaude)
			req.Plugins = nil
			return builder(host{machine(), func(op, p string) error {
				if op == "stat" && strings.HasSuffix(p, "CLAUDE.md") {
					return errInjected
				}
				return nil
			}}), req
		},
		"cannot create a directory": func(t *testing.T) (Builder, launch.OverlayRequest) {
			req := request(t, types.AgentClaude)
			req.Plugins = nil
			return builder(host{machine(), failOn("mkdir")}), req
		},
		"cannot write a file": func(t *testing.T) (Builder, launch.OverlayRequest) {
			req := request(t, types.AgentClaude)
			req.Plugins = nil
			return builder(host{machine(), failOn("write")}), req
		},
	}
	for name, setup := range cases {
		b, req := setup(t)
		if _, err := b.Prepare(req); err == nil {
			t.Errorf("%s: Prepare() error = nil", name)
		}
		if fake, ok := b.Host.(host); ok && name != "no temporary directory" && fake.Exists("/tmp/agent-cli-overlay-1") {
			t.Errorf("%s: a failed overlay stayed on disk", name)
		}
	}
}

type failingDialect struct{}

func (failingDialect) Facts() shared.Facts { return shared.Facts{} }
func (failingDialect) Place(shared.Input) (shared.Placement, error) {
	return shared.Placement{}, errInjected
}

func TestAFailingDialectStopsThePlacement(t *testing.T) {
	b := builder(host{Fake: machine()})
	if _, err := b.place(failingDialect{}, artifact.Set{}, launch.OverlayRequest{WorkDir: "/work/repo"}, "/tmp/o"); !errors.Is(err, errInjected) {
		t.Errorf("place() error = %v", err)
	}
}
