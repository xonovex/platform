package config_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"sigs.k8s.io/yaml"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
	"github.com/xonovex/platform/packages/agent/agent-operator-go/internal/validator"
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
