CREATE TABLE IF NOT EXISTS wallets(
id SERIAL PRIMARY KEY,
user_id INT NOT NULL,
name VARCHAR(255) NOT NULL,
balance decimal(15,2) DEFAULT 0.00,
created_at TIMESTAMP DEFAULT NOW(),
updated_at TIMESTAMP DEFAULT NOW()
);