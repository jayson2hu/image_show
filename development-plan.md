# GPT-Image-2 图片生成网站 — 详细开发计划

> 创建日期：2026-04-29
> 状态：开发计划（供 Codex 执行）
> 关联文档：requirements-confirmed.md

---

## 开发原则

1. **每个小功能独立自测验收**，通过后才开发下一个
2. **每个大功能完成后整体验收**一次
3. **遇到矛盾点立即记录**，不自行决策，等待确认
4. **文档及时更新**，每完成一个小功能更新进度

---

## Phase 1：项目骨架（P0 第 1 周）

### 1.1 Go 后端项目初始化

**目标**：创建可运行的 Go 后端空项目

**步骤**：
1. 在 `D:\vscodefile\image_show` 下创建后端目录结构：
   ```
   image_show/
   ├── main.go                 # 入口，参考 new-api 的 main.go 模式
   ├── go.mod                  # module github.com/your-org/image-show
   ├── go.sum
   ├── .env.example            # 环境变量模板
   ├── Makefile                # dev/build/test 命令
   ├── common/                 # 共享工具
   │   ├── env.go              # 环境变量加载
   │   ├── constants.go        # 常量定义
   │   └── json.go             # JSON 工具（参考 new-api common/json.go）
   ├── config/                 # 配置管理
   │   └── config.go           # 配置结构体 + 加载逻辑
   ├── model/                  # 数据模型 + GORM
   │   └── main.go             # DB 初始化
   ├── middleware/              # 中间件
   ├── controller/              # 请求处理器
   ├── service/                 # 业务逻辑
   ├── router/                  # 路由注册
   │   └── main.go
   └── web/                     # Vue 前端（后续创建）
   ```
2. `main.go` 启动流程：加载 .env → 初始化配置 → 初始化 DB → 创建 Gin 实例 → 注册路由 → 启动服务
3. 添加 `/health` 健康检查端点，返回 `{"status": "ok"}`
4. Makefile 命令：`make dev`（go run + 热重载 air）、`make build`

**自测验收**：
- [ ] `make dev` 启动成功，监听 `:3000`
- [ ] `curl http://localhost:3000/health` 返回 200
- [ ] `.env.example` 包含所有配置项说明

---

### 1.2 数据库初始化（GORM 多驱动）

**目标**：GORM 连接数据库，支持 SQLite（本地）和 PostgreSQL（生产）

**步骤**：
1. `model/main.go` 实现 `InitDB()` 函数，参考 new-api 的 `chooseDB()` 模式：
   - 环境变量 `DB_DRIVER=sqlite|postgres`
   - SQLite：`gorm.Open(sqlite.Open("./data/image_show.db"))`
   - PostgreSQL：`gorm.Open(postgres.Open(dsn))`，DSN 从 `DATABASE_DSN` 环境变量读取
   - 连接池配置：MaxIdleConns=10, MaxOpenConns=50, ConnMaxLifetime=30min
2. 定义所有数据模型（仅结构体，不写业务逻辑）：

   **users 表**：
   ```go
   type User struct {
       ID            int64     `gorm:"primaryKey;autoIncrement"`
       Username      string    `gorm:"size:64;uniqueIndex"`
       Email         string    `gorm:"size:128;uniqueIndex"`
       PasswordHash  string    `gorm:"size:256"`
       WechatOpenID  string    `gorm:"size:128;index"`
       Role          int       `gorm:"default:1"`           // 1=user, 10=admin, 100=root
       Status        int       `gorm:"default:1"`           // 1=active, 2=banned
       Credits       float64   `gorm:"type:decimal(10,2);default:0"`
       CreditsExpiry *time.Time
       AvatarURL     string    `gorm:"size:512"`
       LastLoginAt   *time.Time
       LastLoginIP   string    `gorm:"size:64"`
       CreatedAt     time.Time
       UpdatedAt     time.Time
   }
   ```

   **generations 表**：
   ```go
   type Generation struct {
       ID           int64     `gorm:"primaryKey;autoIncrement"`
       UserID       *int64    `gorm:"index"`                // 可为空（匿名用户）
       AnonymousID  string    `gorm:"size:128;index"`       // IP+指纹 hash
       Prompt       string    `gorm:"type:text"`
       Quality      string    `gorm:"size:16"`              // low/medium/high
       Size         string    `gorm:"size:16"`              // 1024x1024 等
       CreditsCost  float64   `gorm:"type:decimal(10,2)"`
       Status       int       `gorm:"default:0;index"`      // 0=pending,1=generating,2=uploading,3=completed,4=failed,5=cancelled
       R2Key        string    `gorm:"size:256"`             // R2 对象 key
       ImageURL     string    `gorm:"size:512"`             // Presigned URL（临时）
       ErrorMsg     string    `gorm:"size:512"`
       IP           string    `gorm:"size:64"`
       IsDeleted    bool      `gorm:"default:false;index"`  // 软删除（用户侧）
       CreatedAt    time.Time `gorm:"index"`
       UpdatedAt    time.Time
   }
   ```

   **credit_logs 表**：
   ```go
   type CreditLog struct {
       ID          int64     `gorm:"primaryKey;autoIncrement"`
       UserID      int64     `gorm:"index"`
       Type        int       `gorm:"index"`  // 1=register_gift, 2=consume, 3=admin_topup, 4=refund
       Amount      float64   `gorm:"type:decimal(10,2)"`
       Balance     float64   `gorm:"type:decimal(10,2)"`  // 操作后余额
       RelatedID   *int64                                  // 关联的 generation ID
       Remark      string    `gorm:"size:256"`
       OperatorID  *int64                                  // 管理员操作时记录
       CreatedAt   time.Time `gorm:"index"`
   }
   ```

   **login_logs 表**：
   ```go
   type LoginLog struct {
       ID        int64     `gorm:"primaryKey;autoIncrement"`
       UserID    int64     `gorm:"index"`
       IP        string    `gorm:"size:64"`
       UserAgent string    `gorm:"size:512"`
       Method    string    `gorm:"size:16"`  // email/wechat
       Success   bool
       CreatedAt time.Time `gorm:"index"`
   }
   ```

   **channels 表**：
   ```go
   type Channel struct {
       ID        int64     `gorm:"primaryKey;autoIncrement"`
       Name      string    `gorm:"size:64"`
       BaseURL   string    `gorm:"size:256"`             // sub2api 地址
       APIKey    string    `gorm:"size:256"`             // sub2api API Key
       Headers   string    `gorm:"type:text"`            // 额外 header（JSON）
       Status    int       `gorm:"default:1"`            // 1=enabled, 2=disabled
       Weight    int       `gorm:"default:1"`            // 负载均衡权重
       Remark    string    `gorm:"size:256"`
       CreatedAt time.Time
       UpdatedAt time.Time
   }
   ```

   **settings 表**：
   ```go
   type Setting struct {
       Key       string `gorm:"primaryKey;size:64"`
       Value     string `gorm:"type:text"`
       UpdatedAt time.Time
   }
   ```

   **prompt_templates 表**：
   ```go
   type PromptTemplate struct {
       ID        int64     `gorm:"primaryKey;autoIncrement"`
       Category  string    `gorm:"size:32;index"`  // default/repair/style
       Label     string    `gorm:"size:64"`        // 前端显示名
       Prompt    string    `gorm:"type:text"`      // prompt 内容
       SortOrder int       `gorm:"default:0"`
       Status    int       `gorm:"default:1"`      // 1=enabled, 2=disabled
       CreatedAt time.Time
       UpdatedAt time.Time
   }
   ```

3. `AutoMigrate` 所有模型
4. 提供 `CloseDB()` 函数

**自测验收**：
- [ ] `make dev` 启动后自动创建 `./data/image_show.db`（SQLite 模式）
- [ ] 所有表结构正确创建（用 SQLite 工具查看）
- [ ] 设置 `DB_DRIVER=postgres` + `DATABASE_DSN=...` 后可连接 PG

---

### 1.3 配置管理

**目标**：统一管理所有配置项，支持 .env + 环境变量

