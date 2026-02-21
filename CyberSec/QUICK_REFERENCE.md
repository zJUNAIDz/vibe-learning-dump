# Quick Reference Guide 📖

Fast lookup for commands, tools, and concepts used throughout the curriculum.

---

## 🔍 Network Analysis

### nmap — Port Scanning
```bash
# Basic scan
nmap <target>

# Common ports
nmap -F <target>

# Service version detection
nmap -sV <target>

# OS detection
sudo nmap -O <target>

# Scan localhost
nmap localhost
```

### netcat (nc) — Network Swiss Army Knife
```bash
# Listen on port
nc -l 4444

# Connect to port
nc <host> <port>

# Send file
nc -w 3 <host> <port> < file.txt

# Port banner grab
echo "" | nc <host> <port>
```

### tcpdump — Packet Capture
```bash
# Capture on interface
sudo tcpdump -i eth0

# Capture HTTP traffic
sudo tcpdump -i any -A 'tcp port 80'

# Save to file
sudo tcpdump -i any -w capture.pcap

# Read from file
tcpdump -r capture.pcap
```

### ss — Socket Statistics
```bash
# Show listening ports
ss -tuln

# Show all TCP connections
ss -tan

# Show process info
ss -tunap

# Show only listening sockets
ss -tunl | grep LISTEN
```

---

## 🌐 HTTP Testing

### curl — HTTP Client
```bash
# Basic GET
curl https://example.com

# Show headers
curl -i https://example.com

# Follow redirects
curl -L https://example.com

# POST JSON
curl -X POST https://api.example.com/data \
  -H "Content-Type: application/json" \
  -d '{"key":"value"}'

# Send cookie
curl -b "session=abc123" https://example.com

# Show request details
curl -v https://example.com

# Ignore SSL errors (testing only!)
curl -k https://self-signed.example.com
```

### HTTPie — Better HTTP Client
```bash
# Basic GET
http GET https://example.com

# POST JSON (automatic)
http POST https://api.example.com/data key=value

# Custom headers
http GET https://api.example.com Authorization:"Bearer token123"

# Download file
http --download https://example.com/file.zip

# Follow redirects
http --follow https://example.com
```

---

## 🐧 System Debugging

### strace — Syscall Tracing
```bash
# Trace a command
strace ls

# Trace specific syscalls
strace -e open,read,write cat file.txt

# Trace running process
sudo strace -p <PID>

# Follow child processes
strace -f ./program

# Output to file
strace -o trace.log ./program
```

### lsof — List Open Files
```bash
# Show all open files
sudo lsof

# Show network connections
sudo lsof -i

# Show what's using a port
sudo lsof -i :8080

# Show files opened by process
lsof -p <PID>

# Show files opened by user
lsof -u username
```

### ps — Process Listing
```bash
# All processes
ps aux

# Process tree
ps auxf

# Find process
ps aux | grep nginx

# Show process by PID
ps -p <PID> -o comm,pid,ppid,user
```

---

## 🔒 File Permissions

### Basic Permission Commands
```bash
# View permissions
ls -la

# Change permissions (numeric)
chmod 644 file.txt     # rw-r--r--
chmod 755 script.sh    # rwxr-xr-x
chmod 600 private.key  # rw-------

# Change ownership
sudo chown user:group file.txt

# Recursive change
sudo chown -R user:group /path/to/dir

# View file capabilities
getcap /path/to/binary

# Set capabilities
sudo setcap cap_net_raw+ep /path/to/binary
```

---

## 🐳 Docker Security

### Container Commands
```bash
# Run container (foreground)
docker run --rm <image>

# Run container (background)
docker run -d --name myapp <image>

# Run with limited resources
docker run -m 512m --cpus="1.0" <image>

# Run as non-root user
docker run --user 1000:1000 <image>

# Run with read-only filesystem
docker run --read-only <image>

# Inspect container
docker inspect <container>

# Check logs
docker logs <container>

# Execute command in running container
docker exec -it <container> /bin/sh
```

### Image Security
```bash
# Scan image for vulnerabilities
docker scan <image>

# View image layers
docker history <image>

# Remove dangling images
docker image prune
```

---

## 🔐 SSL/TLS Testing

### OpenSSL Commands
```bash
# Test SSL connection
openssl s_client -connect example.com:443

# Show certificate
openssl s_client -connect example.com:443 -showcerts

# Check certificate expiry
echo | openssl s_client -connect example.com:443 2>/dev/null | \
  openssl x509 -noout -dates

# Generate self-signed certificate
openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -days 365 -nodes
```

---

## 📝 Log Analysis

### journalctl — systemd Logs
```bash
# View all logs
sudo journalctl

# Follow logs (tail -f style)
sudo journalctl -f

# Logs from service
sudo journalctl -u nginx

# Logs since boot
sudo journalctl -b

# Logs from today
sudo journalctl --since today

# Show errors only
sudo journalctl -p err
```

### grep — Pattern Matching
```bash
# Find failed login attempts
sudo grep "Failed password" /var/log/auth.log

# Find with context
grep -C 3 "error" logfile.txt

# Recursive search
grep -r "TODO" /path/to/code

# Case insensitive
grep -i "warning" logfile.txt

# Count matches
grep -c "error" logfile.txt
```

