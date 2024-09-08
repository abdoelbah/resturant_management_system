CREATE TABLE items (
    id uuid PRIMARY KEY,
    name varchar(255) NOT NULL,
    img varchar(255),
    price decimal(10, 2) NOT NULL,
    vendor_id uuid REFERENCES users(id) ON DELETE CASCADE,  -- Reference to 'users' table instead of 'vendors'
    created_at timestamp NOT NULL DEFAULT NOW(),
    updated_at timestamp NOT NULL DEFAULT NOW()
);
