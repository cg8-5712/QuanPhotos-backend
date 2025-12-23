-- 000001_init_schema.up.sql
-- QuanPhotos Initial Schema

-- ============================================
-- Functions
-- ============================================

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- ============================================
-- 1. Users Table
-- ============================================

CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'user',
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    can_comment BOOLEAN NOT NULL DEFAULT TRUE,
    can_message BOOLEAN NOT NULL DEFAULT TRUE,
    can_upload BOOLEAN NOT NULL DEFAULT TRUE,
    avatar VARCHAR(500),
    bio VARCHAR(500),
    location VARCHAR(100),
    last_login_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_users_role CHECK (role IN ('guest', 'user', 'reviewer', 'admin', 'superadmin')),
    CONSTRAINT chk_users_status CHECK (status IN ('active', 'banned'))
);

CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_status ON users(status);

CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- ============================================
-- 2. Categories Table
-- ============================================

CREATE TABLE categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    name_en VARCHAR(100) NOT NULL,
    description VARCHAR(500),
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_categories_sort_order ON categories(sort_order);

CREATE TRIGGER update_categories_updated_at
    BEFORE UPDATE ON categories
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- ============================================
-- 3. Photos Table
-- ============================================

CREATE TABLE photos (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    category_id INT REFERENCES categories(id) ON DELETE SET NULL,
    title VARCHAR(200) NOT NULL,
    description TEXT,
    file_path VARCHAR(500) NOT NULL,
    thumbnail_path VARCHAR(500),
    raw_file_path VARCHAR(500),
    file_size BIGINT,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    view_count INT NOT NULL DEFAULT 0,
    like_count INT NOT NULL DEFAULT 0,
    favorite_count INT NOT NULL DEFAULT 0,
    comment_count INT NOT NULL DEFAULT 0,
    share_count INT NOT NULL DEFAULT 0,

    -- Aviation info
    aircraft_type VARCHAR(100),
    airline VARCHAR(100),
    registration VARCHAR(20),
    airport VARCHAR(10),

    -- EXIF Camera info
    exif_camera_make VARCHAR(100),
    exif_camera_model VARCHAR(100),
    exif_serial_number VARCHAR(100),

    -- EXIF Lens info
    exif_lens_make VARCHAR(100),
    exif_lens_model VARCHAR(200),
    exif_focal_length VARCHAR(20),
    exif_focal_length_35mm VARCHAR(20),

    -- EXIF Shooting parameters
    exif_aperture VARCHAR(20),
    exif_shutter_speed VARCHAR(20),
    exif_iso INT,
    exif_exposure_mode VARCHAR(50),
    exif_exposure_program VARCHAR(50),
    exif_metering_mode VARCHAR(50),
    exif_white_balance VARCHAR(50),
    exif_flash VARCHAR(50),
    exif_exposure_bias VARCHAR(20),

    -- EXIF Time and location
    exif_taken_at TIMESTAMP,
    exif_gps_latitude DECIMAL(10, 7),
    exif_gps_longitude DECIMAL(10, 7),
    exif_gps_altitude DECIMAL(10, 2),

    -- EXIF Image info
    exif_image_width INT,
    exif_image_height INT,
    exif_orientation INT,
    exif_color_space VARCHAR(50),
    exif_software VARCHAR(100),

    -- Timestamps
    approved_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_photos_status CHECK (status IN ('pending', 'ai_passed', 'ai_rejected', 'approved', 'rejected'))
);

CREATE INDEX idx_photos_user_id ON photos(user_id);
CREATE INDEX idx_photos_category_id ON photos(category_id);
CREATE INDEX idx_photos_status ON photos(status);
CREATE INDEX idx_photos_aircraft_type ON photos(aircraft_type);
CREATE INDEX idx_photos_airline ON photos(airline);
CREATE INDEX idx_photos_registration ON photos(registration);
CREATE INDEX idx_photos_airport ON photos(airport);
CREATE INDEX idx_photos_created_at ON photos(created_at DESC);
CREATE INDEX idx_photos_exif_taken_at ON photos(exif_taken_at);

