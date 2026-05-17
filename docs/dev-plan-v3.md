# Chat 图片界面优化 - 开发计划

> 基于 UI 审核稿 docs/ui-review-v3.html  
> 技术栈：Vue 3 + Pinia + TailwindCSS  
> 前端目录：web/src/  
> 交付标准：每个子任务独立可验收

---

## 开发进度

| 功能 | 状态 | 提交 | 自测/验收 |
|------|------|------|-----------|
| 1 侧边栏可收起 | 已完成 | `4eb1ece` | `pnpm build` 通过 |
| 2 聊天记录命名修复 | 已完成 | `16e25a3` | `pnpm build`、`go test ./controller` 通过 |
| 3 输入框样式统一 | 已完成 | `6f60ffa` | `pnpm build` 通过 |
| 4 图片操作按钮优化 | 已完成 | `f27393b`（代码随 `4bd1929` 落地） | `pnpm build` 通过 |
| 5 图片尺寸自适应限制 | 已完成 | `4bd1929` | `pnpm build` 通过 |
| 6 页面固定一屏不滚动 | 已完成 | `b8486e6` | `pnpm build` 通过 |

最近更新时间：2026-05-17

---

## 功能 1：侧边栏可收起

### 1.1 新建 UI Store

**文件：** web/src/stores/ui.ts（新建）

**任务：**
- 新建 Pinia store useUiStore
- state: sidebarCollapsed: boolean（默认 false）
- action: 	oggleSidebar() 切换状态
- 初始化时从 localStorage.getItem('sidebar_collapsed') 读取
- 变更时同步写入 localStorage

**验收：**
- [ ] store 可正常导入使用
- [ ] toggleSidebar 切换 sidebarCollapsed
- [ ] 刷新页面后状态保持

**自测：**
1. 调用 toggleSidebar → 值变为 true
2. 刷新页面 → 仍为 true
3. 再次调用 → 变为 false

---

### 1.2 SessionList 展开态改造

**文件：** web/src/components/chat/SessionList.vue

**任务：**
- 引入 useUiStore
- 展开态（!uiStore.sidebarCollapsed）：
  - 顶部：左侧「对话列表」文字 + 右侧两个按钮（+ 新建 + 收起图标）
  - 收起图标：sidebar panel 图标（方框+竖线），点击调用 uiStore.toggleSidebar()
  - 对话列表保持现有样式
- 宽度保持 w-72

**验收：**
- [ ] 顶部显示「对话列表」+ 两个图标按钮
- [ ] 点击 + 创建新对话
- [ ] 点击收起图标 → sidebarCollapsed 变为 true

**自测：**
1. 页面加载 → 展开态正常显示
2. 点击 + → 新对话出现
3. 点击收起 → 侧边栏切换

---

### 1.3 SessionList 收起态

**文件：** web/src/components/chat/SessionList.vue

**任务：**
- 收起态（uiStore.sidebarCollapsed）：
  - 宽度 w-14（56px）
  - 顶部：展开按钮（同一个 sidebar panel 图标）
  - 下方：+ 新建按钮（w-8 h-8 rounded-lg border）
  - 中间：对话列表用首字圆形图标（w-8 h-8 rounded-lg，取 title[0]）
  - 当前选中项 g-mist，其他 hover:bg-slate-100
  - 每个图标 	itle 属性显示完整标题
  - 点击图标切换对话
- 使用 -if/v-else 或条件 class 切换两种模式

**验收：**
- [ ] 收起态宽度 56px
- [ ] 显示展开按钮、新建按钮、首字图标列表
- [ ] 当前对话高亮
- [ ] 点击图标切换对话
- [ ] 点击展开按钮恢复展开态
- [ ] hover 显示 title tooltip

**自测：**
1. 收起态 → 56px 宽
2. 点击对话图标 → 主内容切换
3. 点击展开 → 恢复完整侧边栏
4. hover 图标 → tooltip 显示标题

---

### 1.4 Chat.vue 布局适配

**文件：** web/src/views/Chat.vue

**任务：**
- SessionList 组件根据 sidebarCollapsed 动态切换宽度
- 添加过渡动画 	ransition-all duration-200
- 移动端（lg 以下）行为不变

**验收：**
- [ ] 切换时有平滑过渡
- [ ] 主内容区宽度自适应
- [ ] 移动端不受影响

**自测：**
1. 切换侧边栏 → 动画平滑
2. 主内容区宽度变化
3. 窗口缩小到 lg 以下 → 侧边栏隐藏


---

## 功能 2：聊天记录命名修复

### 2.1 自动命名逻辑

**文件：** web/src/stores/conversation.ts

