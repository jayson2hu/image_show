# 用户个人中心架构与 UI 开发文档

日期：2026-05-07  
状态：待评审  
关联产品方案：`docs/plan-user-account-center.md`  
目标：按架构边界和 UI 设计规范拆分个人中心开发任务，做到小功能完成即自测验收，大功能完成即整体验收并提交代码。

## 1. 架构目标

新增用户个人中心能力，但不把用户侧能力和后台管理能力混在一起。

目标：

- 给登录用户提供统一的账户、积分、历史、通知入口。
- 复用现有模型和接口，减少一次性改动范围。
- 需要新增接口时，只新增用户侧 `/api/account/*`，不复用 `/api/admin/*`。
- 管理员也是用户，但用户中心只展示当前账号自己的用户侧数据。
- 前端形成清晰的信息层级，不做复杂设置页。

非目标：

- 不做订单列表。
- 不做修改邮箱。
- 不做修改密码。
- 不做头像文件上传。
- 不做登录设备管理。
- 不做收藏作品。
- 不调整微信登录主流程。

## 2. 当前系统基础

### 2.1 已有后端模型

可复用字段：

- `User`
  - `id`
  - `username`
  - `email`
  - `wechat_open_id`
  - `role`
  - `status`
  - `credits`
  - `credits_expiry`
  - `avatar_url`
  - `last_login_at`
  - `last_login_ip`
  - `created_at`
  - `updated_at`
- `CreditLog`
  - 用户积分变动记录。
- `Generation`
  - 用户生成历史。
- `Announcement` / `AnnouncementRead`
  - 用户公告与已读状态。
- `LoginLog`
  - 登录日志，后续可补最近登录方式。

### 2.2 已有用户侧接口

- `GET /api/auth/me`
- `GET /api/credits/balance`
- `GET /api/credits/logs`
- `GET /api/generations`
- `GET /api/generations/:id`
- `GET /api/announcements`
- `POST /api/announcements/:id/read`
- `GET /api/packages`

### 2.3 已有前端页面

- `/`
- `/login`
- `/register`
- `/history`
- `/credits`
- `/packages`
- `/console/admin`

## 3. 架构设计

### 3.1 推荐架构

采用“用户中心聚合接口 + 轻量资料更新接口”的方式。

新增后端：

- `controller/account.go`
- `controller/account_test.go`
- `service/account.go`，如逻辑较少可先不新增 service，保持在 controller 内部组合查询。
- `router/main.go` 增加：
  - `GET /api/account/overview`
  - `PUT /api/account/profile`

新增前端：

- `web/src/views/Account.vue`
- `web/src/router/index.ts` 增加 `/account`
- `web/src/stores/user.ts` 扩充用户字段。
- `web/src/App.vue` 顶部菜单增加个人中心入口。

### 3.2 为什么不只靠前端并发请求

可选方案：

1. 前端并发调用 `/auth/me`、`/credits/logs`、`/generations`、`/announcements`。
2. 后端新增 `/api/account/overview` 聚合接口。
3. 前端先并发，后端后续再聚合。

推荐方案：第 2 种。

理由：

- 个人中心是稳定入口，聚合接口能减少前端重复拼装。
- 权限边界集中在后端，更容易保证只返回当前用户数据。
- 后续加统计字段、最近登录方式、公告未读数时不需要频繁改前端数据组合。
- 失败处理更清晰，后端可对单个模块异常做降级。

保守策略：

- 第一版聚合接口可以只查必要数据。
- 最近作品只取 6 条。
- 最近积分流水只取 5 条。
- 公告只取前 5 条。
- 统计字段使用轻量 count，避免全量扫描。

### 3.3 数据流

用户打开 `/account`：

1. 前端路由守卫检查 `userStore.token`。
2. 若本地有 token 但没有 user，调用 `userStore.fetchUser()`。
3. 页面调用 `GET /api/account/overview`。
4. 后端通过 `AuthRequired` 得到当前 `userID`。
5. 后端并按当前 `userID` 查询：
   - 用户基础信息。
   - 最近积分流水。
   - 最近生成记录。
   - 生成统计。
   - 用户公告。
