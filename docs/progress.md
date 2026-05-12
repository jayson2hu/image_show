# 开发进度记录

## 2026-05-13 P2-A/P2-B 场景入口与历史再次生成

- 开发目标：
  - 完成 `docs/tasks-acceptance.md` 中剩余 P2-A S-1 ~ S-4 和 P2-B R-1 ~ R-3。
- 完成：
  - `PromptTemplate` 增加场景元数据字段：`icon`、`recommended_ratio`、`description`，并保留 `category=scene` 作为场景模板分类。
  - 新增 `GET /api/generation/scenes`，默认返回 6 个场景卡片，包含图标、描述、提示词模板、推荐比例和积分。
  - 新增 `SceneCard.vue`，支持 hover、选中、点击反馈、比例 badge 和积分显示。
  - 首页接入场景网格，桌面 3 列、平板 2 列、手机横向 snap；点击场景会聚焦输入框、切换比例并打字机填充提示词。
  - “自由创作”场景只切换 1:1 并聚焦输入框，不填充提示词；再次点击已选场景会取消选择并恢复 1:1。
  - 后台模板管理支持场景分类，scene 类型可编辑图标、推荐比例和描述。
  - 历史页图片卡片 hover 显示“再次生成”悬浮按钮，跳转首页时通过 query 回填提示词和比例，不自动触发生成。
- 自测记录：
  - `go test ./controller -run "TestPromptTemplates|TestGenerationScenes|TestAdminPromptTemplateCRUDAndSettings" -v`：通过。
  - `go test ./...`：通过。
  - `cd web && pnpm exec vue-tsc --noEmit`：通过。
  - `cd web && pnpm build`：通过。
- 验收结论：
  - P2-A/P2-B 自动验收通过，剩余任务已完成，可以提交并推送。

## 2026-05-12 P1-B 登录注册重设计

- 开发目标：
  - 完成 `docs/tasks-acceptance.md` 中 P1-B 的 L-1 ~ L-5，复用现有微信、邮箱登录、邮箱注册 API。
- 完成：
  - 重写 `web/src/views/Login.vue` 为微信优先登录页，页面加载自动请求 `/api/auth/wechat/qrcode`，二维码区域带 skeleton 加载态。
  - 主卡片保留微信验证码输入和登录按钮，登录成功后调用 `fetchUser` 并跳转首页。
  - 邮箱入口改为折叠面板，支持“邮箱登录 / 邮箱注册”Tab，注册 Tab 可发送邮箱验证码。
  - 微信未启用或请求失败时自动展开邮箱面板，并显示“微信登录暂不可用”提示条。
  - `/register` 路由改为重定向 `/login`，`Register.vue` 仅保留兜底跳转，导航文案改为“登录”。
- 自测记录：
  - `cd web && pnpm exec vue-tsc --noEmit`：通过。
  - `cd web && pnpm build`：通过。
- 验收结论：
  - P1-B 自动验收通过，可以提交并推送。

## 2026-05-12 P1-A GPT Image 2 集成：固定输出参数与按比例计费

- 开发目标：
  - 完成 `docs/tasks-acceptance.md` 中 P1-A 的 G-1 ~ G-8：固定生成质量/格式/背景，按 5 种比例后台可配置扣费，并在首页和后台设置页展示。
- 完成：
  - 后端生成接口固定使用 `quality=medium`、`output_format=png`、`background=opaque`，忽略客户端传入的格式、背景和压缩参数。
  - `service.CostForRatio` / `CreditCostsByRatio` 支持从后台设置读取 5 种比例价格，默认方形 1 积分、其余 2 积分。
  - `/api/generation/options` 和 `/api/site/config` 返回配置后的积分价格；生成扣费与前端预估共用同一套设置。
  - 后台设置页“账号与额度”新增 5 个生成积分定价输入，并校验正整数。
  - 首页加载 `/api/site/config` 的 `credit_costs`，比例选择器、生成按钮和计费提示实时展示预估积分。
- 自测记录：
  - `go build ./...`：通过。
  - `go test ./controller -run "TestGeneration|TestSiteConfig|TestAdminPromptTemplateCRUDAndSettings" -v`：通过。
  - `go test ./service -v`：通过。
  - `cd web && pnpm exec vue-tsc --noEmit`：通过。
  - `cd web && pnpm build`：通过。
- 验收结论：
  - P1-A 自动验收通过，可以提交并推送。

## 2026-05-12 P0 管理后台 D3 最终验证

- 开发目标：
  - 根据 `docs/tasks-acceptance.md` 完成 P0 管理后台 D3 最终验证与文档收尾。
- 完成：
  - 确认 `/console/admin` 路由入口加载 `web/src/views/admin/AdminDashboard.vue`，该入口已是 `AdminLayout` 薄壳。
  - 更新 `docs/plan-admin-dashboard-redesign.md`，将整体状态、D3 验证项、回归矩阵和验收清单标记为已完成。
  - 更新 `docs/tasks-acceptance.md`，将 D3-2、D3-4 和 P0 验收项标记为已完成。
- 自测记录：
  - `cd web && pnpm exec vue-tsc --noEmit`：通过。
  - `cd web && pnpm build`：通过，构建产物无错误。
- 验收结论：
  - P0 管理后台 D3 最终验证通过，已通过本地代理完成 GitHub 推送。

## 2026-05-08 管理后台站点配置与个人中心体验优化计划

- 需求来源：
  - 管理后台需要增加全局 SEO、网站标题、关于网站、注册开关、邮箱后缀限制、充值套餐和充值渠道配置。
  - 图片生成渠道管理文案不能只绑定 sub2api。
  - 个人中心需要支持头像本地上传、默认只展示个人信息、最近作品只展示 3 个、历史/流水/购买中心可返回上一层，并修复购买中心初始黑色状态。
- 完成：
  - 新增 `docs/plan-admin-site-account-ops.md`。
  - 按大功能拆分为 A 管理后台站点配置、B 积分充值套餐与充值渠道、C 渠道文案修正、D 个人资料与头像上传、E 个人中心信息架构优化。
  - 每个小功能已写明开发内容、自测命令、验收标准、进度记录和建议提交信息。
- 待确认：
  - 无。
- 决策确认：
  - 充值渠道第一版只做“人工充值/联系管理员”，后台可配置微信号、微信二维码、QQ 和说明文案，暂不接真实支付。
  - 头像上传第一版不走 R2，使用后台可配置的本地头像存储；后续如启用 R2，需要后台提供一键上传全部用户头像并回写头像地址的迁移能力。
  - 管理后台创建用户不受普通注册开关和邮箱后缀限制影响；但如果填写邮箱，必须校验邮箱格式。
- 计划更新：
  - `docs/plan-admin-site-account-ops.md` 已将 B3 拆为人工充值联系方式配置，新增 B4 前台购买中心展示联系方式。
  - D 头像相关拆为 D1 头像上传后端接口、D2 后台头像存储配置 UI、D3 个人中心头像上传 UI、D4 头像 R2 迁移预留方案文档。
- 自测记录：
  - `Get-Content docs/plan-admin-site-account-ops.md`：已执行，可读取。

## 2026-05-08 C1 渠道管理文案去 Sub2API 化

- 开发目标：
  - 后台渠道管理页不再把所有图片生成渠道统称为 Sub2API，避免后续接入其他渠道时产生误导。
- 完成：
  - 将 `web/src/components/admin/ChannelsTab.vue` 的渠道页副标题从“维护 Sub2API 渠道、权重、状态和测试结果”改为“维护图片生成服务渠道、权重、状态和测试结果”。
  - 保留底层 `Sub2APIClient`、环境兜底渠道名和测试用例中的 sub2api 名称，不影响现有调用逻辑。
  - 更新 `docs/plan-admin-site-account-ops.md` 进度表，标记 C1 已完成。
- 自测记录：
  - `rg -n "Sub2API|sub2api|维护 Sub2API|维护 sub2api" web/src`：无匹配，说明前端展示文案已清理。
  - `pnpm.cmd exec vue-tsc --noEmit`：通过。
- 验收结论：
  - C1 局部自测通过，可以提交。

## 2026-05-08 A1 后端设置键和默认值

- 开发目标：
  - 后台设置接口补齐站点基础、SEO 和注册邮箱后缀设置，保证老数据库没有配置时也有默认值。
- 完成：
  - `GET /api/admin/settings` 新增返回 `site_title`、`site_about`、`seo_title`、`seo_keywords`、`seo_description`、`register_email_domain_allowlist`。
  - 新增 `adminSettingDefaults` 集中记录站点、SEO、注册、微信、验证、监控、积分、图片模型和 R2 的默认设置值。
  - 设置保存接口沿用现有 upsert 逻辑，老数据库无需额外迁移。
  - 管理员创建用户现有请求绑定已要求 `email` 格式，本小功能未改变管理员创建用户权限策略。
  - 更新 `docs/plan-admin-site-account-ops.md` 进度表，标记 A1 已完成。
- 自测记录：
  - `gofmt -w controller/admin_template_setting.go controller/admin_template_setting_test.go`：已执行。
  - `go test ./controller -run "TestAdminPromptTemplateCRUDAndSettings" -v`：通过。
  - `go test ./controller -run "TestAdminPromptTemplateCRUDAndSettings|TestRegister" -v`：通过。
- 验收结论：
  - A1 局部自测通过，可以提交。

## 2026-05-08 A2 后台站点与 SEO 设置 UI

- 开发目标：
  - 后台设置页新增清晰的“站点与 SEO”和“注册策略”分组，避免站点配置、注册配置继续混在账号额度里。
- 完成：
  - `web/src/components/admin/SettingsTab.vue` 默认打开“站点与 SEO”分组。
  - 新增“站点与 SEO”分组，包含网站标题、关于网站、SEO 标题、SEO 关键词、SEO 描述。
  - 新增“注册策略”分组，包含注册开关和允许注册邮箱后缀。
  - “账号与额度”分组只保留注册赠送积分和额度用完提示相关配置。
  - 注册邮箱后缀支持多行输入，并增加示例 placeholder。
  - 保存设置时只提交当前分组字段，降低误提交其他分组配置的风险。
  - 更新 `docs/plan-admin-site-account-ops.md` 进度表，标记 A2 已完成。
- 自测记录：
  - `pnpm.cmd exec vue-tsc --noEmit`：通过。
- 验收结论：
  - A2 局部自测通过，可以提交。

## 2026-05-08 A3 前台应用网站标题和 SEO

- 开发目标：
  - 前台读取后台配置的网站标题、关于网站和 SEO 配置，并避免暴露后台敏感设置。
- 完成：
  - 新增 `GET /api/site/config`，只返回 `site_title`、`site_about`、`seo_title`、`seo_keywords`、`seo_description`。
  - 新增后端测试，确认公开接口能返回站点配置，并且不会泄露 `wechat_server_token`。
  - 前端新增 `web/src/api/site.ts`。
  - `App.vue` 启动时读取站点配置，更新 `document.title`、`meta[name=description]`、`meta[name=keywords]`。
  - 顶部品牌标题和副标题改为读取后台站点配置。
  - 更新 `docs/plan-admin-site-account-ops.md` 进度表，标记 A3 已完成。
- 自测记录：
  - `gofmt -w controller/site.go controller/site_test.go router/main.go`：已执行。
  - `go test ./controller -run "TestSiteConfig|TestAdminPromptTemplateCRUDAndSettings" -v`：通过。
  - `pnpm.cmd exec vue-tsc --noEmit`：通过。
- 验收结论：
  - A3 局部自测通过，可以提交。

## 2026-05-08 A4 注册开关和邮箱后缀限制

- 开发目标：
  - 前台注册受后台注册开关和邮箱后缀 allowlist 控制；管理员后台创建用户不受普通注册策略限制，但邮箱格式必须合法。
- 完成：
  - `service.Register` 增加 `register_email_domain_allowlist` 校验。
  - 邮箱后缀 allowlist 支持英文逗号、换行、分号和空格分隔，配置项留空表示不限制。
  - 注册关闭时返回中文友好提示：当前暂未开放注册，请联系管理员。
  - 邮箱后缀不允许时返回中文友好提示：当前邮箱后缀暂不支持注册，请更换邮箱或联系管理员。
  - 后台创建用户保持现有权限逻辑，不读取普通注册策略；补充测试确认 allowlist 不影响后台创建用户。
  - 后台创建用户现有 `binding:"required,email"` 保持邮箱格式校验。
  - 更新 `docs/plan-admin-site-account-ops.md` 进度表，标记 A4 已完成。
- 自测记录：
  - `gofmt -w service/errors.go service/auth.go controller/auth.go controller/auth_test.go controller/admin_user_test.go`：已执行。
  - `go test ./controller -run "TestRegister|TestAdminCreateUserIgnoresRegistrationDomainAllowlist|TestAdminUserManagementAndCredits" -v`：通过。
  - `go test ./service ./controller -run "TestRegister|TestAdminCreateUserIgnoresRegistrationDomainAllowlist|TestAdminUserManagementAndCredits" -v`：通过。
- 验收结论：
  - A4 局部自测通过，可以提交。

## 2026-05-08 大功能 A：管理后台站点配置整体验收

- 范围：
  - A1 后端设置键和默认值。
  - A2 后台站点与 SEO 设置 UI。
  - A3 前台应用网站标题和 SEO。
  - A4 注册开关和邮箱后缀限制。
- 整体验收结果：
  - 后台设置接口可返回并保存站点、SEO、注册开关、邮箱后缀限制。
  - 后台设置页已拆出“站点与 SEO”和“注册策略”分组。
  - 前台通过公开配置接口读取站点配置，并动态更新 title/meta 和品牌文案。
  - 前台注册受开关和邮箱后缀限制，后台创建用户不受普通注册策略影响但校验邮箱格式。
- 自测记录：
  - `go test ./controller -run "TestAdminPromptTemplateCRUDAndSettings|TestSiteConfig|TestRegister|TestAdminCreateUserIgnoresRegistrationDomainAllowlist" -v`：通过。
  - `pnpm.cmd exec vue-tsc --noEmit`：通过。
- 验收结论：
  - 大功能 A 自动验收通过，可以提交整体验收记录并继续大功能 B。

## 2026-05-08 B1 套餐模型和默认套餐核查

- 开发目标：
  - 确保新部署没有套餐时，系统自动创建可用的默认积分套餐，个人中心购买中心后续能直接读取。
- 完成：
  - 核查现有套餐模型、公开 `/api/packages`、后台 `/api/admin/packages` CRUD 已存在。
  - 修正默认套餐初始化逻辑，默认创建 3 个启用套餐：
    - `Starter Pack`：10 积分，9.9，30 天。
    - `Standard Pack`：50 积分，39.9，90 天。
    - `Pro Pack`：100 积分，79.9，180 天。
  - 重写 `model/main.go` 为 ASCII 安全版本，保留原有 DB 初始化、迁移、默认管理员逻辑，避免历史乱码继续影响 Go 编译和补丁维护。
  - 更新套餐测试，断言默认套餐名称、积分和价格。
  - 更新 `docs/plan-admin-site-account-ops.md` 进度表，标记 B1 已完成。
- 自测记录：
  - `gofmt -w model/main.go controller/package_test.go`：已执行。
  - `go test ./controller -run "TestPackage|TestOrder" -v`：通过。
  - `go test ./model -v`：通过。
- 问题记录：
  - 原 `model/main.go` 中默认套餐中文名称已有历史乱码，且影响补丁匹配；本次改为英文 ASCII 套餐名保证稳定。后续如果需要前台显示中文，可在前端展示层或后台套餐管理中手动改名。
- 验收结论：
  - B1 局部自测通过，可以提交。

## 2026-05-08 B2 后台套餐管理 UI 优化

- 开发目标：
  - 管理后台提供明确的套餐管理入口，管理员可新增、编辑、启用/停用、删除积分套餐。
- 完成：
  - 新增 `web/src/components/admin/PackagesTab.vue`。
  - 后台侧边栏新增“套餐”入口。
  - 后台布局接入套餐 Tab。
  - 前端 admin API 增加 `fetchAdminPackages`、`createPackage`、`updatePackage`、`deletePackage`。
  - 前端类型增加 `CreditPackage`。
  - 套餐卡片展示价格、积分、有效期、排序、启用状态和约可生成标准图数量。
  - 新增/编辑使用弹窗表单，删除使用确认弹窗，操作结果使用 Toast。
  - 更新 `docs/plan-admin-site-account-ops.md` 进度表，标记 B2 已完成。
