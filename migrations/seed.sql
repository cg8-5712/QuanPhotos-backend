-- seed.sql
-- QuanPhotos 测试数据种子文件
-- 注意：密码均为 'password123' 的 bcrypt 哈希

-- ============================================
-- 用户数据
-- ============================================

-- 密码: password123
-- bcrypt hash: $2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy

INSERT INTO users (username, email, password_hash, role, status, avatar, bio, location) VALUES
-- 管理员
('admin', 'admin@quanphotos.com', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'admin', 'active', NULL, '系统管理员', '北京'),

-- 审查员
('reviewer01', 'reviewer01@quanphotos.com', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'reviewer', 'active', NULL, '照片审核员', '上海'),
('reviewer02', 'reviewer02@quanphotos.com', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'reviewer', 'active', NULL, '照片审核员', '广州'),

-- 普通用户
('aviator', 'aviator@example.com', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'user', 'active', NULL, '航空摄影爱好者，常驻首都机场', '北京'),
('skywalker', 'skywalker@example.com', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'user', 'active', NULL, '飞机迷，喜欢拍摄各种机型', '上海'),
('planespotter', 'planespotter@example.com', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'user', 'active', NULL, '专业航空摄影师', '深圳'),
('jetfan', 'jetfan@example.com', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'user', 'active', NULL, '波音爱好者', '成都'),
('airbus_lover', 'airbus@example.com', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'user', 'active', NULL, '空客粉丝', '杭州'),

-- 被封禁用户
('banned_user', 'banned@example.com', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'user', 'banned', NULL, '违规用户', '未知');

-- ============================================
-- 标签数据
-- ============================================

INSERT INTO tags (name, photo_count) VALUES
('Boeing', 0),
('Airbus', 0),
('787', 0),
('777', 0),
('A350', 0),
('A380', 0),
('737', 0),
('A320', 0),
('首都机场', 0),
('浦东机场', 0),
('大兴机场', 0),
('白云机场', 0),
('国航', 0),
('东航', 0),
('南航', 0),
('海航', 0),
('着陆', 0),
('起飞', 0),
('滑行', 0),
('停机坪', 0),
('日落', 0),
('夜景', 0),
('雨天', 0),
('雪景', 0);

-- ============================================
-- 照片数据 (使用 UUID 文件名)
-- ============================================

-- 用户 aviator (user_id=4) 的照片
INSERT INTO photos (
    user_id, category_id, title, description, file_path, thumbnail_path, raw_file_path, file_size, status,
    aircraft_type, airline, registration, airport,
    exif_camera_make, exif_camera_model, exif_lens_model,
    exif_focal_length, exif_aperture, exif_shutter_speed, exif_iso,
    exif_taken_at, exif_image_width, exif_image_height,
    view_count, favorite_count, approved_at
) VALUES
-- 照片 1: 787-9 国航
(4, 1, 'Boeing 787-9 国航涂装着陆', '2025年1月拍摄于北京首都机场，国航787-9梦想飞机着陆瞬间',
 '/photos/2025/01/15/a1b2c3d4-e5f6-7890-abcd-ef1234567801.jpg',
 '/thumbnails/2025/01/15/a1b2c3d4-e5f6-7890-abcd-ef1234567801',
 '/raw/2025/01/15/a1b2c3d4-e5f6-7890-abcd-ef1234567801.cr3',
 15728640, 'approved',
 'Boeing 787-9', '中国国际航空', 'B-1234', 'ZBAA',
 'Canon', 'EOS R5', 'RF 100-500mm F4.5-7.1 L IS USM',
 '500mm', 'f/7.1', '1/2000', 400,
 '2025-01-15 10:30:00', 8192, 5464,
 1250, 89, '2025-01-15 14:00:00'),

-- 照片 2: A350-900 东航
(4, 1, 'A350-900 东航着陆', '东航A350-900在首都机场36L跑道着陆',
 '/photos/2025/01/16/b2c3d4e5-f6a7-8901-bcde-f12345678902.jpg',
 '/thumbnails/2025/01/16/b2c3d4e5-f6a7-8901-bcde-f12345678902',
 NULL,
 12582912, 'approved',
 'Airbus A350-900', '中国东方航空', 'B-5678', 'ZBAA',
 'Canon', 'EOS R5', 'RF 100-500mm F4.5-7.1 L IS USM',
 '400mm', 'f/6.3', '1/1600', 320,
 '2025-01-16 15:20:00', 8192, 5464,
 980, 67, '2025-01-16 18:00:00'),

