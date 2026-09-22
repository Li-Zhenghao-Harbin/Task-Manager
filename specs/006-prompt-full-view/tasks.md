# 006 实施清单

## 规格

- [x] `specs/006-prompt-full-view/spec.md`：做什么 / 不做什么 / 验收标准
- [x] `specs/006-prompt-full-view/design.md`：布局、展开语义翻转、编辑队列、剪贴板降级、样式、验证方式

## 前端脚本（`App.vue`）

- [x] 展开状态取反：`expandedAgentIds` → `collapsedAgentIds`，`isAgentExpanded` 默认 true
- [x] 拆分动作：`selectAgent`（只选中）/ `toggleAgentExpand`（只切展开）
- [x] `promptViewOpen` computed（复用 `selectedAgentTask` 的关联校验，解除关联自动回落）
- [x] `openPromptView` / `closePromptView`（`selectedAgentTaskId = ''`）
- [x] 删除 `savedPromptSignature`，改为 `dirtyPrompts` 队列 + `loadedKey` + `promptKey`
- [x] `commitPromptBuffer` / `loadPrompts`（先提交旧缓冲再换新） / `promptPayloadOf` / `storedPrompts`
- [x] `flushSaves`：先提交当前缓冲区，再整队提交，成功后回写 `agent.prompts[taskId]` 并按需重建缓冲区
- [x] `forgetPending` 同步清理队列中指向该任务/该 Agent 的组合
- [x] `copyPrompt(prompt)`：运行时剪贴板 → `navigator.clipboard` → 临时 textarea，三级降级 + 提示
- [x] `deleteAgent` / `deleteTask` 清理改为 `collapsedAgentIds` 与 `closePromptView`

## 前端模板（`App.vue`）

- [x] `.workspace` 加 `:class="{ 'workspace-full': promptViewOpen }"`
- [x] `.workspace` 首个分支：Prompt 管理视图（返回按钮 + 标题 + 添加 + 卡片网格 + 空态）
- [x] Agent 列表行：指示符 `▾/▸` + `@click.stop` 切换展开；行 click 只选中
- [x] 任务行 click → `openPromptView`
- [x] 删除右侧的 Prompt 详情分支；详情区三个 `aside` 收进 `<template v-if="!promptViewOpen">`
- [x] 卡片头部加【复制】按钮（与【删除】同排）

## 样式（`App.css`）

- [x] `.workspace-full` / `.prompt-workspace` / `.prompt-title-line` / `.back-button`
- [x] `.prompt-board` / `.prompt-grid` / `.prompt-card` / `.prompt-card-head` / `.prompt-actions` / `.prompt-copy`
- [x] `.agent-caret` 变为可点（cursor / padding / hover）
- [x] 删除 005 右侧面板专用的 `.prompt-list` / `.prompt-item*` / `.prompt-editor` / `.prompt-add` / `.prompt-footnote`

## 验证

- [x] `gofmt -l .` / `go vet ./...` / `go test ./...`（29 项，无回归）
- [x] `npm --prefix frontend run build`（含 `vue-tsc`）
- [x] 无头截图 7 个视图（`agent` / `row` / `caret` / `prompt` / `back` / `copy` / `switch`），行为断言横幅全绿
- [x] `wails build` 产出 `build/bin/Task Manager.exe`

## 文档

- [x] `outline.md`：3.8 节改写（独立视图 + 默认展开 + 复制），标注 prompt 视图的区域占用
- [x] `.workbuddy/memory/` 当日日志 + `MEMORY.md` 约定（默认展开语义、编辑队列）
