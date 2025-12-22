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
-- Users Table
-- ============================================

CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'user',
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    avatar VARCHAR(500),
    bio VARCHAR(500),
    location VARCHAR(100),
    last_login_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_users_role CHECK (role IN ('guest', 'user', 'reviewer', 'admin')),
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
-- Categories Table
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
-- Photos Table
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
    favorite_count INT NOT NULL DEFAULT 0,

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
-- Tags Table
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
-- Photo Tags Table (Many-to-Many)
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
-- Favorites Table
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
-- Photo Reviews Table
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
-- Tickets Table
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
-- Ticket Replies Table
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
-- Refresh Tokens Table
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
