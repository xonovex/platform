package run_test

import (
	"errors"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/fs"
)

type failingFS struct{}

func (failingFS) ReadFile(string) ([]byte, error)    { return nil, fs.ErrNotExist }
func (failingFS) WriteFile(string, []byte) error     { return errors.New("disk full") }
func (failingFS) ReadDir(string) ([]fs.Entry, error) { return nil, fs.ErrNotExist }
func (failingFS) Stat(string) (fs.Info, error)       { return fs.Info{}, fs.ErrNotExist }
func (failingFS) MkdirAll(string) error              { return nil }
func (failingFS) Remove(string) error                { return nil }