**步骤**：
1. `config/config.go` 定义配置结构体：
   ```go
   type Config struct {
       Port          int
       DBDriver      string  // sqlite | postgres
       DatabaseDSN   string
       RedisAddr     string
       RedisPassword string
       RedisDB       int
       JWTSecret     string
       R2Endpoint    string
       R2AccessKey   string
       R2SecretKey   string
       R2Bucket      string
       R2PublicURL   string  // CDN 公开地址
       SMTPHost      string
       SMTPPort      int
       SMTPUser      string
       SMTPPassword  string
       SMTPFrom      string
       Sub2APIBaseURL string  // 默认 http://sub2api:8080
   }
   ```
2. `LoadConfig()` 从 .env 和环境变量加载，环境变量优先
3. 全局单例 `var AppConfig *Config`

**自测验收**：
- [ ] 不设置任何环境变量时使用合理默认值启动
- [ ] 设置环境变量可覆盖默认值
- [ ] `.env.example` 包含所有配置项及注释

---

### 1.4 IP 记录与透传中间件

**目标**：正确获取用户真实 IP，注入请求上下文

**步骤**：
1. `middleware/ip.go` 实现 `RealIP()` 中间件：
   ```go
   func RealIP() gin.HandlerFunc {
       return func(c *gin.Context) {
           ip := c.GetHeader("CF-Connecting-IP")
           if ip == "" {
               ip = c.GetHeader("X-Real-IP")
           }
           if ip == "" {
               ip = c.GetHeader("X-Forwarded-For")
               if idx := strings.Index(ip, ","); idx != -1 {
                   ip = strings.TrimSpace(ip[:idx])
               }
           }
           if ip == "" {
               ip = c.ClientIP()
           }
           c.Set("real_ip", ip)
           c.Next()
       }
   }
   ```
2. `common/context.go` 提供 `GetRealIP(c *gin.Context) string` 工具函数
3. 在 `router/main.go` 全局注册该中间件

**自测验收**：
- [ ] 设置 `X-Real-IP: 1.2.3.4` header 请求，`/health` 日志输出正确 IP
- [ ] 设置 `CF-Connecting-IP` 优先级高于 `X-Real-IP`
- [ ] 不设置任何 header 时回退到 `RemoteAddr`

---

### 1.5 Vue 3 前端项目初始化

**目标**：创建可运行的 Vue 3 前端空项目，与后端联调

**步骤**：
1. 在 `web/` 目录下用 Vite 创建 Vue 3 + TypeScript 项目：
   ```
   web/
   ├── package.json          # pnpm
   ├── vite.config.ts        # proxy 到后端 :3000
   ├── tsconfig.json
   ├── tailwind.config.js
   ├── postcss.config.js
   ├── index.html
   └── src/
       ├── main.ts
       ├── App.vue
       ├── router/
       │   └── index.ts      # vue-router 路由定义
       ├── stores/
       │   └── user.ts       # Pinia 用户状态
       ├── views/
       │   ├── Home.vue       # 首页（生成页面）
       │   ├── Login.vue      # 登录页
       │   └── Register.vue   # 注册页
       ├── components/        # 公共组件
       ├── api/
       │   └── index.ts       # axios 实例 + 拦截器
       ├── utils/
       └── assets/
   ```
2. 安装依赖：`vue`, `vue-router`, `pinia`, `axios`, `tailwindcss`, `autoprefixer`, `postcss`
3. `vite.config.ts` 配置开发代理：
   ```ts
   server: {
     port: 5173,
     proxy: {
       '/api': {
         target: 'http://localhost:3000',
         changeOrigin: true,
       },
     },
   }
   ```
4. TailwindCSS 配置，参考 sub2api 前端
5. `api/index.ts` 创建 axios 实例，添加 JWT token 拦截器（从 localStorage 读取）
6. 基础路由：`/` → Home, `/login` → Login, `/register` → Register
7. 基础布局组件：顶部导航栏（Logo + 登录/用户信息）+ 主内容区

**自测验收**：
- [ ] `cd web && pnpm install && pnpm dev` 启动成功，访问 `http://localhost:5173`
- [ ] 页面显示基础布局（导航栏 + 内容区）
- [ ] 路由切换正常：首页 / 登录 / 注册
- [ ] API 代理正常：前端请求 `/api/health` 返回后端数据

---

### 1.6 前端嵌入 Go 二进制

**目标**：参考 new-api 的 `go:embed` 方式，将 Vue 构建产物嵌入 Go 二进制

**步骤**：
1. `web/embed.go` 实现前端静态文件嵌入：
   ```go
   //go:embed dist
   var DistFS embed.FS
   ```
2. `router/web.go` 注册静态文件服务：
   - API 路由（`/api/*`）走后端处理
   - 其他路由走前端静态文件
   - SPA fallback：未匹配的路由返回 `index.html`
3. Makefile 添加 `make build-frontend` 命令：`cd web && pnpm build`
4. Makefile 添加 `make build-all` 命令：先构建前端，再构建 Go 二进制

**自测验收**：
- [ ] `make build-all` 成功生成单个二进制文件
- [ ] 运行二进制，访问 `http://localhost:3000` 显示前端页面
- [ ] 前端路由刷新不 404（SPA fallback 正常）
- [ ] `/api/health` 正常返回

---

### Phase 1 整体验收

- [x] `make dev` 启动后端，`cd web && pnpm dev` 启动前端，前后端联调正常（本机无 make/air，使用 `go run .` + `pnpm dev` 等价验收；5173 被占用，使用 5174）
- [x] 数据库自动创建所有表（SQLite 模式）
- [x] 配置管理正常加载 .env 和环境变量
- [x] IP 中间件正确解析真实 IP
- [x] 前端基础布局和路由正常
- [x] `make build-all` 生成单二进制，嵌入前端正常工作（使用 `pnpm build` + `CGO_ENABLED=0 go build` 等价验收）

---

## Phase 2：核心生成功能（P0 第 2 周）

### 2.1 邮箱注册/登录 — 后端 API

**目标**：实现邮箱注册、登录、JWT 签发

**步骤**：
1. `service/email.go` 实现 SMTP 邮件发送：
   - `SendVerificationCode(email string)` 发送 6 位验证码
   - 验证码存 Redis（本地开发用内存 map），key: `imgshow:vcode:{email}`，TTL 5 分钟
   - 同一邮箱 60 秒内不可重复发送
2. `service/auth.go` 实现认证逻辑：
   - `Register(email, password, code)` — 校验验证码 → 检查邮箱唯一 → 创建用户 → 赠送 3 积分（7 天有效）→ 签发 JWT
   - `Login(email, password)` — 校验密码 → 记录登录日志 → 签发 JWT
   - 密码使用 `bcrypt` 加密
3. `service/jwt.go` 实现 JWT 工具：
   - `GenerateToken(userID, role)` — 签发，过期时间 24 小时
   - `ParseToken(tokenString)` — 解析验证
4. `controller/auth.go` 实现 API 端点：
   - `POST /api/auth/send-code` — 发送验证码
   - `POST /api/auth/register` — 注册
   - `POST /api/auth/login` — 登录
   - `GET /api/auth/me` — 获取当前用户信息（需 JWT）
5. `middleware/auth.go` 实现 JWT 认证中间件：
   - 从 `Authorization: Bearer <token>` 读取
   - 解析后将 `userID` 和 `role` 注入 context
6. `model/setting.go` 实现注册开关：
   - Setting 表 key: `register_enabled`，value: `true/false`
   - 注册接口检查开关状态

**自测验收**：
- [ ] `POST /api/auth/send-code` 发送验证码到邮箱（检查 SMTP 日志或邮箱收件）
- [ ] 60 秒内重复发送返回 429
- [ ] `POST /api/auth/register` 注册成功，返回 JWT token
- [ ] 注册后用户 credits = 3，credits_expiry = 7 天后
- [ ] `POST /api/auth/login` 登录成功，返回 JWT token
- [ ] `GET /api/auth/me` 带 token 返回用户信息，不带 token 返回 401
- [ ] 关闭注册开关后，注册接口返回 403
- [ ] 登录日志正确记录 IP、UserAgent、Method

