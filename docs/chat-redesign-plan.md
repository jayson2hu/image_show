# 聊天式重构实施计划（Chat Redesign Implementation Plan）

> 目标读者：Codex / 任何独立开发者
> 视觉参考：`mockups/chat-redesign.html`、`mockups/layered-preview.html`
> 文档约定：每个 **任务（T01-T32）= 一个 PR**，单 PR 工作量控制在半天到 1.5 天

---

## 0. 总览

### 0.1 改造目标
1. 新增 `/chat` 聊天式入口（**设为默认 `/`**），原选项式界面迁到 `/classic` 保留。
2. 引入"会话（conversation） + 消息（message）"概念，每条用户消息绑定一次图像生成，历史可追溯。
3. 支持**图生图任务**（用户附图 + 自然语言，典型场景：老照片修复 / 风格转换），由后端意图识别路由到不同管线。
4. 支持**分层生成**：一次输出 1 张合成图 + N 张图层 PNG，每层 +1 积分，仅供下载不可编辑。
5. 完善**会话管理**：搜索 / 重命名 / 分享 / 删除 / 导出全部。
6. 后台可控**推荐样例**优先级、徽标、显示上限。

### 0.2 核心约束（产品边界）
- ❌ 不提供"在已生成图片上做编辑/修补/换层"的工具能力。
- ✅ 每条用户消息 = 一次独立的新生成，前一张图作为视觉历史保留但不可改。
- ✅ 分层是生成时一次性产物，存储后只可下载、不可编辑。
- ✅ 图生图 = 用户附图作为输入参考，AI 生成一张新图；不算编辑。

### 0.3 架构决策
| 决策点 | 选定方案 | 备选与放弃理由 |
|---|---|---|
| 会话存储 | 单独 `conversations` + `messages` 表，`generations.message_id` 外键 | 不复用 generation 自身分组，因为一条 message 未来可能产出多个产物（合成图 + N 层） |
| 分层产物 | 单独 `generation_assets` 表（kind: composite / layer），同 generation_id 下多行 | 不在 generation 表加 `layers_json` 字段，避免大字段更新风险 |
| 意图识别 | 后端基于"是否有附图 + 关键词正则"做轻量分类，输出 `task_kind`：`text2img` / `img_restore` / `img_style_transfer` / `img2img_generic` | 不引入 ML 分类器（v1 不值得），关键词表后台可维护 |
| 分层管线 | v1 不实做真实分层，后端返回 mock 多层（同一张图 + 不同 mask 占位）；前端先把 UI 和契约打通 | 真实分层需接 SAM 2 / LaMa，工程量大，单开一条 spike 排期 |
| 生成进度推送 | 复用现有 `/api/generations/:id/stream` SSE | 不引入 WebSocket，SSE 满足单向推送即可 |
| 路由 | `/` → Chat，`/classic` → 原 Home，老的 `/history` 保留作为通用历史页 | 不删 Home，保留 fallback |

### 0.4 工作量估算
- 总计 **32 个任务**，分 6 个阶段。
- 累计工程量预估 **18–22 个工作日**（含联调与自测）。
- 强烈建议**按阶段合并主干**，每阶段结束后回归一次。

### 0.5 阶段速览
| 阶段 | 任务区间 | 一句话 | 完成后能干什么 |
|---|---|---|---|
| Phase 0 地基 | T01–T03 | DB schema / 路由切换 / API client 骨架 | 能跑通空壳页面 |
| Phase 1 Chat MVP | T04–T12 | 会话 + 消息 + 文本生图最小闭环 | 用户能在 chat 里发 prompt 出一张图 |
| Phase 2 Composer 对齐 | T13–T16 | 场景 / 样例 / 风格 / 比例 / 附图 / 积分预估 | 老版选项功能全部对齐 |
| Phase 3 会话管理 | T17–T21 | 重命名 / 删除 / 搜索 / 分享 / 导出 | 会话生命周期完整 |
| Phase 4 图生图 | T22–T23 | 意图识别 + 老照片修复对比卡 | 用户附图就能修复老照片 |
| Phase 5 分层 | T24–T29 | Assets 表 / 分层 chip / 下拉分层下载 / 海报场景自动开 | 分层生成全功能上线 |
| Phase 6 管理 + 收尾 | T30–T32 | 后台模板优先级 / 空状态 / 移动端 | 可对外宣发 |

---

## 1. 任务通用约定

### 1.1 每个任务必含
- **ID & 标题**
- **依赖任务**
- **范围（文件路径）**：列出所有要新建 / 修改的文件
- **DB / API 改动**：建表语句、接口签名
- **验收标准（DoD）**：可勾选的功能点
- **自测步骤**：开发者本地能跑通的具体操作
- **预估代码量**：行数级别

### 1.2 命名规范
- Go controller 文件：`controller/{domain}.go`（如 `conversation.go`、`message.go`）
- Vue 组件：`web/src/components/chat/{PascalCase}.vue`
- Pinia store：`web/src/stores/{camelCase}.ts`
- API client：`web/src/api/{kebab}.ts`

### 1.3 测试基线
- 后端：每个新 controller 必须配套 `*_test.go`，覆盖正常路径 + 鉴权失败 + 参数校验失败
- 前端：新组件不强制单测，但每个交互流程要在自测步骤里走通

### 1.4 代码风格
- 沿用项目现有风格（Go: gofmt; TS: Vite 自带 + vue-tsc 检查）
- 提交前必须 `make build`（如有）+ `cd web && npm run build` 全部通过

---

# Phase 0 · 地基（T01–T03）

## T01 · 数据库 schema 扩展

**依赖**：无（最先做）

**范围**
- 新建 / 修改：`model/models.go`
- 迁移逻辑：`model/main.go`（AutoMigrate 调用处）
- 测试：`model/main_test.go` 增加新表迁移用例

**DB 改动**

