package builder

import agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"

// OpencodeCommandBuilder builds command/args for Opencode
type OpencodeCommandBuilder struct{}

func (o *OpencodeCommandBuilder) Command(run *agentv1alpha1.AgentRun) ([]string, []string) {
	var args []string
	if run.Spec.Provider != nil && len(run.Spec.Provider.CliArgs) > 0 {
		args = append(args, run.Spec.Provider.CliArgs...)
	}
	return []string{"opencode"}, args
}
