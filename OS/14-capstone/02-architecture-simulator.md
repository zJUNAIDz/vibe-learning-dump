# Architecture Simulator

**Build Your Own Micro-Cloud**

🔴 **Advanced** | 🎓 **Final Project**

---

## The Challenge

You've learned everything about the OS. Now it's time to put all the legos together. You are going to build a highly-available, containerized web architecture from scratch, using only raw Linux primitives.

**Do not use Docker, Docker Compose, or Kubernetes.**

If you can complete this lab, you understand the Linux OS better than 90% of working software engineers. No cap.

---

## 🏗️ Architecture Requirements

You need to construct the following environment on a single Linux Virtual Machine:

1.  **Network Namespace Isolation:**
    *   Create two Network Namespaces: `netns-web1` and `netns-web2`.
    *   Create a Linux Bridge (`br0`) on the host.
    *   Connect both namespaces to the bridge using `veth` pairs.
    *   Verify they can ping each other across the bridge.
2.  **The Application:**
    *   Write a tiny Go, Python, or Node.js web server that listens on port `8080` and returns `{"server": "webX", "status": "GOAT"}`.
    *   Run one instance of the app inside `netns-web1` and another inside `netns-web2`. (Hint: `ip netns exec netns-web1 node server.js`).
3.  **Process Management:**
    *   Write two `systemd` service files to manage your web processes. Ensure systemd automatically restarts them if they crash.
4.  **Resource Limits:**
    *   Create a cgroup for your services that hard-limits their memory to `128MB`.
5.  **Reverse Proxy / Load Balancing:**
    *   Install `nginx` or configure `iptables` / `IPVS` on the *host* network namespace.
    *   Configure it to listen on host port `80` and round-robin load balance traffic down into the bridge IPs for `web1:8080` and `web2:8080`.
6.  **Security:**
    *   Ensure the web applications run as a non-root user (`www-data` or similar).

---

## 🧪 Acceptance Criteria Tests

When you think you're done, run these exact commands to prove your setup works.

**1. The Load Balancer Test**
```bash
# Run this 5 times. You should see it alternate between web1 and web2.
$ curl http://localhost:80
```

**2. The High Availability Test**
```bash
# Kill one of the backend servers
$ sudo killall node

# Hit the load balancer again immediately. 
# It should seamlessly hand the request to the surviving server.
$ curl http://localhost:80
```

**3. The Systemd Healing Test**
```bash
# Wait 5 seconds after killing the app. Systemd should have revived it.
$ systemctl status my-web1.service
# Should say: Active: (running)
```

**4. The Cgroup Isolation Test**
```bash
# Inspect the memory folder for your cgroup. 
# Ensure the limit is strictly 128MB (134217728 bytes).
$ cat /sys/fs/cgroup/memory/my_web_apps/memory.limit_in_bytes
```

---

## 🏁 Conclusion

If you've made it this far and actually understood the memes and the code constraints, congratulations. You are officially no longer just a "user-space" developer trying to guess why the computer is mad at you. 

You understand the Kernel, the Page Cache, the Socket buffers, the Scheduler, and the OOM Killer. 
When things break in production (and they will break, spectacularly), you won't panic. You'll attach `strace`, look at the `dmesg` logs, analyze the `cgroups`, and fix it while the rest of the team is still trying to restart the VM.

May your uptime be high, and your page faults be low. 👑

---
**End of Curriculum.** You survived.