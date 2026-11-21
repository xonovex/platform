package main

import (
	"io"
	"os"
)

// osFiles is the summaryFiles adapter over the real disk.
type osFiles struct{}

func (osFiles) ReadFile(name string) ([]byte, error) {
	return os.ReadFile(name)
}

func (osFiles) OpenAppend(name string) (io.WriteCloser, error) {
	return os.OpenFile(name, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
}
