# QuanPhotos API 端点文档

## 基础信息

| 项目 | 说明 |
|------|------|
| Base URL | `http://localhost:8080/api/v1` |
| 认证方式 | Bearer Token (JWT) |
| 内容类型 | `application/json` |
| 文件上传 | `multipart/form-data` |

---

## 响应格式

### 成功响应

```json
{
  "code": 0,
  "message": "success",
  "data": { ... }
}
```

### 错误响应

```json
{
  "code": 40001,
  "message": "invalid parameters"
}
```

### 分页响应

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [ ... ],
    "pagination": {
      "page": 1,
      "page_size": 20,
      "total": 100,
      "total_pages": 5
    }
  }
}
```

---

## 错误码定义

| 错误码 | 说明 |
|--------|------|
| 0 | 成功 |
| 40001 | 参数错误 |
| 40002 | 参数验证失败 |
| 40101 | 未登录 |
| 40102 | Token 过期 |
| 40103 | Token 无效 |
| 40301 | 无权限 |
| 40401 | 资源不存在 |
| 40901 | 资源冲突（如用户名已存在） |
| 42201 | 文件格式不支持 |
| 42202 | 文件过大 |
| 42901 | 请求过于频繁 |
| 50001 | 服务器内部错误 |
| 50002 | 数据库错误 |
| 50003 | AI 服务不可用 |

---

## 认证相关 `/auth`

### 用户注册

```
POST /auth/register
```

**请求体**

```json
{
  "username": "string",      // 4-20 字符，字母数字下划线
  "email": "string",         // 有效邮箱
  "password": "string",      // 8-32 字符
  "confirm_password": "string"
}
```

**响应**

```json
{
  "code": 0,
  "message": "registration successful",
  "data": {
    "user_id": 1,
    "username": "aviator"
  }
}
```

**错误情况**
- `40002` 参数验证失败
- `40901` 用户名或邮箱已存在

---

### 用户登录

```
POST /auth/login
```

**请求体**

```json
{
  "username": "string",      // 用户名或邮箱
  "password": "string"
}
```

**响应**

```json
{
  "code": 0,
  "message": "login successful",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIs...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
    "expires_in": 3600,
    "token_type": "Bearer",
    "user": {
      "id": 1,
      "username": "aviator",
      "email": "user@example.com",
      "role": "user",
      "avatar": "https://..."
    }
  }
}
```

**错误情况**
- `40001` 用户名或密码错误
- `40301` 账号已被禁用

---

### 刷新 Token

```
POST /auth/refresh
```

**请求体**

```json
{
  "refresh_token": "string"
}
```

**响应**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIs...",
    "expires_in": 3600
  }
}
```

**错误情况**
- `40102` Refresh Token 过期
- `40103` Refresh Token 无效

---

### 退出登录

```
POST /auth/logout
```

**请求头**

```
Authorization: Bearer <access_token>
```

**响应**

```json
{
  "code": 0,
  "message": "logout successful"
}
```

---

## 用户相关 `/users`

### 获取当前用户信息

```
GET /users/me
```

**请求头**

```
Authorization: Bearer <access_token>
```

**响应**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "username": "aviator",
    "email": "user@example.com",
    "role": "user",
    "avatar": "https://...",
    "bio": "航空摄影爱好者",
    "location": "北京",
    "photo_count": 42,
    "favorite_count": 128,
    "created_at": "2025-01-01T00:00:00Z"
  }
}
```

---

### 更新当前用户信息

```
PUT /users/me
```

**请求头**

```
Authorization: Bearer <access_token>
```

**请求体**

```json
{
  "avatar": "string",        // 头像 URL（可选）
  "bio": "string",           // 个人简介（可选，最多 200 字）
  "location": "string"       // 所在地（可选）
}
```

**响应**

```json
{
  "code": 0,
  "message": "update successful",
  "data": { ... }            // 更新后的用户信息
}
```

---

### 修改密码

```
PUT /users/me/password
```

**请求头**

```
Authorization: Bearer <access_token>
```

**请求体**

```json
{
  "old_password": "string",
  "new_password": "string",
  "confirm_password": "string"
}
```

**响应**

```json
{
  "code": 0,
  "message": "password changed successfully"
}
```

**错误情况**
- `40001` 原密码错误

---

### 获取用户公开信息

```
GET /users/:id
```

**路径参数**

| 参数 | 类型 | 说明 |
|------|------|------|
| id | int | 用户 ID |

**响应**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "username": "aviator",
    "avatar": "https://...",
    "bio": "航空摄影爱好者",
    "photo_count": 42,
    "created_at": "2025-01-01T00:00:00Z"
  }
}
```

