-- Migration: 000003_store_service_init.down.sql

DROP INDEX IF EXISTS idx_store_sessions_active;
DROP INDEX IF EXISTS idx_stores_chain;
DROP INDEX IF EXISTS idx_store_sessions_user_active;

DROP TABLE IF EXISTS store_sessions;
DROP TABLE IF EXISTS store_qr_tokens;
DROP TABLE IF EXISTS stores;
