# What "Cloud" Actually Is

> **Cloud computing is renting someone else's computers, paying by the hour, and pretending they're yours. That's it. Everything else is marketing.**

---

## 🟢 The Core Idea

```
Before cloud:
  "We need 10 servers"
  → Buy servers ($50,000)
  → Wait 6 weeks for delivery
  → Rack them in a data center ($10,000/month)
  → Hire someone to maintain them
  → Server breaks at 3am → you fix it
  → Need 20 servers next month? Buy 10 more. Wait 6 weeks.
  → Only need 5 now? 15 servers sitting idle. Still paying.

With cloud:
  "We need 10 servers"
  → Click a button (or run terraform apply)
  → 10 servers in 30 seconds
  → Pay ~$0.05/hour per server
  → Need 20 next month? Click again.
  → Need 5? Terminate 15. Stop paying.
  → Server breaks? Cloud provider replaces it automatically.
```

---

## 🟢 How It Works: Virtualization

```
Physical server (real hardware):
┌──────────────────────────────────────────┐
│            Physical Machine              │
│     64 CPU cores, 256 GB RAM             │
│                                          │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ │
│  │  VM 1    │ │  VM 2    │ │  VM 3    │ │
│  │ 8 cores  │ │ 16 cores │ │ 4 cores  │ │
│  │ 32 GB    │ │ 64 GB    │ │ 16 GB    │ │
│  │ Ubuntu   │ │ Amazon   │ │ Windows  │ │
│  │          │ │ Linux    │ │ Server   │ │
│  │ Your app │ │ Other    │ │ Another  │ │
│  │          │ │ company  │ │ company  │ │
│  └──────────┘ └──────────┘ └──────────┘ │
│                                          │
│         Hypervisor (splits the hardware) │
└──────────────────────────────────────────┘

Multi-tenancy:
  - Multiple customers share one physical machine
  - Each VM is isolated (can't see other VMs)
  - Hypervisor enforces boundaries
  - You think you have a whole machine — you don't
```

---

## 🟢 CapEx vs OpEx

```
CapEx (Capital Expenditure) — Traditional:
  "Buy $500,000 of servers upfront"
  → Large upfront investment
  → Depreciate over 3-5 years
  → You own the hardware
  → You're stuck with it even if needs change
  → Accountants like: predictable depreciation

OpEx (Operational Expenditure) — Cloud:
  "Pay $8,000/month for compute"
  → No upfront investment
  → Pay as you go
  → Scale up or down anytime
  → Stop paying when you stop using
  → Accountants like: predictable operating costs
  → CFOs sometimes dislike: costs can spike unexpectedly
```

### The Real Reasons Companies Use Cloud

```
Reason 1: Speed
  Traditional: "Need a server" → 6 weeks (procurement, shipping, racking)
  Cloud:       "Need a server" → 30 seconds

Reason 2: Elasticity
  Black Friday: traffic 10x → spin up 10x servers → scale down after
  Traditional: buy 10x servers → 51 weeks of waste

Reason 3: No Maintenance
  You: "Server disk failed"
  Cloud: automatically moved to healthy hardware
  You: didn't even notice

Reason 4: Managed Services
  You: "I need a database"
  Cloud: Here's RDS. We handle backups, patching, failover, replication.
  You: Just connect and query.

Reason 5: Global Reach
  "Deploy in Tokyo, Frankfurt, and Virginia"
  Traditional: Build 3 data centers ($$$$$)
  Cloud: 3 terraform resources
```

---

## 🟡 Cloud Service Models

```
┌─────────────────────────────────────────────────────────┐
│                    You Manage Less →                     │
│                                                         │
│  On-Premise     IaaS          PaaS         SaaS         │
│  ──────────     ────          ────         ────         │
│  Application    Application   Application  ────         │
│  Data           Data          Data         ────         │
│  Runtime        Runtime       ────         ────         │
│  Middleware      Middleware    ────         ────         │
│  OS             OS            ────         ────         │
│  Virtualization ────          ────         ────         │
│  Servers        ────          ────         ────         │
│  Storage        ────          ────         ────         │
│  Networking     ────          ────         ────         │
│                                                         │
│  "I manage      "I manage     "I manage    "I manage    │
│   everything"    app + OS"     app only"    nothing"     │
│                                                         │
│  Example:       EC2           Heroku       Gmail        │
│  Your rack      GCE           App Engine   Salesforce   │
│                 Azure VMs     Cloud Run    Slack        │
└─────────────────────────────────────────────────────────┘

IaaS = Infrastructure as a Service
  → Rent VMs, storage, network
  → You manage everything ON the VM
  → Examples: AWS EC2, GCP Compute Engine, Azure VMs

PaaS = Platform as a Service
  → Give them your code, they run it
  → You only manage the application
  → Examples: Heroku, Google App Engine, AWS Elastic Beanstalk

SaaS = Software as a Service
  → Just use it through a browser
  → You manage nothing
  → Examples: Gmail, Slack, Salesforce

FaaS = Function as a Service (Serverless)
  → Give them a function, they run it when triggered
  → You manage nothing except the function code
  → Examples: AWS Lambda, Google Cloud Functions, Azure Functions
```

---

## 🟡 The Shared Responsibility Model

```
"The cloud provider secures the cloud.
 You secure what's IN the cloud."

AWS's responsibility:                Your responsibility:
─────────────────────                ─────────────────────
Physical data centers                Your application code
Hardware                             Your data
Hypervisor                           IAM (who can access what)
Network infrastructure               Security groups/firewalls
                                     Encryption settings
                                     OS patches (on EC2)
                                     Database passwords

Translation:
  AWS: "The building won't catch fire"
  You: "Lock your apartment door"
```

---

## 🟡 Cloud Economics

### When Cloud Is Cheaper

```
✅ Variable workloads (traffic spikes)
✅ Startups (no upfront investment)
✅ Short-lived projects (experiment and terminate)
✅ Need global presence quickly
✅ Small teams (no infra team)
```

### When Cloud Is More Expensive

```
❌ Predictable, constant workloads (24/7/365)
❌ Very large scale (thousands of servers)
❌ Data transfer heavy (egress fees add up)
❌ Companies with existing data centers
❌ Compliance requiring on-premise data

Example:
  1 server running 24/7 for 3 years:
    AWS EC2 (on-demand): ~$30,000
    AWS EC2 (reserved):  ~$15,000
    Buy your own:        ~$5,000
    
  But: you also need:
    Data center space, power, cooling, network,
    redundant hardware, someone to manage it...
```

---

## 🔴 Multi-Cloud vs Single Cloud

```
Single Cloud (90% of companies):
  ✅ Simpler — one set of tools, one API, one billing
  ✅ Better integration — services work together
  ✅ Deeper expertise — team learns one platform well
  ❌ Vendor lock-in
  ❌ Single point of negotiation leverage

Multi-Cloud (marketing says everyone does it):
  ✅ Avoid lock-in (in theory)
  ✅ Best-of-breed services
  ❌ 2-3x operational complexity
  ❌ Team must learn multiple platforms
  ❌ Integration between clouds is painful
  ❌ Networking between clouds costs $$$

Reality:
  Most companies: "We're multi-cloud" means "We use 
  AWS for everything + one team accidentally used GCP once"
  
Honest advice:
  Pick one cloud. Go deep. Use Kubernetes if you need
  to keep the escape hatch open.
```

---

**Previous:** [README](./README.md)  
**Next:** [02. Compute Abstractions](./02-compute-abstractions.md)