6. 前端渲染概览、积分、作品、通知、安全信息。

用户编辑资料：

1. 前端提交 `PUT /api/account/profile`。
2. 后端校验字段长度和 URL。
3. 后端只更新当前用户的 `username`、`avatar_url`。
4. 返回更新后的 user。
5. 前端刷新 `userStore.user` 和当前页面。

## 4. API 设计

### 4.1 `GET /api/account/overview`

权限：

- 必须登录。
- 只能返回当前用户数据。

返回：

```json
{
  "user": {
    "id": 1,
    "username": "jayson",
    "email": "user@example.com",
    "avatar_url": "",
    "role": 1,
    "status": 1,
    "credits": 10,
    "credits_expiry": null,
    "created_at": "2026-05-07T00:00:00+08:00",
    "last_login_at": "2026-05-07T10:00:00+08:00",
    "last_login_ip": "127.0.0.1"
  },
  "credits": {
    "recent_logs": [
      {
        "id": 1,
        "type": 5,
        "amount": 10,
        "balance": 10,
        "remark": "register gift",
        "created_at": "2026-05-07T00:00:00+08:00"
      }
    ]
  },
  "creations": {
    "total": 12,
    "completed": 10,
    "failed": 2,
    "latest_at": "2026-05-07T12:00:00+08:00",
    "recent_items": [
      {
        "id": 10,
        "prompt": "太空猫",
        "size": "1024x1024",
        "status": 3,
        "image_url": "https://...",
        "credits_cost": 1,
        "created_at": "2026-05-07T12:00:00+08:00"
      }
    ]
  },
  "announcements": {
    "unread_count": 1,
    "recent_items": []
  }
}
```

错误：

- 未登录：`401`
- 用户不存在：`401`
- 查询失败：`500`

### 4.2 `PUT /api/account/profile`

权限：

- 必须登录。
- 只能更新当前用户。

请求：

```json
{
  "username": "新的昵称",
  "avatar_url": "https://example.com/avatar.png"
}
```

校验：

- `username` 最长 64。
- `avatar_url` 最长 512。
- `avatar_url` 非空时必须是 `http://` 或 `https://`。
- 字段缺失时按空字符串处理，不允许更新 `email`、`role`、`status`、`credits`。

返回：

```json
{
  "user": {
    "id": 1,
    "username": "新的昵称",
    "email": "user@example.com",
    "avatar_url": "https://example.com/avatar.png",
    "role": 1,
    "status": 1,
    "credits": 10,
    "credits_expiry": null,
    "created_at": "2026-05-07T00:00:00+08:00",
    "updated_at": "2026-05-07T12:00:00+08:00"
  }
}
```

错误：

- 未登录：`401`
- URL 非法：`400`
- 字段过长：`400`
- 保存失败：`500`

## 5. UI 设计方案

### 5.1 页面整体风格

个人中心应比后台轻，比生成页更偏信息管理。

视觉原则：

- 页面背景使用浅灰，不做复杂渐变。
- 不堆叠卡片，模块之间用清晰留白和浅边框区分。
- 重点突出积分余额、有效期、最近作品。
- 操作按钮数量控制在每个模块 1 到 2 个。
- 移动端优先单列，桌面端两列或三列。

### 5.2 页面布局

桌面端建议结构：

1. 顶部账户横幅
   - 左侧头像、昵称、邮箱、身份 badge。
   - 右侧积分余额、有效期、最近登录。
2. 快捷操作栏
   - 去生成。
   - 查看历史。
   - 购买积分。
   - 查看流水。
3. 内容区双列
   - 左侧：积分与权益、最近流水。
   - 右侧：最近作品。
4. 底部双列
   - 个人资料。
   - 安全与通知。

移动端建议：

