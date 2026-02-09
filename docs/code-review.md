# QuanPhotos 代码审查（阶段性）

> 说明：这是我**已阅读部分**的阶段性问题文档，后续会继续补充。

## 审查范围（本轮已覆盖）

- `cmd/api/main.go`
- `internal/config/*`
- `internal/middleware/*`
- `internal/handler/*`（已看完主要 handler）
- `internal/service/*`（auth/photo/comment/ticket/admin/superadmin/conversation/notification/share/ranking/user/system）
- `internal/repository/postgresql/*`（已看完主要模块：photo/comment/conversation/notification/ranking/share/superadmin/tag/user/token/ticket 的关键文件）
- `internal/model/*`
- `internal/pkg/*`（jwt/hash/storage/logger/i18n/email/exif/imaging/response/security/database）
- `migrations/000001_init_schema.up.sql`

---

## 高优先级问题（建议先修）

1. **Request ID 实现是“固定值”，不是随机值**
   - 位置：`internal/middleware/request_id.go:29`
   - 问题：`randomString` 用 `letters[i%len(letters)]`，每次都会生成同一串字符。
   - 影响：请求链路追踪失效，日志关联不可信；在网关/监控中会造成冲突。

2. **Logout 刷新令牌校验算法用错，导致登出不真正失效 token**
   - 位置：`internal/service/auth/logout.go:27`
   - 问题：保存 refresh token 时用的是 SHA-256（`HashToken`），登出时却用 `CheckPassword(bcrypt)` 比较。
   - 影响：匹配几乎总失败，token 不会被删除，存在会话残留风险。

3. **CORS `Access-Control-Max-Age` 写法错误**
   - 位置：`internal/middleware/cors.go:39`
   - 问题：`string(rune(cfg.MaxAge))` 不是数字字符串，会变成单字符。
   - 影响：预检缓存时间头无效/异常，可能引发前端跨域问题。

4. **限流 key 生成错误（用户 ID 被当作 rune 转字符）**
   - 位置：`internal/middleware/rate_limit.go:120`, `internal/middleware/rate_limit.go:153`
   - 问题：`string(rune(userID))` 会产生不可控字符，且存在碰撞/截断风险。
   - 影响：同用户限流不稳定，可能误限流或绕过限流。

5. **评论“软删除”不会回收统计计数，数据会长期漂移**
   - 位置：`internal/repository/postgresql/comment/repository.go:179`
   - 关联触发器：`migrations/000001_init_schema.up.sql:436`
   - 问题：业务删除评论是 `UPDATE status='deleted'`，但计数递减触发器只在 `DELETE` 时触发。
   - 影响：`photos.comment_count` 与 `photo_comments.reply_count` 会越来越不准。

6. **对外直接回传 `err.Error()`，存在内部信息泄漏风险**
   - 位置示例：`internal/handler/comment_handler.go:63`, `internal/handler/conversation_handler.go:49`, `internal/handler/public_handler.go:84`
   - 问题：数据库/内部错误文本直接返回客户端。
   - 影响：泄漏内部结构、SQL 信息或实现细节，放大攻击面。

7. **生产安全兜底不足：JWT Secret 可为空且未强校验**
   - 位置：`internal/config/config.go:131`, `internal/handler/router.go:92`
   - 问题：配置默认允许空 `JWT_SECRET`，启动阶段未强制校验。
   - 影响：在误配置情况下可能导致弱签名安全。

8. **上传模块降级逻辑不安全：存储初始化失败仍继续启动上传接口**
   - 位置：`internal/handler/router.go:114`, `internal/service/photo/service.go:47`
   - 问题：本地存储初始化失败只 `log.Printf`，上传接口仍保留，最终在服务层返回“uploader not initialized”。
   - 影响：线上出现持续 500，且问题暴露在运行期而非启动期。

9. **权限模型存在“定义了但未真正落地”的逻辑缺口**
   - 位置：`internal/repository/postgresql/superadmin/repository.go:33`, `internal/middleware/rbac.go:105`, `internal/handler/router.go:324`
   - 问题：系统定义了细粒度 permission（`admin_permissions`），路由实际只按角色（admin/superadmin）放行。
   - 影响：权限设计与真实行为不一致，存在越权边界模糊。

10. **用户行为限制字段（can_comment/can_message/can_upload）未真正接入请求链路**
    - 位置：`internal/model/user.go:160`, `internal/handler/router.go:232`
    - 问题：字段可被超级管理员修改，但我在已读代码中未看到统一拦截点调用。
    - 影响：管理后台“禁言/禁传”可能只是写库，不生效或部分场景不生效。

---

## 中优先级问题（建议第二批处理）

