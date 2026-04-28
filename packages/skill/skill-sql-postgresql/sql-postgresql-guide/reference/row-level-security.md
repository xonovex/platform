# row-level-security: Row-Level Security (RLS) Implementation

**Guideline:** Use Row-Level Security (RLS) policies to enforce tenant isolation and access control at the database level. Set session variables to identify the current tenant/user, and create policies that filter data based on these variables.

**Rationale:** RLS provides transparent, centralized access control that applies regardless of how data is accessed. It prevents accidental data leaks between tenants and reduces application-level security logic. Policies are enforced by PostgreSQL itself, making them immune to application bugs.

**Example:**

```sql
-- Define multi-tenant table
CREATE TABLE documents (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    title TEXT NOT NULL,
    content TEXT,
    owner_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Enable RLS
ALTER TABLE documents ENABLE ROW LEVEL SECURITY;

-- Create policies for each operation
CREATE POLICY documents_select_policy ON documents
FOR SELECT
USING (tenant_id = current_setting('app.tenant_id')::UUID);

CREATE POLICY documents_insert_policy ON documents
FOR INSERT
WITH CHECK (
    tenant_id = current_setting('app.tenant_id')::UUID
    AND owner_id = current_setting('app.user_id')::UUID
);

CREATE POLICY documents_update_policy ON documents
FOR UPDATE
USING (
    tenant_id = current_setting('app.tenant_id')::UUID
    AND owner_id = current_setting('app.user_id')::UUID
)
WITH CHECK (
    tenant_id = current_setting('app.tenant_id')::UUID
);

CREATE POLICY documents_delete_policy ON documents
FOR DELETE
USING (
    tenant_id = current_setting('app.tenant_id')::UUID
    AND owner_id = current_setting('app.user_id')::UUID
);

-- Set session variables before queries
SET app.tenant_id = '550e8400-e29b-41d4-a716-446655440000';
SET app.user_id = '660e8400-e29b-41d4-a716-446655440000';

-- Query automatically filtered by RLS
SELECT * FROM documents;  -- Only sees tenant's documents
```

**Techniques:**

- Enable RLS on multi-tenant tables with `ALTER TABLE ... ENABLE ROW LEVEL SECURITY`
- Create policies for each operation type (SELECT, INSERT, UPDATE, DELETE)
- Use `USING` clause to filter which rows are visible/modifiable
- Use `WITH CHECK` clause to validate new/modified rows
- Set session variables (e.g., `app.tenant_id`, `app.user_id`) before queries
- Use `current_setting()` in policies to access session variables
