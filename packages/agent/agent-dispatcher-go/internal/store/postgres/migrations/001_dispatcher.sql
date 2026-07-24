CREATE TABLE IF NOT EXISTS dispatcher_deliveries (
    provider text NOT NULL,
    tenant text NOT NULL,
    delivery_id text NOT NULL,
    event_uuid text NOT NULL DEFAULT '',
    event text NOT NULL,
    action text NOT NULL DEFAULT '',
    payload_digest text NOT NULL,
    payload jsonb NOT NULL,
    headers jsonb NOT NULL,
    state text NOT NULL,
    received_at timestamptz NOT NULL,
    PRIMARY KEY (provider, tenant, delivery_id)
);

CREATE TABLE IF NOT EXISTS dispatcher_workflow_events (
    id bigserial PRIMARY KEY,
    provider text NOT NULL,
    tenant text NOT NULL,
    delivery_id text NOT NULL,
    event_uuid text NOT NULL DEFAULT '',
    kind text NOT NULL,
    action text NOT NULL DEFAULT '',
    repository text NOT NULL DEFAULT '',
    subject_kind text NOT NULL DEFAULT '',
    subject_id text NOT NULL DEFAULT '',
    subject_number bigint NOT NULL DEFAULT 0,
    revision text NOT NULL DEFAULT '',
    actor text NOT NULL DEFAULT '',
    suppressed boolean NOT NULL DEFAULT false,
    payload jsonb NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (provider, tenant, delivery_id)
);

CREATE TABLE IF NOT EXISTS dispatcher_effects (
    id text PRIMARY KEY,
    correlation_id text NOT NULL,
    idempotency_key text NOT NULL,
    provider text NOT NULL,
    tenant text NOT NULL,
    kind text NOT NULL,
    mode text NOT NULL,
    resource_key text NOT NULL,
    effect jsonb NOT NULL,
    state text NOT NULL,
    attempts integer NOT NULL DEFAULT 0,
    next_attempt_at timestamptz NOT NULL DEFAULT now(),
    last_error text NOT NULL DEFAULT '',
    result jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (provider, tenant, idempotency_key)
);

CREATE INDEX IF NOT EXISTS dispatcher_effects_ready_idx
    ON dispatcher_effects (next_attempt_at, created_at)
    WHERE state IN ('queued', 'retry');

CREATE TABLE IF NOT EXISTS dispatcher_resource_leases (
    resource_key text PRIMARY KEY,
    holder text NOT NULL,
    expires_at timestamptz NOT NULL
);

CREATE TABLE IF NOT EXISTS dispatcher_context_records (
    provider text NOT NULL,
    tenant text NOT NULL,
    repository text NOT NULL,
    destination_number bigint NOT NULL,
    context_id text NOT NULL,
    version integer NOT NULL,
    digest text NOT NULL,
    record jsonb NOT NULL,
    native_reference text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (provider, tenant, repository, destination_number, context_id, version),
    UNIQUE (provider, tenant, repository, destination_number, context_id, version, digest)
);

CREATE TABLE IF NOT EXISTS dispatcher_dead_letters (
    effect_id text PRIMARY KEY REFERENCES dispatcher_effects(id),
    reason text NOT NULL,
    effect jsonb NOT NULL,
    failed_at timestamptz NOT NULL DEFAULT now(),
    requeued_at timestamptz
);

CREATE TABLE IF NOT EXISTS dispatcher_audit_events (
    id bigserial PRIMARY KEY,
    correlation_id text NOT NULL,
    event_type text NOT NULL,
    subject text NOT NULL,
    details jsonb NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS dispatcher_audit_correlation_idx
    ON dispatcher_audit_events (correlation_id, created_at);
