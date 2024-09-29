BEGIN;

DROP TABLE IF EXISTS transfers;
DROP TABLE IF EXISTS bank_accounts;
DROP SEQUENCE IF EXISTS bank_accounts_id_seq;
DROP SEQUENCE IF EXISTS transfers_id_seq;

COMMIT;
