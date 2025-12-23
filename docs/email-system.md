# QuanPhotos 邮件系统设计文档

## 概述

邮件系统用于用户通知、账户验证等场景，支持多种邮件模板和异步发送。

---

## 功能需求

### 1. 账户相关

| 功能 | 触发场景 | 优先级 |
|------|----------|--------|
| 注册验证 | 用户注册后发送验证链接 | P0 |
| 找回密码 | 用户请求重置密码时发送重置链接 | P0 |
| 邮箱变更 | 用户修改邮箱时发送验证 | P1 |
| 登录提醒 | 异地登录或新设备登录提醒（可选） | P2 |

### 2. 照片审核相关

| 功能 | 触发场景 | 优先级 |
|------|----------|--------|
| 审核通过 | 照片审核通过后通知用户 | P0 |
| 审核拒绝 | 照片被拒绝后通知用户（含原因） | P0 |
| 入选精选 | 照片被选为精选时通知用户 | P1 |

### 3. 社交互动相关

| 功能 | 触发场景 | 优先级 |
|------|----------|--------|
| 评论通知 | 照片收到新评论时通知作者 | P1 |
| 回复通知 | 评论被回复时通知评论者 | P1 |
| 点赞汇总 | 每日/每周点赞数汇总（可选） | P2 |
| 收藏通知 | 照片被收藏时通知（可选） | P2 |

### 4. 工单相关

| 功能 | 触发场景 | 优先级 |
|------|----------|--------|
| 工单创建 | 用户创建工单后确认邮件 | P1 |
| 工单回复 | 工单有新回复时通知 | P1 |
| 工单状态变更 | 工单状态变更时通知 | P1 |

### 5. 系统通知

| 功能 | 触发场景 | 优先级 |
|------|----------|--------|
| 私信通知 | 收到新私信时通知（可配置） | P2 |
| 系统公告 | 重要系统公告邮件推送 | P2 |
| 账号状态 | 账号被封禁/解封时通知 | P1 |

---

## 数据库设计

### email_verification_tokens - 邮箱验证令牌表

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGSERIAL | PRIMARY KEY | 令牌 ID |
| user_id | BIGINT | NOT NULL REFERENCES users(id) | 用户 ID |
| token | VARCHAR(255) | UNIQUE NOT NULL | 验证令牌 |
| type | VARCHAR(20) | NOT NULL | 类型: register/reset_password/change_email |
| new_email | VARCHAR(255) | | 新邮箱（仅用于邮箱变更） |
| expires_at | TIMESTAMP | NOT NULL | 过期时间 |
| used_at | TIMESTAMP | | 使用时间 |
| created_at | TIMESTAMP | NOT NULL DEFAULT NOW() | 创建时间 |

**索引：**
- `idx_email_verification_tokens_token` ON token
- `idx_email_verification_tokens_user_id` ON user_id
- `idx_email_verification_tokens_expires_at` ON expires_at

### email_logs - 邮件发送日志表

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGSERIAL | PRIMARY KEY | 日志 ID |
| user_id | BIGINT | REFERENCES users(id) | 接收用户 ID（系统邮件可为空） |
| to_email | VARCHAR(255) | NOT NULL | 接收邮箱 |
| template | VARCHAR(50) | NOT NULL | 邮件模板名称 |
| subject | VARCHAR(255) | NOT NULL | 邮件主题 |
| status | VARCHAR(20) | NOT NULL DEFAULT 'pending' | 状态: pending/sent/failed |
| error_message | TEXT | | 发送失败原因 |
| sent_at | TIMESTAMP | | 发送时间 |
| created_at | TIMESTAMP | NOT NULL DEFAULT NOW() | 创建时间 |

**索引：**
- `idx_email_logs_user_id` ON user_id
- `idx_email_logs_status` ON status
- `idx_email_logs_created_at` ON created_at DESC

