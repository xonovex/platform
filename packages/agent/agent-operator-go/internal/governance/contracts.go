package governance

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
)

// ContentAddressedEvidenceReference accepts an opaque provider reference only
// when it terminates in one exact SHA-256 content digest.
func ContentAddressedEvidenceReference(reference string) bool {
	prefix, digest, found := strings.Cut(reference, "#")
	return found && strings.TrimSpace(prefix) != "" && SHA256Digest(digest)
}

// SHA256Digest validates the canonical sha256:<lowercase-hex> representation.
func SHA256Digest(value string) bool {
	digest := strings.TrimPrefix(value, "sha256:")
	if digest == value || len(digest) != 64 || strings.ToLower(digest) != digest {
		return false
	}
	_, err := hex.DecodeString(digest)
	return err == nil
}

// DigestJSON returns a stable digest for JSON-compatible data. encoding/json
// sorts map keys, so equal objects have the same digest regardless of key order.
func DigestJSON(value any) (string, error) {
	canonical, err := json.Marshal(value)
	if err != nil {
		return "", fmt.Errorf("canonicalize JSON value: %w", err)
	}
	digest := sha256.Sum256(canonical)
	return fmt.Sprintf("sha256:%x", digest), nil
}

// AgentRunSubjectRevision identifies the exact immutable execution inputs.
func AgentRunSubjectRevision(run *agentv1alpha1.AgentRun) (string, error) {
	digest, err := DigestJSON(run.Spec)
	if err != nil {
		return "", err
	}
	return "spec-" + digest, nil
}
