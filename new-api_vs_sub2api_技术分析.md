# New API 与 Sub2API 技术分析文档

## 一、项目概览

| 维度 | New API | Sub2API |
|------|---------|---------|
| 定位 | 新一代大模型网关与AI资产管理系统 | AI API 网关平台，面向订阅配额分发 |
| 核心能力 | 聚合 40+ 上游 AI 供应商，统一转换为 OpenAI/Claude/Gemini 兼容格式 | 将 AI 产品订阅配额转化为 API Key 分发，处理认证、计费、负载均衡 |
| 开源协议 | AGPLv3（基于 One API MIT） | 有 CLA 要求 |
| 前身 | 基于 One API 二次开发 | 独立项目 |

---

## 二、技术栈对比

### 2.1 后端

| 维度 | New API | Sub2API |
|------|---------|---------|
| 语言 | Go 1.25+ | Go 1.26+ |
| Web 框架 | Gin | Gin |
| ORM / 数据层 | GORM v2 | Ent (entgo.io) |
| 数据库 | SQLite / MySQL ≥5.7.8 / PostgreSQL ≥9.6（三库兼容） | PostgreSQL 15+（单库） |
| 缓存 | Redis (go-redis/v8) + 内存缓存 | Redis 7+ (go-redis/v9) + Ristretto 本地缓存 |
| 依赖注入 | 无（手动初始化） | Google Wire（编译期依赖注入） |
| 日志 | 自定义 logger | Zap + Lumberjack（结构化日志 + 日志轮转） |
| 认证 | JWT + WebAuthn/Passkeys + OAuth (GitHub/Discord/OIDC/Telegram/微信) | JWT + OAuth (LinuxDO/OIDC/微信) + TOTP |
| 支付 | EPay + Stripe + Creem + Waffo | EasyPay + 支付宝 + 微信支付 + Stripe |
| HTTP 客户端 | 标准库 net/http | imroc/req/v3 + refraction-networking/utls (TLS 指纹) |
| 定时任务 | gopool + 自定义轮询 | robfig/cron/v3 |
| WebSocket | gorilla/websocket | gorilla/websocket + coder/websocket |
| 性能分析 | pprof + Pyroscope | gopsutil/v4 系统监控 |
| 测试 | 少量单元测试 | 大量单元测试 + testcontainers 集成测试 |

### 2.2 前端

| 维度 | New API | Sub2API |
|------|---------|---------|
| 框架 | React 18 | Vue 3.4+ |
| 构建工具 | Vite | Vite 5+ |
| UI 库 | Semi Design (@douyinfe/semi-ui) | TailwindCSS |
| 包管理器 | Bun | pnpm |
| 状态管理 | React Context | Pinia |
| 图表 | VChart (VisActor) | Chart.js + vue-chartjs |
| 国际化 | i18next + react-i18next (zh/en/fr/ru/ja/vi) | vue-i18n |
| 类型检查 | 无 (JSX) | TypeScript + vue-tsc |
| 测试 | 无 | Vitest + @vue/test-utils |

### 2.3 部署

| 维度 | New API | Sub2API |
|------|---------|---------|
| 容器化 | 多阶段 Docker (Bun → Go → Debian slim) | 多阶段 Docker (Node → Go → Alpine) |
| 编排 | docker-compose | docker-compose |
| CI/CD | GitHub Actions (Docker 镜像 + Electron 桌面端 + Release) | GitHub Actions (CI + Release + 安全扫描) |
| 发布 | Docker Hub | GoReleaser + Docker |
| 桌面端 | Electron 打包 | 无 |
| 安装脚本 | 无 | 一键安装脚本（Linux amd64/arm64） |

---

## 三、架构设计对比

### 3.1 New API — 分层 + 适配器模式

```
router/         — HTTP 路由 (API / relay / dashboard / web)
controller/     — 请求处理器
service/        — 业务逻辑
model/          — 数据模型 + DB 访问 (GORM)
relay/          — AI API 代理转发（核心）
  relay/channel/  — 30+ 供应商适配器 (openai/ claude/ gemini/ aws/ ali/ ...)
middleware/     — 认证、限流、CORS、日志、渠道分发
setting/        — 配置管理 (ratio / model / operation / system / performance)
common/         — 共享工具 (JSON / crypto / Redis / env / rate-limit)
dto/            — 数据传输对象
constant/       — 常量 (API 类型 / 渠道类型 / 上下文 Key)
types/          — 类型定义 (relay 格式 / 文件源 / 错误)
i18n/           — 后端国际化 (go-i18n, en/zh)
oauth/          — OAuth 供应商实现
web/            — React 前端
```

