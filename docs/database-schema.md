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
│ role        │       │ status      │             │
│ ...         │       │ exif_*      │             │
└─────────────┘       │ ...         │             │
      │               └─────────────┘             │
      │                     │                     │
      │               ┌─────┴─────┬───────────────┤
      │               ▼           ▼               ▼
      │       ┌─────────────┐ ┌─────────────┐ ┌───────────────────┐
      │       │ photo_tags  │ │  favorites  │ │ reviewer_categories│
      │       ├─────────────┤ ├─────────────┤ ├───────────────────┤
      │       │ photo_id    │ │ user_id     │ │ reviewer_id       │
      │       │ tag_id      │ │ photo_id    │ │ category_id       │
      │       └─────────────┘ └─────────────┘ └───────────────────┘
      │               │
      │               ▼
      │       ┌─────────────┐       ┌─────────────┐
      │       │    tags     │       │ photo_likes │
      │       ├─────────────┤       ├─────────────┤
      │       │ id (PK)     │       │ user_id     │
      │       │ name        │       │ photo_id    │
      │       └─────────────┘       └─────────────┘
      │
      │       ┌─────────────────┐   ┌─────────────────┐
      │       │ photo_comments  │   │  comment_likes  │
      │       ├─────────────────┤   ├─────────────────┤
      │       │ id (PK)         │◄──│ comment_id      │
      │       │ photo_id (FK)   │   │ user_id         │
      │       │ user_id (FK)    │   └─────────────────┘
      │       │ parent_id (FK)  │
      │       │ content         │
      │       └─────────────────┘
      │
      │       ┌─────────────────┐   ┌─────────────────┐
      │       │  photo_shares   │   │ featured_photos │
      │       ├─────────────────┤   ├─────────────────┤
      │       │ id (PK)         │   │ id (PK)         │
      │       │ photo_id (FK)   │   │ photo_id (FK)   │
      │       │ user_id (FK)    │   │ admin_id (FK)   │
      │       └─────────────────┘   └─────────────────┘
      │
┌─────┴─────────┐       ┌─────────────────┐
│ conversations │       │    messages     │
├───────────────┤       ├─────────────────┤
│ id (PK)       │◄──────│ conversation_id │
│ user1_id (FK) │       │ sender_id (FK)  │
│ user2_id (FK) │       │ content         │
└───────────────┘       └─────────────────┘

┌─────────────────┐     ┌─────────────────┐
│  notifications  │     │  announcements  │
├─────────────────┤     ├─────────────────┤
│ id (PK)         │     │ id (PK)         │
│ user_id (FK)    │     │ author_id (FK)  │
│ actor_id (FK)   │     │ title           │
│ type            │     │ content         │
│ content         │     │ status          │
└─────────────────┘     └─────────────────┘

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
| role | VARCHAR(20) | NOT NULL DEFAULT 'user' | 角色: guest/user/reviewer/admin/superadmin |
| status | VARCHAR(20) | NOT NULL DEFAULT 'active' | 状态: active/banned |
| can_comment | BOOLEAN | NOT NULL DEFAULT TRUE | 是否可评论 |
| can_message | BOOLEAN | NOT NULL DEFAULT TRUE | 是否可私信 |
| can_upload | BOOLEAN | NOT NULL DEFAULT TRUE | 是否可上传 |
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
| like_count | INT | NOT NULL DEFAULT 0 | 点赞次数 |
| favorite_count | INT | NOT NULL DEFAULT 0 | 收藏次数 |
| comment_count | INT | NOT NULL DEFAULT 0 | 评论次数 |
| share_count | INT | NOT NULL DEFAULT 0 | 转发次数 |
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

### 11. photo_likes - 照片点赞表

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| user_id | BIGINT | NOT NULL REFERENCES users(id) ON DELETE CASCADE | 用户 ID |
| photo_id | BIGINT | NOT NULL REFERENCES photos(id) ON DELETE CASCADE | 照片 ID |
| created_at | TIMESTAMP | NOT NULL DEFAULT NOW() | 点赞时间 |

