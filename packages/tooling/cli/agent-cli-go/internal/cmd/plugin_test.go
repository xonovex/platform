package cmd

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/plugin"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/system/systemtest"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/trust"
)

const (
	defaultStore = "/home/user/.local/share/agent-cli/plugins"
	helloPlugin  = `{"name":"hello","version":"1.0.0","description":"d"}`
)

func pluginRoot() *testRoot {
	r := newTestRoot()
	r.host.AddFile("/src/hello/plugin.json", helloPlugin, 0o644).
		AddFile("/src/hello/skills/greet/SKILL.md", "---\nname: greet\ndescription: Use when greeting.\n---\n", 0o644).
		AddFile("/market/.claude-plugin/marketplace.json", `{"name":"main","plugins":[{"name":"hello","description":"Says hello.","source":"./hello"},{"name":"quiet","source":"./quiet"}]}`, 0o644).
		AddFile("/market/hello/plugin.json", helloPlugin, 0o644).
		AddFile("/market/quiet/plugin.json", `{"name":"quiet","version":"2.0.0"}`, 0o644)
	return r
}

func (r *testRoot) digest(t *testing.T, source string) string {
	t.Helper()
	err := r.execute("plugin", "install", source)
	if err == nil {
		t.Fatal("an install without an approval passed")
	}
	_, after, found := strings.Cut(err.Error(), "--approve-digest ")
	if !found {
		t.Fatalf("the error names no digest: %v", err)
	}
	return after
}

func TestPluginInstallPreviewsThenInstallsTheApprovedDigest(t *testing.T) {
	r := pluginRoot()
	digest := r.digest(t, "/src/hello")
	if !strings.Contains(r.out.String(), "Plugin: hello 1.0.0") || !strings.Contains(r.out.String(), "Digest: "+digest) || !strings.Contains(r.out.String(), "Files: 2") {
		t.Errorf("preview = %q", r.out.String())
	}
	if r.host.Exists(defaultStore + "/lock.json") {
		t.Error("the preview installed the plugin")
	}

	r.out.Reset()
	if err := r.execute("plugin", "install", "/src/hello", "--approve-digest", digest); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(r.out.String(), "Installed hello 1.0.0 ("+digest+")") {
		t.Errorf("output = %q", r.out.String())
	}

	r.out.Reset()
	if err := r.execute("plugin", "list"); err != nil || !strings.Contains(r.out.String(), "hello 1.0.0 "+digest+" /src/hello") {
		t.Errorf("list = %q, %v", r.out.String(), err)
	}
	r.out.Reset()
	if err := r.execute("plugin", "remove", "hello"); err != nil || r.out.String() != "Removed hello\n" {
		t.Errorf("remove = %q, %v", r.out.String(), err)
	}
	r.out.Reset()
	if err := r.execute("plugin", "list"); err != nil || r.out.String() != "No plugins installed.\n" {
		t.Errorf("list = %q, %v", r.out.String(), err)
	}
}

func TestPluginInstallFailsClosed(t *testing.T) {
	r := pluginRoot()
	r.host.AddFile("/src/hooked/plugin.json", `{"name":"hooked","version":"1.0.0"}`, 0o644).AddFile("/src/hooked/hooks/hooks.json", "{}", 0o644)
	digest := r.digest(t, "/src/hooked")
	if !strings.Contains(r.out.String(), "Carries: hooks") {
		t.Errorf("preview = %q", r.out.String())
	}
	cases := map[string][]string{
		"no grant":      {"plugin", "install", "/src/hooked", "--approve-digest", digest},
		"wrong digest":  {"plugin", "install", "/src/hello", "--approve-digest", digest},
		"bad grant":     {"plugin", "install", "/src/hooked", "--allow", "root"},
		"no source":     {"plugin", "install", "/nowhere"},
		"no argument":   {"plugin", "install"},
		"remove nobody": {"plugin", "remove", "ghost"},
	}
	for name, args := range cases {
		if err := r.execute(args...); err == nil {
			t.Errorf("%s: error = nil", name)
		}
	}
	if err := r.execute("plugin", "install", "/src/hooked", "--approve-digest", digest, "--allow", "hooks"); err != nil {
		t.Errorf("a granted install failed: %v", err)
	}
}

