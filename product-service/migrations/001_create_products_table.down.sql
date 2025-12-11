-- migrations/001_create_products_table.down.sql

-- Drop trigger for automatic updated_at changes
DROP TRIGGER IF EXISTS update_products_updated_at ON products;

-- Drop function used by the trigger
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop indexes explicitly
DROP INDEX IF EXISTS idx_products_stock;
DROP INDEX IF EXISTS idx_products_name;

-- Drop products table
DROP TABLE IF EXISTS products;
