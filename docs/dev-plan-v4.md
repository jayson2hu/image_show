# Chat 界面修复优化 - 开发计划 v4

> 基于截图反馈的 6 个问题修复  
> 前端目录：web/src/  
> 每个子任务独立可验收

---

## 开发进度

| BUG | 状态 | 提交 | 自测/验收 |
|-----|------|------|-----------|
| 1 背景色统一 + 去掉边框感 | 已完成 | `ebe7c17`（sidebar 分隔随 `0beaebb` 落地） | `pnpm build` 通过 |
| 2 新建会话无效 | 已完成 | `0beaebb` | `pnpm build` 通过 |
| 3 分层生成无效（游客模式） | 已完成 | `eadde30` | `pnpm build`、`go test ./controller ./service` 通过 |
| 4 发送按钮改为黑色箭头 | 已完成 | `259b275`（代码随 `ebe7c17` 落地） | `pnpm build` 通过 |
| 5 页面外层滚动条修复 | 已完成 | `68700c9` | `pnpm build` 通过 |

最近更新时间：2026-05-17

---

## BUG 1：背景色统一 + 去掉边框感

### 1.1 左侧 Sidebar 背景色保持白色

**文件：** web/src/components/chat/SessionList.vue

**问题：** 左侧 sidebar 和右侧主内容区颜色不一致，但这是设计意图（左白右浅蓝）。问题是分界线太明显。

**任务：**
- 左侧 sidebar 保持 g-white
- 移除 sidebar 与主内容区之间的 order-r border-slate-200，改为无边框或极淡的分隔（order-r border-slate-100）
- 右侧主内容区保持 g-mist

**验收：**
- [ ] 左侧白色，右侧浅蓝色
- [ ] 左右之间无明显分界线（无粗边框）
- [ ] 视觉上自然过渡

**自测：**
1. 页面加载 → 左白右浅蓝，无明显竖线分隔
2. 收起侧边栏 → 同样无明显分隔

---

### 1.2 去掉 ChatHeader 底部边框

**文件：** web/src/components/chat/ChatHeader.vue

**问题：** header 底部有 order-b border-slate-200，造成上下分隔感。

**任务：**
- 移除 order-b border-slate-200
- header 背景改为透明或与主内容区一致（g-transparent 或 g-mist）
- 保留 ackdrop-blur 效果（可选）

**验收：**
- [ ] header 与消息区域之间无明显横线
- [ ] header 内容正常显示
- [ ] 滚动时 header 不会与消息内容混淆（保留微弱阴影或 backdrop-blur）

**自测：**
1. 页面加载 → header 和消息区无明显分隔线
2. 滚动消息 → header 仍可辨识

---

### 1.3 去掉 Composer 区域的边框感

**文件：** web/src/components/chat/Composer.vue

**问题：** Composer 外层容器有 order border-slate-200 的圆角框，加上 order-t 分隔线，边框感太重。

**任务：**
- Composer 外层容器（ounded-2xl border border-slate-200）改为更淡的边框：order-slate-100 或去掉 border 只保留 shadow-sm
- 移除 Composer 上方的 order-t（如果有的话，在 Chat.vue 中）
- 保持输入框的圆角和阴影，但减弱边框

**验收：**
- [ ] Composer 区域无明显粗边框
- [ ] 输入框仍有圆角和轻微阴影
- [ ] 与消息区域之间无明显横线分隔

**自测：**
1. 页面加载 → 输入框区域边框极淡或无
2. 输入文字 → 正常工作
3. 整体视觉更干净


---

## BUG 2：新建会话无效

### 2.1 修复 createConversation 逻辑

**文件：** web/src/components/chat/SessionList.vue, web/src/stores/conversation.ts

**问题分析：**
SessionList 中 createConversation() 调用的是 conversationStore.ensureConversation()。
ensureConversation() 的逻辑是：
- 未登录时：如果已存在 guest conversation（id=-1），直接选中它，不会创建新的
- 已登录时：如果 currentId 已有值，直接返回，不会创建新的

所以点击「新建」按钮实际上只是切换到已有对话，不会真正创建新对话。

**任务：**
- 修改 SessionList 中的 createConversation() 函数：
  - 已登录用户：直接调用 conversationStore.createLocalConversation()（调用后端 POST /conversations）
  - 未登录用户：清空当前 guest conversation 的消息，重置为空对话状态（或创建新的 guest conversation）
- 确保新建后 currentId 切换到新对话
- 确保新建后 Composer 的 draft 被 reset