---

### 获取用户照片列表

```
GET /users/:id/photos
```

**路径参数**

| 参数 | 类型 | 说明 |
|------|------|------|
| id | int | 用户 ID |

**查询参数**

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| page | int | 否 | 1 | 页码 |
| page_size | int | 否 | 20 | 每页数量（最大 50）|

**响应**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [ ... ],         // 照片列表
    "pagination": { ... }
  }
}
```

---

## 照片相关 `/photos`

### 获取照片列表

```
GET /photos
```

**查询参数**

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| page | int | 否 | 1 | 页码 |
| page_size | int | 否 | 20 | 每页数量（最大 50）|
| sort | string | 否 | newest | 排序：newest/oldest/popular |
| category_id | int | 否 | - | 分类 ID |
| aircraft_type | string | 否 | - | 机型筛选 |
| airline | string | 否 | - | 航空公司 |
| airport | string | 否 | - | 机场代码 |
| registration | string | 否 | - | 注册号 |
| keyword | string | 否 | - | 关键词搜索 |
| user_id | int | 否 | - | 指定用户 |
| date_from | string | 否 | - | 拍摄日期起始（YYYY-MM-DD）|
| date_to | string | 否 | - | 拍摄日期结束（YYYY-MM-DD）|

**响应**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [
      {
        "id": 1,
        "title": "Boeing 787-9 着陆",
        "thumbnail_url": "https://.../thumb/1.jpg",
        "user": {
          "id": 1,
          "username": "aviator",
          "avatar": "https://..."
        },
        "aircraft_type": "Boeing 787-9",
        "airline": "中国国际航空",
        "registration": "B-1234",
        "view_count": 1024,
        "favorite_count": 128,
        "created_at": "2025-01-01T12:00:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "page_size": 20,
      "total": 1000,
      "total_pages": 50
    }
  }
}
```

---

### 上传照片

```
POST /photos
```

**请求头**

```
Authorization: Bearer <access_token>
Content-Type: multipart/form-data
```

**请求体（Form Data）**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| file | file | 是 | 照片文件（JPG/PNG/RAW）|
| raw_file | file | 否 | RAW 文件（与 file 对应）|
| title | string | 是 | 标题（最多 100 字）|
| description | string | 否 | 描述（最多 500 字）|
| aircraft_type | string | 否 | 机型 |
| airline | string | 否 | 航空公司 |
| registration | string | 否 | 注册号 |
| airport | string | 否 | 拍摄机场（ICAO/IATA）|
| category_id | int | 否 | 分类 ID |
| tags | string | 否 | 标签，逗号分隔 |

**响应**

```json
{
  "code": 0,
  "message": "upload successful, pending review",
  "data": {
    "id": 123,
    "status": "pending",
    "title": "Boeing 787-9 着陆"
  }
}
```

**错误情况**
- `42201` 文件格式不支持
- `42202` 文件过大（超过 50MB）
- `42901` 上传过于频繁

---

### 获取照片详情

```
GET /photos/:id
```

**路径参数**

| 参数 | 类型 | 说明 |
|------|------|------|
| id | int | 照片 ID |