- 自测记录：
  - `pnpm.cmd exec vue-tsc --noEmit`：通过。
- 验收结论：
  - B2 局部自测通过，可以提交。
## 2026-05-07 渠道归因与渠道健康统计 1.1：生成记录渠道字段扩展

- 开发目标：
  - 在生成记录中预留渠道归因字段，为后续按渠道统计成功/失败做准备。
- 完成：
  - `model.Generation` 新增 `channel_id` 和 `channel_name`。
  - `channel_id` 可为空并建立索引，用于聚合统计。
  - `channel_name` 保存渠道名称快照，兼容渠道后续改名或删除。
  - 不回填历史数据，旧记录保持空渠道归因。
- 自测记录：
  - `go test ./model -v`：通过。
  - `go test ./...`：通过。
- 问题记录：
  - 暂无。

## 2026-05-07 渠道归因与渠道健康统计 1.2：生成渠道选择结果回传

- 开发目标：
  - 让多渠道调用链在返回图片结果时同步返回实际使用的渠道信息。
- 完成：
  - 新增 `service.ChannelUse`，包含 `ID` 和 `Name`。
  - `ImageGenerationResult` 增加 `Channel` 字段。
  - `GenerateImageViaChannels` 和 `EditImageViaChannels` 成功时写入实际成功渠道。
  - 环境变量兜底渠道名称统一为 `env:SUB2API_BASE_URL`。
  - Mock 模式返回 `mock` 渠道名称，便于测试环境识别。
- 自测记录：
  - `go test ./service -run "Test.*Channel|Test.*GenerateImage" -v`：通过。
  - `go test ./...`：通过。
- 问题记录：
  - 当前仅回传成功渠道；最终失败时的“最后尝试渠道”会在 1.3 写入逻辑中补齐。

## 2026-05-07 渠道归因与渠道健康统计 1.3：生成任务写入渠道归因

- 开发目标：
  - 成功和失败生成任务都写入实际渠道归因，供后台渠道健康统计使用。
- 完成：
  - 新增 `ChannelError`，在多渠道全部失败时携带最后尝试渠道。
  - `runGeneration` 成功调用上游后写入 `channel_id/channel_name`。
  - `runGeneration` 上游失败时写入最后尝试渠道。
  - `runImageEdit` 同步接入渠道归因写入。
  - 保存图片失败时保留此前已写入的上游渠道。
  - 已取消任务不强制写入渠道，避免取消后误更新。
- 自测记录：
  - `go test ./service -run "Test.*Generation|Test.*Channel" -v`：通过。
  - `go test ./controller -run "TestGeneration" -v`：通过。
  - `go test ./...`：通过。
- 问题记录：
  - 失败归因采用“最后尝试渠道”，不是完整重试链路；完整链路需后续新增尝试明细表。

## 2026-05-07 渠道归因与渠道健康统计 1.4：渠道近 24 小时统计接口

- 开发目标：
  - 后台渠道列表接口返回每个渠道近 24 小时成功、失败和失败率。
- 完成：
  - `/api/admin/channels` 返回 `recent_success_count`、`recent_failed_count`、`recent_failure_rate`。
  - 统计窗口为服务器当前时间向前 24 小时。
  - 只统计 `channel_id IS NOT NULL` 且 `status IN (3,4)` 的生成记录。
  - 无统计数据的渠道返回 0。
  - 24 小时外记录不计入。
- 自测记录：
  - `go test ./controller -run "TestAdminChannel" -v`：通过。
- 问题记录：
  - 环境变量兜底渠道不在当前渠道列表中展示，后续如需要可增加单独兜底统计卡片。

## 2026-05-07 渠道归因与渠道健康统计计划文档

- 需求：
  - 针对 Milestone 5.3 渠道近 24 小时成功/失败聚合缺少 `channel_id` 的问题，先输出详细开发计划和自测验收文档，拆分小功能后再开发。
- 完成：
  - 新增 `docs/plan-channel-attribution-and-health.md`。
  - 新增 `docs/acceptance-channel-attribution-and-health.md`。
  - 开发计划拆分为 5 个小功能：生成记录渠道字段扩展、生成渠道选择结果回传、生成任务写入渠道归因、渠道近 24 小时统计接口、后台渠道页展示统计。
  - 自测验收文档覆盖每个小功能的测试点、命令、通过标准，以及大功能整体 Go/No-Go 标准。
  - 明确暂不做历史回填和 `generation_channel_attempts` 明细表，先实现单字段归因版。
- 自测记录：
  - `Get-Content docs/plan-channel-attribution-and-health.md`：可读取。
  - `Get-Content docs/acceptance-channel-attribution-and-health.md`：可读取。
- 问题记录：
  - PowerShell 当前输出中文会乱码，但文档文件按 UTF-8 写入；后续如需要可统一检查文档编码。

## 2026-05-07 产品体验优化 Milestone 4.1：设置项分组梳理和映射

- 开发目标：
  - 按业务场景整理后台设置项，避免所有配置混在同一个长表单里。
- 完成：
  - 复核当前后台已存在设置分组：账号与额度、微信登录、图像生成、图片存储、人机验证、安全与监控、其他配置。
  - 未识别的设置项会进入“其他配置”，不丢失已有配置。
  - 后端设置存储结构保持不变，仍使用 `settings` 表按 key/value 保存。
- 自测记录：
  - `pnpm.cmd exec vue-tsc --noEmit`：通过。
- 问题记录：
  - 当前分组映射在 `AdminDashboard.vue` 内维护，后续如果设置项继续增多，可再拆成独立配置文件。

## 2026-05-07 产品体验优化 Milestone 4.2：设置页分组 UI

- 开发目标：
  - 后台设置页按分组导航展示，并支持单组保存。
- 完成：
  - 复核当前设置页已使用左侧分组导航、右侧字段详情的布局。
  - 本次修正保存逻辑：保存按钮只提交当前选中分组的设置项，不再一次性提交全部设置，降低误改其他组配置的风险。
- 自测记录：
  - `pnpm.cmd exec vue-tsc --noEmit`：通过。
- 问题记录：
  - 页面提示文案仍有历史编码显示问题，未在本小功能中扩大处理，避免引入大范围文本改动。

## 2026-05-07 产品体验优化 Milestone 4.3：敏感字段密文和眼睛按钮

- 开发目标：
  - API Key、Secret、Token 等敏感配置默认隐藏，并可临时查看。
- 完成：
  - 复核当前后台已通过 `isSecretSetting` 识别 `secret`、`password`、`access_key`、`*_token` 等字段。
  - 敏感字段默认使用 password 输入框，并提供眼睛按钮切换显示/隐藏。
  - 复核普通前台微信二维码接口不会返回 `wechat_server_token`、`wechat_server_address` 等敏感配置。
- 自测记录：
  - `go test ./controller -run "TestAdmin.*Setting|TestAdminPromptTemplateCRUDAndSettings" -v`：通过。
  - `pnpm.cmd exec vue-tsc --noEmit`：通过。

## 2026-05-07 产品体验优化 Milestone 4.4：敏感配置保存二次确认

- 开发目标：
  - 修改敏感字段并保存时，要求管理员二次确认。
- 完成：
  - 新增设置原始快照 `originalSettings`，加载后台设置时记录原始值。
  - 保存当前分组前，对比本组敏感字段是否发生变化。
  - 仅当敏感字段变更时弹出二次确认；取消后不会提交保存请求。
  - 保存成功后同步更新原始快照，避免重复提示已确认的变更。
- 自测记录：
  - `pnpm.cmd exec vue-tsc --noEmit`：通过。
- 问题记录：
  - 该交互需要浏览器环境手动点击确认/取消；本次通过类型检查覆盖实现正确性，整体验收阶段继续人工复核。

## 2026-05-07 产品体验优化 Milestone 4.5：IP 黑名单输入优化

- 开发目标：
  - IP 风控配置输入更友好，支持多行示例。
- 完成：
  - 复核当前后台已将 `ip_blacklist` 显示为多行输入。
  - 帮助说明和 placeholder 已包含单 IP、CIDR、多行填写示例。
  - 后端现有 IP 黑名单逻辑保持不变。
- 自测记录：
  - `go test ./middleware -run "TestIP" -v`：通过。
  - `pnpm.cmd exec vue-tsc --noEmit`：通过。

## 2026-05-07 产品体验优化 Milestone 4：后台设置分组和敏感配置保护整体验收

- 范围：
  - 4.1 设置项分组梳理和映射。
  - 4.2 设置页分组 UI 和单组保存。
  - 4.3 敏感字段密文与眼睛按钮。
  - 4.4 敏感配置保存二次确认。
  - 4.5 IP 黑名单输入优化。
- 整体验收结果：
  - 后台设置页按账号与额度、微信登录、图像生成、图片存储、人机验证、安全与监控等业务分组展示。
  - 未归类配置会进入“其他配置”，已有设置不会丢失。
  - 保存动作已收窄为当前分组，避免误提交其他分组配置。
  - 敏感字段默认隐藏，可通过眼睛按钮临时查看。
  - 修改当前分组敏感字段并保存时会二次确认，取消后不提交。
  - IP 黑名单支持多行输入和 IP/CIDR 示例，后端风控逻辑保持可用。
- 自测记录：
  - `go test ./controller -run "TestAdmin.*Setting|TestAdminPromptTemplateCRUDAndSettings" -v`：通过。
  - `go test ./middleware -run "TestIP" -v`：通过。
  - `go test ./...`：通过。
  - `pnpm.cmd exec vue-tsc --noEmit`：通过。
  - `pnpm.cmd build`：通过。沙箱内首次因 esbuild `spawn EPERM` 失败，提权后重跑通过。
- 问题记录：
  - `docs/product-experience-development-plan.md` 和部分历史文案存在编码显示异常；本次没有扩大处理，避免影响业务代码。

## 2026-05-07 产品体验优化 Milestone 5.2：渠道最近测试结果

- 开发目标：
  - 管理员点击渠道测试后，渠道列表能持续显示最近测试时间、成功/失败和错误摘要。
- 完成：
  - `Channel` 模型新增 `last_test_at`、`last_test_success`、`last_test_status`、`last_test_error` 字段，自动迁移保持兼容。
  - `/api/admin/channels/:id/test` 测试成功或失败后会写入最近测试结果。
  - 渠道列表显示最近测试时间、测试可用/失败状态和错误摘要。
  - 前端点击测试后会重新加载渠道列表，保证持久化结果立即可见。
- 自测记录：
  - `go test ./controller -run "TestAdminChannel" -v`：通过。
  - `pnpm.cmd exec vue-tsc --noEmit`：通过。
- 问题记录：
  - 首次自测发现后端局部变量重复声明和前端日期函数引用错误，已修复后重跑通过。

## 2026-05-07 产品体验优化 Milestone 5.1：概览失败率和今日核心指标

- 开发目标：
  - 后台概览和监控页能直接看到今日生成成功/失败、失败率、新增用户和积分消耗。
- 完成：
  - 复核已有监控汇总接口已统计今日生成、成功、失败、新增用户、积分消耗、支付订单和支付金额。
  - 后端 `MonitorSummary` 新增 `failure_rate`，按今日失败数 / 今日生成总数计算，无数据时为 0。
  - 前端监控指标卡新增失败率展示。
- 自测记录：
  - `go test ./controller -run "TestAdminMonitor" -v`：通过。
  - `pnpm.cmd exec vue-tsc --noEmit`：通过。

## 2026-05-07 产品体验优化 Milestone 5.4：监控页失败原因聚合

- 开发目标：
  - 管理员能在监控页看到失败原因分布和最近失败任务，便于判断上游、存储或限流问题。
- 完成：
  - 后端按错误摘要分类失败原因：上游超时、上游不可用、上游限流、存储失败、积分相关、用户取消、其他失败。
  - `MonitorSummary` 新增 `failure_reasons` 和 `recent_failures`。
  - 最近失败任务返回时间、任务 ID、尺寸、错误摘要和分类标签。
  - 前端监控页新增失败原因列表和最近失败任务列表。
- 自测记录：
  - `go test ./controller -run "TestAdminMonitor|TestAdminChannel" -v`：通过。
  - `pnpm.cmd exec vue-tsc --noEmit`：通过。
- 问题记录：
  - 实现过程中曾把失败原因块插入侧栏头部，已清理并放回监控页正确区域；重跑前端类型检查通过。

## 2026-05-07 产品体验优化 Milestone 5.3：渠道成功/失败聚合阻塞记录

- 开发目标：
  - 渠道页展示近 24 小时成功/失败数量。
- 当前结论：
  - 当前 `generations` 记录没有 `channel_id` 或渠道名称字段，无法准确把生成成功/失败归因到具体渠道。
  - 文档要求“如果当前生成记录没有 channel_id，需要先确认是否补字段；不明确时暂停”。
- 处理：
  - 本次不自行扩展生成记录与渠道的关联字段，避免改变生成调度和统计口径。
  - 已先完成渠道最近测试结果持久化，管理员可通过最近测试时间/结果判断渠道连通性。
- 问题记录：
  - 需要你确认后续是否允许在生成任务中记录实际使用的渠道 ID；确认后才能继续做每个渠道近 24 小时成功/失败聚合。

## 2026-05-07 产品体验优化 Milestone 5：渠道健康和后台运营指标整体验收

- 范围：
  - 5.1 概览失败率和今日核心指标。
  - 5.2 渠道最近测试结果。
  - 5.3 渠道成功/失败聚合阻塞记录。
  - 5.4 监控页失败原因聚合。
- 整体验收结果：
  - 监控汇总包含今日生成数、成功数、失败数、失败率、新增用户、积分消耗、支付订单和支付金额。
  - 渠道测试后会持久化最近测试时间、成功状态、HTTP 状态码和错误摘要，前端渠道列表可见。
  - 监控页展示失败原因聚合和最近失败任务，错误分类由后端提供。
  - 暂未实现按渠道聚合近 24 小时成功/失败，因为当前生成记录没有渠道关联字段，已按文档要求记录阻塞点。
- 自测记录：
  - `go test ./controller -run "TestMonitor|TestAdminChannel" -v`：通过。
  - `go test ./...`：通过。
  - `pnpm.cmd exec vue-tsc --noEmit`：通过。
  - `pnpm.cmd build`：通过。沙箱内首次因 esbuild `spawn EPERM` 失败，提权后重跑通过。
- 问题记录：
  - 若后续要完成 5.3，需要先确认是否在生成任务中记录实际使用的渠道 ID，并同步考虑历史数据为空时的展示口径。

## 2026-05-07 产品体验优化 Milestone 6.1：顶部登录态和角色展示整理

- 开发目标：
  - 顶部登录状态清晰展示普通用户/管理员/游客，避免重复登录入口。
- 完成：
  - 顶部不再单独展示“历史”按钮，登录用户统一通过账号菜单进入资产入口。
  - 登录用户菜单显示角色，普通用户同步显示当前积分。
  - 管理员仍不在普通导航暴露后台入口，后台通过 `/console/admin` 访问。
- 自测记录：
  - `pnpm.cmd exec vue-tsc --noEmit`：通过。

## 2026-05-07 产品体验优化 Milestone 6.2：用户菜单下拉入口

- 开发目标：
  - 登录用户菜单聚合历史、积分流水、套餐和退出登录。
- 完成：
  - 顶部账号区域改为 hover 下拉菜单。
  - 普通用户可进入历史记录、积分流水、积分套餐，并可退出登录。
  - 管理员菜单仅保留历史记录和退出登录，避免普通导航暴露后台入口。
- 自测记录：
  - `pnpm.cmd exec vue-tsc --noEmit`：通过。

## 2026-05-07 产品体验优化 Milestone 6.3：用户侧积分流水接口复核

- 开发目标：
  - 普通用户只能查看自己的积分流水。
- 完成：
  - 复核已有 `GET /api/credits/logs` 接口使用登录态 `userID` 查询，不接收用户 ID 参数。
  - 接口支持分页，未登录访问由路由中间件拦截。
- 自测记录：
  - `go test ./controller -run "TestCredit|TestAuth" -v`：通过。

## 2026-05-07 产品体验优化 Milestone 6.4：积分流水页面

