# E2 最近作品预览限制进度记录

日期：2026-05-08

## 开发目标

- 个人中心概览最多只预览 3 个最近作品。
- 完整图片历史通过二级页面查看。

## 完成内容

- `GET /api/account/overview` 的 `creations.recent_items` 从最多 6 条调整为最多 3 条。
- 保留 `creations.total`、`completed`、`failed` 统计，默认页仍可展示摘要。
- 新增后端测试覆盖 5 条生成记录时只返回 3 条预览。
- `docs/plan-admin-site-account-ops.md` 进度表已标记 E2 完成。

## 自测记录

- `gofmt -w controller/account.go controller/account_test.go`：通过。
- `go test ./controller -run "TestAccountOverview" -v`：通过。
- `pnpm.cmd exec vue-tsc --noEmit`：通过。

## 验收结论

E2 局部自测通过，可以提交。
