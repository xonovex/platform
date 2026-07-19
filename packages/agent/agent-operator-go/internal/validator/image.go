package validator

import (
	"encoding/hex"
	"fmt"
	"strings"
)

// ValidatePinnedImageReference requires a named image plus a full immutable
// SHA-256 digest. A human-readable tag may remain before the digest.
func ValidatePinnedImageReference(image string) error {
	parts := strings.SplitN(image, "@sha256:", 2)
	if len(parts) != 2 || parts[0] == "" || len(parts[1]) != 64 {
		return fmt.Errorf("image must use an immutable @sha256: digest")
	}
	if _, err := hex.DecodeString(parts[1]); err != nil {
		return fmt.Errorf("image must use an immutable @sha256: digest")
	}
	return nil
}
