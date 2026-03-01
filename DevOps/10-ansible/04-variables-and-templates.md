# Variables and Templates

> **Variables make playbooks reusable. Templates make config files dynamic. Together, they let one codebase configure staging AND production — just swap the variable file.**

---

## 🟢 Variable Types and Precedence

Ansible has **22 levels** of variable precedence. You only need to know 5:

```
Priority (highest to lowest):

  1. -e "var=value"           ← Command line (always wins)
  2. vars/ in role            ← Role internals (hard to override)
  3. playbook vars            ← In the playbook itself
  4. host_vars/hostname.yml   ← Per-host variables
  5. group_vars/groupname.yml ← Per-group variables
  6. defaults/ in role        ← Role defaults (easy to override)

Rule of thumb:
  - defaults/ → things users SHOULD customize
  - vars/     → things users should NOT touch
  - group_vars/ → environment-specific config
  - host_vars/  → machine-specific overrides
```

---

## 🟢 Defining Variables

### In Playbooks

```yaml
---
- name: Deploy application
  hosts: webservers
  
  vars:
    app_name: myapp
    app_version: "1.2.3"
    app_port: 3000
    app_env: production
    
    # Complex variables
    app_config:
      database:
        host: db.example.com
        port: 5432
        name: myapp_prod
      redis:
        host: redis.example.com
        port: 6379
  
  tasks:
    - name: Deploy {{ app_name }} version {{ app_version }}
      debug:
        msg: "Deploying to port {{ app_port }}"
```

### In group_vars

```yaml
# inventory/production/group_vars/all.yml
---
# Applied to ALL hosts in production
env: production
ntp_servers:
  - ntp1.example.com
  - ntp2.example.com
domain: example.com
monitoring_enabled: true
```

```yaml
# inventory/production/group_vars/webservers.yml
---
# Applied only to hosts in [webservers] group
http_port: 80
https_port: 443
nginx_worker_processes: auto
app_instances: 4
```

```yaml
# inventory/production/group_vars/dbservers.yml
---
# Applied only to hosts in [dbservers] group
postgresql_version: 15
postgresql_max_connections: 200
postgresql_shared_buffers: 4GB
backup_enabled: true
backup_schedule: "0 2 * * *"
```

### In host_vars

```yaml
# inventory/production/host_vars/web1.example.com.yml
---
# Applied only to web1.example.com
# Override group-level settings for this specific machine
app_instances: 8               # This server has more CPU
custom_ssl_cert: /path/to/cert # Custom cert for this host
```

---

## 🟢 Variable Data Types

```yaml
# Strings
app_name: "myapp"
greeting: 'hello'
multiline: |
  This is
  multiple lines
  preserved as-is

# Numbers
port: 3000
timeout: 30.5

# Booleans
debug_enabled: true
ssl_enabled: false

# Lists
packages:
  - nginx
  - nodejs
  - postgresql-client

# Dictionaries / Maps
database:
  host: db.example.com
  port: 5432
  name: myapp

# Nested
servers:
  web:
    count: 3
    type: t3.medium
  db:
    count: 2
    type: r5.large
```

### Accessing Variables

```yaml
tasks:
  # Simple variable
  - debug: msg="{{ app_name }}"
  
  # Dictionary access
  - debug: msg="{{ database.host }}"
  - debug: msg="{{ database['host'] }}"     # Bracket notation
  
  # List access
  - debug: msg="{{ packages[0] }}"          # First item
  
  # Nested
  - debug: msg="{{ servers.web.count }}"
```

---

## 🟢 Jinja2 Templates

Templates use Jinja2 syntax: `{{ variable }}`, `{% logic %}`, `{# comment #}`.

### Basic Template

```jinja2
{# templates/app.conf.j2 #}
# Application Configuration
# Managed by Ansible — DO NOT EDIT MANUALLY

APP_NAME={{ app_name }}
APP_PORT={{ app_port }}
APP_ENV={{ app_env }}

DATABASE_HOST={{ database.host }}
DATABASE_PORT={{ database.port }}
DATABASE_NAME={{ database.name }}
DATABASE_URL=postgresql://{{ db_user }}:{{ db_password }}@{{ database.host }}:{{ database.port }}/{{ database.name }}

{% if redis_enabled %}
REDIS_HOST={{ redis.host }}
REDIS_PORT={{ redis.port }}
{% endif %}

{% if app_env == 'production' %}
LOG_LEVEL=warn
DEBUG=false
{% else %}
LOG_LEVEL=debug
DEBUG=true
{% endif %}
```

### Filters

