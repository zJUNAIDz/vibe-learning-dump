# Ansible Basics

> **Ansible is SSH + YAML + Python. That's it. No agents, no daemons, no complex setup. Push a playbook, configure a server.**

---

## 🟢 What Is Ansible?

```
Traditional server configuration:
  1. SSH into server
  2. Run commands manually
  3. Forget what you did
  4. Repeat for 50 servers (slightly differently each time)

Ansible:
  1. Write a playbook (YAML file)
  2. ansible-playbook site.yml
  3. All 50 servers configured identically
  4. Run it again → nothing changes (idempotent)
```

### How Ansible Works

```
┌──────────────────┐
│  Control Machine │    Your laptop or CI server
│  (has Ansible)   │    
└────────┬─────────┘
         │ SSH (no agent needed on target!)
    ┌────┼────────────────────┐
    ▼    ▼                    ▼
┌───────┐ ┌───────┐    ┌───────┐
│Server │ │Server │    │Server │
│  1    │ │  2    │    │  N    │
│       │ │       │    │       │
│(just  │ │(just  │    │(just  │
│ SSH   │ │ SSH   │    │ SSH   │
│ +     │ │ +     │    │ +     │
│Python)│ │Python)│    │Python)│
└───────┘ └───────┘    └───────┘

Ansible:
  1. Connects via SSH
  2. Copies Python modules to the target
  3. Executes modules
  4. Collects results
  5. Disconnects
  
  No agent installed. No daemon running. Just SSH + Python.
```

---

## 🟢 Inventory Files

An **inventory** defines WHICH servers Ansible should manage.

### INI-Style Inventory

```ini
# inventory/hosts.ini

# Individual hosts
web1.example.com
web2.example.com
db1.example.com

# Groups
[webservers]
web1.example.com
web2.example.com
web3.example.com ansible_port=2222    # Custom SSH port

[dbservers]
db1.example.com
db2.example.com

[loadbalancers]
lb1.example.com

# Group of groups
[production:children]
webservers
dbservers
loadbalancers

# Variables for a group
[webservers:vars]
http_port=80
max_connections=1000

[dbservers:vars]
db_port=5432
```

### YAML Inventory (More Flexible)

```yaml
# inventory/hosts.yml
all:
  children:
    production:
      children:
        webservers:
          hosts:
            web1.example.com:
              http_port: 80
            web2.example.com:
              http_port: 80
          vars:
            max_connections: 1000
        
        dbservers:
          hosts:
            db1.example.com:
              db_port: 5432
            db2.example.com:
              db_port: 5432
    
    staging:
      children:
        webservers:
          hosts:
            staging-web1.example.com:
```

---

## 🟢 Playbooks

A playbook is a YAML file that describes what to do on which servers.

### Basic Playbook

```yaml
# site.yml
---
- name: Configure web servers
  hosts: webservers
  become: true              # Run as root (sudo)
  
  tasks:
    - name: Install nginx
      apt:
        name: nginx
        state: present
        update_cache: true
    
    - name: Start nginx
      service:
        name: nginx
        state: started
        enabled: true        # Start on boot
    
    - name: Copy nginx config
      template:
        src: templates/nginx.conf.j2
        dest: /etc/nginx/sites-available/default
        owner: root
        group: root
        mode: '0644'
      notify: Restart nginx       # Trigger handler if this changes
  
  handlers:
    - name: Restart nginx
      service:
        name: nginx
        state: restarted
```

### Running a Playbook

```bash
# Run against inventory
ansible-playbook -i inventory/hosts.ini site.yml

# Dry run (check mode — show what WOULD change)
ansible-playbook -i inventory/hosts.ini site.yml --check

# Run with verbose output
ansible-playbook -i inventory/hosts.ini site.yml -v
ansible-playbook -i inventory/hosts.ini site.yml -vvv  # Extra verbose

# Limit to specific hosts
ansible-playbook -i inventory/hosts.ini site.yml --limit web1.example.com

# Run with extra variables
ansible-playbook -i inventory/hosts.ini site.yml -e "http_port=8080"
```

---

## 🟢 Tasks

Tasks are individual actions. Each task calls one **module**.

