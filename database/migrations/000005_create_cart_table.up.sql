CREATE TABLE carts (
    id uuid PRIMARY KEY,
    user_id uuid REFERENCES users(id) ON DELETE CASCADE,  -- Corrected to reference 'id' from users table
    total_price decimal(10, 2) NOT NULL,
    quantity int NOT NULL,
    created_at timestamp NOT NULL DEFAULT NOW(),
    updated_at timestamp NOT NULL DEFAULT NOW()
);