1. **标签搜索参数文档与实现不一致**
   - 位置：`internal/handler/tag_handler.go:66`, `internal/service/tag/service.go:90`
   - 问题：接口文档写 `q`，实际绑定字段是 `keyword`。
   - 影响：调用方按文档请求会失败（400）。

2. **删除照片未删除物理文件，存在存储泄漏**
   - 位置：`internal/service/photo/service.go:426`
   - 问题：代码里有 `TODO`，仅删数据库记录。
   - 影响：磁盘持续增长，备份和运维成本上升。

3. **会话列表存在 N+1 查询**
   - 位置：`internal/service/conversation/service.go:142`
   - 问题：每个会话单独查一次最后消息。
   - 影响：会话数增大后接口性能明显下降。

4. **管理员用户列表返回结构有 `photo_count` 字段但未赋值**
   - 位置：`internal/service/admin/service.go:69`, `internal/service/admin/service.go:103`
   - 问题：字段定义存在但未填充。
   - 影响：返回数据与接口语义不一致。

5. **多处错误被静默吞掉，观测性不足**
   - 位置示例：`internal/service/auth/login.go:39`, `internal/service/auth/refresh.go:65`, `internal/service/photo/service.go:167`
   - 问题：`_ = ...` 直接忽略错误且无日志。
   - 影响：故障定位困难，异常行为难追踪。

6. **中文文案疑似编码损坏（通知/i18n/邮件模板）**
   - 位置：`internal/service/notification/service.go:196`, `internal/pkg/i18n/zh_cn.go:6`, `internal/pkg/email/service.go:225`
   - 问题：字符串呈现乱码（mojibake）。
   - 影响：用户侧文案不可读，国际化与通知质量受损。

---

## 续审补充（从 `ticket/get.go` 开始）

### 新增高优先级问题

1. **工单状态更新“未命中记录”时仍返回成功，导致结果误报**
   - 位置：`internal/repository/postgresql/ticket/update.go:23`, `internal/repository/postgresql/ticket/admin.go:128`
   - 问题：`rowsAffected == 0` 时直接 `return nil`，调用方会误判“更新成功”。
   - 影响：并发下（或数据已不存在时）会出现假成功，管理端状态变更可观测性失真。

2. **工单详情读取图片信息时吞掉数据库错误**
   - 位置：`internal/service/ticket/service.go:144`
   - 问题：`photoBrief, _ = ...` 直接丢弃错误。
   - 影响：真实数据库异常会被伪装成“无图片信息”，故障定位困难，且可能返回不完整数据而无告警。

### 新增中优先级问题

1. **创建工单对非法 `photo_id` 的错误语义不友好（容易返回 500）**
   - 位置：`internal/service/ticket/service.go:49`, `internal/handler/ticket_handler.go:53`
   - 问题：外键错误未被转换为业务错误，最终统一走 InternalError。
   - 影响：客户端收到 500，而不是可理解的参数错误/资源不存在错误。

2. **工单 OpenAPI 与实现不一致（影响联调）**
   - 位置：`docs/openapi.yaml:2251`, `docs/openapi.yaml:2301`, `docs/openapi.yaml:2310`, `docs/openapi.yaml:2667`, `docs/openapi.yaml:2682`, `docs/openapi.yaml:2687`, `docs/openapi.yaml:299`
   - 问题：
     - 文档把 `photo_id` 标记为必填，但实现中是可选（`*int64`）。
     - 文档工单类型仅列出 `appeal/report/other`，实现还支持 `bug/feedback`。
     - `GET /tickets` 文档缺少 `type` 查询参数。
     - `GET /admin/tickets` 文档缺少 `user_id` 查询参数，且 `status` 默认值写成 `open`（实现默认不过滤，相当于 `all`）。
   - 影响：调用方按文档构造请求会产生误用，文档可信度下降。

3. **工单仓储层存在未使用方法，维护成本增加**
   - 位置：`internal/repository/postgresql/ticket/get.go:77`, `internal/repository/postgresql/ticket/update.go:10`
   - 问题：`IsOwnedBy`、`UpdateStatus` 在当前服务链路中未被调用。
   - 影响：后续重构时容易形成“看起来有能力、实际未接线”的错觉，增加维护负担。

---

## 续审补充（superadmin 与接口文档一致性）

### 新增高优先级问题

1. **superadmin API 文档与真实路由严重漂移（有“文档存在但代码不存在”）**
   - 位置：`docs/openapi.yaml:3167`, `docs/api-endpoints.md:2701`, `internal/handler/router.go:356`
   - 问题：文档定义了 `PUT /superadmin/users/{id}/role`，但路由层并未注册该接口。
   - 影响：前端/SDK 按文档调用会直接 404，且造成权限管理能力认知偏差。

