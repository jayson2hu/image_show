# 微信公众号验证码登录配置与验收

## 当前登录流程

本项目采用 new-api 风格的“公众号验证码换 OpenID”模式，不直接接微信公众号 OAuth 回调。

浏览器流程：

1. 用户进入 `/login`。
2. 点击“获取验证码”。
3. 页面弹出公众号二维码。
4. 用户扫码关注公众号。
5. 公众号向用户返回一串验证码。
6. 用户把验证码填入登录页。
7. 前端请求 `GET /api/auth/wechat/callback?code=验证码`。
8. 后端请求外部 WeChat Server：`GET {wechat_server_address}/api/wechat/user?code=验证码`。
9. WeChat Server 返回 OpenID 后，系统完成登录；首次使用会自动创建用户并发放后台配置的新用户积分。

## 管理后台配置位置

进入后台：

```text
/console/admin
```

在“设置”页维护以下配置：

| 配置项 | 后台名称 | 说明 |
| --- | --- | --- |
| `wechat_auth_enabled` | 微信登录开关 | 设置为“开启”后，普通登录页才允许微信验证码登录。 |
| `wechat_qrcode_url` | 微信登录二维码 | 展示给用户扫码关注的公众号二维码。可以填写图片 URL，也可以直接选择本地图片上传保存。 |
| `wechat_server_address` | 微信服务地址 | 外部 WeChat Server 地址，不带末尾路径，例如 `https://wechat.example.com`。 |
| `wechat_server_token` | 微信服务 Token | 后端请求 WeChat Server 时写入 `Authorization` 请求头。 |
| `register_gift_credits` | 注册赠送积分 | 首次微信验证码登录自动创建用户时发放的积分。 |

配置保存后立即生效，后台设置优先级高于环境变量。

## 环境变量兜底配置

如果还没有进入后台配置，也可以用环境变量作为兜底：

```env
WECHAT_AUTH_ENABLED=true
WECHAT_QRCODE_URL=
WECHAT_SERVER_ADDRESS=https://wechat.example.com
WECHAT_SERVER_TOKEN=change-me
```

`WECHAT_QRCODE_URL` 可以为空，之后在后台上传二维码图片即可。

## WeChat Server 接口要求

项目后端会请求：

```http
GET /api/wechat/user?code=用户输入的验证码
Authorization: 后台配置的 wechat_server_token
```

成功响应必须返回：

```json
{
  "success": true,
  "data": "openid_xxx"
}
```

失败响应可以返回：

```json
{
  "success": false,
  "message": "code expired"
}
```

说明：

- `data` 必须是稳定的微信 OpenID 或等价唯一用户标识。
- 同一个 OpenID 再次登录会复用原账号，不会重复创建用户。
- 外部服务返回非 2xx、`success=false` 或 `data` 为空时，前端会提示“微信验证码无效或已过期”。

## 本地验收方式

后端单元测试已经用模拟 WeChat Server 覆盖核心链路：

```powershell
go test ./controller -run "TestWeChatQRCodeAndLoginCreatesUser|TestWeChatInvalidCodeFromServer|TestWeChatBindAndUnbind|TestWeChatDisabled" -v
```

完整自测：

```powershell
go test ./...
cd web
pnpm.cmd exec vue-tsc --noEmit
pnpm.cmd build
```

## 人工验收步骤

1. 管理员访问 `/console/admin`，进入“设置”页。
2. 开启“微信登录开关”。
3. 配置“微信服务地址”和“微信服务 Token”。
4. 在“微信登录二维码”选择公众号二维码图片并保存。
5. 普通用户访问 `/login`。
6. 点击“获取验证码”，确认弹出公众号二维码。
7. 手机扫码关注公众号，复制公众号返回的验证码。
8. 在页面输入验证码，点击“登录 / 注册”。
9. 首次登录应自动创建用户并获得 `register_gift_credits` 配置的积分。
10. 同一个微信再次登录，应进入同一个账号，不应重复创建用户。

## 常见问题

- 点“获取验证码”提示微信登录未开启：检查 `wechat_auth_enabled` 是否开启。
- 二维码不显示：检查 `wechat_qrcode_url` 是否已保存；没有图片地址时可直接在后台选择本地图片。
- 输入验证码后提示无效：检查 WeChat Server 是否能通过 `/api/wechat/user?code=...` 返回 `success=true` 和稳定 OpenID。
- 后台配置了环境变量但不生效：检查后台设置表是否已有同名配置，后台保存值会覆盖环境变量。
