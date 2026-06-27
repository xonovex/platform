package cmd

import (
	"testing"

	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/validation"
)

func TestParseNetwork(t *testing.T) {
	if _, err := parseNetwork("proxy"); err != nil {
		t.Errorf("parseNetwork(proxy) = %v", err)
	}
	if _, err := parseNetwork("bogus"); err == nil {
		t.Error("parseNetwork(bogus) = nil, want error")
	}
}

func TestWorktreeBranchValidation(t *testing.T) {
	tests := []struct {
		name    string
		branch  string
		wantErr bool
	}{
		{"valid simple", "feature/my-work", false},
		{"valid main", "main", false},
		{"valid release", "release-1.0", false},
		{"invalid semicolon", "branch;rm -rf /", true},
		{"invalid pipe", "branch|evil", true},
		{"invalid dollar", "branch$(whoami)", true},
		{"invalid backtick", "branch`id`", true},
		{"invalid spaces", "branch name", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validation.ValidateBranch(tt.branch)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateBranch(%q) error = %v, wantErr %v", tt.branch, err, tt.wantErr)
			}
		})
	}
}