**验收：**
- [ ] 已登录：点击新建 → 创建新对话，列表新增一条，切换到新对话
- [ ] 未登录：点击新建 → 清空当前对话消息，回到空状态
- [ ] 收起侧边栏时点击 + 按钮同样有效
- [ ] 展开侧边栏时点击 + 按钮同样有效

**自测：**
1. 已登录 → 点击 + → 列表新增「新对话」，主内容区变为空状态
2. 未登录 → 点击 + → 当前对话清空，回到欢迎页
3. 收起态点击 + → 同样有效
4. 多次点击 → 不报错（已登录时每次创建新对话）

---

## BUG 3：分层生成无效（游客模式）

### 3.1 游客模式传递 layered 参数

**文件：** web/src/stores/conversation.ts, web/src/api/generation.ts

**问题分析：**
在 sendMessage 中，游客模式走的是 createGeneration() 或 createImageEdit()。
但 createGeneration 的 payload 只有 prompt 和 size，没有传 layered 和 layer_count。
所以即使前端开启了分层，后端也收不到这个参数。

**任务：**
- 修改 CreateGenerationPayload 接口，添加可选字段：layered?: boolean, layer_count?: number, style_id?: string
- 修改 createGeneration 函数，将这些字段传给后端
- 修改 conversation.ts 中游客模式的调用，传入完整参数

**验收：**
- [ ] 游客模式开启分层后发送 → 请求 payload 包含 layered: true 和 layer_count: 5
- [ ] 游客模式选择风格后发送 → 请求 payload 包含 style_id
- [ ] 已登录模式不受影响（走 createMessage）

**自测：**
1. 未登录 → 开启分层 → 发送 → F12 Network 查看请求 body 包含 layered=true, layer_count=5
2. 未登录 → 选择写实风格 → 发送 → 请求包含 style_id
3. 已登录 → 分层正常工作（走 createMessage 路径）


---

## BUG 4：发送按钮改为黑色箭头

### 4.1 发送按钮样式优化

**文件：** web/src/components/chat/Composer.vue

**问题：** 当前发送按钮是 teal 色圆形 + "发送" 文字，需要改为黑色圆形 + 白色上箭头图标（参考 image7 ChatGPT 风格）。

**任务：**
- 修改发送按钮样式：
  - 背景色：g-ink（黑色），hover 时 g-ink/80
  - disabled 时：g-slate-300
  - 尺寸保持 size-9（36px）圆形
  - 内容：移除 "发送" 文字，改为白色上箭头 SVG 图标
  - 箭头 SVG：<path d="M12 19V5m0 0l-5 5m5-5l5 5" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
- loading 状态保持 spinner 不变

**验收：**
- [ ] 发送按钮为黑色圆形 + 白色上箭头
- [ ] hover 时颜色微变
- [ ] disabled 时灰色
- [ ] loading 时显示 spinner
- [ ] 无 "发送" 文字

**自测：**
1. 页面加载 → 发送按钮为黑色圆形箭头
2. 输入文字 → 按钮可点击（黑色）
3. 清空文字 → 按钮 disabled（灰色）
4. 点击发送 → spinner 出现
5. 发送完成 → 恢复箭头

---

## BUG 5：页面外层滚动条修复

### 5.1 确保页面 h-screen overflow-hidden

**文件：** web/src/views/Chat.vue, web/index.html

**问题：** 从 image6 可以看到页面右侧出现了滚动条，说明内容超出了视口高度。

**任务：**
- 确认 Chat.vue 根容器为 h-screen overflow-hidden（当前已有 lex h-screen w-screen overflow-hidden）
- 检查是否有子元素撑出了 flex 容器：
  - ChatHeader 需要 shrink-0
  - Composer 需要 shrink-0
  - 消息区域需要 lex-1 overflow-y-auto min-h-0（关键：min-h-0 防止 flex 子项不收缩）
- 检查 web/index.html 的 body/html 是否有 overflow: hidden 或 h-full
- 在 index.html 或全局 CSS 中添加：
  `css
  html, body { height: 100%; overflow: hidden; margin: 0; }
  `

**验收：**
- [ ] 页面无外部滚动条（浏览器右侧无滚动条）
- [ ] 消息区域内部可滚动
- [ ] Composer 始终可见在底部
- [ ] ChatHeader 始终可见在顶部
- [ ] 不同屏幕尺寸下均无外部滚动

**自测：**
1. 打开页面 → 无外部滚动条
2. 生成多张图片 → 消息区内部滚动，页面不滚动
3. 缩小浏览器窗口 → 仍无外部滚动条
4. F12 检查 html/body 无 scrollbar

---

### 5.2 消息区域添加 min-h-0

**文件：** web/src/views/Chat.vue