2. **superadmin 限制更新对“目标为 superadmin”返回 500 且暴露内部错误文本**
   - 位置：`internal/service/superadmin/service.go:296`, `internal/handler/superadmin_handler.go:305`
   - 问题：服务层返回通用 `errors.New(...)`，处理层未做业务映射，最终走 `InternalError(err.Error())`。
   - 影响：语义错误（应为业务拒绝而非 500）且带来错误信息泄漏风险。

### 新增中优先级问题

1. **superadmin 文档遗漏多个已实现的 GET 接口**
   - 位置：`internal/handler/router.go:362`, `internal/handler/router.go:366`, `internal/handler/router.go:372`, `internal/handler/router.go:377`, `docs/openapi.yaml:3003`, `docs/openapi.yaml:3099`, `docs/openapi.yaml:3201`
   - 问题：已实现的以下接口在文档中缺失或方法缺失：
     - `GET /superadmin/permissions`
     - `GET /superadmin/admins/{id}/permissions`
     - `GET /superadmin/reviewers/{id}/categories`
     - `GET /superadmin/users/{id}/restrictions`
   - 影响：客户端无法完整感知可用能力，联调成本高。

2. **superadmin 列表响应结构文档与实现不一致**
   - 位置：`internal/service/superadmin/service.go:65`, `internal/service/superadmin/service.go:71`, `internal/service/superadmin/service.go:31`, `internal/service/superadmin/service.go:41`, `docs/openapi.yaml:2996`, `docs/openapi.yaml:3092`, `docs/openapi.yaml:538`
   - 问题：实现返回 `admins/reviewers + total`，而文档写成 `list + pagination`，且字段定义（如 `role/created_at`）与实际返回（如 `avatar`）不一致。
   - 影响：接口契约不稳定，自动生成类型/SDK 易出错。

3. **工单/管理接口的响应文档与实现有偏差**
   - 位置：`internal/service/ticket/service.go:180`, `docs/openapi.yaml:2399`, `docs/openapi.yaml:322`, `internal/service/admin/service.go:445`, `docs/openapi.yaml:2709`, `internal/service/admin/service.go:496`, `docs/openapi.yaml:2779`
   - 问题：
     - 工单回复实际返回 `{id, content, created_at}`，文档却引用含 `user` 的 `TicketReply` 结构。
     - 管理端处理工单实际返回 `status/reply_id` 数据，文档仅给 `BaseResponse`。
     - 设为精选实际返回 `{photo_id, message}`，文档写成 `{id, photo_id, featured_at}`。
   - 影响：前端解析与后端真实载荷错位，出现字段缺失/反序列化问题。

4. **superadmin 批量授权/撤销与分类分配未使用事务，存在部分成功风险**
   - 位置：`internal/service/superadmin/service.go:134`, `internal/service/superadmin/service.go:161`, `internal/service/superadmin/service.go:227`, `internal/service/superadmin/service.go:249`
   - 问题：按条循环写库，任一中途失败会留下部分已生效的数据。
   - 影响：权限与分类分配可能处于半成功状态，回滚成本高。

5. **superadmin 列表查询存在 N+1 模式**
   - 位置：`internal/repository/postgresql/superadmin/repository.go:91`, `internal/repository/postgresql/superadmin/repository.go:105`, `internal/repository/postgresql/superadmin/repository.go:172`, `internal/repository/postgresql/superadmin/repository.go:187`
   - 问题：先查用户列表，再逐个查询权限/分类。
   - 影响：管理员/审核员数量增大后，接口时延与数据库压力线性上升。

---

## 续审补充（conversations / notifications）

### 新增中优先级问题

1. **会话消息接口文档参数与返回结构和实现不一致**
   - 位置：`internal/service/conversation/service.go:221`, `internal/service/conversation/service.go:227`, `docs/openapi.yaml:1601`, `docs/openapi.yaml:1619`, `docs/api-endpoints.md:1580`, `docs/api-endpoints.md:1589`
   - 问题：
     - 文档声明 `before_id` 参数，但实现未接收该参数。
     - 文档返回 `conversation + messages`，实现返回 `list + pagination`。
   - 影响：调用方按文档实现会出现“参数无效+响应反序列化不匹配”。

2. **会话创建/发送接口的状态码与响应载荷文档不一致**
   - 位置：`internal/handler/conversation_handler.go:92`, `internal/handler/conversation_handler.go:181`, `internal/service/conversation/service.go:175`, `docs/openapi.yaml:1560`, `docs/openapi.yaml:1574`, `docs/openapi.yaml:1653`, `docs/api-endpoints.md:1549`, `docs/api-endpoints.md:1552`
   - 问题：
     - 实现返回 HTTP 201，文档写 200。
     - 创建会话实际返回 `message` 对象，文档写 `message_id`。
   - 影响：网关校验、客户端 SDK 及前端状态处理容易出错。