- 开发目标：
  - 用户可以从菜单和套餐页进入完整积分流水视图。
- 完成：
  - 新增 `/credits` 路由和 `Credits.vue` 页面。
  - 页面展示当前余额、分页流水、类型、变动金额、余额、备注和时间。
  - 未登录访问 `/credits` 会跳转登录页。
  - 套餐页“查看我的积分流水”入口改为跳转 `/credits`，不再使用旧的最近 10 条弹窗。
- 自测记录：
  - `pnpm.cmd exec vue-tsc --noEmit`：通过。
  - `go test ./controller -run "TestCredit|TestAuth" -v`：通过。

## 2026-05-07 产品体验优化 Milestone 6：用户菜单和资产入口聚合整体验收

- 范围：
  - 6.1 顶部登录态和角色展示整理。
  - 6.2 用户菜单下拉入口。
  - 6.3 用户侧积分流水接口复核。
  - 6.4 积分流水页面。
- 整体验收结果：
  - 游客只看到登录/注册入口。
  - 登录用户通过顶部账号菜单进入历史记录、积分流水、积分套餐，并可退出登录。
  - 管理员不在普通导航暴露后台入口，仍通过 `/console/admin` 访问。
  - `/credits` 页面只使用当前登录用户的积分流水接口，不接收用户 ID。
  - 套餐页积分流水入口已跳转完整 `/credits` 页面。
- 自测记录：
  - `go test ./controller -run "TestCredit|TestAuth" -v`：通过。
  - `go test ./...`：通过。
  - `pnpm.cmd exec vue-tsc --noEmit`：通过。
  - `pnpm.cmd build`：通过。沙箱内首次因 esbuild `spawn EPERM` 失败，提权后重跑通过。
- 问题记录：
  - 下拉菜单当前使用 hover 展示，后续移动端若需要更强点击体验，可以再改成显式开关状态。


## 2026-05-07 公告系统第一阶段：公告中心与弹窗通知

- 需求：
  - 参考 `D:\vscodefile\sub2api`，分两步开发公告系统；第一步先完成公告中心、未读和 popup 通知。
- 完成：
  - 后端公告表增加通知方式、开始时间、结束时间。
  - 新增公告已读表，支持登录用户标记公告已读。
  - 新增 `/api/announcements` 公告列表接口，按启用状态和时间窗口过滤。
  - 新增 `/api/announcements/:id/read` 已读接口。
  - 保留旧 `/api/announcement` 单条接口兼容。
  - 管理后台公告表单增加通知方式、开始时间、结束时间。
  - 前台全局导航增加公告铃铛、未读红点、公告中心弹窗。
  - `popup` 类型未读公告会弹出强提醒，点击“我知道了”后标记已读。
  - 移除生成页右侧旧公告横幅，避免和公告中心重复展示。
- 自测记录：
  - `go test ./controller -run "TestAnnouncementAdminCRUDAndPublicActive|TestUserAnnouncementsAndReadStatus" -v`：通过。
  - `pnpm.cmd exec vue-tsc --noEmit`：通过。
  - `go test ./controller -run "TestAnnouncementAdminCRUDAndPublicActive|TestUserAnnouncementsAndReadStatus|TestAdminPromptTemplateCRUDAndSettings" -v`：通过。
  - `go test ./...`：通过。
  - `pnpm.cmd build`：通过。沙箱内首次因 esbuild `spawn EPERM` 失败，提升权限后重跑通过。

## 2026-05-07 管理后台全屏壳层修复

- 需求：
  - 后台应像 `D:\vscodefile\sub2api` 一样全屏显示，当前尺寸不对、风格不对。
- 问题定位：
  - `/console/admin` 被 `App.vue` 的普通页面容器包住，存在 `max-w-6xl`、外层 padding 和全站顶部导航。
  - 后台自身高度使用 `calc(100vh - 65px)`，进一步让后台不像独立管理应用。
- 完成：
  - `App.vue` 将 `admin` 路由纳入 full-bleed 页面，不再使用普通页面最大宽度和 padding。
  - 管理后台页隐藏全站顶部导航，改为后台自身侧边栏和内容区全屏承载。
  - 后台根容器、布局容器和侧边栏高度改为 `100vh` / `min-h-screen`。
- 自测记录：
  - `pnpm.cmd exec vue-tsc --noEmit`：通过。
  - `pnpm.cmd build`：通过。沙箱内首次因 esbuild `spawn EPERM` 失败，提升权限后重跑通过。

## 2026-05-07 管理后台整体风格对齐 sub2api

- 需求：
  - 管理后台整体界面按 `D:\vscodefile\sub2api` 的风格重新看齐。
- 参考结论：
  - sub2api 后台主要是浅色应用壳、白色侧边栏、灰白背景、圆角 2xl 卡片、浅边框、轻阴影、顶部工具条、表格独立滚动区、主按钮使用品牌色渐变。
- 完成：
  - 后台外壳从深色侧栏改为浅色侧栏和灰白内容背景。
  - 侧边栏改为 Logo/品牌区、菜单分组、浅色激活态、底部管理员信息。
  - 顶部区域改为白色轻卡片，弱化厚重分割线。
  - 卡片、列表、指标、设置页容器统一为浅边框、圆角和轻阴影。
  - 按钮、输入框、textarea、图标按钮统一为 sub2api 类似的圆角和 focus 反馈。
  - 设置分类选中态从黑底改为品牌色浅底，避免风格断裂。
- 自测记录：
  - `pnpm.cmd exec vue-tsc --noEmit`：通过。
  - `pnpm.cmd build`：通过。沙箱内首次因 esbuild `spawn EPERM` 失败，提升权限后重跑通过。

## 2026-05-07 管理后台用户页记录与充值交互拆分

- 需求：
  - 用户页点击“记录”时不应同时把充值区域带出来。
  - 原页面底部展开记录和充值后界面拥挤，需要重新设计。
- 完成：
  - 参考 `D:\vscodefile\sub2api` 后台用户页的结构：顶部工具条、独立表格区、头像首字母用户列、弹窗处理详情操作。
  - 用户列表不再在页面底部展开记录/充值区域。
  - “记录”改为打开生成记录弹窗，展示用户邮箱、生成状态、时间、Prompt 和图片入口。
  - “充值”改为打开充值弹窗，展示当前积分、充值积分和备注。
  - 充值弹窗增加当前用户信息和充值后积分预览，减少误操作。
  - 两个弹窗互不干扰，关闭后会清理对应状态，用户列表保持紧凑。
- 自测记录：
  - `pnpm.cmd exec vue-tsc --noEmit`：通过。
  - `pnpm.cmd build`：通过。沙箱内首次因 esbuild `spawn EPERM` 失败，提升权限后重跑通过。

## 2026-05-07 后台设置可视化与公告能力增强

- 需求：
  - 后台密文配置增加眼睛按钮，可临时显示/隐藏。
  - “账号与额度”需要更明显地配置新用户注册赠送积分。
  - “安全与监控”增加含义和示例，IP 黑名单输入更友好。
  - 用户列表增加最近一次使用时间。
  - 后台增加公告发布能力，生成图片页面展示公告通知。
- 完成：
  - 新增 `Announcement` 数据模型、自动建表和后台公告 CRUD 接口。
  - 新增公开公告接口，前台生成页读取启用公告并展示通知，可手动关闭当前公告。
  - 后台新增“公告”栏目，支持发布、编辑、禁用、删除公告。
  - 设置页敏感项增加眼睛图标按钮，支持显示/隐藏密文。
  - 注册赠送积分文案改为“新用户注册赠送积分”，说明初始额度含义和示例。
  - IP 黑名单改为多行输入并增加 IP/CIDR 示例；安全监控阈值和告警日期补充说明。
  - 用户列表增加“最近使用”列，展示最近登录时间。
- 自测记录：
  - `go test ./controller -run TestAnnouncementAdminCRUDAndPublicActive -v`：通过。
  - `go test ./controller -run "TestAnnouncementAdminCRUDAndPublicActive|TestAdminUserManagementAndCredits|TestAdminChannelCRUDAndTest|TestAdminPromptTemplateCRUDAndSettings" -v`：通过。
  - `pnpm.cmd exec vue-tsc --noEmit`：通过。
  - `go test ./...`：通过。
  - `pnpm.cmd build`：通过。沙箱内首次因 esbuild `spawn EPERM` 失败，提升权限后重跑通过。

## 2026-05-07 管理后台开阔版重设计

- 需求：
  - 管理后台看起来不够大气，页面里有太多小框，要求按 UI 设计专家思路重新设计。
- 完成：
  - 后台主背景、侧边栏、顶部区域重做为更开阔的控制台视觉。
  - 概览指标从多个小卡片改为连续指标带。
  - 用户、渠道、模板的顶部操作区统一为工具栏。
  - 渠道和模板列表改为统一列表行，减少小卡片堆叠。
  - 设置页保留分类逻辑，但改为更开阔的导航区和内容区。
  - 保留用户、渠道、模板的弹窗新增/编辑交互。
- 自测记录：
  - `pnpm.cmd exec vue-tsc --noEmit`：通过。
  - `go test ./controller -run "TestAdminUserManagementAndCredits|TestAdminChannelCRUDAndTest|TestAdminPromptTemplateCRUDAndSettings" -v`：通过。
  - `go test ./...`：通过。
  - `pnpm.cmd build`：通过。沙箱内首次因 esbuild `spawn EPERM` 失败，提升权限后重跑通过。

## 2026-05-07 后台操作弹窗格式统一

- 需求：
  - 新增用户已经改为弹窗后，其他新增/编辑功能也需要统一格式。
- 完成：
  - 渠道页移除右侧常驻新增/编辑表单，改为顶部“新增渠道”按钮和弹窗表单。
  - 渠道列表中的“编辑”改为打开同一弹窗。
  - 模板页移除右侧常驻新增/编辑表单，改为顶部“新增模板”按钮和弹窗表单。
  - 模板列表中的“编辑”改为打开同一弹窗。
  - 用户、渠道、模板的创建/编辑交互统一为按钮触发弹窗。
- 自测记录：
  - `pnpm.cmd exec vue-tsc --noEmit`：通过。
  - `go test ./controller -run "TestAdminUserManagementAndCredits|TestAdminChannelCRUDAndTest|TestAdminPromptTemplateCRUDAndSettings" -v`：通过。
  - `go test ./...`：通过。
  - `pnpm.cmd build`：通过。沙箱内首次因 esbuild `spawn EPERM` 失败，提升权限后重跑通过。

## 2026-05-06 新增用户弹窗优化

- 需求：
  - 新增用户常驻表单太占空间且不好看，需要改成用户页操作入口和弹窗创建。
- 完成：
  - 用户页搜索栏新增“新建用户”按钮。
  - 移除右侧常驻新增用户表单。
  - 新增用户改为居中弹窗，支持邮箱、用户名、初始密码、角色、状态和初始积分。
  - 创建成功后自动关闭弹窗、清空表单并刷新用户列表。
- 自测记录：
  - `pnpm.cmd exec vue-tsc --noEmit`：通过。
  - `go test ./controller -run "TestAdminUserManagementAndCredits" -v`：通过。
  - `go test ./...`：通过。
  - `pnpm.cmd build`：通过。沙箱内首次因 esbuild `spawn EPERM` 失败，提升权限后重跑通过。

## 2026-05-06 管理后台 UI 与用户管理优化

- 需求：
  - 管理后台界面重新优化一版。
  - 设置项太多太杂，需要重新分类，简单易懂。
  - 后台需要能添加用户并继续管理用户，不能遗漏现有管理能力。
- 现状核对：
  - 原后台已有用户列表、搜索、充值、封禁/解封、角色切换和查看生成记录。
  - 原后台缺少管理员手动新增用户能力。
- 完成：
  - 后台侧边栏和顶部区域改为更清晰的控制台风格。
  - 设置页改为左侧分类、右侧字段的布局。
  - 设置分类为账号与额度、微信登录、图像生成、图片存储、人机验证、安全与监控、其他配置。
  - 新增 `POST /api/admin/users` 管理员创建用户接口，支持邮箱、用户名、初始密码、角色、状态和初始积分。
  - 用户页新增“新增用户”表单，保留原有搜索、充值、封禁/解封、角色切换和生成记录查看能力。
  - 生成中转圈速度优化继续保留。
- 自测记录：
  - `go test ./controller -run "TestAdminUserManagementAndCredits|TestAdminPromptTemplateCRUDAndSettings" -v`：通过。
  - `pnpm.cmd exec vue-tsc --noEmit`：通过。
  - `go test ./...`：通过。
  - `pnpm.cmd build`：通过。沙箱内首次因 esbuild `spawn EPERM` 失败，提升权限后又发现 Tailwind 不存在 `border-teal-100`，已改为项目已有颜色类后重跑通过。

## 2026-05-06 生成中转圈速度优化

- 需求：
  - 生成图片本身耗时较长，当前生成中转圈速度太快，体感不匹配。
- 完成：
  - 将生成中预览占位的旋转动效从 Tailwind 默认 `animate-spin` 改为自定义 `generation-spin`。
  - 转速调整为 `2.8s` 一圈，保留持续反馈但降低焦躁感。
- 自测记录：
  - `pnpm.cmd exec vue-tsc --noEmit`：通过。
  - `go test ./...`：通过。
  - `pnpm.cmd build`：通过。沙箱内首次因 esbuild `spawn EPERM` 失败，提升权限后重跑通过。

## 2026-05-06 WeChat Server 敏感配置不暴露

- 需求：
  - WeChat Server 的访问凭证等敏感信息不能发送到普通前端显示。
- 核对结论：
  - 普通接口 `/api/auth/wechat/qrcode` 只返回 `enabled`、`qrcode_url`、`mode`，不返回 `wechat_server_address` 和 `wechat_server_token`。
  - `wechat_server_address` 和 `wechat_server_token` 仅通过管理员接口 `/api/admin/settings` 返回，该接口需要登录且需要管理员权限。
- 完成：
  - 为微信二维码公开接口补充回归测试，断言响应中不包含 `wechat_server_token`、`wechat_server_address`、实际 token 或实际服务地址。
- 自测记录：
  - `go test ./controller -run "TestWeChatQRCodeAndLoginCreatesUser|TestWeChatInvalidCodeFromServer|TestWeChatBindAndUnbind|TestWeChatDisabled" -v`：通过。
  - `go test ./...`：通过。

## 2026-05-06 微信公众号验证码登录验收

- 需求：
  - 按“点击获取验证码 -> 弹出公众号二维码 -> 用户扫码关注 -> 公众号返回验证码 -> 用户输入验证码 -> 验证通过登录/注册”的流程核对现有微信登录逻辑。
  - 输出配置和操作文档，方便后续接真实公众号/WeChat Server。
- 鉴别结论：
  - 当前项目采用 new-api 风格的验证码换 OpenID 模式，不直接接微信公众号 OAuth 回调。
  - 后端核心链路已经符合“公众号返回验证码，后端用验证码换 OpenID”的模式。
  - 前端原先进入登录页会自动展示二维码，不是“点击获取验证码后弹出二维码”，已调整为点击按钮后弹窗展示。
- 完成：
  - 登录页改为“获取验证码”按钮，点击后弹出公众号二维码弹窗。
  - 验证码输入框改为等待用户拿到公众号验证码后手动提交。
  - 后端补充 WeChat Server 非 2xx 响应校验，避免异常响应被当成正常验证码结果解析。
  - 新增 `docs/wechat-verification-login.md`，说明后台配置位置、环境变量、WeChat Server 接口协议和人工验收步骤。
- 自测记录：
  - `go test ./controller -run "TestWeChatQRCodeAndLoginCreatesUser|TestWeChatInvalidCodeFromServer|TestWeChatBindAndUnbind|TestWeChatDisabled" -v`：通过。
  - `pnpm.cmd exec vue-tsc --noEmit`：通过。
  - `go test ./...`：通过。
  - `pnpm.cmd build`：通过。沙箱内首次因 esbuild `spawn EPERM` 失败，提升权限后重跑通过。

## 2026-05-06 后台二维码本地图片配置

- 需求：
  - 微信登录二维码和联系二维码不能只要求填写网站图片地址；管理员暂时没有图片网站地址，需要能直接提供图片。