**响应**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "title": "Boeing 787-9 着陆",
    "description": "2025年1月1日拍摄于北京首都机场",
    "image_url": "https://.../photos/1.jpg",
    "thumbnail_url": "https://.../thumb/1.jpg",
    "has_raw": true,
    "status": "approved",
    "user": {
      "id": 1,
      "username": "aviator",
      "avatar": "https://..."
    },
    "aircraft_type": "Boeing 787-9",
    "airline": "中国国际航空",
    "registration": "B-1234",
    "airport": "ZBAA",
    "category": {
      "id": 1,
      "name": "民航客机"
    },
    "tags": ["787", "国航", "首都机场"],
    "exif": {
      "camera_make": "Canon",
      "camera_model": "EOS R5",
      "lens_model": "RF 100-500mm F4.5-7.1 L IS USM",
      "focal_length": "500mm",
      "focal_length_35mm": "500mm",
      "aperture": "f/7.1",
      "shutter_speed": "1/2000",
      "iso": 400,
      "exposure_mode": "Manual",
      "metering_mode": "Evaluative",
      "white_balance": "Auto",
      "flash": "Off",
      "taken_at": "2025-01-01T10:30:00Z",
      "gps_latitude": 40.0799,
      "gps_longitude": 116.6031,
      "image_width": 8192,
      "image_height": 5464
    },
    "view_count": 1024,
    "favorite_count": 128,
    "is_favorited": false,
    "created_at": "2025-01-01T12:00:00Z",
    "approved_at": "2025-01-01T14:00:00Z"
  }
}
```

---

### 删除照片

```
DELETE /photos/:id
```

**请求头**

```
Authorization: Bearer <access_token>
```

**响应**

```json
{
  "code": 0,
  "message": "deleted successfully"
}
```

**错误情况**
- `40301` 无权限（非本人照片）
- `40401` 照片不存在

---

### 收藏照片

```
POST /photos/:id/favorite
```

**请求头**

```
Authorization: Bearer <access_token>
```

**响应**

```json
{
  "code": 0,
  "message": "added to favorites",
  "data": {
    "favorite_count": 129
  }
}
```

---

### 取消收藏

```
DELETE /photos/:id/favorite
```

**请求头**

```
Authorization: Bearer <access_token>
```

**响应**

```json
{
  "code": 0,
  "message": "removed from favorites",
  "data": {
    "favorite_count": 128
  }
}
```

---

### 获取我的收藏

```
GET /photos/favorites
```

**请求头**

```
Authorization: Bearer <access_token>
```

**查询参数**

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| page | int | 否 | 1 | 页码 |
| page_size | int | 否 | 20 | 每页数量 |

**响应**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [ ... ],
    "pagination": { ... }
  }
}
```

---

### 获取我的上传

```
GET /photos/mine
```

**请求头**

```
Authorization: Bearer <access_token>
```

**查询参数**

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| page | int | 否 | 1 | 页码 |
| page_size | int | 否 | 20 | 每页数量 |
| status | string | 否 | all | 状态筛选：all/pending/approved/rejected |

**响应**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [ ... ],
    "pagination": { ... }
  }
}
```

---

## 分类相关 `/categories`

### 获取分类列表

```
GET /categories
```

**查询参数**

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| include_empty | bool | 否 | false | 是否包含无照片的分类 |

**响应**

```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": 1,
      "name": "民航客机",
      "name_en": "Commercial Aircraft",
      "description": "商业航班客机",
      "photo_count": 5000,
      "sort_order": 1,
      "created_at": "2025-01-01T00:00:00Z"
    },
    {
      "id": 2,
      "name": "货运飞机",
      "name_en": "Cargo Aircraft",
      "description": "货运航班飞机",
      "photo_count": 1200,
      "sort_order": 2,
      "created_at": "2025-01-01T00:00:00Z"
    }
  ]
}
```

---

### 获取分类详情

```
GET /categories/:id
```

**路径参数**

| 参数 | 类型 | 说明 |
|------|------|------|
| id | int | 分类 ID |

**响应**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "name": "民航客机",
    "name_en": "Commercial Aircraft",
    "description": "商业航班客机",
    "photo_count": 5000,
    "sort_order": 1,
    "created_at": "2025-01-01T00:00:00Z",
    "updated_at": "2025-01-01T00:00:00Z"
  }
}
```

