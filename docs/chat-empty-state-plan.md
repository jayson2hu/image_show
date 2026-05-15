# 空状态实现计划（E01 · 独立任务）

> 任务编号：**E01**
> 视觉参考：`mockups/chat-empty-state.html`
> 执行时机：在主计划 `chat-redesign-plan.md`（T01–T30 + T32）全部完成之后开始
> 阅读须知：**本文档自包含**，执行 E01 时无需阅读其它文档

---

## 0. 上下文：你接手时仓库的状态

执行本任务前，主计划已经完成以下产物。理解这些"已经存在的东西"对完成 E01 至关重要。

### 0.1 已存在的路由

```ts
// web/src/router/index.ts
{ path: '/',         name: 'chat',    component: () => import('@/views/Chat.vue') },
{ path: '/classic',  name: 'classic', component: () => import('@/views/Home.vue') },
// ... 其它路由不变
```

### 0.2 已存在的视图：`web/src/views/Chat.vue`

```vue
<script setup lang="ts">
import { computed } from 'vue'
import { useConversationStore } from '@/stores/conversation'
import SessionList from '@/components/chat/SessionList.vue'
import ChatHeader from '@/components/chat/ChatHeader.vue'
import MessageList from '@/components/chat/MessageList.vue'   // 历史消息流
import Composer from '@/components/chat/Composer.vue'

const conversationStore = useConversationStore()
// ... 其它逻辑
</script>

<template>
  <div class="flex h-screen w-screen overflow-hidden">
    <SessionList />
    <main class="flex min-w-0 flex-1 flex-col">
      <ChatHeader />
      <div class="scroll-y flex-1 overflow-y-auto">
        <!-- 当前：只有 MessageList，需要在 E01 改成条件渲染 -->
        <MessageList />
      </div>
      <Composer />
    </main>
  </div>
</template>
```

### 0.3 已存在的 Store：`web/src/stores/conversation.ts`

```ts
import { defineStore } from 'pinia'
import type { Conversation, Message } from '@/api/types'

export const useConversationStore = defineStore('conversation', {
  state: () => ({
    list: [] as Conversation[],
    currentId: null as number | null,
    messages: {} as Record<number, Message[]>,
    loading: false,
    searchQuery: '',
  }),
  getters: {
    currentConversation: (state) => state.list.find(c => c.id === state.currentId) || null,
    currentMessages: (state) => state.currentId ? (state.messages[state.currentId] || []) : [],
  },
  actions: {
    selectConversation(id: number) { /* ... */ },
    sendMessage(payload: {...}) { /* ... */ },
    // ...
  },
})
```

类型定义（`web/src/api/types.ts`）：
```ts
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
  // ...
}

export interface Scene {
  id: number
  name: string
  icon: string
  description: string
  prompt_template: string
  recommended_ratio: string
  credit_cost: number
  default_layered: boolean
  default_layer_count: number
  sort_order: number
}
```

### 0.4 已存在的 Composer：`web/src/components/chat/Composer.vue`

当前 Composer 的对外契约（**E01 要在它上面加一个 prop**）：

```vue
<script setup lang="ts">
interface Props {
  // 暂无 prop，所有形态都一样
}
const props = defineProps<Props>()

// 内部：
// - textarea v-model 写到 composerStore.draft.prompt
// - 第一行：场景 chip 横滚 + 推荐样例 popover
// - 第二行：textarea
// - 第三行：内联 chip（风格 / 比例 / 分层 / 质量 / 附图）+ 积分预估 + 发送按钮
</script>
```

### 0.5 已存在的 Composer Store：`web/src/stores/composer.ts`

```ts
export const useComposerStore = defineStore('composer', {
  state: () => ({
    draft: {
      prompt: '',
      size: 'square',
      quality: 'medium',
      style_id: '',
      scene_id: '',
      layered: false,
      layer_count: 5,
      attachment: null as File | null,
    },
  }),
  actions: {
    setDraft(patch: Partial<typeof this.draft>) { Object.assign(this.draft, patch) },
    focusInput() { /* emit 一个事件让 Composer 内的 textarea 拿焦点 */ },
    reset() { /* 重置 draft */ },
  },
})
```

### 0.6 已存在的用户 Store：`web/src/stores/user.ts`

```ts
useUserStore().user                  // { id, email, username, credits, ... } | null
useUserStore().user?.username        // 可能为空
useUserStore().user?.email           // 可能为空
useUserStore().user?.credits         // number
```

