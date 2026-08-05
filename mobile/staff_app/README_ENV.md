# Staff App Environment Configuration Guide

## Overview
The Zippyra Staff App resolves environment configuration based on build mode (`debug` vs `release`) and target environment (`development`, `staging`, `production`).

---

## Variable Classification

| Variable Name | Safe Dotenv Only (Dev/Debug) | `--dart-define-from-file` Required (Release Builds) | Description |
|---|---|---|---|
| `APP_ENV` | Yes | Yes | Target environment (`development`, `staging`, `production`) |
| `APP_VERSION_NAME` | Yes | Yes | Mobile app version display string |
| `API_BASE_URL_DEV` | Yes | Yes | Local / LAN development gateway URL |
| `API_BASE_URL_STAGING` | Yes | Yes | Staging gateway URL |
| `API_BASE_URL_PROD` | Yes | Yes | Production gateway URL |
| `CERT_PIN_SHA256_PRIMARY` | No | **YES (REQUIRED IN PROD)** | Primary SSL Certificate SHA-256 Pin |
| `CERT_PIN_SHA256_BACKUP` | No | **YES (REQUIRED IN PROD)** | Backup SSL Certificate SHA-256 Pin |
| `LOCAL_CATALOG_SYNC_ON_SHIFT_START` | **Yes** | No | Auto sync store catalog at shift start |
| `OFFLINE_QUEUE_RETRY_INTERVAL_SECONDS` | **Yes** | No | Retry interval for offline transaction queue |
| `OFFLINE_QUEUE_MAX_RETRY_ATTEMPTS` | **Yes** | No | Maximum retries before flagging item for manual reconciliation |
| `MQTT_BROKER_URL` | No | **YES (Release)** | AWS IoT Core MQTT Endpoint (wss://) |
| `MQTT_CLIENT_CERT_PATH` | **Yes** | No | Local asset path to device TLS certificate |
| `MQTT_CLIENT_KEY_PATH` | **Yes** | No | Local asset path to device TLS private key |
| `MQTT_TOPIC_PREFIX` | **Yes** | No | MQTT topic namespace |
| `STAFF_SESSION_IDLE_TIMEOUT_MINUTES` | **Yes** | No | Auto-logout idle timeout on shared POS devices |
| `SENTRY_DSN_MOBILE` | No | **YES (Release)** | Sentry error tracking DSN |
| `SENTRY_TRACES_SAMPLE_RATE` | **Yes** | No | Sentry performance trace sample rate |
| `DEFAULT_LOCALE` | **Yes** | No | Default language code (e.g. `en`) |

---

## Compiling Release Builds
For staging/production releases, inject variables via `--dart-define-from-file`:

```bash
flutter build ipa --release --dart-define-from-file=env/prod.json
```

Example `env/prod.json` (git-ignored):
```json
{
  "APP_ENV": "production",
  "APP_VERSION_NAME": "1.0.0",
  "API_BASE_URL_PROD": "https://api.zippyra.com",
  "CERT_PIN_SHA256_PRIMARY": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
  "CERT_PIN_SHA256_BACKUP": "BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB=",
  "MQTT_BROKER_URL": "wss://a1b2c3d4e5f6-ats.iot.ap-south-1.amazonaws.com/mqtt",
  "SENTRY_DSN_MOBILE": "https://sample@sentry.io/12345"
}
```

> ⚠️ **Production Safety Assertion**: If `APP_ENV == 'production'` and `CERT_PIN_SHA256_PRIMARY` is empty, `AppConfig.validate()` throws `StateError` during app initialization to prevent insecure release builds.
