# cte-patterns: Common Table Expression Patterns

Decompose complex queries into named CTEs (`WITH`) instead of nested subqueries; name them descriptively (`active_users`, not `cte1`). Recursive CTEs use `WITH RECURSIVE` with a base case `UNION ALL` a recursive case that references the CTE.

```sql
WITH RECURSIVE org_tree AS (
    SELECT id, name, manager_id, 1 AS level FROM employees WHERE manager_id IS NULL
    UNION ALL
    SELECT e.id, e.name, e.manager_id, ot.level + 1
    FROM employees e JOIN org_tree ot ON e.manager_id = ot.id
)
SELECT * FROM org_tree ORDER BY level, name;
```