### user_email_preferences - 用户邮件偏好设置表

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| user_id | BIGINT | PRIMARY KEY REFERENCES users(id) | 用户 ID |
| review_result | BOOLEAN | NOT NULL DEFAULT TRUE | 接收审核结果通知 |
| comment_notification | BOOLEAN | NOT NULL DEFAULT TRUE | 接收评论通知 |
| reply_notification | BOOLEAN | NOT NULL DEFAULT TRUE | 接收回复通知 |
| like_notification | BOOLEAN | NOT NULL DEFAULT FALSE | 接收点赞通知 |
| favorite_notification | BOOLEAN | NOT NULL DEFAULT FALSE | 接收收藏通知 |
| message_notification | BOOLEAN | NOT NULL DEFAULT TRUE | 接收私信通知 |
| ticket_notification | BOOLEAN | NOT NULL DEFAULT TRUE | 接收工单通知 |
| system_notification | BOOLEAN | NOT NULL DEFAULT TRUE | 接收系统通知 |
| weekly_digest | BOOLEAN | NOT NULL DEFAULT FALSE | 接收每周摘要 |
| updated_at | TIMESTAMP | NOT NULL DEFAULT NOW() | 更新时间 |

---

## 用户表扩展

在 `users` 表中添加以下字段：

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| email_verified | BOOLEAN | NOT NULL DEFAULT FALSE | 邮箱是否已验证 |
| email_verified_at | TIMESTAMP | | 邮箱验证时间 |

---

## 邮件模板

### 模板列表

| 模板名称 | 说明 | 变量 |
|----------|------|------|
| register_verify | 注册验证邮件 | username, verify_url, expire_hours |
| reset_password | 密码重置邮件 | username, reset_url, expire_hours |
| change_email | 邮箱变更验证 | username, new_email, verify_url |
| photo_approved | 照片审核通过 | username, photo_title, photo_url |
| photo_rejected | 照片审核拒绝 | username, photo_title, reason |
| photo_featured | 照片入选精选 | username, photo_title, photo_url |
| new_comment | 新评论通知 | username, photo_title, commenter, comment_content, photo_url |
| comment_reply | 评论回复通知 | username, replier, reply_content, photo_url |
| ticket_created | 工单创建确认 | username, ticket_title, ticket_id |
| ticket_reply | 工单回复通知 | username, ticket_title, reply_content |
| ticket_status | 工单状态变更 | username, ticket_title, old_status, new_status |
| account_banned | 账号封禁通知 | username, reason |
| account_unbanned | 账号解封通知 | username |
| weekly_digest | 每周摘要 | username, stats, top_photos |

### 模板示例（注册验证）

```html
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>验证您的 QuanPhotos 账号</title>
</head>
<body style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto;">
    <div style="background: #1a73e8; padding: 20px; text-align: center;">
        <h1 style="color: white; margin: 0;">QuanPhotos</h1>
    </div>
    <div style="padding: 30px; background: #f9f9f9;">
        <h2>您好，{{.Username}}！</h2>
        <p>感谢您注册 QuanPhotos 航空摄影社区。</p>
        <p>请点击下方按钮验证您的邮箱地址：</p>
        <div style="text-align: center; margin: 30px 0;">
            <a href="{{.VerifyURL}}" style="background: #1a73e8; color: white; padding: 12px 30px; text-decoration: none; border-radius: 4px;">验证邮箱</a>
        </div>
        <p style="color: #666; font-size: 14px;">此链接将在 {{.ExpireHours}} 小时后过期。</p>
        <p style="color: #666; font-size: 14px;">如果您没有注册 QuanPhotos 账号，请忽略此邮件。</p>
    </div>
    <div style="padding: 20px; text-align: center; color: #999; font-size: 12px;">
        <p>© 2025 QuanPhotos. All rights reserved.</p>
    </div>
</body>
</html>
```

---

## 接口设计

### 邮箱验证相关

```
POST   /api/v1/auth/send-verification    发送验证邮件
POST   /api/v1/auth/verify-email         验证邮箱
POST   /api/v1/auth/forgot-password      发送密码重置邮件
POST   /api/v1/auth/reset-password       重置密码
```

### 用户邮件偏好

