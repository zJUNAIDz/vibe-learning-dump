# Migrations as Team Practice: Ownership and Coordination

## Who Owns Migrations?

**Wrong answer**: "The person who writes them."

**Right answer**: "The whole team. But especially senior engineers."

### Why Migrations Are a Senior Responsibility

Junior engineer thinking:
> "I need to add a column. I'll write an ALTER TABLE and push it."

Senior engineer thinking:
> "I need to add a column. Let me consider:
> - What's the safest approach?
> - How does this affect the team?
> - What breaks if this goes wrong?
> - Who needs to be involved?
> - What's the communication plan?"

**Migrations are high-leverage, high-risk changes.** They deserve senior-level scrutiny.

### The Ownership Model

```
Junior Engineer:
  - Writes initial migration
  - Tests locally
  - Documents intent

Mid-Level Engineer:
  - Reviews migration
  - Considers edge cases
  - Tests on production-sized data

Senior Engineer:
  - Final review
  - Identifies risks
  - Approves for production
  - Often present during deployment

Staff/Principal Engineer:
  - Consulted for complex migrations
  - Defines migration standards
  - Incident response
```

**This doesn't mean juniors can't write migrations.** It means migrations should be reviewed and approved at a senior level.

## Team Coordination

### The Communication Problem

```
Engineer A: Deploying migration that drops username column
Engineer B: [unaware] Deploying feature that uses username
Both: Deploy simultaneously
Result: 💥 Production breaks
```

### Coordination Strategies

#### 1. Migration Announcement Protocol

Before running a migration:

```markdown
**Migration Announcement**: Add phone column to users

**Who**: @alice
**When**: Tuesday 2024-02-20, 14:00 UTC (2 PM)
**What**: Adding nullable `phone` column to `users` table
**Duration**: ~30 seconds (tested in staging)
**Risk**: Low (metadata-only, nullable column)
**Requires**: No code changes required
**Impact**: None (backward compatible)
**Rollback**: Can drop column if needed

**Asks**:
- @bob: Your feature PR modifies users table, please hold merge until after this
- @on-call: Please be available for 30 minutes after deployment

**Links**:
- Migration PR: #1234
- Staging test results: [link]
```

Post in team Slack/Discord before running.

#### 2. Migration Calendar

Shared calendar of scheduled migrations:

```
Week of Feb 19:
  Mon - No migrations (post-deploy stabilization)
  Tue 14:00 - @alice: Add phone to users
  Wed -      Reserved for hotfixes only
  Thu 10:00 - @bob: Create notifications table
  Fri -      No migrations (pre-weekend freeze)
```

Prevents conflicts, gives visibility.

#### 3. Migration Freeze Windows

```
Freeze periods (no migrations allowed):
  - Fridays after 12pm
  - Weekends
  - Major holidays
  - During incidents
  - Feature launch days
  - Peak traffic events (Black Friday, etc.)
```

Reduces risk during high-stakes periods.

## Code Review for Migrations

### What Reviewers Should Check

#### Technical Review (Mid-Level+)

```markdown
## Migration Review Checklist

### Correctness
- [ ] Migration does what it claims
- [ ] SQL syntax is correct
- [ ] Idempotent (can run twice safely)

### Safety
- [ ] Uses CONCURRENTLY for indexes
- [ ] Uses NOT VALID for constraints
- [ ] Doesn't add NOT NULL without preparation
- [ ] Doesn't drop columns still in use
- [ ] Lock duration is acceptable

### Testing
- [ ] Tested on production-sized data
- [ ] Lock duration measured
- [ ] Tested in staging
- [ ] Rollback tested

### Compatibility
- [ ] Backward compatible with current code
- [ ] Forward compatible with new code
- [ ] No breaking changes

### Documentation
- [ ] Comments explain why
- [ ] Risk assessment included
- [ ] Rollback plan documented
```

#### Senior Review

```markdown
## Senior Review Checklist

### Architecture
- [ ] Aligns with long-term schema direction
- [ ] Doesn't create tech debt
- [ ] Follows team conventions

### Risk Assessment
- [ ] Failure modes identified
- [ ] Impact of failure is acceptable
- [ ] Monitoring plan exists
- [ ] On-call is aware

### Alternatives Considered
- [ ] Is this the best approach?
- [ ] Were other options considered?
- [ ] Why was this chosen?

### Team Impact
- [ ] Other PRs affected?
- [ ] Communication plan exists
- [ ] Timing is appropriate
```

### The Review Process