### 0.7 已存在的接口：`GET /api/generation/scenes`

返回：
```json
{
  "items": [
    {
      "id": 1,
      "name": "小红书封面",
      "icon": "📸",
      "description": "...",
      "prompt_template": "...",
      "recommended_ratio": "portrait_3_4",
      "credit_cost": 2,
      "default_layered": false,
      "default_layer_count": 5,
      "sort_order": 40
    },
    ...
  ]
}
```

前端已封装在 `web/src/api/scene.ts`：
```ts
export async function listScenes(): Promise<{ items: Scene[] }>
```

### 0.8 已存在的色板（`web/tailwind.config.js`）

```js
colors: {
  ink:   '#17202a',
  mist:  '#eef4f7',
  coral: '#e56f5a',
  teal:  '#177e89',
  // E01 实现时无需新增颜色
}
```

---

## 1. 设计原则

**极简，不做引导教学**。用户已经在产品里，不需要 hero 横幅、6 大瓦片、推荐宫格、上传引导、使用提示。空状态只需要让用户**最快开始第一条消息**，把决策成本降到接近零。

参考模型：ChatGPT / Claude 的空对话页——一句问候 + 几个建议 chip + 居中输入框。

---

## 2. 视觉结构（与 `mockups/chat-empty-state.html` 完全对齐）

```
左栏（SessionList，不动）        |  主区
                                |
                                |   ──────── ChatHeader ────────
                                |   "新对话"          [切换经典版]
                                |   ─────────────────────────────
                                |
                                |        ┌ 垂直居中容器（max-w 720px）─┐
                                |        │                              │
                                |        │  问候语                       │
                                |        │  你好，准备好创作了吗？        │
                                |        │  输入一句话，或选下方场景      │
                                |        │                              │
                                |        │  场景小按钮（5 个，居中排列） │
                                |        │  [📸小红书] [🛒商品] [🎨海报]  │
                                |        │  [👤头像]   [✨自由创作]       │
                                |        │                              │
                                |        │  输入框（Composer compact）   │
                                |        │  ┌─────────────────────────┐  │
                                |        │  │ textarea...             │  │
                                |        │  │ [chips] [积分预估] [↑]  │  │
                                |        │  └─────────────────────────┘  │
                                |        │                              │
                                |        │  Enter发送 · 余 N 积分 · 计费 │
                                |        └──────────────────────────────┘
```

**关键约束**
- 主区**垂直水平居中**，不是顶部对齐（用 `flex items-center justify-center`）
- 场景小按钮是 **chip 风格**，不是大瓦片，**不带描述文字、不带缩略图**
- 场景按钮**只显示 5 个**（包括"海报设计"和"自由创作"），多了反而干扰
- 输入框**和对话状态共用同一个 Composer 组件**，靠 `compact` prop 切换形态

---

## 3. 与对话状态的差异（实现关键）

| 维度 | 空状态 | 对话进行中 |
|---|---|---|
| 主区内容 | 居中的"问候 + 场景 chips + 居中输入框" | 顶部对齐的消息流 + 底部固定输入框 |
| 输入框位置 | **垂直水平居中** | **底部固定** |
| Composer prop | `compact: true` | `compact: false`（默认） |
| Composer 内：场景 chip 行 | **隐藏** | 显示横向 6 个 chip |
| Composer 内：推荐样例 popover | **隐藏** | 显示 chip + popover |
| Composer 内：textarea + 内联 chip 行 + 积分 + 发送 | **保留** | 保留 |
| 顶栏标题 | "新对话" + "未保存"灰标签 | 会话标题 + 元信息 |

---

## 4. 实现步骤

### 4.1 新建组件 `web/src/components/chat/ChatEmptyState.vue`

完整代码骨架：

