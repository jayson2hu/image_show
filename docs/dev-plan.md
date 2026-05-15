# Chat UI 优化开发计划

> 目标：完成 Chat 界面 8 个优化点的开发  
> 技术栈：Vue 3 + Pinia + TailwindCSS + Go/Gin 后端  
> 交付标准：每个子任务独立可验收，附自测 checklist

---

## 全局约定

- 前端代码目录：web/src/
- 后端代码目录：项目根目录（Go）
- 样式方案：TailwindCSS，自定义色值在 	ailwind.config.js（ink/mist/coral/teal）
- 状态管理：Pinia stores（web/src/stores/）
- API 层：web/src/api/
- 组件目录：web/src/components/chat/
- 路由：web/src/router/index.ts

---

## 功能 1：新建创作按钮 UI 优化

### 1.1 改造 SessionList 顶部布局

**涉及文件：** web/src/components/chat/SessionList.vue

**任务描述：**
- 移除当前的全宽黑色「新建创作」按钮
- 替换为顶部横排布局：左侧文字「对话列表」+ 右侧 + 图标按钮
- + 按钮样式：w-8 h-8 rounded-lg border border-slate-200，hover 时 order-teal bg-teal/5
- 图标使用 inline SVG：<path d="M12 5v14M5 12h14"/> stroke-width=2

**验收标准：**
- [ ] 顶部显示「对话列表」文字 + 右侧 + 图标按钮
- [ ] 点击 + 按钮调用 conversationStore.createLocalConversation()
- [ ] hover 时按钮边框变为 teal 色
- [ ] 按钮尺寸 32x32px，图标 16x16px

**自测：**
1. 页面加载后侧边栏顶部显示正确布局
2. 点击 + 创建新对话，对话列表新增一条
3. 多次点击不报错
4. 移动端（lg 以下）侧边栏隐藏时不影响主内容

---

## 功能 2：对话重命名

### 2.1 自动命名（首条消息后）

**涉及文件：** web/src/stores/conversation.ts

**任务描述：**
- 修改 sendLocalMessage action
- 当对话的 msg_count 从 0 变为 1 时，自动将 title 设为 prompt 的前 12 个字符
- 超过 12 字加 ... 后缀

**验收标准：**
- [ ] 新对话默认 title 为「新对话」
- [ ] 发送第一条消息后，title 自动更新为 prompt 前 12 字
- [ ] 超过 12 字显示 ...
- [ ] 第二条及之后的消息不再自动改名
- [ ] 空 prompt 不触发改名

**自测：**
1. 创建新对话 → title 显示「新对话」
2. 输入 "未来城市夜景，湿润街道反射霓虹灯" 发送 → title 变为 "未来城市夜景，湿润街道反..."
3. 再发一条消息 → title 不变
4. 输入 5 个字发送 → title 为完整 5 个字，无 ...


### 2.2 右键菜单（更多按钮）

**涉及文件：** web/src/components/chat/SessionList.vue

**任务描述：**
- 每个对话项右侧添加三点更多按钮（hover 时显示）
- 点击弹出下拉菜单，包含「重命名」和「删除」两个选项
- 菜单样式：g-white rounded-lg shadow-lg border border-slate-200 py-1
- 点击外部区域关闭菜单

**验收标准：**
- [ ] hover 对话项时右侧出现三点图标
- [ ] 点击三点图标弹出菜单
- [ ] 菜单包含「重命名」「删除」两项
- [ ] 删除项为红色文字
- [ ] 点击菜单外部关闭菜单
- [ ] 同时只能打开一个菜单

**自测：**
1. hover 对话项 → 三点按钮出现
2. 点击三点 → 菜单弹出
3. 点击空白处 → 菜单关闭
4. 打开 A 的菜单后点击 B 的三点 → A 关闭，B 打开

### 2.3 重命名编辑态

**涉及文件：** web/src/components/chat/SessionList.vue