```go
// 新增表：会话
type Conversation struct {
    ID         int64     `gorm:"primaryKey;autoIncrement" json:"id"`
    UserID     int64     `gorm:"index;not null" json:"user_id"`
    Title      string    `gorm:"size:128" json:"title"`         // 默认取首条 prompt 前 30 字
    LastMsgAt  time.Time `gorm:"index" json:"last_msg_at"`     // 用于左栏排序
    MsgCount   int       `gorm:"default:0" json:"msg_count"`
    TotalCost  float64   `gorm:"type:numeric;default:0" json:"total_cost"`
    IsLayered  bool      `gorm:"default:false;index" json:"is_layered"` // 该会话是否启用过分层
    IsDeleted  bool      `gorm:"default:false;index" json:"is_deleted"`
    ShareToken string    `gorm:"size:64;index" json:"share_token"` // 公开分享 token，空=未分享
    CreatedAt  time.Time `gorm:"index" json:"created_at"`
    UpdatedAt  time.Time `json:"updated_at"`
}

// 新增表：消息（用户输入 + AI 回复在前端是两条 bubble，但后端只存用户那一条；AI 那一条 = generation）
type Message struct {
    ID             int64     `gorm:"primaryKey;autoIncrement" json:"id"`
    ConversationID int64     `gorm:"index;not null" json:"conversation_id"`
    UserID         int64     `gorm:"index" json:"user_id"`
    Prompt         string    `gorm:"type:text" json:"prompt"`
    AttachmentKey  string    `gorm:"size:256" json:"attachment_key"` // R2 key，附图时填
    AttachmentURL  string    `gorm:"size:512" json:"attachment_url"`
    TaskKind       string    `gorm:"size:32;default:text2img;index" json:"task_kind"`
    Size           string    `gorm:"size:16" json:"size"`
    Quality        string    `gorm:"size:16" json:"quality"`
    StyleID        string    `gorm:"size:64" json:"style_id"`
    SceneID        string    `gorm:"size:64" json:"scene_id"`
    Layered        bool      `gorm:"default:false" json:"layered"`
    LayerCount     int       `gorm:"default:0" json:"layer_count"`
    GenerationID   *int64    `gorm:"index" json:"generation_id"` // 关联生成任务
    CreatedAt      time.Time `gorm:"index" json:"created_at"`
}

// 新增表：生成产物（一个 generation 可有 1 张合成 + N 张层）
type GenerationAsset struct {
    ID           int64     `gorm:"primaryKey;autoIncrement" json:"id"`
    GenerationID int64     `gorm:"index;not null" json:"generation_id"`
    Kind         string    `gorm:"size:16;not null" json:"kind"`     // composite / layer
    LayerName    string    `gorm:"size:64" json:"layer_name"`        // 背景/主体/文案/光效 等
    LayerOrder   int       `gorm:"default:0" json:"layer_order"`     // 0 = composite, 1..N = z 序
    R2Key        string    `gorm:"size:256" json:"r2_key"`
    URL          string    `gorm:"size:512" json:"url"`
    BytesSize    int64     `gorm:"default:0" json:"bytes_size"`
    Width        int       `json:"width"`
    Height       int       `json:"height"`
    CreatedAt    time.Time `json:"created_at"`
}

// 修改 Generation：增加 message_id 外键（保留原有字段，向后兼容）
type Generation struct {
    // ... 已有字段
    MessageID *int64 `gorm:"index" json:"message_id"`  // 新增
    TaskKind  string `gorm:"size:32;default:text2img" json:"task_kind"` // 新增
}

// 修改 PromptTemplate：补 v1 需要的字段
type PromptTemplate struct {
    // ... 已有字段
    Priority  int    `gorm:"default:0;index" json:"priority"`   // 越小越靠前
    Badge     string `gorm:"size:16" json:"badge"`              // pinned / hot / new / ""
}

// 修改 Setting：新增 site config 项（不改表结构，应用层约定 key）
// key: chat.sample_max_display      value: 8
// key: chat.layered.cost_per_layer  value: 1
// key: chat.intent.keywords.restore value: 修复,复原,去划痕,高清化,模糊,老照片
```

**API 改动**：无（纯 schema 任务）

**验收标准**
- [ ] `go build ./...` 编译通过
- [ ] `go test ./model/...` 全部通过
- [ ] 启动服务后，PostgreSQL/MySQL 中新出现 3 张表（`conversations`、`messages`、`generation_assets`）
- [ ] 旧 `generations` 表多出 `message_id`、`task_kind` 两列且默认值正确
- [ ] 旧 `prompt_templates` 表多出 `priority`、`badge` 两列
- [ ] 现有 `generations` 数据未损坏（导出抽样对比）

**自测步骤**
1. `git stash` 保存当前未提交代码
2. 启动服务连接到本地测试库，记录现有 `generations` 行数 N
3. `git stash pop` 应用本任务代码，重启服务触发迁移
4. SQL 查询 `SELECT COUNT(*) FROM generations;` 仍为 N
5. SQL 查询 `\d conversations`（pg）或 `DESC conversations`（mysql）确认列结构
6. 手动 `INSERT INTO conversations (user_id, title) VALUES (1, 'test');` 应成功
7. 跑 `go test ./model/... -v` 全绿

**预估代码量**：~120 行（schema 定义 + 测试）

---

## T02 · 前端路由切换（/ ↔ /chat ↔ /classic）

**依赖**：无（与 T01 可并行）

**范围**
- 修改：`web/src/router/index.ts`
- 新建：`web/src/views/Chat.vue`（先放占位，仅一行 `<div>Chat 占位</div>`）
- 修改：所有现有跳转到 `name: 'home'` 的引用——直接全局搜索 `name: 'home'`、`router.push('/')`、`<router-link to="/">` 检查语义
- 修改：`web/src/App.vue` 如果导航栏写死了"首页"链接，需要同步

**API 改动**：无

**路由清单（改动后）**
```ts
{ path: '/',         name: 'chat',    component: () => import('@/views/Chat.vue') }  // 默认聊天式
{ path: '/classic',  name: 'classic', component: Home }                              // 原选项式
{ path: '/login',    ... }
{ path: '/account',  ... }
{ path: '/history',  ... }    // 保留作为通用历史页
{ path: '/credits',  ... }
{ path: '/packages', ... }
{ path: '/console/admin/...', ... }
```

**验收标准**
- [ ] 浏览器打开 `http://localhost:5173/` 看到 "Chat 占位"
- [ ] 浏览器打开 `http://localhost:5173/classic` 看到原 Home.vue 全功能界面
- [ ] `npm run build` 类型检查通过（vue-tsc 无报错）
- [ ] 顶部导航条（如果有）里有"经典版"入口
- [ ] 历史依赖 `name: 'home'` 的跳转改为 `name: 'classic'` 或 `name: 'chat'`，语义对齐

**自测步骤**
1. `cd web && npm run dev`
2. 浏览器访问 `/` → 显示占位
3. 浏览器访问 `/classic` → 显示原 Home，能正常输入 prompt 生成图
4. 浏览器访问 `/login` → 正常
5. 登录后访问 `/account` → 正常
6. `npm run build` 通过

**预估代码量**：~30 行

---

<!-- Phase 1 starts after T03 -->

## T03 · 前端 API 客户端骨架

**依赖**：无（可与 T01、T02 并行）

**范围**
- 新建：`web/src/api/conversation.ts`
- 新建：`web/src/api/message.ts`
- 新建：`web/src/api/asset.ts`
- 修改：`web/src/api/index.ts`（保持现有 axios 实例不变；本任务只添加类型）

**API 改动**：仅前端 TS 定义，后端不动（接口签名按 T04-T06 设计预先约定）

**契约草案**（先定义 TS 类型 + 函数签名，实现先抛 `not implemented`）

```ts
// conversation.ts
export interface Conversation {
  id: number
  title: string
  last_msg_at: string
  msg_count: number
  total_cost: number
  is_layered: boolean
  share_token: string
  created_at: string
}

export async function listConversations(params: { q?: string; cursor?: string; limit?: number }): Promise<{ items: Conversation[]; next_cursor: string }>
export async function getConversation(id: number): Promise<Conversation>
export async function createConversation(payload: { title?: string }): Promise<Conversation>
export async function renameConversation(id: number, title: string): Promise<Conversation>
export async function deleteConversation(id: number): Promise<void>
export async function shareConversation(id: number): Promise<{ share_url: string; share_token: string }>
export async function unshareConversation(id: number): Promise<void>
export async function exportConversation(id: number): Promise<Blob>  // 返回 ZIP

// message.ts
export interface Message {
  id: number
  conversation_id: number
  prompt: string
  attachment_url: string
  task_kind: 'text2img' | 'img_restore' | 'img_style_transfer' | 'img2img_generic'
  size: string
  layered: boolean
  layer_count: number
  generation_id: number | null
  created_at: string
}

export interface PostMessagePayload {
  conversation_id?: number     // 不传则创建新会话
  prompt: string
  attachment?: File            // 附图
  size: string
  quality?: string
  style_id?: string
  scene_id?: string
  layered?: boolean
  layer_count?: number         // 仅 layered=true 时使用
}

export async function listMessages(conversationId: number): Promise<{ items: Message[] }>
export async function postMessage(payload: PostMessagePayload): Promise<{ message: Message; generation_id: number }>

// asset.ts
export interface GenerationAsset {
  id: number
  generation_id: number
  kind: 'composite' | 'layer'
  layer_name: string
  layer_order: number
  url: string
  width: number
  height: number
  bytes_size: number
}

export async function listAssets(generationId: number): Promise<{ items: GenerationAsset[] }>
export async function downloadAsset(assetId: number): Promise<Blob>
export async function downloadZip(generationId: number): Promise<Blob>
```

