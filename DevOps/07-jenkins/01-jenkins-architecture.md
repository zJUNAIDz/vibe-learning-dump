# Jenkins Architecture

> **Jenkins isn't a single program. It's a distributed system — a controller that orchestrates agents.**

---

## 🟢 The Big Picture

```
┌──────────────────────────────────────────────────────────────────┐
│                     JENKINS CONTROLLER                           │
│                                                                  │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────────┐    │
│  │ Web UI   │  │ REST API │  │ Scheduler│  │ Plugin Engine│    │
│  │ (Port    │  │ (JSON/   │  │ (Picks   │  │ (1800+ avail)│    │
│  │  8080)   │  │  XML)    │  │  jobs)   │  │              │    │
│  └──────────┘  └──────────┘  └──────────┘  └──────────────┘    │
│                         │                                        │
│                    Build Queue                                   │
│              ┌──┬──┬──┬──┬──┐                                   │
│              │J1│J2│J3│J4│J5│                                   │
│              └──┴──┴──┴──┴──┘                                   │
└──────────────────┬───────────────────────────────────────────────┘
                   │ Distributes work
        ┌──────────┼──────────────┐
        ▼          ▼              ▼
   ┌─────────┐ ┌─────────┐  ┌─────────┐
   │ Agent 1 │ │ Agent 2 │  │ Agent 3 │
   │ (Linux) │ │ (Docker)│  │ (K8s)   │
   │         │ │         │  │         │
   │ 2 exec  │ │ 4 exec  │  │ dynamic │
   └─────────┘ └─────────┘  └─────────┘
```

---

## 🟢 Controller (Master)

The Jenkins **controller** is the brain. It does NOT (ideally) run your builds.

**What the controller does:**
- Serves the web UI
- Manages configuration (job definitions, credentials, plugins)
- Schedules builds
- Distributes work to agents
- Stores build results and logs
- Manages the plugin lifecycle

**What the controller should NOT do:**
- Run builds directly (security risk + performance bottleneck)
- Store large artifacts (use external artifact storage)
- Be a single point of failure without backups

```
# Jenkins home directory on the controller
$JENKINS_HOME/
├── config.xml              # Global Jenkins config
├── jobs/                   # Job definitions
│   ├── my-app/
│   │   ├── config.xml      # Job config
│   │   └── builds/         # Build history
│   │       ├── 1/
│   │       ├── 2/
│   │       └── lastSuccessfulBuild → 2
│   └── another-app/
├── plugins/                # Installed plugins (.jpi files)
├── users/                  # User configs
├── secrets/                # Encrypted secrets
├── nodes/                  # Agent configurations
└── workspace/              # Build workspaces (if running locally)
```

### Why You Should Never Build on the Controller

```
Scenario: controller runs builds directly

1. Build runs `npm install` → downloads 500MB of node_modules
2. Build runs tests → uses 2GB of RAM
3. Build leaks memory → Jenkins controller OOMs
4. Controller crashes → ALL builds stop, UI is down
5. No one can see what happened because the logs are gone too

Result: Single point of failure + no isolation
```

---

## 🟢 Agents (Nodes)

Agents are **worker machines** that execute the actual builds.

### Types of Agents

| Type | How It Works | Best For |
|------|-------------|----------|
| **Permanent Agent** | A server with Jenkins agent installed, always connected | Dedicated build servers |
| **SSH Agent** | Controller SSHes into a machine to launch agent | Linux/Mac machines |
| **JNLP Agent** | Agent connects to controller (outbound connection) | Behind firewalls |
| **Docker Agent** | Spins up a Docker container per build | Isolated, clean builds |
| **Kubernetes Agent** | Creates a pod per build in a K8s cluster | Scalable, dynamic |

### Permanent vs Dynamic Agents

```
Permanent Agents:
┌──────────────────────────┐
│ Agent Machine             │
│                          │
│  Build #101              │
│  Build #102              │  ← Same machine, contamination possible
│  Build #103              │
│  leftover from #99...    │  ← Old files still here!
│                          │
└──────────────────────────┘

Dynamic Agents (Docker/K8s):
┌──────────────────────────┐
│ Build #101               │  ← Fresh container, clean environment
│ (destroyed after build)  │
└──────────────────────────┘

┌──────────────────────────┐
│ Build #102               │  ← Fresh container, no contamination
│ (destroyed after build)  │
└──────────────────────────┘
```

