package postgres

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/xonovex/platform/packages/agent/agent-dispatcher-go/internal/domain"
	"github.com/xonovex/platform/packages/agent/agent-dispatcher-go/internal/store"
)

//go:embed migrations/*.sql
var migrations embed.FS

type Store struct {
	pool *pgxpool.Pool
}

func Open(ctx context.Context, databaseURL string) (*Store, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse PostgreSQL configuration: %w", err)
	}
	config.MaxConns = 20
	config.MinConns = 2
	config.MaxConnIdleTime = 5 * time.Minute
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("open PostgreSQL pool: %w", err)
	}
	return &Store{pool: pool}, nil
}

func (database *Store) Close() {
	database.pool.Close()
}

func (database *Store) Ping(ctx context.Context) error {
	return database.pool.Ping(ctx)
}

func (database *Store) Migrate(ctx context.Context) error {
	content, err := migrations.ReadFile("migrations/001_dispatcher.sql")
	if err != nil {
		return fmt.Errorf("read dispatcher migration: %w", err)
	}
	if _, err := database.pool.Exec(ctx, string(content)); err != nil {
		return fmt.Errorf("apply dispatcher migration: %w", err)
	}
	return nil
}

func (database *Store) Ingest(
	ctx context.Context,
	delivery domain.Delivery,
	event domain.WorkflowEvent,
) (bool, error) {
	headers, err := json.Marshal(delivery.Headers)
	if err != nil {
		return false, fmt.Errorf("encode delivery headers: %w", err)
	}
	transaction, err := database.pool.Begin(ctx)
	if err != nil {
		return false, fmt.Errorf("begin delivery transaction: %w", err)
	}
	defer func() {
		_ = transaction.Rollback(ctx)
	}()

	tag, err := transaction.Exec(ctx, `
		INSERT INTO dispatcher_deliveries (
			provider, tenant, delivery_id, event_uuid, event, action, payload_digest,
			payload, headers, state, received_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		ON CONFLICT DO NOTHING`,
		delivery.Provider, delivery.Tenant, delivery.DeliveryID, delivery.EventUUID,
		delivery.Event, delivery.Action, delivery.PayloadDigest, delivery.Payload,
		headers, delivery.State, delivery.ReceivedAt,
	)
	if err != nil {
		return false, fmt.Errorf("insert delivery: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return false, nil
	}
	_, err = transaction.Exec(ctx, `
		INSERT INTO dispatcher_workflow_events (
			provider, tenant, delivery_id, event_uuid, kind, action, repository,
			subject_kind, subject_id, subject_number, revision, actor, suppressed, payload
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`,
		event.Provider, event.Tenant, event.DeliveryID, event.EventUUID, event.Kind,
		event.Action, event.Repository, event.SubjectKind, event.SubjectID,
		event.SubjectNumber, event.Revision, event.Actor, event.Suppressed, event.Payload,
	)
	if err != nil {
		return false, fmt.Errorf("insert workflow event: %w", err)
	}
	if err := audit(ctx, transaction, event.DeliveryID, "delivery.ingested", event.Repository, map[string]any{
		"provider":       event.Provider,
		"tenant":         event.Tenant,
		"kind":           event.Kind,
		"action":         event.Action,
		"subject_kind":   event.SubjectKind,
		"subject_id":     event.SubjectID,
		"subject_number": event.SubjectNumber,
		"revision":       event.Revision,
		"actor":          event.Actor,
		"suppressed":     event.Suppressed,
	}); err != nil {
		return false, err
	}
	if err := transaction.Commit(ctx); err != nil {
		return false, fmt.Errorf("commit delivery transaction: %w", err)
	}
	return true, nil
}

func (database *Store) EnqueueEffect(
	ctx context.Context,
	effect domain.Effect,
) (domain.Effect, bool, error) {
	encoded, err := json.Marshal(effect)
	if err != nil {
		return domain.Effect{}, false, fmt.Errorf("encode effect: %w", err)
	}
	transaction, err := database.pool.Begin(ctx)
	if err != nil {
		return domain.Effect{}, false, fmt.Errorf("begin effect enqueue: %w", err)
	}
	defer func() {
		_ = transaction.Rollback(ctx)
	}()
	tag, err := transaction.Exec(ctx, `
		INSERT INTO dispatcher_effects (
			id, correlation_id, idempotency_key, provider, tenant, kind, mode,
			resource_key, effect, state, attempts, next_attempt_at, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,0,$11,$12,$12)
		ON CONFLICT (provider, tenant, idempotency_key) DO NOTHING`,
		effect.ID, effect.CorrelationID, effect.IdempotencyKey, effect.Provider,
		effect.Tenant, effect.Kind, effect.Mode, effect.ResourceKey(), encoded,
		domain.EffectQueued, effect.NextAttemptAt, effect.CreatedAt,
	)
	if err != nil {
		return domain.Effect{}, false, fmt.Errorf("enqueue effect: %w", err)
	}
	if tag.RowsAffected() == 1 {
		if err := audit(ctx, transaction, effect.CorrelationID, "effect.queued", effect.ID, map[string]any{
			"provider": effect.Provider,
			"tenant":   effect.Tenant,
			"kind":     effect.Kind,
			"mode":     effect.Mode,
			"target":   effect.Target,
		}); err != nil {
			return domain.Effect{}, false, err
		}
		if err := transaction.Commit(ctx); err != nil {
			return domain.Effect{}, false, fmt.Errorf("commit effect enqueue: %w", err)
		}
		return effect, true, nil
	}
	if err := transaction.Rollback(ctx); err != nil {
		return domain.Effect{}, false, fmt.Errorf("rollback duplicate effect enqueue: %w", err)
	}
	existing, _, err := database.getEffectByIdempotency(ctx, effect.Provider, effect.Tenant, effect.IdempotencyKey)
	if err != nil {
		return domain.Effect{}, false, err
	}
	if !sameEffectIntent(existing, effect) {
		return domain.Effect{}, false, store.ErrIdempotencyConflict
	}
	return existing, false, nil
}

func (database *Store) GetEffect(ctx context.Context, id string) (domain.Effect, domain.EffectResult, error) {
	return database.getEffect(ctx, `
		SELECT effect, state, attempts, next_attempt_at, result
		FROM dispatcher_effects WHERE id=$1`, id)
}

func (database *Store) getEffectByIdempotency(
	ctx context.Context,
	provider domain.Provider,
	tenant string,
	key string,
) (domain.Effect, domain.EffectResult, error) {
	return database.getEffect(ctx, `
		SELECT effect, state, attempts, next_attempt_at, result
		FROM dispatcher_effects WHERE provider=$1 AND tenant=$2 AND idempotency_key=$3`,
		provider, tenant, key,
	)
}

func (database *Store) getEffect(
	ctx context.Context,
	query string,
	arguments ...any,
) (domain.Effect, domain.EffectResult, error) {
	var encodedEffect []byte
	var encodedResult []byte
	var state domain.EffectState
	var attempts int
	var nextAttemptAt time.Time
	err := database.pool.QueryRow(ctx, query, arguments...).Scan(
		&encodedEffect, &state, &attempts, &nextAttemptAt, &encodedResult,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Effect{}, domain.EffectResult{}, store.ErrNotFound
	}
	if err != nil {
		return domain.Effect{}, domain.EffectResult{}, fmt.Errorf("read effect: %w", err)
	}
	var effect domain.Effect
	if err := json.Unmarshal(encodedEffect, &effect); err != nil {
		return domain.Effect{}, domain.EffectResult{}, fmt.Errorf("decode effect: %w", err)
	}
	effect.State = state
	effect.Attempts = attempts
	effect.NextAttemptAt = nextAttemptAt
	var result domain.EffectResult
	if len(encodedResult) > 0 {
		if err := json.Unmarshal(encodedResult, &result); err != nil {
			return domain.Effect{}, domain.EffectResult{}, fmt.Errorf("decode effect result: %w", err)
		}
	}
	return effect, result, nil
}

func (database *Store) ClaimEffect(
	ctx context.Context,
	holder string,
	leaseDuration time.Duration,
) (domain.Effect, error) {
	transaction, err := database.pool.Begin(ctx)
	if err != nil {
		return domain.Effect{}, fmt.Errorf("begin effect claim: %w", err)
	}
	defer func() {
		_ = transaction.Rollback(ctx)
	}()
	var id string
	var resourceKey string
	var encoded []byte
	err = transaction.QueryRow(ctx, `
		SELECT effect.id, effect.resource_key, effect.effect
		FROM dispatcher_effects AS effect
		WHERE effect.state IN ('queued','retry')
			AND effect.next_attempt_at <= now()
			AND NOT EXISTS (
				SELECT 1
				FROM dispatcher_resource_leases AS lease
				WHERE lease.resource_key=effect.resource_key AND lease.expires_at > now()
			)
		ORDER BY effect.next_attempt_at, effect.created_at
		FOR UPDATE SKIP LOCKED
		LIMIT 1`).Scan(&id, &resourceKey, &encoded)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Effect{}, store.ErrNotFound
	}
	if err != nil {
		return domain.Effect{}, fmt.Errorf("select ready effect: %w", err)
	}
	tag, err := transaction.Exec(ctx, `
		INSERT INTO dispatcher_resource_leases (resource_key, holder, expires_at)
		VALUES ($1,$2,$3)
		ON CONFLICT (resource_key) DO UPDATE
		SET holder=EXCLUDED.holder, expires_at=EXCLUDED.expires_at
		WHERE dispatcher_resource_leases.expires_at <= now()`,
		resourceKey, holder, time.Now().UTC().Add(leaseDuration),
	)
	if err != nil {
		return domain.Effect{}, fmt.Errorf("acquire resource lease: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.Effect{}, store.ErrNotFound
	}
	_, err = transaction.Exec(ctx, `
		UPDATE dispatcher_effects
		SET state='processing', attempts=attempts+1, updated_at=now()
		WHERE id=$1`, id)
	if err != nil {
		return domain.Effect{}, fmt.Errorf("mark effect processing: %w", err)
	}
	var effect domain.Effect
	if err := json.Unmarshal(encoded, &effect); err != nil {
		return domain.Effect{}, fmt.Errorf("decode claimed effect: %w", err)
	}
	effect.State = domain.EffectProcessing
	effect.Attempts++
	if err := audit(ctx, transaction, effect.CorrelationID, "effect.claimed", effect.ID, map[string]any{
		"attempt": effect.Attempts,
		"holder":  holder,
	}); err != nil {
		return domain.Effect{}, err
	}
	if err := transaction.Commit(ctx); err != nil {
		return domain.Effect{}, fmt.Errorf("commit effect claim: %w", err)
	}
	return effect, nil
}

func (database *Store) RenewEffectLease(
	ctx context.Context,
	effect domain.Effect,
	holder string,
	leaseDuration time.Duration,
) error {
	tag, err := database.pool.Exec(ctx, `
		UPDATE dispatcher_resource_leases
		SET expires_at=$3
		WHERE resource_key=$1 AND holder=$2 AND expires_at > now()`,
		effect.ResourceKey(), holder, time.Now().UTC().Add(leaseDuration),
	)
	if err != nil {
		return fmt.Errorf("renew effect lease: %w", err)
	}
	if tag.RowsAffected() != 1 {
		return store.ErrNotFound
	}
	return nil
}

func (database *Store) ValidateContextSequence(ctx context.Context, effect domain.Effect) error {
	if effect.Kind != domain.EffectContextPublish || effect.Payload.Context == nil {
		return nil
	}
	record := effect.Payload.Context
	var existingDigest string
	err := database.pool.QueryRow(ctx, `
		SELECT digest
		FROM dispatcher_context_records
		WHERE provider=$1 AND tenant=$2 AND repository=$3 AND destination_number=$4
			AND context_id=$5 AND version=$6`,
		effect.Provider, effect.Tenant, effect.Target.Repository, effect.Target.Number,
		record.ID, record.Version,
	).Scan(&existingDigest)
	if err == nil {
		if existingDigest != record.Digest {
			return store.ErrContextConflict
		}
		return nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("read current context sequence: %w", err)
	}
	if record.Version == 1 {
		return nil
	}
	var latestVersion int
	err = database.pool.QueryRow(ctx, `
		SELECT version
		FROM dispatcher_context_records
		WHERE provider=$1 AND tenant=$2 AND repository=$3 AND destination_number=$4
			AND context_id=$5
		ORDER BY version DESC
		LIMIT 1`,
		effect.Provider, effect.Tenant, effect.Target.Repository, effect.Target.Number, record.ID,
	).Scan(&latestVersion)
	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("%w: context version %d requires stored predecessor version %d",
			store.ErrContextConflict, record.Version, record.Version-1)
	}
	if err != nil {
		return fmt.Errorf("read prior context sequence: %w", err)
	}
	if latestVersion != record.Version-1 {
		return fmt.Errorf("%w: context version %d requires stored predecessor version %d",
			store.ErrContextConflict, record.Version, record.Version-1)
	}
	return nil
}

