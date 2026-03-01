# K9s Advanced Features

> **Once you've mastered the basics, K9s has deeper capabilities that make cluster management effortless**

---

## 🟡 Custom Skins (Theming)

K9s supports custom color schemes via skins.

### Skin Configuration

```bash
# K9s config directory
~/.config/k9s/

# Create or edit skin file
mkdir -p ~/.config/k9s/skins
vim ~/.config/k9s/skins/my-theme.yaml
```

**Example skin:**

```yaml
# ~/.config/k9s/skins/my-theme.yaml
k9s:
  body:
    fgColor: white
    bgColor: black
    logoColor: cyan
  prompt:
    fgColor: white
    bgColor: black
    suggestColor: green
  info:
    fgColor: yellow
    sectionColor: green
  dialog:
    fgColor: white
    bgColor: black
    buttonFgColor: black
    buttonBgColor: cyan
    buttonFocusFgColor: white
    buttonFocusBgColor: blue
    labelFgColor: yellow
    fieldFgColor: white
  frame:
    border:
      fgColor: white
      focusColor: cyan
    menu:
      fgColor: white
      keyColor: cyan
      numKeyColor: green
    crumbs:
      fgColor: black
      bgColor: cyan
      activeColor: yellow
    status:
      newColor: green
      modifyColor: yellow
      addColor: cyan
      errorColor: red
      highlightColor: orange
      killColor: darkgray
      completedColor: gray
    title:
      fgColor: white
      bgColor: blue
      highlightColor: cyan
      counterColor: yellow
      filterColor: green
  views:
    table:
      fgColor: white
      bgColor: black
      cursorFgColor: black
      cursorBgColor: cyan
      headerFgColor: cyan
      headerBgColor: black
    yaml:
      keyColor: cyan
      colonColor: white
      valueColor: white
    logs:
      fgColor: white
      bgColor: black
      indicator:
        fgColor: white
        bgColor: blue
```

### Apply the Skin

```yaml
# ~/.config/k9s/config.yaml
k9s:
  ui:
    skin: my-theme    # Name of your skin file (without .yaml)
```

### Built-in Skins

