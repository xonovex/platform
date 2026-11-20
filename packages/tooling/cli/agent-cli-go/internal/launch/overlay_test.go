package launch

import (
	"errors"
	"slices"
	"strings"
	"testing"
)

type fakeOverlays struct {
	result  OverlayResult
	err     error
	request OverlayRequest
	called  bool
	cleaned bool
}

func (f *fakeOverlays) Prepare(request OverlayRequest) (OverlayResult, error) {
	f.called, f.request = true, request
	result := f.result
	result.Cleanup = func() { f.cleaned = true }
	return result, f.err
}

func overlayOptions() Options {
	options := bwrapOptions()
	options.Plugins = []string{"hello"}
	options.OverlayDirs = []string{"extras"}
	options.Store = "store"
	options.AgentArgs = []string{"review"}
	return options
}

func TestNewPlanResolvesTheOverlayRequest(t *testing.T) {
	p := mustPlan(t, overlayOptions(), "")
	if p.Overlay == nil {
		t.Fatal("no overlay request")
	}
	if p.Overlay.WorkDir != "/work/repo" || p.Overlay.Store != "/work/repo/store" || !slices.Equal(p.Overlay.Directories, []string{"/work/repo/extras"}) || !slices.Equal(p.Overlay.Plugins, []string{"hello"}) || p.Overlay.Agent.Binary != "claude" {
		t.Errorf("overlay = %+v", p.Overlay)
	}

	if p := mustPlan(t, baseOptions(), ""); p.Overlay != nil {
		t.Errorf("a run without plugins has the overlay %+v", p.Overlay)
	}
	options := overlayOptions()
	options.Store = ""
	options.DryRun = true
	if p := mustPlan(t, options, ""); p.Overlay.Store != "" || !p.Overlay.DryRun {
		t.Errorf("overlay = %+v", p.Overlay)
	}
}

func TestNewPlanRejectsAnInvalidOverlay(t *testing.T) {
	cases := map[string]func(*Options){
		"plugin name": func(o *Options) { o.Plugins = []string{"Bad Name"} },
		"empty dir":   func(o *Options) { o.OverlayDirs = []string{""} },
		"login agent": func(o *Options) { o.Agent = "kiro" },
	}
	for name, mutate := range cases {
		options := overlayOptions()
		mutate(&options)
		if _, err := plan(t, options, ""); err == nil {
			t.Errorf("%s: NewPlan() error = nil", name)
		}
	}
}

func TestExecuteAppliesTheOverlayBeforeTheUserArguments(t *testing.T) {
	f := newExecuteFixture()
	overlays := &fakeOverlays{result: OverlayResult{
		Root:        "/tmp/o",
		Environment: []string{"A=1"},
		Arguments:   []string{"--add-dir", "/tmp/o"},
		RoBindPaths: []string{"/tmp/o"},
		BindPaths:   []string{"/tmp/w"},
	}}
	deps := f.deps()
	deps.Overlays = overlays

	if err := Execute(deps, overlayOptions()); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	run := f.sandbox.isolator.commandRun
	if !slices.Equal(run.Process.Argv, []string{"claude", "--permission-mode", "bypassPermissions", "--add-dir", "/tmp/o", "review"}) || !slices.Contains(run.Process.Env, "A=1") || !slices.Contains(run.RoBindPaths, "/tmp/o") || !slices.Contains(run.BindPaths, "/tmp/w") {
		t.Errorf("run = %+v", run)
	}
	if !overlays.called || !overlays.cleaned || overlays.request.DryRun {
		t.Errorf("called, cleaned, dry run = %v %v %v", overlays.called, overlays.cleaned, overlays.request.DryRun)
	}
}

func TestExecuteCleansAnEmptyOverlayOfATerminalSession(t *testing.T) {
	f := newExecuteFixture()
	empty := &fakeOverlays{}
	deps := f.deps()
	deps.Overlays = empty
	options := overlayOptions()
	options.Terminal = "tmux"

	// A terminal run whose overlay placed nothing has nothing to keep.
	if err := Execute(deps, options); err != nil || !empty.cleaned {
		t.Errorf("Execute() = %v, cleaned = %v", err, empty.cleaned)
	}
}

