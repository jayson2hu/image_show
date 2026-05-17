# Chat 接后端 + 图生图全流程开发计划

> 任务编号前缀：**F01–F10**
> 视觉参考：`mockups/chat-img2img.html`
> 阅读须知：**本文档自包含**，无需阅读其它任何文档
> 单任务粒度：半天 ~ 1.5 天，可独立提 PR

---

## 0. 背景定位与非目标

### 0.1 这个计划要解决什么

现状：Chat 模块（路由 `/`）的前端 UI 已就位（SessionList、ChatHeader、ChatEmptyState、Composer、MessageList），但**整个聊天流是纯前端 mock**——会话用 `Date.now()` 当 id，消息只 push 到本地 state，**完全不调后端，也没有 AI 图片回复渲染**。

目标：把 Chat 真正接到后端生成管线，并支持"用户附图 + 自然语言 → AI 出新图"的图生图能力。完成后用户在 `/` 输入提示词（或附图）能真正生成图片，历史会话持久化到后端。

### 0.2 明确的非目标（不要做）

| 非目标 | 决策依据 |
|---|---|
| ❌ 不要后端意图分类器（不识别"修复 / 风格化 / 换背景"等关键词） | 让底层模型自己理解 prompt，前端无需路由不同管线 |
| ❌ 不要"原图 vs 结果"对比预览卡 | 用户上滚就能看到自己上传的原图，无需专门 UI |
| ❌ 不要按 task_kind 分支渲染（PhotoRestoreCard 等组件） | AI 回复一律走极简图片卡（GPT 风） |
| ❌ 不要为附图任务设计"分析 → 抠图 → 重建"等阶段进度 | 统一用百分比 + spinner 即可 |
| ❌ 不要"再画一张 / 编辑此图 / 作为参考"等按钮 | AI 生成结果**只能下载**，不能修改 |
| ❌ 不要质量选择器、不要自动分层提示条 | 全局已决议（dev-plan 3.1 / 3.5） |

**如果实现过程中觉得"加点 X 会更好"，先在 PR 评论里问，不要自由发挥。**

---

## 1. 当前代码现状（重要 · Codex 必读）

### 1.1 后端

**已存在：**
- `model/Conversation` 表 + CRUD 5 个接口（GET / POST / GET:id / PATCH / DELETE）
- `model/Generation` 完整生成流水线，含 SSE 流（`/api/generations/:id/stream`）
- `service.CallImageGeneration` 调 GPT Image 2，已支持 `source_image_url`（旧 Home.vue 留下的图片编辑能力）
- R2 上传（用户头像、生成产物都走它）

**不存在：**
- ❌ `model/Message` 表
- ❌ `controller/message.go`
- ❌ `POST /api/conversations/:id/messages` 路由
- ❌ Conversation 与 Generation 的关联字段

### 1.2 前端

**已存在：**
- `web/src/views/Chat.vue`（41 行，三栏布局骨架已搭好）
- `web/src/components/chat/SessionList.vue`（完整，含侧栏收起 / 重命名 / 删除）
- `web/src/components/chat/ChatHeader.vue`（22 行极简）
- `web/src/components/chat/ChatEmptyState.vue`（完整，含游客积分 / 场景按钮）
- `web/src/components/chat/Composer.vue`（291 行，含附图上传 UI ✅）
- `web/src/stores/composer.ts`（draft 含 `attachment: File | null`）
- `web/src/stores/conversation.ts`（**100% 本地 mock**，所有 action 都不调 API）
- `web/src/api/types.ts`（`Conversation` / `Message` / `Scene` 类型已定义）

**不存在 / 严重缺失：**
- ❌ `web/src/api/conversation.ts`（API 客户端）
- ❌ `web/src/api/message.ts`
- ❌ `MessageList.vue` 只有 **19 行**，只渲染 prompt 文字气泡，后跟静态 "正在准备生成任务..." 占位，**完全没有 AI 图片回复**
- ❌ 没有独立的 `MessageBubble.vue` 或 `ImageReply.vue` 组件
- ❌ `conversationStore.sendLocalMessage(prompt)` 只接 prompt 参数，**完全不传 attachment**
- ❌ 整个 conversation 状态用 `Date.now()` 做 id，刷新即丢

### 1.3 关键文件签名

```ts
// web/src/api/types.ts —— 已存在
export interface Conversation {
  id: number
  title: string
  msg_count: number
  last_msg_at: string
  is_layered: boolean
  total_cost: number
}

export interface Message {
  id: number
  conversation_id: number
  prompt: string
  task_kind: 'text2img' | 'img_restore' | 'img_style_transfer' | 'img2img_generic'
  created_at?: string
}
```

```ts
// web/src/stores/composer.ts —— 已存在
type Draft = {
  prompt: string
  size: string
  quality: string
  style_id: string
  scene_id: string
  layered: boolean
  layer_count: number
  attachment: File | null   // dev-plan 4 已完成
}
```

```ts
// web/src/stores/conversation.ts —— 已存在但全部是本地 mock
// 关键 action 现状：
sendLocalMessage(prompt: string) {
  // 只接 prompt，没有 attachment
  // 只 push 到本地 state，不调 API
  // 自动改名：首条消息后 title = prompt.slice(0,12) + '...'
}
```