**任务：**
- 修改 sendLocalMessage action
- 当对话 msg_count 从 0 变为 1 时，title 设为 prompt.slice(0, 12) + (prompt.length > 12 ? '...' : '')
- 后续消息不再自动改名

**验收：**
- [ ] 新对话默认 title「新对话」
- [ ] 首条消息后 title 更新为 prompt 前 12 字
- [ ] 超 12 字加 ...
- [ ] 第二条消息不改名

**自测：**
1. 新建对话 → title = "新对话"
2. 发送 "生成一张市集，有摊位，有人群" → title = "生成一张市集，有摊位，..."
3. 再发一条 → title 不变
4. 发送 5 字 → title 为完整 5 字无 ...

---

### 2.2 对话项右键菜单

**文件：** web/src/components/chat/SessionList.vue

**任务：**
- 每个对话项右侧添加三点更多按钮（hover 时显示，当前选中项始终显示）
- 点击弹出下拉菜单：「重命名」「删除」
- 删除项红色
- 点击外部关闭菜单（@click.outside 或 document click listener）
- 同时只能打开一个菜单

**验收：**
- [ ] hover 显示三点按钮
- [ ] 点击弹出菜单
- [ ] 菜单含重命名和删除
- [ ] 点击外部关闭
- [ ] 同时只开一个

**自测：**
1. hover → 三点出现
2. 点击三点 → 菜单弹出
3. 点击空白 → 关闭
4. 打开 A 后点 B 的三点 → A 关，B 开

---

### 2.3 重命名编辑态

**文件：** web/src/components/chat/SessionList.vue

**任务：**
- 点击「重命名」→ 对话项变为 input 编辑态
- input 边框 order-2 border-teal，自动获得焦点
- 下方「确认」「取消」按钮
- Enter 确认，Esc 取消
- 确认时更新 store 中 title（如后端 API 可用则调用 PATCH）
- 空 title 不允许保存
- 最大 128 字符

**验收：**
- [ ] 点击重命名 → input 出现并获焦
- [ ] Enter/确认 → 保存
- [ ] Esc/取消 → 恢复
- [ ] 空值不可保存
- [ ] 超 128 字截断

**自测：**
1. 重命名 → input 出现
2. 改名 → 确认 → 列表更新
3. 改名 → Esc → 恢复原名
4. 清空 → 确认按钮 disabled

---

### 2.4 删除对话

**文件：** web/src/components/chat/SessionList.vue, web/src/stores/conversation.ts

**任务：**
- 点击「删除」→ 确认对话框（使用已有 ConfirmDialog 或 window.confirm）
- 确认后从 list 移除，清除 messages
- 删除当前对话 → 自动选中第一条
- 列表为空 → 创建新对话
- 如后端 API 可用则调用 DELETE

**验收：**
- [ ] 删除弹确认
- [ ] 确认后移除
- [ ] 删除当前对话 → 切换
- [ ] 列表空 → 新建
- [ ] 取消不执行

**自测：**
1. 删除非当前 → 列表减少，当前不变
2. 删除当前 → 切换到第一条
3. 删除最后一条 → 自动新建
4. 取消 → 无变化


---

## 功能 3：输入框样式统一

### 3.1 Composer 始终使用统一样式

**文件：** web/src/components/chat/Composer.vue, web/src/views/Chat.vue

**任务：**
- 移除 Composer 的 compact prop 及相关逻辑
- Composer 始终使用完整样式：mx-auto max-w-3xl rounded-2xl border border-slate-200 bg-white p-2 shadow-sm
- 外层容器始终为：order-t border-slate-200 bg-white/90 p-4（或 g-mist p-4，与页面背景协调）
- 场景标签行和底部提示行保留（在有消息时也显示）

**验收：**
- [ ] 空状态和有消息状态的输入框样式完全一致
- [ ] 输入框居中，最大宽度 max-w-3xl
- [ ] 有圆角边框和阴影
- [ ] 工具栏（风格/比例/分层/附图/发送）正常显示

**自测：**
1. 空状态 → 输入框居中大框
2. 发送消息后 → 输入框样式不变
3. 输入文字 → 正常发送
4. 工具栏按钮正常工作

---

### 3.2 Chat.vue 布局调整

**文件：** web/src/views/Chat.vue

**任务：**
- 修改布局：无论空状态还是有消息，都显示 MessageList（空时为空列表）+ 底部 Composer
- ChatEmptyState 改为在 MessageList 区域内显示（当消息为空时）
- 结构变为：
  `
  <main class="flex flex-col h-full">
    <ChatHeader />
    <div class="flex-1 overflow-y-auto">
      <ChatEmptyState v-if="isEmpty" />
      <MessageList v-else />
    </div>
    <Composer />  <!-- 始终在底部 -->
  </main>
  `
