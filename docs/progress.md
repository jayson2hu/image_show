# 开发进度记录

## 2026-04-30 小尺寸生成兼容 GPT Image

- 问题：
  - 用户选择 `512x512` 生成时，上游返回 `sub2api status 502: error code: 502`。
  - 查阅 OpenAI 官方 Image API 文档后确认，GPT image models 当前生成尺寸只支持 `1024x1024`、`1024x1536`、`1536x1024` 和 `auto`；`512x512` 不能直接传给 GPT Image 上游。

- 决策：
  - 保留前端和业务层的小尺寸选项。
  - 后端调用上游时将小尺寸/非官方尺寸映射为 GPT Image 支持的最接近方向尺寸：方图 `1024x1024`、竖图 `1024x1536`、横图 `1536x1024`。
  - 上游返回图片后，后端再按用户选择的目标尺寸缩放并存储/返回。

- 完成：
  - 生成协程改为分离“用户请求尺寸”和“上游生成尺寸”。
  - `StoreGeneratedImage` 支持按目标尺寸缩放 base64 或 URL 图片结果。
  - 新增 `ProviderImageSize`、`ShouldResizeImage`、`ResizeImageBytes` 和尺寸解析工具。
  - 新增测试覆盖 `512x512`、竖图、横图上游尺寸映射和缩放后图片尺寸。

- 自测记录：
  - `go test ./service ./controller`：通过。
  - `go test ./...`：通过。
  - `pnpm.cmd build`：通过。沙箱内首次仍因 esbuild `spawn EPERM` 失败，授权后重跑通过。

## 2026-04-30 预览铺满和尺寸权限

- 问题：
  - 生成后的 1024 图片被结果区的两列网格和 `max-w` 限制，显示不够大。
  - “下载 / 下载全部”位置靠上且样式普通。
  - 尺寸选项缺少更小规格，且未登录用户不应开放大尺寸。

- 完成：
  - 首页结果预览改为铺满左侧预览区，图片使用整块可视区域展示。
  - 下载按钮移动到图片底部浮层工具条，并重新设计为深色玻璃按钮 + 白色主按钮。
  - 默认尺寸列表调整为 `512x512,768x768,1024x1024,1024x1536,1536x1024,1024x1792,1792x1024,1536x1536`。
  - `/api/generation/options` 改为可选登录识别：未登录只返回宽高都不超过 1024 的尺寸，登录后返回后台启用的完整尺寸。
  - 创建生成任务时后端同步拦截未登录用户使用大尺寸，避免绕过前端。
  - 后台设置页的尺寸配置说明补充未登录限制。

- 自测记录：
  - 新增测试覆盖未登录尺寸过滤、登录完整尺寸、未登录大尺寸生成拦截。
  - `go test ./...`：通过。
  - `pnpm.cmd build`：通过。沙箱内首次仍因 esbuild `spawn EPERM` 失败，授权后重跑通过。
  - 已重启本地后端，`http://localhost:3000/health` 返回 `{"status":"ok"}`。
  - 未登录访问 `GET /api/generation/options` 返回 `512x512,768x768,1024x1024`。

## 2026-04-30 Cloudflare R2 后台配置

- 问题：
  - R2 上传代码已存在，但配置只读取 `.env` / 环境变量，后台没有可编辑入口。
  - 未配置 R2 时，生成图片会以 data URL 兜底返回，不会上传到 Cloudflare R2。

- 完成：
  - 后台设置新增 `r2_endpoint`、`r2_access_key`、`r2_secret_key`、`r2_bucket`、`r2_public_url`。
  - R2 客户端改为优先读取后台设置，后台未填时回退环境变量 `R2_ENDPOINT`、`R2_ACCESS_KEY`、`R2_SECRET_KEY`、`R2_BUCKET`、`R2_PUBLIC_URL`。
  - 后台设置页为 R2 字段增加中文标签、密码输入和填写说明。
  - 部署文档补充后台 R2 配置说明。

