# GPT-Image-2 图片生成网站 — 需求审核与优先级规划

> 确认日期：2026-04-29
> 状态：需求审核阶段（v2，合并用户反馈）

---

## 一、矛盾点解决

| # | 矛盾项 | 需求文档原文 | 用户新要点 | 最终决议 |
|---|--------|------------|-----------|---------|
| 1 | 数据库 | SQLite | PG + Redis 共用 sub2api | **用 sub2api 的 PostgreSQL 18 + Redis 8**，放弃 SQLite |
| 2 | 前端部署 | Cloudflare Pages | 前后端同容器 | **前端 embed 进 Go 二进制**（参考 new-api/sub2api 做法），前面用 Caddy 反代 |
| 3 | 未登录试用 | 无此设计 | 免费试用 1 次 | **未登录可通过 IP+指纹免费生成 1 次** |
| 4 | 图片修复 | MVP 仅文生图 | prompt 图片修复选项 | **MVP 包含**：本质是文生图 + 预设 prompt 选项，不是真正的图片编辑 |
| 5 | 限流实现 | Go 内存限流 | 用同一个 Redis | **Redis 限流**，共用 sub2api 的 Redis 实例 |
| 6 | ORM | 待定 Ent/GORM | 用户明确 GORM | **GORM**，与 new-api 一致，快速上手 |

---

## 二、需求要点统一（15 条逐项确认）

| # | 用户要点 | 最终确认 |
|---|---------|---------|
| 1 | sub2api 内部调用 | 同一 Docker 网络内，image-show 通过 `http://sub2api:8080` 内网调用，不走公网。渠道设置参考 new-api 的渠道管理模式（API Key + Base URL + Header 透传配置） |
| 2 | PG + Redis 共用 | 共用 sub2api 的 PG（独立数据库 `image_show`）和 Redis（Key 加 `imgshow:` 前缀，用独立 DB 编号如 db=1） |
| 3 | IP 记录和透传 | Caddy 层透传 `X-Real-IP` / `X-Forwarded-For` / `CF-Connecting-IP`（参考 sub2api 的 Caddyfile）。后端优先读 `CF-Connecting-IP → X-Real-IP → X-Forwarded-For`。IP 记录到生成日志和登录日志。参考 new-api 渠道 header 透传机制 |
| 4 | 日志记录和查询 | 生成日志、登录日志、API 调用日志。管理后台支持按时间范围查询和**批量删除**（指定时间条件删除，防止日志过大） |
| 5 | 渠道添加 | 管理后台支持添加/编辑/删除 sub2api 渠道（API Key、Base URL、状态开关） |
| 6 | 微信登录 + 邮箱登录 | 微信登录参考 new-api 实现（公众号测试号或 WxPusher）。邮箱注册登录 + SMTP 配置。后台注册开关，参考 new-api 模式 |
| 7 | Prompt 默认 + UI | 提供默认 prompt 模板标签（管理后台可配置）。**图片修复选项**本质是预设 prompt 模板（如"修复模糊"、"提升分辨率"等），MVP 阶段以文生图 prompt 选项形式实现 |
| 8 | 图片展示 + SSE | SSE 推送生成状态。前端展示：排队动画 → 进度提示 → 图片预览 |
| 9 | 容器部署 + 本地开发 | Docker 容器部署（与 sub2api 同一 compose 网络）。**本地开发用 SQLite**（零依赖快速调试），生产环境用 PG。GORM 多驱动兼容 |
| 10 | 安全限制 + 限流 + IP | Redis 限流：用户级 10次/分 + IP 级 20次/分。每日 API 预算熔断。管理员封禁 |
| 11 | 免费试用 + 积分体系 | 未登录 1 次免费（IP+指纹）→ 注册送 3 积分（7天有效）→ 消耗：low 0.2 / medium 1 / high 4 → 管理员手动充值 → 后续开放套餐 |
| 12 | 前端交互 | 图片生成慢（~30s），**不给预估时间**，改为趣味状态文案轮播：「正在创建图片」→「正在打草稿」→「生成初稿中」→「正在设置场景」→「正在润饰细节」→「即将完成」→「正在做最后润色」→「最后微调一下」。配合骨架屏/加载动画，生成中可取消，错误友好提示 |
| 13 | Prompt 图片修复选项 | **前端给选项即可**，如"修复模糊"、"提升清晰度"、"增强色彩"等标签，点击后拼接到 prompt 中。后台可配置选项列表。不涉及图片上传编辑 |
| 14 | Cloudflare R2 存储 | aws-sdk-go-v2 对接 R2，Presigned URL 防盗链，CDN 分发不走服务器带宽。**支持删除 R2 图片**（管理后台批量删除 + 用户删除自己的图片） |
| 15 | 同服务器容器部署 | image-show 加入 sub2api 的 Docker 网络（external network），共享 PG/Redis。前面统一用 Caddy 反代 |
| 16 | 用户图片归属追踪 | **新增**：每张图片记录生成者标识（用户ID / 匿名标识 IP+指纹）。统计每个用户生成的图片数量。匿名用户注册后，将其匿名期间生成的图片归属到新账号。管理后台可按用户查看/删除图片 |

