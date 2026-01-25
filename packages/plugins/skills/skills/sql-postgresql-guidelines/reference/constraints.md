# constraints: Database Constraint Patterns

**Guideline:** Enforce data integrity at the database level using constraints (PRIMARY KEY, FOREIGN KEY, UNIQUE, CHECK, NOT NULL, EXCLUSION). Define referential actions (ON DELETE/UPDATE) for foreign keys.

**Rationale:** Database constraints ensure data validity regardless of which application or process modifies the data.

**Example:**

```sql
-- Primary key
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT NOT NULL UNIQUE,
    age INTEGER CHECK (age >= 18)
);

-- Foreign key with cascading delete
CREATE TABLE orders (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    total NUMERIC(10, 2) NOT NULL CHECK (total >= 0),
    UNIQUE (user_id, id)
);

-- Composite unique constraint
ALTER TABLE users ADD UNIQUE (email, tenant_id);

-- Exclusion constraint for overlapping ranges
CREATE TABLE bookings (
    id UUID PRIMARY KEY,
    room_id UUID NOT NULL,
    start_time TIMESTAMPTZ NOT NULL,
    end_time TIMESTAMPTZ NOT NULL,
    EXCLUDE USING GIST (
        room_id WITH =,
        tstzrange(start_time, end_time) WITH &&
    )
);
```

**Techniques:**
- Use PRIMARY KEY for unique row identifiers
- Use FOREIGN KEY with appropriate ON DELETE/UPDATE actions
- Use UNIQUE for columns requiring uniqueness
- Use CHECK for value validation and business rules
- Use NOT NULL for required fields
- Use EXCLUSION constraints to prevent overlapping ranges or conflicting data
