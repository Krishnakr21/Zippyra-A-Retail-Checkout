> [!CAUTION]
> **LEGAL NOTICE**: This is an initial operational draft reflecting Zippyra platform technical architecture and DPDP Act 2023 compliance logic. Final legal documents must be reviewed and approved by a qualified legal professional specializing in Indian Law before production deployment.

# Zippyra Privacy Policy

**Effective Date:** August 1, 2026  
**Notice Version:** v1.2 (`DPDP_PRIVACY_NOTICE_VERSION=v1.2`)  
**Data Fiduciary:** Zippyra Retail Technologies Private Limited, India

At Zippyra, we are committed to safeguarding your personal data in full compliance with the **Digital Personal Data Protection Act, 2023 (DPDP Act)** and applicable Indian regulations. This Privacy Policy details the exact personal data we collect, how it is processed, data localization commitments, and your statutory rights.

---

## 1. Personal Data We Collect

We collect only the minimum necessary personal data required to operate our autonomous self-checkout retail platform:

- **Identity & Contact Details** (`auth-service`): Mobile Phone Number, Email Address, Full Name (optional).
- **Point-in-Time Store Location** (`store-service`): Point-in-time device location verified **only** at store entry geofence binding. *We do NOT perform continuous background location tracking.*
- **Purchase & Transaction History** (`order-service`): Cart items scanned, items purchased, invoice totals, discounts, GST tax breakdown (CGST, SGST, IGST), and exit pass tokens.
- **Device & Push Identifiers** (`notification-service`): Firebase Cloud Messaging (FCM) tokens and device OS type for order receipts and security alerts.
- **Payment Method Metadata** (`payment-service`): Payment method type (e.g., UPI, Card, NetBanking) and transaction reference IDs. **We NEVER store full card numbers, CVVs, or UPI PINs** — all payment processing occurs directly on PCI-DSS certified gateway infrastructure (Razorpay/Cashfree).

---

## 2. Specified Purposes of Data Processing

Your data is processed strictly for specified, lawful retail operations:

1. **Self-Checkout & Order Fulfillment**: Item scanning, cart total calculation, digital exit pass generation, and digital invoice generation.
2. **Statutory GST Tax Compliance**: Issuing B2C/B2B tax invoices with Government E-Invoice IRN QR Codes via GST GSP/IRP portals.
3. **Store Security & Theft Prevention**: Validating item scans against exit gate sensors before issuing encrypted exit QR passes.
4. **Customer Support & Returns**: Resolving order queries, processing 24-hour return requests, and issuing refunds.
5. **Optional Marketing & Loyalty (Consent-Driven)**: Sending transactional updates, digital receipts via WhatsApp, and optional promotional offers.

---

## 3. Data Sharing & Infrastructure Partners

We share personal data strictly with trusted service providers bound by data processing agreements:

- **Payment Gateways** (*Razorpay / Cashfree*): For secure payment authorization and instant refund processing.
- **Cloud Infrastructure** (*AWS India - ap-south-1 Region*): All primary databases, Redis caches, Kafka brokers, and TimescaleDB analytics run exclusively on AWS infrastructure within India, fulfilling Indian **Data Localization** mandates.
- **Digital Receipt Delivery** (*WhatsApp Business API / Meta*): Delivering e-invoices and order confirmations via SMS/WhatsApp.
- **Retail Store Operations**: Staff at the specific retail store where you shop receive transaction details for item picking, returns processing, and exit validation. Store staff cannot access data from other retail chains or stores.

---

## 4. Data Retention Periods

In accordance with our DPDP retention schedule (`DPDP_DATA_RETENTION_DAYS`), your data is stored for the following durations:

- **Active Account & Consent Records**: Retained until account closure or consent withdrawal.
- **Completed Order & Tax Invoices**: Retained for **7 years** (2,555 days) to comply with statutory Indian GST & Income Tax Recordkeeping requirements.
- **Analytics & Anonymized Aggregates**: Transformed into anonymized metrics after **90 days** in ClickHouse/TimescaleDB storage.

---

## 5. Your Statutory Rights under DPDP Act 2023

You possess full control over your personal data:

- **Right to Access**: Request a comprehensive export of all personal data held about you across all microservices (delivered via a secure presigned JSON download).
- **Right to Correction & Erasure**: Request correction of inaccurate data or complete account & PII erasure across all service databases.
- **Right to Nominate**: Nominate another individual to exercise your rights in the event of death or incapacity.
- **Right of Grievance Redressal**: Contact our statutory Grievance Officer for prompt resolution of privacy concerns.

### Grievance Officer Contact Information

- **Grievance Officer:** Ms. Ananya Verma
- **Designation:** Chief Data Protection Officer
- **Email:** `grievance-officer@zippyra.com` (`DPDP_GRIEVANCE_OFFICER_EMAIL`)
- **Address:** Zippyra HQ, 4th Floor, Retail Tech Park, Outer Ring Road, Bengaluru, KA - 560103
- **Statutory SLA:** Written acknowledgment within 48 hours; full resolution within 30 days.

---

## 6. Consent Management & Re-confirmation

You can manage your consent preferences anytime under **Profile -> Privacy Settings** in the Zippyra Customer App:
- `MARKETING_COMMS`: Toggle promotional notifications and WhatsApp updates.
- `LOCATION_TRACKING`: Toggle store-entry geofence verification.
- `ANALYTICS_SHARING`: Toggle personalized recommendation analytics.

Whenever this Privacy Policy is updated (indicated by a version bump of `DPDP_PRIVACY_NOTICE_VERSION`), you will receive an in-app prompt requiring re-confirmation before continuing platform usage.

---

## 7. TRAI Commercial Communications & DND Registry Compliance

In accordance with the Telecom Regulatory Authority of India (TRAI) TCCCPR regulations:
- **Transactional Messages** (OTP, order receipts, exit passes) are sent under registered transactional headers (`ZIPPYR`) and bypass DND filtering as required for service delivery.
- **Promotional / Marketing Messages** require explicit opt-in (`MARKETING` preference set to `OPT-IN`). Promotional communications default to **OFF (NONE)** for all new users and are sent via dedicated promotional headers (`PRMZIP`) honoring National Do Not Disturb (NDNC) registry preferences.