- 顶部账户横幅单列。
- 快捷操作使用 2 列按钮。
- 最近作品横向滚动或 2 列网格。
- 表格类信息改为列表。

### 5.3 组件拆分

第一版可以先写在 `Account.vue` 内，后续再拆组件。若文件超过 500 行，再拆为：

- `AccountHero.vue`
- `AccountMetricCard.vue`
- `AccountCreditSummary.vue`
- `AccountRecentCreations.vue`
- `AccountProfileForm.vue`
- `AccountSecurityNotice.vue`

建议第一阶段：

- 保持 `Account.vue` 单文件，避免提前抽象。
- 复用现有 Tailwind 风格和顶部菜单样式。
- 不引入新 UI 库。

### 5.4 空状态与错误态

必须覆盖：

- 没有头像。
- 没有昵称。
- 没有积分流水。
- 没有历史作品。
- 没有公告。
- `credits_expiry` 为空。
- overview 加载失败。
- profile 保存失败。

用户文案方向：

- 无作品：“还没有作品，去生成第一张图片。”
- 无流水：“暂无积分变动，注册赠送或生成图片后会出现在这里。”
- 无公告：“暂无新的通知。”
- 加载失败：“个人中心加载失败，请稍后重试。”

### 5.5 顶部菜单调整

当前顶部登录态菜单已有：

- 历史记录
- 积分流水
- 积分套餐
- 退出登录

新增后建议：

- 个人中心
- 历史记录
- 积分流水
- 积分套餐
- 退出登录

规则：

- 管理员也显示个人中心。
- 管理后台入口不放回右上角公开展示，仍通过 `/console/admin`。
- 退出登录保持点击菜单交互，不再使用 hover 展开。

## 6. 开发拆分与验收流程

开发时必须遵循：

- 每完成一个小功能，先做小功能自测验收。
- 每完成一个大功能，做整体自测验收。
- 每个大功能完成后提交一次代码。
- 每次完成小功能和发现问题，都更新 `docs/progress.md`。
- 如果出现产品、架构、权限、UI 冲突，先记录到 `docs/progress.md`，并询问确认，不自行决定。

## 7. 大功能 1：个人中心只读概览

目标：

- 登录用户能进入 `/account`，集中看到账户信息、积分、最近流水、最近作品、公告摘要。

### 小功能 1.1：新增路由和顶部入口

改动范围：

- `web/src/router/index.ts`
- `web/src/App.vue`
- 新增 `web/src/views/Account.vue` 占位页面

实现要求：

- 新增 `/account` 路由。
- 未登录访问 `/account` 跳转 `/login`。
- 顶部用户菜单增加“个人中心”。
- 点击进入后不影响管理员后台登录状态。

自测：

- 游客访问 `/account`，跳转登录。
- 普通用户登录后菜单显示“个人中心”。
- 管理员登录后菜单显示“个人中心”，访问 `/account` 不自动跳 `/console/admin`。
- 点击退出后无法访问 `/account`。

验收标准：

- 路由和入口可用。
- 未登录保护生效。
- 没有重复登录/注册入口。

进度记录：

- 在 `docs/progress.md` 增加“小功能 1.1 完成”记录。

### 小功能 1.2：后端概览接口

改动范围：

- `router/main.go`
- 新增 `controller/account.go`
- 新增 `controller/account_test.go`

实现要求：

- 新增 `GET /api/account/overview`。
- 使用 `middleware.AuthRequired()`。
- 返回当前用户基础信息。
- 返回最近 5 条积分流水。
- 返回最近 6 条生成记录。
- 返回生成统计：总数、完成数、失败数、最近生成时间。
- 返回公告未读数和最近公告。
- 不返回 `password_hash`、微信 token、后台设置、渠道信息。

自测：

- 未登录请求返回 401。
- 登录请求返回自己的用户信息。
- 用户 A 看不到用户 B 的历史和流水。
- 无数据时返回空数组和 0，不返回 null 导致前端崩溃。

