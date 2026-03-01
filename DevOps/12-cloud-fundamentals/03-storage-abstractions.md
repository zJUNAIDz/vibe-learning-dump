# Storage Abstractions

> **Object storage for files, block storage for disks, file storage for shared filesystems. Know the difference and you'll never pick the wrong one.**

---

## 🟢 Three Types of Cloud Storage

```
┌───────────────────────────────────────────────────────┐
│                   Storage Types                        │
│                                                       │
│  Object Storage     Block Storage     File Storage    │
│  ─────────────     ─────────────     ────────────     │
│  S3, GCS           EBS, PD           EFS, Filestore   │
│                                                       │
│  Store: files      Store: raw disk   Store: shared    │
│  Access: HTTP API  Access: mount     Access: NFS/SMB  │
│                    to ONE server     to MANY servers   │
│                                                       │
│  Like: Google      Like: USB hard    Like: shared     │
│  Drive API         drive attached    network drive    │
│                    to your PC        at the office    │
│                                                       │
│  Unlimited size    Fixed size         Scales           │
│  Cheapest          Mid-range          Most expensive   │
│  Highest latency   Lowest latency     Mid latency      │
└───────────────────────────────────────────────────────┘
```

---

## 🟢 Object Storage (S3, GCS, Azure Blob)

Objects are files stored with metadata, accessed via HTTP API. No filesystem, no directories (just prefixes that look like directories).

```
S3 bucket structure:
  my-bucket/
  ├── images/photo1.jpg       ← Key: "images/photo1.jpg"
  ├── images/photo2.jpg       ← Key: "images/photo2.jpg"
  ├── backups/db-2024-01.gz   ← Key: "backups/db-2024-01.gz"
  └── config.json             ← Key: "config.json"

"images/" is NOT a folder. It's part of the key name.
S3 is a flat key-value store that looks like a filesystem.
```

### Common Operations

```bash
# AWS CLI
aws s3 cp file.txt s3://my-bucket/file.txt          # Upload
aws s3 cp s3://my-bucket/file.txt ./file.txt         # Download
aws s3 ls s3://my-bucket/                            # List
aws s3 rm s3://my-bucket/file.txt                    # Delete
aws s3 sync ./build/ s3://my-bucket/static/          # Sync directory

# Presigned URL (temporary access without credentials)
aws s3 presign s3://my-bucket/private-file.pdf --expires-in 3600
```

### Terraform Example

```hcl
resource "aws_s3_bucket" "assets" {
  bucket = "myapp-assets-production"
}

resource "aws_s3_bucket_versioning" "assets" {
  bucket = aws_s3_bucket.assets.id
  versioning_configuration {
    status = "Enabled"    # Keep old versions (protect against deletion)
  }
}

resource "aws_s3_bucket_lifecycle_configuration" "assets" {
  bucket = aws_s3_bucket.assets.id
  
  rule {
    id     = "archive-old-logs"
    status = "Enabled"
    filter {
      prefix = "logs/"
    }
    
    transition {
      days          = 30
      storage_class = "STANDARD_IA"     # Cheaper, less frequent access
    }
    transition {
      days          = 90
      storage_class = "GLACIER"          # Very cheap, retrieval takes hours
    }
    expiration {
      days = 365                         # Delete after 1 year
    }
  }
}
```

### Storage Classes (Cost Tiers)

```
S3 Standard               → Frequently accessed    $0.023/GB/month
S3 Standard-IA             → Infrequent access     $0.0125/GB/month
S3 One Zone-IA             → Infrequent, one AZ    $0.01/GB/month
S3 Glacier Instant         → Archive, instant       $0.004/GB/month
S3 Glacier Flexible        → Archive, minutes-hours $0.0036/GB/month
S3 Glacier Deep Archive    → Rarely accessed        $0.00099/GB/month

Rule of thumb:
  Hot data (accessed daily)     → Standard
  Warm data (accessed monthly)  → Standard-IA
  Cold data (accessed yearly)   → Glacier
  Compliance/audit logs         → Glacier Deep Archive
```

### When to Use Object Storage

```
✅ Static website hosting (HTML, CSS, JS)
✅ User uploads (images, videos, documents)
✅ Application logs
✅ Database backups
✅ Data lake (analytics, ML training data)
✅ Terraform state files
✅ CI/CD artifacts

❌ Database storage (need block storage)
❌ Application code at runtime (need filesystem)
❌ Low-latency reads (use CDN in front of S3)
```

---

## 🟢 Block Storage (EBS, Persistent Disks)

Block storage is a virtual hard drive attached to a VM. It looks and acts like a local disk.

```
EC2 Instance
├── Root volume (EBS): /dev/xvda → / (OS)
│   └── 20 GB, gp3
├── Data volume (EBS): /dev/xvdb → /data (application data)
│   └── 100 GB, gp3
└── Ephemeral storage: /dev/nvme0n1 → /tmp (lost on stop!)
    └── Instance store (fast but temporary)
```

### EBS Volume Types

```
gp3 (General Purpose SSD):
  → Most workloads: web apps, dev environments
  → 3000 baseline IOPS, up to 16,000
  → $0.08/GB/month

io2 (Provisioned IOPS SSD):
  → Databases requiring consistent IOPS
  → Up to 64,000 IOPS
  → $0.125/GB/month + $0.065/IOPS/month

st1 (Throughput Optimized HDD):
  → Big data, log processing
  → Sequential reads, cheap bulk storage
  → $0.045/GB/month

sc1 (Cold HDD):
  → Infrequent access, cheapest block storage
  → $0.015/GB/month
```