func TestPluginInstallShowsTheMarketplaceAndDependencies(t *testing.T) {
	r := pluginRoot()
	if err := r.execute("marketplace", "add", "/market"); err != nil {
		t.Fatal(err)
	}
	r.host.AddFile("/market/hello/hooks/hooks.json", "{}", 0o644).AddFile("/market/hello/plugin.json", `{"name":"hello","version":"1.0.0","dependencies":["quiet"]}`, 0o644)
	r.out.Reset()
	_ = r.execute("plugin", "install", "hello@main")
	for _, want := range []string{"Marketplace: main", "Depends on: quiet", "Carries: hooks"} {
		if !strings.Contains(r.out.String(), want) {
			t.Errorf("preview lacks %q: %s", want, r.out.String())
		}
	}
}

func TestMarketplaceCommands(t *testing.T) {
	r := pluginRoot()
	if err := r.execute("marketplace", "list"); err != nil || r.out.String() != "No marketplaces registered.\n" {
		t.Errorf("list = %q, %v", r.out.String(), err)
	}
	r.out.Reset()
	if err := r.execute("marketplace", "add", "/market"); err != nil || !strings.Contains(r.out.String(), "Registered marketplace main (/market) with 2 plugins") {
		t.Errorf("add = %q, %v", r.out.String(), err)
	}
	r.out.Reset()
	if err := r.execute("marketplace", "list"); err != nil || r.out.String() != "main /market\n" {
		t.Errorf("list = %q, %v", r.out.String(), err)
	}
	r.out.Reset()
	if err := r.execute("marketplace", "list", "main"); err != nil || r.out.String() != "hello - Says hello.\nquiet\n" {
		t.Errorf("browse = %q, %v", r.out.String(), err)
	}

	digest := r.digest(t, "hello@main")
	if err := r.execute("plugin", "install", "hello@main", "--approve-digest", digest); err != nil {
		t.Fatal(err)
	}
	r.out.Reset()
	if err := r.execute("plugin", "list"); err != nil || !strings.Contains(r.out.String(), "(marketplace main)") {
		t.Errorf("plugin list = %q, %v", r.out.String(), err)
	}
	if err := r.execute("marketplace", "remove", "main"); err == nil {
		t.Error("a marketplace in use was removed")
	}
	_ = r.execute("plugin", "remove", "hello")
	r.out.Reset()
	if err := r.execute("marketplace", "remove", "main"); err != nil || r.out.String() != "Removed marketplace main\n" {
		t.Errorf("remove = %q, %v", r.out.String(), err)
	}
	for _, args := range [][]string{{"marketplace", "add", "/nowhere"}, {"marketplace", "list", "ghost"}, {"marketplace", "remove", "ghost"}, {"marketplace", "add"}} {
		if err := r.execute(args...); err == nil {
			t.Errorf("%v: error = nil", args)
		}
	}
}

func TestTheStoreFlagAndTheEnvironmentChooseTheStore(t *testing.T) {
	r := pluginRoot()
	digest := r.digest(t, "/src/hello")
	if err := r.execute("plugin", "install", "/src/hello", "--approve-digest", digest, "--store", "relative/store"); err != nil {
		t.Fatal(err)
	}
	if !r.host.Exists("/work/relative/store/lock.json") {
		t.Error("a relative store is not resolved against the working directory")
	}
	r.host.Env["XDG_DATA_HOME"] = "/data"
	if err := r.execute("plugin", "install", "/src/hello", "--approve-digest", digest); err != nil || !r.host.Exists("/data/agent-cli/plugins/lock.json") {
		t.Errorf("the environment store: %v", err)
	}
	if err := r.execute("plugin", "list", "--store", "/abs/../x"); err != nil {
		t.Errorf("an unclean store path is cleaned: %v", err)
	}
}

type brokenHost struct {
	*systemtest.Fake
	cwd, home, relative bool
}

func (b brokenHost) Getwd() (string, error) {
	if b.cwd {
		return "", errors.New("no cwd")
	}
	if b.relative {
		return "relative", nil
	}
	return b.Fake.Getwd()
}

func (b brokenHost) UserHomeDir() (string, error) {
	if b.home {
		return "", errors.New("no home")
	}
	return b.Fake.UserHomeDir()
}