测试命令：

- `go test ./controller -run "TestAccount" -v`
- `go test ./router ./controller -run "Test"`.

验收标准：

- 接口返回结构稳定。
- 权限隔离正确。
- 测试覆盖未登录、正常用户、跨用户隔离、空数据。

进度记录：

- 在 `docs/progress.md` 增加“小功能 1.2 完成”记录。

### 小功能 1.3：个人中心页面 UI 骨架

改动范围：

- `web/src/views/Account.vue`
- `web/src/stores/user.ts`

实现要求：

- 页面加载调用 `/account/overview`。
- 顶部展示头像、昵称、邮箱、身份、注册时间。
- 指标展示积分余额、有效期、最近登录。
- 增加快捷操作：去生成、历史记录、积分流水、积分套餐。
- 兜底展示：
  - 无昵称显示邮箱前缀。
  - 无头像显示首字母头像。
  - 无有效期显示“暂无到期时间”。
  - 无最近登录显示“暂无记录”。

自测：

- 有昵称、无昵称都显示正常。
- 有头像、无头像都显示正常。
- 页面刷新后仍能加载。
- 移动端不横向溢出。

测试命令：

- `pnpm.cmd exec vue-tsc --noEmit`

验收标准：

- UI 信息层级清楚。
- 不出现空白字段。
- 不出现文本溢出或按钮挤压。

进度记录：

- 在 `docs/progress.md` 增加“小功能 1.3 完成”记录。

### 小功能 1.4：最近流水与最近作品

改动范围：

- `web/src/views/Account.vue`

实现要求：

- 展示最近 5 条积分流水。
- 展示最近 6 张作品。
- 最近作品可点击进入 `/history` 或打开历史详情，第一阶段建议跳 `/history`。
- 无流水、无作品显示空状态和行动按钮。
- 失败作品展示友好状态，不直接显示上游错误原文。

自测：

- 无数据用户显示空状态。
- 有数据用户显示缩略图和流水。
- 点击“查看全部历史”进入 `/history`。
- 点击“查看全部流水”进入 `/credits`。

测试命令：

- `pnpm.cmd exec vue-tsc --noEmit`

验收标准：

- 最近资产入口完整。
- 空状态有下一步动作。

进度记录：

- 在 `docs/progress.md` 增加“小功能 1.4 完成”记录。

### 大功能 1 整体验收

自动自测：

- `go test ./controller -run "TestAccount" -v`
- `go test ./router ./controller -run "Test"`
- `pnpm.cmd exec vue-tsc --noEmit`

手动验收：

- 游客访问 `/account`。
- 普通用户访问 `/account`。
- 管理员访问 `/account`。
- 无历史用户。
- 有历史用户。
- 无积分流水用户。
- 有积分流水用户。

提交：

- 提交信息：`feat: add account overview`
- 推送到当前分支。

## 8. 大功能 2：个人资料编辑

目标：

- 用户能维护昵称和头像 URL，保存后顶部菜单与个人中心同步更新。

### 小功能 2.1：后端资料更新接口

改动范围：

- `router/main.go`
- `controller/account.go`
- `controller/account_test.go`

实现要求：

- 新增 `PUT /api/account/profile`。
- 只允许更新当前用户。
- 支持 `username` 和 `avatar_url`。
- 校验字段长度。
- `avatar_url` 非空时必须是 `http://` 或 `https://`。
- 返回更新后的用户。

自测：

- 未登录返回 401。
- 正常更新昵称成功。
- 正常更新头像 URL 成功。
- 非法 URL 返回 400。
- 超长字段返回 400。
- 请求中带 `role/status/credits/email` 不应生效。

测试命令：

- `go test ./controller -run "TestAccountProfile" -v`
- `go test ./router ./controller -run "Test"`

验收标准：

- 资料更新权限正确。
- 不允许越权改敏感字段。

进度记录：

- 在 `docs/progress.md` 增加“小功能 2.1 完成”记录。

