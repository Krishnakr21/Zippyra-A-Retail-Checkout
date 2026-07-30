-- Migration 000021 Down: Integration Service Initial Schema

DROP INDEX IF EXISTS idx_erp_sync_jobs_pending;
DROP TABLE IF EXISTS erp_sync_jobs;

DROP TABLE IF EXISTS erp_webhook_events;

DROP INDEX IF EXISTS idx_erp_connections_chain;
DROP TABLE IF EXISTS erp_connections;
