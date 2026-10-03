package jj

import "testing"

func TestStrategyBuildsQuotedWorkspaceScript(t *testing.T) {
	strategy := &Strategy{}

	script := strategy.WorktreeScript("/workspace/my tree", "ignored", "change@origin")

	want := "jj workspace add '/workspace/my tree' --revision 'change@origin'\n"
	if script != want {
		t.Errorf("WorktreeScript() = %q, want %q", script, want)
	}
}

func TestStrategyIdentifiesJujutsuWorkspaceContainer(t *testing.T) {
	strategy := &Strategy{}

	name := strategy.InitContainerName()

	if name != "jj-workspace" {
		t.Errorf("InitContainerName() = %q, want jj-workspace", name)
	}
}

func TestStrategyInitializesColocatedRepository(t *testing.T) {
	strategy := &Strategy{}

	script := strategy.PostCloneScript()

	if script != "jj git init --colocate\n" {
		t.Errorf("PostCloneScript() = %q, want colocated init", script)
	}
}
