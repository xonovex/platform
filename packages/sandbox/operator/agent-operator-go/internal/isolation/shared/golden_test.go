package shared

import (
	"embed"
	"flag"
	"os"
	"path/filepath"
	"testing"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	agentv1alpha1 "github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/api/v1alpha1"
)

var updateGolden = flag.Bool("update-golden", false, "regenerate pod-spec golden files")

// goldenFiles holds the goldens in the test binary, so a normal run reads no
// file; only -update-golden writes to testdata.
//
//go:embed testdata/*.golden.yaml
var goldenFiles embed.FS

// goldenJobs returns representative Jobs whose full serialized pod spec is frozen
// in testdata/*.golden.yaml. The goldens were generated from the post-merge
// builder, so they lock the pod spec against FUTURE drift; the merge itself is
// verified by the field-assertion tests in this package (job_test, workspace_job_test).
func goldenJobs(t *testing.T) map[string]*batchv1.Job {
	stdRun := func() *agentv1alpha1.AgentRun {
		return &agentv1alpha1.AgentRun{
			ObjectMeta: metav1.ObjectMeta{Name: "golden-run", Namespace: "default"},
			Spec: agentv1alpha1.AgentRunSpec{
				Workspace: &agentv1alpha1.WorkspaceSpec{
					Repository: agentv1alpha1.RepositorySpec{URL: "https://github.com/example/repo.git", Branch: "main"},
				},
			},
		}
	}
	wsRun := &agentv1alpha1.AgentRun{
		ObjectMeta: metav1.ObjectMeta{Name: "golden-ws-run", Namespace: "default"},
		Spec:       agentv1alpha1.AgentRunSpec{WorkspaceRef: "my-workspace"},
	}
	return map[string]*batchv1.Job{
		"standalone-claude-git": mustBuildJob(t, stdRun(), nil, "golden-pvc", "node:trixie-slim", time.Hour,
			agentv1alpha1.AgentTypeClaude, agentv1alpha1.WorkspaceTypeGit, nil, nil),
		"workspace-git": mustBuildJob(t, wsRun, nil, "my-workspace-ws", "node:trixie-slim", time.Hour,
			agentv1alpha1.AgentTypeClaude, agentv1alpha1.WorkspaceTypeGit, nil,
			&WorkspaceBinding{WorktreeBranch: "feature", WorkspaceRef: "my-workspace"}),
		"nix-image": mustBuildJob(t, stdRun(), nil, "golden-pvc", "ghcr.io/xonovex/agent@sha256:abc", time.Hour,
			agentv1alpha1.AgentTypeClaude, agentv1alpha1.WorkspaceTypeGit, nil, nil),
	}
}

func TestBuildJobGolden(t *testing.T) {
	for name, job := range goldenJobs(t) {
		got, err := yaml.Marshal(job)
		if err != nil {
			t.Fatalf("marshal %s: %v", name, err)
		}
		path := "testdata/" + name + ".golden.yaml"

		if *updateGolden {
			if err := os.MkdirAll("testdata", 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.FromSlash(path), got, 0o644); err != nil {
				t.Fatal(err)
			}
			continue
		}

		want, err := goldenFiles.ReadFile(path)
		if err != nil {
			t.Fatalf("read golden %s (run with -update-golden to generate): %v", path, err)
		}
		if string(got) != string(want) {
			t.Errorf("pod-spec golden mismatch for %q; run `go test -run TestBuildJobGolden -update-golden` if the change is intended", name)
		}
	}
}
