# QuanPhotos 后端开发 TODO

> 本文档根据 `docs/development-plan.md` 需求，对比现有代码实现状态生成。
>
> 优先级说明：P0（核心必须）、P1（重要）、P2（可选）

---

## 开发进度概览

| 阶段 | 内容 | 状态 |
|------|------|------|
| 第一阶段 | 基础架构搭建 | ✅ 完成 |
| 第二阶段 | 用户系统 | ✅ 完成 |
| 第三阶段 | 照片管理核心 | ✅ 完成 |
| 第四阶段 | AI 审核集成 | 🔴 未开始 |
| 第五阶段 | 工单系统 | ✅ 完成 |
| 第六阶段 | 管理后台接口 | ✅ 完成 |
| 第七阶段 | 功能增强 | ✅ 完成 |
| 第八阶段 | 社交互动功能 | ✅ 完成 |
| 第九阶段 | 私信与站内信 | ✅ 完成 |
| 第十阶段 | 安全与性能 | ✅ 完成（不含缓存）|
| 第十一阶段 | 部署与运维 | ✅ 基本完成 |

---

## 第三阶段：照片管理核心

> ✅ 已完成

### 存储服务 `internal/pkg/storage/`

- [x] **P0** 实现 `Storage` 接口定义
- [x] **P0** 实现 `LocalStorage` 本地文件存储
- [x] **P1** 文件上传到临时目录
- [x] **P1** 文件移动到正式目录（按日期组织）
- [x] **P1** 文件删除功能
- [x] **P2** 预留 OSS/S3 云存储接口

### EXIF 解析 `internal/pkg/exif/`

- [x] **P0** 集成 `goexif` 或 `go-exif` 库
- [x] **P0** 解析相机信息（Make, Model, SerialNumber）
- [x] **P0** 解析镜头信息（LensModel, FocalLength）
- [x] **P0** 解析拍摄参数（Aperture, ShutterSpeed, ISO）
- [x] **P1** 解析 GPS 坐标
- [x] **P1** 解析拍摄时间

### 图像处理

- [x] **P0** 集成 `imaging` 库
- [x] **P0** 图像方向校正（根据 EXIF Orientation）
- [x] **P1** 生成缩略图 (sm: 300x200, md: 800x533, lg: 1600x1067)
- [x] **P1** 原图尺寸压缩（最大 4096px）
- [ ] **P2** 渐进式 JPEG 输出

### 照片上传接口

- [x] **P0** `POST /api/v1/photos` 上传接口实现
- [x] **P0** 文件格式验证（JPG/PNG）
- [x] **P0** 文件大小限制（50MB）
- [ ] **P1** 上传频率限制
- [x] **P2** RAW 格式支持 (CR2, CR3, NEF, ARW, RAF, ORF, RW2, DNG)
- [ ] **P2** RAW + JPG 配对上传验证

---

## 第四阶段：AI 审核集成

### AI 客户端 `internal/pkg/ai/`

- [ ] **P0** HTTP 客户端封装
- [ ] **P0** 请求超时配置
- [ ] **P1** 失败重试机制
- [ ] **P1** 熔断降级处理

### 审核流程

- [ ] **P0** 上传后自动触发 AI 审核
- [ ] **P0** 审核状态流转（pending → ai_passed/ai_rejected）
- [ ] **P1** 审核结果存储到 `photo_reviews` 表
- [ ] **P2** 异步审核（后台任务）
- [ ] **P2** AI 服务回调接口

---

## 第五阶段：工单系统

> 数据库表 `tickets`, `ticket_replies` 已创建

### Repository 层 `internal/repository/postgresql/ticket/`

- [x] **P0** 创建工单
- [x] **P0** 获取工单列表（用户自己的）
- [x] **P0** 获取工单详情
- [x] **P1** 添加工单回复
- [x] **P1** 更新工单状态

### Service 层 `internal/service/ticket/`

- [x] **P0** 创建工单业务逻辑
- [x] **P0** 工单列表查询
- [x] **P1** 工单回复逻辑
- [x] **P1** 工单状态流转验证

