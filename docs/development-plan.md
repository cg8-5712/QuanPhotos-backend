# QuanPhotos 后端开发计划

## 项目背景

QuanPhotos 是一个专业的航空摄影社区平台，为航空爱好者提供照片分享、浏览和交流的空间。本文档描述后端 API 服务的开发计划。

---

## 开发阶段

### 第一阶段：基础架构搭建

**目标**：完成项目骨架和核心基础设施

| 任务 | 描述 | 优先级 |
|------|------|--------|
| 项目初始化 | 创建项目结构、go.mod、基础目录 | P0 |
| 配置管理 | 实现配置加载（godotenv）、配置结构体定义 | P0 |
| 日志系统 | 集成 zap 结构化日志 | P0 |
| 数据库连接 | PostgreSQL 连接池配置、sqlx 集成 | P0 |
| HTTP 框架 | Gin 框架搭建、路由注册 | P0 |
| 统一响应 | 定义标准响应格式和错误码 | P0 |

**交付物**：
- 可运行的空白服务
- 配置文件模板 `.env.example`
- 基础中间件（日志、恢复）

---

### 第二阶段：用户系统

**目标**：完成用户注册、登录、权限管理

| 任务 | 描述 | 优先级 |
|------|------|--------|
| 用户模型 | users 表设计、模型定义 | P0 |
| 数据库迁移 | 用户表迁移文件 | P0 |
| 密码加密 | bcrypt 密码哈希 | P0 |
| JWT 认证 | Access Token / Refresh Token 实现 | P0 |
| 注册接口 | POST /api/v1/auth/register | P0 |
| 登录接口 | POST /api/v1/auth/login | P0 |
| Token 刷新 | POST /api/v1/auth/refresh | P1 |
| 用户信息 | GET /api/v1/users/me | P1 |
| 角色权限 | 游客/用户/管理员/超管角色定义 | P1 |
| 权限中间件 | 基于角色的访问控制 | P1 |

**交付物**：
- 完整的用户认证流程
- 角色权限体系
- 用户相关 API

---

### 第三阶段：照片管理核心

**目标**：实现照片上传、存储、浏览基础功能

| 任务 | 描述 | 优先级 |
|------|------|--------|
| 照片模型 | photos 表设计（含 EXIF 字段） | P0 |
| 存储服务 | 本地文件存储实现 | P0 |
| 图片上传 | POST /api/v1/photos 上传接口 | P0 |
| EXIF 解析 | 集成 EXIF 解析库、提取元数据 | P0 |
| 缩略图生成 | 上传时自动生成缩略图 | P1 |
| 照片列表 | GET /api/v1/photos 分页查询 | P0 |
| 照片详情 | GET /api/v1/photos/:id | P0 |
| 照片删除 | DELETE /api/v1/photos/:id | P1 |
| RAW 格式支持 | CR2/CR3/NEF/ARW 等格式处理 | P1 |
| RAW-JPG 关联 | 同一照片 RAW 与 JPG 对应关系 | P2 |

**交付物**：
- 照片 CRUD 接口
- EXIF 解析功能
- 文件存储服务

---

### 第四阶段：AI 审核集成

**目标**：对接 AI 微服务实现照片自动初审

| 任务 | 描述 | 优先级 |
|------|------|--------|
| AI 客户端 | HTTP 客户端封装、重试机制 | P0 |
| 审核流程 | 上传后自动触发 AI 审核 | P0 |
| 审核状态 | pending/ai_passed/ai_rejected/approved/rejected | P0 |
| 审核结果存储 | 存储 AI 返回的审核详情 | P1 |
| 异步审核 | 后台任务处理（可选） | P2 |
| 审核回调 | AI 服务回调接口（可选） | P2 |

**交付物**：
- AI 服务集成
- 完整的审核工作流

---

### 第五阶段：工单系统

**目标**：用户可对 AI 审核结果提出申诉

| 任务 | 描述 | 优先级 |
|------|------|--------|
| 工单模型 | tickets 表设计 | P0 |
| 创建工单 | POST /api/v1/tickets | P0 |
| 工单列表 | GET /api/v1/tickets（用户查看自己的） | P0 |
| 工单详情 | GET /api/v1/tickets/:id | P0 |
| 工单回复 | 用户/管理员回复功能 | P1 |
| 工单状态流转 | open/processing/resolved/closed | P1 |

**交付物**：
- 工单 CRUD 接口
- 工单与照片关联

---

### 第六阶段：管理后台接口

**目标**：为管理员提供审核、用户管理功能

| 任务 | 描述 | 优先级 |
|------|------|--------|
| 待审核列表 | GET /api/v1/admin/reviews | P0 |
| 人工审核 | POST /api/v1/admin/reviews/:id | P0 |
| 用户管理 | 用户列表、封禁、角色调整 | P1 |
| 照片管理 | 删除违规照片 | P1 |
| 工单处理 | 管理员工单处理接口 | P1 |
| 数据统计 | 基础统计接口（可选） | P2 |

