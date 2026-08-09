# Google Sign-In Platform Setup Guide

## Overview
This feature integrates Google OAuth ("Continue with Google") in the Flutter customer application, authenticating users via Google ID token verified server-side by `auth-service`.

---

## 1. Google Cloud Console Setup

1. Create/Open project in [Google Cloud Console](https://console.cloud.google.com/).
2. Configure **OAuth Consent Screen**:
   - User Type: External
   - Add scopes: `openid`, `email`, `profile`.
3. Create **OAuth 2.0 Client IDs**:
   - **Web Application Client ID**: This is the **Server Client ID**.
     - Set environment variable `GOOGLE_OAUTH_CLIENT_ID` on backend `auth-service`.
     - Pass to Flutter via `--dart-define=GOOGLE_OAUTH_CLIENT_ID=your_web_client_id`.

---

## 2. Android Configuration

1. Get your Android Debug / Release SHA-1 fingerprints:
   ```bash
   keytool -list -v -keystore ~/.android/debug.keystore -alias androiddebugkey -storepass android -keypass android
   ```
2. In Google Cloud Console, add an **Android Client ID**:
   - Package name: `com.krishnakumar.zippyra.customer` (or `com.example.customerApp`)
   - SHA-1 Certificate fingerprint: Paste output from `keytool`.
3. Download `google-services.json` and place it at:
   `mobile/customer_app/android/app/google-services.json`

---

## 3. iOS Configuration

1. In Google Cloud Console, create an **iOS Client ID**:
   - Bundle ID: `com.krishnakumar.zippyra.customer`
2. Download `GoogleService-Info.plist` and add it to `ios/Runner/GoogleService-Info.plist` in Xcode.
3. Add the `REVERSED_CLIENT_ID` URL scheme to `ios/Runner/Info.plist`:

```xml
<key>CFBundleURLTypes</key>
<array>
	<dict>
		<key>CFBundleTypeRole</key>
		<string>Editor</string>
		<key>CFBundleURLSchemes</key>
		<array>
			<!-- Replace with your REVERSED_CLIENT_ID from GoogleService-Info.plist -->
			<string>com.googleusercontent.apps.123456789012-abcdefghijklmnopqrstuvwxyz</string>
		</array>
	</dict>
</array>
```

---

## 4. Running the Flutter App

Run Flutter with `GOOGLE_OAUTH_CLIENT_ID` provided:

```bash
flutter run --dart-define=GOOGLE_OAUTH_CLIENT_ID=your_server_web_client_id.apps.googleusercontent.com
```
