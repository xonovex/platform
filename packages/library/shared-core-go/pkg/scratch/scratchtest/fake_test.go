package scratchtest_test

import (
	"errors"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/scratch/scratchtest"
)

func TestMakeNumbersDirectoriesAndRemoveRecords(t *testing.T) {
	f := scratchtest.New()
	a, _ := f.Make("p-")
	b, _ := f.Make("p-")
	if a != "/scratch/p-1" || b != "/scratch/p-2" {
		t.Fatal(a, b)
	}
	_ = f.Remove(a)
	if len(f.Made()) != 2 || len(f.Removed()) != 1 || f.Removed()[0] != a {
		t.Fatal(f.Made(), f.Removed())
	}
}

func TestMakeFailure(t *testing.T) {
	f := scratchtest.New()
	f.MakeErr = errors.New("disk full")
	if _, err := f.Make("x"); err == nil {
		t.Fatal("want error")
	}
}