func TestAMissingWorkingOrHomeDirectoryStopsTheCommand(t *testing.T) {
	commands := [][]string{
		{"plugin", "install", "/src"}, {"plugin", "list"}, {"plugin", "remove", "a"},
		{"marketplace", "add", "/m"}, {"marketplace", "list"}, {"marketplace", "list", "a"}, {"marketplace", "remove", "a"},
	}
	hosts := []brokenHost{{Fake: systemtest.New(), cwd: true}, {Fake: systemtest.New(), home: true}, {Fake: systemtest.New(), relative: true}}
	for _, host := range hosts {
		for _, args := range commands {
			if host.relative {
				args = append(append([]string{}, args...), "--store", "s")
			}
			root := NewRootCommand(newTestRoot().deps, PluginPorts{Host: host, Out: io.Discard}, RecordPorts{}, &recordingBridge{})
			root.SetArgs(args)
			root.SetOut(io.Discard)
			root.SetErr(io.Discard)
			if err := root.Execute(); err == nil {
				t.Errorf("%v with %+v: error = nil", args, host)
			}
		}
	}
	// A relative store needs no home directory.
	host := brokenHost{Fake: systemtest.New(), home: true}
	root := NewRootCommand(newTestRoot().deps, PluginPorts{Host: host, Out: io.Discard}, RecordPorts{}, &recordingBridge{})
	root.SetArgs([]string{"plugin", "list", "--store", "s"})
	root.SetOut(io.Discard)
	root.SetErr(io.Discard)
	if err := root.Execute(); err != nil {
		t.Errorf("a relative store: %v", err)
	}
}

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) { return 0, errors.New("closed") }

func TestAFailedWriteToTheOutputIsAnError(t *testing.T) {
	r := pluginRoot()
	digest := r.digest(t, "/src/hello")
	if err := r.execute("plugin", "install", "/src/hello", "--approve-digest", digest); err != nil {
		t.Fatal(err)
	}
	r.ports.Out = failingWriter{}
	_ = r.execute("marketplace", "add", "/market") // registered while the output works
	for _, args := range [][]string{
		{"plugin", "install", "/src/hello", "--approve-digest", digest},
		{"plugin", "list"}, {"plugin", "remove", "hello"},
		{"marketplace", "add", "/market"}, {"marketplace", "list"}, {"marketplace", "list", "main"}, {"marketplace", "remove", "main"},
	} {
		root := NewRootCommand(r.deps, r.ports, RecordPorts{}, r.bridge)
		root.SetArgs(args)
		root.SetOut(io.Discard)
		root.SetErr(io.Discard)
		_ = root.Execute()
	}
}

func TestThePreviewNamesTheRevisionOfARepository(t *testing.T) {
	var out strings.Builder
	printPreview(&out, plugin.Fetched{Source: "https://h/r.git#v1", Resolved: "abc123"})
	if !strings.Contains(out.String(), "Revision: abc123") {
		t.Errorf("preview = %q", out.String())
	}
}

func TestTheAllowFlagCompletesTheCapabilities(t *testing.T) {
	r := pluginRoot()
	root := NewRootCommand(r.deps, r.ports, RecordPorts{}, r.bridge)
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(io.Discard)
	root.SetArgs([]string{"__complete", "plugin", "install", "--allow", ""})
	if err := root.Execute(); err != nil || !strings.Contains(out.String(), "hooks") || !strings.Contains(out.String(), "mcp") {
		t.Errorf("completion = %q, %v", out.String(), err)
	}
}

func terminalRoot(answers ...string) *testRoot {
	r := pluginRoot()
	r.host.Terminal = true
	r.host.Input = answers
	return r
}

func TestPluginInstallAsksOnATerminalAndInstallsOnYes(t *testing.T) {
	for _, answer := range []string{"y", "yes", "Y", " YES \r"} {
		r := terminalRoot(answer)
		if err := r.execute("plugin", "install", "/src/hello"); err != nil {
			t.Fatalf("answer %q: %v", answer, err)
		}
		out := r.out.String()
		if !strings.Contains(out, "Digest: sha256:") || !strings.Contains(out, "Approve this digest? [y/N] ") || !strings.Contains(out, "Installed hello 1.0.0") {
			t.Errorf("answer %q: output = %q", answer, out)
		}
		if !r.host.Exists(defaultStore + "/lock.json") {
			t.Errorf("answer %q: the plugin was not installed", answer)
		}
		if len(r.host.Input) != 0 {
			t.Errorf("answer %q: input left = %q", answer, r.host.Input)
		}
	}
}

