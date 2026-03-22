package worktree

// VCSType represents the type of version control system
type VCSType string

const (
	VCSGit     VCSType = "git"
	VCSJujutsu VCSType = "jj"
	VCSDefault VCSType = VCSGit
)

// IsValid returns true if vt is a recognised VCS type
func (vt VCSType) IsValid() bool {
	return vt == VCSGit || vt == VCSJujutsu
}
