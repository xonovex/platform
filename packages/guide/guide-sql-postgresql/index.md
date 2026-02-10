---
name: sql-postgresql-guidelines
description: Trigger on `.sql` files, PostgreSQL migrations, schema files. Use when writing PostgreSQL 15+ queries and schemas. Apply for CTEs, indexing, constraints, JSONB, RLS, performance optimization. Keywords: PostgreSQL, SQL, CTEs, GIN index, JSONB, RLS, role-based access, migrations, EXPLAIN ANALYZE, constraints.
---

# PostgreSQL Coding Guidelines

## Requirements

- PostgreSQL ≥ 15.

## Essentials

- **Query composition** - Use CTEs to decompose complex queries, name descriptively, see [reference/cte-patterns.md](reference/cte-patterns.md)
- **Schema design** - Choose precise types, index for access paths, avoid over-indexing, see [reference/data-types.md](reference/data-types.md), [reference/indexing.md](reference/indexing.md)
- **Data integrity** - Enforce with PK/FK/UNIQUE/CHECK/NOT NULL constraints, see [reference/constraints.md](reference/constraints.md)
- **Semi-structured data** - Use JSONB with GIN indexes, see [reference/jsonb.md](reference/jsonb.md)
- **Multi-tenancy** - Apply RLS and role-based access for data isolation, see [reference/row-level-security.md](reference/row-level-security.md), [reference/role-based-access.md](reference/role-based-access.md)
- **Performance** - Analyze query plans and tune, see [reference/performance.md](reference/performance.md)

## Progressive disclosure

- Read [reference/cte-patterns.md](reference/cte-patterns.md) - When breaking down complex queries or improving readability
- Read [reference/data-types.md](reference/data-types.md) - When choosing column types or avoiding type mismatches
- Read [reference/indexing.md](reference/indexing.md) - When optimizing slow queries or query planning
- Read [reference/constraints.md](reference/constraints.md) - When enforcing data integrity rules
- Read [reference/jsonb.md](reference/jsonb.md) - When storing semi-structured or dynamic data
- Read [reference/row-level-security.md](reference/row-level-security.md) - When implementing multi-tenant data isolation
- Read [reference/role-based-access.md](reference/role-based-access.md) - When configuring database user permissions
- Read [reference/performance.md](reference/performance.md) - When analyzing query plans or tuning performance