### Terraform Example

```hcl
resource "aws_ebs_volume" "data" {
  availability_zone = "us-east-1a"
  size              = 100        # GB
  type              = "gp3"
  iops              = 3000
  throughput        = 125        # MB/s
  encrypted         = true
  
  tags = {
    Name = "myapp-data"
  }
}

resource "aws_volume_attachment" "data" {
  device_name = "/dev/xvdb"
  volume_id   = aws_ebs_volume.data.id
  instance_id = aws_instance.web.id
}
```

### Key Concepts

```
Snapshots:
  → Point-in-time backup of an EBS volume
  → Stored in S3 (cross-AZ durable)
  → Incremental (only changed blocks)
  → Can create new volumes from snapshots
  
Limitations:
  → Attached to ONE instance at a time (usually)
  → Must be in the SAME availability zone as the instance
  → Cannot easily share between instances
  → Size is fixed (can expand, cannot shrink)
```

### When to Use Block Storage

```
✅ Database storage (PostgreSQL, MySQL data directory)
✅ OS disk for VMs
✅ Application data requiring filesystem semantics
✅ High-IOPS workloads (transactional databases)

❌ Shared storage between multiple servers (use EFS)
❌ Archival/backup (use S3 — much cheaper)
❌ Static content serving (use S3 + CloudFront)
```

---

## 🟡 File Storage (EFS, Filestore, Azure Files)

A shared filesystem that multiple servers can mount simultaneously (like a network drive).

```
                    EFS (Elastic File System)
                    ┌─────────────────────┐
                    │   /shared/          │
                    │   ├── configs/      │
                    │   ├── uploads/      │
                    │   └── models/       │
                    └────────┬────────────┘
                             │ NFS protocol
              ┌──────────────┼──────────────┐
              ▼              ▼              ▼
         ┌─────────┐   ┌─────────┐   ┌─────────┐
         │  EC2 #1  │   │  EC2 #2  │   │  EC2 #3  │
         │ (mount   │   │ (mount   │   │ (mount   │
         │  /mnt/   │   │  /mnt/   │   │  /mnt/   │
         │  shared) │   │  shared) │   │  shared) │
         └─────────┘   └─────────┘   └─────────┘
         
All three servers see the SAME files.
Write on EC2 #1 → instantly visible on #2 and #3.
```

### Terraform Example

```hcl
resource "aws_efs_file_system" "shared" {
  creation_token = "myapp-shared"
  encrypted      = true
  
  performance_mode = "generalPurpose"    # or "maxIO"
  throughput_mode  = "bursting"          # or "provisioned"
  
  lifecycle_policy {
    transition_to_ia = "AFTER_30_DAYS"    # Move cold files to cheaper tier
  }
  
  tags = {
    Name = "myapp-shared"
  }
}

resource "aws_efs_mount_target" "shared" {
  count           = length(var.subnet_ids)
  file_system_id  = aws_efs_file_system.shared.id
  subnet_id       = var.subnet_ids[count.index]
  security_groups = [aws_security_group.efs.id]
}
```

### When to Use File Storage

```
✅ Shared config files across multiple servers
✅ CMS content (WordPress, etc.)
✅ Machine learning model files
✅ Legacy apps requiring POSIX filesystem
✅ Kubernetes PersistentVolumes (ReadWriteMany)

❌ High-performance databases (use EBS)
❌ Static web content (use S3)
❌ Backup/archive (use S3 Glacier)
❌ Cost-sensitive workloads (EFS is expensive)
```

---

## 🟡 Comparison Table

| Feature | Object (S3) | Block (EBS) | File (EFS) |
|---------|------------|-------------|------------|
| **Access** | HTTP API | Mount to 1 VM | Mount to many VMs |
| **Protocol** | REST API | Block device | NFS/SMB |
| **Max size** | Unlimited | 64 TB per volume | Petabytes |
| **Performance** | High throughput | Highest IOPS | Moderate |
| **Durability** | 99.999999999% | 99.8-99.999% | 99.999999999% |
| **Price (per GB)** | $0.023 | $0.08 | $0.30 |
| **Use case** | Files, backups, static | Databases, OS disks | Shared storage |
| **Kubernetes** | Via CSI (read-only) | PV (ReadWriteOnce) | PV (ReadWriteMany) |

---

## 🔴 Common Mistakes

```
❌ Using EBS for backups
   → S3 is 3x cheaper and more durable. Use EBS snapshots + S3.

❌ Using EFS for everything
   → EFS is 10x more expensive than S3. Only use when you NEED
     shared filesystem semantics.

❌ Not enabling S3 versioning
   → Accidental deletion = data loss. Enable versioning on 
     important buckets.

❌ Ignoring S3 lifecycle policies
   → Storing old logs at Standard tier costs 10x more than Glacier.
     Set up lifecycle rules.

❌ Putting database on instance store
   → Instance store is EPHEMERAL. Stop the instance = data gone.
     Always use EBS for databases.
```

---

**Previous:** [02. Compute Abstractions](./02-compute-abstractions.md)  
**Next:** [04. Networking Abstractions](./04-networking-abstractions.md)