3. **通知未读统计接口返回结构与文档不一致**
   - 位置：`internal/service/notification/service.go:155`, `docs/openapi.yaml:1756`, `docs/api-endpoints.md:1767`
   - 问题：实现仅返回 `{count}`，文档写 `{total, by_type}`。
   - 影响：前端若依赖按类型统计会直接失效。

4. **“全部标记已读”接口文档声明可按类型筛选，但实现忽略该参数**
   - 位置：`internal/handler/notification_handler.go:123`, `internal/repository/postgresql/notification/repository.go:177`, `docs/openapi.yaml:1800`, `docs/api-endpoints.md:1819`
   - 问题：文档存在 `type` 请求体字段，实际接口不解析请求体，且总是全量标记。
   - 影响：调用方以为可局部标记，实际会误操作为“全部已读”。

5. **通知类型枚举文档不完整**
   - 位置：`internal/repository/postgresql/notification/repository.go:20`, `docs/openapi.yaml:1709`, `docs/api-endpoints.md:1698`
   - 问题：实现支持 `reply/share/featured/review/message`，文档枚举未覆盖。
   - 影响：客户端过滤器与类型映射不全，出现“未知类型”处理分支。

6. **会话软删除语义与访问控制不一致**
   - 位置：`internal/repository/postgresql/conversation/repository.go:180`, `internal/repository/postgresql/conversation/repository.go:192`
   - 问题：`IsParticipant` 不检查 `user*_deleted` 标志；用户软删除后仍可通过会话 ID 继续读写会话。
   - 影响：删除语义被绕过，出现“列表不可见但仍可操作”的状态不一致。

---

## 续审补充（announcements / categories / tags）

### 新增高优先级问题

1. **公告更新在特定分支下“找不到记录”会误返回 500**
   - 位置：`internal/repository/postgresql/photo/admin.go:320`, `internal/repository/postgresql/photo/admin.go:323`, `internal/service/admin/service.go:669`, `internal/handler/admin_handler.go:645`
   - 问题：更新公告时若 `status=published` 且记录不存在，仓储层先查 `published_at` 直接返回 `sql.ErrNoRows`，未转换为 `ErrNotFound`。
   - 影响：客户端收到 500 而不是 404，错误语义错误且影响排障。

### 新增中优先级问题

1. **公共公告接口文档与实现不一致（过滤参数与返回字段）**
   - 位置：`internal/handler/public_handler.go:170`, `internal/handler/public_handler.go:207`, `internal/handler/public_handler.go:245`, `docs/openapi.yaml:1836`, `docs/openapi.yaml:1835`, `docs/openapi.yaml:495`, `docs/api-endpoints.md:1847`, `docs/api-endpoints.md:1890`
   - 问题：
     - 文档声明 `is_pinned` 过滤，但实现未接收该参数。
     - 文档默认 `page_size=10`，实现默认是 20。
     - 文档示例含 `author/created_at/updated_at`，实现公开结构侧重 `published_at`，字段契约不同。
   - 影响：前端筛选与展示字段容易出现偏差，联调成本升高。

2. **管理员公告接口文档缺漏与响应契约漂移**
   - 位置：`internal/handler/router.go:350`, `docs/openapi.yaml:2904`, `docs/api-endpoints.md:2367`, `internal/handler/admin_handler.go:574`, `docs/openapi.yaml:2884`, `internal/handler/admin_handler.go:649`, `docs/openapi.yaml:2936`
   - 问题：
     - 代码已实现 `GET /admin/announcements/:id`，文档未覆盖。
     - 创建接口实际返回 201 + 完整对象，文档写 200 + 精简对象。
     - 更新接口实际返回对象数据，文档仅写 `BaseResponse`。
   - 影响：SDK 生成与前端类型定义不稳定，调用时容易出现字段/状态码不匹配。

3. **分类接口文档与实现存在多处不一致**
   - 位置：`internal/handler/category_handler.go:31`, `internal/handler/category_handler.go:42`, `internal/service/category/service.go:58`, `docs/openapi.yaml:1990`, `docs/openapi.yaml:2007`, `docs/api-endpoints.md:720`, `docs/api-endpoints.md:728`, `internal/handler/category_handler.go:176`, `internal/handler/category_handler.go:191`, `docs/openapi.yaml:2124`, `docs/api-endpoints.md:902`
   - 问题：
     - 实现走 `page/page_size + pagination`，文档写 `include_empty + array`。
     - 删除实现参数是 `force`，文档写 `move_to`。
   - 影响：调用参数错误率高，删除行为预期与实际不一致（存在误操作风险）。