**错误情况**
- `40401` 分类不存在

---

### 创建分类（管理员）

```
POST /categories
```

**请求头**

```
Authorization: Bearer <access_token>
```

**请求体**

```json
{
  "name": "公务机",
  "name_en": "Business Jet",
  "description": "私人及公务飞机",
  "sort_order": 10
}
```

**响应**

```json
{
  "code": 0,
  "message": "category created",
  "data": {
    "id": 10,
    "name": "公务机",
    "name_en": "Business Jet",
    "description": "私人及公务飞机",
    "sort_order": 10,
    "created_at": "2025-01-01T12:00:00Z"
  }
}
```

**错误情况**
- `40002` 参数验证失败
- `40301` 无权限（需要管理员）
- `40901` 分类名称已存在

---

### 更新分类（管理员）

```
PUT /categories/:id
```

**请求头**

```
Authorization: Bearer <access_token>
```

**路径参数**

| 参数 | 类型 | 说明 |
|------|------|------|
| id | int | 分类 ID |

**请求体**

```json
{
  "name": "公务机",
  "name_en": "Business Jet",
  "description": "私人及公务飞机（更新）",
  "sort_order": 5
}
```

**响应**

```json
{
  "code": 0,
  "message": "category updated",
  "data": {
    "id": 10,
    "name": "公务机",
    "name_en": "Business Jet",
    "description": "私人及公务飞机（更新）",
    "sort_order": 5,
    "updated_at": "2025-01-02T10:00:00Z"
  }
}
```

**错误情况**
- `40301` 无权限
- `40401` 分类不存在
- `40901` 分类名称已存在

---

### 删除分类（管理员）

```
DELETE /categories/:id
```

**请求头**

```
Authorization: Bearer <access_token>
```

**路径参数**

| 参数 | 类型 | 说明 |
|------|------|------|
| id | int | 分类 ID |

**查询参数**

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| move_to | int | 否 | - | 将该分类下照片移至目标分类 ID |

**响应**

```json
{
  "code": 0,
  "message": "category deleted"
}
```

**错误情况**
- `40001` 分类下有照片且未指定 move_to
- `40301` 无权限
- `40401` 分类不存在

---

## 标签相关 `/tags`

### 获取热门标签

```
GET /tags
```

**查询参数**

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| limit | int | 否 | 50 | 返回数量（最大 100）|

**响应**

```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": 1,
      "name": "787",
      "photo_count": 1200
    },
    {
      "id": 2,
      "name": "首都机场",
      "photo_count": 800
    }
  ]
}
```

---

### 搜索标签

```
GET /tags/search
```

**查询参数**

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| keyword | string | 是 | - | 搜索关键词 |
| limit | int | 否 | 20 | 返回数量 |

**响应**

```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": 1,
      "name": "787-9",
      "photo_count": 500
    },
    {
      "id": 2,
      "name": "787-10",
      "photo_count": 200
    }
  ]
}
```

---

### 获取标签下的照片

```
GET /tags/:id/photos
```

**路径参数**

| 参数 | 类型 | 说明 |
|------|------|------|
| id | int | 标签 ID |

**查询参数**

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| page | int | 否 | 1 | 页码 |
| page_size | int | 否 | 20 | 每页数量 |