```
1. Engineer creates migration PR
   ↓
2. Automated checks run
   - Lint migration SQL
   - Check for dangerous patterns
   - Run on test database
   ↓
3. Peer review (mid-level engineer)
   - Technical correctness
   - Best practices
   ↓
4. Senior review
   - Architecture alignment
   - Risk assessment
   - Approve or request changes
   ↓
5. Staging deployment
   - Run on production-like data
   - Monitor for issues
   ↓
6. Production approval
   - Senior sign-off
   - Schedule deployment
   ↓
7. Production deployment
   - Active monitoring
   - Team on standby
```

### Review Comments That Help

**Bad review comment**:
> "Looks good 👍"

**Good review comment**:
> "This adds NOT NULL immediately on a 10M row table. That will lock for ~30 seconds. Recommend using the multi-step pattern:
> 1. Add nullable column
> 2. Deploy app code
> 3. Backfill
> 4. Add NOT NULL
>
> See [link to docs] for pattern."

**Great review comment**:
> "This migration drops the `username` column. I see the app code no longer uses it, but:
> 
> - Background worker `process_email_queue.ts` still queries it (line 45)
> - Analytics dashboard has a query that joins on username
> - Our partner API contract includes username in responses
>
> We need to:
> 1. Update background worker
> 2. Notify analytics team
> 3. Version the API to deprecate username
>
> Timeline should be ~2 weeks, not immediate drop."

**The difference**: Specificity, alternatives, and consideration of broader impact.

## Dev/Prod Drift Prevention

### The Problem

```
Scenario 1:
- Engineer tests locally (empty database)
- Migration works great
- Pushes to production
- 💥 Fails on production data

Scenario 2:
- Engineer manually tweaks staging database
- Staging works
- Production has different schema
- 💥 Migration fails

Scenario 3:
- Migration runs in production
- Never applied to staging
- Staging and production diverge
- 💥 Next migration fails in staging
```

### Prevention Strategies

#### 1. Schema Validation in CI

```typescript
// scripts/validate-schema.ts
import { exec } from 'child_process';
import { promisify } from 'util';

const execAsync = promisify(exec);

async function validateSchema() {
  // Apply all migrations to fresh database
  await execAsync('docker run -d --name test-postgres postgres');
  await execAsync('sleep 3');  // Wait for postgres to start
  await execAsync('migrate -database "postgres://localhost/test" up');
  
  // Compare schema hash with expected
  const { stdout } = await execAsync('pg_dump -s test | md5sum');
  const currentHash = stdout.trim();
  
  const expectedHash = readFileSync('.schema-hash', 'utf-8').trim();
  
  if (currentHash !== expectedHash) {
    throw new Error('Schema drift detected! Migrations changed.');
  }
}
```

Run in CI on every PR.

#### 2. Periodic Schema Dumps

```bash
# Weekly job: dump production schema
pg_dump -s production > schema_dumps/schema_$(date +%Y%m%d).sql

# Compare with previous week
diff schema_dumps/schema_20240215.sql schema_dumps/schema_20240220.sql
```

Catches manual changes.

#### 3. Migration Hygiene

```markdown
## Team Rules

1. **Never** manually alter production schema without migration
2. **Always** run migrations in order: local → staging → production
3. **Always** keep staging in sync with production schema
4. **Never** skip migrations or mark as applied without running
5. **Always** restore from production backup to debug staging issues
```

#### 4. Automated Sync Checks

```yaml
# .github/workflows/schema-check.yml
name: Schema Consistency Check

on:
  schedule:
    - cron: '0 9 * * *'  # Daily

jobs:
  check-drift:
    runs-on: ubuntu-latest
    steps:
      - name: Dump production schema
        run: pg_dump -s $PROD_DB > prod_schema.sql
      
      - name: Dump staging schema
        run: pg_dump -s $STAGING_DB > staging_schema.sql
      
      - name: Compare
        run: diff prod_schema.sql staging_schema.sql || exit 1
      
      - name: Notify on drift
        if: failure()
        run: |
          curl -X POST $SLACK_WEBHOOK \
            -d '{"text": "⚠️ Schema drift detected between prod and staging!"}'
```

## Documentation and Tribal Knowledge

### The Problem

Six months later:

```sql
SELECT * FROM users;

-- Huh, why does this table have a `legacy_id` column?
-- What's it for?
-- Can I drop it?
-- Who added it?
```

Nobody remembers. Original engineer left. No documentation.

**Tribal knowledge is lost.**

### The Solution: Document Everything

#### 1. Migration Comments

```sql
-- Migration: Add legacy_id column
-- Date: 2024-02-15
-- Author: Alice (@alice)
-- Ticket: PROJ-1234
--
-- Context:
-- We're migrating from legacy user system. This column stores
-- the old system's user ID for data reconciliation.
--
-- Expected lifespan: 6 months (until migration complete)
-- Can drop after: 2024-08-15
-- Dependencies: legacy_sync background job
--
-- Related migrations:
-- - 20240220_create_legacy_user_mapping.sql
--

ALTER TABLE users ADD COLUMN legacy_id VARCHAR(50);
CREATE INDEX CONCURRENTLY idx_users_legacy_id ON users(legacy_id);
```

