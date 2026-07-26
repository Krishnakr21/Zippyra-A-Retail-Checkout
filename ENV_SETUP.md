# Zippyra Monorepo Environment Configuration & Secrets Architecture (`ENV_SETUP.md`)

This document defines the unified, typed, fail-fast environment configuration architecture across all Zippyra surfaces (**Backend Go Microservices**, **Web Next.js Apps**, and **Mobile Flutter Apps**).

---

## 🔒 Golden Security Principle
> **No environment ever reads a committed secret.**
> - **Local Development**: Uses local `.env` files (strictly `.gitignore`d).
> - **Backend Staging / Production**: AWS Secrets Manager (`zippyra/{environment}/{service}/`) with hot-reloading.
> - **Web Staging / Production**: CI/CD-injected build-time & runtime environment variables.
> - **Mobile Release Builds**: Injected via `--dart-define-from-file=env/prod.json` (never embedded in dotenv assets inside release APK/IPA).

---

## 🖥️ Surface Guides & Configuration Code

### 1. Backend (Go Microservices)
- **Reference Template**: [backend/.env.example](file:///Users/krishna/Downloads/Fatima/Zippyra/backend/.env.example)
- **Documentation**: [backend/README-env.md](file:///Users/krishna/Downloads/Fatima/Zippyra/backend/README-env.md)
- **Typed Config Code**: [backend/shared/config/config.go](file:///Users/krishna/Downloads/Fatima/Zippyra/backend/shared/config/config.go)
- **Secrets Manager**: [backend/shared/config/secretsmanager.go](file:///Users/krishna/Downloads/Fatima/Zippyra/backend/shared/config/secretsmanager.go)
- **Key Features**:
  - Fail-fast startup validator (`config.Load(serviceName)`).
  - Aggregated error reporting (lists ALL missing required variables in one pass).
  - Production Safety Invariants (`DB_SSL_MODE=require`, `RDS_IAM_AUTH_ENABLED=true`, `WAF_ENABLED=true`, `KAFKA_SECURITY_PROTOCOL!=PLAINTEXT`).
  - DPDP Data Localization Enforcement (`AWS_REGION` must be `ap-south-1` or `ap-south-2`).
  - PII-masked startup integration logging.

### 2. Web Applications (Next.js - Retailer, Admin, HQ)
- **Reference Template**: [web/.env.example](file:///Users/krishna/Downloads/Fatima/Zippyra/web/.env.example)
- **Typed Env Package**: [web/packages/env/index.ts](file:///Users/krishna/Downloads/Fatima/Zippyra/web/packages/env/index.ts)
- **Key Features**:
  - `serverSchema` vs `clientSchema` (`NEXT_PUBLIC_*`).
  - Build-time validation using Zod schemas.
  - Automatic secret leak prevention (`assertNoSecretLeaks`).
  - NextAuth secret length enforcement (minimum 32 characters in production).

### 3. Mobile Applications (Flutter - Customer App & Staff App)
- **Customer App Reference**: [mobile/customer_app/.env.example](file:///Users/krishna/Downloads/Fatima/Zippyra/mobile/customer_app/.env.example)
- **Customer App Guide**: [mobile/customer_app/README_ENV.md](file:///Users/krishna/Downloads/Fatima/Zippyra/mobile/customer_app/README_ENV.md)
- **Staff App Reference**: [mobile/staff_app/.env.example](file:///Users/krishna/Downloads/Fatima/Zippyra/mobile/staff_app/.env.example)
- **Staff App Guide**: [mobile/staff_app/README_ENV.md](file:///Users/krishna/Downloads/Fatima/Zippyra/mobile/staff_app/README_ENV.md)
- **Typed Core Config Code**: [mobile/packages/zippyra_core/lib/config/app_config.dart](file:///Users/krishna/Downloads/Fatima/Zippyra/mobile/packages/zippyra_core/lib/config/app_config.dart)
- **Key Features**:
  - `AppConfig.baseUrl` dynamically resolves endpoint based on `APP_ENV` (`development` | `staging` | `production`).
  - Debug mode uses `flutter_dotenv` (`.env`).
  - Release builds use `--dart-define-from-file=env/prod.json`.
  - Startup assertion (`AppConfig.validate()`): Throws `StateError` if `APP_ENV == 'production'` and `CERT_PIN_SHA256_PRIMARY` is empty.
