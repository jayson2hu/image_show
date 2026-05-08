# D2 后台头像存储配置 UI 进度记录

日期：2026-05-08

## 开发目标

- 后台设置页增加头像存储配置入口。
- 第一版只开放本地头像存储配置，不误导为已经接入 R2。

## 完成内容

- 后台设置页新增“头像存储”分组。
- 支持配置：
  - `avatar_storage_driver`，当前只支持 `local`。
  - `avatar_max_size_mb`，头像最大大小。
  - `avatar_allowed_types`，允许上传的头像格式。
- 增加说明文案：
  - 当前头像保存到后端 `uploads/avatars`。
  - 后续启用 R2 前需要先执行迁移。
- `docs/plan-admin-site-account-ops.md` 进度表已标记 D2 完成。

## 自测记录

- `pnpm.cmd exec vue-tsc --noEmit`：通过。

## 验收结论

D2 局部自测通过，可以提交。