func (database *Store) CompleteEffect(
	ctx context.Context,
	effect domain.Effect,
	result domain.EffectResult,
) error {
	encodedResult, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("encode effect result: %w", err)
	}
	transaction, err := database.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin effect completion: %w", err)
	}
	defer func() {
		_ = transaction.Rollback(ctx)
	}()
	if effect.Kind == domain.EffectContextPublish && effect.Payload.Context != nil {
		record, err := json.Marshal(effect.Payload.Context)
		if err != nil {
			return fmt.Errorf("encode context record: %w", err)
		}
		tag, err := transaction.Exec(ctx, `
			INSERT INTO dispatcher_context_records (
				provider, tenant, repository, destination_number, context_id,
				version, digest, record, native_reference
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
			ON CONFLICT DO NOTHING`,
			effect.Provider, effect.Tenant, effect.Target.Repository, effect.Target.Number,
			effect.Payload.Context.ID, effect.Payload.Context.Version, effect.Payload.Context.Digest,
			record, result.NativeReference,
		)
		if err != nil {
			return fmt.Errorf("store context record: %w", err)
		}
		if tag.RowsAffected() == 0 {
			var digest string
			err := transaction.QueryRow(ctx, `
				SELECT digest FROM dispatcher_context_records
				WHERE provider=$1 AND tenant=$2 AND repository=$3 AND destination_number=$4
					AND context_id=$5 AND version=$6`,
				effect.Provider, effect.Tenant, effect.Target.Repository, effect.Target.Number,
				effect.Payload.Context.ID, effect.Payload.Context.Version,
			).Scan(&digest)
			if err != nil {
				return fmt.Errorf("read context conflict: %w", err)
			}
			if digest != effect.Payload.Context.Digest {
				return store.ErrContextConflict
			}
		}
	}
	tag, err := transaction.Exec(ctx, `
		UPDATE dispatcher_effects
		SET state='succeeded', result=$2, last_error='', updated_at=now()
		WHERE id=$1 AND state='processing'`,
		effect.ID, encodedResult,
	)
	if err != nil {
		return fmt.Errorf("complete effect: %w", err)
	}
	if tag.RowsAffected() != 1 {
		return store.ErrNotFound
	}
	if _, err := transaction.Exec(ctx, `DELETE FROM dispatcher_resource_leases WHERE resource_key=$1`, effect.ResourceKey()); err != nil {
		return fmt.Errorf("release completed effect lease: %w", err)
	}
	if err := audit(ctx, transaction, effect.CorrelationID, "effect.succeeded", effect.ID, result); err != nil {
		return err
	}
	if err := transaction.Commit(ctx); err != nil {
		return fmt.Errorf("commit effect completion: %w", err)
	}
	return nil
}

