package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"testing"
	"time"
)

func TestVerifyGitHub(t *testing.T) {
	body := []byte(`{"action":"opened"}`)
	mac := hmac.New(sha256.New, []byte("secret"))
	_, _ = mac.Write(body)
	signature := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	if err := VerifyGitHub("secret", body, signature); err != nil {
		t.Fatal(err)
	}
	if !errors.Is(VerifyGitHub("wrong", body, signature), ErrInvalidSignature) {
		t.Error("expected invalid GitHub signature")
	}
	if !errors.Is(VerifyGitHub("secret", body, ""), ErrMissingSignature) {
		t.Error("expected missing GitHub signature")
	}
}

func TestVerifyGitLabStandardAndLegacy(t *testing.T) {
	key := []byte("signing-secret")
	token := "whsec_" + base64.StdEncoding.EncodeToString(key)
	body := []byte(`{"object_kind":"issue"}`)
	now := time.Unix(1_800_000_000, 0)
	messageID := "delivery"
	timestamp := "1800000000"
	mac := hmac.New(sha256.New, key)
	_, _ = mac.Write([]byte(messageID + "." + timestamp + "." + string(body)))
	signature := "v1," + base64.StdEncoding.EncodeToString(mac.Sum(nil))
	if err := VerifyGitLabStandard(token, messageID, timestamp, body, "v1,bad "+signature, now, time.Minute); err != nil {
		t.Fatal(err)
	}
	if !errors.Is(VerifyGitLabStandard(token, messageID, "1700000000", body, signature, now, time.Minute), ErrStaleDelivery) {
		t.Error("expected stale GitLab delivery")
	}
	if !errors.Is(VerifyGitLabStandard("bad", messageID, timestamp, body, signature, now, time.Minute), ErrInvalidSignature) {
		t.Error("expected invalid signing token")
	}
	if err := VerifyGitLabLegacy("legacy", "legacy"); err != nil {
		t.Fatal(err)
	}
	if !errors.Is(VerifyGitLabLegacy("legacy", "wrong"), ErrInvalidSignature) {
		t.Error("expected invalid legacy token")
	}
}

func TestNormalizeGitHub(t *testing.T) {
	body := []byte(`{
		"action":"synchronize",
		"repository":{"full_name":"owner/repo"},
		"sender":{"login":"dispatcher[bot]"},
		"pull_request":{"id":9,"number":7,"head":{"sha":"abc"}}
	}`)
	event, err := NormalizeGitHub("tenant", "delivery", "pull_request", body, func(login string) bool {
		return login == "dispatcher[bot]"
	})
	if err != nil {
		t.Fatal(err)
	}
	if event.Repository != "owner/repo" || event.SubjectKind != "pull_request" ||
		event.SubjectNumber != 7 || event.Revision != "abc" || !event.Suppressed {
		t.Fatalf("unexpected normalized event: %+v", event)
	}
	if _, err := NormalizeGitHub("tenant", "delivery", "issues", []byte(`{}`), func(string) bool { return false }); err == nil {
		t.Error("expected missing repository error")
	}
}

func TestNormalizeGitLab(t *testing.T) {
	body := []byte(`{
		"object_kind":"merge_request",
		"event_type":"merge_request",
		"project":{"path_with_namespace":"group/project"},
		"user":{"username":"reviewer"},
		"object_attributes":{"id":8,"iid":3,"action":"update","last_commit":{"id":"def"}}
	}`)
	event, err := NormalizeGitLab("tenant", "delivery", "event", "Merge Request Hook", body, func(string) bool { return false })
	if err != nil {
		t.Fatal(err)
	}
	if event.Kind != "merge_request" || event.Repository != "group/project" ||
		event.SubjectNumber != 3 || event.Revision != "def" || event.Actor != "reviewer" {
		t.Fatalf("unexpected normalized event: %+v", event)
	}
}
