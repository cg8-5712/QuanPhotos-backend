-- 000001_init_schema.down.sql
-- Rollback QuanPhotos Initial Schema

-- Drop tables in reverse order of creation (respecting foreign key constraints)
DROP TABLE IF EXISTS refresh_tokens;
DROP TABLE IF EXISTS ticket_replies;
DROP TABLE IF EXISTS tickets;
DROP TABLE IF EXISTS photo_reviews;
DROP TABLE IF EXISTS favorites;
DROP TABLE IF EXISTS photo_tags;
DROP TABLE IF EXISTS tags;
DROP TABLE IF EXISTS photos;
DROP TABLE IF EXISTS categories;
DROP TABLE IF EXISTS users;

-- Drop functions
DROP FUNCTION IF EXISTS update_updated_at_column();
DROP FUNCTION IF EXISTS increment_tag_count();
DROP FUNCTION IF EXISTS decrement_tag_count();
DROP FUNCTION IF EXISTS increment_favorite_count();
DROP FUNCTION IF EXISTS decrement_favorite_count();
