# Common Jenkins Disasters

> **Every Jenkins admin has war stories. Learn from their pain so you don't repeat it.**

---

## 🟢 Disaster 1: Plugin Hell

### The Scenario

```
Monday:  "Jenkins says 15 plugins have updates. Let me update them all."
Monday (5 minutes later): Jenkins won't start.

Error log:
  java.lang.NoSuchMethodError: 
    'void org.jenkinsci.plugins.workflow.steps.StepExecution.applyAll'
  Caused by: Plugin 'pipeline-model-definition' requires 
    'workflow-cps' version 3700.v but installed version is 3600.v
```

### Why It Happens

```
Plugin A v2.0 requires Plugin B >= v3.0
Plugin C v4.0 requires Plugin B <= v2.5
You install both → 💥 Version conflict

Plugin Dependency Graph:
  pipeline-model-definition:2.2150
    └── workflow-cps:3700+ ────────┐
  some-other-plugin:1.5            │ CONFLICT!
    └── workflow-cps:<=3600 ───────┘
```

### Prevention

```bash
# 1. Never update all plugins at once
#    Update ONE plugin, test, then the next

# 2. Use a staging Jenkins
#    Production Jenkins: plugins locked
#    Staging Jenkins: test updates here first

# 3. Pin plugin versions
# plugins.txt
kubernetes:4029.v5712230ccb_f8
workflow-aggregator:596.v8c21c963d92d
git:5.2.0

# 4. Back up before updating
tar -czf jenkins-backup-$(date +%Y%m%d).tar.gz $JENKINS_HOME/plugins/
```

### Recovery

```bash
# Option 1: Rollback plugins
cd $JENKINS_HOME/plugins/
cp problematic-plugin.jpi.bak problematic-plugin.jpi
systemctl restart jenkins

# Option 2: Start Jenkins in safe mode (no plugins)
java -jar jenkins.war --safe

# Option 3: Remove the problematic plugin
rm $JENKINS_HOME/plugins/bad-plugin.jpi
rm -rf $JENKINS_HOME/plugins/bad-plugin/
systemctl restart jenkins
```

---

## 🟢 Disaster 2: Disk Space Exhaustion

### The Scenario

```
Tuesday 3 AM:  Jenkins stops accepting builds
Error: "java.io.IOException: No space left on device"

Investigation:
$ df -h /var/jenkins_home
Filesystem      Size  Used Avail Use%
/dev/sda1       100G  100G    0  100%

$ du -sh /var/jenkins_home/*/  | sort -rh | head
72G    /var/jenkins_home/jobs/
18G    /var/jenkins_home/workspace/
8G     /var/jenkins_home/plugins/
2G     /var/jenkins_home/war/
```

### Why It Happens

```
Every build keeps:
  - Console output log (can be 10MB+)
  - Artifacts (if archived)
  - Test results
  - Build metadata

100 builds × 50 jobs × 10MB = 50GB just in logs

Plus:
  - Docker images cached on agents
  - node_modules in workspaces
  - Go build cache
  - Downloaded dependencies
```

### Prevention

```groovy
// 1. Limit build history
pipeline {
    options {
        buildDiscarder(logRotator(
            numToKeepStr: '10',        // Keep last 10 builds
            daysToKeepStr: '30',       // Or builds from last 30 days
            artifactNumToKeepStr: '3'  // Keep artifacts from last 3
        ))
    }
}

// 2. Clean workspace after every build
post {
    always {
        cleanWs()     // Delete workspace contents
    }
}

// 3. Don't archive unnecessarily
// ❌ 
archiveArtifacts artifacts: '**/*'           // Everything!

// ✅
archiveArtifacts artifacts: 'dist/*.tar.gz'  // Only the release artifact
```

```bash
# 4. Cron job to clean old Docker images on agents
# Run weekly on each build agent
0 2 * * 0 docker system prune -af --filter "until=168h"

# 5. Monitor disk usage
# Add to Jenkins monitoring
$ du -sh $JENKINS_HOME/jobs/*/builds/ | sort -rh | head -20
```

### Emergency Recovery

```bash
# Find what's eating disk space
du -sh $JENKINS_HOME/*/ | sort -rh

# Delete old builds for a specific job
cd $JENKINS_HOME/jobs/my-heavy-job/builds/
ls -d */ | sort -n | head -50 | xargs rm -rf

# Clean ALL workspaces
rm -rf $JENKINS_HOME/workspace/*

# Clean Docker
docker system prune -af
```

