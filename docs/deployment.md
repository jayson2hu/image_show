# Image Show 容器化部署文档

> 更新日期：2026-04-30  
> 适用范围：后续使用 Docker / docker-compose 部署 `image_show` 单体服务。  
> 当前项目采用 Go 后端内嵌 Vue 构建产物，容器运行时只需要启动一个二进制服务。

## 1. 部署架构

推荐生产部署形态：

```text
Nginx / Caddy / Cloudflare
        |
        v
image-show:3000
        |
        +-- PostgreSQL
        +-- Redis
        +-- Sub2API
        +-- Cloudflare R2
        +-- SMTP
        +-- EPay / Turnstile / WeChat Server
```

说明：

- `image-show` 同时提供 API 和前端页面，不需要单独部署 Vite dev server。
- 前端通过 `go:embed` 打进 Go 二进制，访问 `/`、`/login`、`/admin` 等路径由后端返回静态资源。
- `/api/*` 走 Gin API。
- 本地可以用 SQLite，生产建议用 PostgreSQL。
- Redis 可选，但生产建议启用，用于限流等能力。

## 2. 构建镜像

项目根目录建议准备以下 `Dockerfile`：

```dockerfile
# Stage 1: build frontend
FROM node:20-alpine AS frontend
WORKDIR /app/web
RUN corepack enable
COPY web/package.json web/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile
COPY web/ ./
RUN pnpm build

# Stage 2: build backend
FROM golang:1.22-alpine AS backend
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
COPY --from=frontend /app/web/dist ./web/dist
RUN CGO_ENABLED=0 go build -o image-show .

# Stage 3: runtime
FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=backend /app/image-show ./image-show
EXPOSE 3000
CMD ["./image-show"]
```

构建命令：

```powershell
docker build -t image-show:latest .
```

## 3. 推荐 docker-compose

如果复用 `sub2api-network`、外部 PostgreSQL 和 Redis，可使用：

```yaml
services:
  image-show:
    image: image-show:latest
    container_name: image-show
    restart: unless-stopped
    ports:
      - "3000:3000"
    env_file:
      - .env.production
    environment:
      - TZ=Asia/Shanghai
    networks:
      - sub2api-network
    healthcheck:
      test: ["CMD", "wget", "-q", "-T", "5", "-O", "/dev/null", "http://localhost:3000/health"]
      interval: 30s
      timeout: 10s
      retries: 3

networks:
  sub2api-network:
    external: true
```

如果需要同时启动独立 PostgreSQL 和 Redis，可使用：

```yaml
services:
  postgres:
    image: postgres:16-alpine
    container_name: image-show-postgres
    restart: unless-stopped
    environment:
      POSTGRES_DB: image_show
      POSTGRES_USER: image_show
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
      TZ: Asia/Shanghai
    volumes:
      - image-show-postgres:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    container_name: image-show-redis
    restart: unless-stopped
    command: ["redis-server", "--appendonly", "yes", "--requirepass", "${REDIS_PASSWORD}"]
    volumes:
      - image-show-redis:/data

  image-show:
    image: image-show:latest
    container_name: image-show
    restart: unless-stopped
    depends_on:
      - postgres
      - redis
    ports:
      - "3000:3000"
    env_file:
      - .env.production
    environment:
      DB_DRIVER: postgres
      DATABASE_DSN: postgres://image_show:${POSTGRES_PASSWORD}@postgres:5432/image_show?sslmode=disable
      REDIS_ADDR: redis:6379
      REDIS_PASSWORD: ${REDIS_PASSWORD}
      REDIS_DB: 1
      TZ: Asia/Shanghai
    healthcheck:
      test: ["CMD", "wget", "-q", "-T", "5", "-O", "/dev/null", "http://localhost:3000/health"]
      interval: 30s
      timeout: 10s
      retries: 3

volumes:
  image-show-postgres:
  image-show-redis:
```

启动命令：

```powershell
docker compose --env-file .env.production up -d
```

## 4. 生产环境变量

建议从 `.env.example` 复制：

```powershell
Copy-Item .env.example .env.production
```

核心必填项：

```env
APP_ENV=production
PORT=3000

DB_DRIVER=postgres
DATABASE_DSN=postgres://image_show:${POSTGRES_PASSWORD}@postgres:5432/image_show?sslmode=disable

REDIS_ADDR=redis:6379
REDIS_PASSWORD=change-me
REDIS_DB=1

JWT_SECRET=change-to-long-random-secret

ADMIN_EMAIL=admin@example.com
ADMIN_PASSWORD=change-to-strong-admin-password

SUB2API_BASE_URL=http://sub2api:8080

R2_ENDPOINT=https://xxx.r2.cloudflarestorage.com
R2_ACCESS_KEY=
R2_SECRET_KEY=
R2_BUCKET=image-show
R2_PUBLIC_URL=https://cdn.yourdomain.com

SMTP_HOST=smtp.qq.com
SMTP_PORT=465
SMTP_USER=
SMTP_PASSWORD=
SMTP_FROM=noreply@yourdomain.com

SERVER_ADDRESS=https://yourdomain.com
PAY_ADDRESS=https://your-epay-gateway.example.com
EPAY_ID=
EPAY_KEY=
EPAY_PAY_METHODS=alipay,wxpay

TURNSTILE_SITE_KEY=
TURNSTILE_SECRET=

WECHAT_AUTH_ENABLED=false
WECHAT_SERVER_ADDRESS=
WECHAT_SERVER_TOKEN=
WECHAT_QRCODE_URL=

MOCK_SUB2API=false
```