**验收标准**
- [ ] 4 个 TS 文件创建完成，导出函数签名清晰
- [ ] 所有函数 body 抛 `throw new Error('NOT_IMPLEMENTED')`，让后续任务渐进实现
- [ ] `npm run build` 类型通过
- [ ] 在 Chat.vue 占位里能 `import { listConversations } from '@/api/conversation'` 不报错

**自测步骤**
1. `cd web && npm run build` 通过
2. 在 Chat.vue 加一行 `import { listConversations } from '@/api/conversation'`，build 通过
3. console 调用 `listConversations({})` 应抛 `NOT_IMPLEMENTED`

**预估代码量**：~150 行（纯类型 + 空壳函数）

---

# Phase 1 · Chat MVP（T04–T12）

> 目标：跑通 **"用户登录 → 进入 /  → 点新建 → 输入 prompt → 出图 → 历史保留可见"** 最小闭环。
> 暂不实现：附图、分层、场景 chip、风格预设、样例、搜索、分享。

## T04 · 后端：Conversation CRUD

**依赖**：T01

**范围**
- 新建：`controller/conversation.go`
- 新建：`controller/conversation_test.go`
- 修改：`router/main.go`（注册新路由）

**API**
```
GET    /api/conversations           列表（按 last_msg_at desc，支持 q/cursor/limit）
POST   /api/conversations           创建（body: {title?}）
GET    /api/conversations/:id       详情
PATCH  /api/conversations/:id       重命名（body: {title}）
DELETE /api/conversations/:id       软删（is_deleted=true）
```

所有接口要求 `middleware.AuthRequired()`；权限校验：`user_id` 必须等于当前登录用户。

**验收标准**
- [ ] 5 个接口全部上线，对应 HTTP 状态码：200 / 201 / 200 / 200 / 204
- [ ] 鉴权失败返回 401；越权（访问别人会话）返回 404（不暴露存在性）
- [ ] 列表支持 `?q=关键词` 模糊匹配 title（LIKE %xx%）
- [ ] 列表支持 `?limit=20&cursor=<last_id>` 游标分页
- [ ] 软删后列表查不到，但 SQL 中行还在
- [ ] `controller/conversation_test.go` 覆盖：列表、创建、重命名、删除、越权 401、不存在 404

**自测步骤**
```bash
# 启动服务并登录获得 token
TOKEN="...your token..."

# 创建
curl -X POST http://localhost:8080/api/conversations \
  -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
  -d '{"title":"测试会话"}' | jq

# 列表
curl http://localhost:8080/api/conversations -H "Authorization: Bearer $TOKEN" | jq

# 重命名
curl -X PATCH http://localhost:8080/api/conversations/1 \
  -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
  -d '{"title":"改名后"}' | jq

# 软删
curl -X DELETE http://localhost:8080/api/conversations/1 -H "Authorization: Bearer $TOKEN" -w "%{http_code}\n"

# 越权（用 token A 访问 token B 的会话）应返回 404
```

**预估代码量**：~250 行（含 test）

---

## T05 · 后端：Message 创建 + 列表

**依赖**：T01、T04，T01 已新增 `Generation.MessageID`

**范围**
- 新建：`controller/message.go`
- 新建：`controller/message_test.go`
- 修改：`router/main.go`
- **修改：`controller/generation.go`** —— `CreateGeneration` 函数支持接收 `message_id` 参数，存入 generation 表

**API**
```
GET  /api/conversations/:id/messages              列出某会话的所有消息（按 created_at asc）
POST /api/messages                                创建消息 + 触发生成（核心接口）
```

**POST /api/messages 行为**
- 入参（multipart/form-data）：
  - `conversation_id`（可选，没有则新建会话，标题取 prompt 前 30 字）
  - `prompt`（必填）
  - `attachment`（可选，File）
  - `size`、`quality`、`style_id`、`scene_id`、`layered`、`layer_count`
- 出参：`{ message: {...}, generation_id: 123 }`
- 内部流程：
  1. 鉴权 + 参数校验
  2. 若无 conversation_id，创建会话
  3. 若有附图，先上传到 R2（复用现有 `service.UploadToR2`）
  4. 创建 message 行（task_kind 暂时一律存 `text2img`，T22 再做分类）
  5. 复用现有 `controller.CreateGeneration` 内部逻辑触发生成，传入 message.id 写入 `generations.message_id`
  6. 更新 conversation：`last_msg_at = now()`、`msg_count++`、`total_cost += credits_cost`、`is_layered |= layered`
  7. 返回 message 和 generation_id

**验收标准**
- [ ] GET messages 返回数组，单条 message 包含其 generation_id（即可由前端 join 出 generation 详情）
- [ ] POST 不带 conversation_id 时自动建会话，返回的 message.conversation_id 指向新会话
- [ ] POST 带不存在的 conversation_id 返回 404
- [ ] POST 越权（用别人的 conversation_id）返回 404
- [ ] 生成失败时 message 仍存在，generation.status=failed，前端能通过 generation_id 查到错误
- [ ] 积分扣减逻辑沿用现有 `controller.CreateGeneration`，不重复扣
- [ ] `controller/message_test.go` 覆盖：创建（带/不带 conversation_id）、列表、附图上传、越权 404

**自测步骤**
```bash
# 创建消息 + 新建会话
curl -X POST http://localhost:8080/api/messages \
  -H "Authorization: Bearer $TOKEN" \
  -F 'prompt=一只在星空下的猫咪' -F 'size=square' -F 'quality=medium' | jq

# 复用上一步返回的 conversation_id 继续发
curl -X POST http://localhost:8080/api/messages \
  -H "Authorization: Bearer $TOKEN" \
  -F 'conversation_id=1' -F 'prompt=再来一张' -F 'size=square' | jq

# 列表
curl http://localhost:8080/api/conversations/1/messages -H "Authorization: Bearer $TOKEN" | jq
```

**预估代码量**：~350 行（含 test，含对 generation.go 的小幅改动）

---

## T06 · 前端：填充 API 客户端实现

**依赖**：T03、T04、T05

**范围**
- 修改：`web/src/api/conversation.ts`（替换 `NOT_IMPLEMENTED`）
- 修改：`web/src/api/message.ts`
- 修改：`web/src/api/asset.ts`（assets 接口先留空，T26 实现）

**验收标准**
- [ ] 所有函数走真实 HTTP 请求
- [ ] 401 沿用现有 axios 拦截器，跳登录
- [ ] `postMessage` 支持 `attachment: File`（用 FormData）
- [ ] `exportConversation` 返回 Blob 可触发浏览器下载

**自测步骤**
1. 在 Chat.vue 占位里临时加按钮 `<button @click="test">Test</button>`：
   ```ts
   async function test() {
     const list = await listConversations({})
     console.log(list)
     const conv = await createConversation({ title: '手动测试' })
     console.log(conv)
     await deleteConversation(conv.id)
   }
   ```
2. 浏览器登录后点击按钮，看 Network 面板请求成功，console 打印数据

**预估代码量**：~120 行

---

## T07 · 前端：Conversation Store（Pinia）

**依赖**：T06

**范围**
- 新建：`web/src/stores/conversation.ts`

**Store 形状**
```ts
interface ConversationState {
  list: Conversation[]
  currentId: number | null
  messages: Record<number, Message[]>  // conversationId -> messages
  loading: boolean
  searchQuery: string
}

actions:
- fetchList(q?: string)
- selectConversation(id: number)
- fetchMessages(conversationId: number)
- sendMessage(payload: PostMessagePayload)  // 调 postMessage，本地乐观插入
- renameConversation(id, title)
- deleteConversation(id)
- pollMessageGeneration(messageId)   // 轮询其 generation 状态，更新本地
```