4. **标签列表/标签照片接口文档与实现不一致**
   - 位置：`internal/service/tag/service.go:33`, `internal/service/tag/service.go:56`, `internal/service/tag/service.go:131`, `docs/openapi.yaml:2145`, `docs/openapi.yaml:2162`, `docs/openapi.yaml:2198`, `docs/api-endpoints.md:930`, `docs/api-endpoints.md:1007`
   - 问题：
     - 标签列表实现支持 `page/page_size/order_by` 且返回分页，文档写 `limit` 且返回数组。
     - 标签照片实现支持 `sort_by/sort_order`，文档未体现。
   - 影响：前端筛选/排序能力被文档误导，接口能力无法被正确使用。

5. **category 模块存在未接线路径与未使用字段，增加维护噪音**
   - 位置：`internal/service/category/service.go:21`, `internal/service/category/service.go:170`, `internal/repository/postgresql/category/repository.go:237`, `internal/handler/router.go:257`
   - 问题：`Service.baseURL` 当前未被使用，且 category 的 `ListPhotos` 请求/仓储能力未暴露到路由层。
   - 影响：后续维护容易误判“功能已上线”，增加重构和排查成本。

---

## 续审补充（rankings / featured）

### 新增中优先级问题

1. **排行榜 period/sort_by 枚举与文档不一致，会导致“按文档请求却返回错误统计口径”**
   - 位置：`internal/service/ranking/service.go:25`, `internal/repository/postgresql/ranking/repository.go:60`, `docs/openapi.yaml:1897`, `docs/openapi.yaml:1909`, `docs/api-endpoints.md:1914`, `docs/api-endpoints.md:1916`
   - 问题：
     - 实现使用 `day/week/month/all`，文档写 `daily/weekly/monthly/all`。
     - 实现 `sort_by` 是 `like_count/view_count/comment_count/share_count`，文档写 `score/likes/views/favorites`。
   - 影响：调用方按文档传参时，后端可能回落到默认逻辑（如全时段/默认排序），结果与预期不一致。

2. **排行榜响应结构与文档不一致**
   - 位置：`internal/service/ranking/service.go:50`, `internal/service/ranking/service.go:126`, `internal/service/ranking/service.go:152`, `docs/openapi.yaml:1924`, `docs/openapi.yaml:1975`, `docs/api-endpoints.md:1925`, `docs/api-endpoints.md:1970`
   - 问题：实现返回 `period + sort_by + 扁平 list`，文档示例/Schema 包含 `period_start/period_end`、嵌套 `photo/user`、`score` 等不同字段。
   - 影响：前端解析与展示模型难以稳定，易出现字段缺失或类型不匹配。

3. **精选照片响应结构与文档不一致**
   - 位置：`internal/handler/public_handler.go:32`, `internal/handler/public_handler.go:88`, `docs/openapi.yaml:404`, `docs/openapi.yaml:1494`, `docs/api-endpoints.md:1456`
   - 问题：实现返回扁平字段（`id/title/thumbnail_url/like_count/view_count/user_id`），文档写的是 `photo + featured_reason + featured_at` 结构。
   - 影响：调用方如果按文档解包，将无法直接获得实际返回数据。

---

## 续审补充（photos / users）

### 新增高优先级问题

1. **`raw_file` 缺少大小/类型校验，可被用作存储打满入口**
   - 位置：`internal/handler/photo_handler.go:73`, `internal/service/photo/upload.go:89`, `internal/service/photo/upload.go:395`
   - 问题：主图会走大小/类型校验，但可选 `raw_file` 在读取后直接落盘，未复用 `maxUploadSize` 与类型白名单校验。
   - 影响：可通过超大或非预期文件持续写盘，形成存储资源消耗风险。

2. **点赞/收藏与收藏列表未校验照片审核状态，存在“未过审内容可交互/可见”风险**
   - 位置：`internal/service/photo/service.go:359`, `internal/service/photo/service.go:382`, `internal/repository/postgresql/photo/favorite.go:68`, `internal/repository/postgresql/photo/list.go:217`
   - 问题：点赞/收藏只做 `id` 存在性校验，不限制 `approved`；收藏列表查询也未过滤 `p.status='approved'`。
   - 影响：非作者可对未过审照片进行交互，且可通过收藏列表看到未公开内容，破坏审核边界。

3. **修改密码后未回收该用户已有 refresh token，会话仍可延续**
   - 位置：`internal/service/user/service.go:87`, `internal/service/user/service.go:109`, `internal/repository/postgresql/token/delete.go:51`
   - 问题：`ChangePassword` 仅更新密码哈希，没有联动删除用户已有 refresh token（仓储层已有 `DeleteByUserID` 能力）。
   - 影响：密码修改后历史会话仍可继续刷新 access token，降低账号被盗后的止损效果。

### 新增中优先级问题

