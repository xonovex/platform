package validator

import "testing"

func TestValidatePinnedImageReference(t *testing.T) {
	valid := "docker.io/alpine/git:2.54.1@sha256:2306e2b05ee0ef99f5d25b0779f645b24ae9f91ba187107503592fcdfe78da3a"
	if err := ValidatePinnedImageReference(valid); err != nil {
		t.Fatalf("valid image rejected: %v", err)
	}
	for _, image := range []string{"alpine/git:latest", "alpine/git@sha256:short", "@sha256:697cb1c85aefc5724febaec2202a974e0d66f6abb6be91a9a86d0c8757af692a"} {
		if err := ValidatePinnedImageReference(image); err == nil {
			t.Fatalf("moving or malformed image accepted: %q", image)
		}
	}
}
