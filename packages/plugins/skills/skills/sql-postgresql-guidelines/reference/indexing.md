# indexing: Index Strategy and Optimization

**Guideline:** Create indexes based on actual query patterns and access paths. Use appropriate index types (B-tree, GIN, GiST, BRIN) for different data types and query patterns. Avoid over-indexing as each index has write overhead.

**Rationale:** Proper indexing dramatically improves query performance by reducing full table scans. Different index types are optimized for different use cases: B-tree for equality and range queries, GIN for JSONB and arrays, GiST for full-text search, BRIN for large time-series data.

**Example:**

```sql
-- B-tree index for common queries
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_orders_user_id ON orders(user_id);

-- Composite index for multi-column queries
CREATE INDEX idx_orders_user_created ON orders(user_id, created_at DESC);

-- Partial index for filtered queries
CREATE INDEX idx_active_users_email ON users(email)
WHERE status = 'active' AND deleted_at IS NULL;

-- Expression index for computed values
CREATE INDEX idx_users_lower_email ON users(LOWER(email));

-- GIN index for JSONB
CREATE INDEX idx_products_metadata ON products USING GIN (metadata);
CREATE INDEX idx_products_tags ON products USING GIN (tags);

-- GiST index for full-text search
CREATE INDEX idx_products_search ON products USING GiST (
    to_tsvector('english', name || ' ' || coalesce(description, ''))
);

-- BRIN index for large time-series data
CREATE INDEX idx_logs_created_at ON logs USING BRIN (created_at);
```

**Techniques:**
- Create B-tree indexes (default) for frequently queried columns
- Use partial indexes to index only relevant subsets
- Use composite indexes for multi-column queries
- Use expression indexes for computed values
- Use GIN indexes for JSONB, arrays, and full-text search
- Use GiST indexes for full-text search with ranking
- Use BRIN indexes for large, naturally ordered data (timestamps, IDs)
- Review query plans with EXPLAIN ANALYZE before adding indexes