**任务描述：**
- 点击菜单中「重命名」后，对话项进入编辑态
- 编辑态：title 变为 input 输入框，border 为 teal
- 输入框下方显示「确认」「取消」两个小按钮
- 确认：调用 conversationStore 更新 title（本地更新，如有后端 API 则同步调用 PATCH）
- 取消：恢复原 title
- Enter 键确认，Esc 键取消
- title 最大 128 字符

**验收标准：**
- [ ] 点击重命名后进入编辑态，input 获得焦点
- [ ] input 显示当前 title，可编辑
- [ ] 点击确认或按 Enter → 保存新 title，退出编辑态
- [ ] 点击取消或按 Esc → 恢复原 title，退出编辑态
- [ ] 空 title 不允许保存（按钮 disabled 或提示）
- [ ] 超过 128 字截断

**自测：**
1. 重命名 → input 出现，焦点在 input
2. 修改文字 → 点确认 → 列表显示新名称
3. 修改文字 → 按 Esc → 恢复原名
4. 清空 input → 确认按钮不可点击
5. 输入超长文字 → 自动截断到 128 字

### 2.4 删除对话

**涉及文件：** web/src/components/chat/SessionList.vue, web/src/stores/conversation.ts

**任务描述：**
- 点击菜单中「删除」后，弹出确认对话框（使用已有的 ConfirmDialog 组件）
- 确认后从 list 中移除该对话，清除对应 messages
- 如果删除的是当前选中对话，自动选中列表第一条；如果列表为空则创建新对话
- 如有后端 API 则调用 DELETE

**验收标准：**
- [ ] 点击删除弹出确认框
- [ ] 确认后对话从列表消失
- [ ] 删除当前对话后自动切换到第一条
- [ ] 列表为空时自动创建新对话
- [ ] 取消不执行删除

**自测：**
1. 删除非当前对话 → 列表减少一条，当前对话不变
2. 删除当前对话 → 切换到列表第一条
3. 删除最后一条 → 自动创建新对话
4. 点取消 → 无变化

---

## 功能 3：分层/风格/比例配置优化

### 3.1 移除对话框上方的自动分层提示

**涉及文件：** web/src/components/chat/ChatEmptyState.vue

**任务描述：**
- 在场景卡片中，移除 自动分层 的 badge 标签（<span v-if="scene.default_layered" ...>自动分层</span>）
- 选择海报设计场景时不再自动开启分层（不设置 layered: true）
- 分层完全由用户在底部工具栏手动控制

**验收标准：**
- [ ] 场景卡片上不再显示「自动分层」标签
- [ ] 点击海报设计场景后，分层开关保持关闭状态
- [ ] 其他场景属性（prompt_template, recommended_ratio）正常填充

**自测：**
1. 页面加载 → 场景卡片无「自动分层」文字
2. 点击海报设计 → composer 中分层按钮为关闭状态
3. 点击其他场景 → 正常填充 prompt 和比例

### 3.2 风格下拉选择（含「无」选项）

**涉及文件：** web/src/components/chat/Composer.vue, web/src/stores/composer.ts

**任务描述：**
- 将原来的「风格」纯文字按钮改为下拉选择器
- 下拉第一项为「无（模型决定）」，value 为空字符串
- 其他选项从后端 /generation/options 接口获取（已有 API）
- 前端 fallback 默认值：写实、动漫、幻想、赛博朋克、水彩、插画
- 选中后按钮文字更新为当前风格名
- 下拉面板样式：g-white rounded-lg shadow-lg border border-slate-200 py-1
- 点击外部关闭下拉

**验收标准：**
- [ ] 点击风格按钮弹出下拉面板
- [ ] 第一项为「无（模型决定）」
- [ ] 选中某项后按钮文字更新
- [ ] 默认选中「无」
- [ ] 后端有配置时使用后端数据，无配置时使用 fallback
- [ ] 点击外部关闭下拉
- [ ] 选中项有 ✓ 标记

**自测：**
1. 点击风格按钮 → 下拉出现
2. 选择「写实」→ 按钮显示「写实」，draft.style_id 更新
3. 选择「无」→ 按钮显示「风格：无」，draft.style_id 为空
4. 点击外部 → 下拉关闭
5. 断网时 → 使用 fallback 列表

