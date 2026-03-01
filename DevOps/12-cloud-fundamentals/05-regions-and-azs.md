# Regions and Availability Zones

> **A Region is a geographic area. An Availability Zone is an isolated data center within a region. Deploy across AZs for resilience. Deploy across regions for disaster recovery.**

---

## 🟢 The Physical Hierarchy

```
Cloud Provider (AWS)
├── Region: us-east-1 (N. Virginia)
│   ├── AZ: us-east-1a (Data Center Cluster A)
│   ├── AZ: us-east-1b (Data Center Cluster B)
│   ├── AZ: us-east-1c (Data Center Cluster C)
│   ├── AZ: us-east-1d
│   ├── AZ: us-east-1e
│   └── AZ: us-east-1f
├── Region: eu-west-1 (Ireland)
│   ├── AZ: eu-west-1a
│   ├── AZ: eu-west-1b
│   └── AZ: eu-west-1c
├── Region: ap-southeast-1 (Singapore)
│   ├── AZ: ap-southeast-1a
│   ├── AZ: ap-southeast-1b
│   └── AZ: ap-southeast-1c
└── ... (30+ regions worldwide)

Key facts:
  - Each AZ is physically separate (different building, power, cooling)
  - AZs within a region connected by high-speed private fiber
  - Latency between AZs: ~1-2ms
  - Latency between regions: 50-300ms (depends on distance)
```

---

## 🟢 Why Multiple AZs Matter

```
Single AZ deployment:
  ┌───────────────────────┐
  │     us-east-1a        │
  │  ┌──────┐ ┌──────┐   │
  │  │ App  │ │  DB  │   │
  │  └──────┘ └──────┘   │
  └───────────────────────┘
  
  AZ-a has power outage → EVERYTHING is down 💥
  No redundancy. Full outage.

Multi-AZ deployment:
  ┌───────────────────────┐  ┌───────────────────────┐
  │     us-east-1a        │  │     us-east-1b        │
  │  ┌──────┐ ┌──────┐   │  │  ┌──────┐ ┌──────┐   │
  │  │ App  │ │DB    │   │  │  │ App  │ │DB    │   │
  │  │  #1  │ │Primary│  │  │  │  #2  │ │Standby│  │
  │  └──────┘ └──────┘   │  │  └──────┘ └──────┘   │
  └───────────────────────┘  └───────────────────────┘
          │                            │
          └────────── ALB ─────────────┘
  
  AZ-a has power outage:
    → App #1 down, DB failover to standby in AZ-b
    → ALB routes all traffic to App #2
    → Users: maybe notice 30s of errors during failover
    → Recovery: automatic ✅
```

---

## 🟢 Choosing a Region

### Decision Factors

```
1. LATENCY — Where are your users?
   Users in Europe → eu-west-1 (Ireland) or eu-central-1 (Frankfurt)
   Users in US     → us-east-1 (Virginia) or us-west-2 (Oregon)
   Users in Asia   → ap-southeast-1 (Singapore) or ap-northeast-1 (Tokyo)

2. COMPLIANCE — Where must data stay?
   GDPR (EU data) → eu-* regions
   Data sovereignty laws → local region
   Government contracts → specific approved regions (GovCloud)

3. COST — Regions have different prices!
   us-east-1 (Virginia) → Often cheapest (largest, most capacity)
   sa-east-1 (São Paulo) → 40-80% more expensive
   
   Same service can cost very different amounts:
     t3.medium in us-east-1: $0.0416/hr
     t3.medium in ap-southeast-1: $0.0468/hr

4. SERVICE AVAILABILITY — Not all services in all regions
   New AWS services launch in us-east-1 first
   Some regions have fewer services
   Check: AWS regional services page

5. DISASTER RECOVERY — Secondary region?
   Primary: us-east-1 → DR: us-west-2
   Primary: eu-west-1 → DR: eu-central-1
   Failover should be geographically separated
```

---

## 🟡 Architecture Patterns

### Single Region, Multi-AZ (Most Common)

```
Best for: 99.99% of applications

Region: us-east-1
┌────────────────────────────────────────────────┐
│                                                │
│  AZ-a                    AZ-b                  │
│  ┌──────────────┐  ┌──────────────┐           │
│  │ App (x2)     │  │ App (x2)     │           │
│  │ Worker (x1)  │  │ Worker (x1)  │           │
│  │ DB Primary   │  │ DB Standby   │           │
│  │ Redis Primary│  │ Redis Replica│           │
│  └──────────────┘  └──────────────┘           │
│                                                │
│  ALB spans both AZs                           │
│  S3 automatically multi-AZ                    │
│  RDS multi-AZ = automatic failover            │
│                                                │
└────────────────────────────────────────────────┘

Benefits:
  ✅ Survives single AZ failure
  ✅ Low complexity
  ✅ Low latency between AZs (~1ms)
  ✅ Affordable

Limitations:
  ❌ Entire region outage = downtime
  ❌ Single geographic presence
```

