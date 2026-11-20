package configtest_test

import (
	"errors"
	"io"
	iofs "io/fs"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/configtest"
)

func TestFiles(t *testing.T) {
	files := configtest.NewFiles(map[string]string{"/a/b/c.json": "x", "/empty": ""})
	if data, err := files.ReadFile("/a/b/c.json"); err != nil || string(data) != "x" {
		t.Fatalf("ReadFile = %q, %v", data, err)
	}
	if _, err := files.ReadFile("/a/b"); !errors.Is(err, iofs.ErrNotExist) {
		t.Fatalf("ReadFile of a directory = %v", err)
	}
	if info, err := files.Stat("/a/b/c.json"); err != nil || !info.Mode.IsRegular() {
		t.Fatalf("Stat file = %v, %v", info, err)
	}
	if info, err := files.Stat("/empty"); err != nil || !info.Mode.IsRegular() {
		t.Fatalf("Stat empty file = %v, %v", info, err)
	}
	if info, err := files.Stat("/a"); err != nil || !info.Mode.IsDir() {
		t.Fatalf("Stat dir = %v, %v", info, err)
	}
	if _, err := files.Stat("/z"); !errors.Is(err, iofs.ErrNotExist) {
		t.Fatalf("Stat missing = %v", err)
	}
	if err := files.WriteFile("/s/t.json", []byte("y")); err != nil {
		t.Fatal(err)
	}
	if data, _ := files.ReadFile("/s/t.json"); string(data) != "y" {
		t.Fatalf("written = %q", data)
	}
	if err := files.WriteFile("/a", []byte("y")); err == nil {
		t.Fatal("writing over a directory must fail")
	}
}

func TestPlaces(t *testing.T) {
	errBoom := errors.New("boom")
	places := configtest.Places{Managed: "/m", User: "/u", State: "/s", Wd: "/w", Err: errBoom}
	if places.ManagedDir() != "/m" {
		t.Fatal("ManagedDir")
	}
	if dir, err := places.UserDir(); dir != "/u" || err != errBoom {
		t.Fatal("UserDir")
	}
	if dir, err := places.StateDir(); dir != "/s" || err != errBoom {
		t.Fatal("StateDir")
	}
	if dir, err := places.Getwd(); dir != "/w" || err != errBoom {
		t.Fatal("Getwd")
	}
}

func TestPrompter(t *testing.T) {
	prompter := &configtest.Prompter{Terminal: true, Answers: []string{"y"}}
	if !prompter.IsTerminal() {
		t.Fatal("IsTerminal")
	}
	if answer, err := prompter.ReadLine(); answer != "y" || err != nil {
		t.Fatalf("ReadLine = %q, %v", answer, err)
	}
	if _, err := prompter.ReadLine(); err != io.EOF {
		t.Fatalf("ReadLine after the answers = %v", err)
	}
}
