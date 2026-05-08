# 头像 R2 迁移预留方案

日期：2026-05-08

## 当前状态

- 当前头像上传使用本地存储。
- 文件保存目录：`uploads/avatars`。
- 用户表保存访问地址：`users.avatar_url`，例如 `/uploads/avatars/1-xxx.png`。
- 后台配置 `avatar_storage_driver` 当前仅使用 `local`。

## 后续启用 R2 的目标

- 将已有本地头像上传到 R2。
- 上传成功后批量回写用户 `avatar_url` 为 R2 公网访问地址。
- 迁移失败的用户记录错误，不影响已经成功迁移的用户。
- 前端不需要手动刷新头像来源，只依赖 `/api/auth/me` 和 `/api/account/overview` 返回的最新 `avatar_url`。

## 建议后台能力

### 1. 迁移预检查

- 检查 R2 配置是否完整：
  - `r2_endpoint`
  - `r2_access_key`
  - `r2_secret_key`
  - `r2_bucket`
  - `r2_public_url`
- 检查本地 `uploads/avatars` 是否存在。
- 统计待迁移用户数量：
  - `avatar_url` 以 `/uploads/avatars/` 开头。
  - 本地文件存在。

### 2. 一键迁移任务

- 后台按钮：“上传全部本地头像到 R2”。
- 后端逐个用户处理：
  - 读取本地头像文件。
  - 按稳定 key 上传到 R2，例如 `avatars/{user_id}/{filename}`。
  - 生成公网 URL：`{r2_public_url}/avatars/{user_id}/{filename}`。
  - 回写 `users.avatar_url`。
- 每个用户独立事务或独立更新，避免单个失败导致全部回滚。

### 3. 迁移结果记录

- 建议新增迁移结果响应：
  - `total`
  - `success`
  - `failed`
  - `skipped`
  - `errors`
- `errors` 至少包含：
  - `user_id`
  - `avatar_url`
  - `reason`

### 4. 回滚策略

- 迁移前导出用户 ID 和旧 `avatar_url`。
- 如果 R2 公网访问异常，可以按导出数据批量回写本地 URL。
- 不建议迁移完成后立即删除本地文件，至少保留一个发布周期。

## 风险点

- R2 公网域名或缓存配置错误会导致头像无法显示。
- 本地文件缺失时，只能跳过并记录错误。
- 用户迁移过程中重新上传头像时，可能出现旧任务覆盖新头像，需要在更新前确认用户当前 `avatar_url` 仍等于任务开始时的本地 URL。

## 验收标准

- 迁移前能看到待迁移数量。
- 迁移后用户头像地址变为 R2 公网 URL。
- `/api/auth/me` 和 `/api/account/overview` 返回新头像地址。
- 失败用户有明确错误记录。
- 本地文件未被立即删除，可回滚。
