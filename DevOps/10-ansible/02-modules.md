# Ansible Modules

> **Modules are the actual tools Ansible uses. `apt` installs packages. `copy` copies files. `service` manages services. Each module knows how to be idempotent — you don't have to.**

---

## 🟢 What Are Modules?

Modules are small programs that Ansible pushes to target servers, executes, and then removes. Each module:
1. Accepts parameters
2. Checks current state
3. Makes changes only if needed (idempotent)
4. Returns JSON results (changed/ok/failed)

```
Think of modules as specialized tools:
  apt     → Package installer
  copy    → File copier
  service → Service manager
  user    → User manager
  
  You don't write HOW to install a package.
  You say WHAT package should be present.
  The module figures out the rest.
```

---

## 🟢 Package Management Modules

### apt (Debian/Ubuntu)

```yaml
tasks:
  # Install a single package
  - name: Install nginx
    apt:
      name: nginx
      state: present          # Ensure installed (don't care about version)
      update_cache: true      # apt-get update first
      cache_valid_time: 3600  # Don't update cache if updated in last hour

  # Install specific version
  - name: Install Node.js 18
    apt:
      name: nodejs=18.*
      state: present

  # Install multiple packages
  - name: Install build dependencies
    apt:
      name:
        - build-essential
        - curl
        - git
        - python3-pip
        - unzip
      state: present

  # Remove a package
  - name: Remove Apache (we use nginx)
    apt:
      name: apache2
      state: absent           # Ensure NOT installed
      purge: true             # Also remove config files

  # Upgrade all packages
  - name: Upgrade all packages
    apt:
      upgrade: safe           # safe = only upgrade, don't remove
      update_cache: true
```

### yum/dnf (RHEL/CentOS/Fedora)

```yaml
tasks:
  - name: Install packages on RHEL
    yum:
      name:
        - nginx
        - git
      state: present

  # Using dnf (Fedora/RHEL 8+)
  - name: Install with dnf
    dnf:
      name: nginx
      state: latest           # Install OR upgrade to latest
```

### package (Distro-Agnostic)

```yaml
tasks:
  # Works on both Debian and RHEL
  - name: Install git (any distro)
    package:
      name: git
      state: present
```

---

## 🟢 File Operation Modules

### copy

```yaml
tasks:
  # Copy from control machine to remote
  - name: Copy application config
    copy:
      src: files/app.conf            # Local path (on control machine)
      dest: /etc/myapp/app.conf      # Remote path
      owner: appuser
      group: appuser
      mode: '0644'
      backup: true                    # Create backup of existing file

  # Create file with inline content
  - name: Create health check endpoint
    copy:
      content: |
        {
          "status": "ok",
          "version": "{{ app_version }}"
        }
      dest: /var/www/health.json
      mode: '0644'
```

### template (Jinja2 Templates)

```yaml
tasks:
  # Render template with variables
  - name: Configure nginx virtual host
    template:
      src: templates/nginx-vhost.conf.j2
      dest: /etc/nginx/sites-available/{{ domain }}
      owner: root
      group: root
      mode: '0644'
      validate: nginx -t -c %s       # Validate before applying!
    notify: Reload nginx
```

```jinja2
{# templates/nginx-vhost.conf.j2 #}
server {
    listen {{ http_port | default(80) }};
    server_name {{ domain }};
    
    location / {
        proxy_pass http://{{ app_host }}:{{ app_port }};
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
    
    {% if ssl_enabled %}
    listen 443 ssl;
    ssl_certificate /etc/nginx/ssl/{{ domain }}.crt;
    ssl_certificate_key /etc/nginx/ssl/{{ domain }}.key;
    {% endif %}
}
```

### file

