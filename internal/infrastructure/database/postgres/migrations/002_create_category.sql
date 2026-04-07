-- 002_create_categories_table.sql
CREATE TABLE IF NOT EXISTS categories (
    id          UUID PRIMARY KEY,
    name        VARCHAR(120) NOT NULL UNIQUE,
    description TEXT,
    created_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_categories_name ON categories(name);