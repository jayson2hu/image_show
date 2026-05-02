# 开发进度记录

## 2026-05-02 生成阶段离散显示修复

- 问题：
  - 用户继续反馈后端仍在“正在生成图片”时，前端视觉上仍像到了“保存图片”。
  - 原因是连续进度条百分比和阶段高亮表达仍可能让用户误判当前所在阶段。

- 决策：
  - 进度条改为按离散阶段节点显示：创建、生成、保存、完成。
  - 当前阶段必须独立显示，`status=1` 时只显示“当前阶段：生成”，保存节点不高亮。

- 完成：
  - 新增 `displayStage`，由当前 SSE status 单向派生。
  - 进度条只推进到当前阶段节点：生成阶段为 33.333%，保存阶段为 66.666%。
  - 当前节点增加 ring 高亮，并新增“当前阶段”文本。

- 自测记录：
  - `pnpm.cmd build`：通过。

## 2026-05-02 结果图全屏与参数面板收起

- 需求：
  - 图片生成完成后支持全屏查看。
  - 右侧 prompt/参数面板可以收起，提供更大的图片查看视野。

- 完成：
  - 生成完成后自动收起右侧参数面板。
  - 新增“展开参数”悬浮按钮，可随时恢复右侧面板。
  - 结果图操作区新增“全屏查看”按钮。
  - 新增全屏预览层，使用完整图片展示，并保留下载和关闭操作。

- 自测记录：
  - `pnpm.cmd build`：通过。

## 2026-05-02 生成进度条阶段位置修复

- 问题：
  - 后端状态仍是“正在生成图片”时，前端进度条视觉上已经接近或进入“保存”位置。
  - 原因是 `status=1` 的进度百分比设置为 56%，在四段阶段条中太靠近“保存”节点；同时后端状态文案在无 message 的事件下可能保留上一条。

- 决策：
  - 进度条百分比必须落在当前阶段范围内：生成阶段不能靠近保存节点。
  - “后端状态”文案必须和当前 status 同步，即使 SSE 事件没有 message 也要用当前 status 的默认文案覆盖。

- 完成：
  - `status=1` 进度从 56% 调整为 38%，明确停留在“生成”区间。
  - `status=2` 进度调整为 68%，只有后端真正进入保存状态才显示保存区间。
  - 后端状态文案改为每次 status 事件都同步刷新。

- 自测记录：
  - `pnpm.cmd build`：通过。

## 2026-05-02 进度条与后端状态一致性修复

- 问题：
  - 用户反馈进度条已经到“保存”，但后端状态仍显示“正在生成图片 / 模型正在处理提示词和画面内容”。
  - 原因是前端此前加入状态队列和最小展示时间后，可能出现视觉进度与最新后端状态不完全同步。

- 决策：
  - 前端进度条、阶段高亮、标题和“后端状态”必须严格来自同一个当前 SSE status。
  - 不再在前端缓存状态队列或延迟推进，避免视觉状态与后端状态脱节。

- 完成：
  - 移除 `GenerationProgress` 的状态队列和延迟展示逻辑。
  - 前端收到 SSE status 后立即同步当前阶段、进度条和后端状态。

- 自测记录：
  - `pnpm.cmd build`：通过。

## 2026-05-02 下载按钮无响应修复

- 问题：
  - 结果页“下载”和“下载全部”点击后没有反应。
  - 原因是直接使用 `<a download>` 依赖浏览器对当前 URL 的下载支持，跨域 R2/CDN 或签名 URL 场景下可能被浏览器忽略。

- 决策：
  - 前端改为主动 `fetch` 图片为 Blob，再用本地 object URL 触发下载。
  - 如果跨域策略阻止 fetch，则退回新窗口打开图片，让用户仍能保存。

- 完成：
  - 新增 `web/src/utils/download.ts` 通用下载工具。
  - 首页结果页两个下载按钮改为调用 Blob 下载。
  - 历史页下载也复用同一工具。

- 自测记录：
  - `pnpm.cmd build`：通过。

## 2026-05-02 生成状态跳过生成中修复

- 问题：
  - 点击生成后，前端状态可能从“任务已创建”直接跳到“正在保存结果”。
  - 原因是后端创建任务后立即启动生成 goroutine，`status=1` 可能在前端 SSE 建连前已经发布并写入完成；前端连上时读取到的初始状态已经是 `status=2`。

- 决策：
  - 后端创建任务后增加短暂启动缓冲，让前端有时间建立 SSE 连接。
  - 前端仍只展示真实收到的状态，但对非终态增加最小展示时间，避免真实事件连续到达时视觉上直接跳过。

- 完成：
  - 后端生成任务启动前增加 600ms 建连缓冲。
  - 前端 `GenerationProgress` 增加状态展示节流，非终态至少展示约 700ms，终态立即处理。
  - 前端状态处理改为队列，确保真实收到的 `0 -> 1 -> 2` 不会被后续事件覆盖跳过。