- 自测记录：
  - 新增测试覆盖后台设置优先于环境变量。
  - 新增测试覆盖后台设置接口返回 R2 字段。

## 2026-04-30 生成失败与风格可选修复

- 问题：
  - 用户提交生成任务后，SSE 返回 `generation failed`，底层错误为 `Post "https://ai.laikankan8.top/v1/images/generations": unexpected EOF`，属于上游图片接口连接被提前断开。
  - 首页选择风格后会自动把风格 Prompt 追加到用户输入，用户已经自己写完整 Prompt 时无法选择“不追加风格”。
  - SSE 查看器中中文状态出现乱码，响应头未明确 UTF-8。

- 完成：
  - 首页风格预设新增“无”，并默认选中“无”；用户填入 Prompt 后可以不选择任何风格。
  - Sub2API 图片生成请求超时放宽到 300 秒，禁用 HTTP/2 和 keep-alive，并对 `unexpected EOF`、超时、连接重置、429/502/503/504 做最多 3 次短重试。
  - SSE 响应头改为 `text/event-stream; charset=utf-8`，并手写 `event:status` 帧，避免 Gin 覆盖 charset。
  - 如果订阅 SSE 时任务已经完成/失败/取消，发送初始状态后立即结束连接，不再持续 keepalive。

- 自测记录：
  - `go test ./...`：通过。
  - `pnpm.cmd build`：通过。沙箱内首次仍因 esbuild `spawn EPERM` 失败，授权后重跑通过。
  - 已重启本地后端，`http://localhost:3000/health` 返回 `{"status":"ok"}`。

## 2026-04-30 管理员登录修复

- 问题：
  - `admin@image-show.local / Admin123456` 无法登录时，现场检查发现 `http://localhost:3000` 后端未运行。
  - 代码层面也发现当前启动流程只会自动初始化默认套餐，不会自动初始化管理员；如果 SQLite 数据库被重建，管理员账号会丢失。

- 完成：
  - `config.Config` 增加 `ADMIN_EMAIL` 和 `ADMIN_PASSWORD` 配置。
  - 开发环境默认自动确保 `admin@image-show.local / Admin123456` 可用。
  - 生产环境不使用默认管理员，只有显式配置 `ADMIN_EMAIL` / `ADMIN_PASSWORD` 时才会创建或重置管理员。
  - `model.InitDB()` 启动时会幂等确保管理员账号为启用状态、`role=10`，并同步配置中的密码。
  - `.env.example` 和 `docs/deployment.md` 已补充管理员环境变量说明。

- 自测记录：
  - `go test ./...`：通过。
  - `pnpm.cmd build`：通过。沙箱内首次仍因 esbuild `spawn EPERM` 失败，授权后重跑通过。
  - 已启动后端并验证 `http://localhost:3000/health` 返回 `{"status":"ok"}`。
  - `POST /api/auth/login` 使用 `admin@image-show.local / Admin123456` 登录成功，返回用户 `role=10`。

## 2026-04-30 首页风格和推荐样例后台化

- 计划：
  1. 复用现有 `prompt_templates` 表和后台模板管理，不新增重复配置表。
  2. 使用 `category=style` 控制首页“风格预设”，使用 `category=sample` 控制首页“推荐样例”。
  3. 后台模板页增加“首页推荐样例”分类，并说明启用状态会影响前台展示。
  4. 首页从 `/api/prompt-templates` 动态加载风格和样例，接口异常时保留默认兜底。

- 完成：
  - `controller/prompt_template.go` 默认模板扩展为更完整的风格预设和推荐样例 Prompt。
  - `web/src/views/Home.vue` 移除固定写死的风格/样例展示，改为读取后台启用模板。
  - `web/src/views/admin/AdminDashboard.vue` 模板管理新增“首页推荐样例”分类，列表显示中文分类和启用状态。
  - `web/src/components/PromptTags.vue` 类型补充 `sample` 分类。
  - `controller/prompt_template_test.go` 新增默认模板分类测试。

