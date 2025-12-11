-- migrations/001_create_products_table.up.sql

CREATE TABLE IF NOT EXISTS products (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price DECIMAL(10, 2) NOT NULL CHECK (price >= 0),
    stock INTEGER NOT NULL DEFAULT 0 CHECK (stock >= 0),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
    );

-- Index for fast search by product name
CREATE INDEX idx_products_name ON products (name);

-- Index for filtering by products that are in stock
CREATE INDEX idx_products_stock ON products (stock) WHERE stock > 0;

-- Function to automatically update the updated_at column
CREATE
OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at
= CURRENT_TIMESTAMP;
RETURN NEW;
END;
$$
LANGUAGE plpgsql;

-- Trigger to automatically update updated_at on UPDATE
CREATE TRIGGER update_products_updated_at
    BEFORE UPDATE
    ON products
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Seed sample data for development
INSERT INTO products (name, description, price, stock)
VALUES ('Lenovo ThinkPad Laptop', 'Professional laptop for developers', 85000.00, 15),
       ('Mechanical Keyboard', 'Mechanical keyboard with RGB backlight', 8500.00, 50),
       ('Wireless Mouse', 'Ergonomic wireless mouse', 2500.00, 100),
       ('27-inch Monitor', '4K monitor with 144Hz refresh rate', 45000.00, 8),
       ('Headset with Microphone', 'Gaming headset with noise cancellation', 5500.00, 30);
