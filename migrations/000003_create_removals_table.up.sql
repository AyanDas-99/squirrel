CREATE TABLE IF NOT EXISTS removals (
    id SERIAL PRIMARY KEY,
    item_id INTEGER NOT NULL,
    quantity INTEGER NOT NULL,
    removed_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    remarks TEXT NOT NULL,
    CONSTRAINT fk_item
        FOREIGN KEY (item_id) 
        REFERENCES items(id)
        ON DELETE CASCADE
);

CREATE INDEX removals_item_id_idx ON removals(item_id);