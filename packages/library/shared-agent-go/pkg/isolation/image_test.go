package isolation

import "testing"

func TestIsDigestPinnedImage(t *testing.T) {
	digest := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	tests := []struct {
		name  string
		image string
		want  bool
	}{
		{name: "digest", image: "ghcr.io/xonovex/agent@sha256:" + digest, want: true},
		{name: "tag and digest", image: "ghcr.io/xonovex/agent:v1@sha256:" + digest, want: true},
		{name: "mutable tag", image: "ghcr.io/xonovex/agent:v1", want: false},
		{name: "short digest", image: "ghcr.io/xonovex/agent@sha256:abc", want: false},
		{name: "missing name", image: "@sha256:" + digest, want: false},
		{name: "non hexadecimal", image: "ghcr.io/xonovex/agent@sha256:gggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggg", want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := IsDigestPinnedImage(test.image); got != test.want {
				t.Errorf("IsDigestPinnedImage(%q) = %v, want %v", test.image, got, test.want)
			}
		})
	}
}
