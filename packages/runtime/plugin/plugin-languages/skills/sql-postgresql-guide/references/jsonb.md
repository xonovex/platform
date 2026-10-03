# jsonb: JSONB Storage and Querying

Use `JSONB` (not `JSON`) for queryable/indexable semi-structured data; index with GIN. Operators: `->` extracts JSONB, `->>` extracts TEXT, `@>` containment, `?` key exists, `?|` any-key-exists. Mutate with `jsonb_set`, delete a key with `- 'key'`; build/aggregate with `jsonb_build_object` / `jsonb_agg`.

```sql
SELECT data->'user'->>'name' FROM events;
SELECT * FROM events WHERE data @> '{"status": "completed"}';
SELECT * FROM events WHERE data ?| array['email', 'phone'];
UPDATE events SET data = jsonb_set(data, '{status}', '"completed"') WHERE id = $1;
UPDATE events SET data = data - 'temp_field';
CREATE INDEX idx_events_data ON events USING GIN (data);
```