func TestExecuteCleansTheOverlayOnlyWhenTheTerminalReportsTheProgramEnded(t *testing.T) {
	cases := []struct {
		name        string
		waits       bool
		detach      bool
		err         error
		wantCleaned bool
	}{
		{name: "ended", waits: true, wantCleaned: true},
		{name: "ended with a failure", waits: true, err: errors.New("wait failed"), wantCleaned: true},
		{name: "runs on", waits: false},
		{name: "runs on after a failure", waits: false, err: errors.New("attach failed")},
		{name: "detached", waits: true, detach: true},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			f := newExecuteFixture()
			f.terminal.waits, f.terminal.err = test.waits, test.err
			overlays := &fakeOverlays{result: OverlayResult{Root: "/tmp/o"}}
			deps := f.deps()
			deps.Overlays = overlays
			options := overlayOptions()
			options.Terminal = "tmux"
			options.TerminalDetach = test.detach

			if err := Execute(deps, options); (err != nil) != (test.err != nil) {
				t.Fatalf("Execute() error = %v, want %v", err, test.err)
			}
			if overlays.cleaned != test.wantCleaned || strings.Contains(f.log.String(), "The overlay stays at /tmp/o") == test.wantCleaned {
				t.Errorf("cleaned = %v, log = %q, want cleaned %v", overlays.cleaned, f.log.String(), test.wantCleaned)
			}
		})
	}
}

func TestExecuteCleansTheOverlayWhenALaterStepFails(t *testing.T) {
	f := newExecuteFixture()
	f.vcs.err = errors.New("checkout failed")
	overlays := &fakeOverlays{result: OverlayResult{Root: "/tmp/o"}}
	deps := f.deps()
	deps.Overlays = overlays
	options := overlayOptions()
	options.WorktreeBranch = "feature"

	if err := Execute(deps, options); err == nil || !overlays.cleaned {
		t.Errorf("Execute() = %v, cleaned = %v", err, overlays.cleaned)
	}
}

func TestExecuteFailsClosedOnAnOverlayError(t *testing.T) {
	f := newExecuteFixture()
	failure := errors.New("plugin changed on disk")
	deps := f.deps()
	deps.Overlays = &fakeOverlays{err: failure}
	if err := Execute(deps, overlayOptions()); !errors.Is(err, failure) || f.sandbox.isolator.ran {
		t.Errorf("Execute() = %v, ran = %v", err, f.sandbox.isolator.ran)
	}

	deps.Overlays = nil
	if err := Execute(deps, overlayOptions()); err == nil || !strings.Contains(err.Error(), "not available") {
		t.Errorf("Execute() without the port = %v", err)
	}
}

func TestDryRunReportsTheOverlayWithoutRunning(t *testing.T) {
	f := newExecuteFixture()
	overlays := &fakeOverlays{result: OverlayResult{Root: "/tmp/dry", Summary: []string{"  placed 3 files", "skipped skill x"}}}
	deps := f.deps()
	deps.Overlays = overlays
	options := overlayOptions()
	options.DryRun = true

	if err := Execute(deps, options); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !overlays.request.DryRun || f.sandbox.isolator.ran || !overlays.cleaned {
		t.Errorf("dry run = %v, ran = %v, cleaned = %v", overlays.request.DryRun, f.sandbox.isolator.ran, overlays.cleaned)
	}
	for _, want := range []string{"Overlay:", "placed 3 files", "skipped skill x"} {
		if !strings.Contains(f.log.String(), want) {
			t.Errorf("log lacks %q: %s", want, f.log.String())
		}
	}

	// An overlay without a summary adds no section.
	g := newExecuteFixture()
	deps = g.deps()
	deps.Overlays = &fakeOverlays{}
	if err := Execute(deps, options); err != nil || strings.Contains(g.log.String(), "Overlay:") {
		t.Errorf("Execute() = %v, log = %s", err, g.log.String())
	}
}
