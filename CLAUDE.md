# QuanPhotos Backend

## 项目概述

QuanPhotos 是一个类似 JetPhotos 的航空摄影社区平台，本仓库为后端 API 服务。

### 核心功能

- 用户系统：游客浏览、注册用户（浏览/收藏/上传）、管理员、超级管理员
- 照片管理：上传、审核、发布、分类浏览
- AI 审核集成：调用独立的 AI 微服务进行照片初审
- 工单系统：用户对 AI 审核结果提出申诉
- EXIF 解析：详细记录照片元数据

### 照片工作流

上传 → AI 初审 → 人工复审 → 发布

AI 初审失败可提交工单申诉。

## 技术栈

- **语言**: Go 1.21+
- **框架**: Gin
- **数据库**: PostgreSQL
- **ORM**: sqlx（非 ORM，SQL 优先）
- **存储**: 本地文件存储
- **部署**: Docker
- **API 风格**: RESTful

## 项目结构

```
QuanPhotos-backend/
├── cmd/
│   └── api/
│       └── main.go           # 应用入口
├── internal/
│   ├── config/               # 配置管理
│   ├── handler/              # HTTP 处理器
│   ├── middleware/           # 中间件
│   ├── model/                # 数据模型
│   ├── repository/           # 数据访问层
│   ├── service/              # 业务逻辑层
│   └── pkg/
│       ├── exif/             # EXIF 解析
│       ├── storage/          # 文件存储
│       ├── i18n/             # 国际化
│       └── response/         # 统一响应格式
├── migrations/               # 数据库迁移
├── api/                      # API 文档 (OpenAPI/Swagger)
├── scripts/                  # 部署脚本
├── Dockerfile
├── docker-compose.yml
└── go.mod
```

## 环境变量配置

所有配置通过 `.env` 文件或环境变量读取，禁止硬编码敏感信息。

### 服务配置

```env
# 应用基础配置
APP_ENV=development              # 运行环境: development / production
APP_PORT=8080                    # 服务端口
APP_DEBUG=true                   # 调试模式
APP_NAME=QuanPhotos              # 应用名称

# 日志配置
LOG_LEVEL=info                   # 日志级别: debug / info / warn / error
LOG_FORMAT=json                  # 日志格式: json / text
LOG_OUTPUT=stdout                # 输出位置: stdout / file
LOG_FILE_PATH=./logs/app.log     # 日志文件路径（当 LOG_OUTPUT=file 时）
```

### 数据库配置

```env
# PostgreSQL 连接
DB_HOST=localhost
DB_PORT=5432
DB_NAME=quanphotos
DB_USER=postgres
DB_PASSWORD=your_password
DB_SSLMODE=disable               # disable / require / verify-full
DB_MAX_OPEN_CONNS=25             # 最大打开连接数
DB_MAX_IDLE_CONNS=5              # 最大空闲连接数
DB_CONN_MAX_LIFETIME=300         # 连接最大生命周期（秒）
```

### JWT 认证配置

```env
# JWT 配置
JWT_SECRET=your_jwt_secret_key   # JWT 签名密钥（至少 32 位随机字符串）
JWT_ACCESS_EXPIRE=3600           # Access Token 过期时间（秒），默认 1 小时
JWT_REFRESH_EXPIRE=604800        # Refresh Token 过期时间（秒），默认 7 天
JWT_ISSUER=quanphotos            # Token 签发者
```

### 存储配置

```env
# 本地存储
STORAGE_TYPE=local               # 存储类型: local（后续可扩展 oss/s3）
STORAGE_PATH=./uploads           # 本地存储根目录（绝对路径或相对路径）
STORAGE_MAX_SIZE=52428800        # 单文件最大大小（字节，默认 50MB）
STORAGE_ALLOWED_TYPES=jpg,jpeg,png,cr2,cr3,nef,arw,raf,orf,rw2,dng

# 存储子目录（相对于 STORAGE_PATH）
STORAGE_PHOTOS_DIR=photos        # 照片存储目录
STORAGE_RAW_DIR=raw              # RAW 文件存储目录
STORAGE_THUMBS_DIR=thumbnails    # 缩略图存储目录
STORAGE_TEMP_DIR=temp            # 临时文件目录
```

### AI 服务配置

```env
# AI 审核服务
AI_SERVICE_URL=http://localhost:8000    # AI 服务地址
AI_SERVICE_TIMEOUT=30                   # 请求超时时间（秒）
AI_SERVICE_RETRY=3                      # 失败重试次数
AI_SERVICE_API_KEY=                     # AI 服务 API Key（如需要）
```

### 邮件配置（可选）

