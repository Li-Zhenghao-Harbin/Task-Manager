# 项目长期记忆：TaskManager

## 技术栈事实（以代码为准，不要照抄 outline.md）

- Wails v2 + Go 1.26（后端单文件 `app.go`）+ **Vue 3 `<script setup>`**（`frontend/src/App.vue`）。
  `outline.md` / `setup.md` 里写的「React + TypeScript」是早期规划，**实际是 Vue**，不要被误导。
- 数据持久化：`DataStore` 结构体（`units` / `defaultUnitId` / `tasks` / `agents` / `settings`）序列化到可执行文件同目录的 `task-manager-data.json`，写盘走「临时文件 + rename」。无 SQLite。
- 前端路由是单一 `section` ref（`'tasks' | 'output' | 'agent' | 'units'`），**没有 vue-router**。
- `App.vue` 与 `App.css` 刻意保持紧凑：模板大量压成超长单行。改样式前先确认是否已有同类 class，追加到 `App.css` 末尾即可。
- 样式变量统一用 `:root` 里的 `--blue / --blue-50 / --line / --ink / --muted / --page-background`，配色为蓝白主题。

## 约定

- SDD：规格放 `specs/NNN-<slug>/`，每号三件套 `spec.md`（做什么/不做什么/验收标准）、`design.md`、`tasks.md`。
- 领域词汇：`TaskType = code | document | support`（界面文案分别是 代码 / 文档 / 实施）。
- Agent 模块基础规则见 `specs/002-agent-module/spec.md`；**关联模型已被 003 修订**，见下。
- **关联关系只有一个事实来源：`Task.AgentID`。** `AgentApp.ProjectIDs` 是派生视图，
  由 `syncAgentProjectsLocked()` 按任务重建。任何改动关系的代码都必须维持这个不变量，
  不要再往 `ProjectIDs` 里直接写数据。`repairAgentLinksLocked()` 在 `load()`/`ImportData()` 里做幂等修复。
- 右侧详情面板（任务 / Agent）是**即时保存**，没有保存按钮，也**不显示任何保存状态提示**（004 移除）。
  新增面板字段时不用加保存逻辑，但新字段要进 `taskPayload()` / `agentPayload()` 的签名，否则自动保存不会触发。
  **签名必须排除 `updatedAt`/`createdAt`**，否则会形成保存死循环。
- 任务列表渲染走 `taskGroups` / `orphanChildren` 两个 computed（004），**不要退回模板里内联过滤**。
  `taskGroups` 会把「子任务命中筛选但上级未命中」的上级任务作为 `context:true` 分组头带出来，
  这是「筛进行中要看到已完成上级」的机制；`.task-row.context` 样式刻意不覆盖 background，以保留 `.completed` 绿底。
- Agent 快捷冷却：`setAgentCooldown(target, hours)` 生成 `YYYY-MM-DDTHH:mm` **本地时间**（对齐后端 `agentResetLayout`，
  不能用 `toISOString()`）。面板里的按钮一律放在 `<label>` **之外**，避免 label 嵌套 button 导致误触发 checkbox。
- Agent 通知：`App.notifier` 是可注入字段，测试里替换它避免真弹通知；`notify.go` 里是 Windows 实现。
- **Prompt 挂在 (Agent, 任务) 组合上**（005）：`AgentApp.Prompts map[taskID][]AgentPrompt`，键是任务 ID。
  接口只有 `ListAgentPrompts` / `UpdateAgentPrompts`，后者是**整体替换**——加/删/改都是前端对本地列表的操作，
  不要新增 add/update/delete 这类细粒度接口。
  - 生命周期：解除关联**保留** prompt；删任务时清掉它在各 Agent 下的 prompt；`UpdateAgentApp` 不得动 prompt
    （靠「复制结构再改字段」天然成立，别改成整体赋值）。
  - 前端 `prompts` ref 是编辑副本，签名基于**过滤掉空条目**的 payload（`promptPayload`）：
    点「添加」不发请求，把已有内容清空会触发保存让后端删掉那条。
  - `selectedAgentTask` 这个 computed 里带 `agent.projectIds.includes(taskId)` 校验，
    「取消关联后面板回落」靠它兜住，不要另写清理逻辑。
- **Agent 列表默认展开**（006）：状态是 `collapsedAgentIds`（记「哪些行被收起」，默认空 = 全展开），**不持久化**。
  点行的其它位置只 `selectAgent`，只有点行首 `<span class="agent-caret">` 才 `toggleAgentExpand`（靠 `@click.stop`，
  因为 `.agent-row` 是 `<button>`，不能嵌按钮）。
- **Prompt 是独立整区视图**（006），不在右侧面板：`.workspace` 加 `workspace-full`（`grid-column:2 / -1`）横跨两列，
  详情区三个 `aside` 包在 `<template v-if="!promptViewOpen">` 里**整块不渲染**（只隐藏会留下栅格项把 workspace 挤到下一行）。
  打开条件 `promptViewOpen = section==='agent' && !!selectedAgent && !!selectedAgentTask`。
- **Prompt 待落盘改动按 `(agentId, taskId)` 排队**（006）：`dirtyPrompts: Map<`${agentId}::${taskId}`, {agentId,taskId,list}>`
  + `loadedPromptKey`。`commitPromptBuffer()` 在 `loadPrompts()` 与 `flushSaves()` 开头调用（先把旧缓冲收进队列再换新的），
  **并且自己 `scheduleSave()`**——输入与「返回/切任务」在同一次渲染里发生时，`queuePrompts` 会因为 `loadedPromptKey` 已被清空而直接 return。
  005 的「单缓冲区 + 单签名」写法会静默丢编辑，**不要退回去**。
- 复制 prompt 走三级降级：`ClipboardSetText`（`wailsjs/runtime/runtime`）→ `navigator.clipboard` → 临时 textarea，
  三级都要包 try/catch（预览环境没有 `window.runtime`）。

## 构建与验证

- 改了 Go 方法签名后必须重建绑定，否则前端类型报错：`wails generate module`。
- 日常验证顺序：`go test ./...` → `npm --prefix frontend run build`（含 vue-tsc 类型检查）→ `wails build`。
- 前端界面验证走用户级技能 `wails-vue-ui-verify`（mock 绑定 + 无头截图 + 行为断言），
  注意 `window.runtime` 也要给桩，探针要等目标视图真正渲染完再操作。
  预览服务必须用后台任务常驻（`(cmd &)` 会随那次工具调用一起被杀，截图全是 `ERR_CONNECTION_REFUSED`）。