### 1.4 视觉风格基线（GPT 极简风 · 强制对齐）

参考 `mockups/chat-img2img.html`，AI 回复必须用以下样式：

- **用户气泡**：`bg-slate-100 rounded-3xl` 浅灰胶囊，右对齐
- **用户附图任务**：缩略卡 + 文字胶囊**垂直堆叠**，两块独立浅灰元素
- **AI preamble**：`知画 · 已生成 X.Xs ›` 单行小灰字，可点击展开详情
- **AI 图片**：**无白卡外壳**，图直接 `rounded-2xl` 显示
- **下载按钮**：**浮在图片右下角**，`bg-black/55 rounded-full backdrop-blur`
- **下方小图标行**：复制提示词 / 分享 / 收藏 / 更多，4 个 `h-8 w-8` 灰色 icon，hover 时 `bg-slate-100`
- **AI 名字**：**知画**（取代 ImageShow）

---

## 2. 设计原则

1. **附图不是新功能，是消息的一个可选属性**：无论用户附不附图，都走同一条 `POST /api/conversations/:id/messages` 路径，后端通过有无 `attachment` 字段决定调文生图还是图生图。
2. **前端不感知 task_kind**：所有 AI 回复都走同一个 `ImageReply.vue` 组件，不按 task 分支。
3. **每条用户消息 = 一次独立生成**：不做"编辑上一张"的能力，前一张图作为视觉历史保留。
4. **失败任务自动退积分**：沿用现有 generation 失败处理逻辑。

---

## 3. 任务总览

| 编号 | 标题 | 依赖 | 工作量 |
|---|---|---|---|
| F01 | 后端：Message 模型 + 表迁移 | — | 0.5 天 |
| F02 | 后端：Message CRUD API（不含附图） | F01 | 1 天 |
| F03 | 后端：消息创建时触发 Generation | F02 | 1 天 |
| F04 | 后端：消息附图上传 + 图生图分支 | F03 | 1 天 |
| F05 | 前端：Conversation store 改调真实 API | — | 0.5 天 |
| F06 | 前端：sendMessage 调真实 API + 传附图 | F04, F05 | 0.5 天 |
| F07 | 前端：MessageList 渲染（GPT 极简风） | F06 | 1.5 天 |
| F08 | 前端：生成进度 SSE 对接 + 状态机 | F07 | 1 天 |
| F09 | 前端：用户气泡 + AI 回复样式细节打磨 | F07 | 0.5 天 |
| F10 | 联调与回归 | 全部 | 0.5 天 |
| **总计** | | | **8 天** |

> 后端 F01–F04 可与前端 F05 并行开工；F06 起必须串行。

---

## 4. 后端任务（F01–F04）

### F01 · 后端：Message 模型 + 表迁移

**涉及文件**
- `model/models.go`（新增 Message struct）
- `model/main.go`（AutoMigrate 列表加入 Message）
- `model/main_test.go`（迁移用例）

**Message 模型定义**

```go
type Message struct {
    ID             int64     `gorm:"primaryKey;autoIncrement" json:"id"`
    ConversationID int64     `gorm:"index;not null" json:"conversation_id"`
    UserID         int64     `gorm:"index" json:"user_id"`           // 0 = 匿名
    AnonymousID    string    `gorm:"size:128;index" json:"anonymous_id"`
    Prompt         string    `gorm:"type:text" json:"prompt"`
    AttachmentKey  string    `gorm:"size:256" json:"attachment_key"` // R2 key
    AttachmentURL  string    `gorm:"size:512" json:"attachment_url"`
    AttachmentName string    `gorm:"size:128" json:"attachment_name"`
    AttachmentSize int64     `json:"attachment_size"`
    TaskKind       string    `gorm:"size:32;default:text2img;index" json:"task_kind"`
    Size           string    `gorm:"size:16" json:"size"`
    StyleID        string    `gorm:"size:64" json:"style_id"`
    SceneID        string    `gorm:"size:64" json:"scene_id"`
    Layered        bool      `gorm:"default:false" json:"layered"`
    LayerCount     int       `gorm:"default:0" json:"layer_count"`
    GenerationID   *int64    `gorm:"index" json:"generation_id"`
    CreatedAt      time.Time `gorm:"index" json:"created_at"`
}
```

**Generation 表加 message_id 关联**（轻量改动）

```go
type Generation struct {
    // ... 已有字段
    MessageID *int64 `gorm:"index" json:"message_id"` // 新增
}
```

**验收标准**
- [ ] `go build ./...` 通过
- [ ] 启动服务后 `messages` 表自动创建
- [ ] `generations` 表多出 `message_id` 列，老数据不损坏
- [ ] `go test ./model/...` 通过

**自测**
1. 备份现有数据库
2. 启动服务，看日志确认 AutoMigrate 成功
3. SQL: `\d messages` / `DESC messages` 验证列结构
4. SQL: `INSERT INTO messages (conversation_id, user_id, prompt) VALUES (1, 1, 'test');` 成功
5. 现有 `generations` 行数对比改动前后一致

**工作量**：~80 行（含测试）

---

### F02 · 后端：Message CRUD API（不含附图）

**涉及文件**
- `controller/message.go`（新建）
- `controller/message_test.go`（新建）
- `router/main.go`（注册路由）

**新增路由**

