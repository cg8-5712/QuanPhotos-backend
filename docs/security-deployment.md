# 安全与部署注意事项

本文档记录 QuanPhotos 后端在安全和部署方面需要注意的事项。

## 安全功能

### 1. 限流中间件

位置：`internal/middleware/rate_limit.go`

#### 已实现的限流器

| 限流器 | 限制 | 应用场景 |
|--------|------|----------|
| `LoginRateLimiter` | 5次/分钟（每IP） | 登录接口，防止暴力破解 |
| `UploadRateLimiter` | 10次/小时（每用户） | 文件上传，防止滥用 |
| `APIRateLimiter` | 100次/分钟（每IP） | 全局 API 限流 |
| `GlobalRateLimiter` | 可配置 | 自定义限流场景 |

#### 部署注意事项

1. **当前实现是内存存储**：限流计数存储在内存中，这意味着：
   - 服务重启后计数重置
   - 多实例部署时每个实例独立计数

2. **生产环境建议**：
   - 如果需要多实例部署，建议实现基于 Redis 的分布式限流
   - 可以修改 `RateLimiter` 结构使用 Redis 作为后端存储

3. **配置调整**：
   ```go
   // 可根据实际需求调整限流参数
   middleware.NewRateLimiter(middleware.RateLimiterConfig{
       Limit:   100,          // 请求次数
       Window:  time.Minute,  // 时间窗口
       KeyFunc: middleware.IPKeyFunc, // 限流键
   })
   ```

### 2. 文件上传安全

位置：`internal/pkg/security/file_validator.go`

#### 已实现的安全检查

1. **文件扩展名验证**：只允许预定义的图片格式
2. **魔数（Magic Number）验证**：检查文件头确保内容与扩展名匹配
3. **MIME 类型验证**：验证 Content-Type 头
4. **路径遍历防护**：防止 `../` 等路径攻击
5. **文件名清理**：移除危险字符

#### 支持的文件格式

| 类型 | 扩展名 |
|------|--------|
| JPEG | `.jpg`, `.jpeg` |
| PNG | `.png` |
| GIF | `.gif` |
| WebP | `.webp` |
| BMP | `.bmp` |
| TIFF | `.tiff`, `.tif` |
| RAW | `.cr2`, `.cr3`, `.nef`, `.arw`, `.raf`, `.orf`, `.rw2`, `.dng` |

#### 使用方法

```go
import "QuanPhotos/internal/pkg/security"

// 创建图片验证器（50MB 限制）
validator := security.NewImageValidator(50 * 1024 * 1024)

// 验证上传文件
if err := validator.ValidateFile(fileHeader); err != nil {
    switch err {
    case security.ErrInvalidFileType:
        // 不支持的文件类型
    case security.ErrFileTooLarge:
        // 文件过大
    case security.ErrMagicNumberMismatch:
        // 文件内容与扩展名不匹配
    case security.ErrPathTraversal:
        // 检测到路径遍历攻击
    }
}

// 清理文件名
safeFilename := security.SanitizeFilename(originalFilename)
```

## 部署检查清单

### 环境变量配置

确保以下环境变量已正确配置：

```bash
# 必须配置
APP_ENV=production
JWT_SECRET=<强随机密钥，至少32字符>
DB_HOST=<数据库地址>
DB_PORT=5432
DB_NAME=quanphotos
DB_USER=<数据库用户>
DB_PASSWORD=<数据库密码>

# 存储配置
STORAGE_PATH=/var/lib/quanphotos/uploads
STORAGE_BASE_URL=https://your-domain.com/uploads
STORAGE_MAX_SIZE=52428800  # 50MB

# 可选配置
CORS_ENABLED=true
CORS_ALLOWED_ORIGINS=https://your-domain.com
```

### 安全检查清单

- [ ] JWT_SECRET 使用强随机密钥（不要使用默认值）
- [ ] 数据库密码使用强密码
- [ ] 生产环境禁用 Debug 模式
- [ ] 配置正确的 CORS 策略
- [ ] 上传目录有正确的文件权限
- [ ] 数据库连接使用 SSL（如适用）
- [ ] 配置 HTTPS（通过反向代理）

### 反向代理配置（Nginx 示例）

```nginx
server {
    listen 443 ssl http2;
    server_name your-domain.com;

    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;

    # 安全头
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;

    # 上传文件大小限制
    client_max_body_size 50M;

    # API 代理
    location /api/ {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # 静态文件（上传的图片）
    location /uploads/ {
        alias /var/lib/quanphotos/uploads/;
        expires 30d;
        add_header Cache-Control "public, immutable";
    }
}
```

### 数据库维护

1. **定期备份**：
   ```bash
   pg_dump -h localhost -U quanphotos quanphotos > backup_$(date +%Y%m%d).sql
   ```

2. **索引维护**：
   ```sql
   REINDEX DATABASE quanphotos;
   VACUUM ANALYZE;
   ```

## 监控建议

### 关键指标

1. **API 响应时间**：监控各端点的 P95/P99 延迟
2. **错误率**：4xx 和 5xx 错误的比率
3. **限流触发次数**：监控限流器的触发频率
4. **数据库连接池**：活跃连接数和等待队列
5. **磁盘使用**：上传目录的磁盘使用率

### 日志

应用使用 Zap 结构化日志，建议配置日志聚合：

```bash
# 日志输出到文件
./api 2>&1 | tee -a /var/log/quanphotos/app.log
```

## 未来改进建议

### 高优先级

1. **分布式限流**：使用 Redis 实现，支持多实例部署
2. **Token 黑名单**：Redis 存储已登出的 Token
3. **请求签名**：对敏感操作增加请求签名验证

### 中优先级

1. **CDN 集成**：图片资源使用 CDN 加速
2. **图片压缩**：上传时自动生成多种尺寸
3. **异步处理**：大文件上传和图片处理异步化

### 低优先级

1. **WebSocket**：实时消息推送
2. **全文搜索**：集成 Elasticsearch

---

*文档版本：v1.0*
*更新日期：2025-12-27*