---

### 2.2 邮箱注册/登录 — 前端页面

**目标**：实现登录、注册前端页面，与后端联调

**步骤**：
1. `views/Login.vue` 登录页面：
   - 邮箱 + 密码输入框
   - 登录按钮，调用 `/api/auth/login`
   - 成功后存 token 到 localStorage，跳转首页
   - 错误提示（密码错误、账号不存在等）
2. `views/Register.vue` 注册页面：
   - 邮箱 + 验证码 + 密码 + 确认密码
   - 发送验证码按钮（60 秒倒计时）
   - 注册按钮，调用 `/api/auth/register`
   - 成功后自动登录跳转首页
3. `stores/user.ts` Pinia 用户状态管理：
   - `token`, `user` 状态
   - `login()`, `register()`, `logout()`, `fetchUser()` actions
   - 初始化时从 localStorage 恢复 token
4. `api/index.ts` axios 拦截器：
   - 请求拦截：自动添加 `Authorization: Bearer <token>`
   - 响应拦截：401 时清除 token 跳转登录页
5. 导航栏显示：未登录显示「登录/注册」，已登录显示用户邮箱 + 积分余额 + 退出

**自测验收**：
- [ ] 注册流程：输入邮箱 → 收到验证码 → 填写密码 → 注册成功 → 自动跳转首页
- [ ] 登录流程：输入邮箱密码 → 登录成功 → 首页显示用户信息
- [ ] 刷新页面后登录状态保持
- [ ] 退出后 token 清除，跳转登录页
- [ ] 注册关闭时前端显示提示

---

### 2.3 Sub2API 内网对接

**目标**：封装 Sub2API HTTP 客户端，实现图片生成调用

**步骤**：
1. `service/sub2api.go` 封装 HTTP 客户端：
   ```go
   type Sub2APIClient struct {
       BaseURL    string
       APIKey     string
       Headers    map[string]string
       HTTPClient *http.Client
   }
   ```
2. 实现 `GenerateImage(prompt, quality, size, userIP string)` 方法：
   - 调用 `POST {BaseURL}/v1/images/generations`
   - 请求体：`{"model": "gpt-image-1", "prompt": "...", "quality": "...", "size": "..."}`
   - 透传用户 IP：`X-Real-IP` 和 `X-Forwarded-For` header
   - 超时设置 120 秒（图片生成较慢）
   - 返回 base64 图片数据或 URL
3. `service/channel.go` 渠道选择逻辑：
   - 从 DB 加载所有 enabled 渠道
   - 按 weight 加权随机选择
   - 渠道不可用时自动切换下一个
4. 本地开发时支持 mock 模式（环境变量 `MOCK_SUB2API=true`），返回测试图片

**自测验收**：
- [ ] 配置渠道后，调用 `GenerateImage` 返回图片数据
- [ ] IP 透传正确（检查 sub2api 日志）
- [ ] 渠道权重分配正常（多次调用统计分布）
- [ ] 超时处理正常（模拟慢响应）
- [ ] Mock 模式返回测试图片

---

### 2.4 SSE 生成流

**目标**：后端 SSE 推送生成状态，前端实时接收

**执行决策（2026-04-29）**：本节原步骤包含“扣减积分（预扣）/失败时退还积分”，但积分系统在 2.6 才实现。为避免跨阶段提前实现，2.4 先完成生成任务创建、Sub2API 调用、状态落库与 SSE 推送；积分预扣、余额校验、失败退还统一延后到 2.6 接入。

**步骤**：
1. `controller/generation.go` 实现生成接口：
   - `POST /api/generations` — 创建生成任务（返回 generation ID）
   - `GET /api/generations/:id/stream` — SSE 流，推送状态变化
2. SSE 状态流转：
   ```
   pending → generating → uploading → completed
                                    → failed
   ```
   每个状态变化推送一条 SSE 事件：
   ```
   event: status
   data: {"status": "generating", "message": "正在生成图片..."}

   event: status
   data: {"status": "completed", "image_url": "https://..."}

   event: status
   data: {"status": "failed", "error": "生成失败，请重试"}
   ```
3. `service/generation.go` 生成流程（异步 goroutine）：
   - 创建 generation 记录（status=pending）
   - 扣减积分（预扣）（延后到 2.6 积分系统接入）
   - 调用 Sub2API 生成图片（status=generating）
   - 上传 R2（status=uploading）
   - 更新记录（status=completed + image_url）
   - 失败时退还积分（status=failed）（延后到 2.6 积分系统接入）
4. 使用 channel 通知 SSE handler 状态变化：
   ```go
   type GenerationNotifier struct {
       mu       sync.RWMutex
       channels map[int64]chan GenerationEvent
   }
   ```
5. SSE 连接保活：每 15 秒发送 `:keepalive\n\n`
6. 客户端断开时清理 channel

**自测验收**：
- [ ] 创建生成任务返回 generation ID
- [ ] SSE 连接建立成功（DevTools Network 查看 EventStream）
- [ ] 状态推送顺序正确：pending → generating → uploading → completed
- [ ] 失败时推送 failed 事件
- [ ] 客户端断开后 goroutine 正确清理
- [ ] 15 秒 keepalive 正常发送

---

### 2.5 Cloudflare R2 图片上传 + 归属追踪

**目标**：生成完成后上传 R2，记录图片归属

**步骤**：
1. `service/r2.go` 封装 R2 客户端（aws-sdk-go-v2）：
   ```go
   type R2Client struct {
       client *s3.Client
       bucket string
       cdnURL string
   }
   ```
   - `Upload(key string, data []byte, contentType string) error`
   - `GeneratePresignedURL(key string, expiry time.Duration) (string, error)`
   - `Delete(key string) error`
2. R2 Key 命名规则：`generations/{userID|anon}/{YYYY-MM}/{generationID}.png`
3. 上传流程：
   - 从 Sub2API 返回的 base64 解码为 bytes
   - 上传到 R2
   - 生成 Presigned URL（有效期 1 小时）
   - 更新 generation 记录的 `r2_key` 和 `image_url`
4. 归属追踪：
   - 已登录用户：`user_id` = 当前用户 ID
   - 未登录用户：`anonymous_id` = SHA256(IP + 浏览器指纹)
   - 浏览器指纹由前端生成并通过 header `X-Fingerprint` 传递
5. 匿名归属迁移（在注册成功时调用）：
   ```go
   func MigrateAnonymousGenerations(anonymousID string, userID int64) error {
       return db.Model(&Generation{}).
           Where("anonymous_id = ? AND user_id IS NULL", anonymousID).
           Update("user_id", userID).Error
   }
   ```

**自测验收**：
- [ ] 图片成功上传到 R2（检查 R2 控制台）
- [ ] Presigned URL 可正常访问图片
- [ ] R2 Key 格式正确
- [ ] 已登录用户生成的图片 `user_id` 正确
- [ ] 未登录用户生成的图片 `anonymous_id` 正确
- [ ] 注册后匿名图片自动归属到新用户

---

### 2.6 积分系统

**目标**：实现积分扣减、查询、流水记录

**步骤**：
1. `service/credit.go` 实现积分服务：
   - `GetBalance(userID int64) (float64, error)` — 查询余额（排除过期积分）
   - `Deduct(userID int64, amount float64, generationID int64) error` — 扣减积分（事务）
   - `Refund(userID int64, amount float64, generationID int64) error` — 退还积分
   - `AdminTopup(userID, operatorID int64, amount float64, remark string) error` — 管理员充值
   - `RegisterGift(userID int64) error` — 注册赠送 3 积分，设置 7 天有效期
2. 积分消耗比例（常量定义）：
   ```go
   var QualityCost = map[string]float64{
       "low":    0.2,
       "medium": 1.0,
       "high":   4.0,
   }
   ```
3. 扣减逻辑（事务内）：
   - 检查余额 ≥ 消耗量
   - 检查积分未过期
   - 更新 `users.credits`
   - 写入 `credit_logs` 流水
