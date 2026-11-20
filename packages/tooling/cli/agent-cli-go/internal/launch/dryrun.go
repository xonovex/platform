package launch

import (
	"fmt"
	"io"
	"strings"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/network"
	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/logging"
	credshared "github.com/xonovex/platform/packages/library/shared-launch-go/pkg/credential"
)

// printDryRun reports the run and writes the sandbox command line to out.
func printDryRun(log logging.Logger, out io.Writer, plan Plan, overlay OverlayResult, provenance, command []string) error {
	log.Info("Dry run - would execute:")
	log.Info("  Agent: " + plan.Agent.DisplayName)
	log.Info("  Protocol: " + describeProtocol(plan.Agent))
	log.Info("  Model access: " + describeAccess(plan.Agent))
	log.Info("  Version probe: " + describeVersionProbe(plan.Agent))
	if plan.Provider != nil {
		log.Info("  Provider: " + plan.Provider.DisplayName)
		for _, slot := range plan.Provider.Slots {
			log.Info("  Model " + describeSlot(slot))
		}
	}
	log.Info("  Work directory: " + plan.DisplayDir())
	if len(plan.AgentArgs) > 0 {
		log.Info(fmt.Sprintf("  Arguments: %v", plan.AgentArgs))
	} else {
		log.Info("  Arguments: (none)")
	}
	logOverlay(log, overlay)
	logNetwork(log, plan)
	logProvenance(log, provenance)
	if len(command) == 0 {
		return nil
	}
	if _, err := fmt.Fprintln(out, strings.Join(command, " ")); err != nil {
		return fmt.Errorf("write dry run: %w", err)
	}
	return nil
}

// logOverlay reports the overlay of a dry run.
func logOverlay(log logging.Logger, result OverlayResult) {
	if len(result.Summary) == 0 {
		return
	}
	log.Info("  Overlay:")
	for _, line := range result.Summary {
		log.Info("    " + strings.TrimSpace(line))
	}
}

// logProvenance reports, one line per configuration value, the layer that decided
// it.
func logProvenance(log logging.Logger, provenance []string) {
	if len(provenance) == 0 {
		return
	}
	log.Info("  Configuration:")
	for _, line := range provenance {
		log.Info("    " + line)
	}
}

// logNetwork reports the network mode, the allowed hosts and the credential
// bindings of a dry run. A binding shows where the credential lives and never what
// it holds.
func logNetwork(log logging.Logger, plan Plan) {
	log.Info("  Network: " + string(plan.Run.Network))
	if plan.Run.Network == network.NetworkManaged {
		var hosts []string
		for _, rule := range plan.Allow.Rules() {
			hosts = append(hosts, rule.String())
		}
		log.Info("  Allowed hosts: " + strings.Join(hosts, ", "))
		log.Info("  Egress transport: " + string(plan.Run.Egress.Transport))
	}
	for _, binding := range plan.Credentials {
		log.Info("  Credential: " + binding.Name + " from " + binding.Scheme + ":" + binding.Locator + " (" + credshared.Redacted + ")")
	}
}