Now anyone can understand why this exists.

#### 2. Schema Comments

```sql
-- Add persistent schema documentation
COMMENT ON COLUMN users.legacy_id IS 
  'Legacy system user ID. Used for data reconciliation during migration. 
   Can be removed after 2024-08-15 when migration is complete. See: PROJ-1234';
```

Visible in database tools:

```sql
\d+ users
```

Shows comments.

#### 3. Architecture Decision Records (ADRs)

```markdown
# ADR-015: Split users table into users and user_profiles

## Status
Accepted

## Context
Users table has grown to 50+ columns. Query performance degrading.
Profile fields (bio, avatar, etc.) rarely queried with auth fields.

## Decision
Split into:
- `users`: Auth-related fields (id, email, password_hash)
- `user_profiles`: Profile fields (bio, avatar, settings)

## Consequences
Positive:
- Smaller users table, faster auth queries
- Profile queries don't lock auth data

Negative:
- Requires JOIN for full user data
- Migration complexity (5-phase rollout)

## Migration Plan
See: migrations/20240215_split_users_table/PLAN.md

## Timeline
- Week 1: Create user_profiles table
- Week 2: Dual-write
- Week 3: Backfill
- Week 4: Switch reads
- Week 5: Drop old columns
```

#### 4. Migration README

```markdown
# Migration: Split Users Table

## Overview
Split `users` table into `users` and `user_profiles` for performance.

## Files
- `001_create_user_profiles.sql` - Create new table
- `002_backfill_profiles.sql` - Copy data
- `003_add_foreign_key.sql` - Add constraint
- `004_drop_old_columns.sql` - Clean up

## Running
Must be run in order. Each migration must complete before next.

## Rollback
Rollback is complex after step 2 (data exists in both tables).
See ROLLBACK.md for detailed procedure.

## Monitoring
Watch for:
- JOIN query performance (should improve)
- Write latency (may temporarily increase during dual-write phase)

## Dependencies
- Application code must be updated between migrations 2 and 3
- Background jobs updated: see deploy/background-jobs.md

## Contacts
- Primary: @alice
- Backup: @bob
- On-call: Notify before running
```

## Incident Response

### The Migration Incident Playbook

When a migration goes wrong in production:

#### Step 1: Assess (60 seconds)

```markdown
Questions:
- Is the migration still running?
- Is the application down?
- Are users affected?
- What's the error rate?
```

#### Step 2: Decide (30 seconds)

```markdown
Options:
A. Let it complete (if close to done, low impact)
B. Kill it (if clearly stuck, high impact)
C. Rollback app (if app code broken, migration succeeded)
```

#### Step 3: Act (immediately)

**If killing migration**:

```sql
-- Find the migration process
SELECT pid, query 
FROM pg_stat_activity 
WHERE query LIKE '%ALTER TABLE%';

-- Kill it
SELECT pg_terminate_backend(12345);  -- Use actual PID
```

**If rolling back app**:

```bash
kubectl rollout undo deployment/api
# Or your deployment tool
```

#### Step 4: Communicate (within 5 minutes)

```markdown
Subject: Incident - Migration failure on users table

Status: Active incident
Started: 14:32 UTC
Current: Investigating

What happened:
- Migration to add index on users table started
- Index creation not using CONCURRENTLY
- Writes blocked for 5 minutes
- Application error rate 100%

Actions taken:
- Terminated migration at 14:37 UTC
- Writes recovering
- Application recovering

Next steps:
- Verify data integrity
- Clean up partial index
- Prepare corrected migration

ETA to resolution: 15:00 UTC

Updates: Every 15 minutes or significant change
```

Post in:
- Team chat
- Status page (if customer-facing)
- Stakeholders

#### Step 5: Recover

```sql
-- Check for partial index
\di
-- If index is marked INVALID, drop it
DROP INDEX idx_users_email;

-- Verify table health
SELECT COUNT(*) FROM users;

-- Check for locks
SELECT COUNT(*) FROM pg_stat_activity WHERE wait_event_type = 'Lock';
```

#### Step 6: Post-Mortem (within 24 hours)

See [Rollback and Failure Recovery](./07_rollback_and_failure_recovery.md) for postmortem template.

## Multi-Team Coordination

### Scenario: Shared Database

Multiple teams use the same database:

```
Team A: User Service (owns users table)
Team B: Billing Service (owns invoices table, FKs to users)
Team C: Analytics (reads everything)
```

**Team A wants to change users table. What now?**

#### The Process

