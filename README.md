# QuanPhotos Backend

航空摄影社区平台后端服务

## 简介

QuanPhotos 是一个专业的航空摄影社区平台，为航空爱好者提供照片分享、浏览和交流的空间。本仓库为后端 API 服务，提供用户管理、照片管理、AI 审核集成、工单系统等核心功能。

## 功能特性

- **用户系统**：注册登录、角色权限（游客/用户/管理员/超级管理员）
- **照片管理**：上传、分类、搜索、收藏
- **AI 审核集成**：对接 AI 微服务进行照片自动初审
- **工单系统**：用户申诉 AI 审核结果
- **EXIF 解析**：完整提取照片元数据
- **多语言支持**：中文、英文

## 技术栈

| 类别 | 技术 |
|------|------|
| 语言 | Go 1.21+ |
| 框架 | Gin |
| 数据库 | PostgreSQL |
| 数据访问 | sqlx |
| 存储 | 本地文件存储 |
| 部署 | Docker |

## 项目结构

```
├── cmd/
│   └── api/
│       └── main.go
├── internal/
│   ├── config/
│   ├── handler/
│   ├── middleware/
│   ├── model/
│   ├── repository/
│   ├── service/
│   └── pkg/
│       ├── exif/
│       ├── storage/
│       ├── i18n/
│       └── response/
├── migrations/
├── api/
├── scripts/
├── Dockerfile
├── docker-compose.yml
└── go.mod
```

## 快速开始

### 环境要求

- Go 1.21+
- PostgreSQL 15+
- Docker & Docker Compose（可选）

### 本地开发

1. 克隆仓库

```bash
git clone https://github.com/yourname/QuanPhotos-backend.git
cd QuanPhotos-backend
```

2. 配置环境变量

```bash
cp .env.example .env
# 编辑 .env 填写数据库连接等配置
```

3. 初始化数据库

```bash
# 执行迁移
make migrate-up
```

4. 启动服务

```bash
go run cmd/api/main.go
```

服务默认运行在 `http://localhost:8080`

### Docker 部署

```bash
docker-compose up -d
```

## 配置说明

| 环境变量 | 说明 | 默认值 |
|----------|------|--------|
| `PORT` | 服务端口 | 8080 |
| `DB_HOST` | 数据库地址 | localhost |
| `DB_PORT` | 数据库端口 | 5432 |
| `DB_NAME` | 数据库名称 | quanphotos |
| `DB_USER` | 数据库用户 | postgres |
| `DB_PASSWORD` | 数据库密码 | - |
| `STORAGE_PATH` | 图片存储路径 | ./uploads |
| `AI_SERVICE_URL` | AI 审核服务地址 | http://localhost:8000 |
| `JWT_SECRET` | JWT 密钥 | - |

## API 文档

启动服务后访问 Swagger 文档：

```
http://localhost:8080/swagger/index.html
```

### 主要接口

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/auth/register` | 用户注册 |
| POST | `/api/v1/auth/login` | 用户登录 |
| GET | `/api/v1/photos` | 照片列表 |
| POST | `/api/v1/photos` | 上传照片 |
| GET | `/api/v1/photos/:id` | 照片详情 |
| POST | `/api/v1/tickets` | 创建工单 |
| GET | `/api/v1/admin/reviews` | 待审核列表 |

## 照片工作流

```
上传 → AI 初审 → 人工复审 → 发布
         ↓
      初审失败 → 用户提交工单 → 人工复审
```

## 支持的图片格式

**标准格式**
- PNG、JPG、JPEG

**RAW 格式**
- Canon: CR2, CR3
- Nikon: NEF, NRW
- Sony: ARW
- Fujifilm: RAF
- Olympus: ORF
- Panasonic: RW2
- 通用: DNG

## 相关项目

| 项目 | 说明 |
|------|------|
| [QuanPhotos-web](https://github.com/yourname/QuanPhotos-web) | 前端 Web 应用 |
| [QuanPhotos-ai](https://github.com/yourname/QuanPhotos-ai) | AI 审核微服务 |

## 开发指南

```bash
# 运行测试
make test

# 代码格式化
make fmt

# 代码检查
make lint

# 生成 Swagger 文档
make swagger
```

## License

MIT License