1. **照片列表筛选参数缺少格式校验，非法日期会直接落到 500**
   - 位置：`internal/service/photo/service.go:63`, `internal/repository/postgresql/photo/list.go:131`, `internal/repository/postgresql/photo/list.go:137`, `internal/handler/photo_handler.go:160`
   - 问题：`taken_from/taken_to` 直接作为字符串拼入 SQL 条件，未在入口做日期格式验证；非法值会在数据库层报错并返回 500。
   - 影响：客户端参数错误被误报为服务端故障，影响可观测性与调用体验。

2. **照片上传 `category_id` 解析失败被静默降级为 0**
   - 位置：`internal/handler/photo_handler.go:93`
   - 问题：`strconv.ParseInt` 错误被忽略，非法 `category_id` 会被当作未设置分类继续处理。
   - 影响：输入错误无法被及时拦截，导致数据质量问题和调用方误判“提交成功”。

3. **照片详情聚合过程吞掉多处仓储错误，容易“带错返回成功”**
   - 位置：`internal/service/photo/service.go:167`, `internal/service/photo/service.go:186`, `internal/service/photo/service.go:196`, `internal/service/photo/service.go:204`
   - 问题：浏览计数递增、分类查询、标签查询、用户点赞/收藏状态查询均存在忽略错误路径。
   - 影响：服务可能在依赖异常时仍返回 200，但字段不完整或状态错误，增加线上排障难度。

4. **photos 接口文档与实现存在多处契约漂移（参数/响应结构）**
   - 位置：`internal/handler/photo_handler.go:138`, `internal/service/photo/service.go:65`, `internal/repository/postgresql/photo/list.go:49`, `internal/handler/photo_handler.go:332`, `internal/model/photo.go:109`, `docs/openapi.yaml:898`, `docs/openapi.yaml:928`, `docs/openapi.yaml:1133`, `docs/openapi.yaml:1182`, `docs/api-endpoints.md:396`, `docs/api-endpoints.md:403`, `docs/api-endpoints.md:605`, `docs/api-endpoints.md:1222`
   - 问题：
     - 列表查询实现是 `sort_by/sort_order + taken_from/taken_to`，文档写 `sort + date_from/date_to + user_id`。
     - 实现 `page_size` 上限为 100，文档示例写 50。
     - 点赞/收藏实现返回 message，文档写返回 `like_count/favorite_count`。
     - 详情实现返回 `PhotoDetail`（无 `share_count`、用户为 `UserBrief`），文档按 `Photo` Schema 描述。
   - 影响：SDK、前端类型与后端真实行为不一致，联调阶段高频出现“字段缺失/参数不生效”。

5. **users 接口文档与实现返回模型不一致**
   - 位置：`internal/handler/user_handler.go:42`, `internal/handler/user_handler.go:80`, `internal/handler/user_handler.go:110`, `internal/model/user.go:47`, `internal/model/user.go:58`, `docs/openapi.yaml:79`, `docs/openapi.yaml:114`, `docs/openapi.yaml:744`, `docs/openapi.yaml:776`, `docs/api-endpoints.md:239`, `docs/api-endpoints.md:340`
   - 问题：
     - `/users/me` 实际返回 `UserProfile`（含 `status/can_*/last_login_at/updated_at`），文档按 `User`（含 `photo_count/favorite_count`）描述。
     - `/users/{id}` 实际返回 `UserPublicInfo`（含 `role/location`），文档按 `UserPublic`（含 `photo_count`）描述。
   - 影响：用户域接口数据模型在文档层面不可依赖，客户端类型定义频繁返工。

6. **用户资料更新缺少输入长度约束，与文档声明不一致**
   - 位置：`internal/service/user/service.go:61`, `internal/handler/user_handler.go:75`, `docs/openapi.yaml:762`
   - 问题：`UpdateProfileRequest` 未设置 `binding` 约束，`bio/location/avatar` 长度不会按文档限制校验（文档声明 `bio` 最长 200）。
   - 影响：可写入超长文本，导致前后端规则不一致，增加脏数据与展示异常风险。

---

## 续审补充（auth / comments / share）

### 新增高优先级问题

1. **评论回复未校验父评论归属照片，允许“跨照片挂接回复”**
   - 位置：`internal/service/comment/service.go:202`, `internal/repository/postgresql/comment/repository.go:210`, `internal/repository/postgresql/comment/repository.go:156`
   - 问题：创建回复时仅校验 `parent_id` 是否存在（且可见），未校验父评论与当前 `photo_id` 是否一致。
   - 影响：可把 A 照片回复挂到 B 照片评论树，导致 `reply_count` 污染、上下文串线及数据一致性破坏。