-- 用户 skywalker (user_id=5) 的照片
-- 照片 3: 777-300ER 南航
(5, 1, '777-300ER 南航起飞', '南航波音777-300ER从浦东机场起飞',
 '/photos/2025/01/18/c3d4e5f6-a7b8-9012-cdef-123456789003.jpg',
 '/thumbnails/2025/01/18/c3d4e5f6-a7b8-9012-cdef-123456789003',
 NULL,
 14680064, 'approved',
 'Boeing 777-300ER', '中国南方航空', 'B-2088', 'ZSPD',
 'Nikon', 'Z9', 'NIKKOR Z 100-400mm f/4.5-5.6 VR S',
 '400mm', 'f/5.6', '1/2500', 500,
 '2025-01-18 08:45:00', 8256, 5504,
 756, 45, '2025-01-18 12:00:00'),

-- 照片 4: 747-8F 国货航
(5, 2, 'Boeing 747-8F 国货航', '国货航747-8货机在浦东机场',
 '/photos/2025/01/19/d4e5f6a7-b8c9-0123-defa-234567890104.jpg',
 '/thumbnails/2025/01/19/d4e5f6a7-b8c9-0123-defa-234567890104',
 '/raw/2025/01/19/d4e5f6a7-b8c9-0123-defa-234567890104.nef',
 18874368, 'approved',
 'Boeing 747-8F', '中国国际货运航空', 'B-2422', 'ZSPD',
 'Nikon', 'Z9', 'NIKKOR Z 100-400mm f/4.5-5.6 VR S',
 '280mm', 'f/5.0', '1/1000', 200,
 '2025-01-19 14:30:00', 8256, 5504,
 542, 38, '2025-01-19 17:00:00'),

-- 用户 planespotter (user_id=6) 的照片
-- 照片 5: A380 南航
(6, 1, 'A380 南航巨无霸', '南航空客A380在广州白云机场',
 '/photos/2025/01/20/e5f6a7b8-c9d0-1234-efab-345678901205.jpg',
 '/thumbnails/2025/01/20/e5f6a7b8-c9d0-1234-efab-345678901205',
 '/raw/2025/01/20/e5f6a7b8-c9d0-1234-efab-345678901205.arw',
 20971520, 'approved',
 'Airbus A380-800', '中国南方航空', 'B-6138', 'ZGGG',
 'Sony', 'A1', 'FE 200-600mm F5.6-6.3 G OSS',
 '600mm', 'f/6.3', '1/3200', 640,
 '2025-01-20 11:00:00', 8640, 5760,
 2100, 156, '2025-01-20 15:00:00'),

-- 照片 6: G650ER 公务机
(6, 3, 'Gulfstream G650ER', '湾流G650ER公务机',
 '/photos/2025/01/21/f6a7b8c9-d0e1-2345-fabc-456789012306.jpg',
 '/thumbnails/2025/01/21/f6a7b8c9-d0e1-2345-fabc-456789012306',
 NULL,
 10485760, 'approved',
 'Gulfstream G650ER', 'Private', 'N650GX', 'ZGGG',
 'Sony', 'A1', 'FE 200-600mm F5.6-6.3 G OSS',
 '450mm', 'f/5.6', '1/2000', 400,
 '2025-01-21 09:15:00', 8640, 5760,
 430, 28, '2025-01-21 13:00:00'),

-- 照片 7: 待审核 (pending)
(4, 1, '787-10 南航新涂装', '南航787-10新涂装首航',
 '/photos/2025/01/22/a7b8c9d0-e1f2-3456-abcd-567890123407.jpg',
 '/thumbnails/2025/01/22/a7b8c9d0-e1f2-3456-abcd-567890123407',
 NULL,
 16777216, 'pending',
 'Boeing 787-10', '中国南方航空', 'B-2099', 'ZBAA',
 'Canon', 'EOS R5', 'RF 100-500mm F4.5-7.1 L IS USM',
 '500mm', 'f/7.1', '1/2000', 400,
 '2025-01-22 10:00:00', 8192, 5464,
 0, 0, NULL),

