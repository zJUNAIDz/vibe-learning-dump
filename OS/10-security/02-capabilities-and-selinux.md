# Capabilities and SELinux

**Revoking Sudo's Blank Check**

🔴 **Advanced**

---

## 1. Capabilities: Splitting the Root Atom ⚛️

In traditional Linux security (the classic `ugo/rwx` model from Part 1), there is a massive design flaw.
There are regular users (low privilege), and there is `root` (God mode). 

Let's say you want to run an Nginx web server on port `80`. 
Under fundamental Linux rules, *only* `root` is allowed to open sockets on ports below 1024. So, you have to launch `nginx` as `root` using `sudo`.

But what if there is a vulnerability in Nginx? What if a hacker sends a malicious HTTP request that exploits a buffer overflow?
Since Nginx is running as `root`, the hacker now has a `root` shell. They can wipe your hard drive, steal your SSH keys, and ruin your life. 

It makes no sense to give Nginx the power to format hard drives just because it needs to open Port 80. That's giving someone the nuclear launch codes because they asked to borrow a lighter.

### Enter Linux Capabilities

The kernel devs realized `root` was too monolithic. So they sliced the "God mode" into about 40 specific, granular powers called **Capabilities**.

Instead of making Nginx run as `root`, you run it as a standard, unprivileged `nginx-user`. But, you grant that specific binary *one single capability*: `CAP_NET_BIND_SERVICE`.

```bash
# Give the nginx binary the power to bind ports < 1024, but NOTHING else.
$ sudo setcap 'cap_net_bind_service=+ep' /usr/sbin/nginx
```

Now, if a hacker exploits Nginx, they are trapped as an unprivileged user. They can't touch `/etc/shadow`. They can't reboot the machine. You just contained the blast radius.

**Common Capabilities You Actually Care About:**
*   `CAP_NET_BIND_SERVICE`: Open ports under 1024.
*   `CAP_CHOWN`: Change file owners.
*   `CAP_KILL`: Send signals (like `SIGKILL`) to processes owned by other users.
*   `CAP_SYS_ADMIN`: The "catch-all" capability. (Docker uses this for a lot of things. Try to avoid giving containers this).

**How to check a process's capabilities:**
```bash
# Output looks like unreadable hex codes unless you use a decoder tool like `capsh`
$ grep Cap /proc/1234/status
```

---

## 2. Mandatory Access Control (MAC): AppArmor and SELinux 🛡️

Standard permissions (rwx) are called **Discretionary Access Control (DAC)**. That means the *owner* of the file has the discretion to change the rules. If I own `passwords.txt`, I can run `chmod 777 passwords.txt` and expose it to the world.

In enterprise and government environments, that’s a huge "nope." Organizations want **Mandatory Access Control (MAC)**.
MAC says: "I don't care if you own the file. I don't even care if you are `root`. The Central Policy says the `nginx` process is not allowed to read files in `/home`. Access Denied."

Linux has two main implementations of MAC:
1.  **AppArmor** (Default on Ubuntu/Debian) - Profile-based. Pretty chill, easy to write policies for.
2.  **SELinux** (Security-Enhanced Linux. Default on RHEL/Fedora/CentOS) - Originally developed by the NSA. Extremely strict, label-based, and the source of 90% of DevOps engineer tears.

### Surviving SELinux without turning it off

The biggest meme in Linux system administration is:
> *Step 1 of installing any software: Disable SELinux.*

(Please don't actually do this. It's meant to stop zero-day exploits payload execution).

SELinux works by attaching a **Context Label** to every single file, process, and port on the system.
If the policy doesn't explicitly allow `<Process_Label>` to interact with `<File_Label>`, it blocks it, even for `root`.

**The Classic Scenario:**
You mount a new SSD to `/var/www/html` to hold all your website images. `nginx` is running. You set `chown nginx:nginx /var/www/html`. You set `chmod 755`. Standard permissions are perfect.
You hit the website. **403 Forbidden**. 
You pull your hair out for 3 hours. 

Check SELinux:
```bash
# The 'Z' flag shows SELinux contexts
$ ls -lZ /var/www/html
drwxr-xr-x. root root unconfined_u:object_r:default_t:s0 /var/www/html

# Oh. The files are labeled as `default_t`. 
# Nginx expects them to be labeled as `httpd_sys_content_t`.
```

**The Fix:**
Tell SELinux to restore the correct labels.
```bash
# Relabel the directory so Nginx is legally allowed to touch it
$ sudo restorecon -Rv /var/www/html
```

When in doubt, check the audit logs:
```bash
$ sudo grep AVC /var/log/audit/audit.log
```
SELinux will literally tell you exactly what it blocked and why. It’s tough love. Embrace it.

---
**Next:** [Module 11: Debugging](../11-debugging/01-observability-tools.md)
