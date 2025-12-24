# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

QuanPhotos 是一个航空摄影社区平台后端 API 服务，类似 JetPhotos。

**核心功能**: 用户系统（guest/user/reviewer/admin/superadmin）、照片管理、AI 审核集成、工单申诉、EXIF 解析

**照片工作流**: 上传 → AI 初审 → 人工复审 → 发布（AI 初审失败可提工单申诉）

## 常用命令

```bash
# 开发运行
go run cmd/api/main.go          # 直接运行
make dev                        # 热重载（需要 air）

# 构建
make build                      # 构建到 bin/api

# 测试
make test                       # 运行所有测试
go test -v ./internal/service/auth/...  # 运行单个包测试

# 代码质量
make fmt                        # 格式化代码
make lint                       # 代码检查（需要 golangci-lint）

# 数据库迁移（需要 golang-migrate，环境变量需先设置）
make migrate-up                 # 执行迁移
make migrate-down               # 回滚迁移
make migrate-create name=xxx    # 创建新迁移文件

# Docker
make docker-run                 # docker-compose up -d
make docker-stop                # docker-compose down
```

## 架构分层

```
cmd/api/main.go                 # 入口：配置加载 → 日志初始化 → 数据库连接 → 自动迁移 → 路由设置
    ↓
internal/handler/router.go      # 依赖注入中心：初始化 repo → service → handler，注册路由
    ↓
internal/handler/*.go           # HTTP 处理器：请求解析、调用 service、统一响应
    ↓
internal/service/*/*.go         # 业务逻辑层：事务控制、业务规则
    ↓
internal/repository/postgresql/*/*.go  # 数据访问层：原生 SQL 查询（sqlx）
    ↓
internal/model/*.go             # 数据模型：与数据库表映射
```

**关键文件**:
- `internal/handler/router.go` - 所有路由定义和依赖注入
- `internal/config/config.go` - 配置结构和环境变量加载
- `internal/pkg/response/response.go` - 统一响应格式和错误码
- `internal/model/user.go` - 用户角色和权限定义

## 代码规范

### 响应格式
所有 API 使用统一响应格式，通过 `internal/pkg/response` 包：
```go
response.Success(c, data)
response.SuccessWithPagination(c, list, page, pageSize, total)
response.NotFound(c, "user not found")
```

### 错误码
定义在 `internal/pkg/response/response.go`，格式：`4XYYY` (客户端错误) 或 `5XYYY` (服务端错误)

### 用户角色层级
guest(0) < user(1) < reviewer(2) < admin(3) < superadmin(4)

### 数据库
- 使用 sqlx 进行原生 SQL 查询，非 ORM
- 迁移文件在 `migrations/` 目录，按序号命名
- 开发环境自动运行迁移和种子数据

### 命名
- 文件名：`photo_handler.go`
- 包名：`handler`, `service`
- 表名：复数形式 `users`, `photos`

## 环境变量

通过 `.env` 文件配置，关键变量：
- `APP_ENV`: development/production
- `DB_*`: 数据库连接
- `JWT_SECRET`: JWT 签名密钥（必须设置）
- `STORAGE_PATH`: 文件存储路径

配置结构定义在 `internal/config/config.go`

## 开发注意事项

- **禁止使用 `> nul`**：Windows 上会创建无法删除的文件，使用 `> /dev/null 2>&1`
- 日志使用 zap structured logger
- 所有新增中间件在 `internal/middleware/` 目录
