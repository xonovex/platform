# role-based-access: Role-Based Access Control (RBAC)

**Guideline:** Create database roles with appropriate permissions and grant them to users. Use roles to enforce least-privilege access at the database level. Define read-only, read-write, and admin roles.

**Rationale:** Database-level RBAC provides defense-in-depth security by limiting what each database user can do. It prevents accidental or malicious modifications, supports compliance requirements, and allows different applications or users to have different access levels.

**Example:**

```sql
-- Create roles
CREATE ROLE app_readonly;
CREATE ROLE app_readwrite;
CREATE ROLE app_admin;

-- Grant permissions for read-only role
GRANT CONNECT ON DATABASE myapp TO app_readonly;
GRANT USAGE ON SCHEMA public TO app_readonly;
GRANT SELECT ON ALL TABLES IN SCHEMA public TO app_readonly;
ALTER DEFAULT PRIVILEGES IN SCHEMA public
    GRANT SELECT ON TABLES TO app_readonly;

-- Grant permissions for read-write role
GRANT CONNECT ON DATABASE myapp TO app_readwrite;
GRANT USAGE ON SCHEMA public TO app_readwrite;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO app_readwrite;
GRANT USAGE ON ALL SEQUENCES IN SCHEMA public TO app_readwrite;
ALTER DEFAULT PRIVILEGES IN SCHEMA public
    GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO app_readwrite;

-- Grant permissions for admin role
GRANT ALL PRIVILEGES ON DATABASE myapp TO app_admin;

-- Create users and assign roles
CREATE USER api_user WITH PASSWORD 'secure_password';
GRANT app_readwrite TO api_user;

CREATE USER analyst WITH PASSWORD 'secure_password';
GRANT app_readonly TO analyst;
```

**Techniques:**
- Create roles for different access levels (readonly, readwrite, admin)
- Grant minimal necessary permissions to each role
- Use `ALTER DEFAULT PRIVILEGES` to apply permissions to future objects
- Create database users and assign them to appropriate roles
- Use `GRANT` to assign roles to users
- Avoid granting permissions directly to users; use roles instead
