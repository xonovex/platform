package agents

import (
	"fmt"
	"slices"
	"strings"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/types"
)

// HeadlessArgs returns the arguments of a headless run that writes the output
// format: the non-interactive arguments, the arguments of the format, then the
// hermetic arguments.
func HeadlessArgs(agent *types.AgentConfig, format string) ([]string, error) {
	output, ok := agent.Headless.Output[format]
	if !ok {
		return nil, fmt.Errorf("harness %s declares no %s output", agent.Type, format)
	}
	args := slices.Clone(agent.Headless.Args)
	args = append(args, output...)
	return append(args, agent.Headless.HermeticArgs...), nil
}

// Option returns the arguments of a named option with value in place of the
// placeholder.
func Option(agent *types.AgentConfig, name, value string) ([]string, error) {
	template, ok := agent.Headless.Options[name]
	if !ok {
		return nil, fmt.Errorf("harness %s declares no %s option", agent.Type, name)
	}
	args := make([]string, len(template))
	for i, arg := range template {
		args[i] = strings.ReplaceAll(arg, types.OptionValue, value)
	}
	return args, nil
}

// Setting is one named option of a headless run and its value.
type Setting struct {
	Name, Value string
}

// Options returns the arguments of the settings, in order.
func Options(agent *types.AgentConfig, settings ...Setting) ([]string, error) {
	var args []string
	for _, setting := range settings {
		option, err := Option(agent, setting.Name, setting.Value)
		if err != nil {
			return nil, err
		}
		args = append(args, option...)
	}
	return args, nil
}

// SkillPrompt returns the prompt that invokes skill with the request.
func SkillPrompt(agent *types.AgentConfig, skill, prompt string) (string, error) {
	if agent.Headless.SkillPrompt == "" {
		return "", fmt.Errorf("harness %s declares no skill prompt", agent.Type)
	}
	return strings.NewReplacer(types.SkillPlaceholder, skill, types.PromptPlaceholder, prompt).Replace(agent.Headless.SkillPrompt), nil
}