```vue
<script setup lang="ts">
import { computed, onMounted, ref, nextTick } from 'vue'
import { useUserStore } from '@/stores/user'
import { useComposerStore } from '@/stores/composer'
import { listScenes } from '@/api/scene'
import type { Scene } from '@/api/types'
import Composer from './Composer.vue'

const userStore = useUserStore()
const composerStore = useComposerStore()
const scenes = ref<Scene[]>([])

// 4.1.1 动态问候语
const greeting = computed(() => {
  const hour = new Date().getHours()
  const name = userStore.user?.username || userStore.user?.email?.split('@')[0] || ''
  const timePart = hour < 6
    ? '夜深了'
    : hour < 12
      ? '早上好'
      : hour < 18
        ? '下午好'
        : '晚上好'
  if (name) {
    return `${timePart}，${name}，准备好创作了吗？`
  }
  return '你好，准备好创作了吗？'
})

// 4.1.2 场景按钮：取前 5 个（按 sort_order 升序，后端已排序，前端再 slice 一次以防）
const displayScenes = computed(() => {
  return [...scenes.value]
    .sort((a, b) => a.sort_order - b.sort_order)
    .slice(0, 5)
})

// 4.1.3 加载场景
onMounted(async () => {
  try {
    const res = await listScenes()
    scenes.value = res.items
  } catch (e) {
    console.error('Load scenes failed', e)
  }
})

// 4.1.4 场景点击：填充 composer 草稿 + 聚焦输入框（不直接发送）
function onSceneClick(scene: Scene) {
  composerStore.setDraft({
    prompt: scene.prompt_template,
    size: scene.recommended_ratio,
    scene_id: String(scene.id),
    layered: !!scene.default_layered,
    layer_count: scene.default_layer_count || 5,
  })
  nextTick(() => composerStore.focusInput())
}
</script>

<template>
  <div class="flex h-full items-center justify-center px-6">
    <div class="w-full max-w-[720px]">

      <!-- 问候语 -->
      <div class="mb-6 text-center">
        <h2 class="text-[22px] font-semibold leading-tight text-ink">
          {{ greeting }}
        </h2>
        <p class="mt-1.5 text-[13px] text-muted">
          输入一句话，或选下方的场景快速开始
        </p>
      </div>

      <!-- 场景小按钮：横向居中 -->
      <div class="mb-3 flex flex-wrap items-center justify-center gap-2">
        <button
          v-for="scene in displayScenes"
          :key="scene.id"
          class="scene-chip inline-flex items-center gap-1.5 rounded-full bg-white px-3.5 py-2 text-[13px] text-ink ring-1 ring-line transition hover:-translate-y-px hover:ring-teal/45"
          @click="onSceneClick(scene)"
        >
          <span>{{ scene.icon }}</span>
          <span>{{ scene.name }}</span>
          <span
            v-if="scene.default_layered"
            class="rounded-full bg-coral/10 px-1.5 py-0.5 text-[9.5px] font-medium text-coral"
          >自动分层</span>
        </button>
      </div>

      <!-- 复用 Composer，开 compact 模式 -->
      <Composer compact />

      <!-- 辅助提示 -->
      <div class="mt-3 text-center text-[11px] text-muted">
        Enter 发送 · Shift+Enter 换行 · 余
        <span class="font-semibold text-ink">{{ userStore.user?.credits ?? 0 }}</span>
        积分 ·
        <a href="#" class="text-teal hover:underline">计费规则</a>
      </div>

    </div>
  </div>
</template>
```

### 4.2 修改 Composer：新增 `compact` prop

`web/src/components/chat/Composer.vue` 改动点：

```vue
<script setup lang="ts">
interface Props {
  compact?: boolean   // 新增
}
const props = withDefaults(defineProps<Props>(), { compact: false })

// ... 其它逻辑不变
</script>

<template>
  <div class="...">
    <!-- 第一行：场景 chip 行 + 推荐样例 popover —— compact 时隐藏 -->
    <div v-if="!props.compact" class="mb-2 flex items-center gap-1.5 overflow-x-auto pb-1">
      <!-- 原来的场景 chip 行 -->
      <!-- 原来的推荐样例 popover chip -->
    </div>

    <!-- 海报场景自动启用分层的提示条 —— compact 时也隐藏 -->
    <div v-if="!props.compact && showLayerHint" class="mb-2 ...">
      ...
    </div>

    <!-- 主输入卡：textarea + 内联 chip 行 + 发送 —— 始终显示 -->
    <div class="composer rounded-2xl bg-white p-2 shadow-soft ring-1 ring-line">
      <textarea ... />
      <div class="mt-1 flex flex-wrap items-center gap-1">
        <!-- 风格 / 比例 / 分层 / 质量 / 附图 chip -->
        <!-- 积分预估 -->
        <!-- 发送按钮 -->
      </div>
    </div>

    <!-- 第三行：辅助提示 —— compact 时也隐藏（空状态自己有辅助文字） -->
    <div v-if="!props.compact" class="mt-1.5 flex ...">
      ...
    </div>
  </div>
</template>
```