func (database *Store) RetryEffect(
	ctx context.Context,
	effect domain.Effect,
	reason string,
	nextAttemptAt time.Time,
) error {
	return database.transitionFailure(ctx, effect, domain.EffectRetry, reason, nextAttemptAt)
}

func (database *Store) DeadLetterEffect(ctx context.Context, effect domain.Effect, reason string) error {
	transaction, err := database.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin dead-letter transition: %w", err)
	}
	defer func() {
		_ = transaction.Rollback(ctx)
	}()
	if err := transitionFailure(ctx, transaction, effect, domain.EffectDeadLetter, reason, time.Now().UTC()); err != nil {
		return err
	}
	encoded, err := json.Marshal(effect)
	if err != nil {
		return fmt.Errorf("encode dead-letter effect: %w", err)
	}
	_, err = transaction.Exec(ctx, `
		INSERT INTO dispatcher_dead_letters (effect_id, reason, effect)
		VALUES ($1,$2,$3)
		ON CONFLICT (effect_id) DO UPDATE SET reason=EXCLUDED.reason, failed_at=now()`,
		effect.ID, reason, encoded,
	)
	if err != nil {
		return fmt.Errorf("store dead letter: %w", err)
	}
	if err := audit(ctx, transaction, effect.CorrelationID, "effect.dead_lettered", effect.ID, map[string]string{
		"reason": reason,
	}); err != nil {
		return err
	}
	if err := transaction.Commit(ctx); err != nil {
		return fmt.Errorf("commit dead-letter transition: %w", err)
	}
	return nil
}

