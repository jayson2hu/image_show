# 开发进度记录

## 2026-04-30 Figma Make 首页界面

- 使用 Figma MCP 读取 `https://www.figma.com/make/5IFwPGoEStpt4u4DQHl2FZ/AI-Image-Generation-Website?t=e3ONudqa0VWtSlY1-1`。
- Figma Make 返回源码资源，核心布局来自 `src/app/App.tsx`：顶部导航、左侧 420px 参数面板、右侧生成预览区、紫蓝渐变主按钮、风格预设、推荐示例和高级参数滑块。
- 已在 `web/src/views/Home.vue` 实现对应布局，并保留现有后端生成接口、Turnstile 验证码、SSE 进度监听、取消、重试和用户积分刷新逻辑。
- 已修复首页相关组件中的中文显示问题：`App.vue`、`GenerationProgress.vue`、`PromptTags.vue`、`ImagePreview.vue`。
- 已补充 Figma 风格滑块样式到 `web/src/assets/main.css`。

## 自测记录

- `pnpm.cmd build`：通过。首次在沙箱内因 esbuild `spawn EPERM` 失败，已在授权后重新执行并通过。
- 后端健康检查：`http://localhost:3000/health` 返回 `{"status":"ok"}`。
- 前端开发服务：已启动在 `http://localhost:5180`，未占用用户说明的 `5173`。

## 问题与说明

- Figma MCP 的 `get_screenshot` 不支持 Figma Make 文件，因此本次依据 Make 源码资源分析布局实现。
- 当前后端生成接口仅支持单图任务，界面中的“图片数量 / 创造力 / 步数 / CFG Scale”为 Figma 对齐展示参数，暂未扩展后端协议。

## 2026-04-30 Figma Make 首页新版

- 重新使用 Figma MCP 读取同一 Make 链接，确认新版核心变化：品牌为 `ArtifyAI`，左侧为预览/结果区域，右侧为 420px 控制面板。
- 已将首页布局从“左控制、右预览”调整为“左预览、右控制”。
- 已按新版实现可折叠高级参数、可折叠推荐样例、实心紫色风格选中态、空状态文案和结果悬浮下载样式。
- 已同步全局顶部品牌区为 `ArtifyAI / AI 艺术创作平台`，并保留套餐、历史、管理、登录和退出入口。

## 自测记录

- `pnpm.cmd build`：通过。沙箱内仍因 esbuild `spawn EPERM` 失败，授权后重新执行通过。
- 前端访问检查：`http://localhost:5180` 返回 200。
- 后端健康检查：`http://localhost:3000/health` 返回 `{"status":"ok"}`。

## 2026-04-30 UI 问题修复

- 首页“高级参数”和“推荐样例”默认改为折叠，仅点击后展开。
- 登录页和注册页补齐深色模式下的卡片、文字、输入框、按钮和错误提示对比样式，避免深色模式切换后文字不可见。

## 自测记录

- `pnpm.cmd build`：通过。沙箱内仍因 esbuild `spawn EPERM` 失败，授权后重新执行通过。
- 前端访问检查：`http://localhost:5180` 返回 200。
- 后端健康检查：`http://localhost:3000/health` 返回 `{"status":"ok"}`。
