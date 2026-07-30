# row-level-security: Row-Level Security (RLS) Implementation

Enable per table (`ALTER TABLE ... ENABLE ROW LEVEL SECURITY`), then add a policy per operation. `USING` filters which rows are visible/modifiable; `WITH CHECK` validates rows being written (INSERT/UPDATE). Policies read request context via `current_setting('app.tenant_id')`, set per session/transaction before queries.

```sql
ALTER TABLE documents ENABLE ROW LEVEL SECURITY;

CREATE POLICY documents_select ON documents FOR SELECT
    USING (tenant_id = current_setting('app.tenant_id')::UUID);

CREATE POLICY documents_update ON documents FOR UPDATE
    USING (tenant_id = current_setting('app.tenant_id')::UUID
           AND owner_id = current_setting('app.user_id')::UUID)
    WITH CHECK (tenant_id = current_setting('app.tenant_id')::UUID);

SET app.tenant_id = '550e8400-e29b-41d4-a716-446655440000';
SELECT * FROM documents;  -- only this tenant's rows
```
