CREATE TABLE IF NOT EXISTS issues (
    id SERIAL PRIMARY KEY,
    item_id INTEGER NOT NULL,
    quantity INTEGER NOT NULL,
    issued_to TEXT NOT NULL,
    issued_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_item
        FOREIGN KEY (item_id) 
        REFERENCES items(id)
        ON DELETE CASCADE
);

CREATE INDEX issues_item_id_idx ON issues(item_id);