- 自测记录：
  - `go test ./service ./controller`：通过。
  - `go test ./...`：通过。
  - `pnpm.cmd build`：通过。

## 2026-05-02 sub2api 524 超时处理

- 问题：
  - 生成图片时返回 `sub2api status 524: error code: 524`。
  - `524` 是 Cloudflare 常见的上游源站响应超时，通常表示 sub2api 或其后面的模型服务生成耗时超过代理等待时间，不是前端参数错误。

- 决策：
  - 将 `524` 纳入 sub2api 可重试错误，和 `502/503/504/429` 一起最多重试 3 次。
  - 如果多次重试后仍失败，返回更明确的错误：上游生成超时，建议重试或切换渠道。

- 完成：
  - sub2api 客户端新增 `524` 重试判断。
  - sub2api 错误转换为更清晰的业务错误文案。
  - 新增测试覆盖 524 首次失败后重试成功、连续 524 后错误映射。

- 自测记录：
  - `go test ./service ./controller`：通过。
  - `go test ./...`：通过。

## 2026-05-01 结果图完整显示

- 问题：
  - 1024x1024 结果图在结果页使用 `object-cover` 铺满非方形容器，导致顶部内容被裁切，出现狗狗头部最上面显示不全。
  - 本地后台设置确认 `image_model=gpt-image-2`，模型配置不是本次裁切原因。

- 决策：
  - 结果页前景图片改为 `object-contain`，保证完整显示生成图。
  - 保留一层模糊放大的背景图铺满屏幕，避免页面变成空白留边，同时不裁切主体图片。

- 完成：
  - 首页生成结果展示改为“模糊背景铺满 + 前景完整图片”结构。

- 自测记录：
  - 后台设置接口确认当前 `image_model=gpt-image-2`。
  - `pnpm.cmd build`：通过。

## 2026-04-30 生成中状态与动态占位

- 需求：
  - 生成中不再轮播“正在设置场景”等假阶段文案。
  - 前端状态应根据后端 SSE 的真实 `status` 变化展示。
  - 生成中的图片占位区域需要有低调动态效果，鼠标经过/点击时有更明显的反馈。

- 完成：
  - 移除前端 `GenerationProgress` 的定时文案轮播。
  - 前端改为按 `status=0/1/2/3/4/5` 展示创建、生成、保存、完成、失败、取消状态。
  - 生成中占位图增加缓慢移动的点阵和波纹背景，鼠标移动显示局部高亮，点击触发脉冲反馈。
  - 后端生成状态推送文案改为中文，避免前端展示英文内部状态。

- 自测记录：
  - `go test ./service ./controller`：通过。
  - `go test ./...`：通过。
  - `pnpm.cmd build`：通过。

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

## 2026-04-30 sub2api 503 和生成模型配置

- 问题：
  - `768x768` 生成时，后端已将尺寸映射为 GPT Image 支持的上游尺寸，但 sub2api 返回 `503 Service Unavailable`。
  - 该错误属于上游渠道不可用/模型不可用/限流维护类错误，不是前端请求尺寸直接不支持的 `502` 问题。

- 决策：
  - 不写死生成模型名，后台增加 `image_model` 配置项。
  - 此处当时按旧文档误判默认模型为 `gpt-image-1`，已在后续“gpt-image-2 模型修正”中更正。

- 完成：
  - 新增环境变量 `IMAGE_MODEL`。
  - 生成请求模型名改为读取后台设置 `image_model`，没有后台设置时回退环境变量。
  - 后台设置页增加生成模型配置说明。

- 自测记录：
  - `go test ./service ./controller ./config`：通过。
  - `go test ./...`：通过。
  - `pnpm.cmd build`：通过。

## 2026-04-30 gpt-image-2 模型修正

- 问题：
  - 用户确认生成模型应为 `gpt-image-2`。
  - 重新核对官方文档后确认 `gpt-image-2` 是 OpenAI 官方 Image API 模型，上次记录“没有 gpt-image-2”是错误判断。

- 决策：
  - 默认生成模型改为 `gpt-image-2`。
  - 仍保留后台 `image_model` 配置，便于 sub2api 使用其他模型名。
  - `/v1/images/generations` 的 `size` 参数当前仍应保守使用 `1024x1024`、`1024x1536`、`1536x1024`，其他前端尺寸由后端映射后再缩放。

- 完成：
  - 更新 `IMAGE_MODEL` 默认值、后台设置默认值、前端说明、部署文档和测试断言。
  - 新增 `cmd/admin-reset` 运维命令，用于显式重置管理员账号密码。

- 本地处理：
  - 已执行 `go run ./cmd/admin-reset -email admin@image-show.local -password Admin123456`，本地 SQLite 管理员账号已重置为启用管理员。

- 自测记录：
  - `go test ./service ./controller ./config ./model`：通过。
  - `go test ./...`：通过。
  - `pnpm.cmd build`：通过。

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