**Dynamic agents are almost always better.** They provide:
- Clean environment per build (no leftover files)
- Isolation (one build can't affect another)
- Scalability (spin up more as needed, scale down when idle)
- Security (compromised build can't affect other builds)

### Executors

An **executor** is a slot on an agent that can run one build at a time.

```
Agent with 2 executors:
┌─────────────────────────┐
│       Agent 1            │
│  ┌──────────┐ ┌────────┐│
│  │Executor 1│ │Executor│ │
│  │          │ │   2    │ │
│  │ Build A  │ │Build B │ │
│  │ (running)│ │(running)│ │
│  └──────────┘ └────────┘ │
│                          │
│  Build C → in queue      │
│  (waiting for executor)  │
└─────────────────────────┘
```

**How many executors per agent?**
- Rule of thumb: 1 executor per CPU core
- CPU-intensive builds: fewer executors
- I/O-heavy builds: more executors
- Start with `cores - 1` and adjust

---

## 🟢 Plugins — The Jenkins Ecosystem

Jenkins without plugins is almost useless. Plugins are everything.

### Essential Plugins

| Plugin | What It Does |
|--------|-------------|
| **Pipeline** | Jenkinsfile support (pipeline as code) |
| **Git** | Clone repos from Git |
| **Docker Pipeline** | Build and use Docker in pipelines |
| **Kubernetes** | Dynamic agents in K8s |
| **Credentials Binding** | Inject secrets into builds securely |
| **Blue Ocean** | Modern UI for pipelines |
| **Job DSL** | Define jobs as Groovy code |
| **Matrix Authorization** | Fine-grained permissions |
| **Timestamper** | Add timestamps to console output |
| **Warnings Next Gen** | Parse and display build warnings |

### The Plugin Problem

```
Jenkins Plugin Dependencies:

pipeline-model-definition:2.2150.1
  ├── workflow-cps:3700.v
  │   ├── workflow-api:1283.v
  │   │   └── scm-api:683.v
  │   └── script-security:1326.v
  ├── pipeline-model-api:2.2150.1
  │   └── workflow-api:1283.v          ← same dependency, different requirement?
  └── credentials-binding:604.v
      └── credentials:1311.v
          └── ??? → 💥 version conflict
```

**Plugin hell is real:**
- Plugins depend on other plugins
- Version conflicts happen regularly
- Updating one plugin can break five others
- Some plugins are abandoned (no maintainer)

**Best practices:**
1. Keep plugins to a minimum — only install what you actually use
2. Test plugin updates in a staging Jenkins first
3. Use the `plugin-installation-manager-tool` for declarative plugin management
4. Pin plugin versions in your configuration

```bash
# plugins.txt — Declarative plugin list
pipeline-model-definition:2.2150.1
git:5.2.0
docker-workflow:572.v950f58993843
kubernetes:4029.v5712230ccb_f8
credentials-binding:604.vb_64480b_c56d8
```

---

## 🟡 How a Build Flows Through Jenkins

```
1. TRIGGER
   ├── Git push (webhook)
   ├── Timer (cron)
   ├── Manual (click "Build Now")
   └── Another job (upstream trigger)
        │
        ▼
2. CONTROLLER RECEIVES TRIGGER
   ├── Loads job configuration
   ├── Reads Jenkinsfile from repo
   └── Places build in queue
        │
        ▼
3. QUEUE → AGENT ASSIGNMENT
   ├── Controller finds matching agent (labels, availability)
   ├── Agent has free executor → assign
   └── No free executor → wait in queue
        │
        ▼
4. WORKSPACE SETUP
   ├── Creates workspace directory on agent
   ├── Clones Git repo into workspace
   └── Injects credentials (if configured)
        │
        ▼
5. PIPELINE EXECUTION
   ├── Runs stages sequentially (or parallel)
   ├── Each step is a command (sh, bat, tool, etc.)
   ├── Captures stdout/stderr as build log
   └── If any step fails → pipeline stops (unless catchError)
        │
        ▼
6. POST-BUILD
   ├── Archive artifacts
   ├── Publish test results
   ├── Send notifications (Slack, email)
   ├── Trigger downstream jobs
   └── Clean up workspace
        │
        ▼
7. RESULT
   ├── SUCCESS (green)
   ├── UNSTABLE (yellow — tests failed but build succeeded)
   ├── FAILURE (red)
   └── ABORTED (grey — manually cancelled)
```

---

## 🟡 Labels and Agent Matching

Labels let you control **which agent** runs **which build**.

```groovy
// Jenkinsfile — require a specific agent type
pipeline {
    agent { label 'linux && docker' }
    // This build will ONLY run on agents that have BOTH labels
    stages {
        stage('Build') {
            steps {
                sh 'docker build -t myapp .'
            }
        }
    }
}
```

```groovy
// Different agents for different stages
pipeline {
    agent none   // Don't assign a global agent
    stages {
        stage('Build') {
            agent { label 'linux' }
            steps {
                sh 'go build -o myapp .'
            }
        }
        stage('Test on Linux') {
            agent { label 'linux' }
            steps {
                sh './myapp test'
            }
        }
        stage('Test on Mac') {
            agent { label 'macos' }
            steps {
                sh './myapp test'
            }
        }
    }
}
```

---

## 🟡 Jenkins Controller High Availability

For production Jenkins, you need to worry about:

### Backup Strategy

```bash
# What to back up:
$JENKINS_HOME/config.xml              # Global config
$JENKINS_HOME/jobs/*/config.xml       # ALL job configs
$JENKINS_HOME/users/                  # User data
$JENKINS_HOME/secrets/                # Encryption keys
$JENKINS_HOME/plugins/*.jpi           # Plugin files
$JENKINS_HOME/nodes/                  # Agent configs

# What NOT to back up (too large, regenerable):
$JENKINS_HOME/jobs/*/builds/          # Build history (optional)
$JENKINS_HOME/workspace/              # Workspaces (regenerated on build)
$JENKINS_HOME/plugins/*/              # Plugin extracted dirs (regenerated)
```

### Configuration as Code (JCasC)

Instead of configuring Jenkins through the UI, use YAML:

```yaml
# jenkins.yaml — Jenkins Configuration as Code
jenkins:
  systemMessage: "Jenkins configured via JCasC"
  numExecutors: 0              # Don't run builds on controller!
  securityRealm:
    ldap:
      configurations:
        - server: "ldap.company.com"
  authorizationStrategy:
    roleBased:
      roles:
        global:
          - name: "admin"
            permissions:
              - "Overall/Administer"
          - name: "developer"
            permissions:
              - "Job/Build"
              - "Job/Read"
  nodes:
    - permanent:
        labelString: "linux docker"
        name: "build-agent-01"
        numExecutors: 4
        remoteFS: "/var/jenkins"
        launcher:
          ssh:
            host: "agent-01.internal"
            credentialsId: "agent-ssh-key"

credentials:
  system:
    domainCredentials:
      - credentials:
          - usernamePassword:
              id: "github-creds"
              username: "jenkins-bot"
              password: "${GITHUB_TOKEN}"
```

**Benefits of JCasC:**
- Jenkins config is in Git (version-controlled, reviewable)
- Reproducible setup (spin up a new Jenkins in minutes)
- No more "who changed that setting in the UI?"
- Easy to audit

---

## 🔴 Common Architecture Anti-Patterns

### Anti-Pattern 1: Running Builds on the Controller

```
❌ Don't do this:
pipeline {
    agent any    // May end up on the controller!
    stages { ... }
}

✅ Do this:
pipeline {
    agent { label 'build-agent' }
    stages { ... }
}
```

### Anti-Pattern 2: Snowflake Agents

```
❌ Agent configured manually:
- SSH in, install Java 11
- Install Node 18
- Install Docker
- Install custom certificates
- Two months later: "Why does this agent have Java 8 now?"

✅ Agent from code:
- Docker image defines all tools
- Or Ansible/Terraform provisions agents
- Any agent can be destroyed and rebuilt in minutes
```

### Anti-Pattern 3: No Pipeline as Code

```
❌ Jobs configured in the UI:
- Click "New Item"
- Fill in form
- Add build steps
- "Who configured this job? Nobody knows."

✅ Jenkinsfile in the repo:
- Pipeline defined alongside the code
- Version controlled
- Code-reviewed
- Team owns their pipeline
```

### Anti-Pattern 4: Too Many Plugins

```
❌ 150 plugins installed:
- 30 have security vulnerabilities
- 20 are abandoned
- 10 conflict with each other
- Every update is a gamble

✅ Minimal plugin set:
- Only install what you need
- Review monthly
- Remove unused plugins
- Test updates in staging
```

---

## 🔴 Jenkins vs Modern Alternatives

| Feature | Jenkins | GitHub Actions | GitLab CI |
|---------|---------|---------------|-----------|
| Hosting | Self-hosted | Cloud | Cloud or self-hosted |
| Config | Jenkinsfile (Groovy) | YAML | YAML |
| Setup complexity | High | Low | Medium |
| Scalability | Manual | Automatic | Semi-automatic |
| Plugin ecosystem | Huge (1800+) | Growing | Built-in |
| Learning curve | Steep | Gentle | Moderate |
| Cost | Free + infra costs | Per-minute (free tier) | Per-minute (free tier) |
| When to use | Maximum flexibility, legacy orgs | GitHub repos | GitLab repos |

**Jenkins is still king when:**
- You need maximum customization
- You have complex, legacy pipelines
- You need to run in a private network
- You have strict compliance requirements
- You already have Jenkins expertise

**Jenkins is overkill when:**
- You're using GitHub and want simple CI
- You have < 10 developers
- You don't need extreme customization
- You want zero infrastructure management

---

**Next:** [02. Jenkinsfile (Pipeline as Code)](./02-jenkinsfile-pipeline-as-code.md)