- 完成：
  - 管理后台设置页对 `wechat_qrcode_url` 和 `credit_exhausted_wechat_qrcode_url` 增加“选择图片”入口。
  - 支持选择 PNG、JPG、WebP，本地图片会转为 data URL 存入设置表。
  - 增加二维码预览和清除按钮。
  - 保留手填图片 URL 能力。
  - 限制上传图片不超过 512KB，避免设置表存入过大 base64。
- 自测记录：
  - `pnpm.cmd exec vue-tsc --noEmit`：通过。
  - `go test ./...`：通过。
  - `pnpm.cmd build`：通过。沙箱内首次因 esbuild `spawn EPERM` 失败，提升权限后重跑通过。

## 2026-05-06 后台微信登录配置入口

- 需求：
  - 管理后台没有明显的微信登录配置入口，需要增加微信登录方式配置。
- 现状确认：
  - 后端已经支持从设置表读取 `wechat_auth_enabled`、`wechat_qrcode_url`、`wechat_server_address`、`wechat_server_token`。
  - 管理后台会返回这些 key，但前端设置页没有中文标签和说明，体验上像没有入口。
- 完成：
  - 管理后台设置页补充微信登录开关、微信登录二维码、微信服务地址、微信服务 Token 的中文标签和帮助说明。
  - 对 `*_enabled` 开关使用“开启/关闭”下拉，降低手填 true/false 的错误率。
  - 后台设置测试补充微信登录配置 key 的存在性检查。
- 自测记录：
  - `go test ./controller -run "TestAdminPromptTemplateCRUDAndSettings|TestWeChatQRCodeAndLoginCreatesUser" -v`：通过。
  - `pnpm.cmd exec vue-tsc --noEmit`：通过。
  - `go test ./...`：通过。
  - `pnpm.cmd build`：沙箱内 `spawn EPERM`，提升权限后通过。

## 2026-05-06 管理员作为用户登录

- 需求：
  - 管理员登录入口不在普通登录注册页显示。
  - 管理员也是用户，可以使用普通用户登录入口和前台功能。
  - 需要进入后台时，由管理员手动输入后台路由登录。
- 完成：
  - 普通用户登录页移除后台登录入口提示。
  - 取消普通登录页对管理员账号的拦截，管理员账号可以登录后停留在前台。
  - 取消路由守卫中“管理员访问 `/login` 自动跳后台”的逻辑。
  - 保留 `/console/admin/login` 作为手动输入后台路由时的管理员登录入口。
- 自测记录：
  - `pnpm.cmd exec vue-tsc --noEmit`：通过。
  - `go test ./...`：通过。
  - `pnpm.cmd build`：沙箱内 `spawn EPERM`，提升权限后通过。

## 2026-05-06 登录注册简化

- 需求：
  - 简化优化普通用户登录注册流程。
- 完成：
  - `/login` 和 `/register` 统一使用同一个普通用户入口，减少重复页面。
  - 普通入口主流程收敛为微信验证码“登录 / 注册”，首次使用自动创建账号。
  - 邮箱登录保留给已有账号，但默认折叠，降低干扰。
  - 管理员登录继续保持 `/console/admin/login` 独立入口，避免和用户登录注册混淆。
- 自测记录：
  - `pnpm.cmd exec vue-tsc --noEmit`：通过。
  - `go test ./...`：通过。
  - `pnpm.cmd build`：沙箱内 `spawn EPERM`，提升权限后通过。

## 2026-05-06 管理员登录和用户登录隔离

- 需求：
  - 检查管理员登录和用户登录是否混淆，并确保两者是两个独立入口。
- 问题：
  - 合并 PR 后只有 `/login` 一个登录页，访问 `/console/admin` 未登录时也会跳到普通用户登录页，容易混淆。
- 完成：
  - 新增管理员登录页 `/console/admin/login`，仅支持管理员邮箱密码登录。
  - `/console/admin` 未登录或非管理员访问时跳转到 `/console/admin/login`。
  - 普通 `/login` 登录到管理员账号时会退出并引导去后台登录入口。
  - 管理员登录页如果登录普通账号，会退出并提示使用用户登录入口。
  - 401 拦截会根据当前路径跳转到普通登录或后台登录。
  - 顶部登录/退出按钮根据当前是否处于后台区域选择对应入口。
- 自测记录：
  - `pnpm.cmd exec vue-tsc --noEmit`：通过。
  - `go test ./...`：通过。
  - `pnpm.cmd build`：沙箱内 `spawn EPERM`，提升权限后通过。

## 2026-05-06 合并 feature/parallel-dev

- 来源：
  - 合并 GitHub 远端分支 `origin/feature/parallel-dev`，提交 `d1d11de feat: refine auth flow and admin UI`。
- 冲突处理：
  - `controller/auth_test.go`：PR 将邮箱注册改为禁用，但当前主线需要保留“新注册赠送积分且后台可配置”，因此保留主线注册测试和注册接口能力。
  - `web/src/views/admin/AdminDashboard.vue`：合入 PR 的后台 UI 样式，同时保留额度用完提示文案的 textarea 配置能力。
- 完成：
  - 合入登录页、注册页、路由和后台管理 UI 优化。
  - 保留主线的注册赠送积分、额度用完联系方式配置、宽高比配置等功能。
- 自测记录：
  - `go test ./...`：通过。
  - `pnpm.cmd exec vue-tsc --noEmit`：通过。
  - `pnpm.cmd build`：沙箱内 `spawn EPERM`，提升权限后通过。

## 2026-05-06 注册赠送积分和额度用完联系配置

- 需求：
  - 新注册用户默认赠送 10 积分，赠送数量可在管理后台配置。
  - 用户额度用完时展示温馨提示，并可展示微信二维码或 QQ 联系方式；这些信息可在管理后台配置。
- 完成：
  - 新增设置项 `register_gift_credits`，默认 `10`；注册和微信新用户注册统一读取该设置。
  - 新增公开接口 `GET /api/support/contact`，返回额度用完提示文案、微信二维码 URL 和 QQ 联系方式。
  - 管理后台设置页新增上述配置项的中文名称、帮助说明和输入类型。
  - 额度用完提示卡读取公开配置，展示后台配置的温馨提示、微信二维码和 QQ。
- 自测记录：
  - `go test ./controller -run "TestAuthFlow|TestRegisterGiftCreditsConfigurable|TestAdminPromptTemplateCRUDAndSettings" -v`：通过。
  - `go test ./...`：初次发现微信新用户注册测试仍期望 3 积分，已同步为 10 后通过。
  - `pnpm.cmd exec vue-tsc --noEmit`：通过。
  - `pnpm.cmd build`：沙箱内 `spawn EPERM`，提升权限后通过。

## 2026-05-06 暂时隐藏上传图像入口

- 需求：
  - 上传图像功能暂时隐藏，后续再修改。
- 完成：
  - 首页隐藏“创作方式”切换里的上传图像入口。
  - 首页隐藏图片编辑上传区域。
  - 本地草稿如果保存过图片编辑模式，会自动回退到文本生成，避免隐藏入口后仍进入编辑流程。
  - 后端图片编辑接口和前端已有逻辑暂不删除，便于后续继续改造。
- 自测记录：
  - `pnpm.cmd exec vue-tsc --noEmit`：通过。
  - `pnpm.cmd build`：沙箱内 `spawn EPERM`，提升权限后通过。

## 2026-05-06 旧尺寸配置升级补漏

- 问题：
  - 页面没有显示五个宽高比，原因是本地数据库 `enabled_image_sizes` 仍是更早的旧配置：`1024x1024,1536x1024,1024x1536`。
  - 后端上一版只自动升级旧 5 尺寸和旧 8 尺寸配置，漏掉了旧 3 尺寸配置。
- 完成：
  - 后端新增旧 3 尺寸配置自动升级，统一返回：`square`、`portrait_3_4`、`story`、`landscape_4_3`、`widescreen`。
  - 后台设置接口读取 `enabled_image_sizes` 时也返回升级后的比例 key，避免旧数据库值覆盖前端展示。
  - 补充测试覆盖旧 3/5/8 尺寸配置升级。
- 自测记录：
  - `go test ./controller -run "TestGenerationOptions" -v`：通过。
  - `go test ./controller -run "TestCreateGenerationAcceptsAspectRatioKey|TestCreateGenerationStillAcceptsMappedPixelSize" -v`：通过。
  - `go test ./...`：通过。
  - `pnpm.cmd exec vue-tsc --noEmit`：通过。

## 2026-05-06 宽高比产品化展示

- 来源：
  - 重新对齐 `plan-gpt-image2-full-integration.md` 中“前端只展示比例名称，后端维护比例到像素映射”的设计。
  - 用户确认前端需要展示：方形 1:1、竖版 3:4、故事版 9:16、横版 4:3、宽屏 16:9。
- 处理边界：
  - 文档同时要求恢复输出格式/背景选择，但上一轮已明确“输出格式和背景暂时不需要提供选择”，本次不恢复这两个前端控件。
- 完成：
  - 后端默认 `enabled_image_sizes` 改为五个比例 key：`square`、`portrait_3_4`、`story`、`landscape_4_3`、`widescreen`。
  - 后端新增比例到实际像素映射：`1024x1024`、`1152x1536`、`1008x1792`、`1536x1152`、`1792x1008`。
  - 创建文生图/图片编辑任务时，前端传比例 key，后端转换为真实像素尺寸后保存并调用上游。
  - 后端保留旧像素值兼容：如果旧请求传入当前比例映射出的真实像素值，如 `1024x1024`，仍可正常创建。
  - `/api/generation/options` 返回产品化标签和比例，不再要求前端展示像素。
  - 旧 5 尺寸和旧 8 尺寸后台配置会自动升级为新的五个比例 key。
  - 首页尺寸按钮改为展示“方形 1:1”等产品化信息，结果页和底部积分卡同步使用该展示。
  - 后台设置说明同步为五个比例 key。
- 自测记录：
  - `go test ./controller -run "TestGenerationOptions|TestCreateGenerationAcceptsAspectRatioKey" -v`：通过。
  - `pnpm.cmd exec vue-tsc --noEmit`：通过。

## 2026-05-06 前端隐藏后端状态和积分展示优化

- 需求：
  - 前端界面不再显示后端技术状态。
  - 积分显示需要更清晰。
- 完成：
  - 首页右侧账户卡移除 `后端已连接/后端未连接` 健康检查展示，并取消首页额外 health 请求。
  - 生成进度条区域移除“后端状态：...”文案，只保留用户可理解的当前阶段。
  - 账户卡改为展示用户/访客身份、可用额度、额度说明。
  - 底部生成按钮上方的积分卡优化为“本次预计消耗 + 可用余额/当前额度”，余额不足时使用提示色区分。
- 自测记录：
  - `rg "health|后端状态|后端已连接|后端未连接|currentBackendMessage|creditText" web/src/views/Home.vue web/src/components/GenerationProgress.vue`：确认无 UI 残留。
  - `pnpm.cmd exec vue-tsc --noEmit`：通过。
  - `go test ./controller ./service`：通过。

## 2026-05-06 未登录上传提示和游客配置保留

- 问题：
  - 未登录用户点击“上传图像”后，上传区内已有“仅对登录用户开放”的提示，但底部还会重复显示登录提示。
  - 游客免费额度用完后，如果去登录/注册，返回首页需要重新填写提示词、风格和尺寸等配置。
- 完成：
  - “上传图像”按钮不再写入底部错误提示；上传区内保留唯一提示。
  - 未登录编辑模式下，底部生成提示不再重复展示“请先登录后再使用上传图像编辑”。
  - 新增生成草稿本地保存，自动保留提示词、风格预设、尺寸比例和创作方式；登录/注册回来后可恢复配置。
  - 文件上传本身不做本地持久化，避免浏览器安全限制和隐私风险。
- 自测记录：
  - `rg "请先登录后再使用上传图像编辑|image_show_generation_draft|persistGenerationDraft|restoreGenerationDraft|generateHint" web/src/views/Home.vue`：确认仅保留防御性提交兜底和草稿逻辑。
  - `pnpm.cmd exec vue-tsc --noEmit`：通过。
  - `go test ./controller ./service`：通过。

## 2026-05-06 隐藏输出格式和背景选择

- 需求：
  - 输出格式和背景暂时不需要提供给用户选择。
  - 尺寸比例选择仍然需要保留。
- 完成：
  - 首页参数面板移除“输出格式”和“背景”两个选择区块。
  - 前端创建文生图和图片编辑任务时不再主动发送 `output_format`、`background`。
  - 后端仍保留这些参数的兼容能力和校验，方便后续需要时恢复 UI。
  - 尺寸比例选择、积分展示和 8 个尺寸逻辑保持不变。
- 自测记录：
  - `rg "outputFormat|backgroundOptions|outputFormatOptions|output_format|background:" web/src/views/Home.vue`：无匹配。
  - `pnpm.cmd exec vue-tsc --noEmit`：通过。
  - `go test ./controller -run TestCreateGenerationValidatesOutputOptions -v`：通过。

## 2026-05-06 高清尺寸前端不可见排查

- 问题：
  - 询问新增的 `1920x1080`、`1080x1920`、`2048x2048` 前端是否未加，或是否区分游客和登录用户。
- 结论：
  - 前端 fallback 已包含 8 个尺寸，后端 `/api/generation/options` 也不区分游客和登录用户。
  - 如果页面只显示 5 个尺寸，原因是数据库中已有旧的 `enabled_image_sizes` 设置，后端会优先读取该配置，从而覆盖代码里的新默认值。
- 完成：
  - 后端新增旧默认尺寸兼容升级：当 `enabled_image_sizes` 恰好等于旧默认 5 个尺寸时，接口自动返回新默认 8 个尺寸。
  - 管理员自定义过的其他尺寸配置不被强制覆盖，避免破坏后台配置。
  - 新增 `TestGenerationOptionsUpgradesLegacyDefaultSizeSetting` 覆盖旧默认设置升级。
- 自测记录：
  - `go test ./controller -run "TestGenerationOptions(DefaultSizesIncludeStableRatios|ReturnsSameSizesForAnonymousAndLoggedIn|UpgradesLegacyDefaultSizeSetting)" -v`：通过。
  - `pnpm.cmd exec vue-tsc --noEmit`：通过。

## 2026-05-06 功能规格 v2 终验补齐

- 来源：
  - 对照 `feature-spec-v2-generation-ux-optimization.md` 的验收标准做补查。
- 发现：
  - `CreditExhaustedGuide` 已有浅色样式，但缺少深色模式下的显式可读样式。
  - `/generations/edit` 已支持积分过期错误码，但缺少图片编辑积分过期的测试覆盖。
- 完成：
  - 补齐引导卡片深色模式样式，标题、正文、关闭按钮、图标容器和次要按钮在深色背景下保持可读。
  - 新增 `TestCreateImageEditCreditsExpired`，覆盖图片编辑端点积分过期时返回 `credits_expired`。
- 自测记录：
  - `go build ./...`：通过。
  - `rg "console\\.(error|warn|log)" web/src controller service model`：无匹配。
  - `go test ./controller -run "TestCreateImageEdit(InsufficientCredits|CreditsExpired)|TestAnonymousTrialOnceUsesStandardQuality|TestCreateGenerationCreditsExpired" -v`：通过。
  - `pnpm.cmd exec vue-tsc --noEmit`：通过。

## 2026-05-06 游客额度用尽友好提示验收修复

- 问题：
  - 验收时游客额度用尽仍看到旧英文 `free trial used, please register`，未稳定展示友好引导卡片。
- 完成：
  - 前端创建任务失败处理新增旧错误字符串兜底映射：即使后端或旧进程返回旧文案，也会展示 `CreditExhaustedGuide` 的“免费体验已结束”引导卡片。
  - 后端游客免费试用测试新增验收断言：第二次请求必须返回 HTTP 402、`free_trial_exhausted`、中文 `message`，且不能包含旧英文文案。
- 自测记录：
  - `go test ./controller -run TestAnonymousTrialOnceUsesStandardQuality -v`：通过。
  - `pnpm.cmd exec vue-tsc --noEmit`：通过。

## 2026-05-06 GPT Image 2 输出参数和 8 个尺寸

- 来源：
  - 按 `feature-spec-v2-generation-ux-optimization.md` 执行需求三 Task 3.1-3.6。