### Multi-Region Active-Passive (DR)

```
Primary: us-east-1 (serves all traffic)
DR:      us-west-2 (standby, receives replicated data)

Active Region (us-east-1)     Passive Region (us-west-2)
┌─────────────────────┐       ┌─────────────────────┐
│  App (running)      │       │  App (standby)      │
│  DB (primary)       │──────→│  DB (read replica)  │
│  Cache (active)     │ async │  Cache (cold)       │
│  S3 (primary)       │──────→│  S3 (replicated)    │
└─────────────────────┘       └─────────────────────┘

Failover:
  1. DNS switches from us-east-1 to us-west-2
  2. DB replica promoted to primary
  3. Recovery time: 15-60 minutes
  
Cost: ~1.5x (paying for standby resources)
```

### Multi-Region Active-Active (Global)

```
For: Applications serving users worldwide with low latency

Region: us-east-1            Region: eu-west-1
┌─────────────────┐          ┌─────────────────┐
│  App (active)   │◄────────→│  App (active)   │
│  DB (primary)   │  bi-dir  │  DB (primary)   │
│  Cache          │  replica │  Cache          │
└─────────────────┘          └─────────────────┘
        │                            │
        ▼                            ▼
  US users routed             EU users routed
  here by DNS                 here by DNS

Challenges:
  ❌ Extremely complex
  ❌ Data consistency issues (split-brain)
  ❌ 2-3x cost
  ❌ Need global database (DynamoDB Global Tables, CockroachDB)
  
Only for: Netflix, Spotify, global SaaS at massive scale
```

---

## 🟡 Latency Considerations

```
Same AZ:            0.1 - 0.5 ms
Cross AZ:           1 - 2 ms
Same Region:        1 - 5 ms
Cross Region:       50 - 300 ms (depends on distance)
Cross Continent:    100 - 300 ms

Impact on architecture:
  Database reads:
    Same AZ → 0.5ms per query (great)
    Cross AZ → 2ms per query (fine for most apps)
    Cross Region → 150ms per query (way too slow for sync calls)
    
  API calls between microservices:
    Same AZ: negligible
    Cross AZ: acceptable
    Cross Region: MUST be async or cached
    
Rule: Keep tightly-coupled services in the same region.
      Use async patterns (queues, events) for cross-region.
```

---

## 🟡 Data Transfer Costs

```
AWS data transfer pricing (often the surprise on the bill):

  Inbound (internet → AWS):    FREE
  Within same AZ:               FREE
  Cross AZ (same region):       $0.01/GB each direction
  Cross Region:                 $0.02/GB
  Outbound (AWS → internet):    $0.09/GB (first 10 TB)
  
  Example bill breakdown:
    10 TB/month outbound = $900/month in transfer alone
    Cross-AZ traffic for multi-AZ app: usually $20-100/month
    
  NAT Gateway data processing: $0.045/GB
    50 GB/month through NAT = $2.25 (reasonable)
    5 TB/month through NAT = $225 (ouch)

GCP:
  Slightly cheaper egress
  No charge for cross-AZ within same region!

Advice:
  ✅ Use CDN for static content (reduces egress)
  ✅ Compress data transfers
  ✅ VPC endpoints for AWS service access (avoid NAT)
  ✅ Monitor data transfer in your bill
```

---

## 🔴 Real Outage Examples

```
2017: AWS S3 us-east-1 outage
  → Typo in maintenance command
  → S3 down for 4 hours
  → Half the internet broke (sites depend on S3)
  → Lesson: Multi-region for critical data

2020: AWS Kinesis us-east-1 outage  
  → Affected CloudWatch, Lambda, many services
  → 8+ hours of degraded service
  → Lesson: Services have dependencies you don't see

2023: Azure Australia East outage
  → Power event in data center
  → Services degraded for 12+ hours
  → Lesson: Multi-AZ isn't enough if the region has issues

Common theme: 
  Region-level outages are rare (1-2/year per region)
  AZ-level outages are more common (3-5/year)
  Multi-AZ protects against most failures
  Multi-region only needed for SLA > 99.95%
```

---

**Previous:** [04. Networking Abstractions](./04-networking-abstractions.md)  
**Next:** [06. Cloud Mapping](./06-cloud-mapping.md)
