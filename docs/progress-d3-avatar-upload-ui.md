# D3 个人中心头像上传 UI 进度记录

日期：2026-05-08

## 开发目标

- 将个人中心头像 URL 输入改为文件上传。
- 上传成功后同步个人中心头像和顶部用户状态。

## 完成内容

- 个人资料页新增头像文件选择控件。
- 支持选择 jpg、jpeg、png、webp 图片。
- 调用 `POST /api/account/avatar` 上传头像。
- 上传成功后同步：
  - `profileForm.avatar_url`
  - `overview.user`
  - `userStore.user`
- 上传失败时展示友好提示：
  - 文件过大。
  - 格式不支持。
  - 其他后端错误。
- 上传过程中禁用上传和保存按钮，避免重复操作。
- `docs/plan-admin-site-account-ops.md` 进度表已标记 D3 完成。

## 自测记录

- `pnpm.cmd exec vue-tsc --noEmit`：通过。

## 验收结论

D3 局部自测通过，可以提交。