```yaml
tasks:
  # Create directory
  - name: Create app directory
    file:
      path: /opt/myapp/logs
      state: directory
      owner: appuser
      group: appuser
      mode: '0755'
      recurse: true           # Apply ownership recursively

  # Create symlink
  - name: Link current release
    file:
      src: /opt/myapp/releases/v1.2.3
      dest: /opt/myapp/current
      state: link

  # Delete file
  - name: Remove old config
    file:
      path: /etc/myapp/old.conf
      state: absent

  # Set permissions
  - name: Secure private key
    file:
      path: /etc/ssl/private/server.key
      mode: '0600'
      owner: root
      group: root
```

### lineinfile / blockinfile

```yaml
tasks:
  # Ensure a line exists in a file
  - name: Allow port forwarding in SSH
    lineinfile:
      path: /etc/ssh/sshd_config
      regexp: '^#?AllowTcpForwarding'
      line: 'AllowTcpForwarding yes'
    notify: Restart sshd

  # Add a block of text
  - name: Add custom iptables rules
    blockinfile:
      path: /etc/rc.local
      marker: "# {mark} ANSIBLE MANAGED - firewall rules"
      block: |
        iptables -A INPUT -p tcp --dport 80 -j ACCEPT
        iptables -A INPUT -p tcp --dport 443 -j ACCEPT
```

---

## 🟢 Service Management Modules

### service / systemd

```yaml
tasks:
  # Start and enable service
  - name: Ensure nginx is running
    service:
      name: nginx
      state: started
      enabled: true           # Start on boot

  # Restart service
  - name: Restart application
    service:
      name: myapp
      state: restarted

  # Reload config without restart
  - name: Reload nginx config
    service:
      name: nginx
      state: reloaded

  # Using systemd module (more features)
  - name: Deploy and start new service
    systemd:
      name: myapp
      state: started
      enabled: true
      daemon_reload: true     # Reload systemd after adding new unit file
```

### Create a systemd service file

```yaml
tasks:
  - name: Create systemd service for myapp
    copy:
      content: |
        [Unit]
        Description=My Application
        After=network.target postgresql.service
        Requires=postgresql.service
        
        [Service]
        Type=simple
        User=appuser
        Group=appuser
        WorkingDirectory=/opt/myapp/current
        ExecStart=/opt/myapp/current/myapp serve
        Restart=always
        RestartSec=5
        Environment=NODE_ENV=production
        Environment=PORT=3000
        
        # Security hardening
        NoNewPrivileges=true
        ProtectSystem=strict
        ProtectHome=true
        ReadWritePaths=/opt/myapp/data /var/log/myapp
        
        [Install]
        WantedBy=multi-user.target
      dest: /etc/systemd/system/myapp.service
      mode: '0644'
    notify: Restart myapp

  handlers:
    - name: Restart myapp
      systemd:
        name: myapp
        daemon_reload: true
        state: restarted
```

---

## 🟡 Command Execution Modules

### command vs shell vs raw

```yaml
tasks:
  # command — safe, no shell interpretation
  - name: Check app version
    command: /opt/myapp/myapp --version
    register: app_version
    changed_when: false        # This never "changes" anything

  # shell — uses /bin/sh, supports pipes and redirects
  - name: Count active connections
    shell: ss -tuln | grep ':80' | wc -l
    register: connection_count
    changed_when: false

  # raw — no Python needed on target (for bootstrapping)
  - name: Install Python on minimal server
    raw: apt-get update && apt-get install -y python3
    args:
      executable: /bin/bash
```

**When to use which:**

```
command  → Default choice. No shell features needed.
             Safe from shell injection.
             
shell    → Need pipes (|), redirects (>), globbing (*).
             Be careful with user input!
             
raw      → Target doesn't have Python yet.
             First task to bootstrap a server.
             
script   → Run a local script on the remote machine
             Copies script, executes, removes
```

### creates / removes (Idempotent Commands)

```yaml
tasks:
  # Only run if file doesn't exist
  - name: Initialize database
    command: /opt/myapp/myapp db:init
    args:
      creates: /opt/myapp/data/.initialized   # Skip if this exists

  # Only run if file DOES exist
  - name: Clean old logs
    command: rm -rf /var/log/myapp/old
    args:
      removes: /var/log/myapp/old              # Skip if this doesn't exist
```