### 3.3 比例下拉选择

**涉及文件：** web/src/components/chat/Composer.vue

**任务描述：**
- 将原来的比例纯文字按钮改为下拉选择器
- 选项从后端获取，fallback：1:1 方图、3:4 竖版、9:16 竖屏、4:3 横版、16:9 宽屏
- 选中后按钮文字更新
- 样式与风格下拉一致

**验收标准：**
- [ ] 点击比例按钮弹出下拉
- [ ] 包含 5 个比例选项
- [ ] 选中后按钮文字和 draft.size 更新
- [ ] 默认为 square（1:1 方图）
- [ ] 后端有配置时优先使用

**自测：**
1. 点击比例按钮 → 下拉出现
2. 选择 3:4 竖版 → 按钮更新，draft.size = 'portrait_3_4'
3. 积分估算随比例变化更新


### 3.4 分层开关优化

**涉及文件：** web/src/components/chat/Composer.vue

**任务描述：**
- 保留底部工具栏的「分层」按钮作为开关
- 开启状态：g-coral text-white，显示「分层 · 开」+ 小开关图标（右侧圆点）
- 关闭状态：g-mist text-ink，显示「分层 · 关」+ 小开关图标（左侧圆点）
- 点击切换 composerStore.draft.layered
- 开启时 layer_count 使用默认值 5（后台可配置）

**验收标准：**
- [ ] 分层按钮默认为关闭状态（灰色）
- [ ] 点击切换为开启状态（coral 色）
- [ ] 再次点击切换回关闭
- [ ] 开启时显示「分层 · 开」，关闭时显示「分层 · 关」
- [ ] 开启时积分估算包含分层费用
- [ ] 关闭时积分估算不含分层费用

**自测：**
1. 默认状态 → 灰色「分层 · 关」
2. 点击 → coral 色「分层 · 开」，积分增加
3. 再点击 → 恢复灰色，积分减少
4. 发送消息时 draft.layered 值正确

### 3.5 隐藏质量按钮

**涉及文件：** web/src/components/chat/Composer.vue

**任务描述：**
- 移除工具栏中的「质量」按钮
- draft.quality 保持默认值 'medium'，不暴露给用户
- 后端可通过配置覆盖

**验收标准：**
- [ ] 工具栏中不显示「质量」按钮
- [ ] draft.quality 始终为 'medium'
- [ ] 发送请求时 quality 字段正常传递

**自测：**
1. 页面加载 → 工具栏无「质量」按钮
2. 发送生成请求 → payload 中 quality = 'medium'

---

## 功能 4：附图上传

### 4.1 附图按钮交互

**涉及文件：** web/src/components/chat/Composer.vue

**任务描述：**
- 点击「附图」按钮切换上传区域的显示/隐藏
- 有附图时按钮样式变为：g-teal/10 border border-teal/30 text-teal，显示「附图 ✓」
- 无附图时为普通 mist 样式

**验收标准：**
- [ ] 点击附图按钮展开上传区域
- [ ] 再次点击收起上传区域
- [ ] 有图片时按钮变为 teal 高亮 + ✓
- [ ] 移除图片后按钮恢复普通样式

**自测：**
1. 点击附图 → 上传区域出现
2. 再点击 → 上传区域消失
3. 上传图片后 → 按钮变为「附图 ✓」teal 色
4. 删除图片 → 按钮恢复

### 4.2 图片上传区域

**涉及文件：** web/src/components/chat/Composer.vue

**任务描述：**
- 在 textarea 和工具栏之间插入上传区域
- 未上传时：虚线边框区域，中间显示上传图标 + 「点击或拖拽上传图片」
- 支持点击选择文件和拖拽上传
- 文件限制：JPG/PNG/WebP，最大 10MB
- 上传后显示缩略图预览（80x80px），右上角 × 删除按钮
- 文件存储到 composerStore.draft.attachment

