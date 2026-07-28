> [!CAUTION]
> **LEGAL NOTICE**: This is an initial operational draft reflecting Zippyra platform technical architecture and DPDP Act 2023 compliance logic. Final legal documents must be reviewed and approved by a qualified legal professional specializing in Indian Law before production deployment.

# Zippyra Return & Refund Policy

**Effective Date:** August 1, 2026  
**Version:** v1.2

At Zippyra, we strive to ensure a seamless self-checkout shopping experience. This Return & Refund Policy outlines the exact rules, eligible categories, and refund timelines for purchases made through the Zippyra platform.

---

## 1. Return Window & Eligibility

- **24-Hour Return Window**: Return requests must be initiated via the Zippyra Customer App within **24 hours** from the order completion timestamp (`order-service`).
- **Physical Condition**: Returned items must be unused, in their original packaging, with intact tags, barcodes, and seals.

---

## 2. Non-Returnable Categories (`is_returnable = false`)

In accordance with retail safety and hygiene standards, items flagged as non-returnable (`is_returnable = false` in inventory database) cannot be returned or refunded once purchased:

- Perishable food items, dairy, fresh produce, and frozen goods.
- Personal care, cosmetics, and hygiene products.
- Innerwear, swimwear, and baby care essentials.
- Items marked under clearance or final sale promotions.

---

## 3. Return & Refund Workflow

When you tap **"Request Return"** under Order Details in the app, the following automated workflow is triggered:

1. **Request Submission (`RETURN_REQUESTED`)**: You select the item(s), quantity, and reason for return.
2. **Staff Inspection**: Bring the item(s) to the designated Customer Service Desk at the store where the purchase was made. Store staff verify item condition against your digital return request.
3. **Staff Action (`RETURNED` or `RETURN_REJECTED`)**:
   - If accepted: Staff approve the return in the Retailer Staff App, triggering instant refund processing.
   - If rejected: Staff specify the reason (e.g., damaged packaging, missing barcode), and the request status updates to `RETURN_REJECTED`.

---

## 4. Refund Processing & Timelines

- **Refund Method**: Refunds are credited back exclusively to the original payment source (UPI account, Credit/Debit card) used during checkout.
- **Refund Initiation**: Payment refund calls (`payment-service`) are dispatched immediately upon staff approval.
- **Banking Settlement SLA**: While Zippyra dispatches refunds instantly, banking settlement SLAs depend on gateway processing (*Razorpay / Cashfree*):
  - **UPI Transfers**: Usually credited within 2 to 24 hours.
  - **Credit / Debit Cards & NetBanking**: 5 to 7 business days as per standard Indian banking network processing times.