---

## 🟢 Disaster 3: Credential Leaks

### The Scenario

```
Build log (visible to all team members):

[Pipeline] sh
+ echo machine github.com login jenkins-bot password ghp_Abc123SecretTokenHere
+ docker login -u admin -p SuperSecretPassword123 registry.example.com
Login Succeeded

Oops. Credentials are now in:
  - Jenkins console output (searchable)
  - Archived build logs
  - Slack notifications (if you echo build output)
  - Maybe even in Git if someone copies the Jenkinsfile
```

### Why It Happens

```groovy
// ❌ Credential exposed in sh command
sh "echo ${DOCKER_PASSWORD}"         // Printed in console

// ❌ Password in command line (visible in ps)
sh "docker login -u admin -p ${DOCKER_PASSWORD} registry.example.com"

// ❌ Credentials in Jenkinsfile (committed to Git)
environment {
    DB_PASSWORD = 'actual-password-here'
}

// ❌ Env var dump includes credentials
sh 'env | sort'                       // Prints everything including secrets
sh 'printenv'
```

### Prevention

```groovy
// ✅ Use credentials() binding — Jenkins masks the value in logs
environment {
    DOCKER_CREDS = credentials('docker-hub-creds')
}
stages {
    stage('Push') {
        steps {
            sh '''
                echo "$DOCKER_CREDS_PSW" | docker login -u "$DOCKER_CREDS_USR" --password-stdin
            '''
            // Note: single quotes prevent Groovy interpolation
            // Note: --password-stdin prevents showing password in ps output
        }
    }
}

// ✅ Use withCredentials for one-time use
steps {
    withCredentials([string(credentialsId: 'api-key', variable: 'API_KEY')]) {
        sh 'curl -H "Authorization: Bearer $API_KEY" https://api.example.com'
    }
    // $API_KEY not available outside this block
}

// ✅ Use Mask Passwords plugin
// Automatically masks known credential values in console output
```

### What to Do After a Leak

```
1. IMMEDIATELY rotate the leaked credential
   - Change the password/token
   - Revoke the old one
   - Don't wait — automated scanners find leaked creds in minutes

2. Audit where the credential was used
   - Search build logs for the credential value
   - Check if it was pushed to any artifact

3. Check for damage
   - Was the leaked cred used for unauthorized access?
   - Check audit logs of the affected service

4. Prevent recurrence
   - Add credential scanning to CI pipeline
   - Use tools like gitleaks, trufflehog
   - Enable log sanitization in Jenkins
```

---

## 🟡 Disaster 4: Zombie Builds

### The Scenario

```
Wednesday:
  - Build #47 has been running for 6 hours
  - It was supposed to take 10 minutes
  - Agent is unresponsive
  - "Abort" button doesn't work
  - Build is consuming resources
  - Queue is backing up
```

### Why It Happens

```
Common causes:
1. Process hangs waiting for input (interactive prompt)
   sh 'apt-get install nginx'    ← Waiting for Y/n

2. Infinite loop in test
   while(true) { /* oops */ }

3. Network timeout (downloading from slow mirror)
   npm install ← mirror is down, curl is retrying forever

4. Agent disconnected but build state not updated
   Agent crashed, Jenkins doesn't know

5. Docker container won't stop
   docker run with no timeout/resource limits
```

### Prevention

```groovy
// 1. Always set timeouts
pipeline {
    options {
        timeout(time: 30, unit: 'MINUTES')  // Pipeline-level
    }
    stages {
        stage('Test') {
            options {
                timeout(time: 10, unit: 'MINUTES')  // Stage-level
            }
            steps {
                sh 'npm test'
            }
        }
    }
}

// 2. Use non-interactive mode for all commands
sh 'DEBIAN_FRONTEND=noninteractive apt-get install -y nginx'
sh 'npm ci --no-optional --no-audit'      // Skip optional and audit
sh 'yes | command-that-asks-for-input'

// 3. Set timeouts on network operations
sh 'curl --max-time 30 --retry 3 https://example.com/file.tar.gz'
sh 'wget --timeout=30 --tries=3 https://example.com/file.tar.gz'
```

### Recovery