### 小功能 2.2：前端资料编辑 UI

改动范围：

- `web/src/views/Account.vue`
- `web/src/stores/user.ts`

实现要求：

- 个人资料模块增加编辑按钮。
- 支持昵称输入。
- 支持头像 URL 输入和预览。
- 保存时禁用按钮并展示加载态。
- 保存成功后：
  - 更新页面 overview。
  - 更新 `userStore.user`。
  - 顶部菜单同步。
- 保存失败展示友好错误。

UI 要求：

- 表单不要占据整页，建议使用右侧内联编辑区或弹窗。
- 移动端编辑表单单列。
- 头像预览失败时回退首字母头像。

自测：

- 修改昵称后保存成功。
- 修改头像 URL 后预览成功。
- 清空昵称后保存，页面使用邮箱前缀兜底。
- 输入非法 URL，前端提示或后端错误能被展示。
- 保存过程中重复点击不会发多次请求。

测试命令：

- `pnpm.cmd exec vue-tsc --noEmit`

验收标准：

- 用户可以维护基础资料。
- 顶部菜单同步更新。
- 错误提示明确。

进度记录：

- 在 `docs/progress.md` 增加“小功能 2.2 完成”记录。

### 大功能 2 整体验收

自动自测：

- `go test ./controller -run "TestAccount" -v`
- `go test ./router ./controller -run "Test"`
- `pnpm.cmd exec vue-tsc --noEmit`

手动验收：

- 普通用户编辑昵称。
- 普通用户编辑头像。
- 管理员编辑自己的昵称，不影响角色。
- 非法头像 URL。
- 退出后重新登录，资料仍然存在。

提交：

- 提交信息：`feat: allow users to update profile`
- 推送到当前分支。

## 9. 大功能 3：安全与通知摘要

目标：

- 用户能看到最近登录信息、登录方式和公告未读状态。

### 小功能 3.1：最近登录方式

改动范围：

- `controller/account.go`
- `controller/account_test.go`
- `web/src/views/Account.vue`

实现要求：

- overview 增加 `security.latest_login`。
- 查询最近一条成功的 `LoginLog`。
- 展示：
  - 登录时间。
  - IP。
  - 登录方式：邮箱登录 / 微信验证码。
- 无记录显示“暂无登录记录”。

自测：

- 邮箱登录后显示邮箱登录。
- 微信验证码登录后显示微信验证码。
- 无登录日志用户显示空状态。

测试命令：

- `go test ./controller -run "TestAccountSecurity" -v`
- `pnpm.cmd exec vue-tsc --noEmit`

验收标准：

- 登录方式展示正确。
- 不展示完整 User-Agent，避免信息过长和隐私压力。

进度记录：

- 在 `docs/progress.md` 增加“小功能 3.1 完成”记录。

### 小功能 3.2：公告未读摘要

改动范围：

- `controller/account.go`
- `controller/account_test.go`
- `web/src/views/Account.vue`
- 可复用 `web/src/components/AnnouncementCenter.vue`

实现要求：

- overview 返回未读公告数量。
- 返回最近公告列表。
- 个人中心显示公告摘要。
- 点击“查看公告”打开现有公告中心或跳回页面顶部公告入口。

自测：

- 无公告显示空状态。
- 有公告显示标题。
- 未读数量正确。
- 标记已读后刷新个人中心数量减少。

测试命令：

- `go test ./controller -run "TestAccountAnnouncements" -v`
- `pnpm.cmd exec vue-tsc --noEmit`

验收标准：

- 用户能明确知道有没有新通知。
- 不破坏现有公告弹窗和公告中心。

进度记录：

- 在 `docs/progress.md` 增加“小功能 3.2 完成”记录。

### 大功能 3 整体验收

自动自测：

- `go test ./controller -run "TestAccount" -v`
- `go test ./router ./controller -run "Test"`
- `pnpm.cmd exec vue-tsc --noEmit`

手动验收：

