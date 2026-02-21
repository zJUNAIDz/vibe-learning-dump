# 🛠️ Module 10: Linux CLI Security Tools

**Difficulty:** 🟡 Intermediate  
**Time:** 45 minutes

---

## What You'll Learn

- nmap (network scanning)
- curl (security testing)
- netcat (network debugging)
- ss (socket statistics)
- lsof (open files and connections)
- strace (system call tracing)

---

## nmap — Network Scanning

### What It Does

**Discovers hosts, open ports, and services on a network.**

---

### Basic Scans

```bash
# Scan single host
nmap 192.168.1.100

# Scan range
nmap 192.168.1.1-254

# Scan common ports (fast)
nmap -F example.com

# Scan specific ports
nmap -p 80,443,8080 example.com
```

---

### Service Detection

```bash
# Detect service versions
nmap -sV example.com

# Example output:
# PORT    STATE SERVICE VERSION
# 22/tcp  open  ssh     OpenSSH 8.2p1
# 80/tcp  open  http    nginx 1.18.0
# 443/tcp open  https   nginx 1.18.0
```

---

### OS Detection

```bash
# Guess operating system (requires root)
sudo nmap -O example.com
```

---

### Security Use Cases

#### 1. Verify Firewall Rules

```bash
# Check if port 3306 (MySQL) is exposed
nmap -p 3306 yourdomain.com

# It should be filtered/closed (not accessible from internet)
```

#### 2. Find Unnecessary Open Ports

```bash
# Scan your server
nmap -p 1-65535 yourserver.com

# Are there unexpected open ports?
```

#### 3. Check for Old Services

```bash
# Service version detection
nmap -sV yourserver.com

# Look for outdated versions with known CVEs
```

---

### Ethical Note

**Only scan systems you own or have permission to scan.**

nmap can trigger intrusion detection systems (IDS).

---

## curl — HTTP Security Testing

### Basic Usage

```bash
# GET request
curl https://example.com

# POST JSON
curl -X POST https://api.example.com/login \
  -H "Content-Type: application/json" \
  -d '{"username":"test","password":"test"}'

# Show headers
curl -i https://example.com

# Verbose (see full exchange)
curl -v https://example.com
```

---

### Security Testing

#### 1. Check Security Headers

```bash
curl -sI https://example.com | grep -iE "(Strict-Transport|X-Frame|Content-Security|X-Content)"
```

**Look for:**
- `Strict-Transport-Security`
- `X-Frame-Options`
- `Content-Security-Policy`
- `X-Content-Type-Options`

---

#### 2. Test CORS

```bash
curl -H "Origin: https://evil.com" -I https://api.example.com/data

# Check for:
# Access-Control-Allow-Origin: *  ← Overly permissive!
```

---

#### 3. Test Authentication

```bash
# Without auth
curl https://api.example.com/admin
# Should return 401

# With auth
curl -H "Authorization: Bearer token123" https://api.example.com/admin
```

---

#### 4. Test Rate Limiting

```bash
# Send multiple requests rapidly
for i in {1..100}; do
  curl https://api.example.com/login \
    -d '{"username":"test","password":"wrong"}' &
done

# Are you rate-limited?
```

---

## netcat (nc) — Network Swiss Army Knife

### What It Does

**Create arbitrary TCP/UDP connections.**

---

### Basic Usage

```bash
# Connect to server
nc example.com 80

# Then type:
GET / HTTP/1.1
Host: example.com

# (Press Enter twice)
```

---

### Security Use Cases

#### 1. Port Checking

```bash
# Check if port is open
nc -zv example.com 22

# Output:
# Connection to example.com 22 port [tcp/ssh] succeeded!
```

#### 2. Banner Grabbing

```bash
# Get service banner
echo "" | nc example.com 22

# Output:
# SSH-2.0-OpenSSH_8.2p1 Ubuntu-4ubuntu0.5
```

#### 3. Simple HTTP Server (Testing)

```bash
# Listen on port 8080
while true; do
  echo -e "HTTP/1.1 200 OK\n\n<h1>Hello</h1>" | nc -l 8080
done
```

---

## ss — Socket Statistics

### What It Does

**Shows network connections, listening ports.**

**Replacement for `netstat`.**

---

### Basic Usage

```bash
# Show all TCP connections
ss -tan

# Show listening sockets
ss -tunl

# Show with process names
ss -tunap
```

---

### Output Explanation

