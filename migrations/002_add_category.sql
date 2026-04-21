-- migrations/002_add_category.sql
ALTER TABLE apps ADD COLUMN IF NOT EXISTS category JSONB NOT NULL DEFAULT '[]';
