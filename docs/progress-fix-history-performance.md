# 历史记录加载慢优化记录

## 背景

用户反馈历史记录中生成图片加载较慢，并询问生成图片存在哪里。

## 存储说明

生成记录存储在数据库 `generations` 表：

- `prompt`、`size`、`status`、`credits_cost` 等元数据存在 `generations` 表。
- `r2_key` 保存 Cloudflare R2 对象路径。
- `image_url` 保存图片访问地址。

图片本体的存储逻辑在 `service/r2.go`：

- 配置了 R2 时：图片上传到 Cloudflare R2，数据库保存 `r2_key` 和临时访问 URL。
- 未配置 R2 时：后端会退化为把图片转成 `data:image/...;base64,...` 写入 `generations.image_url`。

未配置 R2 的情况下，历史列表如果直接返回 `image_url`，一页多张图片会把大段 base64 放进 JSON，导致接口响应、JSON 解析和前端渲染都变慢。

## 修复拆分

### 小功能 1：历史列表接口轻量化

状态：已完成

改动：

- `GET /api/generations` 不再直接返回完整 `model.Generation`。
- 列表查询只选择历史页需要的字段。
- 当图片是 R2 对象或本地 data URL 时，列表 `image_url` 返回轻量代理地址 `/api/generations/:id/image`。
- 保持前端字段名兼容：`id`、`prompt`、`quality`、`size`、`status`、`image_url`、`error_msg`、`created_at`。

自测：

- `go test ./controller -run "TestUserGenerationHistory" -v`：通过。

### 小功能 2：增加历史图片读取接口

状态：已完成

改动：

- 新增 `GET /api/generations/:id/image`。
- 用户只能读取自己的、未删除的生成图片。
- 如果图片存 R2，接口刷新临时链接并重定向。
- 如果图片是 data URL，接口解码后直接返回图片二进制，避免列表接口携带大 JSON。

自测：

- `go test ./controller -run "TestUserGenerationHistory" -v`：通过。

### 小功能 3：历史页前端图片加载优化

状态：已完成

改动：

- 缩略图增加 `loading="lazy"` 和 `decoding="async"`。
- 打开详情时先使用当前列表项显示弹窗，再后台刷新详情数据。
- 因 `<img>` 不能自动携带 Bearer Token，历史页会用 Axios 带鉴权请求 `/api/generations/:id/image`，再转换成 object URL 展示缩略图。
- 详情图和下载按钮也复用 object URL，避免鉴权图片地址在浏览器原生请求里缺少 token。
- 页面刷新筛选或离开时释放 object URL，避免内存累积。

自测：

- `pnpm.cmd exec vue-tsc --noEmit`：通过。

## 验收标准

- 历史列表接口不再返回完整 base64 图片数据：已通过测试覆盖。
- 历史图片仍能正常显示、查看详情和下载：图片读取接口已通过测试覆盖，前端显示和下载复用鉴权后的 object URL。
- 图片接口有登录鉴权和用户归属校验：已实现。
- 前端类型检查通过：已通过。

## 问题记录

- 未发现需求矛盾。
- 如果生产环境没有配置 R2，历史图片仍会占用数据库空间；这次优化解决“列表慢”，但长期建议配置 R2 存储图片本体。
