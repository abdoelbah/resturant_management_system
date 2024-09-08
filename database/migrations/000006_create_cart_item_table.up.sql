CREATE TABLE cart_item (
    cart_id uuid REFERENCES carts(id) ON DELETE CASCADE,  -- Corrected to reference 'carts'
    item_id uuid REFERENCES items(id) ON DELETE CASCADE,
    quantity int NOT NULL,
    PRIMARY KEY (cart_id, item_id)
);
