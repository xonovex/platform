package store

import (
	"context"
	"errors"
	"time"

	"github.com/xonovex/platform/packages/agent/agent-dispatcher-go/internal/domain"
)

var (
	ErrNotFound            = errors.New("record not found")
	ErrContextConflict     = errors.New("context identity conflicts with stored content")
	ErrIdempotencyConflict = errors.New("idempotency key conflicts with stored effect")
)

type Counts struct {
	PendingDeliveries int64 `json:"pending_deliveries"`
	QueuedEffects     int64 `json:"queued_effects"`
	RetryEffects      int64 `json:"retry_effects"`
	DeadLetters       int64 `json:"dead_letters"`
}

type Store interface {
	Close()
	Ping(context.Context) error
	Migrate(context.Context) error
	Ingest(context.Context, domain.Delivery, domain.WorkflowEvent) (bool, error)
	EnqueueEffect(context.Context, domain.Effect) (domain.Effect, bool, error)
	GetEffect(context.Context, string) (domain.Effect, domain.EffectResult, error)
	ClaimEffect(context.Context, string, time.Duration) (domain.Effect, error)
	RenewEffectLease(context.Context, domain.Effect, string, time.Duration) error
	ValidateContextSequence(context.Context, domain.Effect) error
	CompleteEffect(context.Context, domain.Effect, domain.EffectResult) error
	RetryEffect(context.Context, domain.Effect, string, time.Time) error
	DeadLetterEffect(context.Context, domain.Effect, string) error
	ReleaseExpiredLeases(context.Context) (int64, error)
	RequeueDeadLetter(context.Context, string) error
	Counts(context.Context) (Counts, error)
}