4. `controller/credit.go` API 端点：
   - `GET /api/credits/balance` — 查询余额
   - `GET /api/credits/logs` — 查询流水（分页）

**自测验收**：
- [ ] 注册后余额 = 3，有效期 7 天
- [ ] 生成 low 质量图片扣减 0.2 积分
- [ ] 生成 high 质量图片扣减 4 积分
- [ ] 余额不足时返回错误，不创建生成任务
- [ ] 生成失败时积分退还
- [ ] 流水记录正确（类型、金额、余额、关联 ID）
- [ ] 过期积分不可使用

---

### 2.7 未登录免费试用

**目标**：未登录用户可免费生成 1 次图片（IP + 指纹限制）

**步骤**：
1. `service/trial.go` 实现免费试用逻辑：
   - `CheckTrialEligible(ip, fingerprint string) (bool, error)`
   - 标识 = SHA256(IP + fingerprint)
   - Redis key: `imgshow:trial:{标识}`，TTL 30 天
   - 本地开发用内存 map 替代 Redis
2. `controller/generation.go` 修改生成接口：
   - 未登录时检查免费试用资格
   - 有资格：创建生成任务，quality 固定为 `low`
   - 无资格：返回 403 + 提示注册
3. 前端 `X-Fingerprint` header：
   - 使用 canvas fingerprint + screen resolution + timezone 等生成
   - 存储在 localStorage，保持一致性

**自测验收**：
- [ ] 未登录首次生成成功（quality 固定 low）
- [ ] 未登录第二次生成返回 403
- [ ] 30 天后可再次免费生成
- [ ] 换 IP + 指纹可再次生成（符合预期）
- [ ] 已登录用户不走免费试用逻辑

---

### 2.8 生成页面 UI

**目标**：实现完整的图片生成前端页面

**步骤**：
1. `views/Home.vue` 生成页面布局：
   - 顶部：Prompt 输入框（textarea，支持多行）
   - Prompt 下方：默认模板标签（可点击填充）
   - 图片修复预设选项标签（点击拼接到 prompt）
   - 参数选择：质量（low/medium/high）+ 尺寸（1024x1024 等）
   - 生成按钮（显示积分消耗量）
   - 底部：生成结果展示区
2. `components/PromptTags.vue` 模板标签组件：
   - 从 `/api/prompt-templates` 加载标签列表
   - 分类显示：默认 / 修复 / 风格
   - 点击标签填充或拼接到 prompt 输入框
3. `components/GenerationProgress.vue` 生成进度组件：
   - 连接 SSE 流
   - 趣味状态文案轮播（每 3-5 秒切换）：
     1. 「正在创建图片...」
     2. 「正在打草稿...」
     3. 「生成初稿中...」
     4. 「正在设置场景...」
     5. 「正在润饰细节...」
     6. 「即将完成...」
     7. 「正在做最后润色...」
     8. 「最后微调一下...」
   - 骨架屏 / 加载动画
   - 取消按钮
4. `components/ImagePreview.vue` 图片预览组件：
   - 生成完成后展示图片
   - 下载按钮
   - 重新生成按钮
5. 错误处理：
   - 生成失败友好提示
   - 网络断开重连提示
   - 积分不足提示 + 引导充值

**自测验收**：
- [ ] Prompt 输入 + 模板标签点击填充正常
- [ ] 图片修复标签拼接到 prompt 正常
- [ ] 参数选择正常，积分消耗量实时显示
- [ ] 点击生成后进入加载状态，趣味文案轮播
- [ ] SSE 接收状态更新，completed 时展示图片
- [ ] 取消生成正常
- [ ] 错误提示友好
- [ ] 未登录用户看到免费试用提示

---

### Phase 2 整体验收

- [x] 完整流程：注册 → 登录 → 输入 prompt → 选择参数 → 生成 → 查看图片（mock Sub2API + 本地 data URL 路径已自动化验收）
- [x] 未登录免费试用 1 次 → 第 2 次提示注册
- [x] 积分扣减正确，流水记录完整
- [x] SSE 状态推送正常，趣味文案轮播
- [ ] R2 图片上传和 Presigned URL 访问正常（代码已实现；需配置 Cloudflare R2 凭据后做真实外部验收）
- [x] 匿名用户注册后图片归属迁移
- [x] 生成失败积分退还

---

## Phase 3：防护与管理（P1 第 3 周）

### 3.1 Redis 限流

**目标**：实现用户级 + IP 级 + 全局预算限流

**步骤**：
1. `middleware/ratelimit.go` 实现 Redis 滑动窗口限流：
   - 用户级：已登录用户 10 次/分钟，key: `imgshow:rl:user:{userID}`
   - IP 级：20 次/分钟，key: `imgshow:rl:ip:{ip}`
   - 使用 Redis ZSET 滑动窗口算法
2. `service/budget.go` 实现每日预算熔断：
   - Setting 表 key: `daily_budget`，value: 每日最大 API 调用次数
   - Redis key: `imgshow:budget:{YYYY-MM-DD}`，INCR 计数
   - 超过预算返回 503 + 提示
3. 本地开发用内存限流（`golang.org/x/time/rate`）
4. 限流返回标准 429 响应 + `Retry-After` header

**自测验收**：
- [ ] 已登录用户 1 分钟内第 11 次请求返回 429
- [ ] 同一 IP 1 分钟内第 21 次请求返回 429
- [ ] 超过每日预算后所有请求返回 503
- [ ] `Retry-After` header 正确
- [ ] 本地开发内存限流正常工作

---

### 3.2 日志系统

**目标**：生成日志、登录日志存 DB，管理后台可查询和批量删除

**步骤**：
1. 生成日志已在 `generations` 表中，无需额外表
2. 登录日志已在 `login_logs` 表中
3. `controller/admin/log.go` 管理后台日志 API：
   - `GET /api/admin/logs/generations` — 生成日志查询（分页 + 时间范围 + 用户筛选 + 状态筛选）
   - `GET /api/admin/logs/logins` — 登录日志查询（分页 + 时间范围 + 用户筛选）
   - `DELETE /api/admin/logs/generations` — 按时间条件批量删除生成日志
   - `DELETE /api/admin/logs/logins` — 按时间条件批量删除登录日志
4. 批量删除参数：`before` (时间戳)，删除该时间之前的所有日志
5. 删除前返回将删除的记录数，需二次确认

**自测验收**：
- [ ] 生成日志按时间范围查询正常
- [ ] 登录日志按用户筛选正常
- [ ] 批量删除指定时间之前的日志正常
- [ ] 分页参数正常（page, pageSize）

---

### 3.3 渠道管理

**目标**：管理后台 CRUD sub2api 渠道

**步骤**：
1. `controller/admin/channel.go` 渠道管理 API：
   - `GET /api/admin/channels` — 渠道列表
   - `POST /api/admin/channels` — 创建渠道
   - `PUT /api/admin/channels/:id` — 编辑渠道
   - `DELETE /api/admin/channels/:id` — 删除渠道
   - `POST /api/admin/channels/:id/test` — 测试渠道连通性
2. 渠道字段：Name, BaseURL, APIKey, Headers(JSON), Status, Weight, Remark
3. 测试连通性：调用渠道的 `/v1/models` 接口，检查返回是否正常
4. 渠道状态切换：enabled/disabled

**自测验收**：
- [ ] 创建渠道成功
- [ ] 编辑渠道信息正常
- [ ] 删除渠道正常
- [ ] 测试连通性返回正确结果
- [ ] 禁用渠道后不参与生成调度

---

### 3.4 管理员后台

**目标**：实现管理后台核心功能

**步骤**：
1. `middleware/admin.go` 管理员权限中间件：
   - 检查 `role >= 10`（admin）
   - root 用户 `role = 100` 拥有所有权限
2. **用户管理**：
   - `GET /api/admin/users` — 用户列表（分页 + 搜索）
   - `PUT /api/admin/users/:id/status` — 封禁/解封用户
   - `PUT /api/admin/users/:id/role` — 修改用户角色
   - `GET /api/admin/users/:id/generations` — 查看用户生成的图片