**验收标准：**
- [ ] 展开后显示虚线上传区域
- [ ] 点击区域触发文件选择器
- [ ] 拖拽文件到区域可上传
- [ ] 仅接受 image/jpeg, image/png, image/webp
- [ ] 超过 10MB 提示错误（toast）
- [ ] 上传后显示缩略图预览
- [ ] 点击 × 删除图片，清空 draft.attachment
- [ ] 重新选择图片替换旧图

**自测：**
1. 展开上传区 → 显示虚线框和提示文字
2. 点击选择 JPG 文件 → 显示缩略图
3. 拖拽 PNG 文件 → 显示缩略图
4. 选择 GIF 文件 → 提示格式不支持
5. 选择 15MB 文件 → 提示超过大小限制
6. 点击 × → 图片消失，draft.attachment = null
7. 发送消息后 → 上传区域收起，attachment 清空

---

## 功能 5：侧边栏收起

### 5.1 侧边栏状态管理

**涉及文件：** web/src/stores/conversation.ts（或新建 web/src/stores/ui.ts）

**任务描述：**
- 新增 sidebarCollapsed 状态（boolean，默认 false）
- 提供 	oggleSidebar() action
- 状态持久化到 localStorage（key: sidebar_collapsed）

**验收标准：**
- [ ] 初始状态从 localStorage 读取，默认 false
- [ ] toggleSidebar 切换状态并同步 localStorage
- [ ] 刷新页面后状态保持

**自测：**
1. 首次加载 → sidebarCollapsed = false
2. 调用 toggleSidebar → 变为 true，localStorage 更新
3. 刷新页面 → 仍为 true

### 5.2 展开态侧边栏改造

**涉及文件：** web/src/components/chat/SessionList.vue

**任务描述：**
- 顶部右侧添加收起按钮（双箭头左图标），在 + 按钮旁边
- 底部添加用户信息区：头像（首字母圆形）+ 积分显示
- 收起按钮点击调用 	oggleSidebar()

**验收标准：**
- [ ] 顶部显示：「对话列表」+ + 按钮 + 收起按钮
- [ ] 底部显示用户头像和积分
- [ ] 游客显示「游客」+ 积分
- [ ] 点击收起按钮切换到收起态

**自测：**
1. 展开态 → 顶部有收起按钮
2. 底部显示用户信息
3. 点击收起 → 切换到窄条模式

### 5.3 收起态侧边栏

**涉及文件：** web/src/components/chat/SessionList.vue

**任务描述：**
- 收起态宽度 56px（w-14）
- 顶部：展开按钮（双箭头右图标）
- 下方：+ 新建按钮
- 中间：对话列表用首字圆形图标表示（取 title 第一个字）
- 当前选中项高亮（bg-mist）
- 底部：用户头像圆形
- 点击展开按钮恢复展开态
- 点击对话图标切换对话

**验收标准：**
- [ ] 收起态宽度 56px
- [ ] 显示展开按钮、新建按钮、对话首字图标、用户头像
- [ ] 当前对话图标高亮
- [ ] 点击对话图标切换对话
- [ ] 点击新建按钮创建新对话
- [ ] 点击展开按钮恢复展开态
- [ ] hover 图标显示 title tooltip

**自测：**
1. 收起态 → 宽度 56px，显示图标列表
2. 点击对话图标 → 主内容切换到该对话
3. 点击 + → 创建新对话
4. 点击展开 → 恢复完整侧边栏
5. hover 图标 → tooltip 显示完整 title

### 5.4 Chat.vue 布局适配

**涉及文件：** web/src/views/Chat.vue

**任务描述：**
- 根据 sidebarCollapsed 状态动态调整 SessionList 宽度
- 主内容区自适应剩余宽度
- 过渡动画：	ransition-all duration-200

**验收标准：**
- [ ] 展开态侧边栏 w-72，收起态 w-14
- [ ] 切换时有平滑过渡动画
- [ ] 主内容区宽度自适应
- [ ] 移动端（lg 以下）行为不变（侧边栏隐藏）