**交付物**：
- 管理后台全部接口

---

### 第七阶段：功能增强

**目标**：完善用户体验相关功能

| 任务 | 描述 | 优先级 |
|------|------|--------|
| 收藏功能 | 用户收藏照片 | P1 |
| 分类/标签 | 照片分类和标签系统 | P1 |
| 搜索功能 | 按机型、机场、日期等搜索 | P1 |
| 国际化 | i18n 多语言支持（中/英） | P1 |
| 邮件通知 | 审核结果通知 | P2 |

**交付物**：
- 收藏、搜索、分类功能
- 多语言支持

---

### 第八阶段：社交互动功能

**目标**：实现用户互动和社区功能

| 任务 | 描述 | 优先级 |
|------|------|--------|
| 点赞功能 | 用户对照片点赞，记录点赞数 | P0 |
| 评论功能 | 用户对照片发表评论，支持回复 | P0 |
| 转发功能 | 用户转发照片到个人主页或站外 | P1 |
| 精选功能 | 管理员精选优质照片，首页展示 | P1 |
| 排行榜 | 热门照片、活跃用户排行榜 | P1 |
| 公告系统 | 管理员发布站内公告 | P1 |

**交付物**：
- 点赞、评论、转发接口
- 精选照片管理
- 排行榜系统

---

### 第九阶段：私信与站内信

**目标**：实现用户间通讯功能

| 任务 | 描述 | 优先级 |
|------|------|--------|
| 私信模型 | 私信表设计、会话管理 | P0 |
| 发送私信 | POST /api/v1/messages 发送私信 | P0 |
| 私信列表 | GET /api/v1/messages 获取私信会话列表 | P0 |
| 私信详情 | GET /api/v1/messages/:id 获取会话消息 | P0 |
| 站内信系统 | 系统通知、审核通知、互动通知 | P1 |
| 未读计数 | 未读消息数统计 | P1 |
| 消息设置 | 用户消息通知偏好设置 | P2 |

**交付物**：
- 私信系统
- 站内通知系统

---

### 第十阶段：安全与性能

**目标**：生产环境准备

| 任务 | 描述 | 优先级 |
|------|------|--------|
| CORS 配置 | 跨域安全配置 | P0 |
| 限流 | 请求频率限制 | P0 |
| 上传限制 | 文件大小、类型、频率限制 | P0 |
| 输入校验 | 请求参数验证 | P0 |
| SQL 注入防护 | 参数化查询检查 | P0 |
| Redis 缓存 | 热点数据缓存（可选） | P2 |
| 接口文档 | Swagger/OpenAPI 生成 | P1 |

**交付物**：
- 安全加固
- API 文档

---

### 第十一阶段：部署与运维

**目标**：实现容器化部署

| 任务 | 描述 | 优先级 |
|------|------|--------|
| Dockerfile | 多阶段构建镜像 | P0 |
| docker-compose | 本地开发编排 | P0 |
| 健康检查 | /health 接口 | P0 |
| Makefile | 常用命令封装 | P1 |
| CI/CD | GitHub Actions（可选） | P2 |
| 生产部署文档 | 部署指南 | P1 |

**交付物**：
- Docker 部署方案
- 运维文档

---

## 数据库设计概要

### 核心表

```
users              - 用户表
photos             - 照片表（含 EXIF 字段）
photo_reviews      - 审核记录表
tickets            - 工单表
ticket_replies     - 工单回复表
favorites          - 收藏表
categories         - 分类表
tags               - 标签表
photo_tags         - 照片-标签关联表
photo_likes        - 照片点赞表
photo_comments     - 照片评论表
comment_likes      - 评论点赞表
photo_shares       - 照片转发表
featured_photos    - 精选照片表
conversations      - 私信会话表
messages           - 私信消息表
notifications      - 站内通知表
announcements      - 系统公告表
reviewer_categories - 审查员分类权限表
admin_permissions  - 管理员权限表
```

---

## 接口规划

### 认证相关
```
POST   /api/v1/auth/register     用户注册
POST   /api/v1/auth/login        用户登录
POST   /api/v1/auth/refresh      刷新 Token
POST   /api/v1/auth/logout       退出登录
```

### 用户相关
```
GET    /api/v1/users/me          当前用户信息
PUT    /api/v1/users/me          更新用户信息
GET    /api/v1/users/:id         用户公开信息
GET    /api/v1/users/:id/photos  用户照片列表
```

### 照片相关
```
GET    /api/v1/photos            照片列表（分页、筛选）
POST   /api/v1/photos            上传照片
GET    /api/v1/photos/:id        照片详情
DELETE /api/v1/photos/:id        删除照片
POST   /api/v1/photos/:id/fav    收藏照片
DELETE /api/v1/photos/:id/fav    取消收藏
```

