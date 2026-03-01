# Ansible Roles

> **Roles are Ansible's version of libraries. Instead of one giant playbook, you split logic into reusable, testable, shareable units. A "nginx" role can be used in every project.**

---

## 🟢 Why Roles?

```
Without roles (monolithic playbook):
  site.yml → 800 lines
    - Install nginx (50 lines)
    - Configure nginx (80 lines)
    - Install PostgreSQL (60 lines)
    - Configure PostgreSQL (90 lines)
    - Deploy app (100 lines)
    - Configure monitoring (70 lines)
    ...
  
  Problems:
    - Can't reuse the nginx setup in another project
    - Hard to test individual pieces
    - One person edits, everyone gets merge conflicts
    - Variables scattered everywhere

With roles:
  site.yml → 20 lines
    roles:
      - nginx
      - postgresql
      - myapp
      - monitoring
  
  Each role is self-contained:
    roles/nginx/        → Everything about nginx
    roles/postgresql/   → Everything about PostgreSQL
    roles/myapp/        → Everything about the app
```

---

## 🟢 Role Directory Structure

```
roles/
└── nginx/
    ├── tasks/
    │   ├── main.yml          ← Entry point (auto-loaded)
    │   ├── install.yml       ← Sub-tasks (included from main)
    │   └── configure.yml
    ├── handlers/
    │   └── main.yml          ← Handlers (auto-loaded)
    ├── templates/
    │   ├── nginx.conf.j2     ← Jinja2 templates
    │   └── vhost.conf.j2
    ├── files/
    │   └── ssl-params.conf   ← Static files to copy
    ├── vars/
    │   └── main.yml          ← Role variables (high priority)
    ├── defaults/
    │   └── main.yml          ← Default variables (low priority, user can override)
    ├── meta/
    │   └── main.yml          ← Role metadata (dependencies, galaxy info)
    ├── tests/
    │   ├── inventory
    │   └── test.yml
    └── README.md
```

**What goes where:**

```
defaults/   → Variables users SHOULD override (ports, versions, paths)
vars/       → Variables users should NOT override (internal logic)
tasks/      → The actual work
handlers/   → Restart/reload services when config changes
templates/  → Config files with dynamic content (Jinja2)
files/      → Static files (certs, scripts, binaries)
meta/       → Dependencies ("this role needs 'common' role first")
```

---

## 🟢 Creating a Role

### Scaffold with ansible-galaxy

```bash
ansible-galaxy role init roles/nginx

# Creates the full directory structure
```

### Role: nginx

```yaml
# roles/nginx/defaults/main.yml
---
nginx_worker_processes: auto
nginx_worker_connections: 1024
nginx_keepalive_timeout: 65
nginx_client_max_body_size: 10m

nginx_vhosts: []
# Example:
#   - server_name: example.com
#     port: 80
#     upstream_host: localhost
#     upstream_port: 3000
#     ssl: false
```

```yaml
# roles/nginx/tasks/main.yml
---
- name: Include install tasks
  include_tasks: install.yml

- name: Include configure tasks
  include_tasks: configure.yml
```

```yaml
# roles/nginx/tasks/install.yml
---
- name: Install nginx
  apt:
    name: nginx
    state: present
    update_cache: true
    cache_valid_time: 3600

- name: Ensure nginx is started and enabled
  service:
    name: nginx
    state: started
    enabled: true
```

```yaml
# roles/nginx/tasks/configure.yml
---
- name: Configure nginx main config
  template:
    src: nginx.conf.j2
    dest: /etc/nginx/nginx.conf
    owner: root
    group: root
    mode: '0644'
    validate: nginx -t -c %s
  notify: Reload nginx

- name: Configure virtual hosts
  template:
    src: vhost.conf.j2
    dest: "/etc/nginx/sites-available/{{ item.server_name }}"
    owner: root
    group: root
    mode: '0644'
  loop: "{{ nginx_vhosts }}"
  notify: Reload nginx

- name: Enable virtual hosts
  file:
    src: "/etc/nginx/sites-available/{{ item.server_name }}"
    dest: "/etc/nginx/sites-enabled/{{ item.server_name }}"
    state: link
  loop: "{{ nginx_vhosts }}"
  notify: Reload nginx

- name: Remove default site
  file:
    path: /etc/nginx/sites-enabled/default
    state: absent
  notify: Reload nginx
```

```yaml
# roles/nginx/handlers/main.yml
---
- name: Reload nginx
  service:
    name: nginx
    state: reloaded

- name: Restart nginx
  service:
    name: nginx
    state: restarted
```

```jinja2
{# roles/nginx/templates/nginx.conf.j2 #}
user www-data;
worker_processes {{ nginx_worker_processes }};
pid /run/nginx.pid;

events {
    worker_connections {{ nginx_worker_connections }};
}

http {
    sendfile on;
    tcp_nopush on;
    tcp_nodelay on;
    keepalive_timeout {{ nginx_keepalive_timeout }};
    types_hash_max_size 2048;
    client_max_body_size {{ nginx_client_max_body_size }};

    include /etc/nginx/mime.types;
    default_type application/octet-stream;

    access_log /var/log/nginx/access.log;
    error_log /var/log/nginx/error.log;

    gzip on;

    include /etc/nginx/conf.d/*.conf;
    include /etc/nginx/sites-enabled/*;
}
```

