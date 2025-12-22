# QuanPhotos 数据库设计文档

## 概述

本文档描述 QuanPhotos 后端服务的数据库设计，使用 PostgreSQL 15+。

---

## ER 图

```
┌─────────────┐       ┌─────────────┐       ┌─────────────┐
│   users     │       │   photos    │       │ categories  │
├─────────────┤       ├─────────────┤       ├─────────────┤
│ id (PK)     │◄──┐   │ id (PK)     │   ┌──►│ id (PK)     │
│ username    │   │   │ user_id(FK) │───┘   │ name        │
│ email       │   └───│ category_id │       │ name_en     │
│ password    │       │ title       │       └─────────────┘
│ role        │       │ status      │
│ ...         │       │ exif_*      │       ┌─────────────┐
└─────────────┘       │ ...         │       │    tags     │
                      └─────────────┘       ├─────────────┤
                            │               │ id (PK)     │
                            │               │ name        │
                      ┌─────┴─────┐         └─────────────┘
                      ▼           ▼               │
              ┌─────────────┐ ┌─────────────┐     │
              │ photo_tags  │ │  favorites  │     │
              ├─────────────┤ ├─────────────┤     │
              │ photo_id    │ │ user_id     │     │
              │ tag_id      │─┼─────────────┼─────┘
              └─────────────┘ │ photo_id    │
                              └─────────────┘

┌─────────────┐       ┌─────────────────┐
│  tickets    │       │ ticket_replies  │
├─────────────┤       ├─────────────────┤
│ id (PK)     │◄──────│ ticket_id (FK)  │
│ user_id(FK) │       │ user_id (FK)    │
│ photo_id(FK)│       │ content         │
│ type        │       └─────────────────┘
│ status      │
└─────────────┘

┌─────────────────┐
│  photo_reviews  │
├─────────────────┤
│ id (PK)         │
│ photo_id (FK)   │
│ reviewer_id(FK) │
│ action          │
│ ai_result       │
└─────────────────┘
```

---

## 表结构详情

### 1. users - 用户表

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGSERIAL | PRIMARY KEY | 用户 ID |
| username | VARCHAR(50) | UNIQUE NOT NULL | 用户名 |
| email | VARCHAR(255) | UNIQUE NOT NULL | 邮箱 |
| password_hash | VARCHAR(255) | NOT NULL | 密码哈希 (bcrypt) |
| role | VARCHAR(20) | NOT NULL DEFAULT 'user' | 角色: guest/user/reviewer/admin |
| status | VARCHAR(20) | NOT NULL DEFAULT 'active' | 状态: active/banned |
| avatar | VARCHAR(500) | | 头像 URL |
| bio | VARCHAR(500) | | 个人简介 |
| location | VARCHAR(100) | | 所在地 |
| last_login_at | TIMESTAMP | | 最后登录时间 |
| created_at | TIMESTAMP | NOT NULL DEFAULT NOW() | 创建时间 |
| updated_at | TIMESTAMP | NOT NULL DEFAULT NOW() | 更新时间 |

**索引：**
- `idx_users_username` ON username
- `idx_users_email` ON email
- `idx_users_role` ON role
- `idx_users_status` ON status

---

### 2. categories - 分类表

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | SERIAL | PRIMARY KEY | 分类 ID |
| name | VARCHAR(100) | UNIQUE NOT NULL | 中文名称 |
| name_en | VARCHAR(100) | NOT NULL | 英文名称 |
| description | VARCHAR(500) | | 分类描述 |
| sort_order | INT | NOT NULL DEFAULT 0 | 排序顺序 |
| created_at | TIMESTAMP | NOT NULL DEFAULT NOW() | 创建时间 |
| updated_at | TIMESTAMP | NOT NULL DEFAULT NOW() | 更新时间 |

**索引：**
- `idx_categories_sort_order` ON sort_order

---

