# Docker Under the Hood

**Wait, So Containers Aren't Real?**

🟡 **Intermediate** | 🔴 **Advanced**

---

## The Big Lie

When you type `docker run -it ubuntu bash`, you probably feel like you just booted up a tiny virtual machine. "Wow, a whole new OS in 500 milliseconds!"

No cap, that's entirely fake. **Containers do not exist in the Linux kernel.** There is no `struct container` in the C code. It's an illusion. 

A container is literally just a normal Linux process (like Node.js or Python) wearing a hyper-realistic disguise constructed from two kernel features: **Namespaces** and **Control Groups (cgroups)**.

---

## 1. Namespaces: The Gaslighting Engine 🪞

Namespaces are how the kernel lies to a process about what else exists on the computer. It gives the process "main character syndrome."

When Docker starts your app, it puts it in a bunch of locked rooms:

*   **PID Namespace:** "You are PID 1!" (It's actually PID 45892 on the host).
*   **Network Namespace:** "Here is your personal `eth0` interface with IP 172.17.0.2!" (It's a `veth` pair connected to a bridge, as we saw in the networking deep dive).
*   **Mount Namespace:** "This is the root filesystem `/`. Look, there is `/bin` and `/etc`!" (It's actually a folded OverlayFS mount sitting in `/var/lib/docker/overlay2/`).

To prove this, try running:
```bash
$ unshare --pid --fork --mount-proc bash
```
Congratulations, you just manually created a PID namespace without Docker. If you type `ps` inside that new bash shell, you will be PID 1. It's giving "I am the captain now" energy.

---

## 2. Cgroups: The Bouncer 🧱

Namespaces isolate *visibility*. They don't isolate *resources*.
If your isolated "container" process decides to run an infinite `while(true)` loop and allocate 500GB of RAM, it will completely crash the host machine. 

This is where **cgroups** come in. Cgroups set hard limits on how much CPU, Memory, and Disk I/O a process and its children are allowed to use.

When you run:
```bash
docker run --memory="256m" --cpus="0.5" nginx
```

Docker is doing two things:
1. Creating a folder in `/sys/fs/cgroup/memory/docker/<container_id>`
2. Writing `268435456` (256MB in bytes) into `memory.limit_in_bytes`
3. Writing the PID of the Nginx process into the `tasks` file in that folder.

If Nginx tries to allocate 257MB... the **OOM Killer** (remember him?) will spawn inside that cgroup and yeet the process to the shadow realm. The rest of the host OS will be completely fine.

---

## 3. OverlayFS: The Infinite Glitch

If every container is just a process, how does `docker images` work? Why doesn't an `ubuntu` container take up 2GB of disk space every time I run it?

Because Docker uses **Union Filesystems** (usually `OverlayFS`).
It takes your base image (Layer A: Ubuntu), your app code (Layer B: Node.js app), and stacks them on top of each other. 

When your container runs, the OS mounts these layers together as **Read-Only**.
Any time your app tries to *write* a file, OverlayFS silently intercepts it, copies the file to a temporary "Upper" layer (the Container Layer), and writes the change there. This is called **Copy-on-Write (CoW)**.

It's essentially Git for your hard drive. Multiple containers can share the exact same read-only base layers on disk. Efficiency = 100%.

---

## Summary
Docker is not magic. It's just a Go binary that acts as a wrapper around `clone()` system calls, `cgroups` `sysfs` files, and `iptables` rules. When you understand the Linux kernel, you realize Docker is just a really good UX designer for kernel features that have existed since 2008. Based.

---
**Next:** [Module 08: systemd](../08-systemd/01-init-and-services.md)
