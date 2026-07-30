package git

import "testing"

func TestStrategyBuildsQuotedWorktreeScript(t *testing.T) {
	strategy := &Strategy{}

	script := strategy.WorktreeScript("/workspace/my tree", "feature/test", "main")

	want := "git worktree add '/workspace/my tree' -b 'feature/test' 'main'\n"
	if script != want {
		t.Errorf("WorktreeScript() = %q, want %q", script, want)
	}
}

func TestStrategyIdentifiesGitWorktreeContainer(t *testing.T) {
	strategy := &Strategy{}

	name := strategy.InitContainerName()

	if name != "git-worktree" {
		t.Errorf("InitContainerName() = %q, want git-worktree", name)
	}
}

func TestStrategyRequiresNoPostCloneScript(t *testing.T) {
	strategy := &Strategy{}

	script := strategy.PostCloneScript()

	if script != "" {
		t.Errorf("PostCloneScript() = %q, want empty script", script)
	}
}