- 完成：
  - 后端生成和图片编辑请求支持 `output_format`、`background`、`output_compression`，空值保持向后兼容不传给上游。
  - 后端新增参数校验：格式仅允许 `png/jpeg/webp`，背景仅允许 `opaque/transparent`，压缩率限制 `0-100`，并拒绝 `jpeg + transparent`。
  - `Generation` 模型保存输出格式、背景和压缩率，历史记录可返回对应字段。
  - 默认可用尺寸扩展为 8 个：`1280x720`、`720x1280`、`1024x1024`、`1536x1024`、`1024x1536`、`1920x1080`、`1080x1920`、`2048x2048`。
  - 上游尺寸改为对合法请求尺寸透传，避免高清尺寸被映射回低尺寸。
  - 首页参数面板新增输出格式和背景选择；选择 JPEG 时自动隐藏并清空透明背景。
  - 后台设置页的尺寸说明同步更新为 8 个默认尺寸。
- 自测记录：
  - `go test ./controller -run TestCreateGenerationValidatesOutputOptions -v`：通过。
  - `go test ./controller ./service`：通过。
  - `pnpm.cmd exec vue-tsc --noEmit`：通过。

## 2026-05-06 积分用尽引导卡片

- 来源：
  - 按 `feature-spec-v2-generation-ux-optimization.md` 执行需求一 Task 1.1-1.3。
- 完成：
  - 后端免费试用用尽改为 HTTP 402，并返回 `free_trial_exhausted` 和中文 `message`。
  - 登录用户积分不足和积分过期拆分为 `insufficient_credits`、`credits_expired`。
  - 新增 `CreditExhaustedGuide.vue`，根据错误类型展示注册/登录/充值引导。
  - 首页创建任务时识别 402 积分错误码，主预览区展示引导卡片，其他错误仍走普通错误提示。
  - 生成服务创建前改为先判断积分过期再判断余额，避免过期积分被误报为余额不足。
- 自测记录：
  - `go test ./controller ./service`：通过。
  - `pnpm.cmd exec vue-tsc --noEmit`：通过。
  - `pnpm.cmd build`：沙箱内因 esbuild `spawn EPERM` 失败，授权后重新执行通过。
## 2026-05-06 生成页参数面板折叠按钮优化

- 来源：
  - 按 `feature-spec-v2-generation-ux-optimization.md` 执行需求二 Task 2.1-2.3。
- 完成：
  - 右侧参数面板折叠按钮从竖线和字符箭头改为 SVG 方向箭头。
  - 提升收起状态按钮可见度和点击宽度，hover 时增加更明确的反馈。
  - 面板收起后增加 3 秒轻微提示动画，提示可再次展开。
- 自测记录：
  - `pnpm.cmd exec vue-tsc --noEmit`：通过。
  - `pnpm.cmd build`：沙箱内因 esbuild `spawn EPERM` 失败，授权后重新执行通过。
## 2026-05-04 gpt-image-2 比例尺寸补齐

- 需求：
  - 复核 OpenAI 官方 Image API 尺寸说明，并补齐产品侧常用比例：`16:9`、`9:16`、`1:1`、`3:2`、`2:3`。
- 结论：
  - 官方公开 API Reference 对 GPT image models 的 `size` 参数仍列出 `1024x1024`、`1536x1024`、`1024x1536` 和 `auto`。
  - 项目当前按 `gpt-image-2` 兼容校验支持灵活尺寸，因此产品侧补齐稳定映射：`16:9=1280x720`、`9:16=720x1280`、`1:1=1024x1024`、`3:2=1536x1024`、`2:3=1024x1536`。
- 完成：
  - 后端默认 `enabled_image_sizes` 更新为五个比例尺寸，并复用同一个默认常量给生成接口和后台设置接口。
  - 前端默认和接口异常 fallback 尺寸同步为五个比例，且不再按游客/登录状态区分 fallback。
  - 后台设置页说明更新为五个比例和像素向上取整积分规则。
  - 后端测试补充默认尺寸、比例 label、以及 `1536x1024/1024x1536=2` 积分校验。
- 自测记录：
  - go test ./controller ./service：通过。
  - go test ./...：通过。
  - pnpm.cmd build：沙箱内因 esbuild spawn EPERM 失败，授权后重新执行通过。

## 2026-05-04 取消深色入口和侧拉手文案优化

- 需求：
  - 暂时不需要深色模式。
  - 右侧侧拉手 hover 显示“参数”不够优雅，需要调整。

- 完成：
  - 移除顶部“深色/浅色”切换入口。
  - 应用启动时强制移除 `html.dark` 和本地 `theme` 记录，避免历史深色偏好继续生效。
  - 侧拉手去掉竖排“参数”文字，改为更轻量的三段式把手线条和方向箭头，仅保留 `title`/`aria-label` 辅助提示。

- 自测记录：
  - `pnpm.cmd build`：通过。沙箱内仍因 esbuild `spawn EPERM` 失败，授权后重新执行通过。

## 2026-05-04 参数面板折叠拉手优化

- 问题：
  - 右侧参数栏的收起/展开按钮过小，收起后入口不明显，点击反馈和视觉层级不够友好。

- 完成：
  - 将圆形按钮改为贴边侧拉手，展开时贴在参数栏分割线上，收起时以半透明状态悬浮在右侧边缘。
  - hover 后拉手会变清晰并显示“参数”提示，方向箭头保持展开/收起语义。
  - 收起时去掉右侧边框占位，让主预览区视觉更干净。

- 自测记录：
  - `pnpm.cmd build`：通过。沙箱内仍因 esbuild `spawn EPERM` 失败，授权后重新执行通过。

## 2026-05-04 生成完成后重试按钮位置优化

- 问题：
  - 生成完成后，“重新生成上一次”仍显示在右侧参数栏内部，和结果图片操作割裂，位置不直观。

- 完成：
  - 生成结果底部工具栏新增“再生成一次”主按钮，和下载操作放在同一区域。
  - 右侧参数栏在已有结果图时不再显示“重新生成上一次”，避免重复入口和视觉干扰。

- 自测记录：
  - `pnpm.cmd build`：通过。沙箱内仍因 esbuild `spawn EPERM` 失败，授权后重新执行通过。

## 2026-05-04 前端尺寸积分兜底修复

- 问题：
  - 后端像素计费已改为向上取整，但前端在默认尺寸、接口失败、或旧接口只返回 `sizes` 时没有 `credit_cost`，会统一兜底显示 `1` 积分。

- 完成：
  - 前端新增与后端一致的尺寸像素计费函数：`ceil(width * height / 1024 / 1024)`，最低 `1` 积分。
  - 默认尺寸、`sizes` fallback、接口异常 fallback、以及缺失 `credit_cost` 的 `size_options` 都会补齐积分。
  - 尺寸按钮展示和底部“本次预计消耗”统一使用补齐后的积分，保证 `1536x1024` 显示 `2` 积分。

- 自测记录：
  - `pnpm.cmd build`：通过。沙箱内仍因 esbuild `spawn EPERM` 失败，授权后重新执行通过。

## 2026-05-03 像素比例计费和空状态文案调整

- 需求：
  - 去除质量计费，改为按尺寸像素比例收费。
  - `1024x1024` 作为基准消耗 `1` 积分，其他尺寸按总像素比例递增。
  - 首页空状态文案不能承诺“几秒钟后看到作品”。

- 完成：
  - 后端生成和图片编辑任务改为按尺寸像素计费：`1024x1024 = 1`，`1536x1024 / 1024x1536 = 1.5`。
  - `/api/generation/options` 的 `size_options` 返回 `credit_cost`，前端尺寸按钮和底部消耗提示同步展示。
  - 前端移除“生成质量”选择，接口仍固定传标准质量给上游保持兼容。
  - 首页空状态副文案改为“写下你脑海里的画面，选择合适比例，系统会持续展示生成进度，直到作品完成。”。

- 自测记录：
  - `go test ./controller ./service`：通过。
  - `go test ./...`：通过。
  - `pnpm.cmd build`：通过。沙箱内仍因 esbuild `spawn EPERM` 失败，授权后重新执行通过。

## 2026-05-03 像素计费向上取整和按钮文案修复

- 问题：
  - 用户生成图片后，右侧主按钮在上传图像模式下显示“开始编辑”，主操作语义不统一。
  - 像素比例计费按小数展示，`1536x1024` 显示为 `1.5` 积分，不符合“依次类推”的整档收费预期。

- 完成：
  - 主按钮文案统一为“开始生成”，上传图像仅作为输入方式，不再改变主按钮文案。
  - 尺寸计费从小数比例改为向上取整：`1024x1024 = 1`，`1536x1024 / 1024x1536 = 2`。

- 自测记录：
  - `go test ./controller ./service`：通过。
  - `go test ./...`：通过。
  - `pnpm.cmd build`：通过。沙箱内仍因 esbuild `spawn EPERM` 失败，授权后重新执行通过。

## 2026-05-03 质量计费和首页引导调整

- 需求：
  - 积分消耗需要根据生成质量不同而不同。
  - 高级参数暂时关闭，不再显示。
  - 左侧空状态“准备好了吗”文案需要更能引导用户开始使用。

- 完成：
  - 后端生成和图片编辑任务恢复按 `quality` 计算积分：快速 `0.2`、标准 `1`、高清 `4`。
  - 接口仍兼容未传 `quality` 的旧请求，默认按标准质量处理。
  - 首页恢复简洁的“生成质量”选择，并在按钮、结果说明和“本次预计消耗”里同步显示当前积分。
  - 首页移除“高级参数”整块 UI，保留提示词、风格、质量、尺寸和推荐样例。
  - 左侧空状态文案改为“把想法变成一张好图”，降低首次使用门槛。

- 自测记录：
  - `go test ./controller ./service`：通过。
  - `go test ./...`：通过。
  - `pnpm.cmd build`：通过。沙箱内仍因 esbuild `spawn EPERM` 失败，授权后重新执行通过。

## 2026-05-03 尺寸计费和后台路由调整

- 需求：
  - 首页去除“质量”选择。
  - 标准收费显示到“尺寸比例”上，并在生成提示区同步展示。
  - 管理员前端路由改为 `/console/admin`。

- 完成：
  - 前端固定使用标准质量 `medium` 调用图片生成/编辑接口，不再显示质量选择控件。
  - `/api/generation/options` 的 `size_options` 增加 `credit_cost` 字段。
  - 后端创建生成和图片编辑任务时，积分消耗改为按尺寸选项计算，当前稳定尺寸统一标准收费 `1` 积分。
  - 首页尺寸比例按钮展示对应积分，底部“本次预计消耗”同步读取当前尺寸的积分。
  - 管理员页面前端路由改为 `/console/admin`，旧 `/admin` 和 `/console/image-show-admin` 均不再进入后台。
  - 部署文档中的后台访问路径同步为 `/console/admin`。

- 自测记录：
  - `go test ./controller ./service`：通过。
  - `go test ./...`：通过。
  - `pnpm.cmd build`：通过。沙箱内仍因 esbuild `spawn EPERM` 失败，授权后重新执行通过。

## 2026-05-03 主题、生成入口和导航优化

- 问题：
  - 首页大量样式写死浅色，点击深色后视觉变化不明显。
  - “开始生成”按钮在右侧面板底部，需要滚动到底才能点击。
  - 管理后台入口暴露在右上角，不希望普通浏览路径里直接可见。
  - 登录/退出区域视觉不统一。

- 完成：
  - 为首页补充暗色主题全局样式，使右侧面板、输入框、按钮等跟随深色切换。
  - 生成按钮改为右侧面板底部 sticky 操作区，并根据提示词/上传图状态显示可生成提示。
  - 移除右上角“管理后台”入口，后台改为隐藏路由 `/console/image-show-admin` 访问。
  - 管理员登录后不再自动跳后台；登录/注册和退出区域统一为圆角账户区。

- 自测记录：
  - `pnpm.cmd build`：通过。

## 2026-05-03 生成状态页简化优化

- 问题：
  - 放大后的生成状态页视觉元素偏多，背景和动效存在抢焦点的问题。

- 完成：
  - 生成状态页改为浅色简洁画布和居中状态面板。
  - 减少点阵、波纹等复杂装饰，只保留轻量背景和鼠标光效。
  - 底部状态栏收敛为当前阶段、后端状态、进度条和阶段节点。

- 自测记录：
  - `pnpm.cmd build`：通过。

## 2026-05-03 生成状态背景放大

- 问题：
  - 点击生成后左侧生成状态背景太小，视觉重心弱，状态展示不够明显。

- 完成：
  - 生成状态区域从居中卡片改为占满左侧工作区的大画布。
  - 放大生成动画和鼠标交互光效。
  - 状态标题、进度条、阶段节点移到底部信息栏，增强可读性。

- 自测记录：
  - `pnpm.cmd build`：通过。

## 2026-05-03 尺寸统一、编辑登录限制和功能栏保持

- 需求：
  - 尺寸选项不再区分游客和登录用户。
  - 生成完成后不要自动收起右侧功能栏，方便继续调整参数。
  - 上传图像编辑只对登录用户开放，未登录用户给友好提示。

- 完成：
  - `/api/generation/options` 对游客和登录用户返回同一组后台启用尺寸。
  - 文生图匿名试用仍保留一次试用和验证码/指纹逻辑，但不再按尺寸区别游客。
  - `/api/generations/edit` 未登录直接返回 `401` 和登录提示，不再消耗匿名试用。
  - 前端上传图像模式未登录时禁用文件选择并显示提示。
  - 生成完成后保留右侧参数栏展开状态。

- 自测记录：
  - `go test ./controller ./service`：通过。
  - `go test ./...`：通过。
  - `pnpm.cmd build`：通过。

## 2026-05-03 稳定尺寸收敛

- 问题：
  - gpt-image-2 支持符合约束的灵活尺寸，但高分辨率和更多比例会增加不稳定性、耗时和上游失败概率。
  - 当前开放选项过多，不利于先保证生成稳定性。

- 决策：
  - 默认只开放 OpenAI 官方常用且稳定的三种尺寸：`1024x1024`、`1536x1024`、`1024x1536`。
  - 游客只开放 `1024x1024`；登录用户开放三种稳定尺寸。
  - 后端继续保留 gpt-image-2 尺寸合法性校验，防止后台误配置。

- 完成：
  - 后端默认 `enabled_image_sizes` 收敛为 `1024x1024,1536x1024,1024x1536`。
  - 前端默认和 fallback 尺寸同步收敛。
  - 后台设置说明改为推荐稳定尺寸。

- 自测记录：
  - `go test ./controller ./service`：通过。
  - `go test ./...`：通过。
  - `pnpm.cmd build`：通过；普通沙箱首次触发 `esbuild spawn EPERM`，提权重跑通过。

## 2026-05-03 gpt-image-2 官方尺寸对齐

- 问题：
  - 上一版为比例展示增加了 `768x432`、`432x768`、`768x768` 等小尺寸。
  - 复核官方 `gpt-image-2` 文档后确认，这些尺寸总像素低于 `655360`，不符合官方约束。

- 决策：
  - 尺寸列表改为官方约束内的比例尺寸：宽高为 16 的倍数、最大边不超过 3840、长短边不超过 3:1、总像素在 `655360` 到 `8294400` 之间。
  - 产品比例映射为：`16:9=1280x720`、`9:16=720x1280`、`1:1=1024x1024`、`3:2=1152x768`、`2:3=768x1152`。
  - 登录用户可继续启用更高合法尺寸，如 `1536x1024`、`1024x1536`、`2048x2048`、`3840x2160`。

- 完成：
  - 后端默认 `enabled_image_sizes` 更新为 gpt-image-2 合法尺寸。
  - 新增后端尺寸兼容校验，拒绝后台误配置的不合法尺寸进入生成接口。
  - 前端默认和 fallback 尺寸更新为官方合法尺寸。
  - 后台设置说明更新，提示 gpt-image-2 尺寸约束。

- 自测记录：
  - `go test ./controller ./service`：通过。
  - `go test ./...`：通过。
  - `pnpm.cmd build`：通过；普通沙箱首次触发 `esbuild spawn EPERM`，提权重跑通过。

## 2026-05-03 真实图片编辑与比例尺寸

