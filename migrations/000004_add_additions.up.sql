CREATE TABLE IF NOT EXISTS additions (
    id SERIAL PRIMARY KEY,
    item_id INT NOT NULL REFERENCES items(id) ON DELETE CASCADE,
    quantity INT NOT NULL,
    remarks TEXT,
    added_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);