```
GET    /api/conversations/:id/messages              列出某会话的消息（按 created_at asc）
POST   /api/conversations/:id/messages              创建消息（本任务先实现纯文本版）
```

**鉴权**：`middleware.OptionalAuth()`（允许匿名，匿名用户用 `X-Fingerprint` header）

**POST 入参**（JSON body，**本任务暂不实现 attachment**，F04 再补 multipart）

```json
{
  "prompt": "一只在星空下的猫咪",
  "size": "square",
  "style_id": "",
  "scene_id": "",
  "layered": false,
  "layer_count": 0
}
```

**POST 出参**

```json
{
  "message": { "id": 123, "conversation_id": 45, ... },
  "generation_id": null
}
```

> 注意：本任务 `generation_id` 始终返回 null，**F03 才把生成挂上去**。

**权限校验**
- 检查 `conversation.id` 的所有权：登录用户匹配 user_id，匿名用户匹配 anonymous_id
- 越权返回 404（不暴露存在性）

**验收标准**
- [ ] GET messages 返回 200 + 数组
- [ ] POST 成功返回 201 + message 对象 + null 的 generation_id
- [ ] POST 不带 prompt 返回 400
- [ ] POST 用别人 conversation_id 返回 404
- [ ] 匿名用户通过 `X-Fingerprint` header 能正常创建消息
- [ ] 测试用例覆盖：列表、创建、参数校验失败、越权 404

**自测**

```bash
TOKEN="..."

# 列表
curl http://localhost:8080/api/conversations/1/messages \
  -H "Authorization: Bearer $TOKEN" | jq

# 创建（纯文本）
curl -X POST http://localhost:8080/api/conversations/1/messages \
  -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
  -d '{"prompt":"一只猫","size":"square"}' | jq

# 匿名访问
curl -X POST http://localhost:8080/api/conversations/1/messages \
  -H "X-Fingerprint: test-fp-001" -H "Content-Type: application/json" \
  -d '{"prompt":"一只猫","size":"square"}' | jq
```

**工作量**：~220 行（含测试）

---

### F03 · 后端：消息创建时触发 Generation

**涉及文件**
- `controller/message.go`（修改：POST 入口加生成调度）
- `controller/generation.go`（轻改：暴露内部生成调度函数供 message 复用，或直接复制核心逻辑）
- `controller/message_test.go`（补测试）

**改动**

POST `/api/conversations/:id/messages` 内部流程改为：

```
1. 鉴权 + 参数校验（已有）
2. 创建 message 行（已有）
3. ★ 新增：调用 generation 内部调度
   ├─ 计算 credits_cost（按 size + layered）
   ├─ 扣积分（登录用户走 credits 表，匿名用户走 anonymous free 计数）
   ├─ 创建 generation 行：
   │   ├─ user_id = message.user_id
   │   ├─ prompt = message.prompt
   │   ├─ size = message.size
   │   ├─ message_id = message.id  ← 关键关联
   │   ├─ task_kind = message.task_kind（暂时全部为 text2img）
   │   └─ source_image_url = ""（F04 才填）
   ├─ 异步触发实际生成（沿用现有 worker）
   └─ 返回 generation_id
4. 更新 conversation：last_msg_at, msg_count++, total_cost += cost
5. 返回 { message, generation_id }
```

**关键决策**
- **不创建新 endpoint**，所有生成都从 message 入口走
- 现有的 `POST /api/generations` 保留给 Classic 版（`/classic` 路由用），不动
- 异步生成结果通过现有 `/api/generations/:id/stream` SSE 推送

**验收标准**
- [ ] POST messages 返回 `generation_id` 非 null
- [ ] DB 中 generation 的 `message_id` 字段正确指向新 message
- [ ] 积分扣减一次（不重复扣）
- [ ] 失败任务（如积分不足）返回 402，message 不创建
- [ ] generation 完成后能通过 `GET /api/generations/:id` 查到结果

**自测**

```bash
# 创建消息 + 触发生成
GEN_RESPONSE=$(curl -s -X POST http://localhost:8080/api/conversations/1/messages \
  -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
  -d '{"prompt":"一只在星空下的猫咪","size":"square"}')
GEN_ID=$(echo $GEN_RESPONSE | jq -r .generation_id)

# 轮询生成状态
sleep 5
curl http://localhost:8080/api/generations/$GEN_ID \
  -H "Authorization: Bearer $TOKEN" | jq

# 看 DB 关联
psql -c "SELECT m.id AS msg_id, g.id AS gen_id, g.status, g.image_url \
         FROM messages m LEFT JOIN generations g ON g.message_id = m.id \
         ORDER BY m.id DESC LIMIT 1;"
```

**工作量**：~180 行（含测试）

---

### F04 · 后端：消息附图上传 + 图生图分支

**涉及文件**
- `controller/message.go`（POST 改为 multipart/form-data）
- `service/generation.go`（在已有 `CallImageGeneration` 入口加分支：有 source_image_url 时走图生图）
- `controller/message_test.go`（补附图用例）

**改动**

POST `/api/conversations/:id/messages` 改为 `multipart/form-data`：

| Field | Type | Required |
|---|---|---|
| `prompt` | string | ✅ |
| `size` | string | ✅ |
| `style_id` | string | optional |
| `scene_id` | string | optional |
| `layered` | bool | optional |
| `layer_count` | int | optional |
| `attachment` | File | optional |

