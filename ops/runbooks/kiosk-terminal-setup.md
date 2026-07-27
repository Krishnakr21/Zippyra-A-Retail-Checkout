# Kiosk Terminal Setup & Device Pairing Runbook

## Overview
This runbook covers hardware provisioning, device-pairing, electron shell execution, crash recovery, and idle timeout behavior for Zippyra unattended self-checkout kiosks.

---

## 1. Initial Device Pairing (`device-mgmt-service`)
1. Turn on the kiosk terminal hardware.
2. Generate a 6-digit pairing code on the Admin Platform or Chain HQ Dashboard.
3. Update `mobile/kiosk_app/electron/kiosk_config.json`:
   ```json
   {
     "store_id": "STORE-BLR-001",
     "device_id": "KIOSK-TERM-BLR-01",
     "pairing_code": "849204",
     "api_base_url": "https://api.zippyra.com",
     "idle_timeout_seconds": 120,
     "kiosk_mode": true
   }
   ```

---

## 2. Electron Kiosk Shell & Auto-Recovery
- Electron runs in `kiosk: true`, `fullscreen: true` mode without OS window frames or navigation bars.
- **Crash Recovery**: If the Chromium renderer crashes, `main.js` automatically reloads the renderer within 1 second without physical technician intervention.
- **Idle Timeout**: Inactivity for >120 seconds automatically clears cart state and returns the terminal to the Welcome screen.
