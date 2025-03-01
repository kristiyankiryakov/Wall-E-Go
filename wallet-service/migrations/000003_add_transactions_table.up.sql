BEGIN;

CREATE TYPE transaction_type AS ENUM('DEPOSIT','WITHDRAW');

CREATE TABLE IF NOT EXISTS transactions(
    id SERIAL PRIMARY KEY,
    amount decimal(15,2) NOT NULL,
    wallet_id integer REFERENCES wallets (id),
    transaction_type TRANSACTION_TYPE NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

COMMIT;