# Secrets and Connection Security

> **What mistake does this prevent?**
> Database credentials in source code, unencrypted connections that expose every query to network sniffers, and secrets management practices that make rotation impossible.

---

## 1. The Credentials Problem

```mermaid
flowchart TD
    A["Where do credentials live?"] --> B["✗ Hardcoded in source"]
    A --> C["✗ .env file committed to git"]
    A --> D["✗ Environment variable\n(visible in ps, /proc)"]
    A --> E["✓ Secrets manager\n(runtime injection)"]
    A --> F["✓ Certificate\nauthentication"]

    style B fill:#f00,color:#fff
    style C fill:#f00,color:#fff
    style D fill:#fa0
    style E fill:#6f6
    style F fill:#6f6
```

### Evolution of Secrets Handling

| Approach | Risk | Rotation cost |
|----------|------|---------------|
| Hardcoded in code | In every git clone, every cache, every backup | Code change + deploy |
| `.env` file | File permissions, accidental git commit | File change + restart |
| Environment variable | Visible in `ps`, process info, container inspect | Restart process |
| Secrets manager (Vault, AWS SM) | Network dependency | API call, zero downtime possible |
| Certificate auth | Certificate management | Cert rotation, no password |
| IAM auth (AWS RDS, GCP Cloud SQL) | Cloud-native only | Automatic |

---

## 2. SSL/TLS for PostgreSQL Connections

### Why It Matters

Without SSL, every query is transmitted as plaintext. Anyone on the network can see:

```
-- What a network sniffer (Wireshark) captures without SSL:
Q: SELECT * FROM users WHERE email = 'alice@company.com'
D: id=1, email=alice@company.com, password_hash=$2b$12$...
   id=2, email=bob@company.com, password_hash=$2b$12$...
```

### Enabling SSL on PostgreSQL Server

```ini
# postgresql.conf
ssl = on
ssl_cert_file = '/etc/postgresql/server.crt'
ssl_key_file = '/etc/postgresql/server.key'
ssl_ca_file = '/etc/postgresql/ca.crt'       # For client cert verification
ssl_min_protocol_version = 'TLSv1.3'        # PG 12+
```

### Connection SSL Modes

```
# Connection string sslmode parameter
postgresql://user:pass@host/db?sslmode=MODE

# Modes (from least to most secure):
# disable    → No SSL. Don't even try.
# allow      → Try non-SSL first, SSL if server requires it.
# prefer     → Try SSL first, fall back to non-SSL. (DEFAULT)
# require    → Must use SSL. Don't verify server certificate.
# verify-ca  → Must use SSL. Verify server cert is signed by trusted CA.
# verify-full→ Must use SSL. Verify cert + hostname. MOST SECURE.
```

```mermaid
flowchart TD
    A["sslmode options"] --> B["disable / allow / prefer\n❌ MITM possible"]
    A --> C["require\n⚠️ Encrypted but no cert check"]
    A --> D["verify-ca\n✓ Cert verified"]
    A --> E["verify-full\n✓ Cert + hostname verified"]

    style B fill:#f00,color:#fff
    style C fill:#fa0
    style D fill:#6f6
    style E fill:#6f6
```

**The common mistake:** Using `sslmode=require` and thinking you're secure. Without certificate verification, a man-in-the-middle can present their own certificate and decrypt traffic.

**Minimum for production:** `verify-ca`. Ideally `verify-full`.

---

## 3. Connection String Hygiene

### What NOT to Do

```typescript
// ❌ Hardcoded credentials
const pool = new Pool({
  connectionString: 'postgresql://admin:MyP@ssw0rd@prod-db.internal:5432/myapp'
});

// ❌ Password in environment (better but still visible)
const pool = new Pool({
  connectionString: process.env.DATABASE_URL  // password in env var
});
```

### What to Do: Secrets Manager

```typescript
// ✓ Fetch from secrets manager at startup
import { SecretsManagerClient, GetSecretValueCommand } from '@aws-sdk/client-secrets-manager';

async function getDatabaseUrl(): Promise<string> {
  const client = new SecretsManagerClient({ region: 'us-east-1' });
  const response = await client.send(
    new GetSecretValueCommand({ SecretId: 'prod/myapp/database' })
  );
  return JSON.parse(response.SecretString!).connectionString;
}

const pool = new Pool({
  connectionString: await getDatabaseUrl()
});
```

### Connection String Components

```
postgresql://user:password@host:port/database?sslmode=verify-full&sslrootcert=/path/to/ca.crt
             │    │        │    │    │         │                     │
             │    │        │    │    │         │                     └─ CA certificate path
             │    │        │    │    │         └─ SSL verification mode
             │    │        │    │    └─ Database name
             │    │        │    └─ Port (default 5432)
             │    │        └─ Host
             │    └─ Password (the secret part)
             └─ Username
```

---

## 4. Password Rotation

