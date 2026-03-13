# journald and systemd Timers

**When Your Cronjobs Go To Therapy**

🟡 **Intermediate**

---

## 1. `journald`: The Central Nervous System 🧠

If you’ve ever managed a Linux server prior to 2012, logging felt like being in a bad group project where everyone dumps text files in `/var/log` but nobody agrees on the format.

Enter `systemd-journald`.

`journald` is systemd's built-in logging service. Instead of writing text files, it writes to an indexed, structured binary database. If you drop a `console.log()` in a Node app managed by systemd, it goes straight to the journal.

Because it's structured, you can query it like Google.

### The `journalctl` Flexes 🏋️‍♀️

**Follow a specific service's logs (like `tail -f`):**
```bash
$ sudo journalctl -u nginx.service -f
```

**Find out why a service crashed 3 days ago:**
```bash
$ sudo journalctl -u my-app.service --since "2024-03-10" --until "2024-03-11" -p err
```
*Note the `-p err` flag. This only shows errors (log levels critical, error). This is a literal lifesaver during outages.*

**See kernel logs (what `dmesg` does, but better):**
```bash
$ sudo journalctl -k
```
This is where you find out the OOM Killer brutally murdered your DB, or an ext4 driver hit bad sectors on your NVMe drive. Big yikes.

**Find logs from a specific process ID:**
```bash
$ sudo journalctl _PID=4583
```

`journald` automatically rotates and compresses old logs, preventing the classic "The server died because a 600GB log file filled up the root partition" disaster. It cleans up after itself, meaning less late-night pages for you. Protect your peace.

---

## 2. systemd Timers > Cronjobs ⏱️

For 40 years, `cron` has been the default way to run scheduled tasks (like DB backups or log parsing). `crontab` files look like this:

`0 4 * * * /usr/bin/python3 /opt/backup.py`

But here's why `cron` is kind of toxic in 2024:
1. If the host reboots right at `4:00 AM`, the job doesn't run. `cron` doesn't care. It just ghosts you.
2. If the `backup.py` script takes 2 hours instead of 1 minute, the next day at 4 AM it will spawn *another* backup instance. Now you have two running at once. Soon your server is on fire.
3. Where are the logs? Maybe mailed to the root user? Maybe going into the void (`/dev/null`)? We don't know.

### The Fix

A `systemd timer` does the same thing as `cron`, but with main character energy. 

You write a quick Timer Unit `backup.timer`:
```ini
[Unit]
Description=Daily DB Backup

[Timer]
OnCalendar=daily
Persistent=true

[Install]
WantedBy=timers.target
```

Because `Persistent=true` is set, if the server was down at midnight, systemd will run the backup *immediately* when the server boots. No missed jobs.

Also, systemd timers implicitly trigger systemd *services*. The timer trigers `backup.service`.
Since it triggers a service:
* You get full `journald` logging for free out of `stdout`/`stderr`.
* It won’t run a second instance if the first one is still running.
* You can set resource limits (`cgroups`) on it, like "The backup job cannot use more than 10% CPU."

`systemd` timers replaced `cron` because they actually care about your infrastructure's mental health.

---
**Next:** [Module 09: Boot Process](../09-boot/01-boot-process.md)