关键说明：

- `SERVER_ADDRESS` 必须是公网访问地址，支付回调和浏览器返回会依赖它。
- `JWT_SECRET` 生产必须改成强随机值，不能用默认值。
- `ADMIN_EMAIL` / `ADMIN_PASSWORD` 用于首次启动或重置管理员账号；生产环境必须显式配置才会写入管理员。
- `MOCK_SUB2API=false`，否则会返回测试图片。
- `R2_PUBLIC_URL` 如果配置 CDN 域名，图片访问会直接走 CDN；为空时后端使用临时签名 URL。
- `EPAY_PAY_METHODS` 当前默认支持 `alipay,wxpay`。前端当前默认创建 `alipay` 订单。

## 5. 数据库与迁移

应用启动时会自动执行 GORM `AutoMigrate`，会创建或补齐以下核心表：

- `users`
- `generations`
- `credit_logs`
- `login_logs`
- `channels`
- `settings`
- `prompt_templates`
- `packages`
- `orders`
- `anonymous_identities`

部署前建议：

```powershell
docker compose exec postgres pg_dump -U image_show image_show > backup.sql
```

首次启动后，确认默认积分套餐已写入 `packages` 表。

## 6. R2 生命周期配置

代码侧规则：

- 免费或匿名图片写入 `generations/free/...`
- 付费用户图片写入 `generations/paid/...`
- 用户支付或管理员充值后，旧的 `free` 图片会迁移到 `paid`

Cloudflare R2 建议规则：

- `generations/free/`：7 天后自动删除
- `generations/paid/`：不设置自动清理

详细说明见 [r2-lifecycle.md](r2-lifecycle.md)。

## 7. 反向代理

Nginx 示例：

```nginx
server {
    listen 80;
    server_name yourdomain.com;

    location / {
        proxy_pass http://127.0.0.1:3000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    location /api/generations/ {
        proxy_pass http://127.0.0.1:3000;
        proxy_http_version 1.1;
        proxy_set_header Connection "";
        proxy_buffering off;
        proxy_cache off;
        proxy_read_timeout 300s;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

如果使用 HTTPS，`SERVER_ADDRESS` 应填写 `https://yourdomain.com`。

## 8. 上线检查

容器启动：

```powershell
docker compose ps
docker compose logs -f image-show
```

健康检查：

```powershell
curl http://localhost:3000/health
curl http://localhost:3000/api/health
```

浏览器验收：

- 打开 `https://yourdomain.com`
- 注册 / 登录正常
- 后台 `/admin` 可访问
- 套餐页 `/packages` 可创建订单
- 支付回调 `/api/payment/notify` 可被支付网关访问
- 生成图片 SSE 进度正常
- 历史图片可查看
- 监控页数据正常

## 9. 常见问题

### 前端静态文件 404

确认镜像构建阶段执行了：

```dockerfile
COPY --from=frontend /app/web/dist ./web/dist
```

并且 `web/dist` 在 Go 构建前存在，否则 `go:embed` 无法嵌入前端产物。

### 支付成功但积分不到账

检查：

- `SERVER_ADDRESS` 是否为公网地址
- 支付网关是否能访问 `/api/payment/notify`
- `EPAY_ID`、`EPAY_KEY` 是否与网关一致
- 订单 `pay_method` 是否与回调 `type` 一致，例如 `alipay` 或 `wxpay`

### 图片生成失败

检查：

- `SUB2API_BASE_URL` 是否从容器内可达
- 后台渠道是否启用
- 渠道 API Key 是否正确
- `MOCK_SUB2API` 是否误设为 `true`

### Turnstile 不显示

检查：

- 后台 `captcha_enabled=true`
- `turnstile_site_key` 和 `turnstile_secret` 是否都已配置
- 页面是否能加载 `https://challenges.cloudflare.com/turnstile/v0/api.js`

### Redis 未配置

应用会降级到本地内存限流，适合本地开发；生产多实例部署必须使用 Redis，否则实例之间限流状态不会共享。

## 10. 发布流程建议

1. 本地运行测试：

   ```powershell
   $env:CGO_ENABLED='0'; go test ./...
   pnpm.cmd --dir web build
   ```

2. 构建镜像：

   ```powershell
   docker build -t image-show:YYYYMMDD .
   ```

3. 更新 compose 镜像标签。

4. 启动新版本：

   ```powershell
   docker compose --env-file .env.production up -d
   ```

5. 验证 `/health`、登录、生成、支付回调和管理台监控。

6. 保留上一版镜像标签，必要时快速回滚。