---

## 🧰 Useful One-Liners

### Network
```bash
# Show public IP
curl -s https://ifconfig.me

# Test if port is open
nc -zv <host> <port>

# List listening ports
ss -tuln | grep LISTEN

# Find process using port
sudo lsof -i :8080
```

### System
```bash
# Find large files
find / -type f -size +100M 2>/dev/null

# Show disk usage
df -h

# Show directory sizes
du -sh *

# Find SUID binaries
find / -perm -4000 2>/dev/null

# Show environment variables
printenv
```

### Web
```bash
# Start simple HTTP server
python3 -m http.server 8000

# Follow redirects and show final URL
curl -sIL -o /dev/null -w '%{url_effective}' https://example.com

# Test CORS
curl -H "Origin: https://evil.com" -I https://api.example.com
```

---

## 🎯 Common Vulnerabilities - Quick Test

### SQL Injection
```bash
# Test in URL parameter
http GET "https://example.com/user?id=1' OR '1'='1"

# Test in POST body
http POST https://example.com/login username="admin' --" password="x"
```

### XSS (Cross-Site Scripting)
```bash
# Test reflected XSS
http GET "https://example.com/search?q=<script>alert(1)</script>"
```

### SSRF (Server-Side Request Forgery)
```bash
# Test SSRF
http POST https://example.com/fetch url="http://localhost:8080/admin"
```

### Command Injection
```bash
# Test command injection
http POST https://example.com/ping host="127.0.0.1; cat /etc/passwd"
```

---

## 🔧 Burp Suite Shortcuts

| Action | Shortcut |
|--------|----------|
| Send to Repeater | Ctrl+R |
| Send to Intruder | Ctrl+I |
| Forward request | Ctrl+F |
| Drop request | Ctrl+D |
| Switch to Target | Ctrl+Shift+T |
| Switch to Proxy | Ctrl+Shift+P |
| Clear proxy history | Ctrl+Shift+Del |

---

## 📊 HTTP Status Codes (Security-Relevant)

| Code | Meaning | Security Implication |
|------|---------|---------------------|
| 200 | OK | Request succeeded |
| 301 | Moved Permanently | Check for open redirects |
| 302 | Found (redirect) | Check for open redirects |
| 400 | Bad Request | May leak validation logic |
| 401 | Unauthorized | Auth required |
| 403 | Forbidden | Auth passed, authz failed |
| 404 | Not Found | May reveal resource existence |
| 500 | Internal Server Error | May leak stack traces |
| 502 | Bad Gateway | Proxy/backend issue |
| 503 | Service Unavailable | Possible DoS or maintenance |

---

## 🛡️ Security Headers Checklist

```bash
# Check security headers
curl -sI https://example.com | grep -iE '(Strict-Transport|X-Frame|X-Content|Content-Security|X-XSS)'
```

**Important headers:**
- `Strict-Transport-Security` — Force HTTPS
- `X-Frame-Options` — Clickjacking protection
- `X-Content-Type-Options` — MIME sniffing protection
- `Content-Security-Policy` — XSS protection
- `X-XSS-Protection` — Legacy XSS filter
- `Referrer-Policy` — Control referrer leakage

---

## 🔑 JWT (JSON Web Token) Debugging

### Decode JWT (using jq)
```bash
# Decode JWT (header and payload)
echo "eyJhbGc...token..." | \
  cut -d. -f1-2 | \
  tr '.' '\n' | \
  while read part; do echo "$part" | base64 -d 2>/dev/null | jq; done
```

### Online Decoder
- https://jwt.io

---

## 📚 Useful Files and Paths

### Linux Security-Relevant Paths
```
/etc/passwd           # User accounts
/etc/shadow           # Password hashes (requires root)
/etc/group            # Group memberships
/etc/sudoers          # Sudo configuration
/var/log/auth.log     # Authentication logs (Debian/Ubuntu)
/var/log/secure       # Authentication logs (RHEL/Fedora)
/proc/<PID>/cmdline   # Process command line
/proc/<PID>/environ   # Process environment
~/.ssh/authorized_keys # SSH public keys
~/.bash_history       # Command history
```

---

## 🚀 Quick Environment Setup

### Start Juice Shop
```bash
docker run -d -p 3000:3000 bkimminich/juice-shop
```

### Start Local Test Server
```bash
# Node.js
npx http-server -p 8000

# Python
python3 -m http.server 8000
```

### Start Burp Suite
```bash
BurpSuiteCommunity
# Configure Firefox proxy: 127.0.0.1:8080
```

---

## 🆘 Emergency Commands

### Kill Process on Port
```bash
# Find PID
sudo lsof -ti:8080

# Kill it
sudo kill -9 $(sudo lsof -ti:8080)
```

### Reset Docker
```bash
docker stop $(docker ps -aq)
docker rm $(docker ps -aq)
docker system prune -af
```

### Clear Browser Cache/Cookies
```bash
# Firefox (Linux)
rm -rf ~/.mozilla/firefox/*.default*/cache2/*
```

---

**Bookmark this page!** You'll reference it constantly.

→ Back to [README](./README.md) | [Start Learning](./00-orientation/00-how-web-apps-get-hacked.md)
