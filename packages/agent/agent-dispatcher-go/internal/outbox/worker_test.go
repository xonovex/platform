package outbox

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/xonovex/platform/packages/agent/agent-dispatcher-go/internal/domain"
	"github.com/xonovex/platform/packages/agent/agent-dispatcher-go/internal/metrics"
	"github.com/xonovex/platform/packages/agent/agent-dispatcher-go/internal/provider"
	"github.com/xonovex/platform/packages/agent/agent-dispatcher-go/internal/store"
)

func TestProcessPreviewAndApply(t *testing.T) {
	for _, mode := range []domain.EffectMode{domain.EffectModePreview, domain.EffectModeApply} {
		t.Run(string(mode), func(t *testing.T) {
			database := &fakeStore{effect: testEffect(mode)}
			adapter := &fakeAdapter{result: domain.EffectResult{StatusCode: 201, Reconciled: true}}
			registry := &metrics.Registry{}
			worker := testWorker(database, adapter, registry, true)
			if err := worker.ProcessOne(context.Background()); err != nil {
				t.Fatal(err)
			}
			if database.completed.ID != database.effect.ID || database.completedResult.StatusCode == 0 {
				t.Fatalf("effect was not completed: %+v %+v", database.completed, database.completedResult)
			}
			if mode == domain.EffectModePreview && registry.EffectsPreviewed.Load() != 1 {
				t.Error("preview metric not incremented")
			}
			if mode == domain.EffectModeApply && (registry.EffectsApplied.Load() != 1 || registry.EffectsReconciled.Load() != 1) {
				t.Error("apply metrics not incremented")
			}
		})
	}
}

func TestProcessRetriesAndDeadLetters(t *testing.T) {
	tests := []struct {
		name      string
		failure   error
		attempts  int
		apply     bool
		wantRetry bool
		wantDead  bool
		missing   bool
	}{
		{"retryable", &provider.RetryableError{StatusCode: 503}, 1, true, true, false, false},
		{"attempts exhausted", &provider.RetryableError{StatusCode: 503}, 8, true, false, true, false},
		{"permanent", &provider.PermanentError{StatusCode: 400, Message: "bad"}, 1, true, false, true, false},
		{"apply disabled", nil, 1, false, false, true, false},
		{"missing adapter", nil, 1, true, false, true, true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			effect := testEffect(domain.EffectModeApply)
			effect.Attempts = test.attempts
			database := &fakeStore{effect: effect}
			adapter := &fakeAdapter{failure: test.failure}
			registry := &metrics.Registry{}
			lookup := AdapterLookup(func(domain.Provider, string) (provider.Adapter, bool) {
				return adapter, !test.missing
			})
			worker := New(Options{
				Store: database, Adapter: lookup, Metrics: registry, Logger: discardLogger(),
				Holder: "worker", Interval: time.Millisecond, Lease: time.Minute, MaxAttempts: 8, Apply: test.apply,
			})
			worker.now = func() time.Time { return time.Unix(100, 0) }
			if err := worker.ProcessOne(context.Background()); err != nil {
				t.Fatal(err)
			}
			if (database.retryReason != "") != test.wantRetry || (database.deadReason != "") != test.wantDead {
				t.Fatalf("retry=%q dead=%q", database.retryReason, database.deadReason)
			}
		})
	}
}

func TestBackoffIsBounded(t *testing.T) {
	if backoff(1) != time.Second || backoff(20) != 512*time.Second {
		t.Fatalf("unexpected backoff: %s %s", backoff(1), backoff(20))
	}
}

type fakeStore struct {
	effect          domain.Effect
	completed       domain.Effect
	completedResult domain.EffectResult
	retryReason     string
	deadReason      string
}

func (*fakeStore) Close()                        {}
func (*fakeStore) Ping(context.Context) error    { return nil }
func (*fakeStore) Migrate(context.Context) error { return nil }
func (*fakeStore) Ingest(context.Context, domain.Delivery, domain.WorkflowEvent) (bool, error) {
	return true, nil
}
func (*fakeStore) EnqueueEffect(_ context.Context, effect domain.Effect) (domain.Effect, bool, error) {
	return effect, true, nil
}
func (database *fakeStore) GetEffect(context.Context, string) (domain.Effect, domain.EffectResult, error) {
	return database.effect, domain.EffectResult{}, nil
}
func (database *fakeStore) ClaimEffect(context.Context, string, time.Duration) (domain.Effect, error) {
	if database.effect.ID == "" {
		return domain.Effect{}, store.ErrNotFound
	}
	return database.effect, nil
}
func (*fakeStore) RenewEffectLease(context.Context, domain.Effect, string, time.Duration) error {
	return nil
}
func (*fakeStore) ValidateContextSequence(context.Context, domain.Effect) error {
	return nil
}
func (database *fakeStore) CompleteEffect(_ context.Context, effect domain.Effect, result domain.EffectResult) error {
	database.completed = effect
	database.completedResult = result
	return nil
}
func (database *fakeStore) RetryEffect(_ context.Context, _ domain.Effect, reason string, _ time.Time) error {
	database.retryReason = reason
	return nil
}
func (database *fakeStore) DeadLetterEffect(_ context.Context, _ domain.Effect, reason string) error {
	database.deadReason = reason
	return nil
}
func (*fakeStore) ReleaseExpiredLeases(context.Context) (int64, error) { return 0, nil }
func (*fakeStore) RequeueDeadLetter(context.Context, string) error     { return nil }
func (*fakeStore) Counts(context.Context) (store.Counts, error)        { return store.Counts{}, nil }

type fakeAdapter struct {
	result  domain.EffectResult
	failure error
}

func (*fakeAdapter) Preview(domain.Effect) (domain.RequestPreview, error) {
	return domain.RequestPreview{Requests: []domain.ProviderRequestPreview{{Method: "POST", URL: "https://example.test"}}}, nil
}
func (adapter *fakeAdapter) Apply(context.Context, domain.Effect) (domain.EffectResult, error) {
	return adapter.result, adapter.failure
}

func testWorker(database store.Store, adapter provider.Adapter, registry *metrics.Registry, apply bool) *Worker {
	return New(Options{
		Store: database,
		Adapter: func(domain.Provider, string) (provider.Adapter, bool) {
			return adapter, true
		},
		Metrics: registry, Logger: discardLogger(), Holder: "worker",
		Interval: time.Millisecond, Lease: time.Minute, MaxAttempts: 8, Apply: apply,
	})
}

func testEffect(mode domain.EffectMode) domain.Effect {
	return domain.Effect{
		ID: "effect", CorrelationID: "correlation", IdempotencyKey: "key",
		Provider: domain.ProviderGitHub, Tenant: "tenant", Kind: domain.EffectTicketCreate,
		Mode: mode, Target: domain.EffectTarget{Repository: "owner/repo"},
		Payload: domain.EffectPayload{Title: "Ticket"}, Attempts: 1,
	}
}

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

var _ store.Store = (*fakeStore)(nil)
var _ provider.Adapter = (*fakeAdapter)(nil)