### Handler 层 `internal/handler/ticket_handler.go`

- [x] **P0** `POST /api/v1/tickets` 创建工单
- [x] **P0** `GET /api/v1/tickets` 获取我的工单列表
- [x] **P0** `GET /api/v1/tickets/:id` 获取工单详情
- [x] **P1** `POST /api/v1/tickets/:id/replies` 回复工单

---

## 第六阶段：管理后台接口

> ✅ 已完成

### 照片审核

- [x] **P0** `GET /api/v1/admin/reviews` 待审核列表
- [x] **P0** `POST /api/v1/admin/reviews/:id` 人工审核操作
- [x] **P1** 审核通过后更新照片状态
- [x] **P1** 审核拒绝时记录原因

### 照片管理

- [x] **P1** `DELETE /api/v1/admin/photos/:id` 管理员删除照片
- [x] **P1** 删除时记录原因

### 工单管理

- [x] **P1** `GET /api/v1/admin/tickets` 工单列表（全部）
- [x] **P1** `PUT /api/v1/admin/tickets/:id` 处理工单

### 精选照片管理

- [x] **P1** `POST /api/v1/admin/featured` 设置精选
- [x] **P1** `DELETE /api/v1/admin/featured/:id` 取消精选

### 公告管理

- [x] **P1** `GET /api/v1/admin/announcements` 公告列表
- [x] **P1** `POST /api/v1/admin/announcements` 创建公告
- [x] **P1** `PUT /api/v1/admin/announcements/:id` 更新公告
- [x] **P1** `DELETE /api/v1/admin/announcements/:id` 删除公告
- [x] **P1** `GET /api/v1/admin/announcements/:id` 获取公告详情

---

## 第七阶段：功能增强

> ✅ 已完成

### 分类系统 `internal/handler/category_handler.go`

- [x] **P1** `GET /api/v1/categories` 获取分类列表
- [x] **P1** `GET /api/v1/categories/:id` 获取分类详情
- [x] **P1** `POST /api/v1/categories` 创建分类（管理员）
- [x] **P1** `PUT /api/v1/categories/:id` 更新分类（管理员）
- [x] **P1** `DELETE /api/v1/categories/:id` 删除分类（管理员）

### 标签系统 `internal/handler/tag_handler.go`

- [x] **P1** `GET /api/v1/tags` 获取热门标签
- [x] **P1** `GET /api/v1/tags/search` 搜索标签
- [x] **P1** `GET /api/v1/tags/:id/photos` 获取标签下的照片

### 搜索功能

- [x] **P1** 照片列表支持关键词搜索
- [x] **P1** 照片列表支持机型、航空公司、机场、注册号筛选
- [x] **P1** 照片列表支持拍摄日期范围筛选

### 国际化 `internal/pkg/i18n/`

- [x] **P1** 集成 i18n 库
- [x] **P1** 中文语言包 (zh-CN)
- [x] **P1** 英文语言包 (en-US)
- [x] **P1** 错误消息国际化

### 邮件系统 `internal/pkg/email/`

- [x] **P2** SMTP 邮件发送服务
- [x] **P2** 邮件模板管理
- [x] **P2** 注册验证邮件
- [x] **P2** 密码重置邮件
- [x] **P2** 审核结果通知邮件

---

## 第八阶段：社交互动功能

> ✅ 已完成

### 评论系统

- [x] **P0** `GET /api/v1/photos/:id/comments` 获取评论列表
- [x] **P0** `POST /api/v1/photos/:id/comments` 发表评论
- [x] **P0** `DELETE /api/v1/comments/:id` 删除评论
- [x] **P1** 评论回复功能（parent_id）
- [x] **P1** `POST /api/v1/comments/:id/like` 点赞评论

### 转发功能

- [x] **P1** `POST /api/v1/photos/:id/share` 转发照片

### 精选照片展示

- [x] **P1** `GET /api/v1/featured` 获取精选照片列表

### 排行榜

- [x] **P1** `GET /api/v1/rankings/photos` 热门照片排行
- [x] **P1** `GET /api/v1/rankings/users` 活跃用户排行
- [x] **P2** 支持日/周/月/全部时间维度

