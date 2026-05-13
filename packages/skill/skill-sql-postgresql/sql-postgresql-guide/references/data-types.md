# data-types: PostgreSQL Data Type Best Practices

**Guideline:** Choose precise, semantically appropriate data types for each column. Use UUID for distributed identifiers, TIMESTAMPTZ for timestamps, NUMERIC for exact decimals, JSONB for semi-structured data, and custom types (ENUMs, DOMAINs) for business logic.

**Rationale:** Proper data types ensure data integrity, optimize storage, enable better indexing, and prevent common errors. TIMESTAMPTZ handles timezones correctly, NUMERIC avoids floating-point precision issues, and custom types enforce domain constraints at the database level.

**Example:**

```sql
-- Precise data types
CREATE TABLE products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    price NUMERIC(10, 2) NOT NULL,  -- Precise decimal, not FLOAT
    quantity INTEGER NOT NULL CHECK (quantity >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),  -- Timezone aware
    metadata JSONB,  -- Semi-structured data
    tags TEXT[] DEFAULT '{}',  -- Array of strings
    status product_status NOT NULL  -- Enum type
);

-- Enum type
CREATE TYPE product_status AS ENUM ('draft', 'active', 'archived');

-- Domain types for validation
CREATE DOMAIN email AS TEXT
CHECK (VALUE ~ '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}$');

CREATE DOMAIN positive_numeric AS NUMERIC(10, 2)
CHECK (VALUE >= 0);

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email email NOT NULL UNIQUE,
    balance positive_numeric NOT NULL DEFAULT 0
);
```

**Techniques:**

- Use UUID for globally unique identifiers in distributed systems
- Use TIMESTAMPTZ (not TIMESTAMP) for all timestamps to preserve timezone information
- Use NUMERIC for monetary values and exact decimals
- Use ENUMs for small, stable sets of values
- Use DOMAINs to create reusable constrained types
- Use JSONB for flexible, semi-structured data
- Use array types (TEXT[], INTEGER[]) for ordered collections