```jinja2
{# String filters #}
{{ hostname | upper }}                    → MYSERVER
{{ hostname | lower }}                    → myserver
{{ hostname | capitalize }}               → Myserver
{{ "  hello  " | trim }}                  → hello
{{ path | basename }}                     → file.txt
{{ path | dirname }}                      → /var/log

{# Default values #}
{{ http_port | default(80) }}             → 80 if http_port undefined
{{ feature_flag | default(false) }}       → false if undefined

{# List filters #}
{{ packages | join(', ') }}               → nginx, nodejs, git
{{ packages | length }}                   → 3
{{ packages | first }}                    → nginx
{{ packages | last }}                     → git
{{ [3, 1, 2] | sort }}                    → [1, 2, 3]
{{ items | unique }}                      → Remove duplicates

{# Math #}
{{ memory_mb | int / 1024 }}              → Convert to GB
{{ workers | default(ansible_processor_vcpus) }}

{# JSON/YAML #}
{{ config_dict | to_json }}               → JSON string
{{ config_dict | to_nice_json }}          → Pretty JSON
{{ config_dict | to_yaml }}               → YAML string

{# IP address filters #}
{{ '192.168.1.0/24' | ipaddr('network') }}
{{ ansible_default_ipv4.address | ipaddr }}

{# Hash/Crypto #}
{{ 'password' | hash('sha256') }}
{{ 'password' | password_hash('sha512') }}
```

### Loops in Templates

```jinja2
{# templates/nginx-upstreams.conf.j2 #}

{% for backend in app_backends %}
upstream {{ backend.name }} {
    {% for server in backend.servers %}
    server {{ server.host }}:{{ server.port }} weight={{ server.weight | default(1) }};
    {% endfor %}
}
{% endfor %}

{# With loop index #}
{% for server in servers %}
# Server {{ loop.index }} of {{ loop.length }}
server {{ server }};
{% if not loop.last %}
{% endif %}
{% endfor %}
```

### Conditionals in Templates

```jinja2
{# templates/haproxy.cfg.j2 #}

global
    maxconn {{ haproxy_maxconn | default(4096) }}
    {% if haproxy_stats_enabled %}
    stats socket /var/run/haproxy.sock mode 660
    {% endif %}

{% for frontend in haproxy_frontends %}
frontend {{ frontend.name }}
    bind *:{{ frontend.port }}
    {% if frontend.ssl | default(false) %}
    bind *:443 ssl crt {{ frontend.ssl_cert }}
    redirect scheme https if !{ ssl_fc }
    {% endif %}
    default_backend {{ frontend.backend }}
{% endfor %}
```

---

## 🟡 Registered Variables

Capture task output for use in later tasks:

```yaml
tasks:
  - name: Get current disk usage
    command: df -h /
    register: disk_result
    changed_when: false

  - name: Show disk usage
    debug:
      msg: "{{ disk_result.stdout }}"

  - name: Fail if disk almost full
    fail:
      msg: "Disk usage is critical!"
    when: disk_result.stdout_lines[1] | regex_search('9[0-9]%|100%')
```

### Register structure

```yaml
# Every registered variable has:
disk_result:
  changed: false
  cmd: ["df", "-h", "/"]
  rc: 0                      # Return code
  stdout: "..."              # Standard output (string)
  stdout_lines: [...]        # Standard output (list of lines)
  stderr: ""                 # Standard error
  stderr_lines: []
  failed: false
  start: "2024-01-01 ..."
  end: "2024-01-01 ..."
  delta: "0:00:00.005"
```

---

## 🟡 Conditionals (when)

```yaml
tasks:
  # OS-specific tasks
  - name: Install on Debian
    apt:
      name: nginx
      state: present
    when: ansible_os_family == "Debian"

  - name: Install on RedHat
    yum:
      name: nginx
      state: present
    when: ansible_os_family == "RedHat"

  # Variable-based
  - name: Configure SSL
    template:
      src: ssl.conf.j2
      dest: /etc/nginx/ssl.conf
    when: ssl_enabled | default(false)

  # Complex conditions
  - name: Only in production with monitoring
    include_role:
      name: monitoring
    when:
      - env == 'production'
      - monitoring_enabled | default(true)

  # Check if variable is defined
  - name: Configure custom DNS
    template:
      src: resolv.conf.j2
      dest: /etc/resolv.conf
    when: custom_dns_servers is defined
```

---

## 🟡 Loops