```env
# SMTP 邮件服务
SMTP_ENABLED=false
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USER=noreply@example.com
SMTP_PASSWORD=your_password
SMTP_FROM=QuanPhotos <noreply@example.com>
SMTP_TLS=true
```

### Redis 配置（可选，用于缓存/会话）

```env
# Redis
REDIS_ENABLED=false
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0
REDIS_POOL_SIZE=10
```

### 安全配置

```env
# CORS 跨域
CORS_ENABLED=true
CORS_ALLOWED_ORIGINS=http://localhost:5173,https://quanphotos.com
CORS_ALLOWED_METHODS=GET,POST,PUT,DELETE,OPTIONS
CORS_ALLOWED_HEADERS=Authorization,Content-Type,Accept-Language
CORS_MAX_AGE=86400

# 限流
RATE_LIMIT_ENABLED=true
RATE_LIMIT_REQUESTS=100          # 每分钟请求数
RATE_LIMIT_BURST=20              # 突发请求数

# 上传限制
UPLOAD_RATE_LIMIT=10             # 每用户每小时上传数量限制
```

### 配置加载规范

- 使用 `godotenv` 加载 `.env` 文件
- 配置结构体定义在 `internal/config/config.go`
- 必须提供 `.env.example` 作为配置模板
- `.env` 文件添加到 `.gitignore`，禁止提交
- 生产环境通过系统环境变量或 Docker secrets 注入

## 代码规范

### Go 风格

- 遵循 Go 标准项目布局
- 使用 `internal/` 存放私有代码
- 错误处理使用 `errors.Wrap` 添加上下文
- 日志使用结构化日志（zap + sugared logger）
- 配置通过环境变量加载，禁止硬编码

### 命名规范

- 文件名：小写下划线 `photo_handler.go`
- 包名：小写单词 `handler`、`service`
- 接口名：动词或名词 `PhotoService`、`Repository`
- 常量：大写下划线 `MAX_UPLOAD_SIZE`

### API 设计

- RESTful 风格
- 版本前缀 `/api/v1/`
- 统一响应格式：
```json
{
  "code": 0,
  "message": "success",
  "data": {}
}
```
- 错误码规范化

### 数据库

- 使用 sqlx 进行原生 SQL 查询
- 迁移文件按时间戳命名
- 表名复数形式 `photos`、`users`
- 字段名小写下划线 `created_at`

## 国际化 (i18n)

- 支持多语言，前期：中文 (zh-CN)、英文 (en-US)
- 错误消息、提示文本均需国际化
- 语言包存放于 `internal/pkg/i18n/locales/`

## EXIF 信息采集

照片上传时需解析并存储以下 EXIF 字段（不限于此）：

### 相机信息
- Make（相机品牌）
- Model（相机型号）
- SerialNumber（机身序列号）

### 镜头信息
- LensModel（镜头型号）
- LensMake（镜头品牌）
- FocalLength（焦距）
- FocalLengthIn35mmFilm（等效 35mm 焦距）

### 拍摄参数
- ExposureTime（快门速度）
- FNumber（光圈值）
- ISO / ISOSpeedRatings（感光度）
- ExposureProgram（曝光程序）
- ExposureMode（曝光模式）
- MeteringMode（测光模式）
- WhiteBalance（白平衡）
- Flash（闪光灯状态）
- ExposureBiasValue（曝光补偿）

### 时间地点
- DateTimeOriginal（拍摄时间）
- GPSLatitude / GPSLongitude（GPS 坐标）
- GPSAltitude（海拔）

### 图像信息
- ImageWidth / ImageHeight（图像尺寸）
- Orientation（方向）
- ColorSpace（色彩空间）
- Software（处理软件）

## 图片格式支持

### 标准格式
- PNG
- JPG / JPEG

### RAW 格式
支持常见相机 RAW 格式：
- Canon: CR2, CR3
- Nikon: NEF, NRW
- Sony: ARW
- Fujifilm: RAF
- Olympus: ORF
- Panasonic: RW2
- Pentax: PEF, DNG
- Leica: DNG
- 通用: DNG

### 上传规则
- 若同时上传 RAW 和 JPG/PNG，两者必须对应同一张照片
- 通过 EXIF 时间戳和序列号验证对应关系

## AI 服务集成

- AI 审核为独立 Python 微服务
- 通过 HTTP API 调用
- 审核内容：
  - 图片质量（清晰度、构图、曝光）
  - 内容识别（是否为飞机、机型识别）
  - 违规检测（水印、敏感内容）
  - 注册号清晰度
  - 主体遮挡检测

## 相关仓库

- 前端：QuanPhotos-web
- AI 服务：QuanPhotos-ai