**核心分层：** Router → Middleware → Controller → Service → Model

**relay 模块（核心亮点）：**
- `relay/channel/` 下每个供应商一个子包（openai、claude、gemini、aws、ali、baidu 等 30+ 个）
- 统一 `Adaptor` 接口，定义了完整的请求生命周期：
  - `Init` → `GetRequestURL` → `SetupRequestHeader` → `ConvertXxxRequest` → `DoRequest` → `DoResponse`
- `TaskAdaptor` 接口处理异步任务（Midjourney、视频生成等），包含完整的计费生命周期：
  - `EstimateBilling` → `AdjustBillingOnSubmit` → `AdjustBillingOnComplete`
- 支持 OpenAI Chat/Responses、Claude Messages、Gemini 原生格式的互相转换

**中间件链：**
RequestId → PoweredBy → I18n → Logger → Session → CORS → Auth → RateLimit → ModelRateLimit → Distributor

### 3.2 Sub2API — 整洁架构 + 依赖注入

```
backend/
├── cmd/server/          — 入口 + Wire 依赖注入
├── ent/                 — Ent ORM schema 与生成代码
├── internal/
│   ├── config/          — 配置管理 (Viper)
│   ├── domain/          — 领域模型与常量
│   ├── handler/         — HTTP 处理器 (admin/ + 用户端)
│   │   └── dto/         — 数据传输对象
│   ├── payment/         — 支付集成
│   ├── pkg/             — 内部工具包 (apicompat, logger, openai)
│   ├── repository/      — 数据访问层
│   ├── server/          — HTTP 服务器 + 路由 + 中间件
│   │   ├── middleware/  — 中间件 (JWT/Admin/APIKey Auth, CORS, CSP, Recovery)
│   │   └── routes/      — 路由注册 (admin, auth, gateway, payment, user)
│   ├── service/         — 业务逻辑层（核心）
│   ├── setup/           — 首次运行向导
│   ├── util/            — 工具函数
│   └── web/             — 嵌入式前端
frontend/                — Vue 3 前端
```

**核心分层：** Router → Middleware → Handler → Service → Repository → Ent

---

## 四、核心技术要点

### 4.1 New API 技术要点

**1. 多供应商适配器体系**
- 通过 `Adaptor` 接口实现了 30+ AI 供应商的统一接入
- 每个适配器独立处理请求转换、响应解析、流式传输
- 支持 OpenAI ↔ Claude ↔ Gemini 三种格式的双向转换

**2. 智能渠道分发 (Distributor)**
- 加权随机选择算法，根据渠道权重分配流量
- Token 级别的模型访问控制
- 渠道自动测试与故障切换
- 支持指定渠道 ID 直连

**3. 多层级计费系统**
- 模型级别的倍率配置（输入/输出/缓存分别计价）
- 用户组差异化定价
- 异步任务的三阶段计费：预估 → 提交调整 → 完成结算
- 批量更新优化数据库写入

**4. 三数据库兼容**
- 所有 SQL 操作同时兼容 SQLite、MySQL、PostgreSQL
- 通过变量抽象处理保留字差异（`group`、`key` 列名引用）
- 布尔值、JSON 存储等跨库差异处理

**5. 前端嵌入与分析注入**
- `go:embed` 将前端构建产物嵌入二进制
- 运行时动态注入 Umami / Google Analytics 脚本
- 单二进制部署，无需额外 Web 服务器

**6. 性能优化**
- 内存缓存 + Redis 双层缓存
- 渠道缓存定时同步
- 配置热更新（无需重启）
- gopool 协程池管理后台任务
- Pyroscope 持续性能分析

### 4.2 Sub2API 技术要点

**1. Google Wire 编译期依赖注入**
- 所有服务通过 Wire ProviderSet 声明依赖
- 编译期生成依赖图，零运行时反射开销
- `wire.go` 清晰展示了完整的依赖拓扑
- 优雅关闭：并行清理应用层服务，顺序关闭基础设施（Redis → Ent）

**2. Ent ORM 类型安全数据访问**
- Schema 定义 → 代码生成，编译期保证查询正确性
- 自动生成 CRUD、关联查询、Where 条件
- 相比 GORM 的运行时反射，Ent 在类型安全和性能上更优

**3. 智能账号调度与粘性会话**
- 多上游账号池管理（OAuth 账号 + API Key 账号）
- 粘性会话（Sticky Session）：同一用户的请求路由到同一上游账号
- 基于负载因子的智能选择算法
- 单账号重试 + 跨账号智能重试

