# Image Show 开发任务验收文档 v3.1

> 日期：2026-05-12 | 总任务：24 个 | 预估：7-9 天

## 全局决策（不再变更）

| 参数 | 决策 |
|------|------|
| 质量 | 固定 `medium`，后端硬编码 |
| 格式 | 固定 `PNG`，后端硬编码 |
| 背景 | 固定 `opaque`，后端硬编码 |
| 积分 | 方形=1，其余比例=2，后台可配置 |
| 游客/登录 | 完全一致，5种比例+积分预估 |

---

## P0 管理后台 D3 最终提交

**优先级**：P0 紧急 | **工作量**：< 0.5 天 | **依赖**：无

### 子任务

| ID | 内容 | 状态 |
|----|------|------|
| D3-1 | 确认路由入口已切换到新 `AdminLayout`，旧 `AdminDashboard.vue` 不再被引用 | 已完成 |
| D3-2 | `vue-tsc --noEmit` + `pnpm build` 零错误零警告 | 已完成 |
| D3-3 | 提交代码并推送 GitHub | 已完成 |
| D3-4 | 更新 `docs/plan-admin-dashboard-redesign.md` 进度表全部标记"已完成" | 已完成 |

### 自测标准

```bash
cd web && pnpm exec vue-tsc --noEmit
cd web && pnpm build
```

- 浏览器访问 `/console/admin`，7 个 Tab 均可正常切换

### 验收标准

- [x] 构建产物无错误无警告
- [x] `/console/admin` 所有 Tab 功能正常
- [x] 文档进度表全部标记"已完成"

### 提交信息

```
feat(admin): complete dashboard redesign D3 final verification

- verify vue-tsc and build pass with zero errors
- update plan-admin-dashboard-redesign.md to completed
```

---
## P1-A GPT Image 2 集成（精简版）

**优先级**：P1 核心 | **工作量**：3-4 天 | **依赖**：P0 完成后

### 积分定价

| 比例 | 显示名 | 像素 | 默认积分 | 后台设置键 |
|------|--------|------|----------|------------|
| 1:1 | 方形 | 1024×1024 | 1 | `credit_cost_square` |
| 3:4 | 竖版 | 1152×1536 | 2 | `credit_cost_portrait` |
| 9:16 | 故事版 | 1008×1792 | 2 | `credit_cost_story` |
| 4:3 | 横版 | 1536×1152 | 2 | `credit_cost_landscape` |
| 16:9 | 宽屏 | 1792×1008 | 2 | `credit_cost_widescreen` |

---

### 阶段一：后端改造

| ID | 内容 | 工作量 | 状态 |
|----|------|--------|------|
| G-1 | `model/models.go` — Generation 新增 `OutputFormat`、`Background` 字段；Quality 始终存 "medium"；GORM AutoMigrate | 小 | 已完成 |
| G-2 | `controller/admin_template_setting.go` 新增 5 个积分设置键；新建 `service/credit.go` — `CostForRatio(ratio string) int` 从 DB 读取，未配置用默认值 | 小 | 已完成 |
| G-3 | `controller/generation.go` — 接收 `prompt + ratio`，硬编码 `quality=medium / format=png / background=opaque`，调用 `CostForRatio` 扣费 | 小 | 已完成 |
| G-4 | `service/generation.go` — 透传固定参数到 OpenAI Images API，存储 `output_format / background` 到 DB | 小 | 已完成 |
| G-5 | `controller/site.go` — `GET /api/site/config` 扩展返回 `credit_costs: { square, portrait, story, landscape, widescreen }` | 小 | 已完成 |

#### 阶段一自测

```bash
go build ./...
go test ./controller/... -v -run TestGeneration
go test ./service/... -v
```

- POST `/api/generation` 只传 `prompt + ratio` → 成功生成，API 请求含 `quality=medium, format=png, background=opaque`
- 积分扣费：方形扣 1、竖版扣 2
- GET `/api/site/config` 返回 `credit_costs`
- 后台修改 `credit_cost_square=3` → 扣费变为 3

---

### 阶段二：前端积分预估 + 后台配置 UI

