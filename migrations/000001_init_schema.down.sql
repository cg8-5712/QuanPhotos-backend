-- 000001_init_schema.down.sql
-- Rollback QuanPhotos Initial Schema

-- Drop tables in reverse order of creation (respecting foreign key constraints)

-- 21. Admin Permissions
DROP TABLE IF EXISTS admin_permissions;

-- 20. Reviewer Categories
DROP TABLE IF EXISTS reviewer_categories;

-- 19. Announcements
DROP TABLE IF EXISTS announcements;

-- 18. Notifications
DROP TABLE IF EXISTS notifications;

-- 17. Messages (must drop before conversations due to FK)
DROP TABLE IF EXISTS messages;

-- 16. Conversations
DROP TABLE IF EXISTS conversations;

-- 15. Featured Photos
DROP TABLE IF EXISTS featured_photos;

-- 14. Photo Shares
DROP TABLE IF EXISTS photo_shares;

-- 13. Comment Likes
DROP TABLE IF EXISTS comment_likes;

-- 12. Photo Comments
DROP TABLE IF EXISTS photo_comments;

-- 11. Photo Likes
DROP TABLE IF EXISTS photo_likes;

-- 10. Refresh Tokens
DROP TABLE IF EXISTS refresh_tokens;

-- 9. Ticket Replies
DROP TABLE IF EXISTS ticket_replies;

-- 8. Tickets
DROP TABLE IF EXISTS tickets;

-- 7. Photo Reviews
DROP TABLE IF EXISTS photo_reviews;

-- 6. Favorites
DROP TABLE IF EXISTS favorites;

-- 5. Photo Tags
DROP TABLE IF EXISTS photo_tags;

-- 4. Tags
DROP TABLE IF EXISTS tags;

-- 3. Photos
DROP TABLE IF EXISTS photos;

-- 2. Categories
DROP TABLE IF EXISTS categories;

-- 1. Users
DROP TABLE IF EXISTS users;

-- Drop functions
DROP FUNCTION IF EXISTS update_updated_at_column() CASCADE;
DROP FUNCTION IF EXISTS increment_tag_count() CASCADE;
DROP FUNCTION IF EXISTS decrement_tag_count() CASCADE;
DROP FUNCTION IF EXISTS increment_favorite_count() CASCADE;
DROP FUNCTION IF EXISTS decrement_favorite_count() CASCADE;
DROP FUNCTION IF EXISTS increment_photo_like_count() CASCADE;
DROP FUNCTION IF EXISTS decrement_photo_like_count() CASCADE;
DROP FUNCTION IF EXISTS increment_photo_comment_count() CASCADE;
DROP FUNCTION IF EXISTS decrement_photo_comment_count() CASCADE;
DROP FUNCTION IF EXISTS increment_comment_like_count() CASCADE;
DROP FUNCTION IF EXISTS decrement_comment_like_count() CASCADE;
DROP FUNCTION IF EXISTS increment_share_count() CASCADE;
