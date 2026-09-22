# 005 实施清单

## 后端（`app.go`）

- [x] `AgentPrompt` 结构体与 `AgentApp.Prompts map[string][]AgentPrompt`
- [x] 常量：`maxPromptsPerPair`(50) / `maxPromptLength`(20000)
- [x] `normalizePrompts([]AgentPrompt) []AgentPrompt`：补 ID（含去重）、丢空白、按 rune 截断
- [x] `normalizeAgentPromptsLocked()`：nil → 空 map，剪掉空键与指向已删除任务的键
- [x] `validatePromptTargetLocked(agentID, taskID) error`：Agent 存在 + 任务已关联 + 类型为 code
- [x] `ListAgentPrompts(agentID, taskID) ([]AgentPrompt, error)`
- [x] `UpdateAgentPrompts(agentID, taskID string, prompts []AgentPrompt) ([]AgentPrompt, error)`
- [x] `DeleteTask` 清理所有 Agent 下该任务的 prompt
- [x] `load` / `ImportData` 接入 `normalizeAgentPromptsLocked`
- [x] 确认 `UpdateAgentApp` / `syncAgentProjectsLocked` / `repairAgentLinksLocked` 不动 prompt
- [x] `CreateAgentApp` 与 `ListAgentApps` 把 `Prompts` 初始化为空 map（序列化为 `{}` 而非 `null`）

## 前端（`App.vue` / `App.css`）

- [x] 类型：`AgentPrompt`；`AgentApp` 增加 `prompts`
- [x] `normalizePrompts` / `normalizeAgent` 归一化 `prompts`
- [x] 状态：`expandedAgentIds` / `selectedAgentTaskId` / `prompts` / `savedPromptSignature`
- [x] `toggleAgentRow` / `openPromptPanel` / `selectedAgentTask` computed / `agentTasks` / `promptCount`
- [x] Agent 列表行：展开指示符 + `expanded` class + 展开区（关联任务行，带 prompt 条数）
- [x] 右侧 Prompt 管理面板：标题（任务名 + 所属应用）+ 返回入口 + 列表 + 添加 + 删除 + 空态
- [x] prompt 即时保存：`promptPayload` / `loadPrompts` / `queuePrompts` / `flushSaves` 第三段
- [x] 删除任务 / 删除 Agent 时清理 `selectedAgentTaskId` 与展开记录；删除确认文案说明 Prompt 影响
- [x] `App.css`：`.agent-caret` / `.agent-task-list` / `.agent-task-row` / `.agent-sub` / `.link-button` / `.prompt-*`
- [x] `.agent-row` 网格加一列（16px）放指示符，媒体查询同步

## 验证

- [x] 新增 10 个 Go 用例（`prompt_test.go`），总数 19 → 29
- [x] `gofmt -l .` / `go vet ./...` / `go test ./...` 全过
- [x] `wails generate module` 重建绑定
- [x] `npm --prefix frontend run build`（含 vue-tsc）
- [x] 无头浏览器验证 5 视图，行为断言横幅全绿：
  - `expand`：展开出 2 个任务行、计数 `2 条 / 0 条 Prompt`、指示符 `▾`
  - `prompt`：面板标题为任务名、所属 `Codex-GPT`、2 条内容正确、返回与添加按钮在位
  - `add`：**点「添加」后保存次数 = 0**（空条目不落盘）；输入后保存 1 次、payload 3 条且无空条目
  - `remove`：删第 1 条后 payload 剩 1 条、面板剩 1 条
  - `empty`：空态文案正确、只展开 TRAE（多行展开互不影响）
- [x] `wails build` 出包

## 收尾

- [x] 归档截图到 `specs/005-agent-task-prompts/screenshots/`（5 张）
- [x] 更新 `outline.md`（3.8 新章节、数据模型、接口、规则）
- [x] 更新工作区记忆
