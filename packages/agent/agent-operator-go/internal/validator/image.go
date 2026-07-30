package validator

import (
	"fmt"

	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/isolation"
)

// ValidatePinnedImageReference requires a named image plus a full immutable
// SHA-256 digest. A human-readable tag may remain before the digest.
func ValidatePinnedImageReference(image string) error {
	if !isolation.IsDigestPinnedImage(image) {
		return fmt.Errorf("image must use an immutable @sha256: digest")
	}
	return nil
}
