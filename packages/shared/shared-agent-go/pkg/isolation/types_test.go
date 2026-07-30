package isolation

import (
	"slices"
	"testing"

	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/types"
)

func TestUserConfigPathsAreScopedToAgent(t *testing.T) {
	tests := []struct {
		name      string
		agentType types.AgentType
		want      []string
	}{
		{name: "claude", agentType: types.AgentClaude, want: []string{".claude", ".claude.json"}},
		{name: "opencode", agentType: types.AgentOpencode, want: []string{".config/opencode"}},
		{name: "unknown", agentType: "unknown", want: nil},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			paths := UserConfigPaths(test.agentType)

			if !slices.Equal(paths, test.want) {
				t.Fatalf("UserConfigPaths() = %v, want %v", paths, test.want)
			}
		})
	}
}