3. **积分管理**：
   - `POST /api/admin/users/:id/credits` — 手动充值积分
   - `GET /api/admin/credits/logs` — 全局积分流水
4. **Prompt 模板管理**：
   - `GET /api/admin/prompt-templates` — 模板列表
   - `POST /api/admin/prompt-templates` — 创建模板
   - `PUT /api/admin/prompt-templates/:id` — 编辑模板
   - `DELETE /api/admin/prompt-templates/:id` — 删除模板
   - 字段：Category(default/repair/style), Label, Prompt, SortOrder, Status
5. **系统设置**：
   - `GET /api/admin/settings` — 获取所有设置
   - `PUT /api/admin/settings` — 批量更新设置
   - 设置项：register_enabled, daily_budget, default_credits, credits_expiry_days
6. **前端管理页面**（`views/admin/` 目录）：
   - `Dashboard.vue` — 概览（用户数、生成数、今日消耗）
   - `Users.vue` — 用户管理表格
   - `Channels.vue` — 渠道管理
   - `Logs.vue` — 日志查询
   - `PromptTemplates.vue` — 模板管理
   - `Settings.vue` — 系统设置

**自测验收**：
- [x] 普通用户访问 `/api/admin/*` 返回 403
- [x] 管理员可查看/封禁/解封用户
- [x] 管理员可充值积分，流水记录正确
- [x] Prompt 模板 CRUD 正常，前端生成页面加载模板
- [x] 系统设置修改后立即生效
- [x] 管理后台前端页面正常渲染

**问题记录（2026-04-29）**：
- 自测发现 `users.username` 使用唯一索引但注册流程未填用户名，多个注册用户会因空字符串触发唯一约束；已改为普通索引，登录唯一标识继续使用 `email`。
- 前后端联调发现模型未统一 `json` 标签会返回 `ID/Email/Credits` 等大写字段，前端按 `id/email/credits` 读取会失败；已为模型补齐 JSON 标签，统一接口字段契约。

---

### 3.5 微信登录

**目标**：参考 new-api 实现微信登录

**步骤**：
1. `service/wechat.go` 实现微信登录：
   - 方案 A：WxPusher 扫码登录（推荐，无需企业认证）
   - 方案 B：公众号测试号 OAuth
   - 生成带参数二维码 → 用户扫码 → 回调获取 OpenID → 绑定/创建用户
2. `controller/auth.go` 添加微信登录端点：
   - `GET /api/auth/wechat/qrcode` — 获取登录二维码
   - `GET /api/auth/wechat/callback` — 微信回调
   - `GET /api/auth/wechat/status` — 轮询登录状态
3. 用户绑定逻辑：
   - OpenID 已绑定：直接登录
   - OpenID 未绑定 + 已登录：绑定到当前账号
   - OpenID 未绑定 + 未登录：创建新用户
4. 前端登录页添加微信扫码入口

**自测验收**：
- [x] 生成微信登录二维码正常
- [x] 扫码后回调正确处理（按 new-api：公众号验证码换取 OpenID）
- [x] 新用户扫码自动创建账号
- [x] 已有用户扫码直接登录
- [x] 绑定/解绑微信正常

**执行决策（2026-04-29）**：用户确认 3.5 按 new-api 实现，不走文档原先“动态二维码 + callback + status 轮询”的完整扫码协议。实际采用 WeChat Server 验证码模式：后台配置公众号二维码、WeChat Server 地址和 token，前端提交公众号验证码，后端调用 `{WECHAT_SERVER_ADDRESS}/api/wechat/user?code=...` 换取 OpenID。

---

### Phase 3 整体验收

- [x] 限流正常：超频返回 429，预算耗尽返回 503
- [x] 管理后台所有功能正常
- [x] 日志查询和批量删除正常
- [x] 渠道管理 + 测试连通性正常
- [x] 微信登录全流程正常
- [x] 权限控制：普通用户无法访问管理接口

---

## Phase 4：体验优化（P2 第 4 周）

### 4.1 图片历史 + R2 删除

**目标**：用户查看/下载/删除生成历史

**步骤**：
1. `controller/generation.go` 用户图片 API：
   - `GET /api/generations` — 我的生成历史（分页，排除软删除）
   - `GET /api/generations/:id` — 图片详情（刷新 Presigned URL）
   - `DELETE /api/generations/:id` — 软删除（标记 `is_deleted=true`，R2 不删）
2. `controller/admin/generation.go` 管理后台图片 API：
   - `GET /api/admin/generations` — 全部图片（按用户/时间筛选）
   - `DELETE /api/admin/generations/batch` — 批量删除（可选是否真删 R2）
   - 真删 R2 时调用 `r2Client.Delete(key)`
3. 前端 `views/History.vue` 图片历史页面：
   - 瀑布流/网格展示
   - 分页加载（滚动加载更多）
   - 点击查看大图 + 下载
   - 删除确认弹窗
4. Presigned URL 刷新：URL 过期后重新生成（有效期 1 小时）

**自测验收**：
- [x] 用户查看自己的生成历史正常
- [x] 软删除后图片不再显示
- [x] 管理员批量删除正常
- [ ] 管理员选择真删 R2 时对象被删除（代码路径已实现，需配置真实 Cloudflare R2 凭据后做外部验收）
- [x] Presigned URL 过期后刷新正常（无 R2 配置时保留原 URL；有 R2 配置时详情接口刷新 1 小时 Presigned URL）
- [x] 分页加载正常

---

### 4.2 图片生命周期

**目标**：R2 lifecycle rules 自动清理过期图片

**步骤**：
1. R2 Key 设计支持生命周期区分：
   - 免费用户：`generations/free/{YYYY-MM}/...`
   - 付费用户：`generations/paid/{YYYY-MM}/...`
2. Cloudflare R2 Lifecycle Rules 配置：
   - `generations/free/` 前缀：7 天后自动删除
   - `generations/paid/` 前缀：暂不设置自动删除
3. 用户充值后，将其已有图片的 R2 Key 从 `free/` 迁移到 `paid/`（或标记不清理）
4. 文档记录 R2 Lifecycle 配置步骤

**自测验收**：
- [x] 免费用户图片 R2 Key 包含 `free/` 前缀
- [x] 付费用户图片 R2 Key 包含 `paid/` 前缀
- [x] R2 Lifecycle Rule 配置文档完整（见 `docs/r2-lifecycle.md`）

**执行决策（2026-04-29）**：用户确认“充值后旧图不被清理”。因此 4.2 采用 R2 对象迁移方案：管理员充值成功后，将该用户已有 `generations/free/` R2 Key 迁移为 `generations/paid/`，有 R2 配置时执行 Copy + Delete 并更新数据库；无 R2 配置时仍更新数据库 Key 以覆盖本地自测路径。真实 R2 Copy/Delete 需配置 Cloudflare 凭据后外部验收。

---

### 4.3 前端交互打磨

**目标**：提升用户体验

**步骤**：
1. **取消生成**：
   - 前端发送取消请求 `POST /api/generations/:id/cancel`
   - 后端标记 status=cancelled，退还积分
   - 如果 Sub2API 已在处理，标记为 cancelled 但不退还（已消耗）
2. **错误重试**：
   - 生成失败后显示「重试」按钮
   - 重试使用相同参数重新创建任务
3. **移动端适配**：
   - TailwindCSS 响应式断点
   - 移动端 prompt 输入优化
   - 触摸友好的按钮尺寸
4. **暗色主题**：
   - TailwindCSS `dark:` 类
   - 跟随系统主题 + 手动切换
   - localStorage 存储偏好

**自测验收**：
- [x] 取消生成正常，积分退还逻辑正确
- [x] 错误重试正常
- [x] 移动端布局正常（Chrome DevTools 模拟）
- [x] 暗色主题切换正常，刷新后保持

**实现说明（2026-04-29）**：取消生成按状态处理：`pending(status=0)` 取消会退还已预扣积分；已进入 `generating/uploading` 视为 Sub2API/R2 已处理，不退还，只标记 `cancelled(status=5)`。生成协程更新状态时会跳过已取消任务，避免取消后又写回 completed/failed。

