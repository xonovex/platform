//go:build integration
// +build integration

package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// pluginEnvironment runs the binary with a private home and store.
type pluginEnvironment struct {
	t     *testing.T
	home  string
	store string
}

func newPluginEnvironment(t *testing.T) pluginEnvironment {
	t.Helper()
	home := t.TempDir()
	return pluginEnvironment{t: t, home: home, store: filepath.Join(home, "store")}
}

func (e pluginEnvironment) run(args ...string) (string, error) {
	e.t.Helper()
	cmd := exec.Command(testBinary(e.t), args...)
	cmd.Env = append(os.Environ(), "HOME="+e.home, "XDG_CONFIG_HOME="+e.home, "XDG_DATA_HOME="+filepath.Join(e.home, "data"))
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func writeFile(t *testing.T, path, content string, mode os.FileMode) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), mode); err != nil {
		t.Fatal(err)
	}
}

func makePlugin(t *testing.T) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "hello")
	writeFile(t, filepath.Join(dir, ".claude-plugin", "plugin.json"), `{"name":"hello","version":"1.0.0"}`, 0o644)
	writeFile(t, filepath.Join(dir, "skills", "greet", "SKILL.md"), "---\nname: greet\ndescription: Use when greeting.\n---\n# Greet\n", 0o644)
	writeFile(t, filepath.Join(dir, "skills", "greet", "run.sh"), "#!/bin/sh\n", 0o755)
	return dir
}

func digestOf(t *testing.T, output string) string {
	t.Helper()
	_, after, found := strings.Cut(output, "--approve-config-digest ")
	if !found {
		t.Fatalf("no digest in %q", output)
	}
	return strings.Fields(after)[0]
}

func TestPluginLifecycleOnADisk(t *testing.T) {
	env := newPluginEnvironment(t)
	plugin := makePlugin(t)

	output, err := env.run("plugin", "install", plugin, "--store", env.store)
	if err == nil {
		t.Fatalf("an install without approval passed: %s", output)
	}
	digest := digestOf(t, output)
	if _, err := os.Stat(filepath.Join(env.store, "lock.json")); err == nil {
		t.Error("the preview installed the plugin")
	}

	if output, err := env.run("plugin", "install", plugin, "--store", env.store, "--approve-config-digest", digest); err != nil {
		t.Fatalf("install: %v\n%s", err, output)
	}
	if output, err := env.run("plugin", "list", "--store", env.store); err != nil || !strings.Contains(output, "hello 1.0.0 "+digest) {
		t.Fatalf("list: %v\n%s", err, output)
	}
	info, err := os.Stat(filepath.Join(env.store, "content", strings.TrimPrefix(digest, "sha256:"), "skills", "greet", "run.sh"))
	if err != nil || info.Mode().Perm()&0o111 == 0 {
		t.Errorf("the stored script lost its executable bit: %v", err)
	}

	output, err = env.run("run", "--dry-run", "--agent", "claude", "--plugin", "hello", "--store", env.store, "--work-dir", t.TempDir())
	if err != nil {
		t.Fatalf("dry run: %v\n%s", err, output)
	}
	for _, want := range []string{"Overlay:", "would place", "--add-dir"} {
		if !strings.Contains(output, want) {
			t.Errorf("dry run lacks %q:\n%s", want, output)
		}
	}

	// A stored file that changed on disk stops the run.
	stored := filepath.Join(env.store, "content", strings.TrimPrefix(digest, "sha256:"), "skills", "greet", "SKILL.md")
	writeFile(t, stored, "tampered", 0o644)
	if output, err := env.run("run", "--dry-run", "--plugin", "hello", "--store", env.store); err == nil || !strings.Contains(output, "changed on disk") {
		t.Errorf("a tampered plugin ran: %v\n%s", err, output)
	}

	if output, err := env.run("plugin", "remove", "hello", "--store", env.store); err != nil {
		t.Fatalf("remove: %v\n%s", err, output)
	}
	if _, err := os.Stat(filepath.Join(env.store, "content", strings.TrimPrefix(digest, "sha256:"))); err == nil {
		t.Error("remove left the stored files")
	}
}

func TestPluginFromAGitRepositoryAndAMarketplace(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}
	env := newPluginEnvironment(t)
	repo := t.TempDir()
	writeFile(t, filepath.Join(repo, ".claude-plugin", "marketplace.json"), `{"name":"main","plugins":[{"name":"hello","source":"./plugins/hello"}]}`, 0o644)
	writeFile(t, filepath.Join(repo, "plugins", "hello", "plugin.json"), `{"name":"hello","version":"1.0.0"}`, 0o644)
	writeFile(t, filepath.Join(repo, "plugins", "hello", "agents", "reviewer.md"), "---\ndescription: Reviews.\n---\nBe strict.\n", 0o644)
	for _, args := range [][]string{{"init", "-q", "-b", "main"}, {"add", "."}, {"-c", "user.name=t", "-c", "user.email=t@example.test", "commit", "-q", "-m", "init"}} {
		cmd := exec.Command("git", args...)
		cmd.Dir = repo
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, output)
		}
	}
	url := "file://" + repo + "#main"

	if output, err := env.run("marketplace", "add", url, "--store", env.store); err != nil || !strings.Contains(output, "Registered marketplace main") {
		t.Fatalf("marketplace add: %v\n%s", err, output)
	}
	output, err := env.run("plugin", "install", "hello@main", "--store", env.store)
	if err == nil || !strings.Contains(output, "Revision: ") {
		t.Fatalf("preview: %v\n%s", err, output)
	}
	digest := digestOf(t, output)
	if output, err := env.run("plugin", "install", "hello@main", "--store", env.store, "--approve-config-digest", digest); err != nil {
		t.Fatalf("install: %v\n%s", err, output)
	}
	if output, err := env.run("marketplace", "remove", "main", "--store", env.store); err == nil {
		t.Errorf("a marketplace in use was removed: %s", output)
	}
}

func TestPluginInstallDoesNotTreatPipedInputAsApproval(t *testing.T) {
	env := newPluginEnvironment(t)
	plugin := makePlugin(t)

	cmd := exec.Command(testBinary(t), "plugin", "install", plugin, "--store", env.store)
	cmd.Env = append(os.Environ(), "HOME="+env.home, "XDG_CONFIG_HOME="+env.home, "XDG_DATA_HOME="+filepath.Join(env.home, "data"))
	cmd.Stdin = strings.NewReader("yes\n")
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("a piped yes installed a plugin: %s", output)
	}
	if strings.Contains(string(output), "Approve this digest?") || !strings.Contains(string(output), "--approve-config-digest ") {
		t.Errorf("output = %q, want the preview and the --approve-config-digest instructions only", output)
	}
	if _, err := os.Stat(filepath.Join(env.store, "lock.json")); err == nil {
		t.Error("a piped yes changed the store")
	}
}