**验收标准**
- [ ] Store 类型严格（无 any）
- [ ] 乐观更新：发消息时立刻在 messages 数组里追加用户气泡，等 generation 完成再更新对应 AI 回复
- [ ] selectConversation 切换时自动 fetchMessages（如未缓存）

**自测步骤**：在 Chat.vue 占位里 import store，console 调用各 action 验证。

**预估代码量**：~180 行

---

## T08 · 前端：Chat 主布局骨架

**依赖**：T02、T07

**范围**
- 修改：`web/src/views/Chat.vue`（删除占位，落实骨架）
- 新建：`web/src/components/chat/ChatHeader.vue`（顶部会话标题栏）

**布局**（参考 `mockups/chat-redesign.html`）
```
.chat-shell  (h-screen flex)
  └─ <SessionList /> (T09)         width: 260px
  └─ <main>                          flex-1 flex flex-col
       ├─ <ChatHeader />             height: 56px
       ├─ <div .chat-stream>         flex-1 overflow-y-auto
       │    ├─ <MessageBubble />     (T10)
       │    ├─ <ImageReply />        (T10)
       │    └─ ... loop messages
       └─ <Composer />               (T11)
```

**ChatHeader 内容**
- 当前会话标题（点击进入重命名内联编辑）
- 元信息：msg_count、total_cost、相对时间
- 右上：`切换经典版` 链接（href=/classic）

**验收标准**
- [ ] 三栏布局响应式：lg 显示三栏，md 隐藏左栏（用抽屉），sm 全屏
- [ ] 空会话时显示居中 hero "输入一句话，开始创作"
- [ ] 没有任何 JS 报错，无未使用变量

**自测步骤**
1. 打开 `/` 看到空骨架
2. 缩窗到 768px 以下，左栏应收起为抽屉

**预估代码量**：~250 行

---

## T09 · 前端：SessionList 左栏

**依赖**：T07、T08

**范围**
- 新建：`web/src/components/chat/SessionList.vue`
- 新建：`web/src/components/chat/SessionItem.vue`

**功能（v1 最小）**
- 顶部"新建创作"按钮 → 调 `createConversation` → 选中
- 搜索框（v1 仅 UI，绑定 store.searchQuery，触发 fetchList）
- 列表按时间分组（今天 / 昨天 / 本周 / 更早）
- 单项 hover 时显示 ⋯ kebab（先放空菜单，T17–T21 填）
- 当前选中条目 teal 高亮
- 底部用户卡（接 useUserStore）

**验收标准**
- [ ] 列表显示真实数据
- [ ] 搜索框输入 debounce 300ms 触发 fetchList
- [ ] 点击条目切换 currentId 并 fetchMessages
- [ ] "新建创作"按钮成功后立刻选中并清空消息流

**自测步骤**
1. 登录后访问 `/`
2. 点击"新建创作" → 左栏出现新条目并自动选中
3. 后端 SQL 插入 3 条测试会话，刷新 → 看到分组列表
4. 搜索框输入"猫"→ Network 看到 `?q=猫` 请求

**预估代码量**：~280 行

---

## T10 · 前端：MessageBubble + ImageReply（v1 无分层）

**依赖**：T07

**范围**
- 新建：`web/src/components/chat/MessageBubble.vue`（用户气泡）
- 新建：`web/src/components/chat/ImageReply.vue`（AI 回复卡）
- 新建：`web/src/composables/useGenerationPoll.ts`（轮询/SSE 单条 generation 状态）

**MessageBubble**
- 右对齐深色气泡 + 下方元信息（比例 / 分层 N 层 / 消耗 N 积分）
- 若 message.attachment_url 存在，气泡顶部嵌一张缩略图

**ImageReply** 三态：
1. `generation.status = pending/running` → 加载卡（带进度环、% 数字、阶段文字、取消按钮）
2. `generation.status = succeeded` → 大图 + 操作条：下载（v1 单图按钮，下拉留到 T27）+ 分享（v1 仅 toast 占位）+ 复制提示词
3. `generation.status = failed` → 红色错误卡 + "重试"按钮（发同样的 prompt）

**useGenerationPoll**：复用现有 `/api/generations/:id/stream` SSE，或轮询 `GET /api/generations/:id` 每 2s 一次直到 succeeded/failed。

**验收标准**
- [ ] 在已有 messages 的会话里能正确渲染历史消息
- [ ] 新发消息后立刻看到用户气泡，2 秒内 AI 回复卡出现并显示加载
- [ ] 生成完成后图片自动显示
- [ ] 下载按钮触发浏览器下载

**自测步骤**
1. 发消息 "一只猫"
2. 观察用户气泡立刻出现
3. 观察 AI 卡从 pending → succeeded
4. 点击下载，浏览器收到 .png 文件
5. 故意把后端 channel 关掉造成失败，AI 卡显示错误 + 重试按钮

**预估代码量**：~400 行

---

## T11 · 前端：Composer 输入区（v1 最小）

**依赖**：T07

**范围**
- 新建：`web/src/components/chat/Composer.vue`
- 新建：`web/src/components/chat/CreditEstimate.vue`（积分预估卡，独立组件方便后续 T16 扩展）

**v1 内容**
- 多行 textarea（Enter 发送，Shift+Enter 换行）
- 右下角发送按钮（coral 圆）
- 占位的 chip：风格 / 比例 / 质量 / 附图 / 分层（先做静态 UI，T13–T16 接逻辑）
- 积分预估：调 `service.CostForSize`（前端复刻一份或读 site config）
- 余额从 useUserStore 取

**验收标准**
- [ ] Enter 触发发送，输入框清空
- [ ] 发送中按钮 disabled（loading 状态）
- [ ] 余额不足时按钮显示"充值"并跳转 `/packages`
- [ ] textarea 自适应高度（min 2 行，max 8 行）

**自测步骤**
1. 在选中的会话里输入"测试"按 Enter
2. 输入框清空，消息出现在流里
3. 余额改成 0，看按钮变成"充值"

**预估代码量**：~280 行

---

## T12 · Phase 1 联调与回归

**依赖**：T01–T11

**范围**：无新代码，纯回归

**验收标准（端到端）**
- [ ] 游客访问 `/` 看到引导登录（或允许 1 次试用，复用现有逻辑）
- [ ] 登录用户访问 `/` 看到空状态
- [ ] 点击"新建创作"→ 输入 prompt → 看到自己气泡 → 看到生成中卡 → 看到结果图
- [ ] 刷新页面，左栏会话仍在，点击进入能看到完整消息历史
- [ ] 切换不同会话能看到各自的消息流
- [ ] 访问 `/classic` 原 Home 功能完全不受影响
- [ ] `npm run build && go build ./...` 全绿

**自测步骤**
1. 走完 5 条端到端的"verify"清单
2. 录一段 30s 屏幕（可选，方便后续 demo）

**Phase 1 完成 = 可以内测**：核心闭环跑通，缺的是 composer chip 联动 + 高级功能。

---

# Phase 2 · Composer 功能对齐（T13–T16）

> 目标：把老版选项面板的 6 大能力（场景 / 风格 / 比例 / 质量 / 附图 / 推荐样例）全部迁移到 Composer 内嵌 chip。

## T13 · 场景 chip + 风格 / 比例 / 质量下拉

**依赖**：T11

**范围**
- 新建：`web/src/components/chat/SceneChips.vue`（横向 chip 行）
- 新建：`web/src/components/chat/StyleSelector.vue`（chip + popover）
- 新建：`web/src/components/chat/SizeSelector.vue`
- 新建：`web/src/components/chat/QualitySelector.vue`
- 修改：`web/src/components/chat/Composer.vue`（接入上述 4 个组件）