**附图处理流程**
1. 校验：format ∈ {jpg, png, webp}, size ≤ 10MB
2. 上传到 R2（复用现有 `service.UploadToR2`）
3. 写入 `messages.attachment_url`、`attachment_key`、`attachment_name`、`attachment_size`
4. 创建 generation 时 `source_image_url = message.attachment_url`

**`service.CallImageGeneration` 分支逻辑**

```go
func CallImageGeneration(ctx context.Context, gen *model.Generation) error {
    if gen.SourceImageURL != "" {
        return callImg2Img(ctx, gen)   // 调 OpenAI /v1/images/edits
    }
    return callText2Img(ctx, gen)      // 现有逻辑
}
```

**`callImg2Img` 内部**
- 下载 `source_image_url` 内容（或直接用 R2 签名 URL）
- POST 到 OpenAI `/v1/images/edits`，body 含原图 + prompt + size
- **不在 prompt 上加任何后缀**，不做关键词识别
- 返回生成的图片，存到 R2

**验收标准**
- [ ] 带 attachment 的 multipart 请求能成功创建 message，DB 中 `attachment_url` 非空
- [ ] 不带 attachment 走原文生图，行为同 F03
- [ ] 附图格式不合法 → 400 + 明确错误
- [ ] 附图 > 10MB → 400 + 明确错误
- [ ] 带附图 generation 的 `source_image_url` 字段非空
- [ ] 图生图 endpoint 真的被调用（在 channel 日志能看到）
- [ ] 测试用例覆盖：带图成功 / 不带图成功 / 超大文件 400 / 错误格式 400

**自测**

```bash
# 带附图
curl -X POST http://localhost:8080/api/conversations/1/messages \
  -H "Authorization: Bearer $TOKEN" \
  -F 'prompt=帮我修复这张老照片' \
  -F 'size=square' \
  -F 'attachment=@./test_old_photo.jpg' | jq

# 不带附图
curl -X POST http://localhost:8080/api/conversations/1/messages \
  -H "Authorization: Bearer $TOKEN" \
  -F 'prompt=一只猫' \
  -F 'size=square' | jq

# 错误格式
curl -X POST http://localhost:8080/api/conversations/1/messages \
  -H "Authorization: Bearer $TOKEN" \
  -F 'prompt=test' -F 'size=square' \
  -F 'attachment=@./readme.txt'
```

**工作量**：~150 行（含测试）

---

## 5. 前端任务（F05–F09）

### F05 · 前端：Conversation store 改调真实 API

**涉及文件**
- `web/src/api/conversation.ts`（新建）
- `web/src/stores/conversation.ts`（重构）

**新建 `web/src/api/conversation.ts`**

```ts
import api from './index'
import type { Conversation } from './types'

export async function listConversations(): Promise<{ items: Conversation[] }> {
  const res = await api.get('/conversations')
  return res.data
}

export async function createConversation(title?: string): Promise<Conversation> {
  const res = await api.post('/conversations', { title })
  return res.data
}

export async function renameConversation(id: number, title: string): Promise<Conversation> {
  const res = await api.patch(`/conversations/${id}`, { title })
  return res.data
}

export async function deleteConversation(id: number): Promise<void> {
  await api.delete(`/conversations/${id}`)
}
```

**重构 `stores/conversation.ts`**

把现有的 4 个 local action 改为调真实 API：
- `createLocalConversation` → `createConversation`（调 API，把返回的 conversation 加入 list）
- `updateConversationTitle` → `renameConversation`（调 PATCH）
- `deleteLocalConversation` → `deleteConversationAction`（调 DELETE）
- 新增 `fetchConversations`（onMounted 时调用，恢复历史会话）

**保留 local 行为**：
- `sendLocalMessage`（F06 才改造）
- `toggleSidebar`
- `selectConversation`

**验收标准**
- [ ] 刷新页面后会话列表从后端恢复（不再用 Date.now() id）
- [ ] 新建会话调 POST，返回真实 id（int64）
- [ ] 重命名调 PATCH，左栏实时更新
- [ ] 删除调 DELETE，列表移除
- [ ] 网络失败时 toast 错误，本地状态不被破坏
- [ ] 现有 SessionList.vue 完全无需改动（store 接口形状保持兼容）

**自测**
1. 登录后访问 `/`，看 Network 面板有 `GET /api/conversations` 请求
2. 点"新建创作"看 `POST /api/conversations`
3. 重命名看 `PATCH /api/conversations/X`
4. 删除看 `DELETE /api/conversations/X`
5. 刷新页面，会话列表持久化

**工作量**：~150 行

---

### F06 · 前端：sendMessage 调真实 API + 传附图

**涉及文件**
- `web/src/api/message.ts`（新建）
- `web/src/stores/conversation.ts`（改 `sendLocalMessage` → `sendMessage`）
- `web/src/components/chat/Composer.vue`（轻改：send() 不再 reset attachment 在调用前）

**新建 `web/src/api/message.ts`**

