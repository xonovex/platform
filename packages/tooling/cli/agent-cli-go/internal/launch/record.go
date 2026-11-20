package launch

import (
	"slices"
	"sort"
	"time"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/policy"
	"github.com/xonovex/platform/packages/library/shared-launch-go/pkg/record"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/runconfig"
)

// writeRecord stores the record of a finished run. A store that fails costs the
// run its record and never its result, so the failure is a warning.
func writeRecord(deps Dependencies, res resolution, started time.Time, result outcome, runErr error) {
	if deps.Records == nil {
		return
	}
	if _, err := deps.Records.Write(buildRecord(res, started, deps.Host.Now(), result, runErr)); err != nil {
		deps.Log.Warning("Could not write the run record: " + err.Error())
	}
}

// buildRecord is the record of a run. It names the credentials and never holds a
// value, and it keeps the layers that decided each key, not the keys' values.
func buildRecord(res resolution, started, ended time.Time, result outcome, runErr error) record.Record {
	plan := res.plan
	r := record.Record{
		StartedAt: started,
		EndedAt:   ended,
		Status:    record.StatusExited,
		Agent: record.Agent{
			Name:         plan.Agent.Binary,
			VersionProbe: describeVersionProbe(plan.Agent),
		},
		Axes: record.Axes{
			Isolation: string(plan.IsolationName),
			Provision: string(plan.ProvisionName),
			Network:   string(plan.Run.Network),
			Workspace: workspaceRecord(plan),
		},
		Guarantees: guaranteeRecords(plan.Capabilities, plan.Policy),
		WorkDir:    plan.DisplayDir(),
		Config:     configRecord(res.config),
	}
	if plan.Provider != nil {
		r.Agent.Provider = plan.Provider.Name
	}
	if plan.Terminal != nil {
		r.Axes.Terminal = string(plan.Terminal.Type)
	}
	for _, binding := range plan.Credentials {
		r.Credentials = append(r.Credentials, binding.Name)
	}
	r.Overlay = overlayRecord(result.overlay)
	if result.started {
		code := result.code
		r.ExitCode = &code
		return r
	}
	r.Status = record.StatusFailed
	r.Error = runErr.Error()
	return r
}

func workspaceRecord(plan Plan) record.Workspace {
	if plan.Checkout == nil {
		return record.Workspace{Mode: record.ModeInPlace}
	}
	return record.Workspace{Mode: record.ModeCheckout, VCS: string(plan.Checkout.VCS), Branch: plan.Checkout.Config.Branch}
}

func guaranteeRecords(caps policy.Capabilities, pol policy.SandboxPolicy) []record.Guarantee {
	var guarantees []record.Guarantee
	for _, outcome := range policy.Evaluate(caps, pol) {
		guarantees = append(guarantees, record.Guarantee{Name: string(outcome.Guarantee), Required: outcome.Required, Provided: outcome.Provided})
	}
	return guarantees
}

func overlayRecord(sources []OverlaySource) record.Overlay {
	var overlay record.Overlay
	for _, source := range sources {
		digest := record.Digest{Name: source.Name, Digest: source.Digest}
		if source.Kind == SourcePlugin {
			overlay.Plugins = append(overlay.Plugins, digest)
		} else {
			overlay.Directories = append(overlay.Directories, digest)
		}
	}
	return overlay
}

func configRecord(resolved runconfig.Resolved) record.Config {
	var out record.Config
	for _, source := range resolved.Sources() {
		out.Sources = append(out.Sources, record.Source{Layer: string(source.Layer), Path: source.Origin})
	}
	out.Decided = map[string][]string{}
	for key, layers := range resolved.Layers() {
		names := make([]string, len(layers))
		for i, layer := range layers {
			names[i] = string(layer)
		}
		sort.Strings(names)
		out.Decided[string(key)] = slices.Compact(names)
	}
	return out
}
