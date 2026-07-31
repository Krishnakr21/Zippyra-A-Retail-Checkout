# Zippyra Platform API Versioning & Deprecation Policy

## Overview & Core Principles

The Zippyra platform enforces an explicit API lifecycle and backward-compatibility policy across all 26 microservices:

1. **Version Immutability (`/v1/` Rules)**:
   - Existing `/v1/` routes and API contracts MUST NOT be mutated in a breaking manner.
   - Adding optional fields or non-breaking response parameters is permitted on `/v1/` endpoints.
   - Any genuinely breaking change (field renaming, type changes, field deletions, semantic contract alterations) MUST introduce a new `/v2/` route.

2. **12-Month Deprecation Grace Period**:
   - When a `/v2/` endpoint replaces a `/v1/` endpoint, the `/v1/` route is marked **Deprecated** via `versioning.Deprecated(sunsetDate, migrationURL)`.
   - The `/v1/` route remains fully operational for **12 months** post-deprecation before being decommissioned.
   - Deprecated responses include standard RFC/IETF headers:
     - `Deprecation: true`
     - `Sunset: <RFC1123 Date>` (12 months from deprecation announcement)
     - `Link: <https://docs.zippyra.com/api/v2/...>; rel="deprecation"`

3. **Version-Compatibility Shims**:
   - Version-compat translation functions live alongside the `/v2/` handler.
   - Compatibility shims accept `/v1/`-shaped payloads, convert them into `/v2/` structures, invoke the core domain logic, and translate responses back to `/v1/` format.

---

## Retroactive Service Extractions (e.g. `admin-store-service`)

During microservice extractions (such as extracting `admin-store-service` out of `store-service`), external endpoint paths are preserved through thin compatibility proxy handlers:
- Legacy `/v1/store/admin/*` paths in `store-service` proxy requests to `admin-store-service`.
- Injected with `versioning.Deprecated(sunsetDate, migrationURL)` response headers to signal callers to migrate to `http://admin-store-service/v1/admin/stores/*`.

---

## CI / Pull-Request Governance

Pull-Requests altering Go structs in `*handlers.go` or `models.go` on existing `/v1/` routes are checked in CI (`.github/workflows/ci.yml`). If a struct is modified without either:
1. Adding a new `/v2/` route, OR
2. Including an explicit `// backward-compatible` code comment,

The CI workflow logs a review warning to alert team members to verify contract stability.
