package builder

import "fmt"

// JujutsuStrategy implements VCSStrategy for Jujutsu (jj)
type JujutsuStrategy struct{}

func (j *JujutsuStrategy) PostCloneScript() string {
	return "jj git init --colocate\n"
}

func (j *JujutsuStrategy) WorktreeScript(path, _, sourceBranch string) string {
	return fmt.Sprintf("jj workspace add %s --revision %s\n", shellQuote(path), shellQuote(sourceBranch))
}

func (j *JujutsuStrategy) InitContainerName() string {
	return "jj-workspace"
}