**响应**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "tag": {
      "id": 1,
      "name": "787",
      "photo_count": 1200
    },
    "list": [ ... ],
    "pagination": { ... }
  }
}
```

---

## 工单相关 `/tickets`

### 创建工单

```
POST /tickets
```

**请求头**

```
Authorization: Bearer <access_token>
```

**请求体**

```json
{
  "photo_id": 123,           // 关联的照片 ID
  "type": "appeal",          // 工单类型：appeal（申诉）/ report（举报）/ other
  "title": "申诉 AI 审核结果",
  "content": "我认为这张照片符合要求，AI 审核有误..."
}
```

**响应**

```json
{
  "code": 0,
  "message": "ticket created",
  "data": {
    "id": 456,
    "status": "open",
    "created_at": "2025-01-01T12:00:00Z"
  }
}
```

---

### 获取我的工单列表

```
GET /tickets
```

**请求头**

```
Authorization: Bearer <access_token>
```

**查询参数**

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| page | int | 否 | 1 | 页码 |
| page_size | int | 否 | 20 | 每页数量 |
| status | string | 否 | all | 状态：all/open/processing/resolved/closed |

**响应**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [
      {
        "id": 456,
        "type": "appeal",
        "title": "申诉 AI 审核结果",
        "status": "processing",
        "photo_id": 123,
        "created_at": "2025-01-01T12:00:00Z",
        "updated_at": "2025-01-02T10:00:00Z"
      }
    ],
    "pagination": { ... }
  }
}
```

---

### 获取工单详情

```
GET /tickets/:id
```

**请求头**

```
Authorization: Bearer <access_token>
```

**响应**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 456,
    "type": "appeal",
    "title": "申诉 AI 审核结果",
    "content": "我认为这张照片符合要求...",
    "status": "processing",
    "photo": {
      "id": 123,
      "title": "Boeing 787-9 着陆",
      "thumbnail_url": "https://..."
    },
    "replies": [
      {
        "id": 1,
        "content": "您好，我们已收到您的申诉...",
        "user": {
          "id": 100,
          "username": "admin",
          "role": "admin"
        },
        "created_at": "2025-01-02T10:00:00Z"
      }
    ],
    "created_at": "2025-01-01T12:00:00Z",
    "updated_at": "2025-01-02T10:00:00Z"
  }
}
```

---

### 回复工单

```
POST /tickets/:id/replies
```

**请求头**

```
Authorization: Bearer <access_token>
```

**请求体**

```json
{
  "content": "补充说明：这张照片是在..."
}
```

**响应**

```json
{
  "code": 0,
  "message": "reply submitted",
  "data": {
    "id": 2,
    "content": "补充说明：这张照片是在...",
    "created_at": "2025-01-02T14:00:00Z"
  }
}
```

---

## 管理接口 `/admin`

> 以下接口需要管理员或超级管理员权限

### 获取待审核列表

```
GET /admin/reviews
```

**请求头**

```
Authorization: Bearer <access_token>
```

**查询参数**

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| page | int | 否 | 1 | 页码 |
| page_size | int | 否 | 20 | 每页数量 |
| status | string | 否 | pending | 状态：pending/ai_passed/ai_rejected |

**响应**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [
      {
        "id": 123,
        "title": "Boeing 787-9 着陆",
        "thumbnail_url": "https://...",
        "status": "ai_passed",
        "ai_result": {
          "score": 0.85,
          "aircraft_detected": true,
          "quality_score": 0.9,
          "issues": []
        },
        "user": {
          "id": 1,
          "username": "aviator"
        },
        "created_at": "2025-01-01T12:00:00Z"
      }
    ],
    "pagination": { ... }
  }
}
```

---

### 审核照片

```
POST /admin/reviews/:id
```

**请求头**

```
Authorization: Bearer <access_token>
```

**请求体**

```json
{
  "action": "approve",       // approve / reject
  "reason": "string"         // 拒绝原因（拒绝时必填）
}
```

**响应**

```json
{
  "code": 0,
  "message": "review completed",
  "data": {
    "id": 123,
    "status": "approved",
    "reviewed_by": 100,
    "reviewed_at": "2025-01-01T14:00:00Z"
  }
}
```

---

### 获取用户列表

```
GET /admin/users
```

