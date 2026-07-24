package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

var (
	ErrMissingSignature = errors.New("webhook signature is missing")
	ErrInvalidSignature = errors.New("webhook signature is invalid")
	ErrStaleDelivery    = errors.New("webhook timestamp is outside the replay window")
)

func VerifyGitHub(secret string, body []byte, signature string) error {
	if signature == "" {
		return ErrMissingSignature
	}
	if !strings.HasPrefix(signature, "sha256=") {
		return ErrInvalidSignature
	}
	received, err := hex.DecodeString(strings.TrimPrefix(signature, "sha256="))
	if err != nil {
		return ErrInvalidSignature
	}
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(body)
	if !hmac.Equal(received, mac.Sum(nil)) {
		return ErrInvalidSignature
	}
	return nil
}

func VerifyGitLabStandard(
	signingToken string,
	messageID string,
	timestamp string,
	body []byte,
	signatures string,
	now time.Time,
	replayWindow time.Duration,
) error {
	if messageID == "" || timestamp == "" || signatures == "" {
		return ErrMissingSignature
	}
	seconds, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return ErrInvalidSignature
	}
	signedAt := time.Unix(seconds, 0)
	if signedAt.Before(now.Add(-replayWindow)) || signedAt.After(now.Add(replayWindow)) {
		return ErrStaleDelivery
	}
	encodedKey := strings.TrimPrefix(signingToken, "whsec_")
	key, err := base64.StdEncoding.DecodeString(encodedKey)
	if err != nil || len(key) == 0 {
		return fmt.Errorf("decode GitLab signing token: %w", ErrInvalidSignature)
	}
	message := messageID + "." + timestamp + "." + string(body)
	mac := hmac.New(sha256.New, key)
	_, _ = mac.Write([]byte(message))
	expected := "v1," + base64.StdEncoding.EncodeToString(mac.Sum(nil))
	for signature := range strings.FieldsSeq(signatures) {
		if subtle.ConstantTimeCompare([]byte(expected), []byte(signature)) == 1 {
			return nil
		}
	}
	return ErrInvalidSignature
}

func VerifyGitLabLegacy(secret string, received string) error {
	if received == "" {
		return ErrMissingSignature
	}
	if subtle.ConstantTimeCompare([]byte(secret), []byte(received)) != 1 {
		return ErrInvalidSignature
	}
	return nil
}