- 需求：
  - 首页支持“输入文本”和“上传图像”两种创作方式。
  - 输入文本从零生成图片；上传图像后按描述进行真实图片编辑。
  - 尺寸选择按产品比例展示：`16:9`、`9:16`、`1:1`、`3:2`、`2:3`。

- 调研：
  - OpenAI Images API 有文生图 `/v1/images/generations` 和图片编辑 `/v1/images/edits` 两类接口。
  - 图片编辑使用 multipart 表单上传 `image` 文件，并传入 `prompt`、`model`、`size`、`quality` 等字段。
  - 官方 GPT Image 尺寸主要是 `1024x1024`、`1536x1024`、`1024x1536` 和 `auto`；产品侧比例需要映射到实际像素尺寸。

- 决策：
  - 新增 `/api/generations/edit`，沿用现有渠道、积分、匿名试用、SSE、R2 保存流程。
  - sub2api 编辑调用按 OpenAI 兼容方式转发到 `/v1/images/edits`。
  - `generation/options` 同时返回原始 `sizes` 和新的 `size_options`，前端显示比例、提交实际尺寸。
  - 后端默认尺寸补齐小尺寸和登录用户高尺寸：`16:9`、`9:16`、`1:1`、`3:2`、`2:3`。

- 完成：
  - `generations` 表新增 `mode`、`source_r2_key`、`source_image_url` 字段，AutoMigrate 自动补列。
  - 新增图片编辑任务创建、编辑状态流转、源图 R2 保存、编辑结果保存。
  - 前端新增“输入文本 / 上传图像”切换、源图上传预览、格式和大小校验。
  - 前端尺寸选择改为比例按钮，并展示实际像素尺寸。

- 自测记录：
  - `go test ./service ./controller`：通过。
  - `go test ./...`：通过。
  - `pnpm.cmd build`：通过；普通沙箱首次触发 `esbuild spawn EPERM`，提权重跑通过。

## 2026-05-02 下载层级、图片体积和游客尺寸修复

- 问题：
  - 首页结果图右下角“下载/下载全部”无效，但全屏后下载有效。
  - 768x768 下载文件约 2MB，体积偏大。
  - 游客用户仍能看到较高尺寸，登录用户应显示全部尺寸。

- 原因：
  - 底部操作层没有明确 z-index，可能被结果图片层盖住导致点击不到。
  - 后端缩放小尺寸时统一输出 PNG，照片/插画类图片 PNG 体积会明显偏大。
  - 游客尺寸过滤条件为宽高 `<=1024`，导致 `1024x1024` 也对游客可见。

- 决策：
  - 结果图底部操作层提高 z-index。
  - 缩放后的图片改为 JPEG 质量 88 输出，降低文件体积。
  - 游客只显示宽高都小于 1024 的尺寸；登录用户仍显示后台启用的全部尺寸。

- 完成：
  - 首页结果图底部操作区增加 `z-20`。
  - 下载工具根据实际图片类型补全文件扩展名，JPEG 输出下载为 `.jpg`。
  - 前端游客默认尺寸改为 `512x512/768x768`，登录状态变化后重新拉取后端尺寸。
  - `ResizeImageBytes` 改为输出 `image/jpeg`。
  - 游客尺寸过滤从 `<=1024` 调整为 `<1024`，无低尺寸配置时回退到 `512x512/768x768`。
  - 更新尺寸过滤和缩放测试。

- 自测记录：
  - `go test ./service ./controller`：通过。
  - `go test ./...`：通过。
  - `pnpm.cmd build`：通过；普通沙箱首次触发 `esbuild spawn EPERM`，提权重跑通过。

## 2026-05-02 侧栏抽屉和全屏按钮修复

- 问题：
  - 用户希望参数栏像侧边抽屉一样，通过旁边把手/箭头收起展开。
  - “全屏查看”只有收起参数后才容易看到，并且点击后没有进入真正全屏。

- 决策：
  - 参数栏改为右侧抽屉，桌面端通过侧边圆形把手收起/展开。
  - 结果图右上角固定显示“全屏查看”，不受参数栏是否展开影响。
  - 全屏查看同时打开应用内预览层并调用浏览器 Fullscreen API；浏览器拒绝时仍保留应用内全屏预览。

- 完成：
  - 移除右下角“展开参数”悬浮按钮。
  - 右侧参数栏改为宽度过渡抽屉，并新增侧边箭头把手。
  - 结果图右上角新增固定操作区，全屏按钮始终可见。
  - 全屏层改为 `v-show` 常驻节点，点击时可调用 `requestFullscreen()`。

- 自测记录：
  - `pnpm.cmd build`：通过。

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

## 2026-05-03 深色模式和积分消耗提示优化

- 问题：
  - 首页点击深色模式后，右侧参数面板、输入区和浅色文字变化不明显，整体仍接近浅色界面。
  - 生成按钮附近只显示一行预计消耗文案，不够醒目，用户在点击前不容易确认本次会消耗多少积分。
- 完成：
  - 首页主容器、预览空状态、参数面板、折叠按钮和底部生成区补充深色模式样式。
  - 增加深色模式兜底样式，覆盖首页面板内常见浅色背景、边框和文字颜色，提升切换后的对比度。
  - 生成按钮上方新增“本次预计消耗”信息块，展示预计积分、生成模式、质量、尺寸和当前余额/免费试用提示。
- 自测记录：
  - `pnpm.cmd build`：通过。沙箱内仍因 esbuild `spawn EPERM` 失败，授权后重新执行通过。

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
## 2026-05-07 公告中心第二阶段：投放范围与阅读统计

- 开发目标：
  - 参考 sub2api 公告管理的清晰度，继续补齐公告后台能力。
  - 当前系统还没有用户分组/订阅体系，因此本阶段先按登录状态和角色实现投放范围，避免引入过大的权限模型。
- 完成：
  - 公告模型新增 `target` 字段，支持 `all`、`guest`、`user`、`admin` 四种投放范围。
  - 前台公告列表按当前访问者身份过滤：游客只看游客/全员公告，普通用户看用户/全员公告，管理员看管理员/用户/全员公告。
  - 后台公告列表新增阅读数量统计。
  - 新增后台接口 `GET /api/admin/announcements/:id/reads`，可查看某条公告的已读用户、角色和阅读时间。
  - 后台公告表单新增“投放范围”选择，公告列表新增投放范围、阅读数和阅读详情入口。
- 自测记录：
  - `go test ./controller -run "TestAnnouncement" -v`：通过。
  - `go test ./...`：通过。
  - `pnpm.cmd exec vue-tsc --noEmit`：通过。
  - `pnpm.cmd build`：通过。沙箱内首次仍因 esbuild `spawn EPERM` 失败，授权后重跑通过。
- 问题记录：
  - 暂未实现“按用户组/订阅计划/标签投放”，因为当前代码没有这些业务对象；后续若新增用户分组，可在 `target` 基础上扩展 `target_ref` 或单独公告投放表。
## 2026-05-07 产品体验优化 Milestone 1.1：首页价值表达和游客规则提示

- 开发目标：
  - 根据 `docs/product-experience-development-plan.md`，先完成首页转化与生成前引导的第一个小功能。
  - 让游客在首页直接理解“能做什么、免费规则是什么、注册后有什么收益”。
- 完成：
  - 首页空态主文案按登录状态区分：游客看到“输入一句话，生成封面、商品图和头像”，登录用户看到继续创作引导。
  - 首页权益卡片文案按登录状态区分：游客显示免费试用和注册保存历史，登录用户显示按尺寸消耗积分。
  - 生成按钮附近的游客提示改为更明确的中文权益说明。
- 自测记录：
  - `pnpm.cmd exec vue-tsc --noEmit`：通过。
- 问题记录：
  - 本小功能只做前端文案和权益提示，不修改免费次数、注册赠送积分和扣费规则。
## 2026-05-07 产品体验优化 Milestone 1.2：场景化 Prompt 模板后端扩展

- 开发目标：
  - 复用现有 `prompt_templates` 能力，为首页场景入口提供后台可维护的数据来源。
- 完成：
  - 默认 Prompt 模板新增 `scenario` 分类。
  - 首批场景包含：小红书封面、商品展示图、头像、海报、壁纸。
  - 后台模板管理的分类文案和下拉选项新增“首页场景入口”。
  - 公开模板接口测试补充 `scenario` 分类断言。
- 自测记录：
  - `go test ./controller -run "TestPromptTemplate|TestAdminPromptTemplate" -v`：通过。
  - `pnpm.cmd exec vue-tsc --noEmit`：通过。
- 问题记录：
  - 本小功能只扩展模板来源，首页展示和点击联动放到 Milestone 1.3 实现。
## 2026-05-07 产品体验优化 Milestone 1.3：首页场景入口 UI 与参数联动

- 开发目标：
  - 让不懂 Prompt 的用户可以先选择用途，再由系统带入提示词、推荐比例和风格。
- 完成：
  - 首页输入框下方新增场景入口区。
  - 场景入口读取 `scenario` 分类模板，接口不可用时使用前端兜底场景。
  - 点击场景入口后自动填入 Prompt，并联动推荐尺寸和风格。
  - 当前场景入口只在文本生成模式展示，图片编辑模式暂不展示。
- 自测记录：
  - `pnpm.cmd exec vue-tsc --noEmit`：通过。
- 问题记录：
  - 当前后端 `prompt_templates` 还没有独立字段保存推荐尺寸和推荐风格，前端暂按场景顺序使用兜底映射；后续如需后台完全自定义联动参数，需要扩展模板字段或新增配置表。
## 2026-05-07 产品体验优化 Milestone 1.4：免费额度耗尽提示优化

- 开发目标：
  - 额度不足时用更友好的中文说明代替生硬错误文案，并给出明确下一步。
- 完成：
  - 游客免费次数用完提示改为“注册领取积分”，并说明提示词和尺寸会保留。
  - 登录用户积分不足提示改为“可换低消耗尺寸或查看积分套餐”。
  - 积分过期提示改为更明确的继续使用引导。
  - 保留后台配置的微信二维码、QQ 和自定义联系文案展示能力。
- 自测记录：
  - `pnpm.cmd exec vue-tsc --noEmit`：通过。
- 问题记录：
  - 本小功能只优化前端引导文案，不调整游客免费次数和积分不足的后端判定。
## 2026-05-07 产品体验优化 Milestone 1.5：登录/注册后保留生成参数

- 开发目标：
  - 游客从首页跳转登录/注册后，返回首页仍保留已填写的生成参数。
- 完成：
  - 复核现有首页已通过 `image_show_generation_draft` 保存 Prompt、风格、尺寸和生成模式。
  - `/register` 当前复用登录页，邮箱登录和微信验证码登录成功后都会回到首页，首页会自动恢复草稿。
  - 本小功能无需新增代码，沿用现有草稿机制即可满足当前验收目标。
- 自测记录：
  - `pnpm.cmd exec vue-tsc --noEmit`：通过。
- 问题记录：
  - 当前草稿只保存文本生成参数，不保存上传图片文件；上传图片编辑功能目前仍隐藏，后续重新开放图片编辑时需要单独设计图片草稿策略。
## 2026-05-07 产品体验优化 Milestone 1：首页转化与生成前引导整体验收

- 范围：
  - 1.1 首页价值表达和游客规则提示。
  - 1.2 场景化 Prompt 模板后端扩展。
  - 1.3 首页场景入口 UI 与参数联动。
  - 1.4 免费额度耗尽提示优化。
  - 1.5 登录/注册后保留生成参数。
- 整体验收结果：
  - 游客首页能看到明确用途、免费试用规则和注册收益。
  - 首页场景入口可用，点击后会带入 Prompt、尺寸和风格。
  - 免费额度耗尽、积分不足、积分过期提示均为中文友好表达。
  - 登录/注册返回首页后沿用现有草稿机制保留生成参数。
  - 后台模板管理支持“首页场景入口”分类。
- 自测记录：
  - `go test ./controller -run "TestPromptTemplate|TestAdminPromptTemplate" -v`：通过。
  - `go test ./...`：通过。
  - `pnpm.cmd exec vue-tsc --noEmit`：通过。
  - `pnpm.cmd build`：通过。沙箱内首次仍因 esbuild `spawn EPERM` 失败，授权后重跑通过。
- 问题记录：
  - 场景模板目前只有分类、名称、Prompt，推荐尺寸和风格仍由前端兜底映射；如后续要求后台完全配置场景联动参数，需要扩展模板字段。
## 2026-05-07 产品体验优化 Milestone 2.1：统一计费展示口径

- 开发目标：
  - 首页预计积分和后端实际扣费保持同一像素计费口径。
- 完成：
  - 复核后端 `service.CostForSize` 已按 `1024x1024 = 1`、像素向上取整计费。
  - 首页前端兜底计费新增比例尺寸到真实像素尺寸的映射。
  - `square`、`portrait_3_4`、`story`、`landscape_4_3`、`widescreen` 的前端兜底计算改为按真实像素向上取整，而不是依赖手写积分。
- 自测记录：
  - `go test ./service -run "Test" -v`：通过。
  - `pnpm.cmd exec vue-tsc --noEmit`：通过。
- 问题记录：
  - 当前首页优先使用 `/api/generation/options` 返回的 `credit_cost`，前端兜底只用于接口异常或旧数据兼容。
## 2026-05-07 产品体验优化 Milestone 2.2：首页计费规则入口

- 开发目标：
  - 用户在点击生成前能理解当前预计积分的计算方式。
- 完成：
  - 首页预计消耗区域新增“查看计费规则”入口。
  - 新增计费说明弹窗，说明标准图、像素向上取整和失败任务退回规则。
  - 弹窗关闭不影响当前 Prompt、尺寸和风格参数。
- 自测记录：
  - `pnpm.cmd exec vue-tsc --noEmit`：通过。
- 问题记录：
  - 失败任务退回说明按当前后端已有 `Refund` 逻辑描述；后续若业务规则调整，需要同步更新说明。
## 2026-05-07 产品体验优化 Milestone 2.3：套餐卡片可生成张数

- 开发目标：
  - 让用户购买前能直观看到每个套餐大约能生成多少张标准图。
- 完成：
  - 套餐卡片新增“约可生成 N 张标准图”。
  - 标准图口径为 `1024 x 1024`、`1 积分/张`。
  - 保持原购买入口和套餐有效期展示不变。
- 自测记录：
  - `pnpm.cmd exec vue-tsc --noEmit`：通过。
- 问题记录：
  - 当前换算使用套餐积分向下取整；如果后续支持小数积分套餐，需要确认是否展示小数张数或仍按整数张展示。
## 2026-05-07 产品体验优化 Milestone 2.4：积分规则说明与流水入口

- 开发目标：
  - 让用户在套餐页能看到积分规则，并能查看自己的积分流水。
- 完成：
  - 套餐页新增“积分使用规则”说明，覆盖尺寸计费、有效期和失败退回。
  - 套餐页新增“查看我的积分流水”入口。
  - 登录用户点击后通过 `/api/credits/logs` 加载最近 10 条流水。
  - 未登录用户点击流水入口会跳转登录。
- 自测记录：
  - `go test ./controller -run "TestCredit" -v`：通过。
  - `pnpm.cmd exec vue-tsc --noEmit`：通过。
- 问题记录：
  - 当前用户侧流水先用弹窗展示最近 10 条；完整分页页面放到后续用户菜单/资产入口 Milestone 继续扩展。
## 2026-05-07 产品体验优化 Milestone 2.5：支付失败/生成失败扣费规则核对

- 开发目标：
  - 核对套餐页和首页计费说明是否与后端扣费/退费逻辑一致。
- 核对结果：
  - 创建生成任务时，登录用户先检查积分，再创建任务并扣除积分。
  - 上游生成失败会调用 `refundGenerationCredits` 退回本次扣除积分。
  - 图片保存失败同样会退回本次扣除积分。
  - 待处理状态取消任务会退回积分。
  - 生成中取消任务当前不会退回积分，因为上游任务可能已经开始消耗。
  - 支付回调已有幂等处理，重复通知不会重复加积分。
- 完成：
  - 新增 service 包测试覆盖失败退回核心函数。
  - 前端计费说明按上述规则表达。
- 自测记录：
  - `go test ./service ./controller -run "Test.*Credit|Test.*Generation" -v`：通过。
  - `pnpm.cmd exec vue-tsc --noEmit`：通过。
