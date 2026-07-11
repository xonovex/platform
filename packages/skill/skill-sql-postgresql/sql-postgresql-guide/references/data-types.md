# data-types: PostgreSQL Data Type Best Practices

Pick precise types: `TIMESTAMPTZ` (never `TIMESTAMP`) for timestamps, `NUMERIC` for money/exact decimals (never `FLOAT`), `UUID` (`gen_random_uuid()`) for distributed ids, `JSONB` for semi-structured data, `TEXT[]` for ordered collections. Wrap stable value sets in `CREATE TYPE ... AS ENUM` and reusable constrained types in `CREATE DOMAIN`.

```sql
CREATE TYPE product_status AS ENUM ('draft', 'active', 'archived');
CREATE DOMAIN email AS TEXT CHECK (VALUE ~ '^[^@]+@[^@]+\.[^@]+$');

CREATE TABLE products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    price NUMERIC(10, 2) NOT NULL,                  -- exact, not FLOAT
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),  -- timezone aware
    metadata JSONB,
    tags TEXT[] DEFAULT '{}',
    status product_status NOT NULL
);
```