- 自测记录：
  - `go test ./...`：通过。
  - `pnpm.cmd build`：通过。沙箱内首次仍因 esbuild `spawn EPERM` 失败，授权后重跑通过。
  - `http://localhost:3000/health`：返回 `{"status":"ok"}`。
  - `http://localhost:3000/api/prompt-templates`：返回 `style` 和 `sample` 两类默认模板。

## 2026-04-30 登录状态、导航和尺寸配置修复

- 计划：
  1. 检查管理员登录后的路由守卫，确认不再强制把管理员从普通页面跳回 `/admin`。
  2. 替换顶部品牌 `ArtifyAI` 及副标题为 `来看看巴`。
  3. 检查前端重复登录/注册入口，保留顶部唯一入口，并从顶部导航移除“套餐”。
  4. 增加可配置图片尺寸：后端新增 `enabled_image_sizes` 设置项，前端从 `/api/generation/options` 动态读取。
  5. 顶部登录态区分管理员和普通用户，分别显示“管理员已登录”和“普通用户已登录”。

- 完成：
  - `web/src/router/index.ts` 移除管理员访问 `/` 时自动跳 `/admin` 的逻辑，仅保留非管理员访问 `/admin` 的拦截，以及管理员登录页跳后台。
  - `web/src/App.vue` 顶部品牌改为 `来看看巴`，移除“套餐”入口，未登录时只显示一个“登录 / 注册”，登录后按角色显示状态。
  - `web/src/views/Home.vue` 删除页面内部额外登录/注册提示入口，高级参数和推荐样例继续默认折叠，尺寸下拉改为读取后端配置。
  - `controller/generation.go` 新增公开 options 接口，并在创建任务时校验尺寸必须属于启用列表。
  - `controller/admin_template_setting.go` 后台设置接口补充 `enabled_image_sizes`，管理员可在后台设置中维护逗号分隔的尺寸列表。
  - `router/main.go` 新增 `GET /api/generation/options`。

- 自测记录：
  - `go test ./...`：通过。
  - `pnpm.cmd build`：通过。沙箱内首次仍因 esbuild `spawn EPERM` 失败，授权后重跑通过。
  - `http://localhost:3000/health`：返回 `{"status":"ok"}`。
  - `http://localhost:5180/admin`：返回 200。

- 问题记录：
  - 首次验收时当前已运行的后端进程尚未加载新增的 `/api/generation/options` 路由，直接访问该接口出现重定向循环；已重启后端并复测通过，接口返回默认尺寸列表。

## 2026-04-30 Figma Make 首页界面

- 使用 Figma MCP 读取 `https://www.figma.com/make/5IFwPGoEStpt4u4DQHl2FZ/AI-Image-Generation-Website?t=e3ONudqa0VWtSlY1-1`。
- Figma Make 返回源码资源，核心布局来自 `src/app/App.tsx`：顶部导航、左侧 420px 参数面板、右侧生成预览区、紫蓝渐变主按钮、风格预设、推荐示例和高级参数滑块。
- 已在 `web/src/views/Home.vue` 实现对应布局，并保留现有后端生成接口、Turnstile 验证码、SSE 进度监听、取消、重试和用户积分刷新逻辑。
- 已修复首页相关组件中的中文显示问题：`App.vue`、`GenerationProgress.vue`、`PromptTags.vue`、`ImagePreview.vue`。
- 已补充 Figma 风格滑块样式到 `web/src/assets/main.css`。

## 自测记录

- `pnpm.cmd build`：通过。首次在沙箱内因 esbuild `spawn EPERM` 失败，已在授权后重新执行并通过。
- 后端健康检查：`http://localhost:3000/health` 返回 `{"status":"ok"}`。
- 前端开发服务：已启动在 `http://localhost:5180`，未占用用户说明的 `5173`。

## 问题与说明

- Figma MCP 的 `get_screenshot` 不支持 Figma Make 文件，因此本次依据 Make 源码资源分析布局实现。
- 当前后端生成接口仅支持单图任务，界面中的“图片数量 / 创造力 / 步数 / CFG Scale”为 Figma 对齐展示参数，暂未扩展后端协议。