### The Problem

If you change a database password, every application using that password must be updated simultaneously.

### Strategy: Dual-Password Rotation

```mermaid
sequenceDiagram
    participant SM as Secrets Manager
    participant App as Application
    participant DB as PostgreSQL

    SM->>DB: 1. ALTER ROLE app_user SET PASSWORD 'new_password'
    Note over DB: Both old and new passwords are valid<br>(PG only keeps one, so dual-user approach needed)
    SM->>SM: 2. Update secret to new_password
    App->>SM: 3. Next connection fetch gets new_password
    App->>DB: 4. Connects with new_password ✓
```

### Dual-User Rotation (Recommended)

```sql
-- Use two roles that alternate
-- Phase 1: app_user_a is active
CREATE ROLE app_user_a LOGIN PASSWORD 'password_1';
GRANT app_access TO app_user_a;

-- Phase 2: Rotate to app_user_b
CREATE ROLE app_user_b LOGIN PASSWORD 'password_2';
GRANT app_access TO app_user_b;

-- Phase 3: Update secrets manager to point to app_user_b
-- Phase 4: After all connections drain, revoke app_user_a
ALTER ROLE app_user_a PASSWORD NULL;  -- Disable old password
```

---

## 5. Certificate-Based Authentication

Eliminate passwords entirely:

```ini
# pg_hba.conf
hostssl  myapp  app_user  10.0.0.0/8  cert  clientcert=verify-full
```

```
# Client connection with certificate
postgresql://app_user@prod-db:5432/myapp?\
  sslmode=verify-full&\
  sslcert=/etc/ssl/client.crt&\
  sslkey=/etc/ssl/client.key&\
  sslrootcert=/etc/ssl/ca.crt
```

### Advantages Over Passwords

| | Passwords | Certificates |
|---|-----------|-------------|
| Rotation | Must update every client | Overlapping validity periods |
| Sniffing | Sent over connection (encrypted by SSL) | Never transmitted |
| Brute force | Possible if exposed | Computationally infeasible |
| Revocation | Change password = disruption | Revoke cert in CA |

---

## 6. Cloud-Specific Authentication

### AWS RDS IAM Authentication

```typescript
import { RDS } from '@aws-sdk/client-rds';

const signer = new RDS.Signer({
  hostname: 'mydb.region.rds.amazonaws.com',
  port: 5432,
  username: 'app_user',
  region: 'us-east-1'
});

// Token is valid for 15 minutes, auto-generated from IAM role
const token = signer.getAuthToken();

const pool = new Pool({
  host: 'mydb.region.rds.amazonaws.com',
  user: 'app_user',
  password: token,
  ssl: { rejectUnauthorized: true, ca: rdsCA }
});
```

### GCP Cloud SQL

```typescript
// Uses Unix socket or Cloud SQL Auth Proxy
// No password needed — authenticated by service account
const pool = new Pool({
  host: '/cloudsql/project:region:instance',  // Unix socket
  user: 'app_user',
  database: 'myapp'
});
```

---

## 7. Logging and Credential Exposure

### PostgreSQL Logging Risks

```ini
# postgresql.conf
log_statement = 'all'  # DANGEROUS: logs all SQL including SET commands

# This logs:
# LOG: statement: SET app.api_key = 'sk-1234567890abcdef'

# Better: use 'ddl' or 'mod' in production
log_statement = 'mod'  # Only INSERT/UPDATE/DELETE/DDL

# Also dangerous:
log_connections = on  # Can log connection strings (though PG redacts passwords)
```

### Application Logging

```typescript
// ❌ Logging connection strings
console.log(`Connecting to ${process.env.DATABASE_URL}`);

// ✓ Log only safe parts
console.log(`Connecting to ${dbHost}:${dbPort}/${dbName} as ${dbUser}`);
```

---

## 8. Thinking Traps Summary

| Trap | What breaks | Prevention |
|------|------------|------------|
| `sslmode=prefer` (default) | Connection falls back to plaintext | Always use `verify-full` |
| Password in environment variable | Visible in proc filesystem, container inspect | Secrets manager or cert auth |
| `sslmode=require` without cert check | MITM can intercept traffic | Use `verify-ca` or `verify-full` |
| `log_statement = 'all'` | Secrets logged to disk | Use `mod` or `ddl` |
| Single-user password rotation | Downtime during rotation | Dual-user rotation strategy |
| Credentials in source code | Every clone has prod credentials | `.gitignore`, pre-commit hooks, secrets scanning |

---

## Related Files

- [Security_and_Governance/03_roles_privileges_least_privilege.md](03_roles_privileges_least_privilege.md) — role design
- [Production_Postgres/01_connection_management.md](../Production_Postgres/01_connection_management.md) — pooling with SSL
- [Security_and_Governance/05_data_privacy_and_pii.md](05_data_privacy_and_pii.md) — protecting sensitive data