**数据源**
- 场景：`GET /api/generation/scenes`（已存在）
- 风格：`GET /api/prompt-templates`（已存在，filter category=style）
- 比例：`GET /api/generation/options`（已存在）
- 质量：硬编码 ['low', 'medium', 'high']

**交互**
- 点击场景 chip → 填充 textarea 为该场景的 prompt_template + 自动切到 recommended_ratio + 高亮该 chip
- 风格 chip 点击展开 popover，选中后 chip 显示风格名
- 比例 / 质量同理
- 选中场景再次点击 = 取消

**验收标准**
- [ ] 老版 Home.vue 里所有场景 / 风格 / 比例都能在 chat 里点出来
- [ ] 切换场景时 textarea 内容覆盖（保留用户已输入内容会引起混淆，v1 直接覆盖；可加 confirm，按交互设计稿不需要）
- [ ] 选中场景的 chip 视觉上 teal 高亮 + 末尾小徽标
- [ ] 切换比例后积分预估同步更新
- [ ] 移动端 chip 行可横向滚动

**自测步骤**
1. 点击"小红书封面"chip → textarea 自动填入小红书 prompt，比例切到 3:4
2. 切换风格"赛博朋克"，发送后查看 generation 记录的 prompt 包含风格后缀
3. 切换比例为 16:9，积分预估变 2
4. 移动端窗口看 chip 行滚动

**预估代码量**：~450 行

---

## T14 · 推荐样例 popover（含后台优先级）

**依赖**：T11、T13、T30（T30 是后台配置；T14 可以先用现有接口，T30 落地后无需改 T14）

**范围**
- 新建：`web/src/components/chat/SamplesPopover.vue`
- 修改：`controller/prompt_template.go` —— `PromptTemplates` 接口按 `priority ASC, sort_order ASC` 排序，并接入 site config 中的 `chat.sample_max_display` 限制返回数量
- 修改：`controller/prompt_template_test.go`

**API 改动**
- `GET /api/prompt-templates?category=sample` 多加可选参数 `limit`（默认读 site config，最大 50）
- 返回项额外字段：`priority`、`badge`

**SamplesPopover 内容**
- chip 点击 → 弹出气泡（参考 mockup），列出 max 8 条（数量后台控制）
- 每条带徽标（pinned / hot / new）
- 超过 max-height 时纵向滚动 + 底部 "向下滑动还有 N 条 ↓" 提示
- 点击某条 → 填充 textarea + 关闭气泡

**验收标准**
- [ ] 后台插入 10 条样例，前端只显示前 max_display 条（默认 8）
- [ ] 修改某条 priority=1 后刷新，该条排第一
- [ ] badge 字段 = "hot" 时显示 🔥 热门徽标
- [ ] 列表超过弹窗高度时出现滚动条 + 底部"向下滑动"提示
- [ ] 点击样例后输入框获得焦点，光标位于末尾

**自测步骤**
1. SQL 插入 12 条不同 priority 和 badge 的 sample
2. 设置 setting `chat.sample_max_display = 6`
3. 前端打开 popover 只看到 6 条
4. 滚动到底部看到"向下滑动..."
5. 点击置顶项 → 填充成功

**预估代码量**：~300 行

---

## T15 · 附图上传

**依赖**：T11

**范围**
- 新建：`web/src/components/chat/AttachmentChip.vue`（chip + 拖拽上传 + 缩略图预览）
- 修改：`web/src/components/chat/Composer.vue`（集成附图 chip + 在 textarea 上方显示已附图缩略卡）
- 修改：`web/src/api/message.ts`（postMessage 已支持 FormData，本任务只确认）

**约束**
- 单图上传，最大 10 MB
- 支持格式：png / jpg / jpeg / webp
- 客户端上传前做尺寸压缩（>4096px 缩放到 4096）—— 使用 canvas

**验收标准**
- [ ] 点击附图 chip 或拖拽到输入框 → 显示已附图缩略卡
- [ ] 缩略卡有 ✕ 移除按钮
- [ ] 超出大小 / 格式不符时显示 toast 拒绝
- [ ] 已附图状态下发送会走 multipart，后端 message.attachment_url 有值
- [ ] 用户气泡渲染时正确显示附图

**自测步骤**
1. 拖一张大图进输入框 → 缩略图出现
2. 拖一个 .txt → toast "格式不支持"
3. 拖一张 20MB 大图 → toast "大小超限"
4. 发送后查看 DB messages.attachment_key 不空

**预估代码量**：~250 行

---

## T16 · 积分预估 + 余额警示

**依赖**：T11、T13

**范围**
- 修改：`web/src/components/chat/CreditEstimate.vue`
- 新建：`web/src/composables/useCreditEstimate.ts`（纯函数：根据 size + layered + layer_count 算积分）
- 后端校验：`controller/message.go` 在 POST 前再算一次以防前端被改

**公式**
```
基础积分 = service.CostForSize(size)
分层附加 = layered ? layer_count * site_config['chat.layered.cost_per_layer'] : 0
总积分   = 基础 + 分层附加
```

**验收标准**
- [ ] composer 右下角实时显示 `本次消耗 X 积分 · 余 Y`
- [ ] 余额不足时数字变红，发送按钮 disabled，鼠标悬浮 tooltip "余额不足，去充值"
- [ ] 切换比例 / 切换分层开关都即时刷新预估
- [ ] 失败任务退积分逻辑沿用现有，前端 toast "已退回 X 积分"

**自测步骤**
1. 关分层 + square → 显示"2 积分"
2. 开分层 + 5 层 → 显示"7 积分"（2 + 5×1）
3. 余额改为 1 → 按钮 disabled + 红字
4. 故意让生成失败 → toast 显示退回

**预估代码量**：~150 行

---

# Phase 3 · 会话管理（T17–T21）

## T17 · 重命名

**依赖**：T04、T09

**范围**
- 修改：`web/src/components/chat/SessionItem.vue`（kebab 菜单加"重命名"）
- 修改：`web/src/components/chat/ChatHeader.vue`（标题点击进入内联编辑）
- 修改：`web/src/stores/conversation.ts`（renameConversation action）

**验收标准**
- [ ] 标题区域 hover 出现 ✎ 图标，点击 → 变 input + 保存/取消
- [ ] Enter 保存，Esc 取消
- [ ] 服务端 PATCH 成功后左栏标题同步更新（乐观更新 + 失败回滚）
- [ ] 名称为空或超过 128 字符拒绝

**自测步骤**
1. 点标题 → 变输入框
2. 输入新名 → Enter → 左栏对应条目标题更新
3. F12 看 Network PATCH 请求 200

**预估代码量**：~120 行

---

## T18 · 删除会话

**依赖**：T04、T09

**范围**
- 修改：`web/src/components/chat/SessionItem.vue`（kebab 加"删除会话"红色项）
- 新建/复用：`web/src/components/ui/ConfirmDialog.vue`（项目已有，复用）

**验收标准**
- [ ] 点击删除弹确认框 "确定删除「XXX」？此操作不可撤销"
- [ ] 确认后调用 DELETE，左栏移除，若是当前选中则切到第一条 or 空状态
- [ ] 后端真的把 is_deleted 置为 true，列表查不到，但 SQL 表中行还在

**自测步骤**
1. 删除当前会话 → 视图切到下一条
2. SQL `SELECT * FROM conversations WHERE is_deleted = true` 能看到刚删的

**预估代码量**：~80 行

---

## T19 · 会话搜索

**依赖**：T04、T09

**范围**
- 修改：`web/src/components/chat/SessionList.vue`（搜索框 + debounce + 高亮匹配）
- 后端：T04 已支持 `?q=`，无需改