| ID | 内容 | 工作量 | 状态 |
|----|------|--------|------|
| G-6 | `web/src/views/Home.vue` — onMounted 读取 `/api/site/config` 的 `credit_costs`；比例选择器旁显示积分；切换比例时实时变化；生成按钮显示预估积分 | 小 | 已完成 |
| G-7 | `web/src/components/admin/SettingsTab.vue` — 「账号与额度」新增「生成积分定价」区域，5 个输入框对应 5 种比例，旁边显示像素尺寸，校验正整数 ≥ 1 | 小 | 已完成 |
| G-8 | 构建验证 + 集成测试：`vue-tsc + pnpm build + go test`；端到端生成不同比例图片，确认积分扣费正确、图片为 PNG | 小 | 已完成 |

#### 阶段二自测

```bash
cd web && pnpm exec vue-tsc --noEmit && pnpm build
```

- 首页切换比例，积分预估实时变化（方形 1、其余 2）
- 后台修改积分价格 → 前端刷新后生效
- 实际扣费与预估一致

### P1-A 整体验收标准

- [x] 前端 5 种比例选择正常，无质量/格式/透明选项
- [x] 切换比例时积分预估实时变化（方形=1，其余=2）
- [x] 实际扣费与预估一致
- [x] 管理员可在后台修改每种比例积分价格
- [x] 生成的图片全部为 PNG 格式
- [x] 所有用户（游客/登录）界面完全一致
- [x] `go test` 和 `vue-tsc / pnpm build` 全部通过

### 提交信息

```
feat(generation): configurable per-ratio credits and fixed quality/format

- extend Generation model with OutputFormat and Background fields
- add per-ratio credit cost settings (admin configurable)
- hardcode quality=medium, format=png, background=opaque
- expose credit costs in /api/site/config for frontend
- default: square=1 credit, others=2 credits
```

```
feat(home,admin): credit estimation display and admin pricing UI

- Home.vue reads credit costs from /api/site/config
- real-time credit preview when switching aspect ratios
- admin settings: per-ratio credit pricing with pixel reference
- integration test: all ratios generate correctly with proper deduction
```

---
## P1-B 登录注册重设计

**优先级**：P1 核心 | **工作量**：2-3 天 | **后端改动**：无（复用现有 API）

### 交互设计规范

**页面结构**：
- 顶部品牌区：渐变背景 + "Image Show" 标题 + 副标题
- 主卡片：微信二维码 + 验证码输入框 + 登录按钮（始终可见）
- 底部折叠区："已有邮箱账号？邮箱登录 ▾" 文字链 → 点击展开
- 展开后：Tab 切换「邮箱登录」(email+password) | 「邮箱注册」(email+password+验证码)

**交互流程**：
1. 页面加载 → 自动调 `GET /api/auth/wechat/qrcode` 获取二维码，加载中显示 skeleton
2. 微信 API 返回 `enabled=false` 或网络错误 → 隐藏二维码，自动展开邮箱，顶部提示"微信登录暂不可用"
3. 登录成功 → 跳转首页，`fetchUser` 更新状态

### 子任务

| ID | 内容 | 工作量 | 状态 |
|----|------|--------|------|
| L-1 | `web/src/views/Login.vue` — 重写：顶部品牌区 + 主卡片（微信二维码+验证码+登录按钮）+ 底部折叠区；onMounted 自动请求二维码 | 中 | 已完成 |
| L-2 | 邮箱折叠区：Tab 切换「邮箱登录」(email+password) / 「邮箱注册」(email+password+邮箱验证码+发送按钮)；折叠动画 max-height transition 300ms ease | 中 | 已完成 |
| L-3 | 微信降级处理：API 返回 `enabled=false` 或请求失败 → 隐藏二维码 → 自动展开邮箱 → 顶部提示条 | 小 | 已完成 |
| L-4 | `router/index.ts`：`/register` → `redirect: '/login'`；`App.vue` 导航"登录/注册"→"登录"；`Register.vue` 清空内容仅做 redirect 兜底 | 小 | 已完成 |
| L-5 | 样式与动效打磨：全站一致圆角(12/16px)、渐变按钮、折叠展开动画、二维码 skeleton、input focus 紫色边框、按钮 hover/active | 小 | 已完成 |

