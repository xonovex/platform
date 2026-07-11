# role-based-access: Role-Based Access Control (RBAC)

Grant privileges to roles (readonly/readwrite/admin), then assign roles to users — never grant directly to users. `ALTER DEFAULT PRIVILEGES` is required so future tables inherit grants; readwrite also needs `USAGE ON … SEQUENCES` for `nextval` on serial/identity columns.

```sql
CREATE ROLE app_readwrite;
GRANT USAGE ON SCHEMA public TO app_readwrite;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO app_readwrite;
GRANT USAGE ON ALL SEQUENCES IN SCHEMA public TO app_readwrite;
ALTER DEFAULT PRIVILEGES IN SCHEMA public
    GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO app_readwrite;

CREATE USER api_user WITH PASSWORD 'secure_password';
GRANT app_readwrite TO api_user;
```