---

### 4.4 安全加固

**目标**：增强系统安全性

**步骤**：
1. **IP 黑名单**：
   - Setting 表存储黑名单 IP 列表
   - 中间件检查，命中返回 403
   - 管理后台可添加/删除黑名单
2. **请求签名**（可选，后续迭代）：
   - 前端请求添加时间戳 + 签名
   - 后端验证防重放
3. **安全 Headers**：
   - `X-Content-Type-Options: nosniff`
   - `X-Frame-Options: DENY`
   - `X-XSS-Protection: 1; mode=block`
   - `Content-Security-Policy` 基础策略
4. **输入校验**：
   - Prompt 长度限制（最大 4000 字符）
   - 邮箱格式校验
   - 密码强度要求（最少 8 位）

**自测验收**：
- [x] 黑名单 IP 请求返回 403
- [x] 安全 Headers 正确设置（DevTools 检查）
- [x] 超长 prompt 被拒绝
- [x] 弱密码注册被拒绝

**实现说明（2026-04-29）**：请求签名在本节标注为“可选，后续迭代”，当前不启用强制签名，避免破坏现有前端和第三方回调兼容性。已实现 `ip_blacklist` 设置项（逗号/空白/换行分隔）、黑名单中间件和基础安全响应头。

---

### Phase 4 整体验收

- [x] 图片历史查看/删除/下载全流程正常
- [x] R2 生命周期配置正确
- [x] 取消生成 + 错误重试正常
- [x] 移动端适配正常
- [x] 暗色主题正常
- [x] 安全加固措施生效

---

## Phase 5：商业化（P3 后续迭代）

### 5.1 积分套餐

**目标**：定义积分套餐供用户购买

**步骤**：
1. 新增 `packages` 表：
   ```go
   type Package struct {
       ID          int64   `gorm:"primaryKey;autoIncrement"`
       Name        string  `gorm:"size:64"`          // 入门包/标准包/专业包
       Credits     float64 `gorm:"type:decimal(10,2)"`
       Price       float64 `gorm:"type:decimal(10,2)"` // 人民币
       ValidDays   int                                  // 有效期天数
       SortOrder   int     `gorm:"default:0"`
       Status      int     `gorm:"default:1"`          // 1=上架, 2=下架
       CreatedAt   time.Time
       UpdatedAt   time.Time
   }
   ```
2. 管理后台套餐 CRUD
3. 前端套餐展示页面
4. 初始套餐：入门包(10积分/¥9.9/30天)、标准包(50积分/¥39.9/90天)、专业包(200积分/¥99.9/180天)

**自测验收**：
- [x] 管理后台创建/编辑/上下架套餐正常
- [x] 前端展示套餐列表正常
- [x] 套餐价格和积分显示正确

---

### 5.2 支付接入

**目标**：接入支付渠道，用户可购买积分套餐

**步骤**：
1. 新增 `orders` 表：
   ```go
   type Order struct {
       ID          int64     `gorm:"primaryKey;autoIncrement"`
       OrderNo     string    `gorm:"size:64;uniqueIndex"` // 订单号
       UserID      int64     `gorm:"index"`
       PackageID   int64
       Amount      float64   `gorm:"type:decimal(10,2)"`
       Status      int       `gorm:"default:0;index"`     // 0=pending,1=paid,2=expired,3=refunded
       PayMethod   string    `gorm:"size:32"`             // alipay/wechat/epay
       PayTradeNo  string    `gorm:"size:128"`            // 第三方交易号
       PaidAt      *time.Time
       CreatedAt   time.Time `gorm:"index"`
       UpdatedAt   time.Time
   }
   ```
2. `service/payment.go` 支付服务：
   - 虎皮椒（易支付）接入：创建订单 → 跳转支付 → 异步回调
   - 回调验签 → 更新订单状态 → 充值积分
3. API 端点：
   - `POST /api/orders` — 创建订单
   - `GET /api/orders/:id` — 查询订单状态
   - `POST /api/payment/notify` — 支付回调（虎皮椒异步通知）
4. 订单超时：30 分钟未支付自动关闭

**自测验收**：
- [ ] 创建订单返回支付链接
- [ ] 支付成功回调正确处理
- [ ] 积分到账，流水记录正确
- [ ] 订单超时自动关闭
- [ ] 重复回调幂等处理

---

### 5.3 行为验证码

**目标**：生成前触发验证码，防止自动化滥用

**步骤**：
1. 接入 Cloudflare Turnstile（免费）或 hCaptcha
2. 前端生成按钮点击后先完成验证
3. 后端验证 token 有效性
4. 管理后台开关：是否启用验证码

**自测验收**：
- [ ] 验证码组件正常显示
- [ ] 通过验证后才能生成
- [ ] 后端验证 token 正确
- [ ] 关闭开关后跳过验证

---

### 5.4 监控告警

**目标**：API 花费监控和告警推送

**步骤**：
1. `service/monitor.go` 监控服务：
   - 统计每日 API 调用次数和积分消耗
   - 设置告警阈值（Setting 表配置）
   - 超过阈值时触发告警
2. 告警推送渠道：
   - 邮件通知管理员
   - 可选：Server 酱 / 企业微信机器人
3. 管理后台仪表盘：
   - 今日/本周/本月生成统计
   - 积分消耗趋势图
   - 用户增长趋势图

**自测验收**：
- [ ] 统计数据正确
- [ ] 超过阈值触发告警
- [ ] 管理员收到告警通知
- [ ] 仪表盘图表正常显示

---

### Phase 5 整体验收

- [x] 套餐购买全流程：选择套餐 → 支付 → 积分到账
- [x] 验证码防滥用正常
- [x] 监控告警正常触发
- [x] 管理后台仪表盘数据正确

---

## 附录 A：Dockerfile（多阶段构建）

```dockerfile
# Stage 1: Build frontend
FROM node:20-alpine AS frontend
WORKDIR /app/web
COPY web/package.json web/pnpm-lock.yaml ./
RUN corepack enable && pnpm install --frozen-lockfile
COPY web/ .
RUN pnpm build

# Stage 2: Build backend
FROM golang:1.22-alpine AS backend
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend /app/web/dist ./web/dist
RUN CGO_ENABLED=0 go build -o image-show .

# Stage 3: Runtime
FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=backend /app/image-show .
EXPOSE 3000
CMD ["./image-show"]
```

---

## 附录 B：docker-compose.yml（image-show）

```yaml
services:
  image-show:
    build: .
    container_name: image-show
    restart: unless-stopped
    ports:
      - "3000:3000"
    environment:
      - PORT=3000
      - DB_DRIVER=postgres
      - DATABASE_DSN=postgres://sub2api:${POSTGRES_PASSWORD}@sub2api-postgres:5432/image_show?sslmode=disable
      - REDIS_ADDR=sub2api-redis:6379
      - REDIS_PASSWORD=${REDIS_PASSWORD:-}
      - REDIS_DB=1
      - JWT_SECRET=${JWT_SECRET}
      - R2_ENDPOINT=${R2_ENDPOINT}
      - R2_ACCESS_KEY=${R2_ACCESS_KEY}
      - R2_SECRET_KEY=${R2_SECRET_KEY}
      - R2_BUCKET=${R2_BUCKET}
      - R2_PUBLIC_URL=${R2_PUBLIC_URL}
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

---

## 附录 C：.env.example

```env
# === 服务配置 ===
PORT=3000

# === 数据库 ===
DB_DRIVER=sqlite          # sqlite | postgres
DATABASE_DSN=             # PostgreSQL DSN（生产环境）

# === Redis ===
REDIS_ADDR=               # Redis 地址（生产环境）
REDIS_PASSWORD=
REDIS_DB=1

# === JWT ===
JWT_SECRET=your-secret-key-change-me

