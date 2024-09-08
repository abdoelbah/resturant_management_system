CREATE TABLE order_item (
    order_id uuid REFERENCES orders(id) ON DELETE CASCADE,
    item_id uuid REFERENCES items(id) ON DELETE CASCADE,
    quantity int NOT NULL,
    price decimal(10, 2) NOT NULL,
    PRIMARY KEY (order_id, item_id)
);
