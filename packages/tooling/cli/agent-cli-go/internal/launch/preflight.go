package launch

import (
	"fmt"
	"strings"

	credshared "github.com/xonovex/platform/packages/library/shared-launch-go/pkg/credential"
	"github.com/xonovex/platform/packages/library/shared-launch-go/pkg/sandbox"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/runconfig"
)

// defaultSource is the source of an axis no layer set.
const defaultSource = "default"

// Report is the outcome of a preflight: what a run would do and whether it could
// do it. It holds no secret.
type Report struct {
	// Ready is true when the run could start: the plan resolved, every required
	// guarantee holds and every check passed.
	Ready bool `json:"ready"`
	// Problem is why the plan did not resolve, empty when it did.
	Problem    string            `json:"problem,omitempty"`
	Axes       []AxisChoice      `json:"axes"`
	Guarantees []GuaranteeStatus `json:"guarantees"`
	Checks     []Check           `json:"checks"`
	// Configuration is the layer that decided each configuration value.
	Configuration []string `json:"configuration"`
}

// AxisChoice is the choice on one axis and the layer that made it.
type AxisChoice struct {
	Axis   string `json:"axis"`
	Choice string `json:"choice"`
	Source string `json:"source"`
}

// GuaranteeStatus is the evaluation of one guarantee. Reason says why a guarantee
// is not provided.
type GuaranteeStatus struct {
	Guarantee string `json:"guarantee"`
	Required  bool   `json:"required"`
	Provided  bool   `json:"provided"`
	Reason    string `json:"reason,omitempty"`
}

// Check is one readiness check of the host.
type Check struct {
	Name   string `json:"name"`
	OK     bool   `json:"ok"`
	Detail string `json:"detail,omitempty"`
}

// Preflight resolves the run the options describe and reports each choice, each
// guarantee and the readiness of the host, without starting anything: it reads
// the configuration, plans, and checks that the tools, credential stores and
// plugins the run needs are there. A credential is checked for its store and never
// read, and the overlay is planned and never written.
func Preflight(deps Dependencies, options Options) Report {
	report := Report{Axes: []AxisChoice{}, Guarantees: []GuaranteeStatus{}, Checks: []Check{}, Configuration: []string{}}
	res, err := resolve(deps, options, true)
	if err == nil {
		err = res.planErr
	}
	if err != nil {
		report.Problem = err.Error()
		return report
	}
	report.Configuration = res.config.Provenance()
	plan := res.plan
	report.Axes = axisChoices(plan, res.config)
	report.Guarantees = guaranteeStatuses(plan)
	report.Checks = hostChecks(deps, plan)
	report.Ready = readyReport(report)
	return report
}

func readyReport(report Report) bool {
	for _, guarantee := range report.Guarantees {
		if guarantee.Required && !guarantee.Provided {
			return false
		}
	}
	for _, check := range report.Checks {
		if !check.OK {
			return false
		}
	}
	return true
}

func axisChoices(plan Plan, resolved runconfig.Resolved) []AxisChoice {
	layers := resolved.Layers()
	source := func(key runconfig.Key) string {
		if names := layers[key]; len(names) > 0 {
			parts := make([]string, len(names))
			for i, name := range names {
				parts[i] = string(name)
			}
			return strings.Join(parts, ", ")
		}
		return defaultSource
	}
	axes := []AxisChoice{
		{"agent", plan.Agent.Binary, source(runconfig.KeyAgent)},
		{"isolation", string(plan.IsolationName), source(runconfig.KeyIsolation)},
		{"provision", string(plan.ProvisionName), source(runconfig.KeyProvision)},
		{"network", string(plan.Run.Network), source(runconfig.KeyNetwork)},
	}
	if plan.Provider != nil {
		axes = append(axes, AxisChoice{"provider", plan.Provider.Name, source(runconfig.KeyProvider)})
		for _, slot := range plan.Provider.Slots {
			value := slot.ID
			if slot.Effort != "" {
				value += " (effort " + slot.Effort + ")"
			}
			axes = append(axes, AxisChoice{"model " + slot.Name, value, slot.Source})
		}
	}
	workspace := "in-place"
	if plan.Checkout != nil {
		workspace = string(plan.Checkout.VCS) + " checkout"
	}
	axes = append(axes, AxisChoice{"workspace", workspace, defaultSource})
	if plan.Terminal != nil {
		axes = append(axes, AxisChoice{"terminal", string(plan.Terminal.Type), defaultSource})
	}
	return axes
}

func guaranteeStatuses(plan Plan) []GuaranteeStatus {
	var statuses []GuaranteeStatus
	for _, outcome := range evaluatePolicy(plan) {
		status := GuaranteeStatus{Guarantee: string(outcome.Guarantee), Required: outcome.Required, Provided: outcome.Provided}
		if outcome.Err != nil {
			status.Reason = outcome.Err.Error()
		}
		statuses = append(statuses, status)
	}
	return statuses
}

// hostChecks tests what the run needs from the host: the isolation tool, the
// terminal, the version control, the credential stores and the overlay sources.
func hostChecks(deps Dependencies, plan Plan) []Check {
	var checks []Check
	isolation := Check{Name: "isolation " + string(plan.IsolationName), OK: true}
	if available, err := plan.Isolator.Available(); err != nil || !available {
		isolation.OK = false
		isolation.Detail = fmt.Sprintf("not available; available isolations: %v", sandbox.AvailableIsolations(deps.Sandboxes))
	}
	checks = append(checks, isolation)
	if plan.Terminal != nil {
		check := Check{Name: "terminal " + string(plan.Terminal.Type), OK: true}
		if _, err := selectTerminal(deps.Terminals, plan.Terminal); err != nil {
			check.OK, check.Detail = false, err.Error()
		}
		checks = append(checks, check)
	}
	if plan.Checkout != nil {
		check := Check{Name: "workspace " + string(plan.Checkout.VCS), OK: true}
		if vcs, err := selectWorkspace(deps.Workspaces, plan.Checkout); err != nil {
			check.OK, check.Detail = false, err.Error()
		} else if !vcs.Available() {
			check.OK, check.Detail = false, "not installed"
		}
		checks = append(checks, check)
	}
	for _, binding := range plan.Credentials {
		checks = append(checks, credentialCheck(deps.Credentials, binding))
	}
	if plan.Overlay != nil {
		checks = append(checks, overlayCheck(deps, plan))
	}
	return checks
}

// credentialCheck tests that the store of a binding exists and is available. It
// does not read the credential, so it prompts no keychain.
func credentialCheck(registry *credshared.Registry, binding credshared.Binding) Check {
	check := Check{Name: "credential " + binding.Name + " from " + binding.Scheme, OK: true}
	if registry == nil {
		check.OK, check.Detail = false, "this build has no credential stores"
		return check
	}
	store, err := registry.Store(binding.Scheme)
	if err != nil {
		check.OK, check.Detail = false, err.Error()
		return check
	}
	if !store.Available() {
		check.OK, check.Detail = false, "the store is not available"
	}
	return check
}

// overlayCheck plans the overlay as a dry run, which reads and verifies each
// plugin and directory and writes nothing.
func overlayCheck(deps Dependencies, plan Plan) Check {
	check := Check{Name: "overlay", OK: true}
	request := *plan.Overlay
	request.DryRun = true
	result, err := prepareOverlay(deps.Overlays, &request)
	if err != nil {
		check.OK, check.Detail = false, err.Error()
		return check
	}
	result.Cleanup()
	return check
}