**问题：** flex 布局中，子元素默认 min-height: auto，可能导致内容撑出容器。

**任务：**
- 在消息区域的 div.flex-1.overflow-y-auto 上添加 min-h-0 class
- 这是 flex 布局中防止子项溢出的关键修复

**验收：**
- [ ] 消息区域不会撑出父容器
- [ ] 内容多时内部滚动
- [ ] 与 BUG 5.1 配合解决外部滚动条问题

**自测：**
1. 生成大图 → 消息区内部滚动
2. 页面无外部滚动条


---

## 开发顺序

| 优先级 | 任务 | 预估 | 说明 |
|--------|------|------|------|
| P0 | 5.1 页面 overflow-hidden | 0.5h | 最影响体验的 bug |
| P0 | 5.2 min-h-0 修复 | 0.5h | 配合 5.1 |
| P0 | 2.1 新建会话修复 | 1h | 功能性 bug |
| P0 | 3.1 分层参数传递 | 0.5h | 功能性 bug |
| P1 | 4.1 发送按钮改箭头 | 0.5h | UI 优化 |
| P1 | 1.1 sidebar 边框 | 0.5h | 视觉优化 |
| P1 | 1.2 header 边框 | 0.5h | 视觉优化 |
| P1 | 1.3 composer 边框 | 0.5h | 视觉优化 |

**总预估：约 4.5h**

---

## 具体代码修改指引

### BUG 2 修复代码参考

SessionList.vue 中 createConversation 改为：

`	ypescript
async function createConversation() {
  closeMenu()
  const userStore = useUserStore()
  if (userStore.token) {
    // 已登录：调用后端创建新对话
    await conversationStore.createLocalConversation()
  } else {
    // 未登录：重置 guest conversation
    conversationStore.messages[-1] = []
    const guest = conversationStore.list.find(c => c.id === -1)
    if (guest) {
      guest.title = '新对话'
      guest.msg_count = 0
    }
    conversationStore.currentId = -1
  }
  // reset composer
  const { useComposerStore } = await import('@/stores/composer')
  useComposerStore().reset()
}
`

### BUG 3 修复代码参考

generation.ts 修改：

`	ypescript
export interface CreateGenerationPayload {
  prompt: string
  size: string
  style_id?: string
  layered?: boolean
  layer_count?: number
  captcha_token?: string
}

export function createGeneration(payload: CreateGenerationPayload) {
  return api.post<CreateGenerationResponse>('/generations', {
    prompt: payload.prompt,
    size: payload.size,
    style_id: payload.style_id || '',
    layered: payload.layered || false,
    layer_count: payload.layer_count || 0,
    captcha_token: payload.captcha_token || '',
  })
}
`

conversation.ts 游客模式调用修改：

`	ypescript
const response = payload.attachment
  ? await createImageEdit({ prompt: normalizedPrompt, size: payload.size, image: payload.attachment })
  : await createGeneration({
      prompt: normalizedPrompt,
      size: payload.size,
      style_id: payload.style_id,
      layered: payload.layered,
      layer_count: payload.layer_count,
    })
`

### BUG 4 修复代码参考

Composer.vue 发送按钮改为：

`html
<button
  class="flex size-9 items-center justify-center rounded-full bg-ink text-white transition hover:bg-ink/80 disabled:cursor-not-allowed disabled:bg-slate-300"
  type="button"
  :disabled="!composerStore.draft.prompt.trim() || conversationStore.sending"
  @click="send"
>
  <svg v-if="conversationStore.sending" class="size-4 animate-spin" fill="none" viewBox="0 0 24 24">
    <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
    <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 0 1 8-8v4a4 4 0 0 0-4 4H4Z" />
  </svg>
  <svg v-else class="size-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 19V5m0 0l-5 5m5-5l5 5" />
  </svg>
</button>
`

### BUG 5 修复代码参考

Chat.vue 消息区域：

`html
<div class="flex-1 overflow-y-auto min-h-0">
  <MessageList />
</div>
`

index.html body 添加：

`html
<style>html, body { height: 100%; overflow: hidden; margin: 0; }</style>
`

---

## 注意事项

1. **不改动经典版**：所有修改仅针对 Chat 界面
2. **边框优化要适度**：不是完全去掉所有边框，而是减弱（order-slate-200 → order-slate-100 或去掉）
3. **新建会话**：需要同时处理登录和未登录两种情况
4. **分层参数**：后端需要确认 /generations 接口是否接受 layered 参数，如果后端不支持则需要后端配合
5. **发送按钮**：箭头方向为向上（↑），参考 ChatGPT 的设计

---

*文档版本：v4 | 修复截图反馈的 6 个问题*