func TestPluginInstallDeclinesOtherAnswers(t *testing.T) {
	for _, answer := range []string{"", "n", "no", "yess", "ok", "y es"} {
		r := terminalRoot(answer)
		err := r.execute("plugin", "install", "/src/hello")
		if !errors.Is(err, trust.ErrDeclined) {
			t.Errorf("answer %q: error = %v, want trust.ErrDeclined", answer, err)
		}
		if r.host.Exists(defaultStore + "/lock.json") {
			t.Errorf("answer %q: a declined install changed the store", answer)
		}
	}
}

func TestPluginInstallDeclinesAtTheEndOfTheInput(t *testing.T) {
	r := terminalRoot()
	if err := r.execute("plugin", "install", "/src/hello"); !errors.Is(err, trust.ErrDeclined) {
		t.Errorf("error = %v, want trust.ErrDeclined", err)
	}
	if r.host.Exists(defaultStore + "/lock.json") {
		t.Error("an unanswered install changed the store")
	}
	if !strings.HasSuffix(r.out.String(), "Approve this digest? [y/N] \n") {
		t.Errorf("output = %q, want the question closed with a line ending", r.out.String())
	}
}

func TestPluginInstallStopsWhenTheAnswerCannotBeRead(t *testing.T) {
	r := terminalRoot()
	r.host.InputErr = errors.New("interrupted")
	err := r.execute("plugin", "install", "/src/hello")
	if err == nil || errors.Is(err, trust.ErrDeclined) || !strings.Contains(err.Error(), "interrupted") {
		t.Errorf("error = %v, want the read failure", err)
	}
	if r.host.Exists(defaultStore + "/lock.json") {
		t.Error("an unreadable answer changed the store")
	}
}

func TestPluginInstallWithoutATerminalOnlyPreviews(t *testing.T) {
	r := pluginRoot()
	r.host.Input = []string{"yes"}
	err := r.execute("plugin", "install", "/src/hello")
	if err == nil || !strings.Contains(err.Error(), "--approve-digest ") {
		t.Fatalf("error = %v, want instructions naming --approve-digest", err)
	}
	if strings.Contains(r.out.String(), "Approve this digest?") {
		t.Errorf("output = %q, a question was asked without a terminal", r.out.String())
	}
	if len(r.host.Input) != 1 {
		t.Error("the input was read without a terminal")
	}
	if r.host.Exists(defaultStore + "/lock.json") {
		t.Error("the preview installed the plugin")
	}
}

func TestPluginInstallWithADigestNeverAsks(t *testing.T) {
	r := terminalRoot("n")
	digest := func() string {
		probe := pluginRoot()
		return probe.digest(t, "/src/hello")
	}()
	if err := r.execute("plugin", "install", "/src/hello", "--approve-digest", digest); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(r.out.String(), "Approve this digest?") || len(r.host.Input) != 1 {
		t.Errorf("output = %q, input left = %q, want no question", r.out.String(), r.host.Input)
	}

	mismatch := terminalRoot("yes")
	if err := mismatch.execute("plugin", "install", "/src/hello", "--approve-digest", "sha256:"+strings.Repeat("0", 64)); err == nil {
		t.Error("a mismatching digest installed")
	}
	if strings.Contains(mismatch.out.String(), "Approve this digest?") || len(mismatch.host.Input) != 1 || mismatch.host.Exists(defaultStore+"/lock.json") {
		t.Errorf("a mismatching digest asked or installed: output = %q", mismatch.out.String())
	}
}

func TestPluginInstallKeepsTheCapabilityGrantSeparateFromTheAnswer(t *testing.T) {
	r := terminalRoot("yes")
	r.host.AddFile("/src/hooked/plugin.json", `{"name":"hooked","version":"1.0.0"}`, 0o644).AddFile("/src/hooked/hooks/hooks.json", "{}", 0o644)
	if err := r.execute("plugin", "install", "/src/hooked"); err == nil {
		t.Fatal("a yes granted a capability")
	}
	if r.host.Exists(defaultStore + "/lock.json") {
		t.Error("a plugin with hooks was installed without --allow hooks")
	}
	if !strings.Contains(r.out.String(), "Carries: hooks (each needs --allow)") {
		t.Errorf("preview = %q", r.out.String())
	}

	granted := terminalRoot("yes")
	granted.host.AddFile("/src/hooked/plugin.json", `{"name":"hooked","version":"1.0.0"}`, 0o644).AddFile("/src/hooked/hooks/hooks.json", "{}", 0o644)
	if err := granted.execute("plugin", "install", "/src/hooked", "--allow", "hooks"); err != nil {
		t.Errorf("an answered and granted install failed: %v", err)
	}
}