### 自测标准

```bash
cd web && pnpm exec vue-tsc --noEmit && pnpm build
```

- `/login` 自动加载微信二维码
- 验证码输入 + 登录可用
- 点击"邮箱登录"展开折叠区，Tab 切换正常
- 邮箱登录/注册均可成功
- `/register` 自动跳转到 `/login`
- 导航显示"登录"
- 微信关闭时邮箱自动展开

### 验收标准

- [x] `/login` 自动加载微信二维码，验证码登录可用
- [x] 邮箱折叠区含登录和注册 Tab，管理员可通过邮箱入口正常登录
- [x] `/register` 重定向 `/login`
- [x] 微信不可用时自动降级到邮箱，顶部提示条显示

### 提交信息

```
feat(auth): redesign login page with WeChat-first layout

- WeChat QR code auto-loads on page open
- email login/register as collapsible secondary entry with tab switch
- graceful fallback when WeChat unavailable
- /register redirects to /login, nav shows "登录"
```

---
## P2-A 首页场景入口

**优先级**：P2 体验增强 | **工作量**：2-3 天 | **依赖**：P1-A 完成后效果最佳

### 场景数据

| 场景 | 图标 | 推荐比例 | 积分 | 提示词方向 |
|------|------|----------|------|------------|
| 小红书封面 | 📸 | 3:4 竖版 | 2 | 精致生活/美食/穿搭风格封面 |
| 商品展示图 | 🛒 | 1:1 方形 | 1 | 白底/场景化商品展示 |
| 社交头像 | 👤 | 1:1 方形 | 1 | 精致人物/动漫风格头像 |
| 海报设计 | 🎨 | 3:4 竖版 | 2 | 活动/促销/艺术创意海报 |
| 手机壁纸 | 📷 | 9:16 故事版 | 2 | 风景/抽象/治愈系壁纸 |
| 自由创作 | ✨ | 1:1 方形 | 1 | 不填充，用户自由输入 |

### 卡片交互规范

| 状态 | 样式 |
|------|------|
| 默认 | 白底 + 1px border(slate-200) + 12px 圆角 + 微阴影 |
| Hover | border → violet-300，scale(1.02)，阴影加深，图标 translateY(-2px) 微跳 |
| 选中 | 2px border violet-500 + 背景 violet-50 + 外发光 + 左上角小勾 |
| 点击 | 0.15s scale(0.97→1) 弹性回弹 |
| 再次点击已选中 | 取消选中，清空提示词，恢复默认比例 1:1 |
| transition | all 0.2s cubic-bezier(0.4, 0, 0.2, 1) |

**提示词填充动画**：点击 → 输入框聚焦 → 打字机效果逐字出现（40ms/字 + 光标闪烁）；已有内容先 0.15s fade out 清空再填入；「自由创作」不填充，仅切换 1:1 并聚焦输入框。

**响应式**：≥1024px 3列网格 / 768-1023px 2列网格 / <768px 横向滚动（scroll-snap-type: x mandatory，显示 2.3 张）

### 子任务

| ID | 内容 | 工作量 | 状态 |
|----|------|--------|------|
| S-1 | `model/models.go` — PromptTemplate 新增 `Category string`（"style"/"scene"）；`controller/generation.go` — `GET /api/generation/scenes` 返回 category=scene 模板列表（含 name/icon/prompt_template/recommended_ratio/description） | 中 | 已完成 |
| S-2 | 新建 `web/src/components/SceneCard.vue` — props: scene数据、selected状态；渲染图标+标题+描述+比例badge+积分；hover/selected/click 状态；emit: select 事件 | 中 | 已完成 |
| S-3 | `web/src/views/Home.vue` — 提示词输入框下方插入场景网格；点击打字机填充+切换比例+更新积分；再次点击取消；响应式布局；onMounted 调用 `/api/generation/scenes` | 中 | 已完成 |
| S-4 | `web/src/components/admin/TemplatesTab.vue` — 模板 Tab 支持 category 筛选（全部/风格/场景）；scene 类型额外显示 icon 和 recommended_ratio；新增/编辑弹窗加 category 选择 | 小 | 已完成 |

