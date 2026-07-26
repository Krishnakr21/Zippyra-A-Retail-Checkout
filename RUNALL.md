# Zippyra — Step-by-Step Developer Execution Guide

## Prerequisites
- **Go:** 1.22+
- **Node.js:** 18+ (PNPM 8+)
- **Flutter:** 3.19+ & Melos (`dart pub global activate melos`)
- **Docker & Docker Compose**

## 1. Setup & Installation
Run the one-shot bootstrap script:
```bash
./install.sh
```

## 2. Running Local Infrastructure
Start Postgres, Redis, Kafka, ClickHouse, and Elasticsearch:
```bash
docker-compose up -d
```

## 3. Running Services
Start all microservices, web dashboards, and mobile emulators:
```bash
./run_all.sh
```

## Service Endpoints
- **Retailer Web:** http://localhost:3000
- **Admin Web:** http://localhost:3001
- **Chain HQ Web:** http://localhost:3002
- **Kong Gateway:** http://localhost:8000
