CREATE TABLE orders (
    id uuid PRIMARY KEY,
    order_total_cost decimal(10, 2) NOT NULL,
    cart_id uuid REFERENCES carts(id) ON DELETE CASCADE,  -- Corrected to reference 'carts'
    customer_id uuid REFERENCES users(id) ON DELETE CASCADE,  -- Corrected to reference 'id' from users table
    created_at timestamp NOT NULL DEFAULT NOW(),
    updated_at timestamp NOT NULL DEFAULT NOW()
);