```ts
import api from './index'
import type { Message } from './types'

export interface SendMessagePayload {
  conversation_id: number
  prompt: string
  size: string
  style_id?: string
  scene_id?: string
  layered?: boolean
  layer_count?: number
  attachment?: File | null
}

export async function listMessages(conversationId: number): Promise<{ items: Message[] }> {
  const res = await api.get(`/conversations/${conversationId}/messages`)
  return res.data
}

export async function sendMessage(payload: SendMessagePayload): Promise<{ message: Message; generation_id: number }> {
  const formData = new FormData()
  formData.append('prompt', payload.prompt)
  formData.append('size', payload.size)
  if (payload.style_id) formData.append('style_id', payload.style_id)
  if (payload.scene_id) formData.append('scene_id', payload.scene_id)
  if (payload.layered) {
    formData.append('layered', 'true')
    formData.append('layer_count', String(payload.layer_count || 5))
  }
  if (payload.attachment) {
    formData.append('attachment', payload.attachment)
  }
  const res = await api.post(`/conversations/${payload.conversation_id}/messages`, formData)
  return res.data
}
```

**重构 `sendLocalMessage` → `sendMessage`**

```ts
async sendMessage(prompt: string) {
  if (!this.currentId) await this.createConversationAction()
  const composer = useComposerStore()
  const conversationId = this.currentId as number

  // 乐观插入用户消息
  const tempMessage: Message = {
    id: Date.now(),
    conversation_id: conversationId,
    prompt,
    task_kind: composer.draft.attachment ? 'img2img_generic' : 'text2img',
    attachment_url: composer.draft.attachment ? URL.createObjectURL(composer.draft.attachment) : undefined,
    created_at: new Date().toISOString(),
    _pending: true,
  }
  this.messages[conversationId] = [...(this.messages[conversationId] || []), tempMessage]

  try {
    const { message, generation_id } = await sendMessageAPI({
      conversation_id: conversationId,
      prompt,
      size: composer.draft.size,
      style_id: composer.draft.style_id,
      scene_id: composer.draft.scene_id,
      layered: composer.draft.layered,
      layer_count: composer.draft.layer_count,
      attachment: composer.draft.attachment,
    })
    // 用服务端返回替换临时消息
    const list = this.messages[conversationId]
    const idx = list.findIndex(m => m.id === tempMessage.id)
    list.splice(idx, 1, { ...message, _generation_id: generation_id })
  } catch (err) {
    // 失败：标记消息为 failed，但不删除（用户能看到自己写过什么）
    const list = this.messages[conversationId]
    const idx = list.findIndex(m => m.id === tempMessage.id)
    if (idx >= 0) list[idx] = { ...list[idx], _error: err.message, _pending: false }
  }
}
```

**`types.ts` 扩展 Message**（小改）

```ts
export interface Message {
  id: number
  conversation_id: number
  prompt: string
  attachment_url?: string
  attachment_name?: string
  attachment_size?: number
  task_kind: 'text2img' | 'img_restore' | 'img_style_transfer' | 'img2img_generic'
  created_at?: string
  _pending?: boolean       // 前端乐观插入标记
  _error?: string          // 前端失败标记
  _generation_id?: number  // 关联生成任务
}
```

**验收标准**
- [ ] 有附图时 Network 看到 `multipart/form-data`，含 attachment 文件块
- [ ] 无附图时也走 multipart（attachment 字段缺省）
- [ ] 发送中用户气泡立刻出现（乐观渲染）
- [ ] 服务端返回后用真实 message 替换临时 message
- [ ] 失败时用户气泡保留 + 显示错误标记
- [ ] 附图发送成功后 `composerStore.draft.attachment` 被清空（reset）

**自测**
1. Composer 上传一张图 + 输入"修复" → 发送 → F12 Network 看到 multipart
2. 不上传 + 输入"一只猫" → 发送 → 也走 multipart 但无 attachment
3. 后端故意返回 500 → 用户气泡保留，有红色错误标记

**工作量**：~120 行

---

### F07 · 前端：MessageList 渲染（GPT 极简风）

**涉及文件**
- `web/src/components/chat/MessageList.vue`（重写，从 19 行扩到 ~80 行）
- `web/src/components/chat/MessageBubble.vue`（新建，用户气泡）
- `web/src/components/chat/ImageReply.vue`（新建，AI 回复）

**MessageList.vue 骨架**

```vue
<script setup lang="ts">
import { useConversationStore } from '@/stores/conversation'
import MessageBubble from './MessageBubble.vue'
import ImageReply from './ImageReply.vue'

const store = useConversationStore()
</script>

<template>
  <div class="mx-auto flex w-full max-w-3xl flex-col gap-4 px-4 py-6">
    <template v-for="msg in store.currentMessages" :key="msg.id">
      <MessageBubble :message="msg" />
      <ImageReply v-if="msg._generation_id || msg._pending" :message="msg" />
    </template>
  </div>
</template>
```

**MessageBubble.vue（用户气泡 · GPT 浅灰胶囊）**

