# "Everything is a File"

**The Ultimate Unix Vibe Check**

🟢 **Beginner-Friendly** | 🟡 **Intermediate**

---

## Introduction

If you spend 5 minutes in any Linux forum, some greybeard with a Tux profile picture will inevitably say: *"In Unix, everything is a file."*

But what the heck does that actually mean? Do they mean literally everything is a `.txt` on a hard drive? 

No. What they mean is: **A file descriptor is the universal interface or API for communicating with the OS.**

---

## Virtual Filesystems: Treating Kernel Data like Text

Linux exposes live, dynamic data from the kernel *as if they were files on disk*. These files don't actually exist on your SSD; they live purely in RAM, generated on the fly when you try to read them. Total illusion. Big matrix energy.

### 1. The `/proc` Filesystem (Process Information)

Want to see exactly what a running process is doing without a specialized GUI tool? Just read a text file.

```bash
# Get the PID of our shell
$ echo $$
1234

# Let's peek into its /proc directory!
$ ls /proc/1234
cmdline   environ  limits    net        root      statm
comm      exe      maps      numa_maps  smaps     status
cpuset    fd       mem       oom_score  smaps_rollup  task
cwd       fdinfo   mounts    pagemap    stat      wchan
```

If you `cat` those files, the kernel literally builds a string of text dynamically and hands it back to you:
- `/proc/1234/cmdline`: The command used to start it.
- `/proc/1234/environ`: All the environment variables.
- `/proc/1234/limits`: The `ulimit` values for this specific process.

This is how tools like `ps` or `top` work. They aren't doing black magic. They are just rapidly reading `/proc` directories and formatting the output nicely.

### 2. The `/sys` Filesystem (Hardware and Kernel Config)

While `/proc` is largely for processes, `/sys` allows you to talk to your hardware and kernel sub-systems directly using `cat` and `echo`. 

Want to change the brightness of your laptop screen from the terminal? Don't download a fancy app. Just write a number to a file.

```bash
# Read max brightness
$ cat /sys/class/backlight/intel_backlight/max_brightness
100

# Set brightness to 50%
$ echo 50 | sudo tee /sys/class/backlight/intel_backlight/brightness
```
*Bro just commanded physical hardware with standard output. Based.*

### 3. The `/dev` Filesystem (Device Nodes)

Hardware devices are exposed as files here. 

- **`/dev/sda`**: Your literal first hard drive.
- **`/dev/tty`**: Your terminal window.
- **`/dev/null`**: The literal black hole. Anything written here is instantly deleted by the kernel. (Like texting your ex. Oof.)
- **`/dev/urandom`**: An infinite stream of cryptographically secure random garbage.
- **`/dev/zero`**: An infinite stream of NULL bytes (`\0`).

## Writing to the Terminal

When you type `echo "Hello"`, where does it go? 
It goes to Standard Output (File Descriptor 1), which is wired to `/dev/tty#`, which prints to your screen.

You can actually troll other users logged into the same Linux machine:

```bash
# Find what terminal my friend is using
$ who
bob     pts/1        2024-02-21 14:00 (192.168.1.100)

# Write directly to his screen
$ echo "you up?" > /dev/pts/1
```
*Disclaimer: Don't actually do this at work unless you want a fun chat with HR.*

## Key Takeaways
1. "Everything is a file" means that the kernel uses the standard `read()` and `write()` API for *everything*: hardware, network configs, memory, random number generation, etc.
2. `/proc` is a fake filesystem that lets you spy on running processes.
3. `/sys` lets you talk to hardware and change kernel behavior by echoing text.
4. `/dev` maps physical and logical devices into the filesystem.

---
**Next:** [Module 04: File Descriptors](../04-file-descriptors/01-fd-fundamentals.md)