### 工单相关
```
GET    /api/v1/tickets           我的工单列表
POST   /api/v1/tickets           创建工单
GET    /api/v1/tickets/:id       工单详情
POST   /api/v1/tickets/:id/reply 回复工单
```

### 管理接口
```
GET    /api/v1/admin/reviews          待审核列表
POST   /api/v1/admin/reviews/:id      审核操作
GET    /api/v1/admin/users            用户管理
PUT    /api/v1/admin/users/:id/role   修改用户角色
DELETE /api/v1/admin/photos/:id       删除照片
GET    /api/v1/admin/tickets          工单管理
PUT    /api/v1/admin/tickets/:id      处理工单
POST   /api/v1/admin/featured         设置精选照片
DELETE /api/v1/admin/featured/:id     取消精选
GET    /api/v1/admin/announcements    公告管理
POST   /api/v1/admin/announcements    创建公告
PUT    /api/v1/admin/announcements/:id 更新公告
DELETE /api/v1/admin/announcements/:id 删除公告
```

### 超级管理员接口
```
GET    /api/v1/superadmin/admins                  管理员列表
POST   /api/v1/superadmin/admins/:id/permissions  授予管理员权限
DELETE /api/v1/superadmin/admins/:id/permissions  撤销管理员权限
GET    /api/v1/superadmin/reviewers               审查员列表
POST   /api/v1/superadmin/reviewers/:id/categories 授权审查员分类
DELETE /api/v1/superadmin/reviewers/:id/categories 撤销审查员分类
PUT    /api/v1/superadmin/users/:id/role          设置用户角色
PUT    /api/v1/superadmin/users/:id/restrictions  禁用用户功能
```

### 社交互动相关
```
POST   /api/v1/photos/:id/like        点赞照片
DELETE /api/v1/photos/:id/like        取消点赞
GET    /api/v1/photos/:id/comments    获取评论列表
POST   /api/v1/photos/:id/comments    发表评论
DELETE /api/v1/comments/:id           删除评论
POST   /api/v1/photos/:id/share       转发照片
GET    /api/v1/featured               获取精选照片
```

### 私信相关
```
GET    /api/v1/conversations          私信会话列表
POST   /api/v1/conversations          创建会话/发送私信
GET    /api/v1/conversations/:id      获取会话消息
POST   /api/v1/conversations/:id      发送消息
DELETE /api/v1/conversations/:id      删除会话
```

### 通知相关
```
GET    /api/v1/notifications          获取通知列表
PUT    /api/v1/notifications/:id/read 标记已读
PUT    /api/v1/notifications/read-all 全部标记已读
GET    /api/v1/notifications/unread   未读数量
```

### 公告相关
```
GET    /api/v1/announcements          获取公告列表
GET    /api/v1/announcements/:id      获取公告详情
```

### 排行榜相关
```
GET    /api/v1/rankings/photos        热门照片排行
GET    /api/v1/rankings/users         活跃用户排行
GET    /api/v1/rankings/weekly        周排行
GET    /api/v1/rankings/monthly       月排行
```

---

## 优先级说明

| 级别 | 说明 |
|------|------|
| P0 | 核心功能，必须实现 |
| P1 | 重要功能，应当实现 |
| P2 | 可选功能，视情况实现 |

---

## 技术要点

1. **分层架构**：Handler → Service → Repository，职责清晰
2. **SQL 优先**：使用 sqlx 原生 SQL，避免 ORM 黑箱
3. **错误处理**：统一错误码，错误信息可追溯
4. **配置外置**：环境变量配置，禁止硬编码
5. **安全第一**：输入校验、SQL 注入防护、XSS 防护

---

## 依赖项

```go
// 核心框架
github.com/gin-gonic/gin

// 数据库
github.com/jmoiron/sqlx
github.com/lib/pq

// 认证
github.com/golang-jwt/jwt/v5

// 配置
github.com/joho/godotenv

// 日志
go.uber.org/zap

// EXIF
github.com/rwcarlsen/goexif/exif

// 图像处理
github.com/disintegration/imaging

// 验证
github.com/go-playground/validator/v10

// API 文档
github.com/swaggo/swag
github.com/swaggo/gin-swagger
```

---

## 风险与注意事项

1. **RAW 格式处理**：部分 RAW 格式可能需要额外库支持
2. **AI 服务依赖**：需确保 AI 服务可用性，做好降级处理
3. **大文件上传**：注意内存占用，考虑流式处理
4. **并发上传**：需要限制单用户并发上传数量
5. **存储扩展**：预留 OSS/S3 存储接口

---

*文档版本：v1.0*
*创建日期：2025-12-22*
