package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/xonovex/platform/packages/agent/agent-dispatcher-go/internal/domain"
)

var (
	ErrConflict         = errors.New("provider content conflicts with the requested effect")
	ErrDuplicateContext = errors.New("provider contains duplicate context records")
)

type Adapter interface {
	Preview(domain.Effect) (domain.RequestPreview, error)
	Apply(context.Context, domain.Effect) (domain.EffectResult, error)
}

type RetryableError struct {
	StatusCode int
	RetryAfter time.Duration
	Message    string
}

func (failure *RetryableError) Error() string {
	return fmt.Sprintf("provider request is retryable: status=%d message=%s", failure.StatusCode, failure.Message)
}

type PermanentError struct {
	StatusCode int
	Message    string
	Body       json.RawMessage
}

func (failure *PermanentError) Error() string {
	return fmt.Sprintf("provider request failed: status=%d message=%s", failure.StatusCode, failure.Message)
}
