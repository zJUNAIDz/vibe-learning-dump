# Roles, Privileges, and Least Privilege

> **What mistake does this prevent?**
> Running your entire application as the PostgreSQL superuser, which means any SQL injection vulnerability gives an attacker complete control over the database server, OS file access, and every other database on the instance.

---

## 1. Why Most Teams Get This Wrong

The default path of least resistance:

```
1. Create PostgreSQL instance
2. Application connects as 'postgres' (superuser)
3. Everything works
4. Nobody changes it
5. Years pass
6. SQL injection vulnerability → full database compromise
```

The `postgres` superuser can:
- Read/write any table in any database
- Create/drop databases
- Create/drop roles
- Read files from the server filesystem
- Execute OS commands via `COPY TO PROGRAM`
- Bypass all RLS policies
- Modify `pg_hba.conf` authentication rules

**Your application needs none of these capabilities.**

---

## 2. The Principle of Least Privilege

```mermaid
flowchart TD
    A["Superuser\n(postgres)"] --> B["Admin role\n(schema changes, maintenance)"]
    B --> C["Application role\n(CRUD on specific tables)"]
    
    D["Migration role\n(DDL permissions)"]
    E["Read-only role\n(analytics, reporting)"]
    F["Monitoring role\n(pg_stat views only)"]

    B --> D
    B --> E
    B --> F

    style A fill:#f00,color:#fff
    style C fill:#6f6
```

### Role Hierarchy

```sql
-- 1. Application role (daily operations)
CREATE ROLE app_user LOGIN PASSWORD 'strong_password';

-- 2. Migration role (schema changes)
CREATE ROLE migrator LOGIN PASSWORD 'another_password';

-- 3. Read-only role (analytics)
CREATE ROLE analyst LOGIN PASSWORD 'readonly_password';

-- 4. Monitoring role (observability)
CREATE ROLE monitor LOGIN PASSWORD 'monitor_password';

-- Keep postgres superuser for emergency access only
```

---

## 3. Granting Privileges — A Complete Example

### Application Role

```sql
-- Connect to the application database
\c myapp

-- Grant access to the schema
GRANT USAGE ON SCHEMA public TO app_user;

-- Grant table access (SELECT, INSERT, UPDATE, DELETE — not DROP, TRUNCATE, ALTER)
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO app_user;

-- Grant sequence access (for SERIAL/GENERATED columns)
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO app_user;

-- Apply to future tables too
ALTER DEFAULT PRIVILEGES IN SCHEMA public
  GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO app_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA public
  GRANT USAGE, SELECT ON SEQUENCES TO app_user;
```

### Migration Role

```sql
-- Can modify schema but shouldn't have runtime DML access
GRANT ALL ON SCHEMA public TO migrator;
GRANT ALL ON ALL TABLES IN SCHEMA public TO migrator;
GRANT ALL ON ALL SEQUENCES IN SCHEMA public TO migrator;

-- Can create tables, indexes, etc.
ALTER DEFAULT PRIVILEGES IN SCHEMA public
  GRANT ALL ON TABLES TO migrator;
```

### Read-Only Role

```sql
GRANT USAGE ON SCHEMA public TO analyst;
GRANT SELECT ON ALL TABLES IN SCHEMA public TO analyst;

ALTER DEFAULT PRIVILEGES IN SCHEMA public
  GRANT SELECT ON TABLES TO analyst;

-- Ensure analyst cannot write
-- (Explicitly, no INSERT/UPDATE/DELETE granted)
```

### Monitoring Role

```sql
-- Grant access to monitoring views
GRANT pg_monitor TO monitor;  -- Built-in role (PG 10+)
-- This gives access to:
--   pg_stat_activity
--   pg_stat_replication
--   pg_stat_wal
--   And other pg_stat_* views
```

---

## 4. Schema-Level Isolation

```sql
-- Create separate schemas for different concerns
CREATE SCHEMA app;        -- Application tables
CREATE SCHEMA analytics;  -- Analytics/reporting views
CREATE SCHEMA internal;   -- Internal/admin tables

-- Application role only accesses app schema
GRANT USAGE ON SCHEMA app TO app_user;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA app TO app_user;

-- Analyst role only accesses analytics schema
GRANT USAGE ON SCHEMA analytics TO analyst;
GRANT SELECT ON ALL TABLES IN SCHEMA analytics TO analyst;

-- Application has no access to internal tables
-- (No GRANT issued = no access)
```

### The `search_path` Setting

```sql
-- Set per-role search path
ALTER ROLE app_user SET search_path = 'app,public';
ALTER ROLE analyst SET search_path = 'analytics,public';
```

---

## 5. Column-Level Privileges

For tables with sensitive columns:

```sql
-- Grant access to non-sensitive columns only
GRANT SELECT (id, name, email, created_at) ON users TO analyst;

-- Deny access to sensitive columns
-- (Simply don't grant SELECT on salary, ssn, etc.)
-- Analyst cannot: SELECT * FROM users (will error on restricted columns)
-- Analyst can: SELECT id, name, email FROM users
```

