package outbox

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"time"

	"github.com/xonovex/platform/packages/agent/agent-dispatcher-go/internal/domain"
	"github.com/xonovex/platform/packages/agent/agent-dispatcher-go/internal/metrics"
	"github.com/xonovex/platform/packages/agent/agent-dispatcher-go/internal/provider"
	"github.com/xonovex/platform/packages/agent/agent-dispatcher-go/internal/store"
)

type AdapterLookup func(domain.Provider, string) (provider.Adapter, bool)

type Worker struct {
	store       store.Store
	adapter     AdapterLookup
	metrics     *metrics.Registry
	logger      *slog.Logger
	holder      string
	interval    time.Duration
	lease       time.Duration
	maxAttempts int
	apply       bool
	now         func() time.Time
}

type Options struct {
	Store       store.Store
	Adapter     AdapterLookup
	Metrics     *metrics.Registry
	Logger      *slog.Logger
	Holder      string
	Interval    time.Duration
	Lease       time.Duration
	MaxAttempts int
	Apply       bool
}

func New(options Options) *Worker {
	return &Worker{
		store:       options.Store,
		adapter:     options.Adapter,
		metrics:     options.Metrics,
		logger:      options.Logger,
		holder:      options.Holder,
		interval:    options.Interval,
		lease:       options.Lease,
		maxAttempts: options.MaxAttempts,
		apply:       options.Apply,
		now:         time.Now,
	}
}

func (worker *Worker) Run(ctx context.Context) error {
	if _, err := worker.store.ReleaseExpiredLeases(ctx); err != nil {
		return fmt.Errorf("recover dispatcher leases: %w", err)
	}
	ticker := time.NewTicker(worker.interval)
	defer ticker.Stop()
	for {
		if err := worker.ProcessOne(ctx); err != nil && !errors.Is(err, store.ErrNotFound) {
			worker.logger.Error("outbox processing failed", "error", err)
		}
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
}

func (worker *Worker) ProcessOne(ctx context.Context) error {
	effect, err := worker.store.ClaimEffect(ctx, worker.holder, worker.lease)
	if err != nil {
		return err
	}
	adapter, ok := worker.adapter(effect.Provider, effect.Tenant)
	if !ok {
		return worker.deadLetter(ctx, effect, "provider adapter is not configured")
	}
	if effect.Mode == domain.EffectModeApply && !worker.apply {
		return worker.deadLetter(ctx, effect, "apply mode is disabled")
	}
	operationContext, cancelOperation := context.WithCancel(ctx)
	renewalFailure := worker.renewLease(operationContext, cancelOperation, effect)
	defer cancelOperation()
	if err := worker.store.ValidateContextSequence(operationContext, effect); err != nil {
		cancelOperation()
		return worker.deadLetter(ctx, effect, err.Error())
	}

	if effect.Mode == domain.EffectModePreview {
		preview, err := adapter.Preview(effect)
		if err != nil {
			cancelOperation()
			return worker.fail(ctx, effect, err)
		}
		body, err := json.Marshal(preview)
		if err != nil {
			cancelOperation()
			return worker.deadLetter(ctx, effect, "encode preview: "+err.Error())
		}
		cancelOperation()
		if renewalError := receiveRenewalFailure(renewalFailure); renewalError != nil {
			return renewalError
		}
		if err := worker.store.CompleteEffect(ctx, effect, domain.EffectResult{
			StatusCode: 200,
			Body:       body,
		}); err != nil {
			return err
		}
		worker.metrics.EffectsPreviewed.Add(1)
		return nil
	}

	result, err := adapter.Apply(operationContext, effect)
	if err != nil {
		cancelOperation()
		if renewalError := receiveRenewalFailure(renewalFailure); renewalError != nil {
			return renewalError
		}
		return worker.fail(ctx, effect, err)
	}
	cancelOperation()
	if renewalError := receiveRenewalFailure(renewalFailure); renewalError != nil {
		return renewalError
	}
	if err := worker.store.CompleteEffect(ctx, effect, result); err != nil {
		return err
	}
	worker.metrics.EffectsApplied.Add(1)
	if result.Reconciled {
		worker.metrics.EffectsReconciled.Add(1)
	}
	return nil
}

func (worker *Worker) renewLease(
	ctx context.Context,
	cancel context.CancelFunc,
	effect domain.Effect,
) <-chan error {
	failure := make(chan error, 1)
	interval := worker.lease / 3
	if interval < 100*time.Millisecond {
		interval = 100 * time.Millisecond
	}
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := worker.store.RenewEffectLease(ctx, effect, worker.holder, worker.lease); err != nil {
					failure <- fmt.Errorf("renew outbox lease: %w", err)
					cancel()
					return
				}
			}
		}
	}()
	return failure
}

func receiveRenewalFailure(failure <-chan error) error {
	select {
	case err := <-failure:
		return err
	default:
		return nil
	}
}

func (worker *Worker) fail(ctx context.Context, effect domain.Effect, failure error) error {
	var retryable *provider.RetryableError
	if errors.As(failure, &retryable) && effect.Attempts < worker.maxAttempts {
		delay := retryable.RetryAfter
		if delay <= 0 {
			delay = backoff(effect.Attempts)
		}
		if err := worker.store.RetryEffect(ctx, effect, failure.Error(), worker.now().UTC().Add(delay)); err != nil {
			return err
		}
		worker.metrics.EffectsRetried.Add(1)
		return nil
	}
	return worker.deadLetter(ctx, effect, failure.Error())
}

func (worker *Worker) deadLetter(ctx context.Context, effect domain.Effect, reason string) error {
	if err := worker.store.DeadLetterEffect(ctx, effect, reason); err != nil {
		return err
	}
	worker.metrics.EffectsDeadLettered.Add(1)
	return nil
}

func backoff(attempt int) time.Duration {
	exponent := math.Min(float64(max(attempt-1, 0)), 9)
	return time.Duration(math.Pow(2, exponent)) * time.Second
}