### 公告展示

- [x] **P1** `GET /api/v1/announcements` 公告列表（公开）
- [x] **P1** `GET /api/v1/announcements/:id` 公告详情（公开）

---

## 第九阶段：私信与站内信

> ✅ 已完成

### 私信系统

- [x] **P0** `GET /api/v1/conversations` 会话列表
- [x] **P0** `POST /api/v1/conversations` 创建会话/发送私信
- [x] **P0** `GET /api/v1/conversations/:id` 获取会话消息
- [x] **P0** `POST /api/v1/conversations/:id` 发送消息
- [x] **P1** `DELETE /api/v1/conversations/:id` 删除会话
- [x] **P1** 未读消息计数

### 站内通知系统

- [x] **P1** `GET /api/v1/notifications` 通知列表
- [x] **P1** `GET /api/v1/notifications/unread` 未读数量
- [x] **P1** `PUT /api/v1/notifications/:id/read` 标记已读
- [x] **P1** `PUT /api/v1/notifications/read-all` 全部标记已读

### 通知触发

- [x] **P1** 点赞时创建通知
- [x] **P1** 评论时创建通知
- [x] **P1** 回复时创建通知
- [x] **P1** 入选精选时创建通知
- [x] **P1** 审核结果通知

---

## 第十阶段：安全与性能

> ✅ 已完成（不含缓存）

### 限流中间件 `internal/middleware/rate_limit.go`

- [x] **P0** 全局请求限流
- [x] **P0** 登录接口限流（防暴力破解）
- [x] **P1** 上传接口限流（每用户每小时）

### 上传安全

- [x] **P0** MIME 类型验证（不仅检查扩展名）
- [x] **P0** 文件头魔数检查
- [x] **P1** 路径遍历防护

### Redis 缓存 `internal/pkg/cache/`

> ⏳ 暂不实现，后续可根据需要添加

- [ ] **P2** Redis 连接配置
- [ ] **P2** 热点照片缓存
- [ ] **P2** 用户信息缓存
- [ ] **P2** Access Token 黑名单（解决登出后 Token 仍有效问题）

### API 文档

- [ ] **P1** Swagger 注解完善
- [ ] **P1** `make swagger` 生成文档
- [ ] **P1** Swagger UI 集成

---

## 超级管理员功能

> ✅ 已完成

### 管理员权限管理

- [x] **P1** `GET /api/v1/superadmin/admins` 管理员列表
- [x] **P1** `POST /api/v1/superadmin/admins/:id/permissions` 授予权限
- [x] **P1** `DELETE /api/v1/superadmin/admins/:id/permissions` 撤销权限

### 审查员管理

- [x] **P1** `GET /api/v1/superadmin/reviewers` 审查员列表
- [x] **P1** `POST /api/v1/superadmin/reviewers/:id/categories` 授权分类
- [x] **P1** `DELETE /api/v1/superadmin/reviewers/:id/categories` 撤销分类

### 用户功能限制

- [x] **P1** `PUT /api/v1/superadmin/users/:id/restrictions` 禁用用户功能
  - can_comment, can_message, can_upload

---

## 数据库迁移待补充

> 以下表在设计文档中但可能需要验证是否完整

- [ ] `email_verification_tokens` 邮箱验证令牌表
- [ ] `email_logs` 邮件发送日志表
- [ ] `user_email_preferences` 用户邮件偏好表
- [ ] `users.email_verified` 邮箱验证状态字段

---

## 建议开发顺序

1. **第三阶段**（照片管理核心）- 这是产品核心功能
2. **第八阶段**（评论功能）- 社交互动基础
3. **第四阶段**（AI 审核）- 与 AI 服务对接
4. **第五阶段**（工单系统）- 申诉流程
5. **第六阶段**（管理后台）- 运营工具
6. **第九阶段**（私信通知）- 用户沟通
7. **第七阶段**（分类标签、搜索、i18n）- 体验优化
8. **第十阶段**（安全性能）- 生产就绪

---

*文档版本：v1.1*
*生成日期：2025-12-27*
