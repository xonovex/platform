package workspace

import "testing"

func TestVCSTypeIsValid(t *testing.T) {
	tests := []struct {
		name string
		vcs  VCSType
		want bool
	}{
		{name: "git", vcs: VCSGit, want: true},
		{name: "jujutsu", vcs: VCSJujutsu, want: true},
		{name: "unknown", vcs: VCSType("unknown"), want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := test.vcs.IsValid(); got != test.want {
				t.Fatalf("IsValid() = %t, want %t", got, test.want)
			}
		})
	}
}
