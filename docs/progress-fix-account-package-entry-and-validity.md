# 个人中心购买入口与套餐时长修复记录

日期：2026-05-08

## 问题

1. 进入个人中心时，顶部“购买积分”入口使用深色主按钮样式，视觉上像被置黑选中。
2. 套餐 `valid_days` 是“时长”语义，但支付到账逻辑在用户已有更晚 `credits_expiry` 时只保留原到期时间，没有叠加新套餐时长。

## 修复

- 移除个人中心顶部“购买积分”的 `account-action-primary` 样式，保持和其他入口一致。
- 删除不再使用的 `account-action-primary` CSS。
- 修正支付到账有效期逻辑：
  - 如果用户当前积分未过期，则从当前 `credits_expiry` 继续叠加套餐 `valid_days`。
  - 如果用户积分已过期或没有到期时间，则从当前时间叠加套餐 `valid_days`。
- 新增测试 `TestPaymentNotifyExtendsExistingCreditsExpiry`，验证已有有效期会继续叠加套餐天数。

## 自测记录

- `gofmt -w service/payment.go controller/order_test.go`：通过。
- `go test ./controller -run "TestPaymentNotify|TestCreateOrder|TestExpiredOrder" -v`：通过。
- `pnpm.cmd exec vue-tsc --noEmit`：通过。

## 验收结论

修复通过。套餐时长逻辑现在与“购买套餐有效期”语义自洽。
