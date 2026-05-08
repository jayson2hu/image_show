# B4 前台购买中心展示人工充值联系方式进度记录

日期：2026-05-08

## 开发目标

- 前台购买中心读取后台配置的人工充值联系方式。
- 第一版不接真实支付，不再让用户误以为点击后会自动完成支付。

## 完成内容

- `/api/support/contact` 公开返回人工充值展示字段：
  - `manual_recharge_enabled`
  - `manual_recharge_wechat_id`
  - `manual_recharge_wechat_qrcode_url`
  - `manual_recharge_qq`
  - `manual_recharge_note`
- `/packages` 页面同时读取套餐和人工充值联系方式。
- 套餐卡片支持选择套餐，右侧展示当前已选套餐和人工充值说明。
- 微信二维码支持图片预览。
- 微信号和 QQ 支持复制。
- 未登录用户选择套餐时跳转登录，避免未登录状态下进入充值操作。
- 去掉原先 `POST /api/orders` + `pay_method: alipay` 的自动支付跳转行为，避免真实支付未接入时误导用户。
- `docs/plan-admin-site-account-ops.md` 进度表已标记 B4 完成。

## 自测记录

- `gofmt -w controller/support.go controller/admin_template_setting_test.go`：通过。
- `go test ./controller -run "TestAdminPromptTemplateCRUDAndSettings|TestPackage" -v`：通过。
- `pnpm.cmd exec vue-tsc --noEmit`：通过。

## 验收结论

B4 局部自测通过，可以提交。
