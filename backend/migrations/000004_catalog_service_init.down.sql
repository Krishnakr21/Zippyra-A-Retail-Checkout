-- Migration: 000004_catalog_service_init.down.sql

DROP TABLE IF EXISTS catalog_import_jobs;
DROP INDEX IF EXISTS idx_products_category;
DROP INDEX IF EXISTS idx_products_store_sync;
DROP INDEX IF EXISTS idx_products_store_barcode;
DROP TABLE IF EXISTS products;
DROP TABLE IF EXISTS hsn_gst_rates;
DROP TABLE IF EXISTS categories;
DROP SEQUENCE IF EXISTS catalog_sync_seq;
