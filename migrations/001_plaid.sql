-- Run on existing databases that were created before Plaid support.

ALTER TABLE transactions
    ADD COLUMN IF NOT EXISTS plaid_transaction_id VARCHAR(255) UNIQUE;

CREATE TABLE IF NOT EXISTS plaid_items (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    item_id VARCHAR(255) NOT NULL,
    access_token TEXT NOT NULL,
    sync_cursor TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (user_id, item_id)
);
