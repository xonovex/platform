package testutil

import (
	"fmt"
	"strings"
)

// ImageDigestFromList reads the digest of the image that `ctr images ls`
// reports under name. Each row holds the name, the media type and the digest.
func ImageDigestFromList(listing, name string) (string, error) {
	for _, line := range strings.Split(listing, "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 3 && fields[0] == name && strings.HasPrefix(fields[2], "sha256:") {
			return fields[2], nil
		}
	}
	return "", fmt.Errorf("image %s is not in the node image list", name)
}