### 自测标准

```bash
go test ./... -v
cd web && pnpm exec vue-tsc --noEmit && pnpm build
```

- 首页显示 6 个场景卡片
- 点击"小红书封面" → 打字机填入提示词 + 切换 3:4 + 积分显示 2
- 再次点击 → 取消选中，提示词清空，恢复 1:1
- "自由创作" → 不填充提示词，切到 1:1，输入框聚焦
- Hover：边框变色 + 微缩放 + 图标微跳
- 手机端横滑正常，snap 定位准确
- 后台可管理 scene 类型模板

### 验收标准

- [x] 首页展示场景卡片网格，交互动效流畅
- [x] 点击填充提示词（打字机效果）+ 切换比例 + 积分联动
- [x] "自由创作"不填充提示词
- [x] 选中态、hover 态、点击反馈完整
- [x] 响应式：桌面 3 列 / 平板 2 列 / 手机横滑
- [x] 后台可管理场景模板

### 提交信息

```
feat(home): add scene entry cards with typewriter prompt fill

- 6 scene cards with icons, descriptions, ratio badges, credit display
- typewriter animation for prompt fill on click
- hover/selected/click states with smooth transitions
- responsive: 3-col desktop, 2-col tablet, scroll mobile
- admin template management supports scene category
```

---
## P2-B 历史"再次生成"

**优先级**：P2 体验增强 | **工作量**：0.5-1 天 | **后端改动**：无

### 子任务

| ID | 内容 | 工作量 | 状态 |
|----|------|--------|------|
| R-1 | `web/src/views/History.vue` — 每张卡片 hover 显示悬浮操作栏，含"再次生成"按钮（紫色渐变小按钮） | 小 | 已完成 |
| R-2 | 点击 → `router.push({ name: 'home', query: { prompt, ratio } })`（quality/format 已固定，无需传） | 小 | 已完成 |
| R-3 | `Home.vue` — onMounted 检查 `route.query` → 回填 prompt + ratio → 清除 URL query（replace 模式）→ Toast "已回填历史参数，确认后点击生成" → **不自动触发生成** | 小 | 已完成 |

### 自测标准

```bash
cd web && pnpm exec vue-tsc --noEmit && pnpm build
```

- 历史页 hover 卡片出现"再次生成"按钮
- 点击 → 跳转首页，提示词和比例已回填
- 未自动触发生成
- URL query 已清除

### 验收标准

- [x] 历史页每张卡片 hover 有"再次生成"按钮
- [x] 跳转首页后提示词和比例已回填
- [x] 不自动触发生成
- [x] Toast 提示"已回填参数"

### 提交信息

```
feat(history): add regenerate button to history cards

- hover overlay with regenerate button on history cards
- route to home with prompt and ratio query params
- Home.vue reads query and fills form without auto-generating
```

---

## 整体执行顺序

| 顺序 | 功能块 | 任务 ID | 任务数 | 预估天数 |
|------|--------|---------|--------|----------|
| 1 | P0 管理后台提交 | D3-1 ~ D3-4 | 4 | 0.5 |
| 2 | P1-A 后端改造 | G-1 ~ G-5 | 5 | 1.5-2 |
| 3 | P1-A 前端 + 后台 | G-6 ~ G-8 | 3 | 1 |
| 4 | P1-B 登录重设计 | L-1 ~ L-5 | 5 | 2-3 |
| 5 | P2-A 场景入口 | S-1 ~ S-4 | 4 | 2-3 |
| 6 | P2-B 再次生成 | R-1 ~ R-3 | 3 | 0.5-1 |
| **合计** | | | **24** | **7-9 天** |

## 每个任务执行流程

1. **开发** — 按任务描述中的文件路径和内容实现
2. **自测** — 执行对应阶段的自测命令，确保全部通过
3. **提交** — 按建议 commit message 提交 + push
4. **记录** — 更新 `docs/progress.md` 记录完成状态
