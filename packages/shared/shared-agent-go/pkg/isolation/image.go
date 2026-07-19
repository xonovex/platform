package isolation

import (
	"encoding/hex"
	"strings"
)

// IsDigestPinnedImage reports whether image names a repository and carries a
// complete immutable SHA-256 digest. A human-readable tag may precede the
// digest.
func IsDigestPinnedImage(image string) bool {
	parts := strings.SplitN(image, "@sha256:", 2)
	if len(parts) != 2 || parts[0] == "" || len(parts[1]) != 64 {
		return false
	}
	_, err := hex.DecodeString(parts[1])
	return err == nil
}
