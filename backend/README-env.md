# Backend Environment Variables Reference

| Service | Environment Variable | Default Value | Description |
|---|---|---|---|
| All | `PORT` | `8080` | HTTP Server Listening Port |
| All | `DATABASE_URL` | `postgres://zippyra_user:zippyra_password@localhost:5432/zippyra` | PostgreSQL Connection String |
| All | `REDIS_URL` | `localhost:6379` | Redis Host & Port |
| All | `KAFKA_BROKERS` | `localhost:9092` | Kafka Bootstrap Brokers |
| Auth Service | `JWT_SECRET` | `zippyra-dev-jwt-secret-key-32bytes` | JWT Secret Key for Ed25519 / HMAC signing |
| Auth Service | `GMAIL_SMTP_USER` | *(none)* | Gmail address for sending Email OTP (STARTTLS 587) |
| Auth Service | `GMAIL_SMTP_APP_PASSWORD` | *(none)* | 16-character Google App Password (requires 2FA) |
| Auth Service | `GOOGLE_OAUTH_CLIENT_ID` | *(none)* | Google OAuth Web Client ID for ID Token verification |
| Auth Service | `SMS_PROVIDER` | `log` | SMS Provider: `log` (non-prod log only) or `twilio` |
| Auth Service | `TWILIO_ACCOUNT_SID` | *(none)* | Twilio Account SID for SMS delivery |
| Auth Service | `TWILIO_AUTH_TOKEN` | *(none)* | Twilio Auth Token for SMS delivery |
| Auth Service | `TWILIO_FROM_PHONE` | *(none)* | Twilio From Phone Number (E.164 format) |
| Admin Auth Service | `ADMIN_TOTP_ENCRYPTION_KEY` | `emlwcHlyYS1kZXYtdG90cC1lbmNyeXB0aW9uLWtleTMy` | Base64-encoded 32-byte AES key for encrypting TOTP secrets at rest |
| Device Mgmt Service | `TIMESCALE_URL` | `postgres://zippyra_user:zippyra_password@localhost:5432/zippyra` | TimescaleDB Connection String for hypertable heartbeats |
