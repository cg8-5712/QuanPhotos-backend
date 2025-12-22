# QuanPhotos 照片存储设计文档

## 概述

本文档描述 QuanPhotos 后端服务的照片存储设计方案，包括目录结构、文件命名、缩略图生成、RAW 文件处理等。

---

## 目录结构

```
uploads/
├── photos/                    # 原图（压缩后的展示图）
│   └── {year}/
│       └── {month}/
│           └── {day}/
│               ├── {uuid}.jpg
│               └── {uuid}.png
├── raw/                       # RAW 原始文件
│   └── {year}/
│       └── {month}/
│           └── {day}/
│               ├── {uuid}.cr3
│               └── {uuid}.nef
├── thumbnails/                # 缩略图
│   └── {year}/
│       └── {month}/
│           └── {day}/
│               ├── {uuid}_sm.jpg    # 小图
│               ├── {uuid}_md.jpg    # 中图
│               └── {uuid}_lg.jpg    # 大图
└── temp/                      # 临时文件（上传中）
    └── {uuid}.tmp
```

### 目录说明

| 目录 | 说明 |
|------|------|
| photos/ | 压缩处理后的展示用原图 |
| raw/ | RAW 格式原始文件，仅供下载 |
| thumbnails/ | 各尺寸缩略图 |
| temp/ | 上传过程中的临时文件 |

---

## 文件命名规则

### 命名格式

```
{uuid}.{extension}

示例：
550e8400-e29b-41d4-a716-446655440000.jpg
6ba7b810-9dad-11d1-80b4-00c04fd430c8.cr3
```

### 缩略图命名

```
{uuid}_{size}.jpg

尺寸后缀：
- _sm  小图
- _md  中图
- _lg  大图

示例：
550e8400-e29b-41d4-a716-446655440000_sm.jpg
550e8400-e29b-41d4-a716-446655440000_md.jpg
550e8400-e29b-41d4-a716-446655440000_lg.jpg
```

### 使用 UUID 的优势

1. **唯一性**：避免文件名冲突
2. **安全性**：无法通过文件名猜测其他文件
3. **解耦**：与原始文件名无关
4. **一致性**：统一的命名规范

---

## 缩略图规格

| 类型 | 代码 | 尺寸 | 质量 | 用途 |
|------|------|------|------|------|
| 小图 | sm | 300×200 | 80% | 列表页、搜索结果 |
| 中图 | md | 800×533 | 85% | 详情页预览 |
| 大图 | lg | 1600×1067 | 90% | Lightbox 浏览 |
| 原图 | - | 最大 4096px | 92% | 高清查看/下载 |

### 缩略图生成规则

1. **保持宽高比**：按比例缩放，不拉伸变形
2. **居中裁剪**：超出部分从中心裁剪
3. **格式统一**：所有缩略图使用 JPEG 格式
4. **渐进式**：使用渐进式 JPEG 提升加载体验

---

## 上传处理流程

```
┌──────────────────────────────────────────────────────────────────┐
│                         上传处理流程                              │
└──────────────────────────────────────────────────────────────────┘

  ┌─────────┐
  │ 用户上传 │
  └────┬────┘
       │
       ▼
  ┌─────────────────┐
  │ 1. 保存到 temp/ │
  │    生成临时文件  │
  └────────┬────────┘
           │
           ▼
  ┌─────────────────┐     ┌─────────────┐
  │ 2. 文件验证     │────►│ 验证失败    │──► 删除临时文件，返回错误
  │ - 格式检查      │     │ 拒绝上传    │
  │ - 大小检查      │     └─────────────┘
  │ - 安全检查      │
  └────────┬────────┘
           │ 验证通过
           ▼
  ┌─────────────────┐
  │ 3. EXIF 解析    │
  │ - 相机信息      │
  │ - 拍摄参数      │
  │ - GPS 坐标      │
  │ - 拍摄时间      │
  └────────┬────────┘
           │
           ▼
  ┌─────────────────┐
  │ 4. 图像处理     │
  │ - 方向校正      │
  │ - 尺寸压缩      │
  │ - 质量优化      │
  └────────┬────────┘
           │
           ▼
  ┌─────────────────┐
  │ 5. 生成缩略图   │
  │ - sm (300x200)  │
  │ - md (800x533)  │
  │ - lg (1600x1067)│
  └────────┬────────┘
           │
           ▼
  ┌─────────────────┐
  │ 6. 移动到正式目录│
  │ photos/{date}/  │
  │ thumbnails/     │
  └────────┬────────┘
           │
           ▼
  ┌─────────────────┐
  │ 7. 写入数据库   │
  │ - 文件路径      │
  │ - EXIF 数据     │
  │ - 状态: pending │
  └────────┬────────┘
           │
           ▼
  ┌─────────────────┐
  │ 8. 触发 AI 审核 │
  └─────────────────┘
```

