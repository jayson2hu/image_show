# 渠道归因与渠道健康统计开发计划

> 计划日期：2026-05-07  
> 关联问题：Milestone 5.3 渠道近 24 小时成功/失败聚合缺少 `channel_id`，无法把生成任务准确归因到具体渠道。  
> 目标：在不破坏现有生成流程的前提下，记录每次生成实际使用的渠道，并在后台渠道页展示近 24 小时成功、失败和失败率。

## 开发原则

1. 每个小功能完成后先做局部自测，并写入 `docs/progress.md`。
2. 大功能完成后做整体自测验收、提交 Git，并推送到 GitHub。
3. 不对历史生成记录做猜测回填，避免错误归因。
4. 先实现“单字段归因版”，记录最终实际使用渠道；暂不做每次重试明细表。
5. 如果开发中发现生成流程无法可靠识别最终渠道，需要暂停确认，不自行改统计口径。

## 业务口径

### 统计对象

- 只统计 `generations` 表中有渠道归因的记录。
- 统计窗口为后台当前服务器时间的最近 24 小时。
- 成功数：`status = 3`。
- 失败数：`status = 4`。
- 失败率：`failed / (completed + failed)`，分母为 0 时显示 0 或暂无统计。

### 渠道归因口径

- 任务成功：记录实际成功返回图片结果的渠道。
- 任务最终失败：记录最后一次实际尝试的渠道。
- 环境变量兜底渠道：`channel_id = null`，`channel_name = "env:SUB2API_BASE_URL"`。
- 历史老数据：`channel_id = null` 且 `channel_name = ""`，后台显示“暂无渠道归属”，不计入单个渠道统计。

## 数据模型设计

### `generations` 表新增字段

- `channel_id`：可空整型，记录后台渠道 ID。
- `channel_name`：字符串，记录渠道名称快照。

设计理由：

- `channel_id` 用于准确聚合当前渠道。
- `channel_name` 用于历史展示，即使渠道后续改名或删除，旧记录仍可读。
- 可空字段兼容旧数据和环境变量兜底渠道。

### `channels` 表不新增统计字段

近 24 小时统计通过查询 `generations` 聚合得出，不写回 `channels`，避免缓存过期和一致性问题。

## 小功能拆分

### 1.1 生成记录渠道字段扩展

开发范围：

- 在 `model.Generation` 新增 `ChannelID *int64` 和 `ChannelName string`。
- 确认 `AutoMigrate` 能自动加列。
- 不改历史数据。

局部验收：

- 新建测试数据库后 `generations` 可迁移成功。
- 旧测试不因新增字段失败。

局部自测：

- `go test ./model -v`
- `go test ./...`

文档更新：

- 在 `docs/progress.md` 记录字段扩展和兼容策略。

### 1.2 生成渠道选择结果回传

开发范围：

- 调整 `service.GenerateImage` / `service.GenerateImageEdit` 相关调用链。
- 让渠道调用结果返回渠道元信息：
  - `channel_id`
  - `channel_name`
  - 是否为环境变量兜底渠道
- 保持现有多渠道轮询逻辑不变。

建议实现：

- 新增轻量结构：
  - `type ChannelUse struct { ID *int64; Name string }`
  - `type ImageGenerationResult struct { Images []GeneratedImage; Channel ChannelUse }`
- 或在现有返回值旁增加 `ChannelUse` 返回值。

局部验收：

- 有后台渠道时，成功调用返回对应渠道 ID 和名称。
- 使用环境变量兜底时，返回 `channel_id = nil` 和兜底名称。
- 多渠道失败后成功时，返回最终成功渠道。

局部自测：

- `go test ./service -run "TestGenerateImage|TestChannel" -v`

文档更新：

- 在 `docs/progress.md` 记录渠道归因回传口径。

### 1.3 生成任务写入渠道归因

开发范围：

- 在生成开始调用上游前不写渠道，避免误标。
- 上游调用成功或最终失败时，把实际渠道写入 `generations`。
- 成功任务、失败任务都要保留归因。

局部验收：

- 生成成功记录有 `channel_id/channel_name`。
- 生成失败记录有最后尝试渠道。
- 保存图片失败时仍保留生成阶段使用的渠道。
- 取消任务不强制写渠道，除非已经实际调用上游。

局部自测：

- `go test ./service -run "Test.*Generation|Test.*Channel" -v`
- `go test ./controller -run "TestGeneration" -v`

文档更新：

- 在 `docs/progress.md` 记录成功、失败、取消三种状态的归因行为。

### 1.4 渠道近 24 小时统计接口

开发范围：

- 后台渠道列表接口 `/api/admin/channels` 返回每个渠道的近 24 小时统计：
  - `recent_success_count`
  - `recent_failed_count`
  - `recent_failure_rate`
- 可选增加环境变量兜底统计项；若不进入渠道列表，则先只在文档说明。

建议后端实现：

- 查询渠道列表后，一次性从 `generations` 按 `channel_id` 聚合。
- 使用服务器当前时间 `now.Add(-24 * time.Hour)`。
- 只统计 `status in (3,4)`。

局部验收：

- 有成功/失败记录时返回正确计数。
- 无记录时返回 0。
- 禁用渠道也能看到历史近 24 小时统计。
- 老数据 `channel_id is null and channel_name = ""` 不归入某个渠道。

局部自测：

- `go test ./controller -run "TestAdminChannel" -v`

文档更新：

- 在 `docs/progress.md` 记录统计口径。

### 1.5 后台渠道页展示统计

开发范围：

- 渠道列表展示：
  - 近 24 小时成功数
  - 近 24 小时失败数
  - 失败率
- 无数据时显示“暂无生成统计”或 0。
- 保留已有最近测试时间/结果展示。

局部验收：

- 渠道测试结果和生成统计不混淆。
- 统计信息在窄屏不挤压操作按钮。
- 失败率格式清晰，例如 `12.5%`。

局部自测：

- `pnpm.cmd exec vue-tsc --noEmit`
- 浏览器手动查看渠道列表有数据/无数据状态。

文档更新：

- 在 `docs/progress.md` 记录前端展示结果。

## 大功能整体验收

整体验收范围：

- 生成成功能记录渠道归因。
- 生成失败能记录最后尝试渠道。
- 后台渠道页能展示近 24 小时成功、失败、失败率。
- 老数据和环境变量兜底不会导致接口报错。
- 原有生成流程、多渠道重试、渠道测试功能不回归。

整体自测命令：

- `go test ./service -run "Test.*Generation|Test.*Channel" -v`
- `go test ./controller -run "TestAdminChannel|TestGeneration" -v`
- `go test ./...`
- `pnpm.cmd exec vue-tsc --noEmit`
- `pnpm.cmd build`

提交规则：

- 小功能完成只更新文档，不单独提交。
- 大功能整体验收通过后提交并推送。
- 建议提交信息：`feat: attribute generations to channels`

## 风险与回滚

### 风险

- 生成服务当前返回值改造可能影响调用链，需要保持接口变更范围小。
- 多渠道重试时，失败归因采用“最后尝试渠道”，不等同于完整失败链路。
- 历史数据无法回填，统计上线初期会只反映新任务。

### 回滚策略

- 新增字段为可空字段，回滚代码后不会影响旧流程读取。
- 如果统计接口出错，可先隐藏前端统计展示，保留字段写入。
- 不删除新增列，避免数据库回滚风险。

## 暂不纳入本次

- `generation_channel_attempts` 渠道尝试明细表。
- 每次重试的耗时、HTTP 状态、错误详情链路。
- 按模型、尺寸、用户分组做渠道质量分析。
- 历史记录渠道归因回填。