-- 照片 8: AI 初审通过 (ai_passed)
(5, 1, 'A321neo 春秋航空', '春秋航空A321neo',
 '/photos/2025/01/22/b8c9d0e1-f2a3-4567-bcde-678901234508.jpg',
 '/thumbnails/2025/01/22/b8c9d0e1-f2a3-4567-bcde-678901234508',
 NULL,
 11534336, 'ai_passed',
 'Airbus A321neo', '春秋航空', 'B-30CY', 'ZSPD',
 'Nikon', 'Z9', 'NIKKOR Z 100-400mm f/4.5-5.6 VR S',
 '350mm', 'f/5.6', '1/1600', 320,
 '2025-01-22 14:30:00', 8256, 5504,
 0, 0, NULL),

-- 照片 9: AI 初审拒绝 (ai_rejected)
(7, 1, '模糊的飞机照片', '测试照片',
 '/photos/2025/01/22/c9d0e1f2-a3b4-5678-cdef-789012345609.jpg',
 '/thumbnails/2025/01/22/c9d0e1f2-a3b4-5678-cdef-789012345609',
 NULL,
 2097152, 'ai_rejected',
 'Unknown', 'Unknown', '', 'ZUUU',
 'Apple', 'iPhone 15 Pro', NULL,
 '77mm', 'f/2.8', '1/500', 800,
 '2025-01-22 16:00:00', 4032, 3024,
 0, 0, NULL);

-- ============================================
-- 照片标签关联
-- ============================================

-- 照片1的标签: Boeing, 787, 首都机场, 国航, 着陆
INSERT INTO photo_tags (photo_id, tag_id) VALUES
(1, 1), (1, 3), (1, 9), (1, 13), (1, 17);

-- 照片2的标签: Airbus, A350, 首都机场, 东航, 着陆
INSERT INTO photo_tags (photo_id, tag_id) VALUES
(2, 2), (2, 5), (2, 9), (2, 14), (2, 17);

-- 照片3的标签: Boeing, 777, 浦东机场, 南航, 起飞
INSERT INTO photo_tags (photo_id, tag_id) VALUES
(3, 1), (3, 4), (3, 10), (3, 15), (3, 18);

-- 照片4的标签: Boeing, 浦东机场, 国航
INSERT INTO photo_tags (photo_id, tag_id) VALUES
(4, 1), (4, 10), (4, 13);

-- 照片5的标签: Airbus, A380, 白云机场, 南航
INSERT INTO photo_tags (photo_id, tag_id) VALUES
(5, 2), (5, 6), (5, 12), (5, 15);

-- 照片6的标签: 白云机场
INSERT INTO photo_tags (photo_id, tag_id) VALUES
(6, 12);

-- ============================================
-- 收藏数据
-- ============================================

INSERT INTO favorites (user_id, photo_id) VALUES
(4, 3),  -- aviator 收藏 skywalker 的照片
(4, 5),  -- aviator 收藏 planespotter 的照片
(5, 1),  -- skywalker 收藏 aviator 的照片
(5, 2),
(6, 1),  -- planespotter 收藏
(6, 3),
(7, 1),  -- jetfan 收藏
(7, 5),
(8, 2),  -- airbus_lover 收藏
(8, 5);

-- ============================================
-- 审核记录
-- ============================================

INSERT INTO photo_reviews (photo_id, reviewer_id, review_type, action, reason, ai_result) VALUES
-- AI 审核记录
(1, NULL, 'ai', 'approve', NULL,
 '{"score": 0.92, "aircraft_detected": true, "quality_score": 0.95, "issues": [], "aircraft_type": "Boeing 787-9", "registration": "B-1234"}'),
(2, NULL, 'ai', 'approve', NULL,
 '{"score": 0.89, "aircraft_detected": true, "quality_score": 0.91, "issues": [], "aircraft_type": "Airbus A350-900", "registration": "B-5678"}'),
(3, NULL, 'ai', 'approve', NULL,
 '{"score": 0.88, "aircraft_detected": true, "quality_score": 0.90, "issues": [], "aircraft_type": "Boeing 777-300ER", "registration": "B-2088"}'),
