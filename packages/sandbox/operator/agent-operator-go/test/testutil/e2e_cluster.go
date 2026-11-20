//go:build e2e || e2e_gvisor || e2e_kata || e2e_coco

package testutil

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

const kindImageNamespace = "docker.io/library/"

// LoadKindImage copies a local Docker Hub library image, such as
// "busybox:1.37", into the control-plane node of a Kind cluster and returns a
// reference that names it by digest, which is the only form an AgentRun
// accepts. `kind load docker-image` exports every platform of the image and
// fails when the containerd image store holds only one, so this loads a
// single-platform archive and tags the imported image under its digest.
func LoadKindImage(cluster, ref string) (string, error) {
	repository, _, tagged := strings.Cut(ref, ":")
	if !tagged || strings.Contains(ref, "/") {
		return "", fmt.Errorf("image %q must have the form name:tag", ref)
	}
	node := cluster + "-control-plane"
	name := kindImageNamespace + ref

	arch, err := output("docker", "version", "--format", "{{.Server.Arch}}")
	if err != nil {
		return "", err
	}
	dir, err := os.MkdirTemp("", "kind-image-*")
	if err != nil {
		return "", err
	}
	defer func() { _ = os.RemoveAll(dir) }()
	archive := filepath.Join(dir, "image.tar")

	if _, err := output("docker", "save", "--platform", "linux/"+strings.TrimSpace(arch), "-o", archive, ref); err != nil {
		return "", err
	}
	if _, err := output("kind", "load", "image-archive", archive, "--name", cluster); err != nil {
		return "", err
	}
	listing, err := output("docker", "exec", node, "ctr", "--namespace=k8s.io", "images", "ls", "name=="+name)
	if err != nil {
		return "", err
	}
	digest, err := ImageDigestFromList(listing, name)
	if err != nil {
		return "", err
	}
	pinned := repository + "@" + digest
	if _, err := output("docker", "exec", node, "ctr", "--namespace=k8s.io", "images", "tag", name, kindImageNamespace+pinned); err != nil {
		return "", err
	}
	return pinned, waitForNodeImage(node, kindImageNamespace+pinned)
}

// waitForNodeImage waits until the container runtime interface of the node
// resolves the digest reference, which follows the tag by a short delay.
func waitForNodeImage(node, ref string) error {
	var err error
	for range 20 {
		if _, err = output("docker", "exec", node, "crictl", "inspecti", ref); err == nil {
			return nil
		}
		time.Sleep(500 * time.Millisecond)
	}
	return err
}

func output(name string, args ...string) (string, error) {
	var stdout, stderr bytes.Buffer
	cmd := exec.Command(name, args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("%s %s: %w\n%s", name, strings.Join(args, " "), err, stderr.String())
	}
	return stdout.String(), nil
}

// CreateAndWait creates obj and waits until c reads it back. The manager client
// reads from the cache that the controllers read from, so a run created after
// this call finds the object it references; without the wait a controller can
// reconcile the run before its cache holds the new object and fail the run.
func CreateAndWait(ctx context.Context, c client.Client, obj client.Object) error {
	if err := c.Create(ctx, obj); err != nil {
		return err
	}
	deadline := time.Now().Add(10 * time.Second)
	for {
		err := c.Get(ctx, client.ObjectKeyFromObject(obj), obj.DeepCopyObject().(client.Object))
		if err == nil {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("%T %s is not readable after create: %w", obj, client.ObjectKeyFromObject(obj), err)
		}
		time.Sleep(50 * time.Millisecond)
	}
}
