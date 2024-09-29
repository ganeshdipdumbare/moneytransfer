BEGIN;

-- Create bank_accounts table
CREATE TABLE IF NOT EXISTS bank_accounts (
    id BIGINT PRIMARY KEY,
    organization_name TEXT NOT NULL,
    iban TEXT NOT NULL UNIQUE,
    bic TEXT NOT NULL,
    balance_cents BIGINT NOT NULL DEFAULT 0
);

-- Create sequence for bank_accounts
CREATE SEQUENCE IF NOT EXISTS bank_accounts_id_seq START WITH 1;

-- Set the sequence as the default for the id column
ALTER TABLE bank_accounts ALTER COLUMN id SET DEFAULT nextval('bank_accounts_id_seq');

-- Create transfers table
CREATE TABLE IF NOT EXISTS transfers (
    id BIGINT PRIMARY KEY,
    counterparty_name TEXT NOT NULL,
    counterparty_iban TEXT NOT NULL,
    counterparty_bic TEXT NOT NULL,
    amount_cents BIGINT NOT NULL,
    bank_account_id BIGINT NOT NULL,
    description TEXT,
    FOREIGN KEY (bank_account_id) REFERENCES bank_accounts(id)
);

-- Create sequence for transfers
CREATE SEQUENCE IF NOT EXISTS transfers_id_seq START WITH 1;

-- Set the sequence as the default for the id column
ALTER TABLE transfers ALTER COLUMN id SET DEFAULT nextval('transfers_id_seq');

-- Insert sample data if the bank_accounts table is empty
INSERT INTO bank_accounts (organization_name, balance_cents, iban, bic)
SELECT 'Acme', 10000000, 'FR10474608000002006107XXXXX', 'BNPAFRPP'
WHERE NOT EXISTS (SELECT 1 FROM bank_accounts);

COMMIT;