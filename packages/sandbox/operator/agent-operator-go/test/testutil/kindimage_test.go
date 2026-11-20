package testutil

import "testing"

func TestImageDigestFromList(t *testing.T) {
	listing := "REF TYPE DIGEST SIZE PLATFORMS LABELS\n" +
		"docker.io/library/other:1 application/vnd.oci.image.index.v1+json sha256:aaaa 1.0 MiB linux/amd64 -\n" +
		"docker.io/library/busybox:1.37 application/vnd.oci.image.index.v1+json sha256:bbbb 2.1 MiB linux/amd64 -\n"

	got, err := ImageDigestFromList(listing, "docker.io/library/busybox:1.37")
	if err != nil || got != "sha256:bbbb" {
		t.Fatalf("ImageDigestFromList() = %q, %v, want sha256:bbbb", got, err)
	}
}

func TestImageDigestFromListRejectsMissingImage(t *testing.T) {
	if _, err := ImageDigestFromList("REF TYPE DIGEST\n", "docker.io/library/busybox:1.37"); err == nil {
		t.Fatal("ImageDigestFromList() accepted a listing without the image")
	}
}