```
State      Recv-Q Send-Q Local Address:Port Peer Address:Port
LISTEN     0      128    0.0.0.0:22         0.0.0.0:*
ESTAB      0      0      192.168.1.100:45678 93.184.216.34:443
```

**States:**
- `LISTEN` — Listening for connections
- `ESTAB` — Established connection
- `TIME-WAIT` — Connection closing

---

### Security Use Cases

#### 1. Find Listening Ports

```bash
# What's listening?
ss -tunl | grep LISTEN

# Unexpected ports? Investigate!
```

#### 2. Find Established Connections

```bash
# Where is your app connecting?
ss -tunap | grep ESTAB
```

#### 3. Check for Suspicious Connections

```bash
# Connections to unusual IPs?
ss -tunap | grep ESTAB | grep -v "192.168\|127.0"
```

---

## lsof — List Open Files

### What It Does

**Shows open files, network connections, process info.**

**Remember:** In Linux, everything is a file (including network sockets).

---

### Basic Usage

```bash
# All open files (lots of output!)
sudo lsof

# Network connections only
sudo lsof -i

# Specific port
sudo lsof -i :8080

# Files opened by process
lsof -p <PID>

# Files opened by user
lsof -u username
```

---

### Security Use Cases

#### 1. Find Process Using Port

```bash
# What's using port 3000?
sudo lsof -i :3000

# Output:
# COMMAND  PID  USER   FD   TYPE DEVICE SIZE/OFF NODE NAME
# node    1234  webapp 21u  IPv4  12345      0t0  TCP *:3000 (LISTEN)
```

#### 2. Find Deleted But Open Files

```bash
# Files deleted but still held open (disk space not freed)
sudo lsof | grep deleted

# Kill process to free space
```

#### 3. Investigate Suspicious Process

```bash
# What files is this process accessing?
sudo lsof -p <suspicious-PID>

# Is it accessing /etc/passwd, SSH keys, etc.?
```

---

## strace — System Call Tracing

### What It Does

**Shows all system calls made by a process.**

---

### Basic Usage

```bash
# Trace a command
strace ls

# Trace specific syscalls
strace -e open,read,write cat file.txt

# Trace running process
sudo strace -p <PID>

# Follow child processes
strace -f ./program
```

---

### Security Use Cases

#### 1. Find Files Being Accessed

```bash
# What files does app read?
strace -e open node server.js 2>&1 | grep "\.env"

# Is it reading .env from unexpected locations?
```

#### 2. Detect Network Connections

```bash
# What connections is app making?
strace -e connect,socket curl https://example.com
```

#### 3. Debug Permission Issues

```bash
# Why can't app write to file?
strace -e open,write ./app 2>&1 | grep "Permission denied"
```

---

## Combining Tools for Investigation

### Scenario: Unknown Process

```bash
# 1. Find suspicious process
ps aux | grep <suspicious-name>

# 2. What ports is it using?
sudo lsof -i -p <PID>

# 3. What files is it accessing?
sudo lsof -p <PID>

# 4. Trace its system calls
sudo strace -p <PID> 2>&1 | head -100

# 5. Where is the binary?
ls -la /proc/<PID>/exe
```

---

## Summary

| Tool | Purpose | Example |
|------|---------|---------|
| **nmap** | Network scanning | `nmap -p 80,443 example.com` |
| **curl** | HTTP testing | `curl -H "Auth: Bearer token" api.com` |
| **nc** | Raw TCP/UDP | `nc -zv example.com 22` |
| **ss** | Socket stats | `ss -tunl | grep LISTEN` |
| **lsof** | Open files/connections | `lsof -i :8080` |
| **strace** | System call tracing | `strace -e open ./app` |

---

## Exercises

### Exercise 1: Scan Your Server
```bash
nmap -sV yourserver.com
```
Are there unexpected open ports?

### Exercise 2: Find Listening Ports
```bash
ss -tunl | grep LISTEN
```
Do you recognize all services?

### Exercise 3: Trace Your App
```bash
strace -e open node yourapp.js 2>&1 | grep "\.env"
```
Where does it load config from?

---

## What's Next?

Now let's explore TypeScript-specific security issues.

→ **Next: [Module 11: Application Security in TypeScript](../11-typescript-security/01-typescript-security.md)**

---

## Further Reading

- [nmap Documentation](https://nmap.org/book/man.html)
- [curl Manual](https://curl.se/docs/manual.html)
- [strace Tutorial](https://jvns.ca/blog/2021/04/03/what-problems-do-people-solve-with-strace/)
