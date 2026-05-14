# 需求清理与验收总结

日期：2026-05-14  
结论：当前主线需求已完成自动化验收；剩余工作主要是外部服务联调、生产配置验收和少量后续增强，不属于当前未开发阻塞项。

## 1. 文档口径

当前可信进度以本文件、`docs/progress.md`、`docs/tasks-acceptance.md`、`docs/plan-admin-site-account-ops.md`、`docs/plan-user-account-center.md` 和 `docs/dev-plan-user-account-center.md` 为准。

`docs/pending-tasks-review.html` 是 2026-05-12 的旧待办快照，里面的“待开发”状态已被后续开发和复验覆盖，只保留为历史归档，不再作为当前开发依据。

`development-plan.md` 是早期总计划，文件中大量 `[ ]` 属于原始验收清单；后续同一文件底部进度表和 `docs/progress.md` 已记录完成状态，因此不能直接把早期 `[ ]` 视为当前未完成。

## 2. 已开发并验证的需求

| 模块 | 需求 | 状态 | 验证依据 |
| --- | --- | --- | --- |
| 后台重构 | `/console/admin` 切换到 `AdminLayout`，覆盖概览、用户、渠道、模板、设置、公告、积分、监控、套餐 | 已开发 | `web/src/views/admin/AdminDashboard.vue`、`web/src/components/admin/AdminLayout.vue` |
| 站点与 SEO | 后台配置站点标题、关于、SEO、注册策略；前台读取 `/api/site/config` | 已开发 | `controller/site.go`、`web/src/api/site.ts`、`web/src/App.vue` |
| 注册策略 | 注册开关、邮箱后缀白名单、后台创建用户不受普通注册策略影响 | 已开发 | `service/auth.go`、`controller/auth.go`、`controller/auth_test.go` |
| 套餐管理 | 默认套餐、公开套餐列表、后台套餐 CRUD、前台购买中心展示 | 已开发 | `controller/package.go`、`web/src/components/admin/PackagesTab.vue`、`web/src/views/Packages.vue` |
| 人工充值 | 后台配置微信号、二维码、QQ、说明；前台 `/support/contact` 展示 | 已开发 | `controller/admin_template_setting.go`、`web/src/views/Packages.vue` |
| 渠道文案 | 后台图片生成渠道不再把所有渠道统称为 Sub2API | 已开发 | `web/src/components/admin/ChannelsTab.vue` |
| 账号中心 | `/account`、账号概览、积分摘要、最近作品、公告、安全摘要 | 已开发 | `router/main.go`、`controller/account.go`、`web/src/views/Account.vue` |
| 资料编辑 | 昵称、头像 URL 更新，敏感字段不允许越权更新 | 已开发 | `controller/account.go`、`controller/account_test.go` |
| 头像上传 | `POST /api/account/avatar`，本地 `uploads/avatars` 存储，类型/大小校验，后台头像存储配置 | 已开发 | `controller/account.go`、`web/src/views/Account.vue`、`controller/account_test.go` |
| 账号信息架构 | `/account` 默认聚焦个人信息，最近作品限制为 3 个，历史/积分/购买中心提供返回入口 | 已开发 | `controller/account.go`、`web/src/views/Account.vue` |
| 固定生成参数 | 固定 medium、PNG、opaque；按比例读取后台积分定价 | 已开发 | `controller/generation.go`、`service/credit.go`、`web/src/views/Home.vue` |
| 场景入口 | 6 个场景、`/api/generation/scenes`、`SceneCard`、后台模板支持 scene 元数据 | 已开发 | `controller/prompt_template.go`、`web/src/components/SceneCard.vue`、`web/src/views/Home.vue` |
| 历史再次生成 | 历史卡片“再次生成”，带 prompt/ratio 回填首页，不自动触发生成 | 已开发 | `web/src/views/History.vue`、`web/src/views/Home.vue` |
| 登录页重构 | 微信优先、邮箱登录/注册折叠入口、`/register` 跳转 `/login` | 已开发 | `web/src/views/Login.vue`、`web/src/router/index.ts` |

## 3. 未开发或未纳入当前范围

| 需求 | 当前结论 | 说明 |
| --- | --- | --- |
| 请求签名强制校验 | 未开发 | `development-plan.md` 中标为可选后续项，当前未启用，避免破坏前端和第三方回调兼容性。 |
| 完整登录设备列表 | 未开发 | 当前只展示最近一次成功登录。 |
| 公告详情页 | 未开发 | 当前复用现有公告中心/公告摘要。 |
| 用户侧完整订单列表 | 未开发 | 当前以套餐和积分流水解释资产变化，订单列表属后续用户资产增强。 |
| 修改邮箱/修改密码 | 未开发 | 涉及验证码、安全校验和风控，仍属于后续账号安全阶段。 |
| 微信绑定/解绑管理页 | 未纳入本轮 | 后端已有微信绑定能力，个人中心完整管理入口未作为本轮交付重点。 |
| 收藏作品、用户偏好、注销账号 | 未开发 | 属于后续资产库/账号设置能力。 |
| 历史全文索引 | 未开发 | 当前使用轻量 SQL `LIKE`，数据量变大后再升级。 |
| 历史错误后端统一分类 | 未开发 | 当前错误友好分类主要在前端实现。 |

## 4. 已开发但仍需外部/人工验收

| 项目 | 当前状态 | 需要的验收条件 |
| --- | --- | --- |
| Cloudflare R2 真实上传、删除、迁移 | 代码路径和单元测试已覆盖 | 配置真实 R2 凭据后，验证上传、Presigned URL、删除、free 到 paid 迁移。 |
| Redis 限流真实路径 | 代码支持 Redis 和本地内存降级 | 接入真实 Redis 后做限流联调。 |
| SMTP 告警邮件 | 代码支持未配置时跳过 | 配置 SMTP 后确认管理员可收到告警。 |
| Turnstile | 代码和测试已覆盖基本路径 | 配置真实 site key/secret 后做浏览器验证。 |
| 支付回调 | 代码和测试覆盖易支付模式 | 生产支付参数和回调地址配置后做真实沙箱/小额验收。 |
| 微信登录 | 代码支持验证码模式 | 配置真实 WeChat Server 后验证二维码、回调、轮询、绑定/解绑。 |
| 浏览器 UI | 自动化构建已通过 | 仍建议手动检查 `/console/admin` 8 个 Tab、`/account`、`/login`、`/history`、`/packages` 的桌面端和移动端表现。 |

## 5. 本次验证命令

```powershell
go test ./controller -run "TestAccount|TestRegister|TestAdminPromptTemplateCRUDAndSettings|Test.*Package|TestGenerationScenes|TestSiteConfig" -v
go test ./service -v
go test ./...
cd web; pnpm.cmd exec vue-tsc --noEmit
cd web; pnpm.cmd build
```

结果：全部通过。

## 6. 当前建议

当前不建议继续按旧 `pending-tasks-review.html` 补开发。下一步应优先做人工 UI 验收和生产外部服务联调；如果要继续开发，建议从“请求签名”“订单列表”“修改密码/邮箱”“登录设备列表”中重新立项，不要混入已完成主线。