**查询参数**

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| page | int | 否 | 1 | 页码 |
| page_size | int | 否 | 20 | 每页数量 |
| role | string | 否 | all | 角色筛选 |
| status | string | 否 | all | 状态：all/active/banned |
| keyword | string | 否 | - | 搜索用户名/邮箱 |

**响应**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [
      {
        "id": 1,
        "username": "aviator",
        "email": "user@example.com",
        "role": "user",
        "status": "active",
        "photo_count": 42,
        "created_at": "2025-01-01T00:00:00Z",
        "last_login_at": "2025-01-10T12:00:00Z"
      }
    ],
    "pagination": { ... }
  }
}
```

---

### 修改用户角色

```
PUT /admin/users/:id/role
```

**请求头**

```
Authorization: Bearer <access_token>
```

**请求体**

```json
{
  "role": "reviewer"         // user / reviewer / admin
}
```

**响应**

```json
{
  "code": 0,
  "message": "role updated"
}
```

---

### 封禁/解封用户

```
PUT /admin/users/:id/status
```

**请求头**

```
Authorization: Bearer <access_token>
```

**请求体**

```json
{
  "status": "banned",        // active / banned
  "reason": "违规上传"        // 封禁原因（封禁时必填）
}
```

**响应**

```json
{
  "code": 0,
  "message": "operation successful"
}
```

---

### 删除照片（管理员）

```
DELETE /admin/photos/:id
```

**请求头**

```
Authorization: Bearer <access_token>
```

**请求体**

```json
{
  "reason": "违规内容"        // 删除原因
}
```

**响应**

```json
{
  "code": 0,
  "message": "deleted successfully"
}
```

---

### 获取工单列表（管理员）

```
GET /admin/tickets
```

**查询参数**

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| page | int | 否 | 1 | 页码 |
| page_size | int | 否 | 20 | 每页数量 |
| status | string | 否 | open | 状态筛选 |
| type | string | 否 | all | 类型筛选 |

**响应**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [ ... ],
    "pagination": { ... }
  }
}
```

---

### 处理工单

```
PUT /admin/tickets/:id
```

**请求头**

```
Authorization: Bearer <access_token>
```

**请求体**

```json
{
  "status": "resolved",      // processing / resolved / closed
  "reply": "您的申诉已通过..."  // 回复内容（可选）
}
```

**响应**

```json
{
  "code": 0,
  "message": "processed successfully"
}
```

---

## 系统接口 `/system`

### 健康检查

```
GET /health
```

> 无需认证

**响应**

```json
{
  "status": "ok",
  "timestamp": "2025-01-01T12:00:00Z"
}
```

---

### 获取系统信息

```
GET /system/info
```

> 无需认证

**响应**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "version": "1.0.0",
    "supported_formats": ["jpg", "jpeg", "png", "cr2", "cr3", "nef", "arw"],
    "max_upload_size": 52428800,
    "languages": ["zh-CN", "en-US"]
  }
}
```

---

## 请求头说明

| Header | 说明 | 示例 |
|--------|------|------|
| Authorization | 认证令牌 | `Bearer eyJhbGci...` |
| Accept-Language | 语言偏好 | `zh-CN` / `en-US` |
| Content-Type | 内容类型 | `application/json` |
| X-Request-ID | 请求追踪 ID | `uuid` |

---

## 状态枚举

### 照片状态 `photo_status`

| 值 | 说明 |
|----|------|
| pending | 待审核 |
| ai_passed | AI 初审通过 |
| ai_rejected | AI 初审拒绝 |
| approved | 已发布 |
| rejected | 已拒绝 |

### 用户角色 `user_role`

| 值 | 说明 |
|----|------|
| guest | 游客（未登录）|
| user | 普通用户 |
| reviewer | 审查员（可审核照片）|
| admin | 管理员（最高权限）|

### 工单状态 `ticket_status`

| 值 | 说明 |
|----|------|
| open | 待处理 |
| processing | 处理中 |
| resolved | 已解决 |
| closed | 已关闭 |

---

*文档版本：v1.0*
*创建日期：2025-12-22*