func (database *Store) transitionFailure(
	ctx context.Context,
	effect domain.Effect,
	state domain.EffectState,
	reason string,
	nextAttemptAt time.Time,
) error {
	transaction, err := database.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin effect failure transition: %w", err)
	}
	defer func() {
		_ = transaction.Rollback(ctx)
	}()
	if err := transitionFailure(ctx, transaction, effect, state, reason, nextAttemptAt); err != nil {
		return err
	}
	if err := audit(ctx, transaction, effect.CorrelationID, "effect.retry_scheduled", effect.ID, map[string]any{
		"reason":          reason,
		"next_attempt_at": nextAttemptAt,
	}); err != nil {
		return err
	}
	if err := transaction.Commit(ctx); err != nil {
		return fmt.Errorf("commit effect failure transition: %w", err)
	}
	return nil
}

func transitionFailure(
	ctx context.Context,
	transaction pgx.Tx,
	effect domain.Effect,
	state domain.EffectState,
	reason string,
	nextAttemptAt time.Time,
) error {
	tag, err := transaction.Exec(ctx, `
		UPDATE dispatcher_effects
		SET state=$2, last_error=$3, next_attempt_at=$4, updated_at=now()
		WHERE id=$1 AND state='processing'`,
		effect.ID, state, reason, nextAttemptAt,
	)
	if err != nil {
		return fmt.Errorf("transition failed effect: %w", err)
	}
	if tag.RowsAffected() != 1 {
		return store.ErrNotFound
	}
	if _, err := transaction.Exec(ctx, `DELETE FROM dispatcher_resource_leases WHERE resource_key=$1`, effect.ResourceKey()); err != nil {
		return fmt.Errorf("release failed effect lease: %w", err)
	}
	return nil
}

