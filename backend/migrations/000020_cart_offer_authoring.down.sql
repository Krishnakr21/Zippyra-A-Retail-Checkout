-- Migration 000020 Down: Rollback Cart Offer Authoring Tables

DROP INDEX IF EXISTS idx_offers_schedule;
DROP INDEX IF EXISTS idx_offers_chain_wide;
DROP INDEX IF EXISTS idx_offers_store;
DROP TABLE IF EXISTS offers;
