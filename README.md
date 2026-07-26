# Zippyra Platform Monorepo

Zippyra is a self-checkout, anti-theft & retail-operations platform comprising:
- **Backend:** Go microservices (REST + Kafka events), polyglot persistence (PostgreSQL, Redis, TimescaleDB, ClickHouse, Elasticsearch).
- **Web:** Next.js 14 (App Router, TypeScript, Tailwind CSS) — Retailer Dashboard, Admin Platform, Chain HQ Dashboard.
- **Mobile:** Flutter monorepo (Melos-managed) sharing `zippyra_core` for Customer and Store Staff apps.
- **Infra:** Terraform IaC (AWS VPC, RDS, MSK, EKS, IAM, Secrets, S3).

## Quick Start

```bash
# 1. Bootstrap environment
./install.sh

# 2. Run local dev environment
./run_all.sh

# 3. Shutdown services cleanly
./stop_all.sh
```

See [RUNALL.md](./RUNALL.md) for full execution steps and [PRE_PILOT_CHECKLIST.md](./PRE_PILOT_CHECKLIST.md) for launch readiness.