**验收标准**
- [ ] 输入框 300ms debounce 触发查询
- [ ] 列表只显示匹配项
- [ ] 标题中匹配的文字高亮（用 `<mark>` 或 span）
- [ ] 搜索为空时恢复全列表
- [ ] 右上显示"找到 N 条" / "未找到匹配会话"

**自测步骤**
1. 创建 5 个不同标题的会话
2. 输入关键字 → 列表筛选
3. 清空 → 恢复

**预估代码量**：~100 行

---

## T20 · 会话分享（只读公开链接）

**依赖**：T04

**范围**
- 后端：`controller/conversation.go` 新增 2 个接口
  - `POST   /api/conversations/:id/share`   → 生成 share_token（UUID），返回 share_url
  - `DELETE /api/conversations/:id/share`   → 清空 share_token
  - `GET    /api/share/:token`              → **匿名可访问**的会话只读详情，返回 messages + assets
- 前端：
  - 新建：`web/src/views/SharedConversation.vue`（路由 `/share/:token`）
  - 修改：`SessionItem.vue` kebab 加"分享会话"
  - 新建：`web/src/components/chat/ShareDialog.vue`（弹出，显示 url + 一键复制 + 取消分享）

**SharedConversation 视图**
- 只读，无输入框
- 顶部"由 [用户名] 分享" + 复制链接按钮 + "在我的会话中再创作"CTA（跳 `/` 并把 prompts 复制到草稿）
- 不显示分享者积分等隐私信息

**验收标准**
- [ ] 分享后 share_url 形如 `https://yourdomain.com/share/abc123def`
- [ ] 匿名浏览器访问 share_url 能看到只读消息流
- [ ] 取消分享后链接立刻 404
- [ ] 分享接口需要登录且只能分享自己的会话
- [ ] 分享时若会话内 generation 还在运行，匿名访问能看到 placeholder 但不能查看实时进度

**自测步骤**
1. 分享一条会话 → 复制链接
2. 隐身窗口打开链接 → 看到完整对话流（无操作按钮）
3. 取消分享 → 隐身窗口刷新 404
4. 用别人 token 分享 → 401

**预估代码量**：~350 行

---

## T21 · 导出全部图片（ZIP）

**依赖**：T04

**范围**
- 后端：`controller/conversation.go` 新增 `GET /api/conversations/:id/export`
  - 流式打包该会话下所有 generation 的产物
  - 文件名 `{conversation_title}_{date}.zip`
  - ZIP 内目录结构：
    ```
    msg_001/
      composite.png
      layer_01_background.png   (若有分层)
      layer_02_subject.png
    msg_002/
      composite.png
    prompts.txt          (所有 prompt 流水)
    ```
- 前端：`SessionItem.vue` kebab 加"导出全部图片"

**验收标准**
- [ ] 触发后浏览器开始下载 ZIP，文件不损坏
- [ ] ZIP 解压后目录结构正确
- [ ] 大会话（100+ 张图）不会内存爆炸（使用 `archive/zip` 流式写入 Response）
- [ ] 鉴权失败 / 越权 404

**自测步骤**
1. 准备一个 5 条消息的会话（含分层）
2. 触发导出
3. 解压验证目录结构 + prompts.txt 完整

**预估代码量**：~200 行

---

# Phase 4 · 图生图任务（T22–T23）

> 目标：用户上传一张图（如老照片）+ 自然语言描述，AI 识别为相应任务并生成结果。

## T22 · 后端：意图分类器

**依赖**：T05、T15

**范围**
- 新建：`service/intent_classifier.go`
- 新建：`service/intent_classifier_test.go`
- 修改：`controller/message.go`（在创建 message 前调分类器，把结果写入 message.task_kind）

**分类器规则（v1：规则匹配，足够稳定）**
```go
func ClassifyTask(hasAttachment bool, prompt string) string {
    if !hasAttachment {
        return "text2img"
    }
    keywordMap := map[string][]string{
        "img_restore":         {"修复", "复原", "去划痕", "去模糊", "高清化", "老照片", "黑白上色"},
        "img_style_transfer":  {"风格", "改成", "变成", "迁移", "卡通", "动漫化", "油画"},
    }
    // 关键词来自 setting，方便后台维护
    for kind, keywords := range loadFromSiteConfig(keywordMap) {
        for _, kw := range keywords {
            if strings.Contains(prompt, kw) {
                return kind
            }
        }
    }
    return "img2img_generic"
}
```

**Setting 项（T01 已预留）**
- `chat.intent.keywords.restore`：默认 `修复,复原,去划痕,去模糊,高清化,老照片,黑白上色`
- `chat.intent.keywords.style`：默认 `风格,改成,变成,迁移,卡通,动漫化,油画`

**API 改动**：无新接口，仅 POST /api/messages 内部行为变更

**验收标准**
- [ ] 不带附图 → 100% 返回 text2img
- [ ] 带附图 + "修复一下老照片" → img_restore
- [ ] 带附图 + "改成动漫风格" → img_style_transfer
- [ ] 带附图 + "随便加点效果" → img2img_generic
- [ ] 关键词大小写、半全角混合也能命中
- [ ] 后台修改 setting 后无需重启即生效（关键词在 service 启动时加载一次 + DB 触发器或定期刷新）

**自测步骤**
```bash
# 不带附图
curl -X POST .../api/messages -F 'prompt=一只猫' ... | jq .message.task_kind   # text2img

# 带附图修复
curl -X POST .../api/messages -F 'attachment=@old_photo.jpg' -F 'prompt=帮我修复' ... | jq .message.task_kind   # img_restore
```

**预估代码量**：~200 行

---

## T23 · 前端：图生图对比卡（PhotoRestoreCard）

**依赖**：T10、T22

**范围**
- 新建：`web/src/components/chat/PhotoRestoreCard.vue`（左右对比卡，左原图右结果）
- 修改：`web/src/components/chat/ImageReply.vue`（根据 message.task_kind 路由到不同 sub-card）

**ImageReply 渲染逻辑**
```ts
const replyComponent = computed(() => {
  switch (message.task_kind) {
    case 'img_restore':         return PhotoRestoreCard
    case 'img_style_transfer':  return StyleTransferCard  // 留空，v1 fallback 到 GenericImageCard
    case 'img2img_generic':     return GenericImageCard
    default:                     return GenericImageCard
  }
})
```

**PhotoRestoreCard 内容**（参考 mockup turn 2）
- 左：原图（message.attachment_url）
- 右：修复中 / 修复结果（generation 状态决定）
- 4 步管线进度：分析 → 划痕去除 → 细节重建 → 色彩还原（v1 用伪进度，每 2s 推进一格直至 generation succeeded）
- 完成后底部：下载结果 + 下载原图 + 分享

**验收标准**
- [ ] 上传老照片 + "修复" 提示词 → AI 卡片是对比卡而非普通图卡
- [ ] 生成完成后右侧显示真实结果，可下载
- [ ] 切换到普通文生图任务时不显示对比卡

**自测步骤**
1. 上传一张老照片，prompt="修复老照片"
2. 看到 PhotoRestoreCard 渲染
3. 等待生成完成，下载结果
4. 同一个会话再发"一只猫"（无附图）→ 看到普通 ImageCard

**预估代码量**：~280 行

---

# Phase 5 · 分层生成（T24–T29）

> 目标：分层生成全功能上线。**前置约束**：v1 后端不接真实 SAM 2，先 mock 返回（同图 + 不同 mask 占位）；前端接口契约打通即可。**T25 同时安排一条 spike** 评估真实分层接入。

## T24 · 后端：Assets 表 + 单图入库改造

**依赖**：T01