```yaml
tasks:
  # Simple loop
  - name: Create multiple users
    user:
      name: "{{ item }}"
      state: present
    loop:
      - alice
      - bob
      - charlie

  # Loop with dictionaries
  - name: Create users with specific shells
    user:
      name: "{{ item.name }}"
      shell: "{{ item.shell }}"
      groups: "{{ item.groups }}"
    loop:
      - { name: alice, shell: /bin/bash, groups: sudo }
      - { name: bob, shell: /bin/zsh, groups: docker }
      - { name: deploy, shell: /bin/bash, groups: "sudo,docker" }

  # Loop over dictionary
  - name: Set sysctl parameters
    sysctl:
      name: "{{ item.key }}"
      value: "{{ item.value }}"
      state: present
    loop: "{{ sysctl_params | dict2items }}"
    vars:
      sysctl_params:
        net.core.somaxconn: 65535
        vm.swappiness: 10
        net.ipv4.tcp_max_syn_backlog: 65535

  # Loop with index
  - name: Create numbered config files
    template:
      src: worker.conf.j2
      dest: "/etc/myapp/worker-{{ idx }}.conf"
    loop: "{{ range(0, worker_count) | list }}"
    loop_control:
      loop_var: idx
```

---

## 🟡 Ansible Facts

Ansible automatically gathers system facts about target hosts:

```yaml
tasks:
  - name: Show system info
    debug:
      msg: |
        Hostname: {{ ansible_hostname }}
        OS: {{ ansible_distribution }} {{ ansible_distribution_version }}
        Architecture: {{ ansible_architecture }}
        CPUs: {{ ansible_processor_vcpus }}
        Memory: {{ ansible_memtotal_mb }} MB
        IP: {{ ansible_default_ipv4.address }}
        Disk: {{ ansible_mounts[0].size_total }}
```

```bash
# View all facts for a host
ansible web1.example.com -m setup

# Filter facts
ansible web1.example.com -m setup -a "filter=ansible_distribution*"
```

### Custom Facts

```yaml
# Create a custom fact on the target
- name: Set custom facts
  copy:
    content: |
      {
        "app_name": "myapp",
        "app_version": "1.2.3",
        "deployed_at": "{{ ansible_date_time.iso8601 }}"
      }
    dest: /etc/ansible/facts.d/myapp.fact
    mode: '0644'

# Access later via: ansible_local.myapp.app_version
```

---

## 🟡 Vault (Encrypted Variables)

```bash
# Encrypt a variable file
ansible-vault encrypt inventory/production/group_vars/all/vault.yml

# Create an encrypted file
ansible-vault create secrets.yml

# Edit encrypted file
ansible-vault edit secrets.yml

# View encrypted file
ansible-vault view secrets.yml

# Run playbook with vault password
ansible-playbook site.yml --ask-vault-pass
ansible-playbook site.yml --vault-password-file ~/.vault_pass
```

### Vault Best Practice: Separate vault files

```yaml
# inventory/production/group_vars/all/vars.yml (not encrypted)
db_user: myapp
db_host: db.example.com
db_name: myapp_prod
db_password: "{{ vault_db_password }}"    # Reference encrypted var

# inventory/production/group_vars/all/vault.yml (encrypted)
vault_db_password: "s3cr3t_p@ssw0rd"
vault_api_key: "abc123..."
vault_ssl_key: |
  -----BEGIN PRIVATE KEY-----
  MIIEvQIBADANBg...
  -----END PRIVATE KEY-----
```

**Why separate files?**
```
vault.yml → Encrypted, hard to diff/review
vars.yml  → Plain text, easy to review

Pattern: vars.yml references vault.yml using vault_ prefix
  db_password: "{{ vault_db_password }}"
  
This way:
  - Code review sees which variables exist
  - Secrets never appear in plain text
  - Git diffs are meaningful for vars.yml
```

---

## 🔴 Anti-Patterns

### ❌ Secrets in plain text

```yaml
# BAD — password visible in version control
db_password: "my_secret_password"

# GOOD — use vault
db_password: "{{ vault_db_password }}"
```

### ❌ Variables scattered everywhere

```yaml
# BAD — same variable defined in 5 places
# playbook vars + group_vars + host_vars + role defaults + command line
# Nobody knows which value actually applies

# GOOD — clear variable hierarchy
# defaults/main.yml → Sensible defaults for the role
# group_vars/       → Environment-specific overrides  
# host_vars/        → Machine-specific (rare)
# -e flag           → Emergency overrides only
```

### ❌ Not using default() filter in templates

```jinja2
{# BAD — crashes if variable undefined #}
port={{ custom_port }}

{# GOOD — fallback value #}
port={{ custom_port | default(3000) }}
```

---

**Previous:** [03. Roles](./03-roles.md)  
**Next:** [05. Idempotency](./05-idempotency.md)
