package launch

import (
	"errors"
	"strings"
	"testing"

	credshared "github.com/xonovex/platform/packages/library/shared-launch-go/pkg/credential"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/runconfig"
)

func findCheck(report Report, prefix string) (Check, bool) {
	for _, check := range report.Checks {
		if strings.HasPrefix(check.Name, prefix) {
			return check, true
		}
	}
	return Check{}, false
}

func findGuarantee(report Report, name string) GuaranteeStatus {
	for _, guarantee := range report.Guarantees {
		if guarantee.Guarantee == name {
			return guarantee
		}
	}
	return GuaranteeStatus{}
}

func TestPreflightReportsAReadyRunWithoutRunningIt(t *testing.T) {
	f := newExecuteFixture()
	options := bwrapOptions()
	options.Explicit = runconfig.KeysOf(runconfig.KeyIsolation)

	report := Preflight(f.deps(), options)

	if !report.Ready || report.Problem != "" {
		t.Fatalf("report = %+v, want ready", report)
	}
	if f.sandbox.isolator.ran || len(f.records.records) != 0 || len(f.vcs.configs) != 0 {
		t.Error("preflight ran the agent, wrote a record or created a checkout")
	}
	axes := map[string]AxisChoice{}
	for _, axis := range report.Axes {
		axes[axis.Axis] = axis
	}
	if axes["isolation"].Choice != "bwrap" || axes["isolation"].Source != "flags" || axes["network"].Source != "default" || axes["workspace"].Choice != "in-place" {
		t.Errorf("axes = %+v", axes)
	}
	if check, ok := findCheck(report, "isolation"); !ok || !check.OK {
		t.Errorf("isolation check = %+v, %v", check, ok)
	}
}

func TestPreflightExplainsAnUnmetGuarantee(t *testing.T) {
	f := newExecuteFixture()
	options := bwrapOptions()
	options.Policy.RequireKernelIsolation = true

	report := Preflight(f.deps(), options)

	if report.Ready {
		t.Fatal("report is ready although kernel isolation is required and absent")
	}
	kernel := findGuarantee(report, "kernel-isolation")
	if !kernel.Required || kernel.Provided || kernel.Reason == "" {
		t.Errorf("kernel guarantee = %+v, want required, unmet and explained", kernel)
	}
	if report.Problem != "" {
		t.Errorf("Problem = %q; an unmet guarantee is not a plan failure", report.Problem)
	}
}

func TestPreflightPassesAGuaranteeThatIsOnlyRequestedWhenItHolds(t *testing.T) {
	f := newExecuteFixture()
	options := bwrapOptions()
	options.Network = "none"
	options.Policy.RequireEgressRestricted = true

	report := Preflight(f.deps(), options)

	if egress := findGuarantee(report, "egress-restricted"); !egress.Required || !egress.Provided || egress.Reason != "" {
		t.Errorf("egress guarantee = %+v, want met", egress)
	}
}

func TestPreflightReportsAMissingHostTool(t *testing.T) {
	f := newExecuteFixture()
	f.sandbox.isolator.unavailable = true
	f.terminal.missing = true
	options := bwrapOptions()
	options.Terminal = "tmux"
	options.WorktreeBranch = "feature"

	report := Preflight(f.deps(), options)

	if report.Ready {
		t.Fatal("report is ready with the isolation tool and the terminal missing")
	}
	for _, prefix := range []string{"isolation", "terminal"} {
		if check, ok := findCheck(report, prefix); !ok || check.OK || check.Detail == "" {
			t.Errorf("%s check = %+v, %v; want a failure with a detail", prefix, check, ok)
		}
	}
	if check, ok := findCheck(report, "workspace"); !ok || !check.OK {
		t.Errorf("workspace check = %+v, %v; want ok", check, ok)
	}
}

