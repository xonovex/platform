---
name: sql-postgresql-guide
description: "Use when editing PostgreSQL 15+ queries, schemas, or migrations. Triggers on `.sql` files, migration files, schema files, and prompts about CTEs, indexing, JSONB, RLS, EXPLAIN ANALYZE, constraints, or partitioning, even when the user doesn't say 'Postgres'. Skip MySQL, SQLite, MSSQL-specific features, and ORM-only work that doesn't reach raw SQL."
---

# PostgreSQL Coding Guidelines

## Requirements

- PostgreSQL ≥ 15.

## Essentials

- **Query composition** - Use CTEs to decompose complex queries, name descriptively, see [references/cte-patterns.md](references/cte-patterns.md)
- **Schema design** - Choose precise types, index for access paths, avoid over-indexing, see [references/data-types.md](references/data-types.md), [references/indexing.md](references/indexing.md)
- **Data integrity** - Enforce with PK/FK/UNIQUE/CHECK/NOT NULL constraints, see [references/constraints.md](references/constraints.md)
- **Semi-structured data** - Use JSONB with GIN indexes, see [references/jsonb.md](references/jsonb.md)
- **Multi-tenancy** - Apply RLS and role-based access for data isolation, see [references/row-level-security.md](references/row-level-security.md), [references/role-based-access.md](references/role-based-access.md)
- **Performance** - Analyze query plans and tune, see [references/performance.md](references/performance.md)

## Gotchas

- Indexes don't help if a query wraps the column in a function: `WHERE lower(email) = …` misses an index on `email`; create a functional index instead
- `ANALYZE` updates planner statistics — bulk inserts without re-analyzing produce stale plans and full scans
- `JSONB` supports indexing (GIN); `JSON` doesn't — pick JSONB unless you specifically need preserved formatting
- DDL inside transactions is allowed (unlike MySQL) — wrap migrations in `BEGIN`/`COMMIT` for atomicity
- `SERIAL`/`BIGSERIAL` is being deprecated in favor of `GENERATED AS IDENTITY` — same effect, cleaner semantics, no sequence-ownership oddities

## Progressive disclosure

- Read [references/cte-patterns.md](references/cte-patterns.md) - When breaking down complex queries or improving readability
- Read [references/data-types.md](references/data-types.md) - When choosing column types or avoiding type mismatches
- Read [references/indexing.md](references/indexing.md) - When optimizing slow queries or query planning
- Read [references/constraints.md](references/constraints.md) - When enforcing data integrity rules
- Read [references/jsonb.md](references/jsonb.md) - When storing semi-structured or dynamic data
- Read [references/row-level-security.md](references/row-level-security.md) - When implementing multi-tenant data isolation
- Read [references/role-based-access.md](references/role-based-access.md) - When configuring database user permissions
- Read [references/performance.md](references/performance.md) - When analyzing query plans or tuning performance