### 3. photos - 照片表

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGSERIAL | PRIMARY KEY | 照片 ID |
| user_id | BIGINT | NOT NULL REFERENCES users(id) | 上传者 ID |
| category_id | INT | REFERENCES categories(id) | 分类 ID |
| title | VARCHAR(200) | NOT NULL | 标题 |
| description | TEXT | | 描述 |
| file_path | VARCHAR(500) | NOT NULL | 文件路径 |
| thumbnail_path | VARCHAR(500) | | 缩略图路径 |
| raw_file_path | VARCHAR(500) | | RAW 文件路径 |
| file_size | BIGINT | | 文件大小 (bytes) |
| status | VARCHAR(20) | NOT NULL DEFAULT 'pending' | 状态 |
| view_count | INT | NOT NULL DEFAULT 0 | 浏览次数 |
| favorite_count | INT | NOT NULL DEFAULT 0 | 收藏次数 |
| **航空信息** |
| aircraft_type | VARCHAR(100) | | 机型 |
| airline | VARCHAR(100) | | 航空公司 |
| registration | VARCHAR(20) | | 注册号 |
| airport | VARCHAR(10) | | 机场代码 (ICAO/IATA) |
| **EXIF 相机信息** |
| exif_camera_make | VARCHAR(100) | | 相机品牌 |
| exif_camera_model | VARCHAR(100) | | 相机型号 |
| exif_serial_number | VARCHAR(100) | | 机身序列号 |
| **EXIF 镜头信息** |
| exif_lens_make | VARCHAR(100) | | 镜头品牌 |
| exif_lens_model | VARCHAR(200) | | 镜头型号 |
| exif_focal_length | VARCHAR(20) | | 焦距 |
| exif_focal_length_35mm | VARCHAR(20) | | 等效 35mm 焦距 |
| **EXIF 拍摄参数** |
| exif_aperture | VARCHAR(20) | | 光圈值 |
| exif_shutter_speed | VARCHAR(20) | | 快门速度 |
| exif_iso | INT | | ISO 感光度 |
| exif_exposure_mode | VARCHAR(50) | | 曝光模式 |
| exif_exposure_program | VARCHAR(50) | | 曝光程序 |
| exif_metering_mode | VARCHAR(50) | | 测光模式 |
| exif_white_balance | VARCHAR(50) | | 白平衡 |
| exif_flash | VARCHAR(50) | | 闪光灯状态 |
| exif_exposure_bias | VARCHAR(20) | | 曝光补偿 |
| **EXIF 时间地点** |
| exif_taken_at | TIMESTAMP | | 拍摄时间 |
| exif_gps_latitude | DECIMAL(10,7) | | GPS 纬度 |
| exif_gps_longitude | DECIMAL(10,7) | | GPS 经度 |
| exif_gps_altitude | DECIMAL(10,2) | | GPS 海拔 |
| **EXIF 图像信息** |
| exif_image_width | INT | | 图像宽度 |
| exif_image_height | INT | | 图像高度 |
| exif_orientation | INT | | 方向 |
| exif_color_space | VARCHAR(50) | | 色彩空间 |
| exif_software | VARCHAR(100) | | 处理软件 |
| **时间戳** |
| approved_at | TIMESTAMP | | 审核通过时间 |
| created_at | TIMESTAMP | NOT NULL DEFAULT NOW() | 创建时间 |
| updated_at | TIMESTAMP | NOT NULL DEFAULT NOW() | 更新时间 |

**状态值：**
- `pending` - 待审核
- `ai_passed` - AI 初审通过
- `ai_rejected` - AI 初审拒绝
- `approved` - 已发布
- `rejected` - 已拒绝

**索引：**
- `idx_photos_user_id` ON user_id
- `idx_photos_category_id` ON category_id
- `idx_photos_status` ON status
- `idx_photos_aircraft_type` ON aircraft_type
- `idx_photos_airline` ON airline
- `idx_photos_registration` ON registration
- `idx_photos_airport` ON airport
- `idx_photos_created_at` ON created_at DESC
- `idx_photos_exif_taken_at` ON exif_taken_at

---

### 4. tags - 标签表

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | SERIAL | PRIMARY KEY | 标签 ID |
| name | VARCHAR(50) | UNIQUE NOT NULL | 标签名称 |
| photo_count | INT | NOT NULL DEFAULT 0 | 关联照片数 |
| created_at | TIMESTAMP | NOT NULL DEFAULT NOW() | 创建时间 |

**索引：**
- `idx_tags_name` ON name
- `idx_tags_photo_count` ON photo_count DESC

---

### 5. photo_tags - 照片标签关联表

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| photo_id | BIGINT | NOT NULL REFERENCES photos(id) ON DELETE CASCADE | 照片 ID |
| tag_id | INT | NOT NULL REFERENCES tags(id) ON DELETE CASCADE | 标签 ID |
| created_at | TIMESTAMP | NOT NULL DEFAULT NOW() | 创建时间 |

**约束：**
- PRIMARY KEY (photo_id, tag_id)

**索引：**
- `idx_photo_tags_tag_id` ON tag_id

---

### 6. favorites - 收藏表

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| user_id | BIGINT | NOT NULL REFERENCES users(id) ON DELETE CASCADE | 用户 ID |
| photo_id | BIGINT | NOT NULL REFERENCES photos(id) ON DELETE CASCADE | 照片 ID |
| created_at | TIMESTAMP | NOT NULL DEFAULT NOW() | 收藏时间 |

**约束：**
- PRIMARY KEY (user_id, photo_id)

**索引：**
- `idx_favorites_photo_id` ON photo_id
- `idx_favorites_created_at` ON created_at DESC

---

### 7. photo_reviews - 照片审核记录表

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGSERIAL | PRIMARY KEY | 审核记录 ID |
| photo_id | BIGINT | NOT NULL REFERENCES photos(id) ON DELETE CASCADE | 照片 ID |
| reviewer_id | BIGINT | REFERENCES users(id) | 审核员 ID (AI 审核为 NULL) |
| review_type | VARCHAR(20) | NOT NULL | 审核类型: ai/manual |
| action | VARCHAR(20) | NOT NULL | 操作: approve/reject |
| reason | TEXT | | 拒绝原因 |
| ai_result | JSONB | | AI 审核详细结果 |
| created_at | TIMESTAMP | NOT NULL DEFAULT NOW() | 审核时间 |