func (database *Store) ReleaseExpiredLeases(ctx context.Context) (int64, error) {
	transaction, err := database.pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("begin lease recovery: %w", err)
	}
	defer func() {
		_ = transaction.Rollback(ctx)
	}()
	tag, err := transaction.Exec(ctx, `
		UPDATE dispatcher_effects
		SET state='retry', next_attempt_at=now(), last_error='worker lease expired', updated_at=now()
		WHERE state='processing' AND resource_key IN (
			SELECT resource_key FROM dispatcher_resource_leases WHERE expires_at <= now()
		)`)
	if err != nil {
		return 0, fmt.Errorf("recover expired effects: %w", err)
	}
	if _, err := transaction.Exec(ctx, `DELETE FROM dispatcher_resource_leases WHERE expires_at <= now()`); err != nil {
		return 0, fmt.Errorf("delete expired leases: %w", err)
	}
	if err := transaction.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit lease recovery: %w", err)
	}
	return tag.RowsAffected(), nil
}

func (database *Store) RequeueDeadLetter(ctx context.Context, effectID string) error {
	transaction, err := database.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin dead-letter requeue: %w", err)
	}
	defer func() {
		_ = transaction.Rollback(ctx)
	}()
	tag, err := transaction.Exec(ctx, `
		UPDATE dispatcher_effects
		SET state='retry', attempts=0, next_attempt_at=now(), last_error='', updated_at=now()
		WHERE id=$1 AND state='dead_letter'`, effectID)
	if err != nil {
		return fmt.Errorf("requeue dead letter: %w", err)
	}
	if tag.RowsAffected() != 1 {
		return store.ErrNotFound
	}
	if _, err := transaction.Exec(ctx, `
		UPDATE dispatcher_dead_letters SET requeued_at=now() WHERE effect_id=$1`, effectID); err != nil {
		return fmt.Errorf("mark dead letter requeued: %w", err)
	}
	if err := transaction.Commit(ctx); err != nil {
		return fmt.Errorf("commit dead-letter requeue: %w", err)
	}
	return nil
}