**约束：**
- PRIMARY KEY (user_id, photo_id)

**索引：**
- `idx_photo_likes_photo_id` ON photo_id
- `idx_photo_likes_created_at` ON created_at DESC

---

### 12. photo_comments - 照片评论表

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGSERIAL | PRIMARY KEY | 评论 ID |
| photo_id | BIGINT | NOT NULL REFERENCES photos(id) ON DELETE CASCADE | 照片 ID |
| user_id | BIGINT | NOT NULL REFERENCES users(id) ON DELETE CASCADE | 评论者 ID |
| parent_id | BIGINT | REFERENCES photo_comments(id) ON DELETE CASCADE | 父评论 ID（回复） |
| content | TEXT | NOT NULL | 评论内容 |
| like_count | INT | NOT NULL DEFAULT 0 | 点赞数 |
| reply_count | INT | NOT NULL DEFAULT 0 | 回复数 |
| status | VARCHAR(20) | NOT NULL DEFAULT 'visible' | 状态: visible/hidden/deleted |
| created_at | TIMESTAMP | NOT NULL DEFAULT NOW() | 创建时间 |
| updated_at | TIMESTAMP | NOT NULL DEFAULT NOW() | 更新时间 |

**索引：**
- `idx_photo_comments_photo_id` ON photo_id
- `idx_photo_comments_user_id` ON user_id
- `idx_photo_comments_parent_id` ON parent_id
- `idx_photo_comments_created_at` ON created_at DESC

---

### 13. comment_likes - 评论点赞表

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| user_id | BIGINT | NOT NULL REFERENCES users(id) ON DELETE CASCADE | 用户 ID |
| comment_id | BIGINT | NOT NULL REFERENCES photo_comments(id) ON DELETE CASCADE | 评论 ID |
| created_at | TIMESTAMP | NOT NULL DEFAULT NOW() | 点赞时间 |

**约束：**
- PRIMARY KEY (user_id, comment_id)

**索引：**
- `idx_comment_likes_comment_id` ON comment_id

---

### 14. photo_shares - 照片转发表

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGSERIAL | PRIMARY KEY | 转发 ID |
| photo_id | BIGINT | NOT NULL REFERENCES photos(id) ON DELETE CASCADE | 照片 ID |
| user_id | BIGINT | NOT NULL REFERENCES users(id) ON DELETE CASCADE | 转发者 ID |
| content | TEXT | | 转发说明 |
| share_type | VARCHAR(20) | NOT NULL DEFAULT 'internal' | 类型: internal/external |
| created_at | TIMESTAMP | NOT NULL DEFAULT NOW() | 转发时间 |

**索引：**
- `idx_photo_shares_photo_id` ON photo_id
- `idx_photo_shares_user_id` ON user_id
- `idx_photo_shares_created_at` ON created_at DESC

---

### 15. featured_photos - 精选照片表

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGSERIAL | PRIMARY KEY | 精选 ID |
| photo_id | BIGINT | UNIQUE NOT NULL REFERENCES photos(id) ON DELETE CASCADE | 照片 ID |
| admin_id | BIGINT | NOT NULL REFERENCES users(id) | 操作管理员 ID |
| reason | VARCHAR(500) | | 入选原因 |
| sort_order | INT | NOT NULL DEFAULT 0 | 排序顺序 |
| featured_at | TIMESTAMP | NOT NULL DEFAULT NOW() | 入选时间 |
| expires_at | TIMESTAMP | | 过期时间（可选） |

**索引：**
- `idx_featured_photos_sort_order` ON sort_order
- `idx_featured_photos_featured_at` ON featured_at DESC

---