---

## 三、最终技术选型

| 组件 | 选型 | 说明 |
|------|------|------|
| 后端框架 | Go + Gin | 与 new-api/sub2api 一致 |
| 前端框架 | Vue 3 + Vite + TailwindCSS | 与 sub2api 前端栈一致 |
| 数据库 | PostgreSQL 18 | 共用 sub2api 实例，独立数据库 `image_show` |
| 缓存/限流 | Redis 8 | 共用 sub2api 实例，Key 前缀 `imgshow:`，可用独立 DB |
| ORM | GORM v2 | 与 new-api 一致，快速上手 |
| 对象存储 | Cloudflare R2 | aws-sdk-go-v2，S3 兼容 |
| 认证 | JWT | 无状态 |
| 实时通信 | SSE | 生成进度推送 |
| 反向代理 | Caddy | 参考 sub2api 的 Caddyfile，统一管理 TLS + 反代 |
| 前端嵌入 | go:embed | 参考 new-api 的 `//go:embed web/dist` 方式 |
| 部署 | Docker + docker-compose | 加入 sub2api 网络 |
| 包管理 | pnpm (前端) | 与 sub2api 一致 |

---

## 四、系统架构

```
┌─────────────────────────────────────────────────────────────────┐
│  日本腾讯轻量云 2C4G 30Mbps                                       │
│                                                                  │
│  ┌─────────────────────────────────────────────────────────────┐ │
│  │  Caddy (反向代理 + TLS + 压缩 + IP 透传)                      │ │
│  │  image.xxx.com → image-show:3000                             │ │
│  │  api.xxx.com   → sub2api:8080                                │ │
│  └────────────┬──────────────────────────┬─────────────────────┘ │
│               │                          │                        │
│  ┌────────────▼────────────┐  ┌──────────▼──────────────┐        │
│  │  image-show 容器 :3000   │  │  sub2api 容器 :8080      │        │
│  │  ┌────────────────────┐ │  │  (已有，API Key 分发)     │        │
│  │  │ Go 后端 (Gin)       │ │  └──────────┬─────────────┘        │
│  │  │ ├─ JWT 认证         │ │             ▲                       │
│  │  │ ├─ 积分管理         │ │  内网调用    │                       │
│  │  │ ├─ SSE 生成流       │─┼─────────────┘                       │
│  │  │ ├─ Redis 限流       │ │                                     │
│  │  │ └─ 管理后台 API     │ │                                     │
│  │  ├────────────────────┤ │  ┌────────────────────────────────┐ │
│  │  │ Vue 3 前端 (embed)  │ │  │  PostgreSQL 18 (共用)           │ │
│  │  └────────────────────┘ │  │  ├─ DB: sub2api                 │ │
│  └─────────────────────────┘  │  └─ DB: image_show              │ │
│                                └────────────────────────────────┘ │
│  ┌─────────────────────────────────────────────────────────────┐ │
│  │  Redis 8 (共用)                                              │ │
│  │  ├─ DB0: sub2api                                             │ │
│  │  └─ DB1: image_show (imgshow: prefix)                        │ │
│  └─────────────────────────────────────────────────────────────┘ │
│                                                                   │
│  Docker Network: sub2api-network (bridge, 所有容器共享)             │
└───────────────────────────────────────────────────────────────────┘
                    │
                    │ 上传图片 (aws-sdk-go-v2)
                    ▼
          ┌──────────────────┐
          │  Cloudflare R2   │
          │  + CDN 分发       │
          │  (Presigned URL) │
          └──────────────────┘
```

---