```vue
<template>
  <div class="flex flex-col items-end gap-1.5">
    <!-- 附图缩略卡（仅当有 attachment_url） -->
    <div v-if="message.attachment_url" class="inline-flex max-w-[68%] items-center gap-2 rounded-2xl bg-slate-100 p-2 pr-3">
      <img :src="message.attachment_url" class="h-12 w-10 shrink-0 rounded-lg object-cover" :alt="message.attachment_name || '附图'" />
      <div class="min-w-0">
        <div class="truncate text-[12.5px] font-medium text-ink">{{ message.attachment_name || '附图' }}</div>
        <div v-if="message.attachment_size" class="text-[10.5px] text-slate-500">{{ formatSize(message.attachment_size) }}</div>
      </div>
    </div>
    <!-- 文字胶囊 -->
    <div class="max-w-[68%] rounded-3xl bg-slate-100 px-4 py-2.5 text-[14px] leading-relaxed text-ink">
      {{ message.prompt }}
    </div>
    <!-- 失败标记 -->
    <div v-if="message._error" class="text-[11px] text-red-500">发送失败：{{ message._error }}</div>
  </div>
</template>
```

**ImageReply.vue（AI 回复 · GPT 极简卡）**

```vue
<script setup lang="ts">
import { useGenerationPoll } from '@/composables/useGenerationPoll'  // F08 实现

const props = defineProps<{ message: Message }>()
const { generation, loading, error } = useGenerationPoll(props.message._generation_id)
</script>

<template>
  <div class="flex items-start gap-3">
    <div class="mt-1 grid h-7 w-7 shrink-0 place-items-center rounded-full bg-white ring-1 ring-line">
      <!-- 知画 logo SVG -->
    </div>
    <div class="min-w-0">
      <button class="mb-2 inline-flex items-center gap-1 text-[12.5px] text-slate-500 hover:text-ink">
        <span class="font-medium">知画</span>
        <span>·</span>
        <span v-if="loading">正在生成…</span>
        <span v-else-if="generation">已生成 {{ generation.duration }}s</span>
        <svg class="h-3 w-3" ...>›</svg>
      </button>

      <!-- 加载中：spinner + 占位 -->
      <div v-if="loading" class="aspect-square w-[320px] animate-pulse rounded-2xl bg-slate-100"></div>

      <!-- 失败 -->
      <div v-else-if="error" class="rounded-2xl bg-red-50 px-4 py-3 text-[13px] text-red-700">
        生成失败：{{ error }}
      </div>

      <!-- 成功：图直显 + 下载浮按钮 -->
      <div v-else-if="generation" class="group relative w-fit">
        <img :src="generation.image_url" class="rounded-2xl" :style="aspectStyle(generation.size)" />
        <button class="btn absolute bottom-3 right-3 grid h-9 w-9 place-items-center rounded-full bg-black/55 text-white opacity-90 backdrop-blur transition hover:bg-black/75 hover:opacity-100" @click="download">
          <svg ...>↓</svg>
        </button>
      </div>

      <!-- 下方小图标行 -->
      <div v-if="generation" class="mt-2 flex items-center gap-0.5 text-slate-500">
        <button class="btn grid h-8 w-8 place-items-center rounded-lg hover:bg-slate-100 hover:text-ink" @click="copyPrompt"><!-- copy icon --></button>
        <button class="btn grid h-8 w-8 place-items-center rounded-lg hover:bg-slate-100 hover:text-ink" @click="share"><!-- share icon --></button>
        <button class="btn grid h-8 w-8 place-items-center rounded-lg hover:bg-slate-100 hover:text-ink" @click="favorite"><!-- bookmark icon --></button>
        <button class="btn grid h-8 w-8 place-items-center rounded-lg hover:bg-slate-100 hover:text-ink" @click="more"><!-- more icon --></button>
        <span class="ml-2 text-[11px] text-slate-400">{{ sizeLabel }} · {{ generation.credits_cost }} 积分</span>
      </div>
    </div>
  </div>
</template>
```

**视觉硬性要求**（必须对齐 `mockups/chat-img2img.html`）
- 用户气泡：`bg-slate-100 rounded-3xl`
- AI 图片：**无白卡**，直接 `rounded-2xl`
- 下载按钮：**浮在图右下角**，`bg-black/55 backdrop-blur`
- 下方图标行：4 个 `h-8 w-8` 灰色 icon
- AI 名字：**知画**

**验收标准**
- [ ] 用户文字消息：浅灰胶囊
- [ ] 用户附图消息：缩略卡 + 文字胶囊垂直堆叠
- [ ] AI 回复加载中：`animate-pulse` 灰色占位 + 转圈或骨架
- [ ] AI 回复成功：图直显 + 下载浮按钮 + 4 个小图标
- [ ] AI 回复失败：红色错误卡
- [ ] 视觉与 `mockups/chat-img2img.html` 一致

**自测**
1. 发"一只猫"→ 用户气泡 → AI 加载 → AI 完成
2. 上传图 + "修复"→ 用户两块气泡 → AI 加载 → AI 完成
3. 故意失败 → AI 红色错误卡

**工作量**：~280 行（3 个组件 + ImageReply 占大头）

---

### F08 · 前端：生成进度 SSE 对接 + 状态机

**涉及文件**
- `web/src/composables/useGenerationPoll.ts`（新建）

**`useGenerationPoll`**

