# E5 购买中心接入后台套餐进度记录

日期：2026-05-08

## 开发目标

- 个人中心购买中心读取后台套餐。
- 没有真实支付渠道时，不误导用户自动支付成功。

## 完成内容

- `/packages` 页面已读取 `GET /api/packages` 展示后台启用套餐。
- `/packages` 页面已读取 `GET /api/support/contact` 展示人工充值联系方式。
- 后台套餐管理由 B2 完成，可新增、编辑、启停、删除套餐。
- 购买中心不再创建 `pay_method: alipay` 的自动支付订单。
- `docs/plan-admin-site-account-ops.md` 进度表已标记 E5 完成。

## 自测记录

- `go test ./controller -run "TestPackage|TestOrder" -v`：通过。
- `pnpm.cmd exec vue-tsc --noEmit`：通过。

## 验收结论

E5 局部自测通过，可以提交。