**自测：**
1. 切换侧边栏 → 有平滑动画
2. 主内容区宽度随之变化
3. 窗口缩小到移动端 → 侧边栏隐藏


---

## 功能 6：游客积分支持（Chat 页面）

### 6.1 前端游客积分显示

**涉及文件：** web/src/components/chat/ChatEmptyState.vue, web/src/stores/user.ts

**任务描述：**
- 未登录用户进入 Chat 页面时，从后端获取游客可用积分（通过 fingerprint）
- 在 ChatEmptyState 中显示体验积分提示：teal 色圆角标签「体验积分：X」
- 积分数量从后端 site config 或 /auth/me（匿名版）获取
- 如果后端未返回，前端 fallback 显示 5

**验收标准：**
- [ ] 未登录用户看到积分提示
- [ ] 积分数量从后端获取
- [ ] 后端无响应时 fallback 为 5
- [ ] 已登录用户不显示「体验积分」标签（显示正常积分）

**自测：**
1. 未登录访问 / → 显示「体验积分：5」
2. 登录后 → 不显示体验积分标签，显示正常积分
3. 断网时 → fallback 显示 5

### 6.2 欢迎标语后台配置

**涉及文件：** web/src/components/chat/ChatEmptyState.vue, web/src/api/site.ts

**任务描述：**
- greeting 文案从后端 site config 接口获取（字段 greeting_text）
- 前端 fallback：根据时间段生成默认问候语（保持现有逻辑）
- 后端有配置时优先使用后端文案

**验收标准：**
- [ ] 后端配置了 greeting_text 时显示后端文案
- [ ] 后端未配置时使用现有时间段问候逻辑
- [ ] 文案支持中文

**自测：**
1. 后端配置 greeting_text = "欢迎使用AI创作" → 页面显示该文案
2. 后端未配置 → 显示「下午好，xxx，准备好创作了吗？」
3. 后端返回空字符串 → 使用默认逻辑

### 6.3 游客积分用完引导

**涉及文件：** web/src/components/chat/ChatEmptyState.vue（或新建组件）

**任务描述：**
- 当游客积分为 0 时，在 Chat 页面显示引导界面
- 显示：图标 + 「体验积分已用完」+ 说明文字 + 注册/登录按钮
- 注册按钮跳转 /login（当前 register 已 redirect 到 login）
- 登录按钮跳转 /login

**验收标准：**
- [ ] 游客积分为 0 时显示引导界面
- [ ] 显示注册和登录按钮
- [ ] 点击按钮跳转到登录页
- [ ] 有积分时不显示引导界面

**自测：**
1. 游客积分 > 0 → 正常显示创作界面
2. 游客积分 = 0 → 显示用完引导
3. 点击注册 → 跳转 /login
4. 点击登录 → 跳转 /login

---

## 功能 7：经典版回新版入口

### 7.1 添加切换新版按钮

**涉及文件：** web/src/views/Home.vue

**任务描述：**
- 在经典版页面顶部区域添加「切换新版」按钮
- 样式：order border-teal/30 bg-teal/5 px-3 py-1.5 text-sm font-medium text-teal rounded-lg
- 带闪电图标 SVG
- 使用 <RouterLink to="/">
- 放置位置：页面顶部右侧（与现有布局协调）

**验收标准：**
- [ ] 经典版页面顶部右侧显示「切换新版」按钮
- [ ] 按钮带闪电图标
- [ ] 点击跳转到 /（新版 Chat 页面）
- [ ] 按钮样式为 teal 色调
- [ ] 不影响经典版现有布局和功能

**自测：**
1. 访问 /classic → 顶部右侧有「切换新版」按钮
2. 点击按钮 → 跳转到 / (Chat 页面)
3. 经典版其他功能正常（生成、风格选择等）
4. 移动端按钮正常显示

---

## 功能 8：经典版全屏显示修复

### 8.1 排查并修复宽度问题

**涉及文件：** web/src/views/Home.vue, web/index.html

