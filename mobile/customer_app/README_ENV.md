# Customer App Environment Configuration Guide

## Overview
The Zippyra Customer App resolves environment configuration based on build mode (`debug` vs `release`) and target environment (`development`, `staging`, `production`).

---

## Variable Classification

| Variable Name | Safe Dotenv Only (Dev/Debug) | `--dart-define-from-file` Required (Release Builds) | Description |
|---|---|---|---|
| `APP_ENV` | Yes | Yes | Target environment (`development`, `staging`, `production`) |
| `APP_VERSION_NAME` | Yes | Yes | Mobile app version display string |
| `APP_MIN_SUPPORTED_VERSION` | Yes | No | Minimum supported client version (checked by backend) |
| `API_BASE_URL_DEV` | Yes | Yes | Local / LAN development gateway URL (`http://10.0.2.2:8080`) |
| `API_BASE_URL_STAGING` | Yes | Yes | Staging gateway URL |
| `API_BASE_URL_PROD` | Yes | Yes | Production gateway URL |
| `WS_BASE_URL_PROD` | Yes | Yes | Production WebSocket endpoint |
| `CERT_PIN_SHA256_PRIMARY` | No | **YES (REQUIRED IN PROD)** | Primary SSL Certificate SHA-256 Pin |
| `CERT_PIN_SHA256_BACKUP` | No | **YES (REQUIRED IN PROD)** | Backup SSL Certificate SHA-256 Pin |
| `GOOGLE_OAUTH_SERVER_CLIENT_ID` | Yes | **YES (Release)** | Google OAuth Server Client ID |
| `GOOGLE_OAUTH_IOS_CLIENT_ID` | Yes | Yes | Google OAuth iOS Client ID |
| `GOOGLE_OAUTH_ANDROID_CLIENT_ID` | Yes | Yes | Google OAuth Android Client ID |
| `RAZORPAY_KEY_ID` | Yes | **YES (Release)** | Razorpay API Key ID (Public Key) |
| `GOOGLE_MAPS_API_KEY_ANDROID` | No | **YES (Release)** | Android Google Maps SDK Key |
| `GOOGLE_MAPS_API_KEY_IOS` | No | **YES (Release)** | iOS Google Maps SDK Key |
| `FCM_VAPID_KEY` | Yes | No | Firebase Cloud Messaging Web VAPID key |
| `PLAY_INTEGRITY_CLOUD_PROJECT_NUMBER` | No | **YES (Release)** | Google Play Integrity API Cloud Project Number |
| `ROOT_JAILBREAK_CHECK_ENABLED` | **Yes** | No | Toggle root/jailbreak detection |
| `SENTRY_DSN_MOBILE` | No | **YES (Release)** | Sentry error tracking DSN |
| `SENTRY_TRACES_SAMPLE_RATE` | **Yes** | No | Performance tracing sample rate |
| `DEFAULT_LOCALE` | **Yes** | No | Default language code (`en`, `hi`) |
| `OFFLINE_SYNC_JITTER_MAX_SECONDS` | **Yes** | No | Max jitter for offline sync retry backoff |
| `OFFER_CACHE_MAX_AGE_SECONDS` | **Yes** | No | Maximum cache TTL for offline offer banners |

---

## Compiling Release Builds
For staging/production releases, inject secret variables via `--dart-define-from-file`:

```bash
flutter build ipa --release --dart-define-from-file=env/prod.json
```

Example `env/prod.json` (git-ignored):
```json
{
  "APP_ENV": "production",
  "APP_VERSION_NAME": "1.0.0",
  "API_BASE_URL_PROD": "https://api.zippyra.com",
  "WS_BASE_URL_PROD": "wss://api.zippyra.com/ws",
  "CERT_PIN_SHA256_PRIMARY": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
  "CERT_PIN_SHA256_BACKUP": "BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB=",
  "GOOGLE_OAUTH_SERVER_CLIENT_ID": "your_prod_google_client_id.apps.googleusercontent.com",
  "RAZORPAY_KEY_ID": "rzp_live_samplekey1234",
  "GOOGLE_MAPS_API_KEY_ANDROID": "AIzaSySampleAndroidKey12345",
  "GOOGLE_MAPS_API_KEY_IOS": "AIzaSySampleIOSKey12345",
  "PLAY_INTEGRITY_CLOUD_PROJECT_NUMBER": "123456789012",
  "SENTRY_DSN_MOBILE": "https://sample@sentry.io/12345"
}
```

> ⚠️ **Production Safety Assertion**: If `APP_ENV == 'production'` and `CERT_PIN_SHA256_PRIMARY` is empty, `AppConfig.validate()` throws `StateError` during app initialization to prevent insecure release builds.