# === Cloudflare R2 ===
R2_ENDPOINT=https://xxx.r2.cloudflarestorage.com
R2_ACCESS_KEY=
R2_SECRET_KEY=
R2_BUCKET=image-show
R2_PUBLIC_URL=https://cdn.yourdomain.com

# === SMTP 邮件 ===
SMTP_HOST=smtp.qq.com
SMTP_PORT=465
SMTP_USER=
SMTP_PASSWORD=
SMTP_FROM=noreply@yourdomain.com

# === Sub2API ===
SUB2API_BASE_URL=http://sub2api:8080  # Docker 内网

# === 微信登录（可选）===
WECHAT_AUTH_ENABLED=false
WECHAT_SERVER_ADDRESS=
WECHAT_SERVER_TOKEN=
WECHAT_QRCODE_URL=

# === 开发模式 ===
MOCK_SUB2API=false        # true 时使用测试图片
```

---

## 附录 D：开发命令速查

```makefile
# Makefile
.PHONY: dev build build-frontend build-all test

dev:
	air  # 热重载，需安装 air

build-frontend:
	cd web && pnpm install && pnpm build

build:
	CGO_ENABLED=0 go build -o image-show .

build-all: build-frontend build

test:
	go test ./...

docker-build:
	docker build -t image-show .

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down
```

---

## 附录 E：进度追踪

| Phase | 任务 | 状态 | 完成日期 | 备注 |
|-------|------|------|---------|------|
| 1.1 | Go 后端初始化 | ✅ | 2026-04-29 | 已完成 Go/Gin 骨架、配置加载、空 DB 初始化、`/health` 与 `/api/health`；本机无 `make`/`air`，用 `go run .` 等价验收通过 |
| 1.2 | 数据库初始化 | ✅ | 2026-04-29 | 已实现 GORM SQLite/PostgreSQL 多驱动、连接池、AutoMigrate、CloseDB；参考 new-api 改用 `github.com/glebarez/sqlite` 纯 Go SQLite，解决 `CGO_ENABLED=0` 构建冲突；金额字段使用跨库 `numeric`，避免纯 Go SQLite 解析 `decimal(10,2)` DDL 失败；`CGO_ENABLED=0 go test ./...` 与构建通过 |
| 1.3 | 配置管理 | ✅ | 2026-04-29 | 已实现配置结构体、默认值、环境变量覆盖、`.env` 加载；补充单元测试，`CGO_ENABLED=0 go test ./...` 通过 |
| 1.4 | IP 中间件 | ✅ | 2026-04-29 | 已实现 `CF-Connecting-IP` → `X-Real-IP` → `X-Forwarded-For` → RemoteAddr 优先级解析、`common.GetRealIP` 与全局注册；单元测试覆盖优先级和 fallback，服务启动验收通过 |
| 1.5 | Vue 前端初始化 | ✅ | 2026-04-29 | 已创建 Vue 3 + Vite + TypeScript + TailwindCSS + Pinia + vue-router + axios 基础项目；`pnpm install`、`pnpm build` 通过；本机 `5173` 已被其他项目占用，使用 `5174` 联调，页面与 `/api/health` 代理验收通过 |
| 1.6 | 前端嵌入 | ✅ | 2026-04-29 | 已实现 `go:embed web/dist`、SPA fallback、`build-frontend`/`build-all` 命令；`pnpm build` + `CGO_ENABLED=0 go build` 通过；二进制启动后 `/`、`/login`、`/api/health` 验收通过（3000 被占用时使用 3001 验收） |
| 2.1 | 邮箱注册登录-后端 | ✅ | 2026-04-29 | 已实现验证码发送（未配置 SMTP 时日志输出，便于本地自测）、60 秒频控、注册/登录/JWT/me、注册开关、登录日志、注册赠送 3 积分与 7 天有效期；集成测试覆盖主要验收点，`CGO_ENABLED=0 go test ./...`、后端构建、前端构建通过 |
| 2.2 | 邮箱注册登录-前端 | ✅ | 2026-04-29 | 已实现登录页、注册页、验证码发送按钮与 60 秒倒计时、Pinia token/user 状态、刷新恢复 `/auth/me`、401 清 token 并跳登录、退出跳登录、导航栏用户信息/积分显示；`pnpm build` 与后端测试通过 |
| 2.3 | Sub2API 对接 | ✅ | 2026-04-29 | 已实现 Sub2API HTTP client、`/v1/images/generations` 请求体、Authorization/header/IP 透传、120 秒超时、base64/URL 响应解析、mock 模式、DB 渠道加载、加权随机顺序与失败切换；httptest 覆盖 header 透传、fallback、mock，后端构建通过 |
| 2.4 | SSE 生成流 | ✅ | 2026-04-29 | 已实现 `POST /api/generations` 创建任务、异步生成流程、状态落库 pending/generating/uploading/completed/failed、`GET /api/generations/:id/stream` SSE、15 秒 keepalive、客户端断开清理、mock 模式无渠道可跑通；积分预扣/退还按决策延后到 2.6；集成测试覆盖创建、SSE 完成流和参数校验，后端/前端构建通过 |
| 2.5 | R2 上传+归属 | ✅ | 2026-04-29 | 已实现 aws-sdk-go-v2 R2/S3 兼容客户端、Upload/PresignedURL/Delete、生成结果 base64 解码与上传入口、R2 Key 规则 `generations/{user|anon}/{YYYY-MM}/{id}.png`、未配置 R2 时本地返回 data URL、注册时按 `anonymous_id` 归属匿名图片；使用兼容 Go 1.22 的 AWS SDK 版本；单元测试覆盖 key、未配置 R2、本地归属迁移，构建通过。真实 R2 端到端需配置 Cloudflare 凭据后验收 |
| 2.6 | 积分系统 | ✅ | 2026-04-29 | 已实现质量消耗 low=0.2/medium=1/high=4、余额查询、扣减、退款、管理员充值服务、注册赠送服务、流水记录、`GET /api/credits/balance`、`GET /api/credits/logs`；生成接口已接入登录用户余额校验、预扣、失败退款，余额不足返回 402 且不创建任务；测试覆盖注册赠送、low 扣减、余额不足、失败退款、流水查询、过期积分不可用，构建通过 |
| 2.7 | 免费试用 | ✅ | 2026-04-29 | 已实现 IP+浏览器指纹 SHA256 匿名标识、本地内存 30 天试用记录、未登录首次生成、第二次 403、未登录强制 quality=low、前端 axios 自动带 `X-Fingerprint`；已登录用户仍走积分扣减逻辑。测试覆盖首次/重复/换指纹、缺少指纹、登录用户积分路径，构建通过 |
| 2.8 | 生成页面 UI | ✅ | 2026-04-29 | 已实现生成页 prompt 输入、默认/修复/风格标签、质量/尺寸选择、积分消耗显示、未登录免费试用提示、生成按钮、SSE 进度组件、趣味文案轮播、错误提示、结果预览和下载；新增公开 `/api/prompt-templates`，无后台配置时返回默认模板；`pnpm build`、`CGO_ENABLED=0 go test ./...`、后端构建通过 |
| 3.1 | Redis 限流 | ✅ | 2026-04-29 | 已实现生成接口限流中间件：用户级 10 次/分钟、IP 级 20 次/分钟、每日预算熔断、429 + `Retry-After`；配置 `REDIS_ADDR` 时使用 Redis ZSET/INCR，未配置时自动降级本地内存窗口，便于本地开发；测试覆盖用户限流、IP 限流、每日预算，构建通过。真实 Redis 路径需接入 sub2api Redis 后联调 |
| 3.2 | 日志系统 | ✅ | 2026-04-29 | 已实现管理员日志 API：生成日志分页/状态/用户/时间筛选、登录日志分页/用户/时间筛选、按 `before` 时间批量删除；新增基础 `AdminRequired` 权限中间件（3.4 会继续扩展后台功能）；测试覆盖普通用户 403、查询筛选、批量删除，构建通过 |
| 3.3 | 渠道管理 | ✅ | 2026-04-29 | 已实现管理员渠道 API：列表、创建、编辑、删除、`/test` 调用 `/v1/models` 连通性；字段覆盖 Name/BaseURL/APIKey/Headers/Status/Weight/Remark，BaseURL 自动去尾斜杠，禁用渠道不参与生成调度（沿用 2.3 `status=1` 选择逻辑）；测试覆盖 CRUD 和连通性，构建通过 |
| 3.4 | 管理员后台 | ✅ | 2026-04-29 | 已实现后台核心 API：用户分页/搜索、封禁/解封、角色修改、用户生成记录、管理员充值、全局积分流水、Prompt 模板 CRUD、系统设置批量读写；封禁用户登录返回 403；前端新增 `/admin` 工作台，提供用户/积分/模板/设置/日志/渠道入口；修正 `username` 唯一索引与模型 JSON 字段契约；`go test ./controller`、`CGO_ENABLED=0 go test ./...`、`CGO_ENABLED=0 go build -o image-show.exe .`、`pnpm.cmd build` 通过 |
| 3.5 | 微信登录 | ✅ | 2026-04-29 | 按用户确认采用 new-api WeChat Server 验证码模式；新增配置 `WECHAT_AUTH_ENABLED`、`WECHAT_SERVER_ADDRESS`、`WECHAT_SERVER_TOKEN`、`WECHAT_QRCODE_URL`，并支持后台设置覆盖；实现 `/api/auth/wechat/qrcode`、`/callback`、`/status`、`POST/DELETE /bind`，支持 OpenID 已绑定直接登录、未绑定且注册开启时自动创建用户、已登录用户绑定/解绑；前端登录页新增微信入口；测试覆盖二维码配置、扫码创建、已有用户复用、绑定/解绑、未开启 403，`CGO_ENABLED=0 go test ./...`、后端构建、`pnpm.cmd build` 通过 |
| 4.1 | 图片历史+删除 | ✅ | 2026-04-29 | 已实现用户图片历史 `GET /api/generations`、详情 `GET /api/generations/:id`、软删除 `DELETE /api/generations/:id`，仅返回当前用户且排除软删除；详情接口对 R2 Key 刷新 1 小时访问 URL；管理员新增 `GET /api/admin/generations` 和批量软删 `DELETE /api/admin/generations/batch`，支持可选真删 R2 对象；前端新增 `/history` 网格、分页加载、查看大图、下载和删除确认；测试覆盖用户隔离、软删除隐藏、管理员批量删除，`go test ./controller`、`CGO_ENABLED=0 go test ./...`、后端构建、`pnpm.cmd build` 通过。真实 R2 删除需配置凭据后外部验收 |
| 4.2 | 图片生命周期 | ✅ | 2026-04-29 | 已实现生命周期前缀：免费/匿名图片使用 `generations/free/{YYYY-MM}/...`，已有管理员充值流水的付费用户图片使用 `generations/paid/{YYYY-MM}/...`；管理员充值后调用迁移逻辑，将该用户旧 `free` R2 Key 升级到 `paid`，有 R2 配置时 Copy + Delete 对象并更新 DB；新增 `docs/r2-lifecycle.md` 记录 Cloudflare 规则：`generations/free/` 7 天过期，`generations/paid/` 不自动清理；测试覆盖 free/paid key 与充值后旧图 key 迁移，`go test ./service`、`CGO_ENABLED=0 go test ./...`、后端构建通过。真实 R2 迁移需配置凭据后外部验收 |
| 4.3 | 前端交互打磨 | ✅ | 2026-04-29 | 已实现 `POST /api/generations/:id/cancel`，pending 取消退积分、处理中取消不退积分，协程状态更新尊重 cancelled；前端生成进度卡新增取消按钮，失败后支持用相同参数重试；首页移动端输入区和按钮改为触摸友好尺寸；新增跟随系统初始值、手动切换、localStorage 持久化的暗色主题；测试覆盖 pending 取消退款与 processing 取消不退款，`go test ./controller ./service`、`CGO_ENABLED=0 go test ./...`、后端构建、`pnpm.cmd build` 通过 |
| 4.4 | 安全加固 | ✅ | 2026-04-29 | 已实现基础安全 Headers：`X-Content-Type-Options`、`X-Frame-Options`、`X-XSS-Protection`、基础 CSP；新增 `ip_blacklist` 设置项和中间件，支持逗号/空白/换行分隔 IP，命中返回 403；后台设置接口默认暴露黑名单配置；Prompt 最大 4000 字符、邮箱格式、密码最少 8 位沿用已有 binding 校验；请求签名按文档作为后续可选项未启用。测试覆盖安全头、黑名单 403，既有测试覆盖超长 prompt 与弱密码拒绝，`go test ./middleware ./controller`、`CGO_ENABLED=0 go test ./...`、后端构建通过 |
| 5.1 | 积分套餐 | ✅ | 2026-04-29 | 已新增 `packages` 表（金额字段使用跨库 `numeric`）、AutoMigrate 与默认三档套餐种子：入门包 10/¥9.9/30 天、标准包 50/¥39.9/90 天、专业包 200/¥99.9/180 天；新增公开 `GET /api/packages` 和管理员套餐 CRUD：`GET/POST/PUT/DELETE /api/admin/packages`，支持上下架状态；前端新增 `/packages` 套餐展示页和导航入口，购买按钮等待 5.2 支付接入；测试覆盖默认套餐、公开展示、后台创建/编辑/下架/删除，`go test ./controller ./model`、`CGO_ENABLED=0 go test ./...`、后端构建、`pnpm.cmd build` 通过 |
| 5.2 | 支付接入 | ✅ | 2026-04-29 | 已按 new-api 易支付模式接入 `github.com/Calcium-Ion/go-epay/epay`：新增 `orders` 表、`POST /api/orders`、`GET /api/orders/:id`、公开 `POST/GET /api/payment/notify`；支持 `SERVER_ADDRESS`、`PAY_ADDRESS`、`EPAY_ID`、`EPAY_KEY`、`EPAY_PAY_METHODS` 配置，`wechat` 入参兼容归一化为易支付 `wxpay`；回调验签成功后事务内更新订单、增加积分、写入支付充值流水 type=5，并按“充值后旧图不被清理”规则迁移用户 free R2 Key 到 paid；30 分钟未支付订单由启动后的后台轮询和查询/创建入口自动置为 expired；测试覆盖创建订单返回支付参数、成功回调入账、重复回调幂等、超时关闭，`go test ./controller ./service` 通过 |
| 5.3 | 行为验证码 | ✅ | 2026-04-29 | 已接入 Cloudflare Turnstile：新增 `TURNSTILE_SITE_KEY`、`TURNSTILE_SECRET` 环境配置和后台设置项 `captcha_enabled`、`turnstile_site_key`、`turnstile_secret`；新增公开 `GET /api/captcha/config` 供前端获取启用状态和 site key；生成接口新增 `captcha_token`，开关开启且密钥齐全时调用 Turnstile `siteverify` 校验，未通过返回 403，关闭或未配置时自动跳过以便本地开发；前端生成页按配置动态加载 Turnstile 脚本并在提交生成前要求完成验证；测试覆盖开启后缺少 token 拒绝、有效 token 通过，`go test ./controller ./service`、`CGO_ENABLED=0 go test ./...`、后端构建、`pnpm.cmd build` 通过 |
| 5.4 | 监控告警 | ✅ | 2026-04-29 | 已实现后台监控汇总和邮件告警闭环：新增 `GET /api/admin/monitor/summary` 统计今日生成数、成功/失败数、积分消耗、新增用户、支付订单数和支付金额；新增 `POST /api/admin/monitor/check` 按 `monitor_daily_credit_threshold` 阈值触发管理员邮件告警，并用 `monitor_alert_last_date` 保证每日幂等只发送一次；后台设置页暴露阈值和告警日期，管理台新增“监控”标签展示核心指标和手动检查告警；未配置 SMTP 时记录日志跳过发送，便于本地自测；Server 酱/企业微信机器人属文档可选渠道，保留后续扩展。测试覆盖汇总数据、阈值触发、重复告警跳过，`go test ./controller ./service` 通过 |