**任务描述：**
- 检查 Home.vue 根容器是否有 max-width 或固定宽度限制
- 确保根容器使用 w-full 或 w-screen + min-h-screen
- 检查 index.html 的 <meta name="viewport"> 是否为标准配置：width=device-width, initial-scale=1.0
- 排查是否有 CSS media query 错误应用移动端样式
- 仅修复布局撑满问题，不改动经典版的视觉风格和交互

**验收标准：**
- [ ] 经典版页面在 Edge 浏览器中撑满整个窗口
- [ ] 无左右大面积留白
- [ ] Chrome/Firefox/Edge 表现一致
- [ ] 移动端响应式正常
- [ ] 经典版原有风格和交互不变

**自测：**
1. Edge 打开 /classic → 页面撑满浏览器窗口
2. Chrome 打开 → 同样撑满
3. 缩放浏览器窗口 → 响应式正常
4. 经典版所有功能正常（生成、样式选择、全屏查看等）
5. F12 检查无 max-width 限制根容器


---

## 开发顺序建议

按依赖关系和风险排序：

| 阶段 | 任务 | 预估工时 | 依赖 |
|------|------|----------|------|
| P0 | 8.1 经典版全屏修复 | 0.5h | 无 |
| P0 | 7.1 经典版回新版入口 | 0.5h | 无 |
| P1 | 1.1 新建按钮改造 | 0.5h | 无 |
| P1 | 5.1 侧边栏状态管理 | 0.5h | 无 |
| P1 | 5.2 展开态改造 | 1h | 5.1 |
| P1 | 5.3 收起态实现 | 1.5h | 5.1 |
| P1 | 5.4 布局适配 | 0.5h | 5.2, 5.3 |
| P2 | 2.1 自动命名 | 0.5h | 无 |
| P2 | 2.2 右键菜单 | 1h | 无 |
| P2 | 2.3 重命名编辑态 | 1h | 2.2 |
| P2 | 2.4 删除对话 | 0.5h | 2.2 |
| P3 | 3.1 移除自动分层提示 | 0.5h | 无 |
| P3 | 3.2 风格下拉 | 1.5h | 无 |
| P3 | 3.3 比例下拉 | 1h | 3.2（复用组件） |
| P3 | 3.4 分层开关优化 | 0.5h | 无 |
| P3 | 3.5 隐藏质量按钮 | 0.5h | 无 |
| P4 | 4.1 附图按钮交互 | 0.5h | 无 |
| P4 | 4.2 图片上传区域 | 1.5h | 4.1 |
| P5 | 6.1 游客积分显示 | 1h | 无 |
| P5 | 6.2 欢迎标语配置 | 0.5h | 无 |
| P5 | 6.3 积分用完引导 | 0.5h | 6.1 |

**总预估：约 15h**

---

## 注意事项

1. **不改经典版风格**：功能 8 仅修复全屏布局问题，不修改 Home.vue 的视觉设计、配色、交互逻辑
2. **后台配置优先**：风格、比例、游客积分、欢迎标语均优先使用后端配置，前端保留 fallback
3. **组件复用**：风格下拉和比例下拉可抽取为通用的 DropdownSelect 组件
4. **localStorage 持久化**：侧边栏状态需持久化，避免刷新丢失
5. **fingerprint 识别**：游客积分通过已有的 X-Fingerprint header 识别身份
6. **API 兼容**：对话重命名/删除如后端 API 已就绪则调用，否则仅本地操作
7. **过渡动画**：侧边栏收起/展开需要平滑过渡，避免闪烁

---

## 后端配置项汇总（需确认是否已有）

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| guest_free_credits | 游客默认积分 | 5 |
| greeting_text | Chat 页面欢迎标语 | 空（使用前端默认） |
| style_presets | 风格选项列表 | 写实/动漫/幻想/赛博朋克/水彩/插画 |
| size_options | 比例选项列表 | square/portrait_3_4/story/landscape_4_3/widescreen |
| default_quality | 默认质量 | medium |
| default_layer_count | 默认分层数 | 5 |

---

*文档版本：v2 | 更新时间：2026-05-15*
