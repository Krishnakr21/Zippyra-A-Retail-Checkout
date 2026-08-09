# Zippyra Unattended Self-Checkout Kiosk Terminal (`mobile/kiosk_app`)

## Overview

`mobile/kiosk_app` is the platform's third mobile target, built as a Flutter Web application wrapped inside an Electron shell (`mobile/kiosk_app/electron`). It enables unattended self-checkout terminals at retail stores without requiring staff assistance.

---

## Architectural & Scope Decisions

### 1. Customer Authentication Strategy
- **Choice**: **Phone-OTP Login at Terminal (Touch Keypad)**.
- **Rationale**: Pure guest checkout without identity forfeits loyalty points, digital receipt history, and customer engagement. The kiosk presents a large touch-screen keypad for rapid 10-digit phone number entry and SMS OTP verification.

### 2. Exit Pass & Receipt Delivery Strategy
- **Choice**: **SMS / Email Exit Pass Delivery (via `notification-service`)**.
- **Rationale**: Integrating native OS thermal receipt printers introduces fragile driver dependencies per hardware model. Sending the exit pass and digital receipt via SMS/email leverages the existing `notification-service`, allowing customers to present the QR code on their own phone at the RFID exit gate.

### 3. Fixed Store Binding & Device Pairing
- **Choice**: **Pre-Configured Fixed `store_id` & `device_id`**.
- **Rationale**: Unlike personal Customer Apps that dynamic-bind via geofence or entrance QR scan, a kiosk terminal is physically anchored to a single store. Configuration is loaded from `mobile/kiosk_app/electron/kiosk_config.json`.

---

## Build & Launch Commands

### Build Flutter Web Target
```bash
cd mobile/kiosk_app
flutter build web --release
```

### Launch Electron Kiosk Shell
```bash
cd mobile/kiosk_app/electron
npm install
npm start
```
