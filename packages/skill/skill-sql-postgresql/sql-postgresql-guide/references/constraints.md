# constraints: Database Constraint Patterns

Enforce integrity in-schema: PRIMARY KEY, FOREIGN KEY (with `ON DELETE`/`ON UPDATE` actions), UNIQUE, CHECK, NOT NULL. For "no overlapping ranges" use an EXCLUSION constraint (`EXCLUDE USING GIST`).

```sql
CREATE TABLE orders (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    total NUMERIC(10, 2) NOT NULL CHECK (total >= 0)
);

CREATE TABLE bookings (
    room_id UUID NOT NULL,
    start_time TIMESTAMPTZ NOT NULL,
    end_time TIMESTAMPTZ NOT NULL,
    EXCLUDE USING GIST (room_id WITH =, tstzrange(start_time, end_time) WITH &&)
);
```