---

## 数据库存储字段

```sql
-- photos 表文件相关字段

file_path       VARCHAR(500)  NOT NULL  -- 原图路径
thumbnail_path  VARCHAR(500)            -- 缩略图路径（不含尺寸后缀）
raw_file_path   VARCHAR(500)            -- RAW 文件路径（可为空）
file_size       BIGINT                  -- 原图文件大小 (bytes)
```

### 路径存储示例

| 字段 | 值 |
|------|-----|
| file_path | /photos/2025/01/22/550e8400-e29b-41d4-a716-446655440000.jpg |
| thumbnail_path | /thumbnails/2025/01/22/550e8400-e29b-41d4-a716-446655440000 |
| raw_file_path | /raw/2025/01/22/550e8400-e29b-41d4-a716-446655440000.cr3 |

### 缩略图 URL 拼接

```go
// 获取不同尺寸缩略图
smallThumb := photo.ThumbnailPath + "_sm.jpg"   // xxx_sm.jpg
mediumThumb := photo.ThumbnailPath + "_md.jpg"  // xxx_md.jpg
largeThumb := photo.ThumbnailPath + "_lg.jpg"   // xxx_lg.jpg
```

---

## 访问 URL

### 静态文件访问

```
# 原图
GET /uploads/photos/2025/01/22/{uuid}.jpg

# 缩略图
GET /uploads/thumbnails/2025/01/22/{uuid}_sm.jpg
GET /uploads/thumbnails/2025/01/22/{uuid}_md.jpg
GET /uploads/thumbnails/2025/01/22/{uuid}_lg.jpg
```

### API 访问

```
# 获取照片信息（含各尺寸 URL）
GET /api/v1/photos/{id}

Response:
{
  "data": {
    "id": 1,
    "image_url": "https://cdn.quanphotos.com/photos/2025/01/22/uuid.jpg",
    "thumbnail_url": "https://cdn.quanphotos.com/thumbnails/2025/01/22/uuid_md.jpg",
    "thumbnails": {
      "sm": "https://cdn.quanphotos.com/thumbnails/2025/01/22/uuid_sm.jpg",
      "md": "https://cdn.quanphotos.com/thumbnails/2025/01/22/uuid_md.jpg",
      "lg": "https://cdn.quanphotos.com/thumbnails/2025/01/22/uuid_lg.jpg"
    },
    "has_raw": true
  }
}

# 下载 RAW 文件（需要权限）
GET /api/v1/photos/{id}/raw
```

---

## RAW 文件处理

### 支持的 RAW 格式

| 格式 | 厂商 | 处理方式 |
|------|------|----------|
| CR2 | Canon | 提取内嵌 JPEG / dcraw 转换 |
| CR3 | Canon | 提取内嵌 JPEG / libraw |
| NEF | Nikon | 提取内嵌 JPEG / dcraw 转换 |
| NRW | Nikon | 提取内嵌 JPEG |
| ARW | Sony | 提取内嵌 JPEG |
| RAF | Fujifilm | 提取内嵌 JPEG / dcraw |
| ORF | Olympus | 提取内嵌 JPEG |
| RW2 | Panasonic | 提取内嵌 JPEG |
| DNG | 通用 | libraw 处理 |

### RAW + JPG 配对上传

```
上传规则：
1. RAW 文件必须和 JPG/PNG 配对上传
2. 通过 EXIF DateTimeOriginal 验证是否为同一张照片
3. 时间戳差异容忍范围：±2 秒
4. RAW 仅存储供下载，展示使用配对的 JPG
```

