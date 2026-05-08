# 大功能 D 个人资料与头像上传验收记录

日期：2026-05-08

## 范围

- D1 头像上传后端接口。
- D2 后台头像存储配置 UI。
- D3 个人中心头像上传 UI。
- D4 头像 R2 迁移预留方案文档。

## 验收结果

- 登录用户可以通过 `POST /api/account/avatar` 上传头像。
- 未登录用户无法上传头像。
- 后端限制头像格式为后台配置允许的类型，默认 `jpg,jpeg,png,webp`。
- 后端限制头像大小，默认 2MB，配置异常时回退默认值，并设置最高 10MB 保护。
- 头像第一版保存到本地 `uploads/avatars`。
- `/uploads` 静态路由可访问本地头像文件。
- 上传成功后回写当前用户 `avatar_url`。
- 后台设置页可以配置头像存储方式、最大大小和允许格式，并明确当前只支持 `local`。
- 个人中心头像 UI 已改为文件上传，上传成功后同步个人中心和全局用户状态。
- R2 后续迁移方案已记录，覆盖预检查、一键迁移、结果记录和回滚。

## 自测命令

- `go test ./controller -run "TestAccount|TestAdminPromptTemplateCRUDAndSettings" -v`：通过。
- `pnpm.cmd exec vue-tsc --noEmit`：通过。

## 验收结论

大功能 D 自动化自测通过，可以进入后续功能开发。