```ts
import { ref, watch, onUnmounted } from 'vue'

export function useGenerationPoll(generationId: number | null | undefined) {
  const generation = ref<any>(null)
  const loading = ref(true)
  const error = ref<string | null>(null)
  let eventSource: EventSource | null = null

  function connect(id: number) {
    eventSource = new EventSource(`/api/generations/${id}/stream`)
    eventSource.onmessage = (e) => {
      const data = JSON.parse(e.data)
      generation.value = data
      if (data.status === 'succeeded') {
        loading.value = false
        eventSource?.close()
      } else if (data.status === 'failed') {
        loading.value = false
        error.value = data.error_msg || '生成失败'
        eventSource?.close()
      }
    }
    eventSource.onerror = () => {
      // fallback 轮询
      fallbackPoll(id)
    }
  }

  async function fallbackPoll(id: number) {
    eventSource?.close()
    const interval = setInterval(async () => {
      try {
        const res = await api.get(`/generations/${id}`)
        generation.value = res.data
        if (res.data.status === 'succeeded' || res.data.status === 'failed') {
          clearInterval(interval)
          loading.value = false
          if (res.data.status === 'failed') error.value = res.data.error_msg
        }
      } catch (e) { /* ignore one tick */ }
    }, 2000)
  }

  watch(() => generationId, (id) => {
    if (id) connect(id)
  }, { immediate: true })

  onUnmounted(() => eventSource?.close())

  return { generation, loading, error }
}
```

**验收标准**
- [ ] SSE 连通后实时收到进度事件
- [ ] 进度从 0% → 100% 平滑（中间状态显示在 ImageReply 里）
- [ ] SSE 断开自动降级轮询（2s 一次）
- [ ] 组件 unmount 时关闭连接
- [ ] 失败任务能识别并显示错误信息

**自测**
1. 发一条消息，F12 Network 看到 EventSource 连接
2. 后端模拟慢生成（10s），前端应该看到 spinner 持续，完成后切换到图片
3. 断网测试：SSE 断了后 fallback 轮询起作用

**工作量**：~120 行

---

### F09 · 前端：用户气泡 + AI 回复样式细节打磨

**涉及文件**
- `web/src/components/chat/MessageBubble.vue`（细调）
- `web/src/components/chat/ImageReply.vue`（细调）

**对齐 mockup 的细节**

打开 `mockups/chat-img2img.html` 在浏览器，**像素级对照** F07 实现的组件，把以下细节抠到位：

- 字号、行高、padding 精确到 px
- 浅灰色值 `bg-slate-100`（不是 `bg-gray-100`）
- 圆角 `rounded-3xl`（用户胶囊）/ `rounded-2xl`（AI 图）
- 下载浮按钮 `bg-black/55`，hover `bg-black/75`
- 4 图标行 `h-8 w-8` + `gap-0.5`
- 知画 preamble 用 `text-[12.5px] text-slate-500`
- 末尾 `›` 角标暗示可展开

**验收标准**
- [ ] 在 1440px 视口下，前端实际渲染与 mockup 视觉一致（截图对比）
- [ ] 在 375px 视口下，气泡和图片不溢出
- [ ] 暗黑模式（如果项目启用）颜色协调（v1 可只做亮色）

**自测**
1. 浏览器并排打开 `localhost:5173/` 和 `mockups/chat-img2img.html`
2. 一项项对照视觉
3. 截图存档放进 PR

**工作量**：~50 行（多为微调）

---

## 6. 联调任务（F10）

### F10 · 端到端联调与回归

**涉及文件**：无新代码，纯测试 + 修 bug

**端到端场景清单**

1. **基线**：登录用户发"一只猫" → 看到用户气泡 → AI 加载 → 显示图 → 下载成功
2. **图生图**：上传老照片 + "修复" → 看到附图+文字两块气泡 → AI 加载 → 显示修复图 → 下载
3. **多轮**：连续发 3 条不同 prompt → 全部正常显示在流里 → 滚动能看到历史
4. **匿名**：登出 → 访问 `/` → 用游客身份发消息 → 走 anonymous fingerprint 流程
5. **失败回滚**：把后端 channel 关掉 → 发消息 → AI 红色错误卡 → 积分自动退回
6. **会话切换**：左栏切到另一个会话 → 消息流切换 → 各自独立
7. **持久化**：刷新页面 → 会话列表恢复 → 切到某会话 → 消息流恢复（调 `GET /messages`）
8. **附图大小**：上传 15MB 图 → toast 拒绝
9. **附图格式**：上传 .txt → toast 拒绝

**验收标准**
- [ ] 9 个场景全部通过并截图
- [ ] `cd web && npm run build` 全绿
- [ ] `go build ./... && go test ./controller/... ./model/...` 全绿
- [ ] 无 console error / warning
- [ ] 移动端（375px）能用

**工作量**：~0.5 天（修小 bug + 截图）

---

## 7. 文件清单

**新建（后端）**：
- `controller/message.go`
- `controller/message_test.go`

**修改（后端）**：
- `model/models.go`（+ Message struct, + Generation.MessageID）
- `model/main.go`（AutoMigrate）
- `router/main.go`（+ 2 路由）
- `service/generation.go`（+ img2img 分支）

**新建（前端）**：
- `web/src/api/conversation.ts`
- `web/src/api/message.ts`
- `web/src/composables/useGenerationPoll.ts`
- `web/src/components/chat/MessageBubble.vue`
- `web/src/components/chat/ImageReply.vue`

**修改（前端）**：
- `web/src/stores/conversation.ts`（重构所有 action）
- `web/src/components/chat/MessageList.vue`（重写）
- `web/src/api/types.ts`（扩展 Message）
- `web/src/components/chat/Composer.vue`（轻改 send 流程）

---

