# B3 人工充值联系方式配置进度记录

日期：2026-05-08

## 开发目标

- 后台新增独立的人工充值联系方式配置，第一版只支持联系管理员充值，不接真实支付渠道。
- 人工充值配置不复用图片生成渠道，避免把生成服务渠道误显示为支付渠道。

## 完成内容

- `GET /api/admin/settings` 新增返回：
  - `manual_recharge_enabled`
  - `manual_recharge_wechat_id`
  - `manual_recharge_wechat_qrcode_url`
  - `manual_recharge_qq`
  - `manual_recharge_note`
- 后端设置默认值增加人工充值配置项，老数据库无需额外迁移。
- 后台设置页新增“人工充值”分组，可配置启用开关、微信号、微信二维码 URL、QQ 和充值说明。
- 充值说明使用多行输入，帮助文案明确“第一版不是自动支付”，减少用户误解。
- `docs/plan-admin-site-account-ops.md` 进度表已标记 B3 完成。

## 自测记录

- `gofmt -w controller/admin_template_setting.go controller/admin_template_setting_test.go`：通过。
- `go test ./controller -run "TestAdminPromptTemplateCRUDAndSettings" -v`：通过。
- `pnpm.cmd exec vue-tsc --noEmit`：通过。

## 验收结论

B3 局部自测通过，可以提交并继续 B4。
