# 2026-05-14 需求清理、功能核验与文档收敛

## 目标

- 清理旧待办快照、早期验收清单和当前完成进度之间的口径冲突。
- 验证后台、站点配置、注册策略、套餐、人工充值、个人中心、头像上传、场景入口、再次生成等需求是否真实开发。
- 形成当前可继续推进的剩余事项列表。

## 结论

- 已新增 `docs/requirements-cleanup-summary.md`，作为当前需求状态、验收结果和剩余风险的集中说明。
- `docs/pending-tasks-review.html` 是旧待办快照，不再作为当前开发依据。
- 主线功能均有代码入口和测试覆盖：`/api/account/overview`、`/api/account/profile`、`/api/account/avatar`、`/api/site/config`、`/api/generation/scenes`、`/api/support/contact`、`/api/packages`、`/api/admin/packages`。
- 当前明确未开发或未纳入本轮的是：请求签名强制校验、完整登录设备列表、公告详情页、用户侧订单列表、修改邮箱/密码、个人中心微信绑定管理、收藏作品、用户偏好、注销账号、历史全文索引和后端统一错误分类。
- 当前已开发但仍需外部环境验收的是：R2、Redis、SMTP、Turnstile、支付、微信登录和浏览器人工 UI。

## 验证命令

```powershell
go test ./controller -run "TestAccount|TestRegister|TestAdminPromptTemplateCRUDAndSettings|Test.*Package|TestGenerationScenes|TestSiteConfig" -v
go test ./service -v
go test ./...
cd web; pnpm.cmd exec vue-tsc --noEmit
cd web; pnpm.cmd build
```

结果：全部通过。