```
GET    /api/v1/users/me/email-preferences     获取邮件偏好设置
PUT    /api/v1/users/me/email-preferences     更新邮件偏好设置
```

---

## 接口详情

### POST /api/v1/auth/send-verification

发送邮箱验证邮件（需登录）

**请求：**
```json
{}
```

**响应：**
```json
{
  "code": 0,
  "message": "验证邮件已发送",
  "data": {
    "expires_in": 86400
  }
}
```

### POST /api/v1/auth/verify-email

验证邮箱

**请求：**
```json
{
  "token": "abc123..."
}
```

**响应：**
```json
{
  "code": 0,
  "message": "邮箱验证成功",
  "data": null
}
```

### POST /api/v1/auth/forgot-password

发送密码重置邮件

**请求：**
```json
{
  "email": "user@example.com"
}
```

**响应：**
```json
{
  "code": 0,
  "message": "如果该邮箱已注册，您将收到密码重置邮件",
  "data": null
}
```

### POST /api/v1/auth/reset-password

重置密码

**请求：**
```json
{
  "token": "abc123...",
  "password": "newpassword123"
}
```

**响应：**
```json
{
  "code": 0,
  "message": "密码重置成功",
  "data": null
}
```

---

## 服务架构

### 邮件服务接口

```go
type EmailService interface {
    // 发送验证邮件
    SendVerificationEmail(ctx context.Context, userID int64, email string) error

    // 发送密码重置邮件
    SendPasswordResetEmail(ctx context.Context, email string) error

    // 发送照片审核结果
    SendPhotoReviewResult(ctx context.Context, userID int64, photoID int64, approved bool, reason string) error

    // 发送评论通知
    SendCommentNotification(ctx context.Context, userID int64, photoID int64, commenterName, content string) error

    // 发送回复通知
    SendReplyNotification(ctx context.Context, userID int64, replierName, content string, photoID int64) error

    // 发送工单通知
    SendTicketNotification(ctx context.Context, userID int64, ticketID int64, event string) error

    // 发送通用邮件
    SendEmail(ctx context.Context, to, template string, data map[string]interface{}) error
}
```

### 邮件队列

对于高并发场景，建议使用消息队列异步发送邮件：

1. **同步发送**（当前阶段）
   - 直接调用 SMTP 发送
   - 简单场景适用

2. **异步发送**（后续优化）
   - 使用 Redis 队列或消息中间件
   - 后台 Worker 处理发送
   - 支持重试机制

---

## 配置项

```env
# SMTP 邮件配置
SMTP_ENABLED=true
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USER=noreply@quanphotos.com
SMTP_PASSWORD=your_password
SMTP_FROM=QuanPhotos <noreply@quanphotos.com>
SMTP_TLS=true

# 邮件验证配置
EMAIL_VERIFY_EXPIRE=24              # 验证链接过期时间（小时）
EMAIL_RESET_EXPIRE=1                # 密码重置链接过期时间（小时）
EMAIL_RATE_LIMIT=5                  # 每用户每小时发送限制

# 前端 URL（用于生成链接）
FRONTEND_URL=https://quanphotos.com
```

---

## 安全考虑

1. **令牌安全**
   - 使用加密安全的随机令牌（32字节）
   - 令牌单次使用后失效
   - 设置合理的过期时间

2. **发送限制**
   - 限制每用户每小时发送次数
   - 防止邮件轰炸

3. **信息泄露防护**
   - 忘记密码接口不透露邮箱是否存在
   - 统一返回成功消息

4. **链接安全**
   - 使用 HTTPS 链接
   - 令牌在 URL 中传递，日志中需脱敏

---

## 开发阶段

| 阶段 | 功能 | 优先级 |
|------|------|--------|
| 阶段一 | 邮箱验证、密码重置 | P0 |
| 阶段二 | 审核结果通知 | P0 |
| 阶段三 | 评论/回复通知 | P1 |
| 阶段四 | 工单通知 | P1 |
| 阶段五 | 用户偏好设置 | P1 |
| 阶段六 | 异步队列优化 | P2 |

---

*文档版本：v1.0*
*创建日期：2025-12-23*