- ChatEmptyState 不再内嵌 Composer（移除其中的 <Composer compact />）

**验收：**
- [ ] Composer 始终在页面底部
- [ ] 空状态时中间显示欢迎语和场景卡片（无输入框）
- [ ] 有消息时中间显示消息列表
- [ ] 整体布局 h-screen 不溢出

**自测：**
1. 空状态 → 中间欢迎语 + 底部输入框
2. 发送消息 → 中间变为消息列表 + 底部输入框不变
3. 页面无外部滚动条

---

### 3.3 ChatEmptyState 移除内嵌 Composer

**文件：** web/src/components/chat/ChatEmptyState.vue

**任务：**
- 移除 import Composer from './Composer.vue' 和 <Composer compact />
- 移除底部的 "Enter 发送 · Shift+Enter 换行" 提示（由外部 Composer 自带）
- 保留欢迎语、场景卡片、积分显示

**验收：**
- [ ] ChatEmptyState 不再包含输入框
- [ ] 欢迎语和场景卡片正常显示
- [ ] 点击场景卡片仍能填充 composerStore.draft

**自测：**
1. 空状态 → 只显示欢迎语和场景
2. 点击场景 → 底部 Composer 的 prompt 被填充
3. 无重复输入框


---

## 功能 4：图片操作按钮优化

### 4.1 调整按钮顺序和可见性

**文件：** web/src/components/chat/ImageReply.vue

**任务：**
- 图片下方操作按钮行改为：编辑 | 复制提示词 | 下载 | 三点更多
- 移除原来直接显示的「收藏」按钮
- 新增「编辑」按钮（铅笔图标），放在第一个位置
- 编辑按钮点击行为：将当前 message 的 prompt 填入 Composer，并附上当前图片作为 attachment（调用 composerStore.setDraft）

**验收：**
- [ ] 按钮顺序：编辑、复制、下载、更多
- [ ] 收藏不再直接显示
- [ ] 编辑按钮有铅笔图标
- [ ] 点击编辑 → Composer 填入 prompt

**自测：**
1. 图片生成后 → 下方显示 4 个按钮
2. 点击编辑 → Composer prompt 被填入
3. 点击复制 → toast "提示词已复制"
4. 点击下载 → 图片下载

---

### 4.2 三点更多菜单内容优化

**文件：** web/src/components/chat/ImageReply.vue

**任务：**
- 三点菜单内容改为：
  - 收藏（图标 + 文字）
  - 复制图片链接（图标 + 文字）
  - 分隔线
  - 信息区（灰色文字）：尺寸、分层状态、任务 ID
- 菜单样式：w-44 rounded-xl border bg-white py-1 shadow-lg
- 点击外部关闭

**验收：**
- [ ] 三点菜单包含收藏、复制链接、信息
- [ ] 收藏和复制链接可点击
- [ ] 信息区为灰色不可点击
- [ ] 点击外部关闭菜单

**自测：**
1. 点击三点 → 菜单弹出
2. 点击收藏 → toast 提示
3. 点击复制链接 → 复制成功
4. 点击外部 → 菜单关闭

---

## 功能 5：图片尺寸自适应限制

### 5.1 根据比例限制图片最大尺寸

**文件：** web/src/components/chat/ImageReply.vue

**任务：**
- 根据 message.size 计算图片的 max-width 和 max-height class
- 规则：
  - square / 1024x1024 → max-w-[300px] max-h-[300px]
  - portrait_3_4 / 1152x1536 → max-h-[400px]（宽度自适应）
  - story / 1008x1792 → max-h-[400px]（宽度自适应）
  - landscape_4_3 / 1536x1152 → max-w-[500px]（高度自适应）
  - widescreen / 1792x1008 → max-w-[500px]（高度自适应）
  - 未知比例 → max-w-[400px] max-h-[400px]
- 图片使用 object-contain rounded-2xl cursor-pointer
- 移除当前的 max-h-[520px]

**验收：**
- [ ] 方图显示不超过 300×300
- [ ] 竖图高度不超过 400px
- [ ] 横图宽度不超过 500px
- [ ] 图片保持原始比例
- [ ] 未知比例有 fallback 限制

**自测：**
1. 生成方图 → 显示 300×300 以内
2. 生成竖图 → 高度 400px 以内
3. 生成横图 → 宽度 500px 以内
4. 图片不变形

---

### 5.2 点击图片放大查看

**文件：** web/src/components/chat/ImageReply.vue

