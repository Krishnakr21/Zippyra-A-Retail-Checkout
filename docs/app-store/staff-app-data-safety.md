# Staff App — Data Safety & App Privacy Declaration

> **Document Owner**: Zippyra Engineering & Legal
> **Last Verified Against Code**: 2026-08-02
> **App**: Zippyra Staff App (com.zippyra.staff)
> **Platforms**: Android (Google Play) + iOS (App Store)

---

## Verification Checklist

Each data-flow claim below was cross-checked against the actual SDK integrations in the codebase.

| SDK / Integration | Package | Verified Data Flow | Evidence |
|---|---|---|---|
| **Sentry Flutter** | `sentry_flutter: ^8.0.0` | Crash reporting. `sendDefaultPii` is NOT enabled (default=false). No `Sentry.setUser()` calls. Crash reports contain stack traces and device info only — no staff PII. | staff_app `main.dart` |
| **Mobile Scanner** | `mobile_scanner: ^5.0.0` | Camera for barcode scanning (inventory counts, QC inspections, exit gate verification). No images stored or transmitted. On-device frame processing only. | Barcode scanning feature |
| **MQTT Client** | Custom `MqttService` via AWS IoT Core | Subscribes to device alert topics per `store_id`. Receives IoT device telemetry (device heartbeats, sensor alerts). **This is NOT personal data about the staff member** — it's device operational data (battery level, connectivity status, temperature). Staff `device_id` used for MQTT connection authentication only. | [mqtt_service.dart](file:///Users/krishna/Downloads/Fatima/Zippyra/mobile/staff_app/lib/core/services/mqtt_service.dart#L35-L55) |
| **Flutter Secure Storage** | via `zippyra_core` (`flutter_secure_storage: ^9.0.0`) | Staff JWT tokens, PIN hash (never plaintext), device pairing credentials, MQTT endpoint stored in platform Keychain/Keystore. | [device_pairing_repository_impl.dart](file:///Users/krishna/Downloads/Fatima/Zippyra/mobile/staff_app/lib/features/device_pairing/data/repositories/device_pairing_repository_impl.dart) |
| **Drift / SQLite** | via `zippyra_core` (`drift: ^2.14.0`) | Offline inventory cache, shift schedule cache. Stored on-device only. | zippyra_core dependency |
| **Flutter Localizations** | `flutter_localizations` (Flutter SDK) | Locale detection for Hindi/English UI. No data collected or transmitted. | Standard Flutter i18n |

**SDKs NOT present in Staff App** (unlike Customer App):
- ❌ No `geolocator` — no location collection
- ❌ No `razorpay_flutter` — no payment processing
- ❌ No `google_sign_in` — staff authenticate via phone + PIN
- ❌ No `share_plus` — no social sharing
- ❌ No `qr_flutter` — QR codes generated server-side for exit passes

---

## Google Play Data Safety Form

### Data Collection Summary

| Data Category | Sub-type | Collected? | Shared? | Purpose | Optional? |
|---|---|---|---|---|---|
| **Location** | Precise location | **No** | — | — | — |
| **Location** | Approximate location | **No** | — | — | — |
| **Personal info** | Name | **Yes** | No | Account management (staff profile in retailer-auth-service) | No |
| **Personal info** | Phone number | **Yes** | No | Account management (PIN-based authentication via retailer-auth-service) | No |
| **Personal info** | Email address | **No** | — | — | — |
| **Financial info** | Payment info | **No** | — | — | — |
| **Financial info** | Purchase history | **No** | — | — | — |
| **App activity** | Other user-generated content | **Yes** (inventory counts, QC inspection results) | No | App functionality | No |
| **App info and performance** | Crash logs | **Yes** | **Yes** (shared with Sentry) | Analytics | No |
| **App info and performance** | Diagnostics | **Yes** | **Yes** (shared with Sentry) | Analytics | No |
| **Device or other IDs** | Device ID | **Yes** | No (device ID used for MQTT pairing and shift binding — not shared with third parties) | App functionality | No |

### Key Declarations

**Location — "No"**: The Staff App does NOT include the `geolocator` package (confirmed absent from `staff_app/pubspec.yaml`). Staff location is not collected, tracked, or transmitted. Store assignment is managed via the backend — the staff member selects their store during shift clock-in, not via GPS.

**Financial Info — "No"**: The Staff App has no payment processing functionality. The `razorpay_flutter` package is not included. Staff process customer returns via the backend order-service API, but no payment instrument data flows through the Staff App.

**MQTT Device Telemetry ≠ Staff Personal Data**: The `MqttService` subscribes to IoT device alert topics (e.g., `store/{store_id}/devices/alerts`). The data received is device operational telemetry — battery levels, sensor readings, connectivity status, firmware versions — NOT personal data about the staff member using the app. This data should NOT be declared as "personal data collected" in the Data Safety form, as it describes equipment, not people.

**Sentry — No PII**: Same as Customer App — `sendDefaultPii=false`, no `setUser()`, crash reports contain device/OS info and stack traces only.

### Data Handling Practices

| Practice | Declaration |
|---|---|
| **Data encrypted in transit** | Yes — TLS 1.2+ with certificate pinning via Dio HTTP client |
| **Data can be deleted by user** | Yes — Staff account deletion available through admin-auth-service. Staff data subject to same DPDP data-deletion pipeline as customer data. |
| **Independent security review** | Planned (VAPT engagement, `docs/security/pentest-scope.md`) |

---

## Apple App Store Privacy "Nutrition Label"

### Data Linked to You

| Apple Category | Data Type | Linked to Identity? | Used for Tracking? | Purpose |
|---|---|---|---|---|
| **Contact Info** | Phone Number | Yes | No | App Functionality (PIN authentication) |
| **Contact Info** | Name | Yes | No | App Functionality (staff profile) |
| **Identifiers** | Device ID | Yes | No | App Functionality (device pairing, MQTT subscription) |
| **Diagnostics** | Crash Data | No | No | Analytics (Sentry — no PII attached) |
| **Diagnostics** | Performance Data | No | No | Analytics (Sentry) |

### Data NOT Collected

| Apple Category | Reason |
|---|---|
| **Location** (Precise or Approximate) | `geolocator` package not included. No location permissions requested. |
| **Financial Info** | No payment SDK. Staff process returns via backend API only. |
| **Health & Fitness** | Not applicable |
| **Sensitive Info** | Not applicable |
| **Contacts** | Not applicable |
| **Email Address** | Staff authenticate via phone + PIN, not email |
| **Photos or Videos** | Camera used for real-time barcode scanning only; no media stored or transmitted |
| **Audio** | Not applicable |
| **Browsing History** | Not applicable |
| **Search History** | Not applicable |
| **Purchases** | Staff don't make purchases; they process customer orders |

### Tracking Declaration

**Does this app track users?** → **No**

The Staff App does NOT:
- Collect advertising identifiers (IDFA)
- Include any advertising SDKs
- Share data with data brokers
- Link staff data with third-party data for advertising

**No App Tracking Transparency (ATT) prompt is required.**

---

## Third-Party SDK Data Collection Summary

| SDK | What IT Collects | Transmitted To | Zippyra's Control |
|---|---|---|---|
| **Sentry Flutter** | Device model, OS version, app version, stack traces, breadcrumbs | Sentry's cloud | PII stripping enforced; `sendDefaultPii=false` |
| **AWS IoT Core (via MQTT)** | MQTT connection metadata (client ID, connection timestamp) | AWS IoT Core (ap-south-1) | Staff device credentials stored in Secure Storage; connection scoped to specific store's device topics |

---

## Comparison: Customer App vs Staff App Data Collection

| Data Category | Customer App | Staff App |
|---|---|---|
| Precise Location | ✅ Yes (foreground, geofence) | ❌ No |
| Phone Number | ✅ Yes (OTP auth) | ✅ Yes (PIN auth) |
| Email | ✅ Optional (Google Sign-In) | ❌ No |
| Name | ✅ Yes | ✅ Yes |
| Payment Info | ❌ No (Razorpay SDK handles) | ❌ No |
| Purchase History | ✅ Yes | ❌ No |
| Device ID / FCM | ✅ Yes | ✅ Yes (device pairing) |
| Crash Logs | ✅ Yes (Sentry) | ✅ Yes (Sentry) |
| IoT Telemetry | ❌ No | ✅ Yes (MQTT — NOT personal data) |
| Tracking | ❌ No | ❌ No |
