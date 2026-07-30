INSERT INTO products (id, store_id, chain_id, barcode, name, price_paise, mrp_paise, hsn_code, is_active)
VALUES
  ('77777777-7777-7777-7777-777777777771', '33333333-3333-3333-3333-333333333333', '11111111-1111-1111-1111-111111111111', '8901030012345', 'Amul Taaza Toned Milk 1L', 6400, 6800, '2201', true),
  ('77777777-7777-7777-7777-777777777772', '33333333-3333-3333-3333-333333333333', '11111111-1111-1111-1111-111111111111', '8901058852101', 'Britannia Good Day Biscuits 200g', 3000, 3500, '2201', true),
  ('77777777-7777-7777-7777-777777777773', '33333333-3333-3333-3333-333333333333', '11111111-1111-1111-1111-111111111111', '8901491101853', 'Lays Magic Masala Potato Chips 50g', 2000, 2000, '2201', true),
  ('77777777-7777-7777-7777-777777777774', '33333333-3333-3333-3333-333333333333', '11111111-1111-1111-1111-111111111111', '8901262010014', 'Tata Iodized Salt 1kg', 2800, 2800, '2201', true),
  ('77777777-7777-7777-7777-777777777775', '33333333-3333-3333-3333-333333333333', '11111111-1111-1111-1111-111111111111', '8901030045612', 'Nescafe Classic Coffee 50g', 18000, 19500, '2201', true)
ON CONFLICT (store_id, barcode) WHERE deleted_at IS NULL DO NOTHING;

-- Seed Staff Members for Store 1
INSERT INTO staff_members (id, store_id, chain_id, phone, name, role, is_active)
VALUES
  ('99999999-9999-9999-9999-999999999901', '33333333-3333-3333-3333-333333333333', '11111111-1111-1111-1111-111111111111', '+9198210338127', 'Ramesh Kumar', 'CASHIER', true),
  ('99999999-9999-9999-9999-999999999902', '33333333-3333-3333-3333-333333333333', '11111111-1111-1111-1111-111111111111', '+9198210338128', 'Suresh Patel', 'STORE_MANAGER', true),
  ('99999999-9999-9999-9999-999999999903', '33333333-3333-3333-3333-333333333333', '11111111-1111-1111-1111-111111111111', '+9198210338129', 'Vikram Singh', 'SECURITY', true)
ON CONFLICT (phone) DO NOTHING;

-- Seed Inter-Store Transfer Orders
INSERT INTO transfer_orders (id, source_store_id, dest_store_id, chain_id, status, requested_by, created_at)
VALUES
  ('c1111111-1111-1111-1111-111111111111', '33333333-3333-3333-3333-333333333333', '44444444-4444-4444-4444-444444444444', '11111111-1111-1111-1111-111111111111', 'REQUESTED', '99999999-9999-9999-9999-999999999902', NOW() - INTERVAL '1 day'),
  ('c2222222-2222-2222-2222-222222222222', '44444444-4444-4444-4444-444444444444', '33333333-3333-3333-3333-333333333333', '11111111-1111-1111-1111-111111111111', 'APPROVED', '99999999-9999-9999-9999-999999999901', NOW() - INTERVAL '3 days')
ON CONFLICT (id) DO NOTHING;

INSERT INTO transfer_line_items (id, transfer_id, barcode, qty_requested, qty_shipped, qty_received)
VALUES
  (gen_random_uuid(), 'c1111111-1111-1111-1111-111111111111', '8901030012345', 20, 0, 0),
  (gen_random_uuid(), 'c2222222-2222-2222-2222-222222222222', '8901491101853', 50, 50, 0)
ON CONFLICT DO NOTHING;

-- Seed Real Daily Orders for the past 30 days
DO $$
DECLARE
    i INT;
    order_date TIMESTAMPTZ;
    ord_id UUID;
BEGIN
    FOR i IN 0..29 LOOP
        order_date := NOW() - (i || ' days')::INTERVAL;
        
        -- Create Order 1 of the day
        ord_id := gen_random_uuid();
        INSERT INTO orders (id, payment_id, user_id, store_id, items, subtotal_paise, discount_paise, total_paise, payment_method, status, created_at)
        VALUES (
            ord_id,
            gen_random_uuid(),
            'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
            '33333333-3333-3333-3333-333333333333',
            '[{"barcode": "8901030012345", "name": "Amul Taaza Toned Milk 1L", "qty": 3, "unit_price_paise": 6400, "total_price_paise": 19200}, {"barcode": "8901058852101", "name": "Britannia Good Day Biscuits 200g", "qty": 2, "unit_price_paise": 3000, "total_price_paise": 6000}]'::jsonb,
            25200 + (i * 300),
            1200,
            24000 + (i * 300),
            'UPI',
            'COMPLETED',
            order_date
        );

        -- Create Order 2 of the day
        ord_id := gen_random_uuid();
        INSERT INTO orders (id, payment_id, user_id, store_id, items, subtotal_paise, discount_paise, total_paise, payment_method, status, created_at)
        VALUES (
            ord_id,
            gen_random_uuid(),
            'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb',
            '33333333-3333-3333-3333-333333333333',
            '[{"barcode": "8901030045612", "name": "Nescafe Classic Coffee 50g", "qty": 1, "unit_price_paise": 18000, "total_price_paise": 18000}, {"barcode": "8901262010014", "name": "Tata Iodized Salt 1kg", "qty": 4, "unit_price_paise": 2800, "total_price_paise": 11200}]'::jsonb,
            29200 + (i * 500),
            1500,
            27700 + (i * 500),
            'CARD',
            'COMPLETED',
            order_date - INTERVAL '4 hours'
        );
    END LOOP;
END $$;