(4, NULL, 'ai', 'approve', NULL,
 '{"score": 0.91, "aircraft_detected": true, "quality_score": 0.88, "issues": [], "aircraft_type": "Boeing 747-8F", "registration": "B-2422"}'),
(5, NULL, 'ai', 'approve', NULL,
 '{"score": 0.95, "aircraft_detected": true, "quality_score": 0.96, "issues": [], "aircraft_type": "Airbus A380-800", "registration": "B-6138"}'),
(6, NULL, 'ai', 'approve', NULL,
 '{"score": 0.87, "aircraft_detected": true, "quality_score": 0.89, "issues": [], "aircraft_type": "Gulfstream G650", "registration": "N650GX"}'),
(8, NULL, 'ai', 'approve', NULL,
 '{"score": 0.85, "aircraft_detected": true, "quality_score": 0.86, "issues": [], "aircraft_type": "Airbus A321neo", "registration": "B-30CY"}'),
(9, NULL, 'ai', 'reject', '图片模糊，质量不达标',
 '{"score": 0.35, "aircraft_detected": false, "quality_score": 0.28, "issues": ["blurry", "low_resolution", "aircraft_not_detected"]}'),

-- 人工审核记录
(1, 2, 'manual', 'approve', NULL, NULL),
(2, 2, 'manual', 'approve', NULL, NULL),
(3, 3, 'manual', 'approve', NULL, NULL),
(4, 2, 'manual', 'approve', NULL, NULL),
(5, 3, 'manual', 'approve', NULL, NULL),
(6, 2, 'manual', 'approve', NULL, NULL);

-- ============================================
-- 工单数据
-- ============================================

INSERT INTO tickets (user_id, photo_id, type, title, content, status) VALUES
-- 申诉工单
(7, 9, 'appeal', '申诉 AI 审核结果',
 '我认为这张照片虽然是手机拍摄，但清晰度是可以接受的，希望人工复审。照片拍摄于成都双流机场，当时光线条件一般。', 'open'),

-- 功能建议工单
(4, NULL, 'other', '功能建议：批量上传',
 '建议增加批量上传功能，方便一次上传多张照片。目前每次只能上传一张，效率较低。', 'processing'),

-- 举报工单 (已解决)
(5, 5, 'report', '举报疑似盗图',
 '这张照片疑似盗图，原作者是微博用户 @航空摄影师XXX，发布于2024年12月。', 'resolved');

-- ============================================
-- 工单回复
-- ============================================

INSERT INTO ticket_replies (ticket_id, user_id, content) VALUES
-- 工单2的回复
(2, 1, '感谢您的建议！批量上传功能已加入我们的开发计划，预计在下个版本中实现。届时您可以一次选择多张照片上传。'),
(2, 4, '好的，期待新功能上线！'),

-- 工单3的回复
(3, 2, '您好，经过核实，该照片确为用户 planespotter 本人原创作品。我们已联系微博平台进行确认，该微博账号也是同一用户。举报不成立，感谢您的监督。'),
(3, 5, '好的，感谢处理。抱歉造成误会。');

-- ============================================
-- 更新统计数据
-- ============================================

-- 重新计算标签的 photo_count (确保触发器正确)
UPDATE tags t SET photo_count = (
    SELECT COUNT(*) FROM photo_tags pt WHERE pt.tag_id = t.id
);

-- 重新计算照片的 favorite_count (确保触发器正确)
UPDATE photos p SET favorite_count = (
    SELECT COUNT(*) FROM favorites f WHERE f.photo_id = p.id
);

-- ============================================
-- 验证数据
-- ============================================

-- 输出统计信息
DO $$
DECLARE
    user_count INT;
    photo_count INT;
    tag_count INT;
    ticket_count INT;
BEGIN
    SELECT COUNT(*) INTO user_count FROM users;
    SELECT COUNT(*) INTO photo_count FROM photos;
    SELECT COUNT(*) INTO tag_count FROM tags;
    SELECT COUNT(*) INTO ticket_count FROM tickets;

    RAISE NOTICE '========== Seed Data Summary ==========';
    RAISE NOTICE 'Users: %', user_count;
    RAISE NOTICE 'Photos: %', photo_count;
    RAISE NOTICE 'Tags: %', tag_count;
    RAISE NOTICE 'Tickets: %', ticket_count;
    RAISE NOTICE '========================================';
END $$;