## 8. PR 描述模板

每个 PR 提交时复制以下模板：

```markdown
## F0X · [任务标题]

[一句话描述这个任务做了什么]

### 改动
- [文件 1]: [改动点]
- [文件 2]: [改动点]

### 截图 / 录屏
- [ ] [关键场景 1]
- [ ] [关键场景 2]

### 验收清单
- [x] [本任务的所有验收点]

### 自测
- [x] 跑通本任务所有自测步骤

### 关联
- 视觉参考：`mockups/chat-img2img.html`
- 任务文档：`docs/chat-img2img-plan.md` 的 F0X 节
```

---

## 9. DoD（完成定义）

整个计划完成的标志：

- [ ] F01 ~ F10 全部合并到主干
- [ ] 9 个端到端场景测试通过
- [ ] `npm run build` + `go build ./...` 全绿
- [ ] 视觉与 `mockups/chat-img2img.html` 对齐
- [ ] 已部署到 staging 并复测一次完整流程
- [ ] 历史 `/classic` 入口仍正常工作（不影响经典版）

---

## 10. 上手指引（给 Codex）

**推荐顺序**：F01 → F02 → F03 → F04（后端打底）→ F05 → F06 → F07 → F08 → F09（前端跟进）→ F10（联调）

**每个任务起手三步**：
1. 完整读本任务节内容（包括"涉及文件"和"验收标准"）
2. 跑一次"自测"清单中第一项，确认起点状态
3. 开始改代码

**遇到不确定时**：
- ❌ 不要自由发挥扩大范围
- ✅ 在 PR 评论里问，或者标注 `// TODO(F0X): need clarification on X`

**禁止跨任务**：F03 不要顺手做 F04 的事，每个 PR 单焦点。

---

**文档版本**：v2.0（基于真实代码现状重写）
**视觉参考**：`mockups/chat-img2img.html`

---

## 11. 游客会话登录后认领同步计划

### 11.1 背景

新版 Chat 允许游客直接生成。游客消息当前存在前端临时会话中，后端生成记录以 `anonymous_id` 落库。如果用户先试用、再注册或登录，应该把这段游客创作同步到登录账号的正式会话中，否则用户会觉得刚生成的内容丢失。

### 11.2 目标

- 游客生成后登录或注册，自动把当前浏览器游客会话同步为登录用户会话。
- 同步成功后，聊天列表出现正式会话，消息和生成图可继续查看。
- 原匿名 `generation` 更新为当前用户所有，进入用户历史。
- 同步失败不阻断登录，只在控制台记录，避免影响认证主流程。

### 11.3 后端实现细节

新增接口：

```text
POST /api/conversations/claim-guest
Authorization: Bearer <token>
X-Fingerprint: <browser fingerprint>
```

请求体：

```json
{
  "title": "游客创作",
  "messages": [
    {
      "generation_id": 123,
      "prompt": "提示词",
      "task_kind": "text2img",
      "size": "1024x1024",
      "style_id": "",
      "scene_id": "",
      "layered": false,
      "layer_count": 0
    }
  ]
}
```

安全策略：

- 后端不信任前端传 `anonymous_id`。
- 后端用 `X-Fingerprint + IP` 重新计算 `anonymous_id`。
- 只认领 `anonymous_id` 匹配、`user_id IS NULL`、`message_id IS NULL`、`is_deleted=false` 的 generation。
- 重复同步时，已绑定过的 generation 不会再次创建消息。
- 没有可认领记录时返回 `claimed=0`，不创建空会话。

写入流程：

1. 读取登录用户 ID 和 fingerprint。
2. 按请求里的 `generation_id` 查询可认领 generation。
3. 创建正式 `Conversation`。
4. 为每个 generation 创建 `Message`，补齐 prompt、size、task_kind、layered 等字段。
5. 更新 generation 的 `user_id` 和 `message_id`。
6. 更新 conversation 的 `msg_count`、`last_msg_at`、`is_layered`。
7. 返回 `{ conversation, messages, claimed }`。

### 11.4 前端实现细节

- 新增 `claimGuestConversation()` API。
- `conversationStore.syncGuestConversation()` 从本地 `-1` 游客会话读取消息并提交。
- 邮箱登录、邮箱注册、微信登录成功后，在跳转首页前调用同步。
- 同步成功后删除本地游客会话，插入后端正式会话和消息，`currentId` 指向正式会话。
- 同步失败时不影响登录。

### 11.5 自测清单

- [x] 游客生成后登录，正式会话创建成功。
- [x] 原 generation 的 `user_id`、`message_id` 正确更新。
- [x] 错误 fingerprint 无法认领。
- [x] 重复同步不重复创建消息。
- [x] `go test ./controller/... ./model/...` 通过。
- [x] `cmd /c npm run build` 通过。

### 11.6 完成记录

- 状态：已实现并自测通过，2026-05-17。
- 后端：新增 `POST /api/conversations/claim-guest`，按 fingerprint + IP 计算 `anonymous_id`，只认领未绑定的游客 generation。
- 前端：登录、注册、微信登录成功后调用 `syncGuestConversation()`，把 `-1` 游客会话转换为正式会话。
- 自测：`go test ./controller/... ./model/...`、`cmd /c npm run build` 均通过。
- 计划提交信息：`feat: claim guest chat after login`