## 五、优先级排列与开发阶段

### P0 — 基础骨架（第 1 周）

| # | 任务 | 对应要点 | 说明 |
|---|------|---------|------|
| 1 | 项目初始化 + 容器化 | 9, 15 | Go+Gin 后端 + Vue3 前端，Dockerfile（多阶段），docker-compose 加入 sub2api 网络，Makefile（dev/build），Caddy 配置。**本地开发用 SQLite，生产用 PG**（GORM 多驱动） |
| 2 | 数据库设计与连接 | 2 | GORM 多驱动（本地 SQLite / 生产 PG），建表：users, credits, credit_logs, generations, login_logs, channels, settings, **anonymous_identities**（匿名标识追踪） |
| 3 | 邮箱注册/登录 | 6 | SMTP 配置、邮箱验证码、JWT 签发、注册开关（管理后台控制） |
| 4 | IP 记录与透传 | 3 | Caddy 透传 header，Go 中间件解析真实 IP，注入请求上下文 |
| 5 | 基础前端布局 | 12 | Vue 3 + TailwindCSS + vue-router + Pinia，响应式布局，登录/注册页 |

### P0 — 核心生成功能（第 2 周）

| # | 任务 | 对应要点 | 说明 |
|---|------|---------|------|
| 6 | Sub2API 内网对接 | 1 | 通过 Docker 内网 `http://sub2api:8080` 调用 gpt-image-2，封装 HTTP 客户端 |
| 7 | SSE 生成流 | 8 | 后端 SSE 推送：`queued → generating → uploading → completed/failed`，前端 EventSource 接收 |
| 8 | 图片上传 R2 + 归属追踪 | 14, 16 | 生成完成后上传 R2，返回 Presigned URL。记录生成者标识（用户ID / 匿名 IP+指纹）。匿名用户注册后自动归属图片 |
| 9 | 积分系统 | 11 | 积分扣减/查询/流水记录，质量消耗比例 low 0.2 / medium 1 / high 4 |
| 10 | 未登录免费试用 | 11 | IP + 浏览器指纹，Redis 记录，限 1 次/30 天 |
| 11 | 生成页面 UI | 7, 8, 12, 13 | Prompt 输入 + 默认模板标签 + 图片修复预设选项（前端标签拼接 prompt）+ 参数选择 + 趣味状态文案轮播动画 |

### P1 — 防护与管理（第 3 周）

| # | 任务 | 对应要点 | 说明 |
|---|------|---------|------|
| 12 | Redis 限流 | 10 | 用户级 10次/分 + IP 级 20次/分 + 全局每日预算熔断 |
| 13 | 日志系统 | 4 | 生成日志、登录日志、API 调用日志。管理后台查询 + 按时间条件批量删除 |
| 14 | 渠道管理 | 5 | 管理后台 CRUD sub2api 渠道（API Key、Base URL、状态、备注） |
| 15 | 管理员后台 | 5, 7 | 用户管理/封禁、积分充值、参数配置、Prompt 模板管理、注册开关、系统设置 |
| 16 | 微信登录 | 6 | 参考 new-api 微信登录实现，WxPusher 或公众号测试号 |

### P2 — 体验优化（第 4 周）

| # | 任务 | 说明 |
|---|------|------|
| 17 | 图片历史 + R2 删除 | 用户查看/下载/删除生成历史（删除同步删 R2 文件），管理后台按用户查看/批量删除图片，分页加载 |
| 18 | 图片生命周期 | R2 lifecycle rules，免费 7 天 / 付费 90 天 |
| 19 | 前端交互打磨 | 取消生成、错误重试、移动端适配、暗色主题 |
| 20 | 安全加固 | IP 黑名单、请求签名、Cloudflare WAF |

### P3 — 商业化（后续迭代）

| # | 任务 | 说明 |
|---|------|------|
| 21 | 积分套餐 | 入门包/标准包/专业包 |
| 22 | 支付接入 | 虎皮椒/支付宝/微信支付 |
| 23 | 行为验证码 | 生成前触发 |
| 24 | 监控告警 | API 花费告警推送 |

---

## 六、关键设计决策

### 6.1 容器网络与部署
- sub2api 已有 `sub2api-network` 网络，image-show 通过 `external: true` 加入
- image-show 的 docker-compose 不再定义 PG/Redis 服务，直接引用 sub2api 的
- Caddy 统一反代：`image.xxx.com → image-show:3000`，`api.xxx.com → sub2api:8080`
- 参考 sub2api 的 Caddyfile 配置 TLS、压缩、IP 透传

