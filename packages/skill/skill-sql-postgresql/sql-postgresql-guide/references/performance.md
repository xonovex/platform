# performance: Query Performance Optimization

Inspect plans with `EXPLAIN ANALYZE`; a sequential scan where an index should apply is the signal to fix. Replace N+1 round-trips with a JOIN + aggregation (`json_agg`). For large-dataset pagination use keyset/cursor (`WHERE created_at < $last ORDER BY … LIMIT n`), not `OFFSET`, which still scans and discards skipped rows.

```sql
EXPLAIN ANALYZE
SELECT u.email, count(o.id) AS order_count
FROM users u LEFT JOIN orders o ON o.user_id = u.id
GROUP BY u.id, u.email HAVING count(o.id) > 10;

-- keyset pagination, not OFFSET
SELECT * FROM products WHERE created_at < $last_seen ORDER BY created_at DESC LIMIT 20;
```
