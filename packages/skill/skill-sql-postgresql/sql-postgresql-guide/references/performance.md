# performance: Query Performance Optimization

Inspect plans with `EXPLAIN ANALYZE`; a sequential scan where an index should apply is the signal to fix. Replace N+1 round-trips with a JOIN + aggregation (`json_agg`).

```sql
EXPLAIN ANALYZE
SELECT u.email, count(o.id) AS order_count
FROM users u LEFT JOIN orders o ON o.user_id = u.id
GROUP BY u.id, u.email HAVING count(o.id) > 10;
```
