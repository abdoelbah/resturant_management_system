CREATE TABLE roles (
    id   SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL
);

-- Insert default roles into the roles table
INSERT INTO roles (id, name)
VALUES
    (1, 'admin'),
    (2, 'vendor'),
    (3, 'customer')
ON CONFLICT (id) DO NOTHING;