func TestPreflightChecksCredentialStoresWithoutReadingThem(t *testing.T) {
	f := newExecuteFixture()
	var locator string
	deps := f.deps()
	deps.Credentials = credshared.NewRegistry().Register("test", func() credshared.Store { return fakeStore{value: "s3cret", seen: &locator} })
	options := bwrapOptions()
	options.Credentials = []string{"TOKEN=test:vault/token"}

	report := Preflight(deps, options)

	if !report.Ready || locator != "" {
		t.Errorf("ready = %v, locator read = %q; want ready and no lookup", report.Ready, locator)
	}

	for name, tc := range map[string]struct {
		credentials *credshared.Registry
		binding     string
		detail      string
	}{
		"no stores":     {nil, "TOKEN=test:x", "no credential stores"},
		"unknown store": {credshared.NewRegistry(), "TOKEN=ghost:x", "ghost"},
	} {
		deps.Credentials = tc.credentials
		options.Credentials = []string{tc.binding}
		report := Preflight(deps, options)
		check, ok := findCheck(report, "credential")
		if report.Ready || !ok || check.OK || !strings.Contains(check.Detail, tc.detail) {
			t.Errorf("%s: check = %+v, ready = %v; want a failure naming %q", name, check, report.Ready, tc.detail)
		}
	}
}

func TestPreflightReportsAnUnavailableCredentialStore(t *testing.T) {
	f := newExecuteFixture()
	deps := f.deps()
	deps.Credentials = credshared.NewRegistry().Register("test", func() credshared.Store { return offlineStore{} })
	options := bwrapOptions()
	options.Credentials = []string{"TOKEN=test:x"}

	report := Preflight(deps, options)

	if check, _ := findCheck(report, "credential"); report.Ready || check.OK || !strings.Contains(check.Detail, "not available") {
		t.Errorf("check = %+v, ready = %v", check, report.Ready)
	}
}

type offlineStore struct{ fakeStore }

func (offlineStore) Available() bool { return false }

func TestPreflightPlansTheOverlayAsADryRun(t *testing.T) {
	f := newExecuteFixture()
	overlays := &fakeOverlays{}
	deps := f.deps()
	deps.Overlays = overlays
	options := bwrapOptions()
	options.Plugins = []string{"plugin-core"}
	if overlays.cleaned {
		t.Fatal("cleaned before use")
	}

	if report := Preflight(deps, options); !report.Ready {
		t.Fatalf("report = %+v, want ready", report)
	}
	if !overlays.called || !overlays.request.DryRun {
		t.Errorf("request = %+v, want a dry run request", overlays.request)
	}

	if !overlays.cleaned {
		t.Error("the dry run overlay was not cleaned")
	}
	overlays.err = errors.New("plugin plugin-core is not installed")
	report := Preflight(deps, options)
	if check, ok := findCheck(report, "overlay"); report.Ready || !ok || check.OK || !strings.Contains(check.Detail, "not installed") {
		t.Errorf("overlay check = %+v, ready = %v", check, report.Ready)
	}

	deps.Overlays = nil
	if report := Preflight(deps, options); report.Ready {
		t.Error("report is ready for an overlay this build cannot build")
	}
}

func TestPreflightReportsAPlanThatDoesNotResolve(t *testing.T) {
	f := newExecuteFixture()
	options := bwrapOptions()
	options.Agent = "no-such-agent"

	report := Preflight(f.deps(), options)

	if report.Ready || !strings.Contains(report.Problem, "no-such-agent") || report.Axes == nil || report.Guarantees == nil || report.Checks == nil {
		t.Errorf("report = %+v, want a problem and empty lists", report)
	}
}

func TestPreflightReportsAConfigurationThatDoesNotLoad(t *testing.T) {
	f := newExecuteFixture()
	f.host.AddFile(managedFile, runFile(`{"network":"none"}`), 0o644)
	options := bwrapOptions()
	options.Network = "host"
	options.Explicit = runconfig.KeysOf(runconfig.KeyNetwork)

	report := Preflight(f.deps(), options)

	if report.Ready || !strings.Contains(report.Problem, "but the managed file") {
		t.Errorf("report = %+v, want the managed contradiction", report)
	}
}

func TestPreflightReportsTheLayerThatDecidedAnAxis(t *testing.T) {
	f := newExecuteFixture()
	f.host.AddFile(userFile, runFile(`{"isolation":"bwrap"}`), 0o600)
	options := baseOptions()

	report := Preflight(f.deps(), options)

	var source string
	for _, axis := range report.Axes {
		if axis.Axis == "isolation" {
			source = axis.Source
		}
	}
	if !report.Ready || source != "user" {
		t.Errorf("ready, isolation source = %v, %q; want ready and user", report.Ready, source)
	}
	if len(report.Configuration) == 0 {
		t.Error("the report holds no provenance")
	}
}
