# indexing: Index Strategy and Optimization

Index for actual access paths; each index adds write cost, so don't over-index. Type selection: B-tree (default) for equality/range; GIN for JSONB, arrays, full-text; GiST for full-text ranking and exclusion constraints; BRIN for large naturally-ordered data (timestamps, append-only ids). Use partial indexes for hot subsets and expression indexes for computed lookups — a plain index on `email` is skipped by `WHERE LOWER(email) = …`.

```sql
CREATE INDEX idx_orders_user_created ON orders(user_id, created_at DESC);  -- composite
CREATE INDEX idx_active_users ON users(email) WHERE status = 'active';     -- partial
CREATE INDEX idx_users_lower_email ON users(LOWER(email));                 -- expression
CREATE INDEX idx_products_metadata ON products USING GIN (metadata);       -- JSONB/array
CREATE INDEX idx_logs_created_at ON logs USING BRIN (created_at);          -- large time-series
```