### 4.3 修改 `Chat.vue`：接入条件渲染

```vue
<script setup lang="ts">
import { computed } from 'vue'
import { useConversationStore } from '@/stores/conversation'
import SessionList from '@/components/chat/SessionList.vue'
import ChatHeader from '@/components/chat/ChatHeader.vue'
import MessageList from '@/components/chat/MessageList.vue'
import ChatEmptyState from '@/components/chat/ChatEmptyState.vue'   // 新增
import Composer from '@/components/chat/Composer.vue'

const conversationStore = useConversationStore()

// 新增：是否为空状态
const isEmpty = computed(() => {
  const conv = conversationStore.currentConversation
  if (!conv) return true
  const msgs = conversationStore.messages[conv.id]
  return !msgs || msgs.length === 0
})
</script>

<template>
  <div class="flex h-screen w-screen overflow-hidden">
    <SessionList />
    <main class="flex min-w-0 flex-1 flex-col">
      <ChatHeader />

      <!-- 空状态：ChatEmptyState 内部自带 Composer(compact)，
           所以这里 main 区不再渲染外层 Composer。
           对话状态：渲染滚动消息流 + 底部固定 Composer。 -->
      <template v-if="isEmpty">
        <ChatEmptyState class="flex-1" />
      </template>
      <template v-else>
        <div class="scroll-y flex-1 overflow-y-auto">
          <MessageList />
        </div>
        <Composer />
      </template>
    </main>
  </div>
</template>
```

> ⚠️ 注意：空状态时只渲染 `<ChatEmptyState>`，**不要在 main 里再额外渲染 Composer**——因为 ChatEmptyState 内部已经渲染了一个 `<Composer compact />`。

---

## 5. 验收标准

### 视觉
- [ ] 空状态主区垂直水平居中（不是顶部对齐）
- [ ] 问候语根据当前时间和用户名动态变化（早上好/下午好/晚上好/夜深了）
- [ ] 场景小按钮居中排列，最多 5 个
- [ ] "海报设计"按钮末尾显示 coral 色 "自动分层" 小徽标（前提：后台已配置 `default_layered=true`）
- [ ] 输入框在场景按钮下方，宽度 max 720px
- [ ] 输入框 `autofocus` 进入页面即获得焦点
- [ ] 输入框底部内联 chip 行**保留**（风格 / 比例 / 分层 / 质量 / 附图 / 积分 / 发送）
- [ ] 输入框上方的"场景 chip 行"和"推荐样例 popover" **隐藏**
- [ ] 空状态下页面 lighthouse 性能 ≥ 90

### 交互
- [ ] 点击场景按钮 → textarea 立刻填充 prompt
- [ ] 点击场景按钮 → 比例 chip 同步更新为该场景的 `recommended_ratio`
- [ ] 点击"海报设计" → 自动开启分层（分层 chip 变 coral 高亮，显示 "5 层"）
- [ ] 点击场景按钮 → 输入框获得焦点，光标在末尾
- [ ] 点击场景按钮 **不会自动发送**，用户需要再确认或修改后点发送
- [ ] 在 textarea 直接打字 + Enter → 发送成功 → 视图切换到对话流（空状态消失）
- [ ] 第一条消息发送后立刻有用户气泡 + 加载中 AI 卡

### 边界
- [ ] 后台只配置 3 个场景 → 只显示 3 个按钮，不留空槽
- [ ] 后台配置 10 个场景 → 前端只取前 5 个（按 sort_order），其它不显示
- [ ] 用户未登录访问 `/`（如允许匿名）→ 问候语回退到"你好，准备好创作了吗？"
- [ ] 移动端 < 640px → 场景按钮自动换行，输入框宽度自适应
- [ ] 接口 `GET /api/generation/scenes` 失败 → 场景区域不显示按钮，但输入框照常可用，console 有 error 但页面不崩

---

## 6. 自测步骤

完成实现后，按以下顺序在本地手动验证：

1. **登录后访问 `/`**：自动新建会话或选中第一个无消息的会话 → 看到居中的空状态
2. **核对问候语**：根据当前小时是否正确（< 6 显示"夜深了"，< 12 显示"早上好"，< 18 显示"下午好"，其它"晚上好"）；尝试登出再访问，应回退到"你好，准备好创作了吗？"
3. **点击"海报设计"按钮**：
   - textarea 自动填入海报 prompt
   - 比例 chip 显示 "3:4 竖版"
   - 分层 chip 变 coral 高亮，显示 "5 层"
   - 积分预估变为 7 积分（2 基础 + 5×1 分层）
   - 输入框获得焦点（光标可见在末尾）