**范围**
- 修改：`controller/generation.go` —— 生成成功后除了写 generations.r2_key，还要往 generation_assets 写一行（kind=composite, layer_order=0）
- 新建：`controller/asset.go`（仅 GET /api/generations/:id/assets，列出所有 asset）
- 修改：`router/main.go`

**API**
```
GET /api/generations/:id/assets    返回 [{ id, kind, layer_name, layer_order, url, ... }]
GET /api/assets/:id/download       单层下载（302 重定向到带签名的 R2 URL）
```

**验收标准**
- [ ] 老数据：现有 generations 都没有对应 asset 行（不补迁移），新生成的会有 1 条 composite
- [ ] GET assets 返回数组，对未分层的 generation 返回 1 条
- [ ] 单层下载 302 跳到签名 URL，10 分钟有效

**自测步骤**
1. 发一条新消息生成图
2. SQL `SELECT * FROM generation_assets WHERE generation_id = X` 看到 1 条
3. GET assets 返回该条
4. GET download 返回 302 + Location 头

**预估代码量**：~180 行

---

## T25 · 后端：分层生成管线（v1 mock）

**依赖**：T24

**范围**
- 新建：`service/layered_pipeline.go`
- 修改：`controller/generation.go` —— 当 message.layered = true 时走分层路径

**v1 行为（mock）**
- 调 gpt-image-2 出 1 张图
- 不调真实分割，假装拆分：把同一张图复制 N 份，分别命名 layer_01_background.png 到 layer_05_decoration.png
- 在每张层图上叠一个浅色 watermark `[Layer N]` 用于视觉区分
- 入库 generation_assets：1 行 composite + N 行 layer
- 异步处理：generation.status 先到 succeeded（composite 完成），assets 表里 layer 行用 `layer_status = pending` 标记，5–10 秒后异步任务填充
- v2 可替换为真实 SAM 2 / Replicate 等

**配套修改**
- Asset 表增加可选字段 `layer_status`（'pending' / 'ready' / 'failed'）
- GET assets 返回当前状态，前端按状态显示骨架 / 真图

**Spike 任务（独立排期）**
- 与 T25 并行：调研 SAM 2 / GroundingDINO 接入方案，输出独立文档 `docs/layered-spike.md`，不阻塞 T25 mock 上线

**验收标准**
- [ ] 发送 layered=true layer_count=5 的消息 → 生成 1 张合成 + 5 张层
- [ ] 5 张层在 generation_assets 表里 layer_order 1–5，layer_name 取自固定字典（背景 / 主体 / 文案 / 装饰 / 阴影 等）
- [ ] 异步任务完成前 layer_status=pending，前端能看到骨架
- [ ] 不开分层时只产出 composite，行为完全等同 T24

**自测步骤**
1. POST messages 带 layered=true layer_count=5
2. 立刻 GET assets → 1 composite + 5 个 pending layer
3. 等 10 秒再 GET → layers 都 ready
4. 下载某 layer，是带水印的图

**预估代码量**：~400 行

---

## T26 · 前端：LayerRailMini 组件

**依赖**：T10、T24、T25

**范围**
- 新建：`web/src/components/chat/LayerRailMini.vue`
- 修改：`web/src/components/chat/ImageReply.vue` 集成（右侧竖排缩略小轨）

**功能**
- 接收 `assets: GenerationAsset[]`
- 第一个槽位永远是 ⭐ 合成图
- 后续按 layer_order 排序展示
- 每槽 14×12 px 缩略，hover 显示 layer_name tooltip
- pending 状态显示 skeleton + ▶ coral 角标
- 当前选中槽位 outline-2 outline-teal
- 槽位点击 → 切换主画布预览（v1 仅 emit 事件，主画布切换在 T27 实现）

**验收标准**
- [ ] 未分层的 generation 不渲染分层小轨（无 layer 时隐藏）
- [ ] 分层完成的 generation 渲染 5 个槽位
- [ ] pending 的层显示 skeleton 动画
- [ ] 鼠标 hover 槽位有提示和 outline

**自测步骤**
1. 看分层生成结果 → 右侧出现 5 槽位
2. layer 全 ready 状态，hover 看 tooltip

**预估代码量**：~180 行

---

## T27 · 前端：下载下拉菜单（合成图 + 单层 + ZIP）

**依赖**：T26、T24

**范围**
- 新建：`web/src/components/chat/DownloadDropdown.vue`
- 修改：`web/src/components/chat/ImageReply.vue` —— 把 T10 的简单"下载"按钮替换为这个组件
- 后端：T20 已有的 export 接口可单独包一个 `GET /api/generations/:id/zip` 用于单条 generation 打 ZIP（不要复用会话级 export）

**下拉菜单内容**（参考 mockup）
- 合成图（推荐） · 单击直接下载
- 分隔线
- 单独下载图层（loop assets where kind='layer'）
- 分隔线
- 打包下载全部（ZIP）

**验收标准**
- [ ] 未分层任务 → 下拉里只有"合成图" + "打包下载（即合成图 + 原图）"
- [ ] 分层任务 → 显示完整列表
- [ ] 单层下载使用 T24 的 302 接口
- [ ] ZIP 下载使用 GET /api/generations/:id/zip
- [ ] 每项显示文件大小估算

**自测步骤**
1. 分层完成的 generation → 打开下拉
2. 点合成图 → 浏览器下载 composite.png
3. 点某层 → 下载该层 PNG
4. 点 ZIP → 下载完整包，解压验证

**预估代码量**：~220 行

---

## T28 · 前端：分层 chip + 层数选择 + 积分动态联动

**依赖**：T11、T16

**范围**
- 修改：`web/src/components/chat/Composer.vue` —— 分层 chip 改为可点击 toggle
- 新建：`web/src/components/chat/LayerCountSelector.vue`（弹出选择 3/4/5/自动）
- 修改：`web/src/composables/useCreditEstimate.ts` —— 分层时按层数加积分

**交互**
- 分层 chip 默认关（白底）
- 点击 → 弹出 popover "选择层数 [3] [4] [5] [自动]" + "每层 +1 积分" 提示
- 选定后 chip 变 coral 高亮显示 "分层 · 5 层"
- 再次点击 chip → 直接关闭分层（不弹 popover）
- 积分预估实时更新：`本次 2 积分 + 分层 5 = 7 积分`

**验收标准**
- [ ] 默认关闭状态，发消息时 layered=false 传给后端
- [ ] 开启后 chip 高亮，发消息时 layered=true, layer_count=N
- [ ] 切换层数实时改预估
- [ ] 余额不足时按钮 disabled

**自测步骤**
1. 关分层发消息 → DB messages.layered=false
2. 开 5 层发消息 → DB layered=true, layer_count=5, credits_cost=7

**预估代码量**：~250 行

---

## T29 · 海报场景自动启用分层 + 提示条

**依赖**：T13、T28

**范围**
- 修改：`controller/prompt_template.go` —— scene 模板增加可选字段 `default_layered`、`default_layer_count`
- 修改：`web/src/components/chat/SceneChips.vue` —— 点击场景时若带 default_layered=true 自动开启分层并显示蓝色提示条

**API 改动**
- `GET /api/generation/scenes` 响应每项增加 `default_layered: bool`、`default_layer_count: int`

**DB 改动**
- `PromptTemplate` 表加两列：
  ```go
  DefaultLayered    bool `gorm:"default:false" json:"default_layered"`
  DefaultLayerCount int  `gorm:"default:5"     json:"default_layer_count"`
  ```
- 默认数据里"海报设计"的 `default_layered=true, default_layer_count=5`