**任务：**
- 点击图片打开全屏预览（modal overlay）
- 全屏预览：黑色半透明背景 + 居中显示原图 + 右上角关闭按钮
- 按 Esc 或点击背景关闭
- 可复用已有的 ImagePreview 组件（web/src/components/ImagePreview.vue），或在 ImageReply 内实现简单 modal

**验收：**
- [ ] 点击图片 → 全屏预览打开
- [ ] 预览显示原始大小图片（受屏幕限制）
- [ ] 点击背景或 Esc 关闭
- [ ] 右上角有关闭按钮

**自测：**
1. 点击缩略图 → 全屏预览出现
2. 按 Esc → 关闭
3. 点击黑色背景 → 关闭
4. 点击关闭按钮 → 关闭


---

## 功能 6：页面固定一屏不滚动

### 6.1 确保 Chat.vue 布局为 h-screen flex-col

**文件：** web/src/views/Chat.vue

**任务：**
- 确认根容器为 lex h-screen w-screen overflow-hidden（当前已是）
- 确认消息区域为 lex-1 overflow-y-auto（当前已是）
- 确认 Composer 为 shrink-0（不被压缩）
- 确认 ChatHeader 为 shrink-0
- 问题可能出在图片太大导致消息区域内容撑出 → 通过功能 5 的图片尺寸限制解决

**验收：**
- [ ] 页面无外部滚动条（body 不滚动）
- [ ] 消息区域内部可滚动
- [ ] Composer 始终可见在底部
- [ ] ChatHeader 始终可见在顶部
- [ ] 新消息自动滚动到底部

**自测：**
1. 生成多张图片 → 页面无外部滚动条
2. 消息区域可上下滚动
3. 输入框始终在底部可见
4. 发送新消息 → 自动滚动到最新

---

### 6.2 消息列表自动滚动到底部

**文件：** web/src/components/chat/MessageList.vue

**任务：**
- 当新消息添加时，自动滚动到底部
- 使用 
extTick + scrollIntoView 或 scrollTop = scrollHeight
- 滚动容器为 MessageList 的父级 div.overflow-y-auto（在 Chat.vue 中）
- 可通过 ef 获取滚动容器，watch messages 变化时触发

**验收：**
- [ ] 发送消息后自动滚动到底部
- [ ] 收到图片回复后自动滚动
- [ ] 用户手动滚动到上方时不强制拉回（可选：仅在底部时自动滚动）

**自测：**
1. 发送消息 → 滚动到底部
2. 图片加载完成 → 滚动到底部
3. 手动滚动到上方 → 不被强制拉回（可选）

---

## 开发顺序

| 阶段 | 任务 | 预估 | 依赖 |
|------|------|------|------|
| P0 | 6.1 页面布局确认 | 0.5h | 无 |
| P0 | 5.1 图片尺寸限制 | 1h | 无 |
| P0 | 5.2 点击放大 | 1h | 5.1 |
| P1 | 3.1 Composer 统一样式 | 1h | 无 |
| P1 | 3.2 Chat.vue 布局调整 | 1h | 3.1 |
| P1 | 3.3 ChatEmptyState 移除 Composer | 0.5h | 3.2 |
| P1 | 6.2 自动滚动 | 0.5h | 3.2 |
| P2 | 4.1 操作按钮调整 | 1h | 无 |
| P2 | 4.2 三点菜单优化 | 0.5h | 4.1 |
| P3 | 1.1 UI Store | 0.5h | 无 |
| P3 | 1.2 展开态改造 | 0.5h | 1.1 |
| P3 | 1.3 收起态实现 | 1.5h | 1.1 |
| P3 | 1.4 布局适配 | 0.5h | 1.2, 1.3 |
| P4 | 2.1 自动命名 | 0.5h | 无 |
| P4 | 2.2 右键菜单 | 1h | 无 |
| P4 | 2.3 重命名编辑态 | 1h | 2.2 |
| P4 | 2.4 删除对话 | 0.5h | 2.2 |

**总预估：约 12.5h**

---

## 注意事项

1. **功能 5 和 6 优先**：图片尺寸限制是解决「页面需要滚动」的根本原因，优先处理
2. **Composer 统一是核心改动**：涉及 Chat.vue 和 ChatEmptyState 的结构变化，需仔细测试场景卡片点击等交互
3. **侧边栏收起**：与之前 dev-plan.md 中的方案一致，可复用
4. **编辑按钮**：当前只做「填入 prompt 到 Composer」，不做图片编辑器
5. **自动滚动**：注意图片异步加载完成后高度变化，可能需要 ResizeObserver 或 img onload 回调

---

*文档版本：v3 | 对应 UI 审核稿：docs/ui-review-v3.html*