2. **评论读写未按照片公开状态做访问控制，审核边界可被绕过**
   - 位置：`internal/repository/postgresql/comment/repository.go:315`, `internal/repository/postgresql/comment/repository.go:78`
   - 问题：评论创建仅检查照片 `id` 是否存在，列表查询也只按 `photo_id` + 评论状态过滤，未限制 `photos.status='approved'`。
   - 影响：在已知照片 ID 的前提下，可对未公开照片进行评论或读取评论，破坏内容审核可见性边界。

### 新增中优先级问题

1. **Refresh token 轮换删除失败被吞掉，存在旧 token 继续可用风险**
   - 位置：`internal/service/auth/refresh.go:65`
   - 问题：刷新流程中删除旧 refresh token 的错误被忽略，删除失败时仍会签发新 token。
   - 影响：异常场景下同一旧 token 可能继续可用，增大会话重放风险并降低轮换策略有效性。

2. **auth 接口文档与实现存在多处契约漂移（校验规则/状态码/请求响应结构）**
   - 位置：`internal/service/auth/service.go:41`, `internal/handler/auth_handler.go:64`, `internal/handler/auth_handler.go:103`, `internal/handler/auth_handler.go:142`, `internal/handler/auth_handler.go:156`, `internal/handler/router.go:207`, `docs/openapi.yaml:598`, `docs/openapi.yaml:611`, `docs/openapi.yaml:646`, `docs/openapi.yaml:659`, `docs/openapi.yaml:703`, `docs/openapi.yaml:708`, `docs/api-endpoints.md:88`, `docs/api-endpoints.md:100`, `docs/api-endpoints.md:123`, `docs/api-endpoints.md:137`, `docs/api-endpoints.md:191`
   - 问题：
     - 注册实现是 `username min=3/max=50`、密码最大 72、返回 201 + `user/tokens`；文档写 4~20、密码最大 32、200 + 精简对象。
     - 登录实现仅按 `username` 查询，文档声明“用户名或邮箱”；实现返回 `tokens.expires_at`，文档写 `expires_in`。
     - 登出实现路由未鉴权且依赖 body 的 `refresh_token` 做失效，文档写 Bearer 鉴权且未声明请求体。
   - 影响：调用方按文档接入会出现参数/状态码/字段级别系统性不匹配，且登出行为预期与实际不一致。

3. **comments 接口文档与实现契约不一致**
   - 位置：`internal/service/comment/service.go:40`, `internal/service/comment/service.go:187`, `internal/handler/comment_handler.go:114`, `internal/handler/router.go:249`, `docs/openapi.yaml:1294`, `docs/openapi.yaml:1343`, `docs/openapi.yaml:1348`, `docs/openapi.yaml:1390`, `docs/api-endpoints.md:1267`, `docs/api-endpoints.md:1326`, `docs/api-endpoints.md:1395`
   - 问题：
     - 列表实现参数是 `sort_by(created_at/like_count)` 且支持 `parent_id`，文档写 `sort(newest/oldest/popular)`。
     - 创建评论实现 `content max=1000` 且返回 201；文档写 `max=500`，OpenAPI 写 200。
     - 代码已实现 `DELETE /comments/:id/like`，文档未覆盖该路由。
   - 影响：评论模块在参数、状态码和路由层面都存在契约偏差，前后端联调成本高且易误判后端行为。

4. **分享平台语义在落库时被折叠，无法保留真实平台维度**
   - 位置：`internal/service/share/service.go:32`, `internal/repository/postgresql/share/repository.go:49`, `migrations/000001_init_schema.up.sql:491`
   - 问题：请求层 `platform` 支持 `twitter/facebook/weibo/wechat/link`，入库时仅映射为 `internal/external` 两类。
   - 影响：后续无法按真实平台做统计、风控和效果分析，接口输入语义在数据层丢失。

5. **分享计数易被重复请求放大，业务指标可信度不足**
   - 位置：`internal/handler/router.go:241`, `internal/service/share/service.go:54`, `migrations/000001_init_schema.up.sql:486`
   - 问题：分享记录按请求直接插入，库表无去重约束（如 `user+photo+platform`），路由也无频控。
   - 影响：同一用户可短时间反复调用接口累加 `share_count`，造成分享指标被刷高。

---

## 续审补充（public / system）

### 新增中优先级问题

1. **`/system/info` 文档与实现响应结构不一致**
   - 位置：`internal/handler/system_handler.go:40`, `internal/service/system/system.go:113`, `docs/openapi.yaml:3277`, `docs/api-endpoints.md:2800`
   - 问题：实现返回 `app/runtime/storage/i18n` 四层嵌套结构，文档写的是扁平 `version/supported_formats/max_upload_size/languages`。
   - 影响：客户端按文档解包会失败，系统信息接口契约不可依赖。