## 2026-04-30 Figma Make 首页新版

- 重新使用 Figma MCP 读取同一 Make 链接，确认新版核心变化：品牌为 `ArtifyAI`，左侧为预览/结果区域，右侧为 420px 控制面板。
- 已将首页布局从“左控制、右预览”调整为“左预览、右控制”。
- 已按新版实现可折叠高级参数、可折叠推荐样例、实心紫色风格选中态、空状态文案和结果悬浮下载样式。
- 已同步全局顶部品牌区为 `ArtifyAI / AI 艺术创作平台`，并保留套餐、历史、管理、登录和退出入口。

## 自测记录

- `pnpm.cmd build`：通过。沙箱内仍因 esbuild `spawn EPERM` 失败，授权后重新执行通过。
- 前端访问检查：`http://localhost:5180` 返回 200。
- 后端健康检查：`http://localhost:3000/health` 返回 `{"status":"ok"}`。

## 2026-04-30 管理员入口调整

- 管理员登录后默认跳转 `/admin`，不再进入生图首页。
- 路由层增加限制：管理员访问 `/` 自动跳转 `/admin`，非管理员访问 `/admin` 会按登录状态跳转到首页或登录页。
- 顶部导航按角色切换：管理员只显示管理后台入口和退出，不展示套餐、历史等普通用户入口。

## 自测记录

- `pnpm.cmd build`：通过。沙箱内仍因 esbuild `spawn EPERM` 失败，授权后重新执行通过。
- 前端访问检查：`http://localhost:5180/admin` 返回 200。
- 后端健康检查：`http://localhost:3000/health` 返回 `{"status":"ok"}`。

## 2026-04-30 后台 UI 重设计

- 后台改为运维控制台布局：左侧分区导航，右侧工作区，去除管理员生图导向。
- 新增概览页，展示今日生成、新增用户、积分消耗和启用渠道等核心指标。
- 用户页重排为清晰表格，保留角色切换、封禁、充值和用户生成记录查看。
- 渠道页补齐新增、编辑、删除、测试连通性能力，可配置 Base URL、API Key、Headers、权重和状态。
- 模板、设置、积分、监控页统一为新的卡片/表格风格，并修复后台中文乱码。

## 自测记录

- `pnpm.cmd build`：通过。沙箱内仍因 esbuild `spawn EPERM` 失败，授权后重新执行通过。
- 前端访问检查：`http://localhost:5180/admin` 返回 200。
- 后端健康检查：`http://localhost:3000/health` 返回 `{"status":"ok"}`。

## 2026-04-30 Figma Make 登录新版

- 重新使用 Figma MCP 读取同一 Make 链接，确认新版主要变化在认证弹窗：新增“邮箱 / 微信”分段切换、微信扫码区域和“欢迎登录”标题。
- 已将真实登录页同步为新版分段登录布局。
- 邮箱登录保留现有 `/auth/login` 流程，微信登录保留现有 `/auth/wechat/qrcode` 和 `/auth/wechat/callback` 验证码流程。
- 保留上次验收要求：首页“高级参数”和“推荐样例”默认折叠。

## 自测记录

- `pnpm.cmd build`：通过。沙箱内仍因 esbuild `spawn EPERM` 失败，授权后重新执行通过。
- 前端访问检查：`http://localhost:5180` 返回 200。
- 后端健康检查：`http://localhost:3000/health` 返回 `{"status":"ok"}`。

## 2026-04-30 UI 问题修复

- 首页“高级参数”和“推荐样例”默认改为折叠，仅点击后展开。
- 登录页和注册页补齐深色模式下的卡片、文字、输入框、按钮和错误提示对比样式，避免深色模式切换后文字不可见。

## 自测记录

- `pnpm.cmd build`：通过。沙箱内仍因 esbuild `spawn EPERM` 失败，授权后重新执行通过。
- 前端访问检查：`http://localhost:5180` 返回 200。
- 后端健康检查：`http://localhost:3000/health` 返回 `{"status":"ok"}`。