**验收标准**
- [ ] 数据库迁移后"海报设计"自动 default_layered=true
- [ ] 点击"海报设计"chip → 自动开分层 + 显示 teal 提示条 "已为你自动启用分层（5 层）…"
- [ ] 提示条有"关闭分层"按钮可一键关
- [ ] 切到其他场景时若分层是被场景自动开的则自动关；用户手动开的不影响

**自测步骤**
1. 点击海报设计 → composer 看到 teal 提示条 + 分层 chip 变高亮
2. 点击"关闭分层" → 提示条消失 + chip 复位
3. 切到小红书 → 不显示提示条
4. 手动开分层 → 切场景不会关掉

**预估代码量**：~180 行

---

# Phase 6 · 管理与收尾（T30–T32）

## T30 · 后台模板管理：优先级 + 徽标 + 显示上限

**依赖**：T01、T14

**范围**
- 修改：`web/src/components/admin/TemplatesTab.vue`（增加 priority / badge 字段编辑 + 显示上限设置）
- 后端：T01 已新增字段，但 `controller/admin_template_setting.go` 需要支持新字段读写
- 修改：`controller/admin_template_setting_test.go`

**TemplatesTab 改动**
- 列表每行增加 priority 输入框（支持拖拽排序更佳，v1 用数字输入）
- badge 字段下拉选项：无 / pinned / hot / new
- 顶部加一个"全局设置"区块，包含 `chat.sample_max_display` 输入框（保存到 setting）

**验收标准**
- [ ] 管理员能编辑 priority 和 badge 并保存
- [ ] 管理员能设置最大显示数量（1–50）
- [ ] 普通用户前端立刻看到效果（无需重启）

**自测步骤**
1. 管理员后台改某模板 priority=1
2. 前端 popover 看到该项排第一
3. 改全局 max_display=3 → popover 只显示 3 条

**预估代码量**：~220 行

---

## T31 · 空状态 + 引导

**依赖**：T08–T11

**范围**
- 修改：`web/src/views/Chat.vue` 当 currentId=null 或当前 conversation messages 为空时显示空状态
- 新建：`web/src/components/chat/ChatEmptyState.vue`

**空状态内容**
- 居中 hero："输入一句话，开始创作"
- 6 个场景大卡（不是小 chip，参考 mockup empty state）
- 推荐样例 4 条（不是 popover 而是直接铺开）
- 附图入口提示"也可以上传老照片做修复"

**验收标准**
- [ ] 新用户第一次访问看到空状态而不是空白
- [ ] 点场景大卡 → 选中场景 + 输入框获得焦点
- [ ] 点样例 → 填充输入框
- [ ] 发出第一条消息后空状态消失

**自测步骤**
1. 新建会话 → 看到空状态
2. 点"海报设计"大卡 → 输入框预填 + 高亮场景
3. 发一条消息 → 空状态消失

**预估代码量**：~250 行

---

## T32 · 移动端响应式 + 收尾

**依赖**：T01–T31 全部

**范围**
- 修改：`web/src/views/Chat.vue`、`SessionList.vue` 等所有 chat 组件
- 测试不同视口

**响应式断点**
- ≥1024px：三栏全开
- 768–1024px：左栏改抽屉（hamburger 触发）
- <768px：全屏 chat + 顶部 hamburger，所有 chip 行可横向滚动；composer 固定底部

**收尾清单**
- [ ] 所有页面 lighthouse 性能 ≥ 80
- [ ] 移动端 Chrome / Safari 实测 OK
- [ ] 暗黑模式（项目已有 darkMode: 'class'）chat 页面适配
- [ ] 国际化预留：所有硬编码中文文案集中在一个 `web/src/locales/zh-CN/chat.json`（v1 不做多语言，但留好出口）
- [ ] 文档：更新 README，补 `/chat` 入口说明
- [ ] 性能：分页 messages（>30 条懒加载），左栏会话超过 200 个自动虚拟滚动

**自测步骤**
1. Chrome devtools 模拟 iPhone 12 → 完整走一遍创建 / 发送 / 切会话
2. 切暗黑 → 看是否有未适配的白色硬编码
3. lighthouse 跑 `/` 看分数

**预估代码量**：~300 行（多为样式调整）

---

# 附录

## A · 接口契约速查表（汇总）

| Method | Path | 鉴权 | 引入任务 |
|---|---|---|---|
| GET | `/api/conversations` | 必须 | T04 |
| POST | `/api/conversations` | 必须 | T04 |
| GET | `/api/conversations/:id` | 必须 | T04 |
| PATCH | `/api/conversations/:id` | 必须 | T04 |
| DELETE | `/api/conversations/:id` | 必须 | T04 |
| GET | `/api/conversations/:id/messages` | 必须 | T05 |
| POST | `/api/messages` | 必须 | T05 |
| POST | `/api/conversations/:id/share` | 必须 | T20 |
| DELETE | `/api/conversations/:id/share` | 必须 | T20 |
| GET | `/api/share/:token` | 匿名 | T20 |
| GET | `/api/conversations/:id/export` | 必须 | T21 |
| GET | `/api/generations/:id/assets` | 必须 | T24 |
| GET | `/api/assets/:id/download` | 必须 | T24 |
| GET | `/api/generations/:id/zip` | 必须 | T27 |

## B · DB 变更速查

新表：`conversations` / `messages` / `generation_assets`
修改表：
- `generations` + `message_id`, `task_kind`
- `prompt_templates` + `priority`, `badge`, `default_layered`, `default_layer_count`

新增 setting key：
- `chat.sample_max_display`（默认 8）
- `chat.layered.cost_per_layer`（默认 1）
- `chat.intent.keywords.restore`（默认见 T22）
- `chat.intent.keywords.style`（默认见 T22）

## C · 风险与开放问题

1. **真实分层管线（SAM 2）**：T25 用 mock，真实接入是另一个 spike，预计 5–8 天工程量（含模型部署 / GPU 成本 / 边界 case 处理）。
2. **匿名用户能否使用 chat**：现有 Home 支持匿名 1 次免费试用，本计划默认 chat 要求登录。如果要支持匿名，需要 T04–T07 全部加 `OptionalAuth` 中间件，且匿名会话存到 localStorage 不入库。
3. **SSE 长连接稳定性**：T10 复用现有 stream，弱网下可能断连，前端需要 fallback 到轮询。
4. **图片版权**：分享会话给匿名用户看时，是否要给图片打水印？v1 不打，可在 T20 上线后视社区反应迭代。
5. **历史数据迁移**：现有 generations 不归属任何 conversation。可选迁移策略：① 不迁，旧数据只在 `/classic` 和 `/history` 看；② 按 user_id + 创建日期把同日的 generations 归到一个自动生成的会话。建议 ①，避免数据语义歧义。

## D · 上线节奏建议

- 阶段 1 完成 → **内测**，开放给团队 5 个人用 3 天
- 阶段 3 完成 → **灰度 10%**，引导用户访问 `/chat`
- 阶段 5 完成 → **公测**，默认 `/` 切到 chat（执行 T02 真正生效）
- 阶段 6 完成 → **GA**，关闭 `/classic` 入口的导航条显示（链接仍可访问）

## E · 给开发者的执行建议

1. **每个任务一条独立分支**：`feat/chat-T04-conversation-crud`
2. **PR 描述模板**：复制本文档对应任务的"验收标准 + 自测步骤"，附测试截图/录屏
3. **不要跨任务**：T04 不要顺手做 T05，避免回归面扩大
4. **依赖被阻塞时**：可以 mock 上游接口先做下游，但 mock 数据要标注 `// TODO(T0X): replace with real API`
5. **每完成一个 Phase**：跑一遍该 Phase 全部任务的自测步骤，作为阶段验收

---

**文档版本**：v1.0
**最后更新**：2026-05-15
**视觉参考**：`mockups/chat-redesign.html`、`mockups/layered-preview.html`

