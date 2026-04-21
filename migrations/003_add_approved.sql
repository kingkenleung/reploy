-- migrations/003_add_approved.sql
ALTER TABLE apps ADD COLUMN IF NOT EXISTS approved BOOLEAN NOT NULL DEFAULT false;