K9s comes with several built-in skins. Check the [K9s skins repository](https://github.com/derailed/k9s/tree/master/skins) for options:

- `dracula` — Dark theme with purple accents
- `monokai` — Classic Monokai colors
- `one-dark` — Atom One Dark inspired
- `gruvbox` — Warm retro colors
- `nord` — Arctic-inspired pastel palette

Apply any built-in skin:

```yaml
k9s:
  ui:
    skin: dracula
```

---

## 🟡 Custom Aliases

Define keyboard shortcuts for frequently used resources and commands.

### Alias Configuration

```yaml
# ~/.config/k9s/aliases.yaml
aliases:
  # Short aliases for resources
  dp: deployments
  sec: secrets
  cm: configmaps
  np: networkpolicies
  pv: persistentvolumes
  pvc: persistentvolumeclaims
  ing: ingresses
  ep: endpoints
  cj: cronjobs
  ds: daemonsets
  sts: statefulsets
  hpa: horizontalpodautoscalers
  sa: serviceaccounts
  rb: rolebindings
  crb: clusterrolebindings

  # Quick access for common views
  pp: pods             # even shorter
  ss: services
  nn: nodes
```

### Using Aliases

```
:dp          # Opens deployments (instead of :deploy)
:sec         # Opens secrets (instead of :secrets)
:np          # Opens network policies
:cj          # Opens cronjobs
```

---

## 🟡 Custom Hotkeys

Define function key shortcuts that jump directly to a specific resource view.

```yaml
# ~/.config/k9s/hotkeys.yaml
hotKeys:
  # Shift+1 → Jump to pods in production
  shift-1:
    shortCut: Shift-1
    description: Production Pods
    command: pods
    namespace: production

  # Shift+2 → Jump to deployments in production
  shift-2:
    shortCut: Shift-2
    description: Production Deployments
    command: deployments
    namespace: production

  # Shift+3 → Jump to services in all namespaces
  shift-3:
    shortCut: Shift-3
    description: All Services
    command: services
    namespace: all

  # Shift+4 → Jump to events (warnings)
  shift-4:
    shortCut: Shift-4
    description: Warning Events
    command: events
```

---

## 🟡 Custom Plugins

K9s supports plugins that add custom actions to resources via keyboard shortcuts.

### Plugin Configuration

```yaml
# ~/.config/k9s/plugins.yaml
plugins:
  # Ctrl+L → Open logs in stern (if installed)
  stern:
    shortCut: Ctrl-L
    description: "Stern logs"
    scopes:
      - pods
    command: stern
    background: false
    args:
      - --tail
      - "50"
      - $NAME
      - -n
      - $NAMESPACE

  # Shift+T → Open a shell with extra debug tools
  debug-shell:
    shortCut: Shift-T
    description: "Debug shell (netshoot)"
    scopes:
      - pods
    command: kubectl
    background: false
    args:
      - debug
      - -it
      - $NAME
      - -n
      - $NAMESPACE
      - --image=nicolaka/netshoot
      - --
      - /bin/bash

  # Ctrl+H → Show pod resource usage (via top)
  pod-top:
    shortCut: Ctrl-H
    description: "Pod resource usage"
    scopes:
      - pods
    command: kubectl
    background: false
    args:
      - top
      - pod
      - $NAME
      - -n
      - $NAMESPACE
      - --containers

  # Shift+R → Restart deployment by doing rollout restart
  rollout-restart:
    shortCut: Shift-R
    description: "Rollout restart"
    scopes:
      - deployments
    command: kubectl
    background: false
    args:
      - rollout
      - restart
      - deployment/$NAME
      - -n
      - $NAMESPACE

  # Ctrl+G → Get events for a specific pod
  pod-events:
    shortCut: Ctrl-G
    description: "Pod events"
    scopes:
      - pods
    command: kubectl
    background: false
    args:
      - get
      - events
      - --field-selector
      - involvedObject.name=$NAME
      - -n
      - $NAMESPACE
```

### Plugin Variables

| Variable | Expands To |
|----------|-----------|
| `$NAME` | Selected resource name |
| `$NAMESPACE` | Selected resource namespace |
| `$CONTEXT` | Current cluster context |
| `$CLUSTER` | Current cluster name |
| `$USER` | Current user |
| `$COL-<header>` | Column value (e.g., `$COL-IP`, `$COL-NODE`) |

---

## 🟡 Benchmarking Pods

K9s has a built-in HTTP benchmarking tool (requires hey or wrk installed).

### Setup

```bash
# Install hey (HTTP load generator)
go install github.com/rakyll/hey@latest

# Or on Fedora
sudo dnf install hey   # if available
```

### Running Benchmarks

```
:pods
# Select a pod
b            # Benchmark
```

```
┌─ Bench ─────────────────────────────┐
│                                     │
│  URL: http://localhost:8080         │
│  Method: GET                        │
│  Concurrency: [50]                  │
│  Requests: [200]                    │
│  Duration: [0s]                     │
│                                     │
│  [OK]  [Cancel]                     │
└─────────────────────────────────────┘
```

**Results:**

```
Summary:
  Total:        0.3456 secs
  Slowest:      0.0289 secs
  Fastest:      0.0012 secs
  Average:      0.0078 secs
  Requests/sec: 578.70

Response time histogram:
  0.001 [1]    |
  0.004 [45]   |■■■■■■■■■■■■■■■
  0.007 [89]   |■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
  0.010 [38]   |■■■■■■■■■■■■■
  0.013 [15]   |■■■■■
  0.016 [7]    |■■
  0.019 [3]    |■
  0.022 [1]    |
  0.029 [1]    |

Status code distribution:
  [200] 200 responses
```

---

## 🟡 RBAC Visualization

K9s can show RBAC (Role-Based Access Control) visually.

### View Cluster Roles

```
:clusterroles        # or :cr
```

### View Role Bindings

```
:rolebindings        # or :rb
:clusterrolebindings # or :crb
```

### View Who Can Do What

```
:rbac                # RBAC matrix
```

**Shows which subjects (users, groups, service accounts) have which permissions:**

```
SUBJECT           APIGROUP   RESOURCE   VERBS
admin             *          *          *
developer         apps       deployments get,list,watch,update
readonly          ""         pods       get,list,watch
ci-service-account apps      deployments get,list,create,update
```

### Policy Viewer

```
:policy              # View policies
# Select a subject
Enter                # View detailed permissions
```

---

## 🟡 Multi-Cluster Management

### Switch Contexts

```
:ctx                 # List contexts
# Select context
Enter                # Switch to it
```

### Configure Multiple Clusters in K9s

```yaml
# ~/.config/k9s/config.yaml
k9s:
  liveViewAutoRefresh: true
  refreshRate: 2                # seconds
  maxConnRetry: 5
  readOnly: false
  noExitOnCtrlC: false
  ui:
    enableMouse: false
    headless: false
    logoless: false
    crumbsless: false
    noIcons: false
    skin: ""
  clusters:
    production:
      namespace:
        active: production
        lockFavorites: true
        favorites:
          - production
          - kube-system
      view:
        active: pods
      readOnly: true           # Read-only for production!
    staging:
      namespace:
        active: default
      view:
        active: pods
      readOnly: false
    development:
      namespace:
        active: default
      readOnly: false
```

**Key safety feature:** Set `readOnly: true` for production clusters to prevent accidental changes.

---

## 🟡 File Dump and Export

### Export Resource YAML

```
:pods
# Select pod
Ctrl+S       # Save/dump YAML to file
```

**YAML is saved to:**

```
/tmp/k9s-screens-<user>/
└── <cluster>/
    └── <resource>/
        └── <name>.yaml
```

### Screen Dumps

```
Ctrl+S       # Dump current screen (any view)
```

**Useful for:**
- Saving pod state before deletion
- Exporting deployment YAML for version control
- Capturing logs for incident reports

---

## 🔴 XRay View (Resource Tree)

### Visualize Resource Dependencies

```
:xray deploy         # Show deployments as tree
:xray svc            # Show services as tree
:xray rs             # Show replica sets as tree
```

**Example output:**

```
▸ deployment/api-server
  ├── replicaset/api-server-7d8f9c
  │   ├── pod/api-server-7d8f9c-abc ✓ Running
  │   ├── pod/api-server-7d8f9c-def ✓ Running
  │   └── pod/api-server-7d8f9c-ghi ✓ Running
  ├── configmap/api-config
  ├── secret/api-secrets
  └── service/api-service
      └── endpoints/api-service
```

**This shows:**
- Which ReplicaSets belong to which Deployments
- Which Pods belong to which ReplicaSets
- Which ConfigMaps and Secrets are mounted
- Which Services route to which Pods

---

## 🟡 Popeye (Cluster Sanitizer)

K9s integrates with [Popeye](https://github.com/derailed/popeye) — a cluster linter.

### Run Popeye

```
:popeye              # Run cluster audit
```

**Output:**

```
┌─ Popeye ──────────────────────────────────────────┐
│                                                   │
│  Score: 72%                                       │
│                                                   │
│  pods (100)                                       │
│    ✓ 80 OK                                        │
│    △ 15 Warning                                   │
│    ✗  5 Error                                     │
│                                                   │
│  deployments (20)                                 │
│    ✓ 18 OK                                        │
│    △  2 Warning                                   │
│                                                   │
│  services (15)                                    │
│    ✓ 15 OK                                        │
│                                                   │
│  Warnings:                                        │
│    - Pod api-x: No resource requests/limits       │
│    - Pod web-y: No liveness probe defined         │
│    - Deploy cache: Only 1 replica (no HA)         │
│    - Pod db-z: Running as root                    │
│                                                   │
└───────────────────────────────────────────────────┘
```

**What Popeye checks:**
- Missing resource requests/limits
- Missing health probes
- Running as root
- Single replicas (no HA)
- Unused ConfigMaps/Secrets
- Over-allocated resources

---

## 🎯 K9s Configuration Summary

| File | Purpose |
|------|---------|
| `~/.config/k9s/config.yaml` | Main config (skin, clusters, refresh rate) |
| `~/.config/k9s/aliases.yaml` | Custom resource aliases |
| `~/.config/k9s/hotkeys.yaml` | Function key shortcuts |
| `~/.config/k9s/plugins.yaml` | Custom actions and commands |
| `~/.config/k9s/skins/` | Color scheme files |
| `~/.config/k9s/views.yaml` | Custom column definitions |
| `~/.config/k9s/benchmarks/` | Benchmark results |

---

## ✅ Hands-On Exercise

### Task: Customize Your K9s Setup

**1. Create a skin:**

```bash
mkdir -p ~/.config/k9s/skins
cat > ~/.config/k9s/skins/dev-theme.yaml << 'EOF'
k9s:
  body:
    fgColor: lightskyblue
    bgColor: "#1a1a2e"
    logoColor: orange
  frame:
    border:
      fgColor: lightskyblue
      focusColor: orange
    crumbs:
      fgColor: black
      bgColor: orange
    title:
      fgColor: lightskyblue
      bgColor: "#16213e"
  views:
    table:
      fgColor: lightskyblue
      bgColor: "#1a1a2e"
      cursorFgColor: black
      cursorBgColor: orange
      headerFgColor: orange
EOF
```

**2. Create aliases:**

```bash
cat > ~/.config/k9s/aliases.yaml << 'EOF'
aliases:
  pp: pods
  dp: deployments
  ss: services
  sec: secrets
  cm: configmaps
  ns: namespaces
  cj: cronjobs
  ds: daemonsets
  sts: statefulsets
  np: networkpolicies
  hpa: horizontalpodautoscalers
EOF
```

**3. Create a plugin:**

```bash
cat > ~/.config/k9s/plugins.yaml << 'EOF'
plugins:
  rollout-restart:
    shortCut: Shift-R
    description: "Rollout restart"
    scopes:
      - deployments
    command: kubectl
    background: false
    args:
      - rollout
      - restart
      - deployment/$NAME
      - -n
      - $NAMESPACE
EOF
```

**4. Apply the skin:**

```bash
# Edit or create ~/.config/k9s/config.yaml
# Add:
#   k9s:
#     ui:
#       skin: dev-theme
```

**5. Restart K9s and verify:**

```bash
k9s
# Try your aliases:  :pp  :dp  :ss
# Try hotkeys and plugins
```

---

## 📚 Summary

| Feature | How |
|---------|-----|
| Custom skins | `~/.config/k9s/skins/<name>.yaml` |
| Custom aliases | `~/.config/k9s/aliases.yaml` |
| Hotkeys | `~/.config/k9s/hotkeys.yaml` |
| Plugins | `~/.config/k9s/plugins.yaml` |
| Benchmarking | `b` key on pods |
| RBAC view | `:rbac`, `:cr`, `:rb` |
| XRay view | `:xray deploy` |
| Popeye audit | `:popeye` |
| Multi-cluster | `:ctx` + per-cluster config |
| Export YAML | `Ctrl+S` |
| Read-only mode | `--readonly` flag or per-cluster config |

**K9s is not just a viewer — it's a fully customizable Kubernetes cockpit. Invest 30 minutes configuring it, save hundreds of hours operating.**

---

**Previous:** [03. Debugging Workflows](./03-debugging-workflows.md)  
**Module:** [05. K9s](./README.md)
