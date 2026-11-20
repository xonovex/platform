package plugin

import (
	"strings"
	"testing"
)

func record(name string, dependencies ...string) Locked {
	return Locked{Name: name, Version: "1.0.0", Source: "/src/" + name, Digest: "sha256:" + strings.Repeat("a", 64), Dependencies: dependencies}
}

func TestLockWithAndWithoutKeepOrder(t *testing.T) {
	lock := EmptyLock().With(record("b")).With(record("a")).With(record("c"))
	replaced := lock.With(Locked{Name: "b", Version: "2.0.0", Source: "/new", Digest: record("b").Digest})
	if len(replaced.Plugins) != 3 || replaced.Plugins[1].Version != "2.0.0" || replaced.Plugins[0].Name != "a" || len(lock.Plugins) != 3 || lock.Plugins[1].Version != "1.0.0" {
		t.Errorf("replaced = %+v, lock = %+v", replaced, lock)
	}
	without, found := lock.Without("b")
	if !found || len(without.Plugins) != 2 || len(lock.Plugins) != 3 {
		t.Errorf("without = %+v, found = %v", without, found)
	}
	if same, found := lock.Without("zzz"); found || len(same.Plugins) != 3 {
		t.Errorf("Without(zzz) = %+v, %v", same, found)
	}
	if _, found := lock.Find("a"); !found {
		t.Error("Find missed a")
	}
	if _, found := lock.Find("zzz"); found {
		t.Error("Find found zzz")
	}
}

func TestLockDependents(t *testing.T) {
	lock := EmptyLock().With(record("core")).With(record("integrations", "core")).With(record("other"))
	if got := lock.Dependents("core"); len(got) != 1 || got[0] != "integrations" {
		t.Errorf("Dependents(core) = %v", got)
	}
	if got := lock.Dependents("other"); got != nil {
		t.Errorf("Dependents(other) = %v", got)
	}
}

func TestLockClosureOrdersDependenciesFirst(t *testing.T) {
	lock := EmptyLock().With(record("core")).With(record("mid", "core")).With(record("top", "mid", "core"))
	closure, err := lock.Closure([]string{"top", "core"})
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, entry := range closure {
		names = append(names, entry.Name)
	}
	if strings.Join(names, ",") != "core,mid,top" {
		t.Errorf("closure = %v", names)
	}
}

func TestLockClosureFailsClosed(t *testing.T) {
	lock := EmptyLock().With(record("a", "missing")).With(record("x", "y")).With(record("y", "x"))
	if _, err := lock.Closure([]string{"nope"}); err == nil || !strings.Contains(err.Error(), "not installed") {
		t.Errorf("a missing plugin: %v", err)
	}
	if _, err := lock.Closure([]string{"a"}); err == nil || !strings.Contains(err.Error(), `depends on "missing"`) {
		t.Errorf("a missing dependency: %v", err)
	}
	if _, err := lock.Closure([]string{"x"}); err == nil || !strings.Contains(err.Error(), "depends on itself") {
		t.Errorf("a cycle: %v", err)
	}
}

func TestLockRoundTrips(t *testing.T) {
	lock := EmptyLock().With(Locked{Name: "a", Version: "1", Source: "/x", Marketplace: "mk", Resolved: "abc", Digest: record("a").Digest, Grants: []string{"hooks"}, Dependencies: []string{"b"}}).With(record("b"))
	data, err := EncodeLock(lock)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(string(data), "}\n") {
		t.Errorf("the file must end with a newline: %q", data)
	}
	back, err := DecodeLock(data)
	if err != nil {
		t.Fatal(err)
	}
	again, _ := EncodeLock(back)
	if string(again) != string(data) {
		t.Errorf("the round trip changed the file:\n%s\n%s", data, again)
	}
	empty, err := EncodeLock(EmptyLock())
	if err != nil || !strings.Contains(string(empty), `"plugins": []`) {
		t.Errorf("an empty lock must list an empty array: %s, %v", empty, err)
	}
}

