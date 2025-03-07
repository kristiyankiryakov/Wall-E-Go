BEGIN;

CREATE TYPE transaction_type AS ENUM('DEPOSIT', 'WITHDRAW');

CREATE TABLE IF NOT EXISTS transactions(
id uuid DEFAULT gen_random_uuid(),
wallet_id uuid NOT NULL,
amount DECIMAL(15,2) NOT NULL,
type TRANSACTION_TYPE NOT NULL,
idempotency_key VARCHAR(255) UNIQUE,
status VARCHAR(20) DEFAULT 'PENDING',
created_at TIMESTAMP DEFAULT NOW(),

PRIMARY KEY(id)
);

COMMIT;