- 问题记录：
  - “生成中取消不退回”是当前真实业务规则，已记录；如果你希望生成中取消也退回，需要重新定义取消策略和上游成本承担规则。
## 2026-05-07 产品体验优化 Milestone 2：计费透明和套餐页优化整体验收

- 范围：
  - 2.1 统一计费展示口径。
  - 2.2 首页计费规则入口。
  - 2.3 套餐卡片可生成张数。
  - 2.4 积分规则说明与流水入口。
  - 2.5 支付失败/生成失败扣费规则核对。
- 整体验收结果：
  - 首页预计消耗、后端实际扣费、套餐页说明均使用 `1024x1024 = 1 积分` 和像素向上取整口径。
  - 首页可查看计费规则。
  - 套餐卡片展示约可生成标准图数量。
  - 套餐页展示积分规则，并支持登录用户查看最近积分流水。
  - 已明确记录生成失败、保存失败、待处理取消、生成中取消的积分处理规则。
- 自测记录：
  - `go test ./service ./controller -run "Test.*Credit|Test.*Package|Test.*Generation" -v`：通过。
  - `go test ./...`：通过。
  - `pnpm.cmd exec vue-tsc --noEmit`：通过。
  - `pnpm.cmd build`：通过。沙箱内首次仍因 esbuild `spawn EPERM` 失败，授权后重跑通过。
- 问题记录：
  - 用户侧积分流水当前只在套餐页弹窗展示最近 10 条；完整分页入口将在 Milestone 6 用户资产菜单中继续做。
  - 生成中取消当前不退回积分，这是当前真实规则；如需调整，需要先确认业务口径。
## 2026-05-07 产品体验优化 Milestone 3.1：历史列表搜索和筛选接口

- 开发目标：
  - 历史列表支持按 Prompt 关键词、状态和尺寸筛选，为资产库化改造打基础。
- 完成：
  - `GET /api/generations` 新增 `keyword`、`status`、`size` 查询参数。
  - 筛选逻辑仍限定当前登录用户和未删除记录，避免越权看到他人历史。
  - 非数字 `status` 返回 400。
  - 新增测试覆盖关键词、状态、尺寸组合筛选和无效状态。
- 自测记录：
  - `go test ./controller -run "TestUserGenerationHistory" -v`：通过。
- 问题记录：
  - 当前关键词使用 SQL `LIKE` 模糊匹配，适合现阶段轻量搜索；后续历史量变大后再考虑全文索引。
## 2026-05-07 产品体验优化 Milestone 3.2：历史页筛选 UI

- 开发目标：
  - 让用户可以在历史页通过关键词、状态和尺寸快速定位作品。
- 完成：
  - 历史页新增 Prompt 搜索框。
  - 新增状态筛选：全部、已完成、失败、生成中、排队中、已取消。
  - 新增尺寸筛选：常用方形、横版、竖版、宽屏、故事版等尺寸。
  - “加载更多”沿用当前筛选条件。
  - 空状态改为“暂无符合条件的历史图片”。
- 自测记录：
  - `pnpm.cmd exec vue-tsc --noEmit`：通过。
- 问题记录：
  - 当前尺寸筛选为前端常用尺寸列表；后续可以改为从 `/api/generation/options` 动态读取。
## 2026-05-07 产品体验优化 Milestone 3.3：历史记录复制 Prompt

- 开发目标：
  - 让用户可以从历史作品快速复用提示词。
- 完成：
  - 历史卡片新增“复制提示词”按钮。
  - 历史详情弹窗新增“复制提示词”按钮。
  - 复制成功显示绿色提示，复制失败显示可操作错误提示。
- 自测记录：
  - `pnpm.cmd exec vue-tsc --noEmit`：通过。
- 问题记录：
  - 复制功能依赖浏览器 Clipboard API；如果浏览器或非安全上下文不支持，会提示用户手动复制。
## 2026-05-07 产品体验优化 Milestone 3.4：再次生成参数带回首页

- 开发目标：
  - 用户可以从历史作品快速带回 Prompt 和尺寸，回到首页手动确认后再次生成。
- 完成：
  - 历史卡片新增“再次生成”按钮。
  - 历史详情弹窗新增“再次生成”按钮。
  - 点击后写入首页草稿 `image_show_generation_draft`，并跳转首页。
  - 不自动创建任务、不自动扣费，仍由用户在首页确认生成。
- 自测记录：
  - `pnpm.cmd exec vue-tsc --noEmit`：通过。
- 问题记录：
  - 历史记录保存的尺寸可能是真实像素尺寸，首页会在加载生成选项后校验；如该尺寸不在当前启用列表，会回退到第一个可用尺寸。
## 2026-05-07 产品体验优化 Milestone 3.5：失败记录原因和重试入口

- 开发目标：
  - 历史失败记录不只显示失败状态，还要告诉用户下一步怎么做。
- 完成：
  - 历史卡片和详情弹窗展示状态中文文案。
  - 失败记录展示友好错误摘要，覆盖超时、502/503、保存失败等常见场景。
  - 失败记录可使用“再次生成”带回首页重试。
  - 避免直接暴露过长上游错误、URL 或密钥信息。
- 自测记录：
  - `go test ./controller -run "TestUserGenerationHistory" -v`：通过。
  - `pnpm.cmd exec vue-tsc --noEmit`：通过。
- 问题记录：
  - 错误分类目前在前端做轻量归类；Milestone 5 监控页会把错误分类后端化，届时可复用统一分类。
## 2026-05-07 产品体验优化 Milestone 3.6：结果区参数快照

- 开发目标：
  - 用户生成完成后能看到本次作品的关键参数，并能继续复用 Prompt。
- 完成：
  - 首页生成完成区域新增 Prompt 摘要。
  - 保留尺寸和积分消耗展示。
  - 新增“复制提示词”按钮，复制当前生成请求使用的完整 Prompt。
  - 复制成功显示轻提示。
- 自测记录：
  - `pnpm.cmd exec vue-tsc --noEmit`：通过。
- 问题记录：
  - 结果区 Prompt 当前以单行摘要展示，避免遮挡图片；完整内容通过复制按钮复用。
## 2026-05-07 产品体验优化 Milestone 3：历史记录资产库第一阶段整体验收

- 范围：
  - 3.1 历史列表搜索和筛选接口。
  - 3.2 历史页筛选 UI。
  - 3.3 复制 Prompt。
  - 3.4 再次生成参数带回首页。
  - 3.5 失败记录原因和重试入口。
  - 3.6 结果区参数快照。
- 整体验收结果：
  - 历史接口支持关键词、状态、尺寸筛选，并保持用户隔离。
  - 历史页支持搜索、筛选、清空筛选和加载更多。
  - 历史卡片和详情支持复制 Prompt。
  - 历史作品可带回首页再次生成，由用户手动确认。
  - 失败记录展示友好原因和重试入口。
  - 首页生成完成区展示 Prompt 摘要、尺寸、消耗积分，并支持复制 Prompt。
- 自测记录：
  - `go test ./controller -run "TestHistory|TestGeneration|TestUserGenerationHistory" -v`：通过。
  - `go test ./...`：通过。
  - `pnpm.cmd exec vue-tsc --noEmit`：通过。
  - `pnpm.cmd build`：通过。沙箱内首次仍因 esbuild `spawn EPERM` 失败，授权后重跑通过。
- 问题记录：
  - 历史尺寸筛选当前为前端常用列表，后续可改为动态读取生成选项。
  - 错误友好分类当前在前端实现，后续 Milestone 5 可改为后端统一分类。
## 2026-05-07 渠道归因与渠道健康统计 1.5：后台渠道页展示统计

- 开发目标：
  - 在后台渠道列表展示每个渠道近 24 小时真实生成成功数、失败数和失败率。
  - 保持“生成统计”和“最近测试”结果分开，避免把手动测试结果误认为真实生成健康度。
- 完成：
  - 后台渠道卡片新增近 24 小时成功、失败和失败率三个统计块。
  - 无统计数据时按 0 和 `0.0%` 展示。
  - 失败率大于 0 时使用红色弱提示，便于管理员快速定位异常渠道。
- 自测记录：
  - `pnpm.cmd exec vue-tsc --noEmit`：通过。
- 问题记录：
  - 浏览器手动查看需要前后端服务运行并有管理员登录态；本次先完成类型检查和模板结构自测，整体验收阶段继续覆盖构建。
## 2026-05-07 渠道归因与渠道健康统计整体验收

- 范围：
  - 1.1 生成记录新增 `channel_id/channel_name` 归因字段。
  - 1.2 渠道调用结果回传实际使用渠道。
  - 1.3 生成成功和最终失败时写入渠道归因。
  - 1.4 后台渠道接口返回近 24 小时成功、失败和失败率。
  - 1.5 后台渠道页展示真实生成健康统计。
- 整体验收结果：
  - 成功生成会记录实际成功渠道。
  - 多渠道全部失败时会记录最后一次实际尝试渠道。
  - Mock 和环境变量兜底渠道有明确渠道名称，不影响旧数据兼容。
  - 后台 `/api/admin/channels` 会按 `channel_id` 聚合近 24 小时 `status=3/4` 的生成记录。
  - 后台渠道页将生成健康统计与手动连通性测试结果分开展示。
- 自测记录：
  - `go test ./service -run "Test.*Generation|Test.*Channel" -v`：通过。
  - `go test ./controller -run "TestAdminChannel|TestGeneration" -v`：通过。
  - `go test ./...`：通过。
  - `pnpm.cmd exec vue-tsc --noEmit`：通过。
  - `pnpm.cmd build`：通过。沙箱内首次仍因 esbuild `spawn EPERM` 失败，授权后重跑通过。
- 问题记录：
  - 当前只统计后台渠道列表中的渠道；`env:SUB2API_BASE_URL` 兜底渠道因没有 `channel_id`，本次不展示为独立渠道卡片。
  - 本次暂不做每次重试的明细表，因此失败归因采用“最后一次实际尝试渠道”，完整链路分析后续可通过 `generation_channel_attempts` 扩展。
## 2026-05-07 顶部账号菜单退出交互修复

- 问题：
  - 生成图片界面顶部账号菜单依赖鼠标悬停展开，鼠标从“管理员”按钮移动到“退出登录”时容易离开悬停区域，导致菜单提前消失。
- 完成：
  - 顶部账号菜单改为点击展开/收起。
  - 点击页面其他位置、路由切换、未授权退出或执行退出登录后会自动关闭菜单。
  - 菜单增加 `aria-expanded`、`aria-haspopup` 和 `role="menu"`，并支持 Esc 关闭。
- 自测记录：
  - `pnpm.cmd exec vue-tsc --noEmit`：通过。
- 问题记录：
  - 本次只修复菜单交互，不改动登录态、角色跳转和后台权限逻辑。
## 2026-05-07 后台公告发布失败提示优化

- 问题：
  - 后台发布公告失败时，前端只显示“操作失败，请检查权限或输入”，无法判断是权限、字段、时间范围还是网络问题。
- 完成：
  - 后台统一操作守卫增加错误提取，兼容后端 `message`、`error`、`detail`、纯文本响应和 HTTP 401/403。
  - 发布公告前增加本地时间校验：开始/结束时间格式不合法会直接提示，结束时间不晚于开始时间会提示“结束时间必须晚于开始时间”。
  - 后端公告测试补充非法时间范围用例，确认接口返回明确 `ends_at must be after starts_at`。
- 自测记录：
  - `go test ./controller -run "TestAnnouncement" -v`：通过。
  - `pnpm.cmd exec vue-tsc --noEmit`：通过。
- 问题记录：
  - 当前没有直接拿到你本机失败请求的响应体；修复后再次发布时，页面会显示真实失败原因，便于继续定位是否是登录态或字段问题。
## 2026-05-07 后台刷新数据容错优化

- 问题：
  - 后台“刷新数据”并发请求用户、积分、模板、设置、渠道、公告和监控 7 个接口；任意一个接口出现网络错误时，整次刷新只显示“请求失败：Network Error”，无法知道是哪一块失败。
- 完成：
  - 刷新数据改为逐项容错加载，成功的模块照常更新，失败模块单独汇总提示。
  - 初始化进入后台时也使用同样的逐项容错加载，避免单个接口失败导致整个后台不可用。
  - 失败提示会显示模块名称，例如“公告：Network Error”或“监控：登录已失效，请重新登录”。
- 自测记录：
  - `http://localhost:3000/health`：返回 `{"status":"ok"}`。
  - `pnpm.cmd exec vue-tsc --noEmit`：通过。
- 问题记录：
  - 本次未直接复现到你的浏览器里的具体失败接口；修复后再次点击刷新，提示会明确失败模块，后续可按模块继续定位代理、登录态或接口问题。
## 2026-05-07 公告接口 301 重定向修复

- 问题：
  - 后台刷新提示“公告：Network Error”。
  - 直接请求旧后端的 `/api/admin/announcements` 和 `/api/announcements` 返回 `301 Location: ./`，不是权限错误；请求没有命中 API 路由，而是掉进前端静态文件兜底。
- 完成：
  - `router/web.go` 增加 `/api/*` 未匹配保护，未知 API 统一返回 JSON：`{"error":"api route not found"}`，不再进入前端静态路由产生 301。
  - 补充测试覆盖未知 API 不应重定向，避免后续新增 API 路由缺失时再次表现为 `Network Error`。
- 自测记录：
  - `go test ./controller -run "TestAnnouncement|TestUnknownAPIRoute" -v`：通过。
  - `go test ./router ./controller -run "Test"`：通过。
  - `pnpm.cmd exec vue-tsc --noEmit`：通过。
- 问题记录：
  - 当前运行中的本地后端需要重启后才能吃到新路由。尝试用后台方式重启时，进程未稳定监听 `:3000`；前台运行可看到新代码已注册 `/api/admin/announcements` 路由。建议用正常终端执行 `go run .` 或重新启动现有后端服务后再测试后台刷新。
## 2026-05-07 后台公告 Network Error 诊断增强

- 问题：
  - 命令行带管理员 token 请求 `http://localhost:5174/api/admin/announcements` 已返回 200，但浏览器后台仍显示“公告：Network Error”，需要区分浏览器实际 origin、请求 URL、axios code 和响应状态。
- 完成：
  - 后台错误详情新增 `url`、`status`、`code`、`origin`。
  - 后台刷新和初始化加载新增“刷新诊断”区域，显示页面来源和失败模块的完整诊断信息。
  - 整页刷新时的用户列表加载不再嵌套 `guarded`，避免并发加载时全局 message 被子任务抢占。
  - 清理重复前端服务，只保留当前测试入口 `http://localhost:5174`，避免 `5174/5180` 混用。
- 自测记录：
  - `pnpm.cmd exec vue-tsc --noEmit`：通过。
  - 命令行带管理员 token 验证 `5174` 下后台刷新 7 个接口均返回 200。
- 问题记录：
  - 如果浏览器仍显示公告 Network Error，需要读取页面“刷新诊断”块里的 `origin/url/status/code` 继续定位；目前接口层和代理层命令行验证均正常。
## 2026-05-07 后台刷新诊断 UI 优化

- 问题：
  - 诊断信息已经能定位问题，但原始展示偏工程调试，不适合长期放在后台管理界面。
- 完成：
  - 刷新诊断改成后台状态条样式。
  - 成功时展示“后台数据已同步”，使用轻量绿色状态。
  - 失败时展示“部分模块需要检查”，失败明细折叠在状态条下方，并提供收起按钮。
  - 保留页面来源和技术细节，便于后续排查浏览器现场问题。
- 自测记录：
  - `pnpm.cmd exec vue-tsc --noEmit`：通过。
  - `go test ./controller -run "TestAnnouncement|TestUnknownAPIRoute" -v`：通过。
  - `go test ./router ./controller -run "Test"`：通过。
## 2026-05-07 生成结果下载操作区 UI 优化

- 问题：
  - 生成图片完成后，右下角“下载 / 下载全部”等操作和底部整条渐变栏占用图片视野，影响看图。
- 完成：
  - 移除覆盖整条底部的黑色渐变操作栏。
  - 生成信息改为左下角轻量信息块，只展示尺寸、积分、提示词摘要和复制提示。
  - 下载、下载全部、复制提示词、再生成一次改为右下角紧凑图标工具条，默认半透明，悬停或键盘聚焦时增强可见性。
  - 移动端“下载全部”保留图标，桌面端显示“全部”文字，减少遮挡面积。