CREATE TRIGGER update_photos_updated_at
    BEFORE UPDATE ON photos
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- ============================================
-- 4. Tags Table
-- ============================================

CREATE TABLE tags (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL,
    photo_count INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_tags_name ON tags(name);
CREATE INDEX idx_tags_photo_count ON tags(photo_count DESC);

-- ============================================
-- 5. Photo Tags Table (Many-to-Many)
-- ============================================

CREATE TABLE photo_tags (
    photo_id BIGINT NOT NULL REFERENCES photos(id) ON DELETE CASCADE,
    tag_id INT NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    PRIMARY KEY (photo_id, tag_id)
);

CREATE INDEX idx_photo_tags_tag_id ON photo_tags(tag_id);

-- Tag count triggers
CREATE OR REPLACE FUNCTION increment_tag_count()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE tags SET photo_count = photo_count + 1 WHERE id = NEW.tag_id;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE OR REPLACE FUNCTION decrement_tag_count()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE tags SET photo_count = photo_count - 1 WHERE id = OLD.tag_id;
    RETURN OLD;
END;
$$ language 'plpgsql';

CREATE TRIGGER trigger_increment_tag_count
    AFTER INSERT ON photo_tags
    FOR EACH ROW
    EXECUTE FUNCTION increment_tag_count();

CREATE TRIGGER trigger_decrement_tag_count
    AFTER DELETE ON photo_tags
    FOR EACH ROW
    EXECUTE FUNCTION decrement_tag_count();

-- ============================================
-- 6. Favorites Table
-- ============================================

CREATE TABLE favorites (
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    photo_id BIGINT NOT NULL REFERENCES photos(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    PRIMARY KEY (user_id, photo_id)
);

CREATE INDEX idx_favorites_photo_id ON favorites(photo_id);
CREATE INDEX idx_favorites_created_at ON favorites(created_at DESC);

-- Favorite count triggers
CREATE OR REPLACE FUNCTION increment_favorite_count()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE photos SET favorite_count = favorite_count + 1 WHERE id = NEW.photo_id;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE OR REPLACE FUNCTION decrement_favorite_count()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE photos SET favorite_count = favorite_count - 1 WHERE id = OLD.photo_id;
    RETURN OLD;
END;
$$ language 'plpgsql';

CREATE TRIGGER trigger_increment_favorite_count
    AFTER INSERT ON favorites
    FOR EACH ROW
    EXECUTE FUNCTION increment_favorite_count();

CREATE TRIGGER trigger_decrement_favorite_count
    AFTER DELETE ON favorites
    FOR EACH ROW
    EXECUTE FUNCTION decrement_favorite_count();

-- ============================================
-- 7. Photo Reviews Table
-- ============================================

CREATE TABLE photo_reviews (
    id BIGSERIAL PRIMARY KEY,
    photo_id BIGINT NOT NULL REFERENCES photos(id) ON DELETE CASCADE,
    reviewer_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    review_type VARCHAR(20) NOT NULL,
    action VARCHAR(20) NOT NULL,
    reason TEXT,
    ai_result JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_photo_reviews_type CHECK (review_type IN ('ai', 'manual')),
    CONSTRAINT chk_photo_reviews_action CHECK (action IN ('approve', 'reject'))
);

CREATE INDEX idx_photo_reviews_photo_id ON photo_reviews(photo_id);
CREATE INDEX idx_photo_reviews_reviewer_id ON photo_reviews(reviewer_id);
CREATE INDEX idx_photo_reviews_created_at ON photo_reviews(created_at DESC);

-- ============================================
-- 8. Tickets Table
-- ============================================

CREATE TABLE tickets (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    photo_id BIGINT REFERENCES photos(id) ON DELETE SET NULL,
    type VARCHAR(20) NOT NULL,
    title VARCHAR(200) NOT NULL,
    content TEXT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'open',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_tickets_type CHECK (type IN ('appeal', 'report', 'other')),
    CONSTRAINT chk_tickets_status CHECK (status IN ('open', 'processing', 'resolved', 'closed'))
);

CREATE INDEX idx_tickets_user_id ON tickets(user_id);
CREATE INDEX idx_tickets_photo_id ON tickets(photo_id);
CREATE INDEX idx_tickets_status ON tickets(status);
CREATE INDEX idx_tickets_type ON tickets(type);
CREATE INDEX idx_tickets_created_at ON tickets(created_at DESC);

CREATE TRIGGER update_tickets_updated_at
    BEFORE UPDATE ON tickets
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- ============================================
-- 9. Ticket Replies Table
-- ============================================

CREATE TABLE ticket_replies (
    id BIGSERIAL PRIMARY KEY,
    ticket_id BIGINT NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ticket_replies_ticket_id ON ticket_replies(ticket_id);
CREATE INDEX idx_ticket_replies_created_at ON ticket_replies(created_at);

-- ============================================
-- 10. Refresh Tokens Table
-- ============================================

CREATE TABLE refresh_tokens (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);

-- ============================================
-- 11. Photo Likes Table
-- ============================================

CREATE TABLE photo_likes (
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    photo_id BIGINT NOT NULL REFERENCES photos(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    PRIMARY KEY (user_id, photo_id)
);

CREATE INDEX idx_photo_likes_photo_id ON photo_likes(photo_id);
CREATE INDEX idx_photo_likes_created_at ON photo_likes(created_at DESC);

-- Photo like count triggers
CREATE OR REPLACE FUNCTION increment_photo_like_count()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE photos SET like_count = like_count + 1 WHERE id = NEW.photo_id;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE OR REPLACE FUNCTION decrement_photo_like_count()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE photos SET like_count = like_count - 1 WHERE id = OLD.photo_id;
    RETURN OLD;
END;
$$ language 'plpgsql';

CREATE TRIGGER trigger_increment_photo_like_count
    AFTER INSERT ON photo_likes
    FOR EACH ROW
    EXECUTE FUNCTION increment_photo_like_count();

CREATE TRIGGER trigger_decrement_photo_like_count
    AFTER DELETE ON photo_likes
    FOR EACH ROW
    EXECUTE FUNCTION decrement_photo_like_count();

-- ============================================
-- 12. Photo Comments Table
-- ============================================

CREATE TABLE photo_comments (
    id BIGSERIAL PRIMARY KEY,
    photo_id BIGINT NOT NULL REFERENCES photos(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    parent_id BIGINT REFERENCES photo_comments(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    like_count INT NOT NULL DEFAULT 0,
    reply_count INT NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'visible',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_photo_comments_status CHECK (status IN ('visible', 'hidden', 'deleted'))
);

CREATE INDEX idx_photo_comments_photo_id ON photo_comments(photo_id);
CREATE INDEX idx_photo_comments_user_id ON photo_comments(user_id);
CREATE INDEX idx_photo_comments_parent_id ON photo_comments(parent_id);
CREATE INDEX idx_photo_comments_created_at ON photo_comments(created_at DESC);

CREATE TRIGGER update_photo_comments_updated_at
    BEFORE UPDATE ON photo_comments
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Photo comment count triggers
CREATE OR REPLACE FUNCTION increment_photo_comment_count()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE photos SET comment_count = comment_count + 1 WHERE id = NEW.photo_id;
    IF NEW.parent_id IS NOT NULL THEN
        UPDATE photo_comments SET reply_count = reply_count + 1 WHERE id = NEW.parent_id;
    END IF;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE OR REPLACE FUNCTION decrement_photo_comment_count()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE photos SET comment_count = comment_count - 1 WHERE id = OLD.photo_id;
    IF OLD.parent_id IS NOT NULL THEN
        UPDATE photo_comments SET reply_count = reply_count - 1 WHERE id = OLD.parent_id;
    END IF;
    RETURN OLD;
END;
$$ language 'plpgsql';

CREATE TRIGGER trigger_increment_photo_comment_count
    AFTER INSERT ON photo_comments
    FOR EACH ROW
    EXECUTE FUNCTION increment_photo_comment_count();

CREATE TRIGGER trigger_decrement_photo_comment_count
    AFTER DELETE ON photo_comments
    FOR EACH ROW
    EXECUTE FUNCTION decrement_photo_comment_count();

-- ============================================
-- 13. Comment Likes Table
-- ============================================

CREATE TABLE comment_likes (
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    comment_id BIGINT NOT NULL REFERENCES photo_comments(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    PRIMARY KEY (user_id, comment_id)
);

CREATE INDEX idx_comment_likes_comment_id ON comment_likes(comment_id);

-- Comment like count triggers
CREATE OR REPLACE FUNCTION increment_comment_like_count()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE photo_comments SET like_count = like_count + 1 WHERE id = NEW.comment_id;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE OR REPLACE FUNCTION decrement_comment_like_count()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE photo_comments SET like_count = like_count - 1 WHERE id = OLD.comment_id;
    RETURN OLD;
END;
$$ language 'plpgsql';

CREATE TRIGGER trigger_increment_comment_like_count
    AFTER INSERT ON comment_likes
    FOR EACH ROW
    EXECUTE FUNCTION increment_comment_like_count();

CREATE TRIGGER trigger_decrement_comment_like_count
    AFTER DELETE ON comment_likes
    FOR EACH ROW
    EXECUTE FUNCTION decrement_comment_like_count();

-- ============================================
-- 14. Photo Shares Table
-- ============================================

CREATE TABLE photo_shares (
    id BIGSERIAL PRIMARY KEY,
    photo_id BIGINT NOT NULL REFERENCES photos(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content TEXT,
    share_type VARCHAR(20) NOT NULL DEFAULT 'internal',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_photo_shares_type CHECK (share_type IN ('internal', 'external'))
);

CREATE INDEX idx_photo_shares_photo_id ON photo_shares(photo_id);
CREATE INDEX idx_photo_shares_user_id ON photo_shares(user_id);
CREATE INDEX idx_photo_shares_created_at ON photo_shares(created_at DESC);

-- Share count trigger
CREATE OR REPLACE FUNCTION increment_share_count()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE photos SET share_count = share_count + 1 WHERE id = NEW.photo_id;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER trigger_increment_share_count
    AFTER INSERT ON photo_shares
    FOR EACH ROW
    EXECUTE FUNCTION increment_share_count();

-- ============================================
-- 15. Featured Photos Table
-- ============================================

CREATE TABLE featured_photos (
    id BIGSERIAL PRIMARY KEY,
    photo_id BIGINT UNIQUE NOT NULL REFERENCES photos(id) ON DELETE CASCADE,
    admin_id BIGINT NOT NULL REFERENCES users(id),
    reason VARCHAR(500),
    sort_order INT NOT NULL DEFAULT 0,
    featured_at TIMESTAMP NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP
);

CREATE INDEX idx_featured_photos_sort_order ON featured_photos(sort_order);
CREATE INDEX idx_featured_photos_featured_at ON featured_photos(featured_at DESC);

-- ============================================
-- 16. Conversations Table
-- ============================================

CREATE TABLE conversations (
    id BIGSERIAL PRIMARY KEY,
    user1_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    user2_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    last_message_id BIGINT,
    user1_unread INT NOT NULL DEFAULT 0,
    user2_unread INT NOT NULL DEFAULT 0,
    user1_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    user2_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_conversations_users CHECK (user1_id < user2_id),
    CONSTRAINT uq_conversations_users UNIQUE (user1_id, user2_id)
);

CREATE INDEX idx_conversations_user1_id ON conversations(user1_id);
CREATE INDEX idx_conversations_user2_id ON conversations(user2_id);
CREATE INDEX idx_conversations_updated_at ON conversations(updated_at DESC);

CREATE TRIGGER update_conversations_updated_at
    BEFORE UPDATE ON conversations
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- ============================================
-- 17. Messages Table
-- ============================================

CREATE TABLE messages (
    id BIGSERIAL PRIMARY KEY,
    conversation_id BIGINT NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    sender_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    is_read BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_messages_conversation_id ON messages(conversation_id);
CREATE INDEX idx_messages_sender_id ON messages(sender_id);
CREATE INDEX idx_messages_created_at ON messages(created_at DESC);

-- Add foreign key for last_message_id after messages table is created
ALTER TABLE conversations
    ADD CONSTRAINT fk_conversations_last_message
    FOREIGN KEY (last_message_id) REFERENCES messages(id) ON DELETE SET NULL;

-- ============================================
-- 18. Notifications Table
-- ============================================

CREATE TABLE notifications (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    actor_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    type VARCHAR(20) NOT NULL,
    title VARCHAR(200) NOT NULL,
    content TEXT,
    related_photo_id BIGINT REFERENCES photos(id) ON DELETE SET NULL,
    related_comment_id BIGINT REFERENCES photo_comments(id) ON DELETE SET NULL,
    is_read BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_notifications_type CHECK (type IN ('like', 'comment', 'reply', 'follow', 'share', 'featured', 'review', 'system', 'message'))
);

CREATE INDEX idx_notifications_user_id ON notifications(user_id);
CREATE INDEX idx_notifications_type ON notifications(type);
CREATE INDEX idx_notifications_is_read ON notifications(is_read);
CREATE INDEX idx_notifications_created_at ON notifications(created_at DESC);

-- ============================================
-- 19. Announcements Table
-- ============================================

CREATE TABLE announcements (
    id BIGSERIAL PRIMARY KEY,
    author_id BIGINT NOT NULL REFERENCES users(id),
    title VARCHAR(200) NOT NULL,
    summary VARCHAR(500),
    content TEXT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    is_pinned BOOLEAN NOT NULL DEFAULT FALSE,
    published_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_announcements_status CHECK (status IN ('draft', 'published'))
);

CREATE INDEX idx_announcements_status ON announcements(status);
CREATE INDEX idx_announcements_is_pinned ON announcements(is_pinned);
CREATE INDEX idx_announcements_published_at ON announcements(published_at DESC);

CREATE TRIGGER update_announcements_updated_at
    BEFORE UPDATE ON announcements
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- ============================================
-- 20. Reviewer Categories Table
-- ============================================

CREATE TABLE reviewer_categories (
    reviewer_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    category_id INT NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    PRIMARY KEY (reviewer_id, category_id)
);

CREATE INDEX idx_reviewer_categories_category_id ON reviewer_categories(category_id);

-- ============================================
-- 21. Admin Permissions Table
-- ============================================

CREATE TABLE admin_permissions (
    admin_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    permission VARCHAR(50) NOT NULL,
    granted_by BIGINT NOT NULL REFERENCES users(id),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    PRIMARY KEY (admin_id, permission),
    CONSTRAINT chk_admin_permissions_permission CHECK (permission IN (
        'manage_announcements',
        'manage_featured',
        'ban_users',
        'mute_comment',
        'mute_message',
        'mute_upload',
        'review_photos',
        'delete_photos',
        'delete_comments',
        'manage_tickets',
        'manage_categories',
        'manage_tags',
        'view_statistics',
        'view_user_details'
    ))
);

CREATE INDEX idx_admin_permissions_permission ON admin_permissions(permission);
