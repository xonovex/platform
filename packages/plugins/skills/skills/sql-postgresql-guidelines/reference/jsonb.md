# jsonb: JSONB Storage and Querying

**Guideline:** Use JSONB (not JSON) for semi-structured data that needs to be queried or indexed. Leverage JSONB operators and functions for efficient querying. Index JSONB columns with GIN indexes.

**Rationale:** JSONB provides flexible schema storage while maintaining queryability. It stores data in binary format for faster processing, supports indexing, and provides rich operators for querying nested structures. Prefer JSONB over JSON for all use cases except when preserving exact text formatting is critical.

**Example:**

```sql
-- Define JSONB table
CREATE TABLE events (
    id UUID PRIMARY KEY,
    type TEXT NOT NULL,
    data JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Query with JSONB operators
SELECT data->'user'->>'name', data->>'email' FROM events;
SELECT * FROM events WHERE data @> '{"status": "completed"}';
SELECT * FROM events WHERE data ? 'user_id';
SELECT * FROM events WHERE data ?| array['email', 'phone'];

-- Build JSONB objects
SELECT jsonb_build_object(
    'id', id,
    'type', type,
    'timestamp', created_at
) FROM events;

-- Update JSONB fields
UPDATE events
SET data = jsonb_set(data, '{status}', '"completed"')
WHERE id = $1;

UPDATE events
SET data = data - 'temp_field'
WHERE type = 'cleanup';

-- Aggregate to JSONB array
SELECT jsonb_agg(data) FROM events WHERE type = 'user_action';

-- GIN index for fast JSONB queries
CREATE INDEX idx_events_data ON events USING GIN (data);
```

**Techniques:**

- Use JSONB column type for semi-structured data
- Use `->` operator to extract JSONB values (returns JSONB)
- Use `->>` operator to extract text values (returns TEXT)
- Use `@>` operator to check containment
- Use `?` operator to check key existence
- Create GIN indexes on JSONB columns for fast querying
- Use JSONB functions (jsonb_build_object, jsonb_agg, jsonb_set) for manipulation
