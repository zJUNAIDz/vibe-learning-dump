# Idempotency

> **An idempotent operation produces the same result whether you run it once or a hundred times. Ansible modules are idempotent by default. Your custom tasks might not be. Know the difference.**

---

## 🟢 What Is Idempotency?

```
Idempotent:
  "Install nginx" → Run once: installs nginx
                   → Run again: "already installed, no change"
                   → Run 50 times: same result, one install

NOT Idempotent:
  "Append line to file" → Run once: one line
                        → Run again: two lines (duplicate!)
                        → Run 50 times: 50 identical lines 😱
```

### Why It Matters

```
Scenario: CI pipeline runs Ansible every 15 minutes

WITHOUT idempotency:
  Run 1:  Server configured correctly ✅
  Run 2:  Config file doubled, extra users created, cron runs twice ❌
  Run 96: Server is a mess, nginx has 96 duplicate lines in config 💥

WITH idempotency:
  Run 1:  Server configured correctly ✅
  Run 2:  "No changes needed" ✅
  Run 96: "No changes needed" ✅

Idempotent playbooks are SAFE to run anytime.
```

---

## 🟢 Modules That Are Naturally Idempotent

```yaml
# apt → checks if package is installed first
- name: Install nginx
  apt:
    name: nginx
    state: present      # "Ensure installed" not "install now"
  # Run 1: installs nginx    → changed
  # Run 2: already installed → ok (no change)

# file → checks current state
- name: Create directory
  file:
    path: /opt/myapp
    state: directory
    mode: '0755'
  # Run 1: creates directory → changed
  # Run 2: already exists    → ok

# user → checks if user exists
- name: Create user
  user:
    name: appuser
    shell: /bin/bash
  # Run 1: creates user → changed
  # Run 2: user exists  → ok

# template → compares rendered output with existing file
- name: Configure nginx
  template:
    src: nginx.conf.j2
    dest: /etc/nginx/nginx.conf
  # Run 1: file doesn't exist or differs → changed
  # Run 2: file matches template output   → ok

# service → checks current state
- name: Start nginx
  service:
    name: nginx
    state: started    # "Ensure running" not "start now"
  # Run 1: not running, starts it → changed
  # Run 2: already running        → ok
```

---

## 🟢 Modules That AREN'T Idempotent (Be Careful!)

### command / shell

```yaml
# ❌ NOT IDEMPOTENT — runs every time
- name: Run migration
  command: /opt/myapp/migrate.sh

# ✅ MADE IDEMPOTENT — only runs if marker missing
- name: Run migration
  command: /opt/myapp/migrate.sh
  args:
    creates: /opt/myapp/.migrated    # Skip if this file exists
```

### lineinfile gotcha

```yaml
# ✅ IDEMPOTENT — updates existing line or adds it once
- name: Set max connections
  lineinfile:
    path: /etc/postgresql/15/main/postgresql.conf
    regexp: '^max_connections'        # Find this line
    line: 'max_connections = 200'     # Replace with this
  # Run 1: adds/replaces → changed
  # Run 2: already set   → ok

# ❌ NOT IDEMPOTENT — no regexp, adds every time!
- name: Add line (WRONG)
  lineinfile:
    path: /etc/myapp/config
    line: 'FEATURE=enabled'
    # No regexp → Ansible checks exact line but may duplicate
```

---

## 🟡 Making Command Tasks Idempotent

### Strategy 1: creates / removes

```yaml
# Only run if output doesn't exist
- name: Build application
  command: make build
  args:
    chdir: /opt/myapp/src
    creates: /opt/myapp/bin/myapp    # Skip if binary exists

# Only run if something needs cleanup
- name: Clean build artifacts
  command: make clean
  args:
    chdir: /opt/myapp/src
    removes: /opt/myapp/src/build    # Skip if build dir gone
```

### Strategy 2: Register + when

```yaml
- name: Check if database initialized
  command: psql -U myapp -d myapp -c "SELECT 1 FROM schema_migrations LIMIT 1"
  register: db_check
  changed_when: false
  failed_when: false              # Don't fail if table doesn't exist

- name: Initialize database
  command: /opt/myapp/myapp db:init
  when: db_check.rc != 0          # Only run if check failed
```