### Function-Level Privileges

```sql
-- Create a function that encapsulates privileged access
CREATE FUNCTION get_user_summary(user_id INT)
RETURNS TABLE (name TEXT, order_count BIGINT) AS $$
  SELECT u.name, COUNT(o.id)
  FROM users u
  LEFT JOIN orders o ON o.user_id = u.id
  WHERE u.id = user_id
  GROUP BY u.name;
$$ LANGUAGE sql SECURITY DEFINER;  -- Runs with function owner's privileges

-- Grant execute to limited role
GRANT EXECUTE ON FUNCTION get_user_summary TO app_user;
REVOKE ALL ON FUNCTION get_user_summary FROM PUBLIC;
```

**`SECURITY DEFINER` warning:** The function runs with the privileges of the user who created it (often superuser). Validate all inputs inside the function — it's a privilege escalation vector.

---

## 6. Common Privilege Mistakes

### Mistake 1: Granting to PUBLIC

```sql
-- BAD: Every role can access this table
GRANT SELECT ON users TO PUBLIC;

-- The PUBLIC pseudo-role means "everyone"
-- Including roles that haven't been created yet
```

**Fix:**
```sql
-- Revoke default PUBLIC permissions
REVOKE ALL ON SCHEMA public FROM PUBLIC;
REVOKE ALL ON ALL TABLES IN SCHEMA public FROM PUBLIC;
```

### Mistake 2: Application as Table Owner

```sql
-- If app_user creates tables, it OWNS them
-- Owners can DROP, ALTER, GRANT — too much privilege

-- Better: migrator creates tables, app_user gets limited access
-- Tables are owned by migrator, used by app_user
```

### Mistake 3: Password in Connection String

```
# BAD: password visible in process list, logs, env dump
postgresql://app_user:SuperSecret123@db.example.com/myapp

# BETTER: use .pgpass file or environment variable
export PGPASSWORD=SuperSecret123  # Still not great

# BEST: use certificate authentication or IAM roles (cloud)
```

---

## 7. Auditing Privileges

```sql
-- What can a role do?
SELECT
  grantee,
  table_schema,
  table_name,
  privilege_type
FROM information_schema.table_privileges
WHERE grantee = 'app_user'
ORDER BY table_schema, table_name;

-- What roles does a user have?
SELECT
  r.rolname,
  r.rolsuper,
  r.rolinherit,
  r.rolcreaterole,
  r.rolcreatedb,
  r.rolcanlogin,
  ARRAY(SELECT b.rolname FROM pg_catalog.pg_auth_members m
        JOIN pg_catalog.pg_roles b ON m.roleid = b.oid
        WHERE m.member = r.oid) AS member_of
FROM pg_catalog.pg_roles r
WHERE r.rolname NOT LIKE 'pg_%'
ORDER BY r.rolname;

-- Who can connect?
SELECT * FROM pg_hba_file_rules;  -- PG 10+
```

---

## 8. pg_hba.conf — Connection Authentication

The `pg_hba.conf` file controls **who can connect** and **how they authenticate**:

```
# TYPE  DATABASE  USER       ADDRESS         METHOD

# Superuser: only from localhost
local   all       postgres                   peer
host    all       postgres   127.0.0.1/32    scram-sha-256

# Application: from app servers only
host    myapp     app_user   10.0.1.0/24     scram-sha-256

# Monitoring: from monitoring servers
host    myapp     monitor    10.0.2.0/24     scram-sha-256

# Analyst: from VPN only
host    myapp     analyst    10.0.3.0/24     scram-sha-256

# Deny everything else
host    all       all        0.0.0.0/0       reject
```

### Authentication Methods

| Method | Security | Use for |
|--------|----------|---------|
| `peer` | High (OS user = DB user) | Local superuser access |
| `scram-sha-256` | High | Network connections |
| `md5` | Medium (legacy) | Avoid if possible |
| `trust` | **None** | Development only, never production |
| `cert` | Very high | Mutual TLS authentication |
| `gss` / `ldap` | Enterprise | Centralized authentication |

---

## 9. Thinking Traps Summary

| Trap | What breaks | Prevention |
|------|------------|------------|
| App connects as superuser | SQL injection = full server compromise | Dedicated app role with minimal privileges |
| Granting to PUBLIC | Every role gets access | `REVOKE ALL FROM PUBLIC` |
| App role owns tables | Can DROP, ALTER schema | Migrator role owns, app role uses |
| `trust` auth in production | Anyone can connect without password | Use `scram-sha-256` or `cert` |
| Not auditing privileges | Privilege creep over time | Periodic review of role grants |
| `SECURITY DEFINER` without input validation | Privilege escalation | Validate all inputs in DEFINER functions |

---

## Related Files

- [Security_and_Governance/01_sql_injection_beyond_basics.md](01_sql_injection_beyond_basics.md) — what injection can do with too many privileges
- [Security_and_Governance/02_row_level_security.md](02_row_level_security.md) — row-level access control
- [Security_and_Governance/04_secrets_and_connection_security.md](04_secrets_and_connection_security.md) — securing credentials
