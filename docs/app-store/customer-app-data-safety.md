# Customer App — Data Safety & App Privacy Declaration

> **Document Owner**: Zippyra Engineering & Legal
> **Last Verified Against Code**: 2026-08-02
> **App**: Zippyra Customer App (com.zippyra.customer)
> **Platforms**: Android (Google Play) + iOS (App Store)

---

## Verification Checklist

Each data-flow claim below was cross-checked against the actual SDK integrations in the codebase — not assumed. This checklist serves as the audit trail.

| SDK / Integration | Package | Verified Data Flow | Evidence |
|---|---|---|---|
| **Geolocator** | `geolocator: ^10.1.0` | Precise foreground-only location. `LocationAccuracy.best` used in `store_session_bloc.dart:183`. No `ACCESS_BACKGROUND_LOCATION` permission in AndroidManifest. No `NSLocationAlwaysUsageDescription` in Info.plist. | [store_session_bloc.dart](file:///Users/krishna/Downloads/Fatima/Zippyra/mobile/customer_app/lib/features/store_session/presentation/bloc/store_session_bloc.dart#L182-L191) |
| **Razorpay Flutter SDK** | `razorpay_flutter: ^1.3.6` | SDK opens Razorpay's own checkout sheet. App passes `order_id`, `amount`, and `key_id` to SDK. Card/UPI details are entered directly in Razorpay's UI, NOT transmitted to Zippyra servers. App receives only `payment_id` on success. Per Razorpay's own docs, the SDK collects device info and payment instrument details directly to Razorpay's PCI-DSS compliant servers. | [razorpay_service.dart](file:///Users/krishna/Downloads/Fatima/Zippyra/mobile/customer_app/lib/features/payment/data/razorpay_service.dart#L33-L34) |
| **Sentry Flutter** | `sentry_flutter: ^8.0.0` | Crash reporting with feature-tagging `beforeSend`. `sendDefaultPii` is NOT enabled (default=false). No `Sentry.setUser()` calls found. Crash reports contain stack traces and device info (OS version, device model) but no user PII (name, email, phone). | [main.dart](file:///Users/krishna/Downloads/Fatima/Zippyra/mobile/customer_app/lib/main.dart#L16-L51) |
| **Google Sign-In** | `google_sign_in: ^6.2.1` | Requests Google ID token for authentication. App sends ID token to `auth-service` backend. Google profile info (name, email, profile photo URL) is received by the app and transmitted to the Zippyra backend for account creation. | [auth_repository_impl.dart](file:///Users/krishna/Downloads/Fatima/Zippyra/mobile/customer_app/lib/features/auth/data/repositories/auth_repository_impl.dart#L48-L54) |
| **FCM (via notification-service)** | Backend-managed; app registers `fcm_token` | App collects FCM device token and `device_id`, sends to Zippyra `notification-service` backend. Tokens shared with Firebase/Google's push infrastructure (Google's own form has a specific carve-out for this). | [device_token_registrar.dart](file:///Users/krishna/Downloads/Fatima/Zippyra/mobile/customer_app/lib/features/notifications/data/device_token_registrar.dart#L46-L53) |
| **Mobile Scanner** | `mobile_scanner: ^5.0.0` | Camera access for barcode scanning only. No images or video are stored, transmitted, or collected. Frames are processed on-device for barcode detection. | Barcode scanning feature only |
| **Flutter Secure Storage** | via `zippyra_core` (`flutter_secure_storage: ^9.0.0`) | JWT tokens, refresh tokens stored in platform Keychain/Keystore. No PII stored in plain SharedPreferences. | zippyra_core dependency |
| **Drift / SQLite** | via `zippyra_core` (`drift: ^2.14.0`) | Offline cart cache, order history cache. Stored on-device only, not transmitted externally. | zippyra_core dependency |

---

## Google Play Data Safety Form

### Data Collection Summary

| Data Category | Sub-type | Collected? | Shared? | Purpose | Optional? |
|---|---|---|---|---|---|
| **Location** | Precise location | **Yes** | No | App functionality (store-entry geofence verification at session bind time) | No (required for store session) |
| **Location** | Approximate location | No | — | — | — |
| **Personal info** | Name | **Yes** | **Yes** (shared with: specific retail store for purchase receipt; Razorpay for payment processing) | Account management, App functionality | No |
| **Personal info** | Email address | **Yes** | **Yes** (shared with: Razorpay for payment receipts) | Account management, App functionality | Yes (Google Sign-In optional) |
| **Personal info** | Phone number | **Yes** | No (used only for OTP authentication) | Account management, App functionality | No (required for OTP login) |
| **Financial info** | Payment info (cards, UPI) | **No** | — | — | — |
| **Financial info** | Purchase history | **Yes** | No | App functionality (order history), Analytics (loyalty tier calculation) | No |
| **App activity** | In-app search history | No | — | — | — |
| **App activity** | Other user-generated content | **Yes** (support feedback, NPS ratings) | No | App functionality | Yes |
| **App info and performance** | Crash logs | **Yes** | **Yes** (shared with Sentry for crash reporting) | Analytics, App functionality | No |
| **App info and performance** | Diagnostics | **Yes** | **Yes** (shared with Sentry) | Analytics | No |
| **Device or other IDs** | Device ID, FCM token | **Yes** | **Yes** (FCM token shared with Google/Firebase for push notifications) | App functionality (push notifications) | No |

### Key Declarations

**Financial Info — "No"**: The Razorpay Flutter SDK (`razorpay_flutter: ^1.3.6`) handles all payment instrument collection (card numbers, UPI VPA, etc.) within Razorpay's own PCI-DSS compliant UI. The Zippyra app passes only `order_id`, `amount`, and Razorpay `key_id` to the SDK. Payment instrument details are transmitted directly from the Razorpay SDK to Razorpay's servers — they never pass through Zippyra's own backend. The app receives only a `payment_id` string on successful payment.

> ⚠️ **Verification note**: This claim is accurate based on `razorpay_flutter`'s standard checkout integration pattern used in [razorpay_service.dart](file:///Users/krishna/Downloads/Fatima/Zippyra/mobile/customer_app/lib/features/payment/data/razorpay_service.dart). The SDK's `open()` method launches Razorpay's hosted checkout. However, per Razorpay's own Data Safety guidance, the *Razorpay SDK itself* collects device identifiers and payment details on behalf of the merchant — Google may require this to be declared as "collected by a third-party SDK." Razorpay provides specific Google Data Safety guidance at their developer docs. Cross-reference that guidance before final submission.

**Location — Precise, Foreground Only**: The app uses `LocationAccuracy.best` (precise) in `store_session_bloc.dart:183` for store geofence verification at session bind time only. There is:
- **No** `ACCESS_BACKGROUND_LOCATION` permission in AndroidManifest.xml
- **No** `NSLocationAlwaysUsageDescription` in Info.plist
- **No** continuous or background location tracking

Location is collected once at store-entry bind, used to verify the customer is within geofence range of the store, and not persisted beyond the session.

**Sentry Crash Reporting — No PII**: `sendDefaultPii` is NOT enabled (confirmed absent from [main.dart](file:///Users/krishna/Downloads/Fatima/Zippyra/mobile/customer_app/lib/main.dart#L16-L51)). No `Sentry.setUser()` calls exist. Crash reports contain device model, OS version, and stack traces — but no user name, email, phone, or account identifiers.

### Data Handling Practices

| Practice | Declaration |
|---|---|
| **Data encrypted in transit** | Yes — TLS 1.2+ with certificate pinning (implemented via Dio HTTP client with cert-pinning interceptor in `zippyra_core`) |
| **Data can be deleted by user** | Yes — DPDP Act "Delete My Account" flow under Profile → Privacy Settings triggers full data deletion via `compliance-service` |
| **Independent security review** | Planned (VAPT engagement scoped in `docs/security/pentest-scope.md`) |

---

## Apple App Store Privacy "Nutrition Label"

### Data Linked to You

| Apple Category | Data Type | Linked to Identity? | Used for Tracking? | Purpose |
|---|---|---|---|---|
| **Contact Info** | Phone Number | Yes | No | App Functionality (OTP authentication) |
| **Contact Info** | Name | Yes | No | App Functionality (account profile) |
| **Contact Info** | Email Address | Yes | No | App Functionality (Google Sign-In, optional) |
| **Location** | Precise Location | Yes | No | App Functionality (store geofence verification — foreground only) |
| **Purchases** | Purchase History | Yes | No | App Functionality (order history, digital receipts) |
| **Identifiers** | Device ID | Yes | No | App Functionality (push notifications via FCM) |
| **Diagnostics** | Crash Data | No | No | Analytics (Sentry crash reporting — no PII attached) |
| **Diagnostics** | Performance Data | No | No | Analytics (Sentry performance monitoring) |

### Data NOT Collected

| Apple Category | Reason |
|---|---|
| **Financial Info** (Payment Info, Credit Info) | Razorpay SDK handles all payment instruments directly; Zippyra app does not collect or transmit card/UPI details |
| **Health & Fitness** | Not applicable |
| **Sensitive Info** | Not applicable |
| **Browsing History** | Not applicable |
| **Search History** | No in-app search history is collected or transmitted |
| **Contacts** | Not applicable |
| **Photos or Videos** | Camera used only for real-time barcode scanning; no images stored or transmitted |
| **Audio** | Not applicable |

### Tracking Declaration

**Does this app track users?** → **No**

The app does NOT use data to track users across other companies' apps or websites for advertising purposes. This matches Apple's specific definition of "tracking" (linking user/device data with third-party data for targeted advertising or sharing with data brokers). Consequently:
- **No App Tracking Transparency (ATT) prompt is required.**
- No advertising identifiers (IDFA) are collected.
- No third-party advertising SDKs are integrated.

---

## Third-Party SDK Data Collection Summary

Per Google Play and Apple's requirements, here is what each third-party SDK in the app collects independently of the app's own data collection:

| SDK | What IT Collects (per its own privacy docs) | Transmitted To | Zippyra's Control |
|---|---|---|---|
| **Razorpay Flutter** | Payment instrument details (card/UPI), device fingerprint, IP address | Razorpay's PCI-DSS servers | Zippyra does not access this data; Razorpay acts as independent data controller for payment instrument data |
| **Sentry Flutter** | Device model, OS version, app version, stack traces, breadcrumbs | Sentry's cloud (EU/US) | PII stripping enforced via `beforeSend`; `sendDefaultPii=false` |
| **Google Sign-In** | Google account profile (name, email, photo) | Google's servers → then Zippyra's auth-service | User-initiated; profile data used for account creation only |
| **Geolocator** | Device GPS coordinates | Zippyra's store-service (for geofence check) | Collected once at session bind; not persisted |