```yaml
tasks:
  # Package management
  - name: Install packages
    apt:
      name:
        - nginx
        - nodejs
        - postgresql-client
      state: present
      update_cache: true

  # File operations
  - name: Create directory
    file:
      path: /opt/myapp
      state: directory
      owner: appuser
      group: appuser
      mode: '0755'

  # Copy files
  - name: Copy application config
    copy:
      src: files/app.conf
      dest: /opt/myapp/config.yml
      owner: appuser
      mode: '0600'         # Only owner can read (secrets!)

  # Download files
  - name: Download binary
    get_url:
      url: https://releases.example.com/myapp-v1.2.3-linux-amd64
      dest: /opt/myapp/myapp
      mode: '0755'
      checksum: sha256:abc123...

  # Run commands
  - name: Run database migration
    command: /opt/myapp/myapp migrate
    args:
      chdir: /opt/myapp

  # Service management
  - name: Enable and start service
    systemd:
      name: myapp
      state: started
      enabled: true
      daemon_reload: true

  # User management
  - name: Create application user
    user:
      name: appuser
      system: true
      shell: /usr/sbin/nologin
      home: /opt/myapp
```

---

## 🟢 Handlers

Handlers are tasks that **only run when notified** — and only run **once** at the end, even if notified multiple times.

```yaml
tasks:
  - name: Update nginx config
    template:
      src: nginx.conf.j2
      dest: /etc/nginx/nginx.conf
    notify: Restart nginx           # ← Trigger handler

  - name: Update SSL certificate
    copy:
      src: files/ssl.crt
      dest: /etc/nginx/ssl/cert.crt
    notify: Restart nginx           # ← Same handler triggered again

  # Handler only runs ONCE at the end, even though notified twice

handlers:
  - name: Restart nginx
    service:
      name: nginx
      state: restarted
```

**Why handlers instead of regular tasks?**
```
Without handlers:
  1. Update config → restart nginx
  2. Update cert   → restart nginx
  Result: nginx restarted TWICE (unnecessary, may cause brief downtime)

With handlers:
  1. Update config → notify "restart nginx"
  2. Update cert   → notify "restart nginx"
  3. End of play   → restart nginx ONCE
  Result: nginx restarted once, only if something actually changed
```

---

## 🟡 Ad-Hoc Commands

For quick one-off tasks, use ad-hoc commands instead of playbooks.

```bash
# Ping all servers
ansible all -i inventory/hosts.ini -m ping

# Check uptime
ansible webservers -i inventory/hosts.ini -m command -a "uptime"

# Install a package
ansible webservers -i inventory/hosts.ini -m apt -a "name=htop state=present" --become

# Copy a file
ansible webservers -i inventory/hosts.ini -m copy -a "src=./fix.sh dest=/tmp/fix.sh"

# Restart a service
ansible webservers -i inventory/hosts.ini -m service -a "name=nginx state=restarted" --become

# Get facts about a server
ansible web1.example.com -i inventory/hosts.ini -m setup
```

---

## 🟡 Ansible vs Shell Scripts

```
Shell Script:
  #!/bin/bash
  apt-get update
  apt-get install -y nginx
  # Run again → "nginx is already the newest version" (works)
  
  mkdir -p /opt/myapp
  # Run again → already exists (works)
  
  useradd appuser
  # Run again → "useradd: user 'appuser' already exists" ERROR! 💥

Ansible:
  - apt: name=nginx state=present      # Idempotent ✅
  - file: path=/opt/myapp state=directory   # Idempotent ✅
  - user: name=appuser state=present   # Idempotent ✅
  # Run again → "ok" (no changes needed)
```

| Feature | Shell Scripts | Ansible |
|---------|-------------|---------|
| Idempotent | Must code carefully | Built-in |
| Multi-server | Loops + SSH | Built-in |
| Error handling | set -e (fragile) | Per-task, with rescue |
| Dry run | Not possible | `--check` mode |
| Readability | Medium | High (YAML) |
| Reporting | Manual | Built-in (changed/ok/failed) |

---

**Previous:** [README](./README.md)  
**Next:** [02. Modules](./02-modules.md)
