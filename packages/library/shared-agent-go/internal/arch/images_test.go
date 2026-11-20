//go:build arch

package arch

import (
	"path/filepath"
	"regexp"
	"testing"
)

// imageModules are the Go modules whose production code names no container
// image: every library module, the agent CLI, the eval CLI and the operator,
// relative to the packages directory. An image is a setting of the user or the
// deployment, never a default in code.
var imageModules = []string{
	"library/shared-core-go",
	"library/shared-contract-go",
	"library/shared-config-go",
	"library/shared-agent-go",
	"library/shared-launch-go",
	"tooling/cli/agent-cli-go",
	"tooling/cli/agent-eval-go",
	"sandbox/operator/agent-operator-go",
}

// imageReference matches a container image reference that a runtime can pull: a
// repository path, a registry host optional, with a tag or a sha256 digest. A
// bare name with a tag, such as "name:1", is not matched, because it reads the
// same as a "key:value" pair.
var imageReference = regexp.MustCompile(`^(?:[a-z0-9]+(?:[.-][a-z0-9]+)*(?::[0-9]+)?/)?[a-z0-9]+(?:(?:[._]|__|-+)[a-z0-9]+)*(?:/[a-z0-9]+(?:(?:[._]|__|-+)[a-z0-9]+)*)*(?::[A-Za-z0-9_][A-Za-z0-9_.-]{0,127})?(?:@sha256:[0-9a-f]{64})?$`)

var (
	// imageDigest matches the sha256 digest that ends an image reference.
	imageDigest = regexp.MustCompile(`@sha256:[0-9a-f]{64}$`)
	// imageTag matches the tag of the last path component of an image reference.
	imageTag = regexp.MustCompile(`/[^/@]*:[A-Za-z0-9_][A-Za-z0-9_.-]*(?:@|$)`)
)

// containerImage flags an image reference that has a repository path and a
// tag, or that has a sha256 digest.
func containerImage(value string) (string, bool) {
	if !imageReference.MatchString(value) {
		return "", false
	}
	return "container image", imageDigest.MatchString(value) || imageTag.MatchString(value)
}

func TestProductionCodeNamesNoContainerImage(t *testing.T) {
	for _, module := range imageModules {
		for _, f := range scan(t, filepath.Join(packagesDir(t), filepath.FromSlash(module)), containerImage) {
			t.Errorf("%s: the %s %q belongs in a setting of the user or the deployment, not in code", f.where, f.kind, f.literal)
		}
	}
}

func TestTheImageScannerRecognizesWhatItForbids(t *testing.T) {
	const digest = "@sha256:95a34da32a840bd9b3b09a5b773591c16923e350174b1c50e1200c75bf15eaa9"
	for _, image := range []string{
		"docker.io/library/node:26.3.0-trixie-slim" + digest,
		"docker.io/alpine/git:2.54.1" + digest,
		"ghcr.io/xonovex/agent:latest",
		"localhost:5000/agent:v1",
		"registry.example/agent" + digest,
		"node" + digest,
		"library/node:22-alpine",
	} {
		if _, forbidden := containerImage(image); !forbidden {
			t.Errorf("containerImage does not flag %q", image)
		}
	}
	for _, ok := range []string{
		"node:22", "key:value", "localhost:8080", "https://example.com/a:b", "claude", "github.com/xonovex/platform",
		"pkg/isolation", "run.image", "%s:%s", "/nix/store/abc-x:ro", "image must use an immutable @sha256: digest",
		"sha256:" + digest[len("@sha256:"):],
	} {
		if _, forbidden := containerImage(ok); forbidden {
			t.Errorf("containerImage flags %q, which names no pullable image", ok)
		}
	}
}
