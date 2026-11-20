package directory

import (
	"errors"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-launch-go/pkg/oshost/oshosttest"
)

func TestAcceptsAndFetchesAnExistingDirectory(t *testing.T) {
	host := oshosttest.New().AddDir("/abs/plugin").AddDir("/work/rel/plugin")
	origin := New(host)
	for spec, want := range map[string]string{"/abs/plugin": "/abs/plugin", "/abs/plugin/": "/abs/plugin", "rel/plugin": "/work/rel/plugin"} {
		if !origin.Accepts(spec) {
			t.Errorf("Accepts(%q) = false", spec)
		}
		checkout, err := origin.Fetch(spec)
		if err != nil || checkout.Dir != want || checkout.Source != want || checkout.Resolved != "" {
			t.Errorf("Fetch(%q) = %+v, %v", spec, checkout, err)
		}
		checkout.Release()
	}
}

func TestRefusesWhatIsNotADirectory(t *testing.T) {
	host := oshosttest.New().AddFile("/a/file", "x", 0o644)
	origin := New(host)
	for _, spec := range []string{"", "/missing", "/a/file", "https://example.test/x.git"} {
		if origin.Accepts(spec) {
			t.Errorf("Accepts(%q) = true", spec)
		}
		if _, err := origin.Fetch(spec); err == nil {
			t.Errorf("Fetch(%q) passed", spec)
		}
	}
}

type brokenCwd struct{ *oshosttest.Fake }

func (brokenCwd) Getwd() (string, error) { return "", errors.New("no cwd") }

func TestAFailingWorkingDirectoryRefusesARelativeSource(t *testing.T) {
	origin := New(brokenCwd{oshosttest.New()})
	if origin.Accepts("rel") {
		t.Error("Accepts passed without a working directory")
	}
	if _, err := origin.Fetch("rel"); err == nil {
		t.Error("Fetch passed without a working directory")
	}
	if _, err := origin.Fetch("/abs"); err == nil {
		t.Error("Fetch passed for a missing absolute directory")
	}
}
