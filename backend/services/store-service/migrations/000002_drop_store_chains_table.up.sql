-- Post-cutover migration: drop chains table from store-service DB.
-- Run ONLY after admin-store-service has been verified in staging AND production
-- for the defined bake period (e.g. 1 week).
-- chains.id values remain valid as store.chain_id column FK references are now
-- maintained at application-level by admin-store-service (chain status validated
-- before calling store-service's internal write endpoints).

DROP TABLE IF EXISTS chains;