- 邮箱登录用户。
- 微信验证码登录用户。
- 无公告。
- 有未读公告。
- 标记公告已读后刷新。

提交：

- 提交信息：`feat: add account security summary`
- 推送到当前分支。

## 10. 质量门禁

每个小功能完成必须：

- 更新 `docs/progress.md`。
- 记录自测命令和结果。
- 如果发现问题，记录“问题记录”。
- 如果涉及 UI，至少检查桌面端和移动端布局。

每个大功能完成必须：

- 跑对应后端测试。
- 跑前端类型检查。
- 做手动验收。
- 提交代码。
- 推送远端。

推荐命令：

```powershell
go test ./controller -run "TestAccount" -v
go test ./router ./controller -run "Test"
pnpm.cmd exec vue-tsc --noEmit
git status --short
git add <changed-files>
git commit -m "<message>"
git push
```

## 11. 风险与处理

### 11.1 聚合接口查询变慢

风险：

- 个人中心聚合多个模块，后续数据量大时可能变慢。

处理：

- 最近数据限制数量。
- 统计只做必要 count。
- 后续如慢查询明显，再增加索引或缓存。

### 11.2 用户字段和顶部菜单不同步

风险：

- 编辑资料后页面显示已更新，但顶部菜单仍旧。

处理：

- 保存成功后同时更新 `userStore.user`。
- 再调用 `userStore.fetchUser()` 兜底。

### 11.3 管理员和普通用户混淆

风险：

- 管理员进入用户中心后误以为是后台。

处理：

- 个人中心只显示“管理员”身份 badge。
- 不在用户中心顶部强展示后台入口。
- 后台仍使用 `/console/admin`。

### 11.4 敏感信息泄露

风险：

- 聚合接口误返回后端敏感配置、微信 token、渠道 key。

处理：

- 只手动构造响应 DTO。
- 不直接返回完整 model 中不需要的字段。
- 测试中断言响应不包含 `password_hash`、`wechat_server_token`、`api_key`。

## 12. 需要确认的问题

以下问题开发前建议确认：

1. 头像第一阶段是否只支持 URL？
   - 建议：是，只支持 URL。
2. 管理员个人中心是否显示“进入后台”的低优先级入口？
   - 建议：暂不显示，避免普通用户侧入口变复杂。
3. 最近作品点击后是跳历史页还是打开详情弹窗？
   - 建议：第一阶段跳历史页，后续再做详情复用。
4. 是否第一阶段展示订单入口？
   - 建议：不展示，先用积分流水解释资产变动。
5. 昵称是否允许重复？
   - 建议：允许，邮箱仍是唯一账号标识。

## 13. 推荐开发顺序

优先级：

1. 大功能 1：个人中心只读概览。
2. 大功能 2：个人资料编辑。
3. 大功能 3：安全与通知摘要。

原因：

- 只读概览能最快让用户看到完整个人资产。
- 资料编辑建立用户维护能力。
- 安全与通知在前两者稳定后补齐可信感。

## 14. 文档更新要求

开发中必须同步维护：

- `docs/progress.md`
  - 每个小功能完成记录。
  - 每次自测记录。
  - 每个问题记录。
- `docs/dev-plan-user-account-center.md`
  - 如果实际开发方案与本文档不同，必须更新对应章节。
- `docs/plan-user-account-center.md`
  - 如果产品范围变化，必须回写产品方案。

## 15. 最终验收清单

- 登录用户能进入 `/account`。
- 游客访问 `/account` 被引导登录。
- 个人中心展示账户、积分、最近流水、最近作品、通知、安全摘要。
- 用户能编辑昵称和头像 URL。
- 顶部菜单显示个人中心入口。
- 普通用户和管理员身份展示不混淆。
- 用户只能看到自己的数据。
- 接口不返回敏感字段。
- 移动端布局可用。
- 前端类型检查通过。
- 后端测试通过。
- 文档已更新。
- 大功能提交已推送。
