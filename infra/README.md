# Zippyra Infrastructure — Terraform Apply Order & Guidelines

## Apply Order

Infrastructure modules must be applied in this exact dependency order:

```
1. VPC          (networking foundation)
2. IAM          (service roles, OIDC provider requires EKS — created together)
3. RDS + ElastiCache  (can be applied in parallel — both depend only on VPC)
4. RDS Proxy    (depends on RDS)
5. MSK          (depends on VPC)
6. Glue Schema Registry (no VPC dependency)
7. Secrets Manager (no VPC dependency)
8. S3           (no VPC dependency)
9. EKS          (depends on VPC, IAM created concurrently)
10. CloudFront  (depends on S3)
11. WAF         (depends on ALB from EKS ingress — apply last)
```

## Quick Start (Pilot)

```bash
cd infra/environments/pilot
terraform init
terraform plan -out=plan.tfplan
terraform apply plan.tfplan
```

## Production Deployment

> ⚠️ **CRITICAL**: `terraform plan` output for the production environment should
> ALWAYS be reviewed by a second person before apply, given how much of the
> platform depends on this being correct.

```bash
cd infra/environments/production
terraform init
terraform plan -out=plan.tfplan
# STOP — have a second engineer review plan.tfplan
terraform apply plan.tfplan
```

## DPDP Data Localization

All Terraform modules enforce `aws_region` via a validation block:

```hcl
variable "aws_region" {
  validation {
    condition     = contains(["ap-south-1", "ap-south-2"], var.aws_region)
    error_message = "Must use an India region for DPDP data localization."
  }
}
```

Attempting to deploy to any non-India region will fail at `terraform plan`.

## Module Reference

| Module | Purpose | Pilot Sizing | Production Sizing |
|--------|---------|-------------|-------------------|
| `vpc` | VPC, 3-AZ subnets, t3.nano NAT instance | 10.0.0.0/16 | 10.1.0.0/16 |
| `rds` | PostgreSQL 16 Multi-AZ | db.t4g.large, 50GB | db.r6g.xlarge, 200GB |
| `rds_proxy` | Connection pooling (660→320 conns) | Enabled | Enabled |
| `elasticache` | 6 Redis clusters (exit_token=noeviction) | cache.t4g.medium ×1 | cache.r6g.large |
| `msk` | Kafka 3.5.1, all topics pre-created | kafka.t3.small ×3 | kafka.m5.large ×3 |
| `glue_schema_registry` | Schema evolution safety | Enabled | Enabled |
| `eks` | Kubernetes 1.29, IRSA | t3.medium ×3 (80% spot) | m5.xlarge ×6 (60% spot) |
| `iam` | Per-service IRSA roles (least-privilege) | 22 roles | 22 roles |
| `secrets_manager` | Per-service secret paths | 22+1 secrets | 22+1 secrets |
| `s3` | invoices/products/media buckets | Glacier@30d, DR replication | Same |
| `cloudfront` | Products CDN, 24h TTL | PriceClass_200 | PriceClass_200 |
| `waf` | Rate limit, SQLi, Razorpay IP allowlist | Enabled | Enabled |

## Sensitive Files

- `infra/environments/production/terraform.tfvars` — **GITIGNORED**. Real values injected via CI, never committed.
- All secrets are in AWS Secrets Manager, never in Terraform state or committed files.
