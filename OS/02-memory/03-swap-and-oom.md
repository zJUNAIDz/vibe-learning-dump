# Swapping & The OOM Killer

**When You Max Out Your Credit Card**

🟡 **Intermediate** | 🔴 **Advanced**

---

## Introduction

So you wrote a memory leak in your Go app, or you tried to load a 10GB CSV into memory in Python on a 2GB server. What happens next?

The kernel has two lines of defense to stop your laptop from turning into a spicy brick:
1. **Swap Space** (The Credit Card)
2. **The OOM Killer** (The Repo Man)

---

## Swap Space: The Illusion of More RAM

When physical RAM is completely full, the kernel has to make a hard choice. It looks around and says, "Which of these pages (chunks of RAM) haven't been used in a while?"

It takes those idle pages and physically writes them to a special area on the disk called **Swap**. It then frees up that RAM for the app that actively needs it.

### The Catch
RAM is insanely fast (nanoseconds). Disks are extremely slow (milliseconds). 
If the app tries to access the memory that was banished to the disk, the kernel has to:
1. Pause the app (Page Fault).
2. Read the page back from the disk into RAM.
3. Resume the app.

If this happens too heavily, your system enters a state called **Thrashing**. The computer spends 99% of its time moving memory back and forth between disk and RAM, and 1% actually executing code. Your system will feel completely frozen. A literal nightmare.

### `vm.swappiness`
Linux relies on a sysctl parameter called `swappiness` (range 0-100).
- `100`: Aggressively swap everything it can.
- `60` (Default): Balanced.
- `0`: Only swap if absolutely necessary to avoid crashing. (Common on database servers where thrashing ruins query latency).

---

## The OOM Killer (Out Of Memory)

If RAM is full, and Swap is full, the kernel's back is against the wall. The Linux kernel will literally act as an assassin to save the system from freezing completely.

Enter the **OOM (Out Of Memory) Killer**. 

It uses a heuristic to calculate an `oom_score` for every running process. It wants to kill a process that:
1. Will free up the most memory.
2. Isn't a critical system service (like `systemd` or `sshd`).
3. Has low privileges.

Once it picks a target, it sends a `SIGKILL` (Signal 9, unstoppable death) to that process.
And your process goes **POOF**.

### Checking the Logs
If your app randomly vanished, check `dmesg`.

```bash
$ dmesg -T | grep -i oom
[Tue Feb 20 14:02:44 2024] Out of memory: Killed process 12345 (node) total-vm:1500M, anon-rss:500M, file-rss:0M...
```
*Bro got sent to the shadow realm.*

### Adjusting the OOM Score (Playing God)
You can tell the kernel, "Hey, whatever happens, please don't kill my database."

```bash
# View the calculated score for a process (higher = more likely to die)
$ cat /proc/<pid>/oom_score
800

# Adjust the score negatively to protect it (requires root)
$ echo -500 > /proc/<pid>/oom_score_adj
```

---

## Swap and Containers (Kubernetes/Docker)

If you're running Kubernetes in production, the absolute standard rule is: **Disable Swap on the host node.**
Wait, what? Disable the safety net?

Yes! If your container has a memory limit of `512Mi`, and swap is enabled on the host, the container could start swapping to disk without you knowing. Performance degrades unpredictably, and you won't get alerted until users complain.
By disabling swap, Kubernetes forces the container to hit its limit and get OOMKilled *immediately*. 

This is known as **"Failing Fast."**
Kubernetes will just restart the pod and generate an `OOMKilled` event so you can actually fix the memory leak instead of hiding it in the disk. Clean, efficient, based.

## Key Takeaways
1. **Swap** acts as overflow RAM by using the hard drive. It prevents crashes but causes extreme slowdowns (thrashing).
2. The **OOM Killer** is the kernel's mechanism of last resort. It brutally shoots the fattest process to save the OS.
3. In containerized environments, we usually disable swap so we get `OOMKilled` fast instead of degrading performance silently.

---
**Next:** [Module 03: Filesystem Fundamentals](../03-filesystems/01-filesystem-fundamentals.md)