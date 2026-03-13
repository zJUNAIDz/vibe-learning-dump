# Load Balancing and Zero Downtime

**Deploying Software Without Waking Up the Users**

🔴 **Advanced**

---

## 1. The Kernel Load Balancer: IPVS

If your company gets famous and one server is handling 50,000 requests a second, Node.js or Nginx is going to start sweating. You scale horizontally by buying 5 servers. But how do you route traffic between them?

You could use a hardware load balancer, or you can use the Linux Kernel itself.

**IP Virtual Server (IPVS)** is a load-balancing module built directly into the Linux kernel (specifically sitting inside netfilter/iptables logic). 
Because it runs purely in Kernel Space, it completely skips the overhead of copying packets up to User Space boundaries (unlike Nginx or HAProxy acting as reverse proxies). 

IPVS can forward millions of packets per second with basically zero CPU overhead. It’s what Kubernetes uses under the hood (in modern setups) to route traffic to your Pods when you create a `Service` using `kube-proxy` in IPVS mode.

Algorithms IPVS supports out of the box:
*   **Round Robin (rr):** You get a request, and you get a request! 
*   **Least Connections (lc):** Give the packet to the server currently handling the fewest active TCP sessions.
*   **Source Hashing (sh):** Stickiness. If Client IP `A` went to Server `B` last time, predictably route them to Server `B` again.

---

## 2. The Graceful Restart (Zero Downtime)

The most terrifying moment in a backend developer's life is typing `systemctl restart my-app`. 
For the 2 seconds it takes your app to shut down, load its config, and start listening on port 8080 again, any customer clicking "Checkout" gets a `502 Bad Gateway` error. That is unacceptable.

How do we restart a process without dropping a single packet?

### The `SO_REUSEPORT` Magic Spell 🔥

Historically, only one process was allowed to listen on a specific port (e.g., `8080`). If a second process tried to start on `8080`, it got an `EADDRINUSE` error.

Linux 3.9 introduced the `SO_REUSEPORT` socket option.
If multiple processes all bind to the same port using this flag, the kernel will physically intercept incoming TCP SYN packets and **load balance them evenly across all the listening processes**.

**How to achieve a Zero-Downtime Deployment in Node/Go:**
1. App `v1.0` is running on `8080` using `SO_REUSEPORT`.
2. You deploy App `v2.0` in the background. It *also* binds to `8080` using `SO_REUSEPORT`.
3. At this exact moment, you have two instances running. The Kernel starts sending 50% of your new customer traffic to `v1.0` and 50% to `v2.0`.
4. You send a signal (like `SIGTERM`) to `v1.0`.
5. `v1.0` stops accepting *new* connections, finishes processing the current requests, and cleanly exits (Graceful Shutdown).
6. Now `v2.0` handles 100% of the traffic.

Zero dropped packets. Zero failed checkouts. Absolute perfection.

---

## 3. High Availability Floating IPs: `keepalived`

What if the primary database server physically bursts into flames? 
If your web servers are hardcoded to point to IP `10.0.0.5`, your app is dead.

Enter **VRRP (Virtual Router Redundancy Protocol)** via a Linux tool called `keepalived`.

You have two DB servers: `DB-Active` and `DB-Standby`. 
You assign a **Virtual IP (VIP)**, let's say `10.0.0.99`.

* `keepalived` runs on both servers. They constantly ping each other saying "I'm alive!" heartbeat over the network.
* `DB-Active` dynamically claims the `10.0.0.99` IP address on its `eth0` interface and answers ARP requests for it. Muted flex. 
* If someone trips over the power cord for `DB-Active`, the heartbeats stop.
* `DB-Standby` realizes `DB-Active` has been ghosted. Within milliseconds, `DB-Standby` assigns `10.0.0.99` to its own `eth0` interface and sends out an aggressive "Gratuitous ARP" broadcast to the network switches: *“Yo! 10.0.0.99 is MY MAC address now. Send all traffic to me."*

Your web servers casually keep sending traffic to `10.0.0.99` and have literally no idea that the underlying hardware just swapped. High Availability achieved.

---
**Next:** [Module 13: Failure Stories](../13-failure-stories/01-when-things-break.md)