4. **直接打字"一只猫"按 Enter** → 看到用户气泡 + 进度卡，空状态消失，进入对话流布局，底部固定 Composer 出现
5. **新建第二个会话**（点左栏"新建创作"）→ 再次看到空状态
6. **后台 SQL** `UPDATE prompt_templates SET sort_order = 999 WHERE label = '社交头像';` → 刷新前端，该场景从空状态消失（因为排到第 6 位以后）
7. **F12 切到 iPhone 12 视口** → 场景按钮换行，输入框宽度自适应，问候语字号仍可读
8. **打开 Network 面板**，把 `GET /api/generation/scenes` 改成 500 错误 → 场景区不渲染按钮，输入框可用，console 有 error 但页面不崩

每一步通过后截图，附在 PR 描述里。

---

## 7. 不做的事（明确排除）

主计划早期讨论过这些扩展，全部**砍掉**，不要在 E01 里做：

- ❌ 不做 Hero glow / 渐变光晕
- ❌ 不做 3×2 场景大瓦片（带描述、缩略图、比例标签那种）
- ❌ 不做推荐样例宫格（推荐样例只在对话流里以 popover 形式出现）
- ❌ 不做附图引导横幅（附图入口就是输入框里的"附图" chip）
- ❌ 不做"使用提示"卡片（计费规则链接已经够了）
- ❌ 不做空状态的"上传图片做修复"独立按钮——附图 chip 已经够
- ❌ 不做拖拽上传到空状态区域的视觉（拖到输入框就行）
- ❌ 不做欢迎弹窗 / 新手引导浮层

如果实现过程中觉得"加一点 X 会更好"，先停下来在 PR 评论里问一下，不要自由发挥。

---

## 8. 工作量估算

| 项 | 行数 | 工时 |
|---|---|---|
| 新建 `ChatEmptyState.vue` | ~140 | 2.5 h |
| 修改 `Composer.vue` 增加 compact prop（包覆 v-if） | ~30 | 0.5 h |
| 修改 `Chat.vue` 接入条件渲染 | ~15 | 0.3 h |
| 自测 + 截图 + PR 描述 | — | 0.7 h |
| **总计** | **~185** | **半天（4 小时）** |

---

## 9. 文件清单

新建：
- `web/src/components/chat/ChatEmptyState.vue`

修改：
- `web/src/components/chat/Composer.vue`（新增 `compact: boolean` prop + 用 v-if 包覆三个非 compact 区域）
- `web/src/views/Chat.vue`（新增 `isEmpty` computed + 条件渲染）

---

## 10. PR 描述模板

提交 PR 时复制以下模板：

```markdown
## E01 · Chat 空状态实现

实现了用户进入 /chat 且当前会话无消息时的极简空状态视图。

### 改动
- 新建 `ChatEmptyState.vue`：居中问候 + 5 个场景小 chip + 居中 Composer(compact)
- `Composer.vue` 新增 `compact: boolean` prop
- `Chat.vue` 根据 `isEmpty` 条件渲染

### 截图
- [ ] 空状态全屏（PC）
- [ ] 点击海报场景后的状态（验证自动分层）
- [ ] 发送第一条消息后切换到对话流（验证状态切换）
- [ ] 移动端视口（iPhone 12）

### 验收清单
- [x] 视觉清单全部勾选
- [x] 交互清单全部勾选
- [x] 边界清单全部勾选
- [x] 自测 8 步全部跑通

### 关联文档
- 视觉参考：`mockups/chat-empty-state.html`
- 任务定义：`docs/chat-empty-state-plan.md`
```

---

## 11. 验收完成定义（DoD）

- [ ] 视觉所有项打勾
- [ ] 交互所有项打勾
- [ ] 边界所有项打勾
- [ ] 自测 8 步全部跑通并截图
- [ ] `cd web && npm run build` 全绿（vue-tsc + vite build 无报错）
- [ ] `go build ./...` 全绿（E01 不动后端，但起码确认没破坏）
- [ ] PR 描述按第 10 节模板填写

---

**文档版本**：v2.0（自包含版）
**最后更新**：2026-05-15
**视觉参考**：`mockups/chat-empty-state.html`