func TestDecodeLockFailsClosed(t *testing.T) {
	digest := record("a").Digest
	for name, data := range map[string]string{
		"not json":      `{`,
		"unknown key":   `{"version":1,"plugins":[],"extra":1}`,
		"version":       `{"version":2,"plugins":[]}`,
		"trailing":      `{"version":1,"plugins":[]} {}`,
		"bad name":      `{"version":1,"plugins":[{"name":"Bad","source":"/x","digest":"` + digest + `"}]}`,
		"twice":         `{"version":1,"plugins":[{"name":"a","source":"/x","digest":"` + digest + `"},{"name":"a","source":"/x","digest":"` + digest + `"}]}`,
		"bad digest":    `{"version":1,"plugins":[{"name":"a","source":"/x","digest":"nope"}]}`,
		"no source":     `{"version":1,"plugins":[{"name":"a","digest":"` + digest + `"}]}`,
		"bad grant":     `{"version":1,"plugins":[{"name":"a","source":"/x","digest":"` + digest + `","grants":["shell"]}]}`,
		"bad dependecy": `{"version":1,"plugins":[{"name":"a","source":"/x","digest":"` + digest + `","dependencies":["Bad"]}]}`,
	} {
		if _, err := DecodeLock([]byte(data)); err == nil {
			t.Errorf("%s: DecodeLock passed", name)
		}
	}
	if _, err := EncodeLock(Lock{Version: 9}); err == nil {
		t.Error("EncodeLock wrote an unsupported version")
	}
}

func TestRegistryWithAndWithout(t *testing.T) {
	registry := EmptyRegistry().With(Source{Name: "b", Source: "/b"}).With(Source{Name: "a", Source: "/a"})
	registry = registry.With(Source{Name: "b", Source: "/b2"})
	if len(registry.Marketplaces) != 2 || registry.Marketplaces[0].Name != "a" || registry.Marketplaces[1].Source != "/b2" {
		t.Errorf("registry = %+v", registry)
	}
	if source, ok := registry.Find("a"); !ok || source.Source != "/a" {
		t.Errorf("Find(a) = %+v, %v", source, ok)
	}
	if _, ok := registry.Find("zzz"); ok {
		t.Error("Find found zzz")
	}
	without, found := registry.Without("a")
	if !found || len(without.Marketplaces) != 1 || len(registry.Marketplaces) != 2 {
		t.Errorf("without = %+v, %v", without, found)
	}
	if same, found := registry.Without("zzz"); found || len(same.Marketplaces) != 2 {
		t.Errorf("Without(zzz) = %+v, %v", same, found)
	}
}

func TestRegistryRoundTripsAndFailsClosed(t *testing.T) {
	data, err := EncodeRegistry(EmptyRegistry().With(Source{Name: "mk", Source: "/x"}))
	if err != nil {
		t.Fatal(err)
	}
	back, err := DecodeRegistry(data)
	if err != nil || len(back.Marketplaces) != 1 {
		t.Fatalf("back = %+v, %v", back, err)
	}
	empty, err := EncodeRegistry(EmptyRegistry())
	if err != nil || !strings.Contains(string(empty), `"marketplaces": []`) {
		t.Errorf("empty = %s, %v", empty, err)
	}
	for name, data := range map[string]string{
		"not json": `{`, "unknown": `{"version":1,"marketplaces":[],"x":1}`, "version": `{"version":3,"marketplaces":[]}`,
		"trailing": `{"version":1,"marketplaces":[]} 1`,
		"name":     `{"version":1,"marketplaces":[{"name":"Bad","source":"/x"}]}`,
		"twice":    `{"version":1,"marketplaces":[{"name":"a","source":"/x"},{"name":"a","source":"/y"}]}`,
		"source":   `{"version":1,"marketplaces":[{"name":"a","source":""}]}`,
	} {
		if _, err := DecodeRegistry([]byte(data)); err == nil {
			t.Errorf("%s: DecodeRegistry passed", name)
		}
	}
	if _, err := EncodeRegistry(Registry{Version: 0}); err == nil {
		t.Error("EncodeRegistry wrote an unsupported version")
	}
}