**4. TLS 指纹伪装**
- 使用 refraction-networking/utls 库
- 模拟真实浏览器的 TLS 握手指纹
- 降低被上游 AI 服务风控识别的概率

**5. 精细化计费**
- Token 级别的用量追踪
- 支持 Priority Service Tier 差异化定价
- 缓存创建/读取分别计价（5分钟/1小时缓存）
- 长上下文倍率（超过阈值后整次会话提价）
- 图片输出 Token 单独计价
- 多时间窗口限速（5h/1d/7d）

**6. 运维可观测性**
- OpsMetricsCollector：指标采集
- OpsAggregationService：数据聚合
- OpsAlertEvaluator：告警评估
- OpsScheduledReport：定时报告
- OpsSystemLogSink：系统日志汇聚
- 实时 WebSocket 推送运维数据

**7. 首次运行向导 (Setup Wizard)**
- 检测是否首次运行，自动启动配置向导
- 支持 CLI 模式和 Web UI 模式
- Docker 部署支持环境变量自动配置

**8. 安全设计**
- CSP 安全头 + 动态 frame-src 注入
- 请求体大小限制
- JWT + Admin 双层认证
- API Key 认证支持 Google 格式兼容
- 非 root 用户运行容器

---

## 五、优秀之处总结

### 5.1 New API 的优秀之处

| 亮点 | 说明 |
|------|------|
| 供应商覆盖广度 | 30+ AI 供应商适配器，几乎覆盖市面所有主流大模型 |
| 格式互转能力 | OpenAI/Claude/Gemini 三大格式双向转换，真正的"万能网关" |
| 三库兼容 | 同时支持 SQLite/MySQL/PostgreSQL，降低部署门槛 |
| 单二进制部署 | 前端嵌入 Go 二进制，一个 Docker 镜像搞定一切 |
| 异步任务计费 | 三阶段计费（预估→调整→结算）处理 Midjourney 等异步场景 |
| 多语言支持 | 前后端均支持国际化，覆盖中英法俄日越 6 种语言 |
| Electron 桌面端 | 提供桌面应用，适合个人用户本地使用 |
| 丰富的 OAuth | GitHub/Discord/Telegram/微信/LinuxDO/OIDC 全覆盖 |

### 5.2 Sub2API 的优秀之处

| 亮点 | 说明 |
|------|------|
| 工程规范性 | Wire 依赖注入 + Ent 类型安全 ORM + 整洁架构分层，工程质量高 |
| 测试覆盖 | 大量单元测试 + testcontainers 集成测试，质量保障完善 |
| 优雅关闭 | 并行清理 20+ 服务，顺序关闭基础设施，超时保护 |
| TLS 指纹伪装 | utls 库模拟浏览器指纹，降低风控风险 |
| 运维可观测性 | 完整的指标采集→聚合→告警→报告→日志链路 |
| 精细化计费 | 支持 Priority Tier、缓存分类、长上下文倍率、图片输出等细粒度计价 |
| 粘性会话调度 | 智能账号选择 + 会话保持，适合 Codex 等有状态场景 |
| Docker 安全 | 非 root 用户 + 健康检查 + pg_dump 版本一致性 |
| Setup Wizard | 首次运行自动引导配置，降低部署难度 |
| TypeScript 前端 | Vue 3 + TS + Vitest，前端代码质量有保障 |

### 5.3 值得互相借鉴的地方

| New API 可借鉴 Sub2API | Sub2API 可借鉴 New API |
|------------------------|------------------------|
| 引入 Wire 依赖注入，替代手动初始化 | 扩展更多 AI 供应商适配器 |
| 使用 Ent 替代 GORM，提升类型安全 | 支持多数据库（SQLite 降低入门门槛） |
| 增加测试覆盖率 | 增加 Electron 桌面端 |
| 结构化日志 (Zap) 替代自定义 logger | 更丰富的 OAuth 供应商支持 |
| 完善运维可观测性体系 | 格式互转能力（Claude ↔ Gemini） |

---

## 六、总结

**New API** 是一个"大而全"的 AI 网关，核心优势在于供应商覆盖广度和格式互转能力，适合需要统一管理多种 AI 服务的场景。架构偏传统，但胜在实用和部署简单。

**Sub2API** 是一个"精而深"的 API 分发平台，核心优势在于工程规范性和运维可观测性，适合需要精细化管理订阅配额、对稳定性和可维护性要求高的商业场景。架构现代，测试完善，但供应商覆盖面相对较窄。

两个项目代表了 Go 语言 AI 网关的两种典型路线：New API 走的是"快速迭代、功能覆盖"路线，Sub2API 走的是"工程质量、架构规范"路线。