### RAW 处理流程

```
┌───────────────┐     ┌───────────────┐
│ RAW + JPG 上传│────►│ 验证配对关系  │
└───────────────┘     └───────┬───────┘
                              │
                    ┌─────────┴─────────┐
                    │                   │
                    ▼                   ▼
              ┌───────────┐       ┌───────────┐
              │ 配对成功  │       │ 配对失败  │
              └─────┬─────┘       └─────┬─────┘
                    │                   │
                    ▼                   ▼
              ┌───────────┐       ┌───────────┐
              │ RAW 存储  │       │ 仅处理 JPG│
              │ raw/{date}│       │ 丢弃 RAW  │
              └─────┬─────┘       └───────────┘
                    │
                    ▼
              ┌───────────┐
              │ JPG 正常  │
              │ 处理流程  │
              └───────────┘
```

---

## 环境配置

```env
# 存储基础配置
STORAGE_TYPE=local                    # 存储类型: local / oss / s3
STORAGE_PATH=./uploads                # 本地存储根目录
STORAGE_MAX_SIZE=52428800             # 单文件最大 50MB
STORAGE_ALLOWED_TYPES=jpg,jpeg,png,cr2,cr3,nef,arw,raf,orf,rw2,dng

# 缩略图配置
THUMB_SIZE_SM=300x200                 # 小图尺寸
THUMB_SIZE_MD=800x533                 # 中图尺寸
THUMB_SIZE_LG=1600x1067               # 大图尺寸
THUMB_QUALITY_SM=80                   # 小图质量
THUMB_QUALITY_MD=85                   # 中图质量
THUMB_QUALITY_LG=90                   # 大图质量

# 原图处理
IMAGE_MAX_DIMENSION=4096              # 原图最大边长
IMAGE_QUALITY=92                      # 原图压缩质量

# CDN 配置（可选）
CDN_ENABLED=false
CDN_BASE_URL=https://cdn.quanphotos.com
```

---

## 文件清理策略

| 文件类型 | 清理规则 | 执行频率 |
|----------|----------|----------|
| temp/ 临时文件 | 创建后 24 小时未使用 | 每小时 |
| 软删除照片 | 删除后 30 天 | 每天凌晨 |
| 孤立文件 | 无数据库记录的文件 | 每周日凌晨 |
| 过期 Token | refresh_tokens 过期记录 | 每天 |

### 清理任务示例

```go
// 清理临时文件
func CleanTempFiles() {
    threshold := time.Now().Add(-24 * time.Hour)
    // 删除 temp/ 下修改时间早于 threshold 的文件
}

// 清理软删除照片
func CleanDeletedPhotos() {
    threshold := time.Now().Add(-30 * 24 * time.Hour)
    // 查询 deleted_at < threshold 的照片
    // 删除对应的 photos/, thumbnails/, raw/ 文件
    // 物理删除数据库记录
}
```

---

## 存储扩展

### 云存储支持（后续）

```go
// Storage 接口定义
type Storage interface {
    Upload(ctx context.Context, file io.Reader, path string) error
    Delete(ctx context.Context, path string) error
    GetURL(path string) string
    Exists(ctx context.Context, path string) bool
}

// 实现
type LocalStorage struct { ... }
type OSSStorage struct { ... }    // 阿里云 OSS
type S3Storage struct { ... }     // AWS S3
```

### 迁移到云存储

1. 配置云存储凭证
2. 修改 `STORAGE_TYPE` 为 `oss` 或 `s3`
3. 运行迁移脚本同步现有文件
4. 更新数据库中的文件路径或配置 CDN

---

## 安全措施

1. **文件类型验证**：检查 MIME 类型和文件头，不仅依赖扩展名
2. **文件大小限制**：单文件最大 50MB
3. **路径安全**：禁止 `..` 等目录遍历
4. **访问控制**：RAW 文件下载需要登录
5. **防盗链**：配置 Referer 白名单（可选）
6. **水印**：可选添加水印保护版权

---

*文档版本：v1.0*
*创建日期：2025-12-22*