func (database *Store) Counts(ctx context.Context) (store.Counts, error) {
	var counts store.Counts
	err := database.pool.QueryRow(ctx, `
		SELECT
			(SELECT count(*) FROM dispatcher_deliveries WHERE state='pending'),
			(SELECT count(*) FROM dispatcher_effects WHERE state='queued'),
			(SELECT count(*) FROM dispatcher_effects WHERE state='retry'),
			(SELECT count(*) FROM dispatcher_effects WHERE state='dead_letter')`).Scan(
		&counts.PendingDeliveries,
		&counts.QueuedEffects,
		&counts.RetryEffects,
		&counts.DeadLetters,
	)
	if err != nil {
		return store.Counts{}, fmt.Errorf("read dispatcher counts: %w", err)
	}
	return counts, nil
}

func audit(
	ctx context.Context,
	transaction pgx.Tx,
	correlationID string,
	eventType string,
	subject string,
	details any,
) error {
	encoded, err := json.Marshal(details)
	if err != nil {
		return fmt.Errorf("encode audit details: %w", err)
	}
	if _, err := transaction.Exec(ctx, `
		INSERT INTO dispatcher_audit_events (correlation_id, event_type, subject, details)
		VALUES ($1,$2,$3,$4)`, correlationID, eventType, subject, encoded); err != nil {
		return fmt.Errorf("insert audit event: %w", err)
	}
	return nil
}

func sameEffectIntent(left domain.Effect, right domain.Effect) bool {
	return left.IdempotencyKey == right.IdempotencyKey &&
		left.Provider == right.Provider &&
		left.Tenant == right.Tenant &&
		left.Kind == right.Kind &&
		left.Mode == right.Mode &&
		reflect.DeepEqual(left.Target, right.Target) &&
		reflect.DeepEqual(left.Payload, right.Payload) &&
		reflect.DeepEqual(left.Preconditions, right.Preconditions)
}