### Strategy 3: changed_when

```yaml
# Tell Ansible WHEN a command actually changed something
- name: Add apt repository
  shell: |
    if ! grep -q "nodesource" /etc/apt/sources.list.d/*; then
      curl -fsSL https://deb.nodesource.com/setup_18.x | bash -
      echo "ADDED"
    else
      echo "EXISTS"
    fi
  register: repo_result
  changed_when: "'ADDED' in repo_result.stdout"
```

---

## 🟡 Testing Idempotency

```bash
# Run playbook twice — second run should show NO changes

# First run:
ansible-playbook site.yml
# PLAY RECAP
# web1: ok=15  changed=12  failed=0

# Second run (MUST be zero changed):
ansible-playbook site.yml
# PLAY RECAP
# web1: ok=15  changed=0   failed=0   ← This is the goal!
```

### Check Mode (Dry Run)

```bash
# Show what WOULD change without actually changing anything
ansible-playbook site.yml --check --diff

# --check → Don't make changes
# --diff  → Show what would change (for files)
```

---

## 🟡 Common Idempotency Traps

### Trap 1: Shell scripts that aren't idempotent

```yaml
# ❌ BAD — appends EVERY run
- name: Add to PATH
  shell: echo 'export PATH=/opt/myapp/bin:$PATH' >> ~/.bashrc

# ✅ GOOD — lineinfile checks first
- name: Add to PATH
  lineinfile:
    path: ~/.bashrc
    line: 'export PATH=/opt/myapp/bin:$PATH'
    regexp: '/opt/myapp/bin'
```

### Trap 2: Downloading files every time

```yaml
# ❌ BAD — downloads every run even if file exists
- name: Download binary
  shell: curl -o /opt/myapp/binary https://example.com/binary

# ✅ GOOD — get_url checks if file exists + checksum
- name: Download binary
  get_url:
    url: https://example.com/binary
    dest: /opt/myapp/binary
    checksum: sha256:abc123...
    mode: '0755'
```

### Trap 3: Service restarts without handlers

```yaml
# ❌ BAD — restarts nginx EVERY run
- name: Update config
  template:
    src: nginx.conf.j2
    dest: /etc/nginx/nginx.conf

- name: Restart nginx
  service:
    name: nginx
    state: restarted     # Runs every time regardless!

# ✅ GOOD — only restart when config actually changes
- name: Update config
  template:
    src: nginx.conf.j2
    dest: /etc/nginx/nginx.conf
  notify: Restart nginx

handlers:
  - name: Restart nginx
    service:
      name: nginx
      state: restarted   # Only runs if template changed
```

### Trap 4: The "latest" trap

```yaml
# ⚠️ DANGEROUS — changes every time a new version is released
- name: Install nginx
  apt:
    name: nginx
    state: latest         # Upgrades if newer version available

# ✅ SAFER — only installs, doesn't upgrade
- name: Install nginx
  apt:
    name: nginx
    state: present        # Install if missing, don't touch if present
```

---

## 🔴 The Real-World Idempotency Test

```yaml
# A truly idempotent playbook passes this test:
#
# 1. Start from scratch → ansible-playbook site.yml → works
# 2. Run again          → ansible-playbook site.yml → zero changes  
# 3. Manually break something (delete a file, stop a service)
# 4. Run again          → ansible-playbook site.yml → fixes ONLY what's broken
# 5. Run again          → ansible-playbook site.yml → zero changes
#
# If any step fails, your playbook has an idempotency bug.
```

### Checklist for Idempotent Playbooks

```
✅ Use modules instead of shell/command when possible
✅ Use creates/removes for command tasks
✅ Use changed_when for shell tasks that are read-only
✅ Use handlers for service restarts
✅ Use state: present instead of state: latest
✅ Use lineinfile with regexp for config modifications
✅ Use template/copy instead of shell echo/cat
✅ Test by running twice — second run = zero changes
```

---

**Previous:** [04. Variables and Templates](./04-variables-and-templates.md)  
**Next:** [06. Ansible + Terraform](./06-ansible-plus-terraform.md)
