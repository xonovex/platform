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
	completion := newCompletionCommand()
	completion.SetOut(failingCompletionWriter{})

	if err := completion.RunE(completion, []string{"bash"}); err == nil {
		t.Fatal("completion command ignored the output error")
	}
}