### 16. conversations - 私信会话表

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGSERIAL | PRIMARY KEY | 会话 ID |
| user1_id | BIGINT | NOT NULL REFERENCES users(id) ON DELETE CASCADE | 用户 1 ID |
| user2_id | BIGINT | NOT NULL REFERENCES users(id) ON DELETE CASCADE | 用户 2 ID |
| last_message_id | BIGINT | | 最后一条消息 ID |
| user1_unread | INT | NOT NULL DEFAULT 0 | 用户 1 未读数 |
| user2_unread | INT | NOT NULL DEFAULT 0 | 用户 2 未读数 |
| user1_deleted | BOOLEAN | NOT NULL DEFAULT FALSE | 用户 1 是否删除 |
| user2_deleted | BOOLEAN | NOT NULL DEFAULT FALSE | 用户 2 是否删除 |
| created_at | TIMESTAMP | NOT NULL DEFAULT NOW() | 创建时间 |
| updated_at | TIMESTAMP | NOT NULL DEFAULT NOW() | 更新时间 |

**约束：**
- UNIQUE (user1_id, user2_id) WHERE user1_id < user2_id

**索引：**
- `idx_conversations_user1_id` ON user1_id
- `idx_conversations_user2_id` ON user2_id
- `idx_conversations_updated_at` ON updated_at DESC

---

### 17. messages - 私信消息表

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGSERIAL | PRIMARY KEY | 消息 ID |
| conversation_id | BIGINT | NOT NULL REFERENCES conversations(id) ON DELETE CASCADE | 会话 ID |
| sender_id | BIGINT | NOT NULL REFERENCES users(id) ON DELETE CASCADE | 发送者 ID |
| content | TEXT | NOT NULL | 消息内容 |
| is_read | BOOLEAN | NOT NULL DEFAULT FALSE | 是否已读 |
| created_at | TIMESTAMP | NOT NULL DEFAULT NOW() | 发送时间 |

**索引：**
- `idx_messages_conversation_id` ON conversation_id
- `idx_messages_sender_id` ON sender_id
- `idx_messages_created_at` ON created_at DESC

---

### 18. notifications - 站内通知表

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGSERIAL | PRIMARY KEY | 通知 ID |
| user_id | BIGINT | NOT NULL REFERENCES users(id) ON DELETE CASCADE | 接收者 ID |
| actor_id | BIGINT | REFERENCES users(id) ON DELETE SET NULL | 触发者 ID |
| type | VARCHAR(20) | NOT NULL | 通知类型 |
| title | VARCHAR(200) | NOT NULL | 通知标题 |
| content | TEXT | | 通知内容 |
| related_photo_id | BIGINT | REFERENCES photos(id) ON DELETE SET NULL | 关联照片 ID |
| related_comment_id | BIGINT | REFERENCES photo_comments(id) ON DELETE SET NULL | 关联评论 ID |
| is_read | BOOLEAN | NOT NULL DEFAULT FALSE | 是否已读 |
| created_at | TIMESTAMP | NOT NULL DEFAULT NOW() | 创建时间 |

**通知类型：**
- `like` - 点赞通知
- `comment` - 评论通知
- `reply` - 回复通知
- `follow` - 关注通知
- `share` - 转发通知
- `featured` - 入选精选通知
- `review` - 审核结果通知
- `system` - 系统通知
- `message` - 私信通知

**索引：**
- `idx_notifications_user_id` ON user_id
- `idx_notifications_type` ON type
- `idx_notifications_is_read` ON is_read
- `idx_notifications_created_at` ON created_at DESC

---

### 19. announcements - 系统公告表

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGSERIAL | PRIMARY KEY | 公告 ID |
| author_id | BIGINT | NOT NULL REFERENCES users(id) | 作者 ID（管理员）|
| title | VARCHAR(200) | NOT NULL | 公告标题 |
| summary | VARCHAR(500) | | 摘要 |
| content | TEXT | NOT NULL | 公告内容（Markdown）|
| status | VARCHAR(20) | NOT NULL DEFAULT 'draft' | 状态: draft/published |
| is_pinned | BOOLEAN | NOT NULL DEFAULT FALSE | 是否置顶 |
| published_at | TIMESTAMP | | 发布时间 |
| created_at | TIMESTAMP | NOT NULL DEFAULT NOW() | 创建时间 |
| updated_at | TIMESTAMP | NOT NULL DEFAULT NOW() | 更新时间 |