**索引：**
- `idx_photo_reviews_photo_id` ON photo_id
- `idx_photo_reviews_reviewer_id` ON reviewer_id
- `idx_photo_reviews_created_at` ON created_at DESC

**ai_result JSON 结构：**
```json
{
  "score": 0.85,
  "aircraft_detected": true,
  "quality_score": 0.9,
  "issues": [],
  "aircraft_type": "Boeing 787-9",
  "registration": "B-1234"
}
```

---

### 8. tickets - 工单表

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGSERIAL | PRIMARY KEY | 工单 ID |
| user_id | BIGINT | NOT NULL REFERENCES users(id) | 创建者 ID |
| photo_id | BIGINT | REFERENCES photos(id) ON DELETE SET NULL | 关联照片 ID |
| type | VARCHAR(20) | NOT NULL | 类型: appeal/report/other |
| title | VARCHAR(200) | NOT NULL | 标题 |
| content | TEXT | NOT NULL | 内容 |
| status | VARCHAR(20) | NOT NULL DEFAULT 'open' | 状态 |
| created_at | TIMESTAMP | NOT NULL DEFAULT NOW() | 创建时间 |
| updated_at | TIMESTAMP | NOT NULL DEFAULT NOW() | 更新时间 |

**状态值：**
- `open` - 待处理
- `processing` - 处理中
- `resolved` - 已解决
- `closed` - 已关闭

**索引：**
- `idx_tickets_user_id` ON user_id
- `idx_tickets_photo_id` ON photo_id
- `idx_tickets_status` ON status
- `idx_tickets_type` ON type
- `idx_tickets_created_at` ON created_at DESC

---

### 9. ticket_replies - 工单回复表

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGSERIAL | PRIMARY KEY | 回复 ID |
| ticket_id | BIGINT | NOT NULL REFERENCES tickets(id) ON DELETE CASCADE | 工单 ID |
| user_id | BIGINT | NOT NULL REFERENCES users(id) | 回复者 ID |
| content | TEXT | NOT NULL | 回复内容 |
| created_at | TIMESTAMP | NOT NULL DEFAULT NOW() | 创建时间 |

**索引：**
- `idx_ticket_replies_ticket_id` ON ticket_id
- `idx_ticket_replies_created_at` ON created_at

---

### 10. refresh_tokens - 刷新令牌表

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGSERIAL | PRIMARY KEY | 令牌 ID |
| user_id | BIGINT | NOT NULL REFERENCES users(id) ON DELETE CASCADE | 用户 ID |
| token_hash | VARCHAR(255) | UNIQUE NOT NULL | 令牌哈希 |
| expires_at | TIMESTAMP | NOT NULL | 过期时间 |
| created_at | TIMESTAMP | NOT NULL DEFAULT NOW() | 创建时间 |

**索引：**
- `idx_refresh_tokens_user_id` ON user_id
- `idx_refresh_tokens_expires_at` ON expires_at

---

## 触发器

### 更新 updated_at 字段

```sql
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';
```

应用到以下表：
- users
- categories
- photos
- tickets

### 更新标签计数

```sql
-- 增加标签计数
CREATE OR REPLACE FUNCTION increment_tag_count()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE tags SET photo_count = photo_count + 1 WHERE id = NEW.tag_id;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- 减少标签计数
CREATE OR REPLACE FUNCTION decrement_tag_count()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE tags SET photo_count = photo_count - 1 WHERE id = OLD.tag_id;
    RETURN OLD;
END;
$$ language 'plpgsql';
```

### 更新收藏计数

```sql
-- 增加收藏计数
CREATE OR REPLACE FUNCTION increment_favorite_count()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE photos SET favorite_count = favorite_count + 1 WHERE id = NEW.photo_id;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- 减少收藏计数
CREATE OR REPLACE FUNCTION decrement_favorite_count()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE photos SET favorite_count = favorite_count - 1 WHERE id = OLD.photo_id;
    RETURN OLD;
END;
$$ language 'plpgsql';
```

---

## 初始数据

### 默认分类

| ID | name | name_en |
|----|------|---------|
| 1 | 民航客机 | Commercial Aircraft |
| 2 | 货运飞机 | Cargo Aircraft |
| 3 | 公务机 | Business Jet |
| 4 | 通用航空 | General Aviation |
| 5 | 军用飞机 | Military Aircraft |
| 6 | 直升机 | Helicopter |
| 7 | 历史飞机 | Historic Aircraft |
| 8 | 机场设施 | Airport Facilities |

### 默认管理员

```sql
INSERT INTO users (username, email, password_hash, role)
VALUES ('admin', 'admin@quanphotos.com', '<bcrypt_hash>', 'admin');
```

---

## 性能优化建议

1. **分区表**：photos 表可按 created_at 进行时间分区
2. **只读副本**：对读多写少的查询使用只读副本
3. **连接池**：使用 PgBouncer 进行连接池管理
4. **缓存**：热点数据使用 Redis 缓存
5. **全文搜索**：对 title、description 建立 GIN 索引

---

*文档版本：v1.0*
*创建日期：2025-12-22*
