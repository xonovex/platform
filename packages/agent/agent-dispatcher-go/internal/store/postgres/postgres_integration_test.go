package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/xonovex/platform/packages/agent/agent-dispatcher-go/internal/domain"
	"github.com/xonovex/platform/packages/agent/agent-dispatcher-go/internal/store"
)

func TestStoreLifecycle(t *testing.T) {
	databaseURL := os.Getenv("DISPATCHER_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DISPATCHER_TEST_DATABASE_URL is not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	database, err := Open(ctx, databaseURL)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(database.Close)
	if err := database.Ping(ctx); err != nil {
		t.Fatalf("Ping() error = %v", err)
	}
	if err := database.Migrate(ctx); err != nil {
		t.Fatalf("Migrate() error = %v", err)
	}

	testID, err := domain.NewID("integration")
	if err != nil {
		t.Fatalf("NewID() error = %v", err)
	}
	tenant := testID
	deliveryID := testID + ":delivery"
	correlationIDs := []string{
		deliveryID,
		testID + ":effect:complete",
		testID + ":effect:dead",
		testID + ":effect:expired",
		testID + ":context:v1",
		testID + ":context:v2",
	}
	t.Cleanup(func() {
		cleanupContext, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cleanupCancel()
		cleanup := []struct {
			query     string
			arguments []any
		}{
			{
				query: `
					DELETE FROM dispatcher_dead_letters
					WHERE effect_id IN (SELECT id FROM dispatcher_effects WHERE tenant=$1)`,
				arguments: []any{tenant},
			},
			{
				query: `
					DELETE FROM dispatcher_resource_leases
					WHERE resource_key IN (SELECT resource_key FROM dispatcher_effects WHERE tenant=$1)`,
				arguments: []any{tenant},
			},
			{query: `DELETE FROM dispatcher_context_records WHERE tenant=$1`, arguments: []any{tenant}},
			{
				query:     `DELETE FROM dispatcher_audit_events WHERE correlation_id=ANY($1)`,
				arguments: []any{correlationIDs},
			},
			{query: `DELETE FROM dispatcher_effects WHERE tenant=$1`, arguments: []any{tenant}},
			{query: `DELETE FROM dispatcher_workflow_events WHERE tenant=$1`, arguments: []any{tenant}},
			{query: `DELETE FROM dispatcher_deliveries WHERE tenant=$1`, arguments: []any{tenant}},
		}
		for _, operation := range cleanup {
			if _, cleanupErr := database.pool.Exec(cleanupContext, operation.query, operation.arguments...); cleanupErr != nil {
				t.Errorf("cleanup integration records: %v", cleanupErr)
			}
		}
	})

	payload := json.RawMessage(`{"action":"opened"}`)
	delivery := domain.Delivery{
		Provider:      domain.ProviderGitHub,
		Tenant:        tenant,
		DeliveryID:    deliveryID,
		Event:         "issues",
		Action:        "opened",
		PayloadDigest: domain.Digest(payload),
		Payload:       payload,
		Headers:       map[string]string{"X-GitHub-Event": "issues"},
		State:         domain.DeliveryNormalized,
		ReceivedAt:    time.Now().UTC(),
	}
	event := domain.WorkflowEvent{
		Provider:      delivery.Provider,
		Tenant:        tenant,
		DeliveryID:    deliveryID,
		Kind:          "ticket",
		Action:        "opened",
		Repository:    "xonovex/integration",
		SubjectKind:   "issue",
		SubjectID:     "1",
		SubjectNumber: 1,
		Payload:       payload,
	}
	inserted, err := database.Ingest(ctx, delivery, event)
	if err != nil || !inserted {
		t.Fatalf("Ingest() = %v, %v; want true, nil", inserted, err)
	}
	inserted, err = database.Ingest(ctx, delivery, event)
	if err != nil || inserted {
		t.Fatalf("duplicate Ingest() = %v, %v; want false, nil", inserted, err)
	}

	complete := ticketEffect(testID+":effect:complete", tenant)
	stored, created, err := database.EnqueueEffect(ctx, complete)
	if err != nil || !created || stored.ID != complete.ID {
		t.Fatalf("EnqueueEffect() = %#v, %v, %v", stored, created, err)
	}
	stored, created, err = database.EnqueueEffect(ctx, complete)
	if err != nil || created || stored.ID != complete.ID {
		t.Fatalf("duplicate EnqueueEffect() = %#v, %v, %v", stored, created, err)
	}
	divergent := complete
	divergent.Payload.Title = "Different intent"
	if _, _, err := database.EnqueueEffect(ctx, divergent); !errors.Is(err, store.ErrIdempotencyConflict) {
		t.Fatalf("divergent EnqueueEffect() error = %v; want %v", err, store.ErrIdempotencyConflict)
	}

	claimed, err := database.ClaimEffect(ctx, testID+":worker", time.Minute)
	if err != nil {
		t.Fatalf("ClaimEffect() error = %v", err)
	}
	if claimed.ID != complete.ID || claimed.Attempts != 1 {
		t.Fatalf("ClaimEffect() = %#v", claimed)
	}
	if err := database.RenewEffectLease(ctx, claimed, testID+":worker", time.Minute); err != nil {
		t.Fatalf("RenewEffectLease() error = %v", err)
	}
	completedResult := domain.EffectResult{NativeReference: "issue:1", StatusCode: 201}
	if err := database.CompleteEffect(ctx, claimed, completedResult); err != nil {
		t.Fatalf("CompleteEffect() error = %v", err)
	}
	completed, result, err := database.GetEffect(ctx, complete.ID)
	if err != nil || completed.State != domain.EffectSucceeded || result.NativeReference != "issue:1" {
		t.Fatalf("GetEffect() = %#v, %#v, %v", completed, result, err)
	}

	dead := ticketEffect(testID+":effect:dead", tenant)
	if _, _, err := database.EnqueueEffect(ctx, dead); err != nil {
		t.Fatalf("enqueue dead-letter candidate: %v", err)
	}
	claimed, err = database.ClaimEffect(ctx, testID+":worker", time.Minute)
	if err != nil {
		t.Fatalf("claim dead-letter candidate: %v", err)
	}
	if err := database.DeadLetterEffect(ctx, claimed, "integration failure"); err != nil {
		t.Fatalf("DeadLetterEffect() error = %v", err)
	}
	counts, err := database.Counts(ctx)
	if err != nil || counts.DeadLetters != 1 {
		t.Fatalf("Counts() = %#v, %v; want one dead letter", counts, err)
	}
	if err := database.RequeueDeadLetter(ctx, dead.ID); err != nil {
		t.Fatalf("RequeueDeadLetter() error = %v", err)
	}
	claimed, err = database.ClaimEffect(ctx, testID+":worker", time.Minute)
	if err != nil || claimed.ID != dead.ID {
		t.Fatalf("claim requeued dead letter = %#v, %v", claimed, err)
	}
	if err := database.CompleteEffect(ctx, claimed, domain.EffectResult{StatusCode: 200}); err != nil {
		t.Fatalf("complete requeued dead letter: %v", err)
	}

	expired := ticketEffect(testID+":effect:expired", tenant)
	expired.Target.Repository = "xonovex/expired"
	if _, _, err := database.EnqueueEffect(ctx, expired); err != nil {
		t.Fatalf("enqueue expired-lease candidate: %v", err)
	}
	claimed, err = database.ClaimEffect(ctx, testID+":worker", -time.Second)
	if err != nil {
		t.Fatalf("claim expired-lease candidate: %v", err)
	}
	recovered, err := database.ReleaseExpiredLeases(ctx)
	if err != nil || recovered != 1 {
		t.Fatalf("ReleaseExpiredLeases() = %d, %v; want 1, nil", recovered, err)
	}
	claimed, err = database.ClaimEffect(ctx, testID+":worker", time.Minute)
	if err != nil || claimed.ID != expired.ID {
		t.Fatalf("claim recovered effect = %#v, %v", claimed, err)
	}
	if err := database.CompleteEffect(ctx, claimed, domain.EffectResult{StatusCode: 200}); err != nil {
		t.Fatalf("complete recovered effect: %v", err)
	}

	contextV1 := contextEffect(testID+":context:v1", tenant, 1, "")
	contextV2 := contextEffect(testID+":context:v2", tenant, 2, contextV1.Payload.Context.Digest)
	if _, _, err := database.EnqueueEffect(ctx, contextV2); err != nil {
		t.Fatalf("enqueue context v2: %v", err)
	}
	claimed, err = database.ClaimEffect(ctx, testID+":worker", time.Minute)
	if err != nil {
		t.Fatalf("claim context v2: %v", err)
	}
	if err := database.ValidateContextSequence(ctx, claimed); !errors.Is(err, store.ErrContextConflict) {
		t.Fatalf("ValidateContextSequence(v2 without v1) = %v", err)
	}
	if err := database.DeadLetterEffect(ctx, claimed, "missing predecessor"); err != nil {
		t.Fatalf("dead-letter context v2: %v", err)
	}

	if _, _, err := database.EnqueueEffect(ctx, contextV1); err != nil {
		t.Fatalf("enqueue context v1: %v", err)
	}
	claimed, err = database.ClaimEffect(ctx, testID+":worker", time.Minute)
	if err != nil {
		t.Fatalf("claim context v1: %v", err)
	}
	if err := database.ValidateContextSequence(ctx, claimed); err != nil {
		t.Fatalf("ValidateContextSequence(v1) error = %v", err)
	}
	if err := database.CompleteEffect(ctx, claimed, domain.EffectResult{
		NativeReference: "comment:1",
		StatusCode:      201,
	}); err != nil {
		t.Fatalf("complete context v1: %v", err)
	}
	if err := database.RequeueDeadLetter(ctx, contextV2.ID); err != nil {
		t.Fatalf("requeue context v2: %v", err)
	}
	claimed, err = database.ClaimEffect(ctx, testID+":worker", time.Minute)
	if err != nil {
		t.Fatalf("reclaim context v2: %v", err)
	}
	if err := database.ValidateContextSequence(ctx, claimed); err != nil {
		t.Fatalf("ValidateContextSequence(v2) error = %v", err)
	}
	if err := database.CompleteEffect(ctx, claimed, domain.EffectResult{
		NativeReference: "comment:2",
		StatusCode:      201,
	}); err != nil {
		t.Fatalf("complete context v2: %v", err)
	}

	var auditCount int
	if err := database.pool.QueryRow(ctx, `
		SELECT count(*) FROM dispatcher_audit_events WHERE correlation_id=ANY($1)`,
		correlationIDs,
	).Scan(&auditCount); err != nil {
		t.Fatalf("read audit count: %v", err)
	}
	if auditCount < 12 {
		t.Fatalf("audit count = %d; want at least 12", auditCount)
	}
}

func ticketEffect(id string, tenant string) domain.Effect {
	now := time.Now().UTC()
	return domain.Effect{
		ID:             id,
		CorrelationID:  id,
		IdempotencyKey: id,
		Provider:       domain.ProviderGitHub,
		Tenant:         tenant,
		Kind:           domain.EffectTicketCreate,
		Mode:           domain.EffectModePreview,
		Target:         domain.EffectTarget{Repository: "xonovex/integration"},
		Payload:        domain.EffectPayload{Title: "Integration test"},
		State:          domain.EffectQueued,
		NextAttemptAt:  now.Add(-time.Second),
		CreatedAt:      now,
	}
}

func contextEffect(id string, tenant string, version int, supersedes string) domain.Effect {
	record := domain.ContextRecord{
		ID:           "decision.integration",
		Version:      version,
		Supersedes:   supersedes,
		Type:         "decision",
		Summary:      "Use the durable dispatcher",
		Rationale:    "Provider writes require replay-safe state",
		Alternatives: "Session-local writes",
		Tradeoffs:    "Requires PostgreSQL",
		AppliesTo:    "dispatcher",
		Source:       "integration-test",
		Status:       "accepted",
		Audience:     "engineering",
		Visibility:   "provider",
	}
	record.Digest = domain.Digest(record.SemanticPayload())
	now := time.Now().UTC()
	return domain.Effect{
		ID:             id,
		CorrelationID:  id,
		IdempotencyKey: id,
		Provider:       domain.ProviderGitHub,
		Tenant:         tenant,
		Kind:           domain.EffectContextPublish,
		Mode:           domain.EffectModePreview,
		Target: domain.EffectTarget{
			Repository:  "xonovex/integration",
			SubjectKind: "issue",
			Number:      1,
		},
		Payload:       domain.EffectPayload{Context: &record},
		State:         domain.EffectQueued,
		NextAttemptAt: now.Add(-time.Second),
		CreatedAt:     now,
	}
}
