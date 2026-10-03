package cmd

import (
	"errors"
	"testing"
)

type failingCompletionWriter struct{}

func (failingCompletionWriter) Write(_ []byte) (int, error) {
	return 0, errors.New("completion output unavailable")
}

func TestCompletionCommand_PropagatesWriterError(t *testing.T) {
	previous := completionCmd.OutOrStdout()
	completionCmd.SetOut(failingCompletionWriter{})
	t.Cleanup(func() { completionCmd.SetOut(previous) })

	if err := completionCmd.RunE(completionCmd, []string{"bash"}); err == nil {
		t.Fatal("completion command ignored the output error")
	}
}
