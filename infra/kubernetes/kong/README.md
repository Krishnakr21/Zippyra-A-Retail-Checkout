# Kong API Gateway Deployment & Migration Guide

## Overview

Kong API Gateway operates as the single, centralized external-facing gateway for the Zippyra platform deployed on AWS EKS `ap-south-1` (Mumbai).

- **Proxy Endpoint (`api.zippyra.com`)**: AWS Network LoadBalancer routing external traffic into Kong.
- **Admin API**: `ClusterIP` on port 8001. Internal-only — **NEVER exposed via Ingress** or outside the cluster boundary.
- **Service Mesh / Internal Traffic**: Internal microservice-to-microservice calls (e.g. cached calls, internal `SYSTEM` JWT calls) continue using direct Kubernetes DNS (`http://service-name.namespace:8080`). Kong sits exclusively at the cluster edge.

---

## 🚫 EXPLICIT GOVERNANCE RULE

> [!CAUTION]
> **NO MANUAL ADMIN API POKING OR KUBECTL EXEC!**
> All Kong configuration changes (services, routes, plugins, consumers) MUST be version-controlled under `infra/kubernetes/kong/declarative-config/`.
>
> Changes are applied automatically via CI/CD using `deck diff` and `deck sync`. Never execute imperative `curl` calls against port 8001 or manual edits inside Kong pods.

---

## Adding a New Route for a Service

1. Open `infra/kubernetes/kong/declarative-config/services.yaml` and add the Kubernetes `ClusterIP` service URL:
   ```yaml
   - name: new-service
     url: http://new-service.zippyra-pilot.svc.cluster.local:80
   ```

2. Open `infra/kubernetes/kong/declarative-config/routes.yaml` and add the explicit public route mapping (DO NOT map internal paths `/v1/*/internal/*`):
   ```yaml
   - name: new-service-public
     service: new-service
     paths: ["/v1/new-service/public-endpoint"]
     strip_path: false
   ```

3. Run `deck validate` locally or open a Pull Request. CI/CD will validate the schema, post a `deck diff` preview, and execute `deck sync` on merge.

---

## 🚀 Migration Rollout Sequence (Zero-Downtime DNS Cutover)

1. **Deploy Kong Parallel Stack**:
   - Deploy Kong Helm chart with `values.yaml` and load declarative config via `deck sync`.
   - The existing direct ALB Ingress rules remain ACTIVE in parallel during deployment.

2. **Staging Environment Validation**:
   - Point staging DNS (`api.staging.zippyra.com`) to Kong's LoadBalancer IP/hostname.
   - Execute synthetic golden path verification:
     ```bash
     TARGET_URL="https://api.staging.zippyra.com" ./scripts/golden-path-check.sh
     ```
   - Execute internal route blocking verification:
     ```bash
     ./scripts/verify-internal-routes-blocked.sh
     ```
   - Confirm complete trace threading in Jaeger via edge-generated `X-Request-ID`.

3. **Production DNS Cutover**:
   - Repoint production CNAME/A-record (`api.zippyra.com`) from direct ALB to Kong LoadBalancer.
   - Retain direct ALB Ingress manifests in `infra/kubernetes/ingress.yaml` as fallback for 7 days.

4. **Teardown Unused Ingress**:
   - After 7 days of clean production operation through Kong, decommission legacy direct ALB Ingress rules.
