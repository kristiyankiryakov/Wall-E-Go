CREATE TABLE IF NOT EXISTS wallets(
id uuid DEFAULT gen_random_uuid(),
user_id INT NOT NULL,
name VARCHAR(255) NOT NULL,
balance decimal(15,2) DEFAULT 0.00 CHECK (balance >= 0),
created_at TIMESTAMP DEFAULT NOW(),
updated_at TIMESTAMP DEFAULT NOW(),

PRIMARY KEY(id)
);