```bash
# Option 1: Force kill from Jenkins UI
# Manage Jenkins → Script Console
Jenkins.instance.getItemByFullName('job-name')
    .getBuildByNumber(47)
    .doKill()

# Option 2: Kill the process on the agent
ssh agent-01 "ps aux | grep 'build-47' | awk '{print \$2}' | xargs kill -9"

# Option 3: Delete the agent pod (if Kubernetes)
kubectl delete pod jenkins-agent-xxxxx --force --grace-period=0

# Option 4: Nuclear — restart Jenkins
systemctl restart jenkins
# Warning: ALL running builds will be lost
```

---

## 🟡 Disaster 5: Jenkins Controller Crash

### The Scenario

```
Thursday 9 AM: niemand can access Jenkins UI
Logs show:
  java.lang.OutOfMemoryError: Java heap space
  
Or:
  SEVERE: Jenkins.cleanUp()
  java.lang.RuntimeException: Shitting down
  (yes, that's a real Jenkins log message)
```

### Why It Happens

```
1. Out of memory (OOM)
   - Too many build logs in memory
   - Too many plugins loaded
   - Java heap too small

2. Deadlock
   - Plugin bug causes thread deadlock
   - Jenkins becomes unresponsive

3. Disk full
   - Can't write to disk
   - Can't serialize state
   - Corrupted state

4. Plugin crash
   - A plugin throws uncaught exception
   - Takes down the entire JVM
```

### Prevention

```bash
# 1. Set proper JVM memory
JAVA_OPTS="-Xms2g -Xmx4g -XX:+UseG1GC"

# 2. Monitor Jenkins memory
# Use Monitoring plugin or external monitoring

# 3. Limit concurrent builds
# Manage Jenkins → Configure System → # of executors on controller = 0

# 4. Regular backups
0 2 * * * tar -czf /backup/jenkins-$(date +\%Y\%m\%d).tar.gz $JENKINS_HOME/
```

### Recovery

```bash
# 1. Check logs
tail -200 $JENKINS_HOME/logs/jenkins.log

# 2. Increase memory and restart
export JAVA_OPTS="-Xms4g -Xmx8g"
systemctl restart jenkins

# 3. If corrupt — restore from backup
systemctl stop jenkins
rm -rf $JENKINS_HOME
tar -xzf /backup/jenkins-latest.tar.gz -C /
systemctl start jenkins

# 4. If no backup — start fresh with JCasC
# This is why Configuration as Code is critical
```

---

## 🔴 Jenkins Security Incidents

### Real-World Jenkins Attack Vectors

```
1. Unauthenticated access
   - Jenkins exposed to internet without auth
   - Script Console accessible → remote code execution
   
2. Credential theft
   - Access $JENKINS_HOME/secrets/ → decrypt all credentials
   - Build logs contain credentials
   
3. Supply chain attack
   - Malicious plugin installed
   - Compromised build agent
   - Modified Jenkinsfile in PR runs malicious code
   
4. Lateral movement
   - Jenkins has network access to production
   - Compromised Jenkins → access to prod via kubectl/SSH
```

### Jenkins Security Checklist

```
□ Jenkins is NOT exposed to the internet (behind VPN/firewall)
□ Authentication enabled (LDAP, OIDC, SAML)
□ Authorization configured (Matrix/RBAC — not "anyone can do anything")
□ Script Console restricted to admins only
□ CSRF protection enabled
□ Agent-to-controller access restricted
□ Builds don't run on the controller
□ Secrets managed via credentials plugin (not plaintext)
□ Plugins reviewed and minimized
□ Jenkins version kept up to date
□ Build logs don't contain secrets
□ PR builds are sandboxed (can't access prod credentials)
```

---

## 🔴 Post-Mortem Template

When a Jenkins disaster happens (and it will), document it:

```markdown
## Incident: [Title]
**Date:** YYYY-MM-DD
**Duration:** X hours
**Severity:** P1/P2/P3

### What Happened
[Description of the incident]

### Timeline
- HH:MM - First symptom noticed
- HH:MM - Investigation started  
- HH:MM - Root cause identified
- HH:MM - Fix applied
- HH:MM - Service restored

### Root Cause
[What actually caused the problem]

### Impact
- X builds failed
- Y developers blocked for Z hours
- Deployment to production delayed by N hours

### What Went Well
- [Things that helped during recovery]

### What Went Wrong
- [Things that made recovery harder]

### Action Items
- [ ] [Preventive measure 1]
- [ ] [Preventive measure 2]
- [ ] [Monitoring improvement]
```

---

**Previous:** [04. Integration with Kubernetes](./04-jenkins-kubernetes-integration.md)