```markdown
1. Proposal Phase
   - Team A proposes change
   - Share design doc with Teams B, C
   - Gather feedback

2. Review Phase
   - Each team reviews impact
   - Team B: "We join on username, don't drop it"
   - Team C: "We can update our queries"

3. Negotiation Phase
   - Discuss alternatives
   - Agree on approach
   - Set timeline

4. Coordination Phase
   - Team A schedules migration
   - Team B updates queries
   - Team C updates dashboards
   - All teams sync deployments

5. Execution Phase
   - Team A runs migration
   - Teams B, C deploy updates
   - All teams monitor
```

#### The Communication

```markdown
**RFC: Rename users.username to users.handle**

**Proposer**: Team A - User Service
**Stakeholders**: Team B (Billing), Team C (Analytics)

**Context**:
We're standardizing terminology. "handle" better reflects usage.

**Proposed Change**:
Rename `users.username` → `users.handle`

**Impact Analysis**:

Team B (Billing):
- 3 queries join on username
- Estimated update: 2 hours
- Requires code deploy

Team C (Analytics):
- 12 dashboard queries use username
- Estimated update: 4 hours
- No deploy required (can update live)

**Timeline**:
- Week 1: Update Team B queries
- Week 2: Update Team C dashboards
- Week 3: Run migration (expand-migrate-contract)
- Week 4: Verify, clean up

**Risks**:
- If team B doesn't update, their queries break
- Analytics dashboards might show empty data temporarily

**Mitigation**:
- Use expand-migrate-contract (both columns exist during transition)
- Feature flag to switch between username/handle
- Gradual rollout

**Questions / Concerns**:
[Discussion thread]
```

## Team Standards and Style Guides

Document your team's conventions:

```markdown
# Team Migration Standards

## Naming Conventions
- Timestamp format: YYYYMMDDHHmmss
- Descriptive names: `add_phone_to_users`, not `update_users`
- Prefix with table name for clarity

## SQL Style
- Use uppercase for keywords: `SELECT`, `FROM`, `WHERE`
- One statement per line for DDL
- Always include comments explaining why

## Safety Requirements
- [ ] All indexes use CONCURRENTLY
- [ ] All constraints use NOT VALID pattern
- [ ] No NOT NULL without multi-step
- [ ] Tested on production-sized data

## Review Requirements
- Peer review (mid-level)
- Senior approval
- Staging validation

## Deployment
- Announce 24h in advance
- Run during low-traffic windows
- Active monitoring required
- On-call notified

## Documentation
- Migration comments required
- Schema comments for new columns
- ADR for major changes
- Update team wiki
```

## The Senior Engineer's Role

As you become senior, your migration responsibilities evolve:

**Junior**:
- Write migrations
- Test locally
- Follow patterns

**Mid**:
- Review migrations
- Catch unsafe patterns
- Test thoroughly

**Senior**:
- Design migration strategies
- Make risk decisions
- Lead incident response
- Teach best practices
- Define team standards

**Staff/Principal**:
- Set organization-wide standards
- Architect complex migrations
- Resolve cross-team conflicts
- Incident prevention

**The progression**: From executing migrations to ensuring migrations across the organization are safe.

## Summary: Team Practices Checklist

For healthy migration practices:

- [ ] Clear ownership model (junior writes, senior approves)
- [ ] Migration announcement protocol
- [ ] Shared migration calendar
- [ ] Freeze windows defined
- [ ] Code review checklist
- [ ] Schema validation in CI
- [ ] Periodic drift checks
- [ ] Documentation standards
- [ ] Incident playbook
- [ ] Multi-team coordination process
- [ ] Team style guide
- [ ] Regular training

**The goal**: Migrations are not solo acts. They're team practices that require coordination, communication, and collective responsibility.

---

## Conclusion: You're Ready

You've made it through the entire migration masterclass. You now know:

1. **Philosophy**: Migrations are contract changes, not just DDL
2. **Basics**: How migration systems work
3. **Locks**: Why ALTER TABLE is scary and how to minimize risk
4. **Data**: How to transform production data safely
5. **Zero-downtime**: Expand-migrate-contract pattern
6. **Indexes**: CONCURRENTLY and constraint staging
7. **Rollbacks**: Why they're often lies, and how to fix forward
8. **Tooling**: Drizzle vs go-migrate and when to use each
9. **Testing**: How to catch issues before production
10. **Real-world**: Battle-tested patterns and postmortems
11. **Mistakes**: What not to do
12. **Team practices**: Coordination and communication

**You're no longer a rookie.**

You understand:
- The risks
- The patterns
- The tradeoffs
- The failure modes

More importantly, you understand **why** migrations are done the way they are.

**The shift**: From copying examples to making confident decisions.

**Go forth and migrate safely.** 🚀

---

*Have questions or feedback? This guide is a living document. Contribute patterns, corrections, and war stories.*
