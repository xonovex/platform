# cte-patterns: Common Table Expression Patterns

**Guideline:** Use CTEs (WITH clauses) to decompose complex queries into readable, named components. Prefer CTEs over subqueries for better organization.

**Rationale:** CTEs improve readability by naming intermediate result sets, making complex queries easier to understand and maintain.

**Example:**

```sql
-- Simple CTE
WITH active_users AS (
    SELECT id, email, name FROM users
    WHERE status = 'active' AND deleted_at IS NULL
)
SELECT * FROM active_users WHERE email LIKE '%@example.com' ORDER BY name;

-- Multiple CTEs chained
WITH active_users AS (SELECT id FROM users WHERE status = 'active'),
orders_last_30d AS (
    SELECT user_id, count(*) AS order_count, sum(total) AS total_amount
    FROM orders WHERE created_at >= now() - interval '30 days'
    GROUP BY user_id
)
SELECT u.id, u.email, coalesce(o.order_count, 0) AS orders
FROM active_users u LEFT JOIN orders_last_30d o ON o.user_id = u.id
ORDER BY orders DESC;

-- Recursive CTE for hierarchies
WITH RECURSIVE org_tree AS (
    SELECT id, name, manager_id, 1 AS level FROM employees WHERE manager_id IS NULL
    UNION ALL
    SELECT e.id, e.name, e.manager_id, ot.level + 1
    FROM employees e INNER JOIN org_tree ot ON e.manager_id = ot.id
)
SELECT * FROM org_tree ORDER BY level, name;
```

**Techniques:**
- Name CTEs descriptively (active_users, not cte1)
- Use CTEs to decompose multi-step logic
- Reference CTEs multiple times to avoid repetition
- Use UNION ALL for recursive base + recursive case
