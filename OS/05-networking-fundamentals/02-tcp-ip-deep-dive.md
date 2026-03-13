# TCP/IP & The Kernel Network Stack

**How the Internet Actually Fits Inside Your RAM**

🟡 **Intermediate** | 🔴 **Advanced**

---

## Introduction

So you wrote an API and called `res.send("Hello World")`. 
You think it instantly teleported to the user's iPhone? Nah, bro. It had to survive the **Linux Network Stack** first.

Networking isn't just cables and routers. A massive chunk of the internet's logic—TCP handshakes, retries, congestion control, window sizing—is literally just a C program running inside the Linux kernel. 

---

## The Journey of a Packet (Outbound)

When your Node/Go/Python app decides to send data over a socket, here is the gauntlet it runs:

### 1. User Space to Socket Buffer (The Drop-off)
Your app calls `write(fd, "Hello", 5)`. 
The kernel takes that string and copies it into a queue inside RAM called the **Socket Send Buffer** (`wmem`).
Once it's in the buffer, the kernel tells your app: *"I got it! You can go back to work."*
(This is why `write()` is so fast—you're just writing to RAM. The data hasn't hit the network yet!)

### 2. The TCP Layer (The Accountant)
The kernel's TCP subsystem wakes up. It looks at the data in the send buffer and says:
- "Let's chop this huge JSON payload into 1500-byte chunks." (Segmentation / MTU).
- "Let's slap a TCP Header on it (Source Port: 8080, Dest Port: 443, Sequence Number: 1234)."
- "Let's start a timer. If I don't get an ACK (Acknowledgement) from the client in 200ms, I'm re-sending this exact chunk." (This is why TCP is "Reliable").

### 3. The IP Layer (The Navigator)
TCP passes the chunk down to the IP layer. The IP layer adds an IP Header (Source IP, Dest IP).
It then checks the **Routing Table** (`ip route`) to figure out which network card (interface) this packet should leave from. 
"Ah, destination `8.8.8.8`? Send it out the Default Gateway on `eth0`."

### 4. The Data Link Layer (MAC Addresses / Arp)
IP addresses are for across the world; MAC addresses are for the local room. 
Before it leaves the `eth0` card to hit your home Router, it needs the Router's physical MAC address. It looks this up in the kernel's **ARP Cache** (`ip neigh`). Adds an Ethernet header.

### 5. Ring Buffers and the NIC Driver (The Exit)
The kernel places the fully-assembled packet into a **Ring Buffer** (a queue shared between the CPU and the physical Network Interface Card).
It rings a bell (an interrupt/doorbell) to tell the hardware: "Yo, ping this over the wire." 
The network card flashes its little LED and sends electrons over the copper. Slay. ✨

---

## The Journey of a Packet (Inbound)

Now imagine your server is receiving a high volume traffic spike. It's reversing the process, but the stakes are higher.

1. **Hardware Interrupt**: The Network Card receives electrons. It writes the packet into RAM (using Direct Memory Access - DMA). It then rudely interrupts the CPU: "PACKET HERE. STOP WHAT YOU'RE DOING."
2. **NAPI (New API) / SoftIRQ**: In the 90s, if you got 100,000 packets per second, you got 100,000 CPU interrupts. Your CPU would literally freeze doing nothing but acknowledging interrupts (Interrupt Storm). 
Linux fixed this with **NAPI**. NAPI says: "Okay, turn off the hardware interrupts. I will just continuously poll the Ring Buffer in a loop until it's empty." (This happens in the `ksoftirqd` kernel threads).

3. It parses the Ethernet header -> IP Header -> TCP Header.
4. TCP puts the payload into the **Socket Receive Buffer** (`rmem`).
5. Your application's `epoll_wait()` loop wakes up! 
6. You call `read(fd)` and pull the data from the kernel's buffer into your app's memory to parse the JSON.

---

## TCP Tuning: When Default Vibes Aren't Enough

If you run a massive backend (like Discord, Netflix, or a high-freq trading app), the default Linux network settings are literally holding you back.

### The Backlog (Drop the line)
When a client tries to connect, TCP does a 3-way handshake (`SYN -> SYN-ACK -> ACK`).
While waiting for the final `ACK`, the connection sits in the `SYN Backlog`. 
If you get hit by a traffic spike and this queue fills up, the Linux kernel will literally start throwing new user connections in the trash. No error message. Just *dropped*.

**The Fix:**
```bash
# Increase the queue size for pending connections
$ sudo sysctl -w net.ipv4.tcp_max_syn_backlog=4096

# Increase the queue size for fully established connections waiting for the app to accept()
$ sudo sysctl -w net.core.somaxconn=4096
```

### Buffer Bloat
If your server sends massive files but the user is on a terrible 3G connection, the user can't read fast enough. TCP realizes this and slows down. The un-sent data backs up in your server's RAM (The Send Buffer). If you have 10,000 slow clients, they might eat all your RAM!
Linux auto-tunes buffer sizes, but you can restrict or expand them.

```bash
# Min, Default, Max sizes for TCP Receive/Send buffers
$ sysctl net.ipv4.tcp_rmem
net.ipv4.tcp_rmem = 4096    131072    6291456
```

## Key Takeaways
1. A packet travels from App -> TCP/UDP -> IP -> Ethernet/ARP -> Hardware ring buffer.
2. The kernel's TCP stack handles all reliability, retries, and fragmentation so you don't have to in user space.
3. Network cards use DMA (Direct Memory Access) and polling (NAPI) to avoid interrupting the CPU 100,000 times a second.
4. "Connection Refused" or dropped packets at scale usually means you hit an OS Queue limit, not an app bug.

---
**Next:** [Module 06: Network Namespaces and veth Pairs](../06-networking-deep/01-namespaces-veth-bridges.md)