---

## 🟡 User and Group Modules

```yaml
tasks:
  # Create system user (for running services)
  - name: Create application user
    user:
      name: appuser
      system: true                 # System user (low UID, no home by default)
      shell: /usr/sbin/nologin    # Can't login interactively
      home: /opt/myapp
      create_home: true

  # Create regular user with SSH key
  - name: Create deploy user
    user:
      name: deploy
      groups:
        - sudo
        - docker
      append: true                 # Add to groups (don't replace existing)
      shell: /bin/bash

  - name: Set SSH key for deploy user
    authorized_key:
      user: deploy
      key: "{{ lookup('file', 'files/deploy_key.pub') }}"
      state: present

  # Create group
  - name: Create app group
    group:
      name: appgroup
      state: present
```

---

## 🟡 Networking Modules

```yaml
tasks:
  # Wait for port to be available
  - name: Wait for app to start
    wait_for:
      port: 3000
      host: localhost
      delay: 5                    # Wait 5s before first check
      timeout: 60                 # Fail if not up in 60s

  # Make HTTP request
  - name: Check health endpoint
    uri:
      url: "http://localhost:3000/health"
      method: GET
      return_content: true
      status_code: 200
    register: health_check
    retries: 5
    delay: 10
    until: health_check.status == 200

  # Download file
  - name: Download application binary
    get_url:
      url: "https://releases.example.com/myapp-{{ app_version }}-linux-amd64.tar.gz"
      dest: /tmp/myapp.tar.gz
      checksum: "sha256:{{ app_checksum }}"
      mode: '0644'

  # Unarchive
  - name: Extract application
    unarchive:
      src: /tmp/myapp.tar.gz
      dest: /opt/myapp/releases/{{ app_version }}
      remote_src: true            # File is already on remote machine
```

---

## 🟡 Docker Modules

```yaml
tasks:
  # Pull Docker image
  - name: Pull application image
    docker_image:
      name: myregistry.com/myapp
      tag: "{{ app_version }}"
      source: pull

  # Run container
  - name: Run application container
    docker_container:
      name: myapp
      image: "myregistry.com/myapp:{{ app_version }}"
      state: started
      restart_policy: unless-stopped
      ports:
        - "3000:3000"
      env:
        DATABASE_URL: "{{ database_url }}"
        NODE_ENV: production
      volumes:
        - /opt/myapp/data:/app/data
      networks:
        - name: app_network

  # Docker compose
  - name: Deploy with docker-compose
    docker_compose:
      project_src: /opt/myapp
      pull: true
      state: present
```

---

## 🔴 Common Module Anti-Patterns

### ❌ Using shell for everything

```yaml
# BAD — shell for everything
- name: Install nginx
  shell: apt-get install -y nginx

- name: Create directory
  shell: mkdir -p /opt/myapp

- name: Create user
  shell: useradd -r myapp || true    # Ignore errors 🤮

# GOOD — use proper modules
- name: Install nginx
  apt:
    name: nginx
    state: present

- name: Create directory
  file:
    path: /opt/myapp
    state: directory

- name: Create user
  user:
    name: myapp
    system: true
```

### ❌ Missing changed_when for commands

```yaml
# BAD — reports "changed" every time even for read-only commands
- name: Get disk usage
  command: df -h

# GOOD — declare it doesn't change anything
- name: Get disk usage
  command: df -h
  changed_when: false
  register: disk_usage
```

### ❌ Not validating config before applying

```yaml
# BAD — broken config → service won't restart
- name: Update nginx config
  template:
    src: nginx.conf.j2
    dest: /etc/nginx/nginx.conf

# GOOD — validate before deploying
- name: Update nginx config
  template:
    src: nginx.conf.j2
    dest: /etc/nginx/nginx.conf
    validate: nginx -t -c %s       # Check syntax before replacing!
```

---

**Previous:** [01. Ansible Basics](./01-ansible-basics.md)  
**Next:** [03. Roles](./03-roles.md)
