# D1 头像上传后端接口进度记录

日期：2026-05-08

## 开发目标

- 新增登录用户头像文件上传接口。
- 第一版本地保存头像，不上传 R2。
- 上传后更新当前用户 `avatar_url`，让 `/api/auth/me` 和 `/api/account/overview` 后续能读取最新头像地址。

## 完成内容

- 新增 `POST /api/account/avatar`。
- 上传字段名为 `avatar`，使用 `multipart/form-data`。
- 上传权限限制为登录用户，只能更新当前登录用户自己的头像。
- 本地保存目录为 `uploads/avatars`。
- 返回头像 URL 格式为 `/uploads/avatars/{filename}`。
- 新增 `/uploads` 静态访问路由，用于访问本地头像文件。
- 上传限制从后台设置读取：
  - `avatar_storage_driver`，默认 `local`。
  - `avatar_max_size_mb`，默认 `2`，上限保护为 `10`。
  - `avatar_allowed_types`，默认 `jpg,jpeg,png,webp`。
- 后台设置接口补齐头像配置默认键，老数据库无需额外迁移。

## 自测记录

- `gofmt -w controller/account.go controller/account_test.go controller/admin_template_setting.go router/main.go router/web.go`：通过。
- `go test ./controller -run "TestAccount.*Avatar|TestAccountProfile|TestAccountOverview" -v`：通过。
- `go test ./controller -run "TestAdminPromptTemplateCRUDAndSettings|TestAccount.*Avatar" -v`：通过。

## 验收结论

D1 局部自测通过，可以提交。
