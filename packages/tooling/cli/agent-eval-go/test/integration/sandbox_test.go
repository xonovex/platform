//go:build integration

package integration

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// baseImage is the image the harness double is copied into. The test never
// pulls it: a machine without it skips the test.
const baseImage = "busybox:1.37"

// containedClaude answers like the fake claude only inside a container, so a
// pass proves that the harness ran in the sandbox. "sleep" outlives any
// timeout of the test.
const containedClaude = `#!/bin/sh
for arg; do last=$arg; done
if [ ! -e /.dockerenv ]; then echo "not in a container" >&2; exit 1; fi
case "$last" in *sleep*) sleep 120 ;; esac
echo '{"type":"system","subtype":"init","skills":["alpha"]}'
case "$last" in
*"load it"*|*hang*)
  echo '{"type":"assistant","message":{"content":[{"type":"tool_use","name":"Skill","input":{"skill":"alpha"}}]}}'
  case "$last" in *hang*) sleep 120 ;; esac
  ;;
*)
  echo '{"type":"assistant","message":{"content":[{"type":"text","text":"ok"}]}}'
  echo '{"type":"result","result":"ok","total_cost_usd":0.01,"usage":{"input_tokens":5,"output_tokens":2}}'
  ;;
esac
`

func docker(t *testing.T, args ...string) (string, error) {
	t.Helper()
	out, err := exec.Command("docker", args...).CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

// harnessImage builds an image that holds the harness double at /bin/claude
// and removes it when the test ends.
func harnessImage(t *testing.T, w *workspace) string {
	t.Helper()
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("docker is not installed")
	}
	if _, err := docker(t, "info"); err != nil {
		t.Skip("the docker daemon is not reachable")
	}
	if _, err := docker(t, "image", "inspect", baseImage); err != nil {
		t.Skipf("the base image %s is not present", baseImage)
	}
	suffix := make([]byte, 6)
	if _, err := rand.Read(suffix); err != nil {
		t.Fatal(err)
	}
	tag := "agent-eval-integration:" + hex.EncodeToString(suffix)
	builder := "agent-eval-integration-" + hex.EncodeToString(suffix)
	w.write(t, "image/claude", containedClaude)
	script := filepath.Join(w.root, "image", "claude")
	if err := os.Chmod(script, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{
		{"create", "--name", builder, baseImage},
		{"cp", script, builder + ":/bin/claude"},
		{"commit", builder, tag},
		{"rm", builder},
	} {
		if out, err := docker(t, args...); err != nil {
			_, _ = docker(t, "rm", "--force", builder)
			t.Fatalf("docker %v: %v\n%s", args, err, out)
		}
	}
	t.Cleanup(func() { _, _ = docker(t, "image", "rm", "--force", tag) })
	return tag
}

// sandboxes lists the containers whose name agent-eval gives.
func sandboxes(t *testing.T) []string {
	t.Helper()
	out, err := docker(t, "ps", "--all", "--filter", "name=agent-eval-", "--format", "{{.Names}}")
	if err != nil {
		t.Fatal(err)
	}
	return strings.Fields(out)
}

// approve writes the project file and approves it.
func (w *workspace) approve(t *testing.T, project string) {
	t.Helper()
	w.write(t, "work/agentic.json", project)
	sum := sha256.Sum256([]byte(project))
	digest := "sha256:" + hex.EncodeToString(sum[:])
	w.write(t, "state/agentic/trust.json", `{"version":1,"approvals":[{"digest":"`+digest+`","path":"`+filepath.Join(w.root, "work", "agentic.json")+`"}]}`)
}

func TestADockerSandboxRunsTheHarnessAndLeavesNoContainer(t *testing.T) {
	w := newWorkspace(t)
	image := harnessImage(t, w)
	before := sandboxes(t)

	// The fake claude on the host PATH would also pass, so the host copy is
	// removed: only the copy in the image can answer.
	if err := os.Remove(filepath.Join(w.bin, "claude")); err != nil {
		t.Fatal(err)
	}
	w.write(t, "config/agentic/agentic.json", `{"version":1,"eval":{"sandbox":{"isolation":"docker","image":"`+image+`","network":"none"}}}`)
	got := w.run(t, "trigger", "alpha", "--runs", "1")
	if got.code != 0 || !strings.Contains(got.stdout, "PASS") {
		t.Fatalf("the harness in the container passes: %d\n%s\n%s", got.code, got.stdout, got.stderr)
	}

	// A harness that keeps the stream open is stopped by the trigger reader, and
	// one that never answers is ended by the timeout. Each leaves its container
	// running until agent-eval removes it.
	w.write(t, "work/plugins/core/skills/alpha/eval-trigger.json", `{"version":1,"skill":"alpha","queries":[
 {"id":"hang","prompt":"hang","shouldTrigger":true,"split":"validation"}]}`)
	if got := w.run(t, "trigger", "alpha", "--runs", "1", "--sandbox-image", image); got.code != 0 {
		t.Fatalf("a stopped run passes: %d\n%s\n%s", got.code, got.stdout, got.stderr)
	}
	w.write(t, "work/plugins/core/skills/alpha/eval-trigger.json", `{"version":1,"skill":"alpha","queries":[
 {"id":"sleep","prompt":"sleep","shouldTrigger":true,"split":"validation"}]}`)
	w.approve(t, `{"version":1,"eval":{"pluginDirs":["plugins"],"profiles":{"ci":{"harness":"claude","model":"fake","timeoutSeconds":1}},"defaultProfile":"ci"}}`)
	if got := w.run(t, "trigger", "alpha", "--runs", "1"); got.code != 2 || !strings.Contains(got.stdout, "errored  sleep") {
		t.Fatalf("a run that times out on every attempt is a fault: %d\n%s\n%s", got.code, got.stdout, got.stderr)
	}

	for _, name := range sandboxes(t) {
		if !slices.Contains(before, name) {
			_, _ = docker(t, "rm", "--force", name)
			t.Errorf("container %s outlived its run", name)
		}
	}
}
