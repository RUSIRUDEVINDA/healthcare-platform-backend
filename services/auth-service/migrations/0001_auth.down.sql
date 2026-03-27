-- Migration: 0001_init.down.sql
-- Rolls back the initial migration
-- Run this to completely wipe auth service tables

DROP TRIGGER IF EXISTS update_users_updated_at ON users;
DROP FUNCTION IF EXISTS update_updated_at_column();
DROP TABLE IF EXISTS refresh_tokens;
DROP TABLE IF EXISTS users;
