# 003 实施任务

- [x] 编写规格与设计（spec.md / design.md）。
- [x] 后端：`syncAgentProjectsLocked()` / `repairAgentLinksLocked()` / `assignAgentProjectsLocked()`。
- [x] 后端：`load()` 与 `ImportData()` 接入数据修复并回写。
- [x] 后端：`CreateTask` / `UpdateTask` / `CreateAgentApp` / `UpdateAgentApp` / `DeleteTask` / `DeleteAgentApp` 保持不变量。
- [x] 后端：`applyAgentResetLocked()` 改为返回被恢复的 Agent 标签。
- [x] 后端：`checkAgentResets()` + `watchAgentResets()` 常驻检查（30s，随 ctx 退出）。
- [x] 后端：`showWindowsToast()`（PowerShell EncodedCommand + 环境变量传参）与可注入 `notifier`。
- [x] 后端：`EventsEmit("agents:reset")` 广播。
- [x] 前端：防抖自动落盘（签名基线 + 预校验）接入任务面板与 Agent 面板。
- [x] 前端：移除两个面板的保存按钮，新增保存状态展示。
- [x] 前端：监听 `agents:reset` 事件并刷新。
- [x] 前端：落盘后双向镜像关联状态。
- [x] 测试：关联双向同步、旧存档修复、同任务被两 Agent 引用、重置通知回调（新增 8 项，合计 19 项）。
- [x] 验证：gofmt / go vet / go test / vue-tsc 构建 / wails build / 无头截图。

## 验证记录

- 用真实存档副本跑数据修复：`Codex-GPT` 的 4 个关联项目全部补记到任务侧；`TRAE-Deepseek`
  同时保留了任务侧原有的 `ClouESP 8.0` 与 Agent 侧的 `Clou ReleaseFlow`，两边不再打架。
- 无头截图（见 screenshots/）：
  - `01-task-live-save.png`：任务面板无保存按钮，自动保存已调用 1 次。
  - `02-agent-live-save.png`：Agent 面板无保存按钮，字段与勾选状态正确。
  - `03-project-toggle-verified.png`：勾选「柒幻 芒种」后落盘 payload 为 `["t5","t4"]`，DOM 勾选状态一致。
  - `04-task-to-agent-sync.png`：任务侧改 Agent 后，Agent 侧项目列表同步出现该任务。
- `notify_manual_test.go`：默认跳过，`TM_TOAST_TEST=1 go test -run TestManualToast` 可发一条真实通知
  验证当前机器的通知通道（本机已验证 PowerShell 返回 0）。

## 实施前的准备

- 已备份用户存档：`build/bin/task-manager-data.backup-20260921-171012.json`。
- 数据修复只在任务 `agentId` 为空时补记，不覆盖任务侧已有的关联，可重复执行。

## 探针自造的假 bug（记录以免下次重踩）

无头探针最初「点击目标 Agent 行」后等的是 `.agent-projects-editor` 存在，而该元素在**任意** Agent 被选中时
就已存在，于是点击复选框发生在面板尚未切到目标 Agent 之前，勾中的是上一个 Agent 的复选框，数据被写坏。
这不是应用缺陷：修正等待条件为「标题里出现目标 Agent 名」后，落盘 payload 与 DOM 勾选状态完全一致。
教训已写入技能 `wails-vue-ui-verify`（第 4、6 节）。