**索引：**
- `idx_announcements_status` ON status
- `idx_announcements_is_pinned` ON is_pinned
- `idx_announcements_published_at` ON published_at DESC

---

### 20. reviewer_categories - 审查员分类权限表

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| reviewer_id | BIGINT | NOT NULL REFERENCES users(id) ON DELETE CASCADE | 审查员用户 ID |
| category_id | INT | NOT NULL REFERENCES categories(id) ON DELETE CASCADE | 分类 ID |
| created_at | TIMESTAMP | NOT NULL DEFAULT NOW() | 授权时间 |

**约束：**
- PRIMARY KEY (reviewer_id, category_id)

**索引：**
- `idx_reviewer_categories_category_id` ON category_id

**说明：**
- 审查员（role='reviewer'）只能审核其被授权的分类下的照片
- 管理员（role='admin'）可以审核所有分类的照片
- 如果审查员没有任何分类授权，则无法审核任何照片

---

### 21. admin_permissions - 管理员权限表

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| admin_id | BIGINT | NOT NULL REFERENCES users(id) ON DELETE CASCADE | 管理员用户 ID |
| permission | VARCHAR(50) | NOT NULL | 权限标识 |
| granted_by | BIGINT | NOT NULL REFERENCES users(id) | 授权者 ID（superadmin）|
| created_at | TIMESTAMP | NOT NULL DEFAULT NOW() | 授权时间 |

**约束：**
- PRIMARY KEY (admin_id, permission)

**索引：**
- `idx_admin_permissions_permission` ON permission

**权限标识列表：**

| 权限标识 | 说明 |
|----------|------|
| `manage_announcements` | 管理公告（创建、编辑、删除）|
| `manage_featured` | 管理精选照片 |
| `ban_users` | 禁用用户账号 |
| `mute_comment` | 禁止用户评论 |
| `mute_message` | 禁止用户私信 |
| `mute_upload` | 禁止用户上传 |
| `review_photos` | 审核照片 |
| `delete_photos` | 删除违规照片 |
| `delete_comments` | 删除违规评论 |
| `manage_tickets` | 处理工单 |
| `manage_categories` | 管理分类 |
| `manage_tags` | 管理标签 |
| `view_statistics` | 查看统计数据 |
| `view_user_details` | 查看用户详细信息（含邮箱等）|

**说明：**
- 只有 `role='admin'` 的用户可以被分配权限
- `role='superadmin'` 自动拥有所有权限，无需在此表中记录
- 管理员的权限由 superadmin 授予和撤销
- 没有任何权限的 admin 只能访问后台但无法执行任何操作

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
- photo_comments
- conversations
- announcements

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

### 更新点赞计数

```sql
-- 增加照片点赞计数
CREATE OR REPLACE FUNCTION increment_photo_like_count()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE photos SET like_count = like_count + 1 WHERE id = NEW.photo_id;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- 减少照片点赞计数
CREATE OR REPLACE FUNCTION decrement_photo_like_count()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE photos SET like_count = like_count - 1 WHERE id = OLD.photo_id;
    RETURN OLD;
END;
$$ language 'plpgsql';

-- 增加评论点赞计数
CREATE OR REPLACE FUNCTION increment_comment_like_count()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE photo_comments SET like_count = like_count + 1 WHERE id = NEW.comment_id;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- 减少评论点赞计数
CREATE OR REPLACE FUNCTION decrement_comment_like_count()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE photo_comments SET like_count = like_count - 1 WHERE id = OLD.comment_id;
    RETURN OLD;
END;
$$ language 'plpgsql';
```

### 更新评论计数

```sql
-- 增加照片评论计数
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

-- 减少照片评论计数
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
```

### 更新转发计数

```sql
-- 增加转发计数
CREATE OR REPLACE FUNCTION increment_share_count()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE photos SET share_count = share_count + 1 WHERE id = NEW.photo_id;
    RETURN NEW;
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
-- 超级管理员（系统初始化时创建）
INSERT INTO users (username, email, password_hash, role)
VALUES ('superadmin', 'admin@quanphotos.com', '<bcrypt_hash>', 'superadmin');
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