2. **`/health` 接口返回语义在文档中不完整**
   - 位置：`internal/handler/system_handler.go:29`, `internal/service/system/system.go:94`, `docs/openapi.yaml:3246`, `docs/openapi.yaml:3253`, `docs/api-endpoints.md:2779`
   - 问题：实现在降级时返回 HTTP 503，且包含 `checks.database` 字段；文档仅描述 200 + `status/timestamp`。
   - 影响：监控与调用方无法依据文档准确处理“降级但可访问”的健康状态。

3. **系统信息接口默认匿名开放，暴露较多运行时与构建细节**
   - 位置：`internal/handler/router.go:199`, `internal/service/system/system.go:114`, `internal/service/system/system.go:121`
   - 问题：匿名可读 `env/git_commit/build_time/go_version/goroutines` 等运行信息。
   - 影响：增加外部探测与指纹识别面，建议按环境收敛或最小化公开字段。

---

## 建议的首修顺序

1. 先修安全与正确性：`request_id`、`logout token 校验`、`CORS Max-Age`、`rate-limit key`、`ticket 状态更新假成功`。
2. 修统计一致性：评论软删除计数问题。
3. 修权限落地：把 permission/can_* 限制真正接入中间件或服务校验。
4. 修可观测性：禁止直接返回 `err.Error()` 给客户端，补全日志。

---

## Top10 必修清单（收敛版）

> 目标：先消除“安全与数据正确性”高风险点，再处理接口契约漂移。

1. **修复 refresh token 失效链路**
   - 涉及：`logout` 哈希校验错误、`refresh` 轮换删除错误被吞、`change password` 不回收历史 token。
   - 目标：确保“登出/改密/刷新”三条路径都能强一致失效旧 token。

2. **修复评论与照片的审核可见性边界**
   - 涉及：评论创建/读取未限制 `approved`、点赞收藏与收藏列表未限制 `approved`。
   - 目标：所有公开侧交互统一以“照片可见性”作为前置条件。

3. **修复评论回复跨照片挂接问题**
   - 涉及：`parent_id` 未校验归属 `photo_id`。
   - 目标：保证评论树拓扑正确，防止 `reply_count` 与上下文污染。

4. **修复评论软删除统计漂移**
   - 涉及：业务软删除与 DB 触发器（仅 `DELETE` 递减）不一致。
   - 目标：`comment_count/reply_count` 长期准确，不随软删除漂移。

5. **修复限流/CORS/请求追踪基础中间件错误**
   - 涉及：`request_id` 固定值、rate-limit key 转换错误、CORS Max-Age 写法错误。
   - 目标：恢复链路追踪、限流准确性和跨域预检稳定性。

6. **修复 ticket 状态更新“假成功”与公告更新 500/404 语义错误**
   - 涉及：仓储层 not found 转换缺失、服务层未做一致性校验。
   - 目标：接口状态码与真实落库结果一致。

7. **修复上传入口的资源风控缺口**
   - 涉及：`raw_file` 无大小/类型校验、上传器初始化失败仍开放上传路由。
   - 目标：避免存储被打满与运行期持续 500。

8. **落实权限模型（permission/can_*）到请求链路**
   - 涉及：细粒度权限仅定义未执法、`can_comment/can_message/can_upload` 未统一拦截。
   - 目标：后台风控配置“可配即生效”。

9. **统一错误处理与可观测性**
   - 涉及：多处 `err.Error()` 直接出参、服务层大量吞错返回 200。
   - 目标：外部错误语义稳定、内部日志可追踪。

10. **收敛 OpenAPI/API 文档与实现契约**
   - 涉及：`auth/photos/users/comments/rankings/announcements/system` 多模块参数、状态码、响应结构漂移。
   - 目标：文档可直接驱动 SDK 与前端类型，不再靠“联调口口相传”。

---

## 修复路线建议（按 PR 批次）

1. **PR-1（安全链路）**：token 失效链路 + 上传风控 + 审核可见性边界。
2. **PR-2（数据一致性）**：评论树归属校验 + 软删除计数修复 + ticket/announcement 状态语义修复。
3. **PR-3（中间件稳定性）**：request_id/rate-limit/CORS 一次性修正并补回归测试。
4. **PR-4（权限落地）**：permission/can_* 统一中间件或 service guard。
5. **PR-5（契约对齐）**：统一修订 OpenAPI + `docs/api-endpoints.md`，与代码同版发布。

---

## 进度记录（方便下次续审）

- **当前已完成**：在既有第一轮审查基础上，已补审 `ticket/superadmin/conversations/notifications/announcements/categories/tags/rankings/featured/photos/users/auth/comments/share/public/system` 模块，并补充了对应问题清单。
- **下次继续建议起点**：按上面的 `PR-1` 开始落修；我可以逐文件给出最小改动补丁草案与回归测试点。

