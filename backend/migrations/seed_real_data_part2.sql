-- 1. Stores
INSERT INTO stores (id, chain_id, name, address, city, state, pincode, lat, lng, capacity_max, status, created_at, updated_at)
VALUES
  ('33333333-3333-3333-3333-333333333333', '11111111-1111-1111-1111-111111111111', 'Downtown Flagship Superstore', '100 Main St', 'Bangalore', 'Karnataka', '560001', 12.9716, 77.5946, 100, 'ACTIVE', NOW(), NOW()),
  ('44444444-4444-4444-4444-444444444444', '11111111-1111-1111-1111-111111111111', 'Koramangala Express Store', '45 80-Feet Rd', 'Bangalore', 'Karnataka', '560095', 12.9352, 77.6245, 50, 'ACTIVE', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- 2. Categories
INSERT INTO categories (id, chain_id, name, created_at, updated_at)
VALUES
  ('55555555-5555-5555-5555-555555555555', '11111111-1111-1111-1111-111111111111', 'Beverages & Soft Drinks', NOW(), NOW()),
  ('66666666-6666-6666-6666-666666666666', '11111111-1111-1111-1111-111111111111', 'Snacks & Confectionery', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- 3. HSN Rates
INSERT INTO hsn_gst_rates (hsn_code, gst_rate_percent, description, created_at)
VALUES ('2201', 18.0, 'Mineral water and carbonated drinks', NOW())
ON CONFLICT (hsn_code) DO NOTHING;

-- 4. Products
INSERT INTO products (id, store_id, chain_id, barcode, name, description, category_id, price_paise, mrp_paise, hsn_code, created_at, updated_at)
VALUES
  ('77777777-7777-7777-7777-777777777777', '33333333-3333-3333-3333-333333333333', '11111111-1111-1111-1111-111111111111', '8901234567890', 'Sparkling Mineral Water 500ml', 'Pure alpine sparkling water', '55555555-5555-5555-5555-555555555555', 4500, 5000, '2201', NOW(), NOW()),
  ('88888888-8888-8888-8888-888888888888', '33333333-3333-3333-3333-333333333333', '11111111-1111-1111-1111-111111111111', '8901234567891', 'Organic Cold Brew Coffee 250ml', 'Single-origin Arabica dark roast', '55555555-5555-5555-5555-555555555555', 12000, 15000, '2201', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;
