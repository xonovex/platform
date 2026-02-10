# performance: Query Performance Optimization

**Guideline:** Use EXPLAIN ANALYZE to understand query execution plans. Avoid N+1 queries by using JOINs. Choose appropriate join types. Limit result sets and use efficient pagination strategies.

**Rationale:** Query performance directly impacts application responsiveness and database load. Understanding query plans helps identify bottlenecks. Proper JOIN usage eliminates multiple round-trips. Keyset pagination outperforms OFFSET for large datasets. Limiting result sets reduces memory usage and network transfer.

**Example:**

```sql
-- EXPLAIN ANALYZE to understand query performance
EXPLAIN ANALYZE
SELECT u.email, count(o.id) AS order_count
FROM users u
LEFT JOIN orders o ON o.user_id = u.id
GROUP BY u.id, u.email
HAVING count(o.id) > 10;

-- Use JOINs instead of N+1 queries
SELECT
    u.*,
    json_agg(o.*) AS orders
FROM users u
LEFT JOIN orders o ON o.user_id = u.id
GROUP BY u.id;

-- Limit result sets with WHERE and LIMIT
SELECT * FROM large_table
WHERE created_at >= now() - interval '7 days'
ORDER BY created_at DESC
LIMIT 100;

-- Keyset pagination (cursor-based) for large datasets
SELECT * FROM products
WHERE created_at < $last_seen_timestamp
ORDER BY created_at DESC
LIMIT 20;
```

**Techniques:**

- Use `EXPLAIN ANALYZE` to inspect query execution plans
- Look for sequential scans where indexes should be used
- Replace N+1 query patterns with JOINs or aggregation
- Use appropriate JOIN types (INNER, LEFT, RIGHT, FULL OUTER)
- Always limit result sets with WHERE clauses and LIMIT
- Use keyset (cursor-based) pagination instead of OFFSET for large datasets
- Aggregate related data with JSON functions to reduce round-trips