- 自测记录：
  - `pnpm.cmd exec vue-tsc --noEmit`：通过。
- 问题记录：
  - 当前“下载全部”仍复用单图下载逻辑；本次只优化生成结果区视觉与交互位置，不改变下载业务逻辑。
## 2026-05-07 用户个人中心大功能 1 小功能 1.1：路由和入口

- 完成：
  - 新增 `/account` 前端路由。
  - 顶部用户菜单增加“个人中心”入口。
  - 未登录访问 `/account` 会跳转 `/login`。
  - 管理员访问 `/account` 不自动跳转后台，避免用户侧和后台侧混淆。
- 自测记录：
  - 待随大功能 1 前端类型检查统一验证。
- 问题记录：
  - 当前后台入口仍保持 `/console/admin`，不重新展示到右上角。
## 2026-05-07 用户个人中心大功能 1 小功能 1.2：后端概览接口

- 完成：
  - 新增 `GET /api/account/overview`。
  - 接口返回当前用户基础信息、最近 5 条积分流水、最近 6 条生成记录、生成统计、公告未读数量和最近公告。
  - 响应使用手动 DTO，不返回密码哈希、渠道密钥、微信服务端 token 等敏感字段。
  - 补充 account controller 测试，覆盖未登录、当前用户数据隔离、空状态和敏感字段保护。
- 自测记录：
  - `go test ./controller -run "TestAccount" -v`：通过。
- 问题记录：
  - 空状态测试中 GORM 会打印 `record not found` 查询日志，接口实际返回 200 和空数组，不影响功能。
## 2026-05-07 用户个人中心大功能 1 小功能 1.3：页面 UI 骨架

- 完成：
  - 新增 `web/src/views/Account.vue`。
  - 展示头像/首字母兜底、昵称/邮箱前缀兜底、邮箱、身份、账号状态、注册时间、积分余额、积分有效期、最近登录时间和 IP。
  - 增加去生成、图片历史、积分流水、购买积分快捷入口。
  - 扩充前端 user store 类型，支持用户名、头像、积分有效期、最近登录等字段。
- 自测记录：
  - 待随大功能 1 前端类型检查统一验证。
- 问题记录：
  - 第一阶段只读展示，不提供资料编辑；编辑资料放到大功能 2。
## 2026-05-07 用户个人中心大功能 1 小功能 1.4：最近流水与最近作品

- 完成：
  - 个人中心展示最近 5 条积分流水。
  - 个人中心展示最近 6 张作品和生成统计。
  - 无流水、无作品、无公告均有空状态和下一步入口。
  - 最近作品点击进入历史页，第一阶段不做详情弹窗复用。
- 自测记录：
  - 待随大功能 1 前端类型检查统一验证。
- 问题记录：
  - 最近作品使用后端返回的 `image_url`，若旧图 URL 过期，后续可考虑概览接口内刷新 URL；本次先保持轻量查询。
## 2026-05-07 用户个人中心大功能 1：只读概览整体验收

- 完成：
  - `/account` 个人中心只读概览已完成。
  - 后端聚合接口已完成。
  - 前端账户信息、积分与权益、最近作品、安全与通知摘要已完成。
  - 产品方案文档和架构/UI 开发文档已新增。
- 自测记录：
  - `go test ./controller -run "TestAccount" -v`：通过。
  - `go test ./router ./controller -run "Test"`：通过。
  - `pnpm.cmd exec vue-tsc --noEmit`：通过。
- 验收结论：
  - 大功能 1 自动验收通过，可以提交。
- 问题记录：
  - 本轮未启动浏览器做人工截图验收；后续启动前后端后可重点检查 `/account` 桌面端和移动端视觉效果。
## 2026-05-07 用户个人中心大功能 2 小功能 2.1：后端资料更新接口

- 完成：
  - 新增 `PUT /api/account/profile`。
  - 支持更新当前用户自己的昵称和头像 URL。
  - 校验昵称长度、头像 URL 长度和 `http://` / `https://` 前缀。
  - 请求中携带 `role`、`credits`、`email` 等敏感字段不会生效。
  - 补充测试覆盖正常更新、非法头像 URL、未登录访问和敏感字段保护。
- 自测记录：
  - `go test ./controller -run "TestAccount" -v`：通过。
- 问题记录：
  - 第一阶段只支持头像 URL，不支持文件上传；文件上传后续结合 R2 上传策略单独设计。
## 2026-05-07 用户个人中心大功能 2 小功能 2.2：前端资料编辑 UI

- 完成：
  - 个人中心增加“个人资料”编辑模块。
  - 支持昵称和头像 URL 编辑。
  - 支持头像预览和首字母头像兜底。
  - 头像 URL 图片加载失败时会回退首字母头像。
  - 保存成功后同步 `userStore.user`，顶部菜单可同步使用最新用户信息。
  - 保存失败时将后端错误改写为用户友好提示。
- 自测记录：
  - 待随大功能 2 前端类型检查统一验证。
- 问题记录：
  - 暂无。
## 2026-05-07 用户个人中心大功能 2：个人资料编辑整体验收

- 完成：
  - 用户个人资料编辑已完成。
  - 后端资料更新接口和权限校验已完成。
  - 前端昵称、头像 URL、预览、保存状态、错误提示已完成。
- 自测记录：
  - `go test ./controller -run "TestAccount" -v`：通过。
  - `go test ./router ./controller -run "Test"`：通过。
  - `pnpm.cmd exec vue-tsc --noEmit`：通过。
- 验收结论：
  - 大功能 2 自动验收通过，可以提交。
- 问题记录：
  - 未做浏览器手动验收；后续启动前端后重点验证头像 URL 输入、保存成功提示和顶部菜单同步。
## 2026-05-07 用户个人中心大功能 3 小功能 3.1：最近登录方式

- 完成：
  - `GET /api/account/overview` 增加 `security.latest_login`。
  - 后端查询当前用户最近一条成功登录日志。
  - 前端安全与通知模块展示登录时间、登录方式和 IP。
  - 无登录日志时保持空状态文案。
- 自测记录：
  - 待随大功能 3 自动验收统一验证。
- 问题记录：
  - 当前只展示最近一次成功登录，不展示完整设备列表。
## 2026-05-07 用户个人中心大功能 3 小功能 3.2：公告未读摘要

- 完成：
  - 个人中心展示公告未读数量。
  - 个人中心展示最近公告标题。
  - 无公告时展示空状态文案。
  - 后端 overview 保持只返回当前用户可见公告，不返回后台公告管理数据。
- 自测记录：
  - 待随大功能 3 自动验收统一验证。
- 问题记录：
  - 第一阶段不新增公告详情页，继续复用现有顶部公告中心。
## 2026-05-07 用户个人中心大功能 3：安全与通知摘要整体验收

- 完成：
  - 个人中心安全与通知摘要已完成。
  - 后端 overview 返回最近成功登录方式、IP 和时间。
  - 前端展示登录方式、IP、公告未读数量和最近公告标题。
- 自测记录：
  - `go test ./controller -run "TestAccount" -v`：通过。
  - `go test ./router ./controller -run "Test"`：通过。
  - `pnpm.cmd exec vue-tsc --noEmit`：通过。
- 验收结论：
  - 大功能 3 自动验收通过，可以提交。
- 问题记录：
  - 未做浏览器手动验收；后续可在 `/account` 验证有/无公告、有/无登录日志两种状态。
## 2026-05-07 管理后台重设计阶段 A：基础设施层

- 完成：
  - 新增 `web/src/types/admin.ts`，抽取后台用户、积分流水、生成记录、渠道、模板、公告、监控等共享类型。
  - 新增 `web/src/api/admin.ts`，封装后台用户、积分、模板、渠道、设置、监控、公告、生成记录等 API。
  - 新增全局 Toast 基础设施：`useToast.ts` 和 `AppToast.vue`，并在 `App.vue` 挂载。
  - 新增通用 UI 组件：`ConfirmDialog.vue`、`Pagination.vue`、`EmptyState.vue`、`SkeletonCard.vue`。
  - 本阶段只新增共享基础设施，未重构现有 `AdminDashboard.vue` 业务逻辑。
- 自测记录：
  - `pnpm.cmd exec vue-tsc --noEmit`：通过。
  - `pnpm.cmd build`：沙箱内因 esbuild `spawn EPERM` 失败；提升权限后通过。
- 验收结论：
  - 阶段 A 自动验收通过，可以提交。
- 问题记录：
  - 当前项目未配置 ESLint 脚本，因此未执行 `npm run lint`。
## 2026-05-07 管理后台重设计阶段 B：布局组件预置

- 完成：
  - 新增 `AdminLayout.vue`，包含管理员权限守卫、当前 Tab 状态和新布局内容区。
  - 新增 `AdminSidebar.vue`，支持桌面端左侧固定导航和移动端横向滚动导航。
  - 侧边栏包含概览、用户、渠道、模板、设置、公告、积分、监控 8 个入口，并显示管理员邮箱和返回前台入口。
- 自测记录：
  - `pnpm.cmd exec vue-tsc --noEmit`：通过。
- 验收结论：
  - B1/B2 自动验收通过。
- 问题记录：
  - 原计划 B3 会把 `AdminDashboard.vue` 改成新布局入口，并接受“原有所有功能暂时不可用”。考虑当前后台仍在使用中，本轮暂不替换入口，避免主分支后台功能中断；等阶段 C 的 Tab 功能迁移完成后再切换入口。
## 2026-05-07 管理后台重设计阶段 C1：概览页

- 完成：
  - 新增 `OverviewTab.vue`。
  - 概览页展示今日生成、新增用户、积分消耗、启用渠道 4 张指标卡。
  - 增加渠道状态列表、今日监控摘要、检查告警按钮和快捷操作区。
  - 数据加载时使用 `SkeletonCard`，无渠道时使用 `EmptyState`。
  - 概览页已接入预置 `AdminLayout`，但当前还未替换线上后台入口。
- 自测记录：
  - `pnpm.cmd exec vue-tsc --noEmit`：通过。
- 验收结论：
  - C1 自动验收通过，可以提交。
- 问题记录：
  - 由于新布局尚未切到 `/console/admin`，本轮未做浏览器实际页面验收；后续阶段 C 全量迁移后统一切换并回归。
## 2026-05-07 管理后台重设计阶段 C2：用户管理页

- 完成：
  - 新增 `UsersTab.vue`。
  - 支持用户搜索、分页、桌面表格和移动端卡片布局。
  - 支持新建用户弹窗。
  - 支持查看用户生成记录弹窗、人工充值弹窗。
  - 用户禁用/启用、角色调整改用 `ConfirmDialog` 确认。
  - 操作结果使用全局 Toast 反馈。
  - C2 已接入预置 `AdminLayout`，但当前还未替换线上后台入口。
- 自测记录：
  - `pnpm.cmd exec vue-tsc --noEmit`：通过。
- 验收结论：
  - C2 自动验收通过，可以提交。
- 问题记录：
  - 新布局尚未切到 `/console/admin`，本轮未做浏览器实际页面验收。
## 2026-05-07 管理后台重设计阶段 C3：渠道管理页

- 完成：
  - 新增 `ChannelsTab.vue`。
  - 支持渠道列表、状态卡片、最近测试状态、失败率展示。
  - 支持新增/编辑渠道弹窗。
  - 支持渠道测试并通过 Toast 反馈结果。
  - 删除渠道改用 `ConfirmDialog` 确认。
  - C3 已接入预置 `AdminLayout`，但当前还未替换线上后台入口。
- 自测记录：
  - `pnpm.cmd exec vue-tsc --noEmit`：通过。
- 验收结论：
  - C3 自动验收通过，可以提交。
- 问题记录：
  - 新布局尚未切到 `/console/admin`，本轮未做浏览器实际页面验收。
## 2026-05-07 管理后台重设计阶段 C4：模板管理页

- 完成：
  - 新增 `TemplatesTab.vue`。
  - 支持模板列表、分类筛选、状态展示和排序展示。
  - 支持新增/编辑模板弹窗。
  - 删除模板改用 `ConfirmDialog` 确认。
  - 操作结果使用全局 Toast 反馈。
  - C4 已接入预置 `AdminLayout`，但当前还未替换线上后台入口。
- 自测记录：
  - `pnpm.cmd exec vue-tsc --noEmit`：通过。
- 验收结论：
  - C4 自动验收通过，可以提交。
- 问题记录：
  - 新布局尚未切到 `/console/admin`，本轮未做浏览器实际页面验收。
## 2026-05-07 管理后台重设计阶段 C5：系统设置页

- 完成：
  - 新增 `SettingsTab.vue`。
  - 设置项按账号与额度、微信登录、图像生成、图片存储、人机验证、安全与监控分组。
  - 敏感项支持显示/隐藏。
  - 常见配置增加说明和示例。
  - 支持保存设置并通过 Toast 反馈结果。
  - C5 已接入预置 `AdminLayout`，但当前还未替换线上后台入口。
- 自测记录：
  - `pnpm.cmd exec vue-tsc --noEmit`：通过。
- 验收结论：
  - C5 自动验收通过，可以提交。
- 问题记录：
  - 设置保存仍沿用现有接口的一次性保存策略；后续如需更强安全性，可补二次确认或 diff 预览。
## 2026-05-07 管理后台重设计阶段 C6：积分流水页

- 完成：
  - 新增 `CreditsTab.vue`。
  - 支持积分流水列表、中文类型文案、变动金额颜色区分。
  - 支持桌面表格和移动端卡片布局。
  - 支持分页和刷新。
  - C6 已接入预置 `AdminLayout`，但当前还未替换线上后台入口。
- 自测记录：
  - `pnpm.cmd exec vue-tsc --noEmit`：通过。
- 验收结论：
  - C6 自动验收通过，可以提交。
- 问题记录：
  - 当前按现有接口只做全局流水分页，暂未增加用户 ID、类型、时间筛选。
## 2026-05-07 管理后台重设计阶段 C7：监控告警页

- 完成：
  - 新增 `MonitorTab.vue`。
  - 展示生成总数、完成数、失败数、失败率、积分消耗、新增用户、支付订单和支付金额。
  - 展示告警状态、失败原因聚合、最近失败任务。
  - 支持刷新和手动检查告警。
  - C7 已接入预置 `AdminLayout`，但当前还未替换线上后台入口。
- 自测记录：
  - `pnpm.cmd exec vue-tsc --noEmit`：通过。
- 验收结论：
  - C7 自动验收通过，可以提交。
- 问题记录：
  - 新布局尚未切到 `/console/admin`，本轮未做浏览器实际页面验收。
## 2026-05-07 管理后台重设计阶段 C 补充：公告通知页

- 完成：
  - 新增 `AnnouncementsTab.vue`。
  - 支持公告列表、已读数、目标用户、状态展示。
  - 支持发布/编辑公告弹窗。
  - 支持静默/弹窗通知、目标范围、时间范围、排序和状态配置。
  - 删除公告改用 `ConfirmDialog` 确认。
  - 补齐公告已读用户列表弹窗。
  - 补齐原开发计划遗漏但旧后台已有的公告功能，避免新后台切换后功能缺失。
- 自测记录：
  - `pnpm.cmd exec vue-tsc --noEmit`：通过。
- 验收结论：
  - 公告 Tab 自动验收通过，可以提交。
- 问题记录：
  - 新布局尚未切到 `/console/admin`，本轮未做浏览器实际页面验收。
## 2026-05-07 管理后台重设计阶段 D：切换新后台入口

- 完成：
  - `AdminDashboard.vue` 已改为新 `AdminLayout` 薄壳。
  - 新后台入口已覆盖概览、用户、渠道、模板、设置、公告、积分、监控。
  - 旧后台单文件巨组件已移除。
- 自测记录：
  - `pnpm.cmd exec vue-tsc --noEmit`：通过。
  - `go test ./router ./controller -run "Test"`：通过。
  - `pnpm.cmd build`：提升权限后通过。
- 验收结论：
  - 管理后台重设计自动验收通过，已可以提交。
- 问题记录：
  - 尚未启动浏览器进行人工 UI 验收；后续建议重点检查 `/console/admin` 下 8 个 Tab 的实际接口数据、弹窗层级和移动端布局。
