-- Seed Chains
INSERT INTO chains (id, name, created_at, updated_at)
VALUES 
  ('11111111-1111-1111-1111-111111111111', 'Zippyra Mega Retail', NOW(), NOW()),
  ('22222222-2222-2222-2222-222222222222', 'Apex Supermarkets', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- Seed Stores
INSERT INTO stores (id, chain_id, name, store_code, address, status, created_at, updated_at)
VALUES
  ('33333333-3333-3333-3333-333333333333', '11111111-1111-1111-1111-111111111111', 'Downtown Flagship Superstore', 'ZMR-001', '100 Main St, Bangalore, India', 'ACTIVE', NOW(), NOW()),
  ('44444444-4444-4444-4444-444444444444', '11111111-1111-1111-1111-111111111111', 'Koramangala Express Store', 'ZMR-002', '45 80-Feet Rd, Bangalore, India', 'ACTIVE', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- Seed Categories
INSERT INTO categories (id, name, created_at, updated_at)
VALUES
  ('55555555-5555-5555-5555-555555555555', 'Beverages & Soft Drinks', NOW(), NOW()),
  ('66666666-6666-6666-6666-666666666666', 'Snacks & Confectionery', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- HSN Code
INSERT INTO hsn_gst_rates (hsn_code, cgst_percent, sgst_percent, igst_percent, description)
VALUES ('2201', 9.0, 9.0, 18.0, 'Mineral water and carbonated drinks')
ON CONFLICT (hsn_code) DO NOTHING;

-- Seed Products
INSERT INTO products (id, store_id, chain_id, barcode, name, description, category_id, price_paise, mrp_paise, hsn_code, created_at, updated_at)
VALUES
  ('77777777-7777-7777-7777-777777777777', '33333333-3333-3333-3333-333333333333', '11111111-1111-1111-1111-111111111111', '8901234567890', 'Sparkling Mineral Water 500ml', 'Pure alpine sparkling water', 4500, 5000, '2201', NOW(), NOW()),
  ('88888888-8888-8888-8888-888888888888', '33333333-3333-3333-3333-333333333333', '11111111-1111-1111-1111-111111111111', '8901234567891', 'Organic Cold Brew Coffee 250ml', 'Single-origin Arabica dark roast', 12000, 15000, '2201', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- Stock Levels
INSERT INTO stock_levels (store_id, barcode, on_hand_qty, reorder_point, reorder_qty, updated_at)
VALUES
  ('33333333-3333-3333-3333-333333333333', '8901234567890', 450, 50, 100, NOW()),
  ('33333333-3333-3333-3333-333333333333', '8901234567891', 18, 25, 50, NOW())
ON CONFLICT (store_id, barcode) DO NOTHING;

-- Devices
INSERT INTO devices (id, store_id, chain_id, device_type, gate_id, label, status, iot_thing_name, cert_arn, cert_id, device_jwt_kid, last_heartbeat_at, firmware_version, created_at, updated_at)
VALUES
  ('99999999-9999-9999-9999-999999999999', '33333333-3333-3333-3333-333333333333', '11111111-1111-1111-1111-111111111111', 'EXIT_GATE_SCANNER', 'GATE-A', 'Main Entrance Scanner 1', 'ONLINE', 'thing-gate-a', 'arn:aws:iot:ap-south-1:123456789012:cert/1', 'cert-1', 'kid-1', NOW(), 'v2.4.1', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- Users
INSERT INTO users (id, phone, email, name, auth_provider_last)
VALUES
  ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', '+919876543210', 'customer1@example.com', 'Aarav Sharma', 'phone'),
  ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', '+919876543211', 'retailer1@zippyra.com', 'Priya Patel', 'email')
ON CONFLICT (id) DO NOTHING;
