package launch

import (
	"fmt"
	"sort"

	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/logging"
	termshared "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/terminal/shared"
)

// resolveAutoTerminal replaces the automatic terminal selection with the one
// wrapper among AutoTerminals whose session the CLI runs in. Any other
// configuration passes unchanged.
func resolveAutoTerminal(deps Dependencies, cfg *termshared.TerminalConfig) (*termshared.TerminalConfig, error) {
	if cfg == nil || cfg.Type != termshared.TerminalAuto {
		return cfg, nil
	}
	var inside []termshared.TerminalType
	for _, name := range deps.AutoTerminals {
		if executor, ok := deps.Terminals[name]; ok && executor.IsInside() {
			inside = append(inside, name)
		}
	}
	name, err := termshared.SelectAuto(inside)
	if err != nil {
		return nil, err
	}
	resolved := *cfg
	resolved.Type = name
	return &resolved, nil
}

// withRunID returns a copy of the terminal configuration that carries the id
// of the run record, so a wrapper can label its session with it.
func withRunID(cfg *termshared.TerminalConfig, runID string) *termshared.TerminalConfig {
	if cfg == nil {
		return nil
	}
	labelled := *cfg
	labelled.RunID = runID
	return &labelled
}

// selectTerminal returns the installed executor of the requested terminal, or
// nil when the run uses no terminal wrapper.
func selectTerminal(terminals map[termshared.TerminalType]termshared.TerminalExecutor, cfg *termshared.TerminalConfig) (termshared.TerminalExecutor, error) {
	if cfg == nil {
		return nil, nil
	}
	executor, ok := terminals[cfg.Type]
	if !ok {
		return nil, fmt.Errorf("unknown terminal type: %s; available types: %v", cfg.Type, availableTerminals(terminals))
	}
	if !executor.IsAvailable() {
		return nil, fmt.Errorf("terminal type %s is not available (not installed)", cfg.Type)
	}
	return executor, nil
}

func availableTerminals(terminals map[termshared.TerminalType]termshared.TerminalExecutor) []string {
	var names []string
	for name, executor := range terminals {
		if executor.IsAvailable() {
			names = append(names, string(name))
		}
	}
	sort.Strings(names)
	return names
}

// startInTerminal starts the confined agent inside the terminal wrapper of the
// plan. The code of a wrapper that fails still counts, and a program that runs
// on in its session after the wrapper returns is reported, so its overlay stays.
func startInTerminal(log logging.Logger, plan Plan, terminal termshared.TerminalExecutor, sandboxed confined) (launched, error) {
	command, env, err := plan.Isolator.TerminalCommand(sandboxed.run, sandboxed.contribution)
	if err != nil {
		return launched{}, fmt.Errorf("prepare terminal command: %w", err)
	}
	if plan.Verbose {
		log.Info("Using terminal wrapper: " + string(plan.Terminal.Type))
	}
	ran, err := terminal.Execute(plan.Terminal, command, env, sandboxed.run.Process.Dir)
	result := launched{code: ran.Code, runsOn: !ran.Ended}
	if err != nil {
		return result, err
	}
	return exited(result)
}