### 6.2 数据库隔离
- 在 sub2api 的 PG 实例中创建独立数据库 `image_show`（不是独立 schema，是独立 DB）
- GORM AutoMigrate 管理表结构
- Redis 使用独立 DB 编号（如 db=1），或统一 db=0 + `imgshow:` 前缀

### 6.3 Sub2API 内网调用
- Docker 内网直连 `http://sub2api:8080`，零公网延迟
- 渠道设置参考 new-api 模式：管理后台配置 API Key + Base URL + Header 透传
- 调用 sub2api 的 OpenAI 兼容接口 `/v1/images/generations`
- 透传用户 IP 到 sub2api（通过 header），sub2api 会记录到使用日志

### 6.4 IP 透传链路
```
用户 → Cloudflare CDN → Caddy → image-show Go 后端
                                    ↓
                              sub2api (内网)
```
- Caddy 设置 `X-Real-IP`、`X-Forwarded-For`、保留 `CF-Connecting-IP`
- Go 后端读取优先级：`CF-Connecting-IP → X-Real-IP → X-Forwarded-For → RemoteAddr`
- 调用 sub2api 时透传 IP header，sub2api 会记录到日志

### 6.5 前端嵌入方式
- 参考 new-api 的 `//go:embed web/dist` + `indexPage` 注入方式
- 构建时 Vue 产物嵌入 Go 二进制，单容器部署
- 本地开发时前后端分离：Go 后端 `:3000` + Vite dev server `:5173`（CORS 配置）

### 6.6 本地开发 vs 生产环境
- **本地开发**：SQLite（零依赖，`make dev` 即可启动），无需 PG/Redis
- **生产环境**：PostgreSQL 18 + Redis 8（共用 sub2api 实例）
- GORM 多驱动：通过环境变量 `DB_DRIVER=sqlite|postgres` 切换
- 本地限流用内存，生产用 Redis

### 6.7 SSE 趣味状态文案
- 不给预估时间，改为趣味状态文案定时轮播（每 3-5 秒切换）：
  1. 「正在创建图片...」
  2. 「正在打草稿...」
  3. 「生成初稿中...」
  4. 「正在设置场景...」
  5. 「正在润饰细节...」
  6. 「即将完成...」
  7. 「正在做最后润色...」
  8. 「最后微调一下...」
- 后端 SSE 推送真实状态（queued/generating/uploading/completed/failed）
- 前端在 generating 阶段展示文案轮播，completed 时展示图片

### 6.8 用户图片归属与匿名追踪
- generations 表包含 `user_id`（可为空）和 `anonymous_id`（IP+指纹 hash）
- 匿名用户生成时记录 `anonymous_id`
- 用户注册后，通过 `anonymous_id` 匹配，将匿名期间的图片 `user_id` 更新为新用户 ID
- 管理后台可按用户/匿名标识查看图片统计和列表

### 6.9 R2 图片删除
- 用户删除自己的图片：**DB 标记软删除（对用户不可见），R2 实际不删除**（保留数据，防误删）
- 管理后台批量删除：按用户/时间范围批量删除（管理员可选择是否真删 R2 对象）
- R2 lifecycle rules 作为兜底：免费用户 7 天自动清理，**付费用户暂不清理**

### 6.10 日志管理
- 日志存 PG 表，不存文件
- 管理后台支持：按时间范围查询、按条件筛选、按时间条件批量删除
- 参考 new-api 的日志删除接口设计

---

## 七、验证方案

1. **本地开发**：`make dev` + `pnpm dev`，完成注册→登录→生成→查看历史全流程
2. **SSE 验证**：DevTools Network 确认 EventStream 连接和状态推送
3. **内网调用**：确认 image-show 容器可通过 `http://sub2api:8080` 访问 sub2api
4. **限流验证**：快速连续请求，确认 Redis 限流返回 429
5. **未登录试用**：无 Cookie 生成 1 次，第 2 次被拒
6. **容器部署**：`docker-compose up` 全部服务，确认 PG/Redis 共用正常
7. **R2 上传**：生成图片后确认 R2 存储和 Presigned URL 可访问
8. **IP 透传**：检查生成日志中记录的 IP 是否为用户真实 IP
