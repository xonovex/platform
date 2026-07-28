package config_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/yaml"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
	"github.com/xonovex/platform/packages/agent/agent-operator-go/internal/validator"
	agentwebhook "github.com/xonovex/platform/packages/agent/agent-operator-go/internal/webhook"
)

func TestSamplesStrictlyDecodeAgainstAgentAPI(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := agentv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("add agent API to scheme: %v", err)
	}
	decoder := serializer.NewCodecFactory(scheme, serializer.EnableStrict).UniversalDeserializer()
	paths, err := filepath.Glob(filepath.Join("samples", "*.yaml"))
	if err != nil {
		t.Fatalf("list samples: %v", err)
	}
	if len(paths) == 0 {
		t.Fatal("no sample manifests found")
	}

	for _, path := range paths {
		t.Run(filepath.Base(path), func(t *testing.T) {
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read sample: %v", err)
			}
			if _, _, err := decoder.Decode(data, nil, nil); err != nil {
				t.Fatalf("strictly decode sample: %v", err)
			}
		})
	}
}

func TestWorkspaceSampleMountsConfigurationInRuntimeHome(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("samples", "agentworkspace_sample.yaml"))
	if err != nil {
		t.Fatalf("read workspace sample: %v", err)
	}
	var workspace agentv1alpha1.AgentWorkspace
	if err := yaml.UnmarshalStrict(data, &workspace); err != nil {
		t.Fatalf("decode workspace sample: %v", err)
	}

	for _, volume := range workspace.Spec.SharedVolumes {
		if !strings.HasPrefix(volume.MountPath, "/home/agent/") {
			t.Errorf("shared volume %q mountPath = %q, want runtime home", volume.Name, volume.MountPath)
		}
	}
}

func TestSampleAgentRunPassesAdmissionWithSamplePolicyAndProvider(t *testing.T) {
	policy := decodeSample[agentv1alpha1.AgentPolicy](t, "agentpolicy_sample.yaml", map[string]string{
		"<digest>": strings.Repeat("a", 64),
	})
	provider := decodeSample[agentv1alpha1.AgentProvider](t, "agentprovider_sample.yaml", nil)
	run := decodeSample[agentv1alpha1.AgentRun](t, "agentrun_sample.yaml", nil)
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatalf("add core API to scheme: %v", err)
	}
	if err := agentv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("add agent API to scheme: %v", err)
	}
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "anthropic-credentials", Namespace: "ai-agents"},
		Data:       map[string][]byte{"api-key": []byte("test-token")},
	}
	fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(policy, provider, secret).Build()
	admissionWebhook := &agentwebhook.AgentRunWebhook{Client: fakeClient}

	warnings, err := admissionWebhook.ValidateCreate(context.Background(), run)

	if err != nil {
		t.Fatalf("sample AgentRun admission failed: %v", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("sample AgentRun admission warnings = %v, want none", warnings)
	}
}

func TestSampleAgentToolchainPassesAdmission(t *testing.T) {
	toolchain := decodeSample[agentv1alpha1.AgentToolchain](t, "agenttoolchain_sample.yaml", map[string]string{
		"<nixpkgsRev>": strings.Repeat("b", 40),
		"<digest>":     strings.Repeat("a", 64),
	})
	toolchainWebhook := &agentwebhook.AgentToolchainWebhook{}

	warnings, err := toolchainWebhook.ValidateCreate(context.Background(), toolchain)

	if err != nil {
		t.Fatalf("sample AgentToolchain admission failed: %v", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("sample AgentToolchain admission warnings = %v, want none", warnings)
	}
}

func TestSampleAgentHarnessPassesAdmission(t *testing.T) {
	harness := decodeSample[agentv1alpha1.AgentHarness](t, "agentharness_sample.yaml", map[string]string{
		"<digest>": strings.Repeat("a", 64),
	})
	harnessWebhook := &agentwebhook.AgentHarnessWebhook{}

	warnings, err := harnessWebhook.ValidateCreate(context.Background(), harness)

	if err != nil {
		t.Fatalf("sample AgentHarness admission failed: %v", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("sample AgentHarness admission warnings = %v, want none", warnings)
	}
}

func decodeSample[T any](t testing.TB, name string, replacements map[string]string) *T {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("samples", name))
	if err != nil {
		t.Fatalf("read sample %q: %v", name, err)
	}
	contents := string(data)
	for oldValue, newValue := range replacements {
		contents = strings.ReplaceAll(contents, oldValue, newValue)
	}
	var value T
	if err := yaml.UnmarshalStrict([]byte(contents), &value); err != nil {
		t.Fatalf("decode sample %q: %v", name, err)
	}
	return &value
}

func TestManagerImageUsesImmutableDigest(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("manager", "manager.yaml"))
	if err != nil {
		t.Fatalf("read manager manifest: %v", err)
	}
	var deployment appsv1.Deployment
	if err := yaml.UnmarshalStrict(data, &deployment); err != nil {
		t.Fatalf("decode manager manifest: %v", err)
	}
	if len(deployment.Spec.Template.Spec.Containers) != 1 {
		t.Fatalf("manager containers = %d, want 1", len(deployment.Spec.Template.Spec.Containers))
	}
	image := deployment.Spec.Template.Spec.Containers[0].Image
	if err := validator.ValidatePinnedImageReference(image); err != nil {
		t.Fatalf("manager image %q is mutable: %v", image, err)
	}
}