### Using the Role

```yaml
# site.yml
---
- name: Configure web servers
  hosts: webservers
  become: true
  
  roles:
    - role: nginx
      vars:
        nginx_worker_connections: 2048
        nginx_vhosts:
          - server_name: api.example.com
            port: 80
            upstream_host: localhost
            upstream_port: 3000
```

---

## 🟡 Role Dependencies

Define in `meta/main.yml`:

```yaml
# roles/myapp/meta/main.yml
---
dependencies:
  - role: common           # Always run "common" first
  - role: nginx            # Then "nginx"
    vars:
      nginx_vhosts:
        - server_name: "{{ app_domain }}"
          port: 80
          upstream_host: localhost
          upstream_port: "{{ app_port }}"
```

```
Execution order with dependencies:
  1. common (dependency of myapp)
  2. nginx (dependency of myapp)
  3. myapp (the role you actually called)
```

---

## 🟡 Ansible Galaxy

Galaxy is Ansible's package registry for sharing roles.

```bash
# Install a role from Galaxy
ansible-galaxy install geerlingguy.docker
ansible-galaxy install geerlingguy.postgresql

# Install roles from requirements file
ansible-galaxy install -r requirements.yml

# List installed roles
ansible-galaxy list
```

### requirements.yml

```yaml
# requirements.yml
---
roles:
  # From Galaxy
  - name: geerlingguy.docker
    version: "6.1.0"

  - name: geerlingguy.postgresql
    version: "3.4.0"

  # From GitHub
  - name: custom_nginx
    src: https://github.com/myorg/ansible-role-nginx.git
    version: v2.0.0

  # From a tarball
  - name: proprietary_app
    src: https://internal.example.com/roles/app-v1.0.tar.gz

collections:
  - name: community.docker
    version: "3.4.0"
  - name: community.postgresql
    version: "3.0.0"
```

---

## 🟡 Organizing a Multi-Role Project

```
ansible-project/
├── ansible.cfg               ← Ansible configuration
├── inventory/
│   ├── production/
│   │   ├── hosts.yml
│   │   ├── group_vars/
│   │   │   ├── all.yml
│   │   │   ├── webservers.yml
│   │   │   └── dbservers.yml
│   │   └── host_vars/
│   │       └── web1.example.com.yml
│   └── staging/
│       ├── hosts.yml
│       └── group_vars/
│           └── all.yml
├── roles/
│   ├── common/              ← Base setup for all servers
│   ├── nginx/               ← Web server
│   ├── postgresql/          ← Database
│   ├── myapp/               ← Application
│   └── monitoring/          ← Prometheus node_exporter
├── playbooks/
│   ├── site.yml             ← Main entry point
│   ├── webservers.yml       ← Web server setup
│   ├── dbservers.yml        ← Database setup
│   └── deploy.yml           ← App deployment only
├── requirements.yml         ← Galaxy dependencies
└── Makefile                 ← Convenience commands
```

### ansible.cfg

```ini
# ansible.cfg
[defaults]
inventory = inventory/production/hosts.yml
roles_path = roles
retry_files_enabled = false
host_key_checking = false    # Only in dev/CI!
stdout_callback = yaml       # Readable output
forks = 20                   # Parallel connections

[privilege_escalation]
become = true
become_method = sudo

[ssh_connection]
pipelining = true            # Faster! Fewer SSH connections
ssh_args = -o ControlMaster=auto -o ControlPersist=60s
```

### Main Playbook

```yaml
# playbooks/site.yml
---
- name: Common setup for all servers
  hosts: all
  roles:
    - common

- name: Configure web servers
  hosts: webservers
  roles:
    - nginx
    - myapp

- name: Configure database servers
  hosts: dbservers
  roles:
    - postgresql

- name: Configure monitoring
  hosts: all
  roles:
    - monitoring
```

---

## 🔴 Role Anti-Patterns

### ❌ God role

```yaml
# BAD — one role does everything
roles/
└── setup_everything/
    └── tasks/
        └── main.yml    # 500+ lines: nginx + postgres + app + monitoring

# GOOD — one role per concern
roles/
├── common/
├── nginx/
├── postgresql/
├── myapp/
└── monitoring/
```

### ❌ Hardcoded values instead of defaults

```yaml
# BAD — hardcoded in tasks
- name: Configure nginx
  template:
    src: nginx.conf.j2
    dest: /etc/nginx/nginx.conf
  # Template has: worker_connections 1024; (hardcoded)

# GOOD — use defaults, template uses variables
# defaults/main.yml
nginx_worker_connections: 1024

# Template uses: worker_connections {{ nginx_worker_connections }};
# User can override: roles: [{role: nginx, vars: {nginx_worker_connections: 4096}}]
```

### ❌ Not using meta/main.yml for dependencies

```yaml
# BAD — manual ordering in site.yml
- hosts: webservers
  roles:
    - common     # "You have to remember to include this first"
    - nginx
    - myapp

# GOOD — roles declare their own dependencies
# roles/myapp/meta/main.yml
dependencies:
  - common
  - nginx
```

---

**Previous:** [02. Modules](./02-modules.md)  
**Next:** [04. Variables and Templates](./04-variables-and-templates.md)
