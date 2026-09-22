# 005 设计：Agent 关联任务的 Prompt 管理

## 1. 数据模型

### 1.1 Go

```go
// AgentPrompt is one prompt stored against a (Agent application, task) pair.
// Text is the whole payload: a prompt has no title or note of its own.
type AgentPrompt struct {
    ID   string `json:"id"`
    Text string `json:"text"`
}

type AgentApp struct {
    ID         string                    `json:"id"`
    Name       string                    `json:"name"`
    Model      string                    `json:"model"`
    ProjectIDs []string                  `json:"projectIds"`
    Available  bool                      `json:"available"`
    ResetAt    string                    `json:"resetAt"`
    // Prompts maps a task ID to the prompts written for it on this application.
    // A key with an empty slice is dropped on save.
    Prompts   map[string][]AgentPrompt `json:"prompts"`
    CreatedAt string                   `json:"createdAt"`
    UpdatedAt string                   `json:"updatedAt"`
}
```

`Prompts` 用 `map[taskID]` 而不是在 `Task` 上开字段，理由：

- prompt 的语义主体是「**用这个 Agent 跑这个任务**」，换 Agent 就是另一套 prompt；
- 反过来，一个代码任务只能关联一个 Agent（003 的模型），所以 map 的键天然唯一；
- 不污染 `Task` —— `Task` 是关联关系的唯一事实来源，prompt 是关系之上的用户内容，两者不该混在一个结构里。

`Prompts` **不参与** `syncAgentProjectsLocked` / `repairAgentLinksLocked` 的关系推导，
它只跟着「任务被删除」这一个事件清理。

### 1.2 序列化约定

- 空 map 序列化为 `{}`，不是 `null`：`load` 归一化时若为 `nil` 就置为空 map，
  和 `ProjectIDs` 用空数组而非 null 是同一个考虑（前端不必到处判空）。
- 旧存档没有 `prompts` 字段 → `json.Unmarshal` 得到 nil → 归一化为空 map，不报错、不丢其它数据。
- **`Prompts` 不进 `AgentAppInput`**：`UpdateAgentApp` 只碰应用名/模型名/关联项目/可用性/重置时间，
  复制语义（`updated := a.agents[idx]` 再改字段）天然保留了 prompt，不会被清空。

### 1.3 约束

| 约束 | 值 | 校验位置 |
| --- | --- | --- |
| 单个 (Agent, 任务) 的 prompt 条数上限 | 50 | `normalizePrompts` |
| 单条 prompt 文本长度上限 | 20000 字符（按 rune 计） | `normalizePrompts` |
| 任务必须已关联到该 Agent | 硬校验，违反报错 | `validatePromptTargetLocked` |
| 任务类型必须是 `code` | 硬校验（关联关系已隐含，仍显式检查） | `validatePromptTargetLocked` |

超限的处理是**截断**（条数与长度）而非报错，避免用户粘贴超长文本时直接丢内容、还不知道为什么。

## 2. 后端接口

```go
// ListAgentPrompts returns the prompts stored for one (application, task) pair.
func (a *App) ListAgentPrompts(agentID, taskID string) ([]AgentPrompt, error)

// UpdateAgentPrompts replaces the whole prompt list of one (application, task) pair.
// Entries with a blank ID get a generated one; entries whose text is blank are dropped.
func (a *App) UpdateAgentPrompts(agentID, taskID string, prompts []AgentPrompt) ([]AgentPrompt, error)
```

### 2.1 为什么是「整体替换」而不是 add / update / delete 三个接口

右侧面板已经是**即时保存**模型（003）：本地维护编辑中的列表，防抖后整体下发。
用整体替换，前端不需要为「添加」「删除」「改文本」各维护一条链路，
也不会出现「本地顺序与后端顺序不一致」的问题。加/删/改在界面上都是对本地列表的操作。

### 2.2 `normalizePrompts` 的规则

1. 逐个条目：`ID` 为空则生成 UUID；
2. `Text` 去掉首尾空白后为空 → **丢弃该条**；
3. `Text` 只做右侧空白清理（`TrimRight`），保留用户的换行与缩进，过长则按 rune 截断；
4. 合计最多 50 条，超出丢弃；
5. 返回的切片一定非 nil（空列表也要能序列化成 `[]`）。

### 2.3 清理路径

| 触发 | 实现 |
| --- | --- |
| `DeleteTask` | 遍历所有 Agent，`delete(agent.Prompts, taskID)` |
| `DeleteAgentApp` | 无需专门处理：Agent 整个从 `a.agents` 移除，prompt 随之消失 |
| `assignAgentProjectsLocked`（解除关联） | **不动** prompt，这是刻意的（见 spec 第 4 条） |
| `load` / `ImportData` | `normalizeAgentPromptsLocked` 统一归一化：nil → 空 map，清掉空键 |

## 3. 前端设计

### 3.1 新增状态

```ts
type AgentPrompt = { id: string; text: string }
type AgentApp = { /* ...既有字段 */ prompts: Record<string, AgentPrompt[]> }

const expandedAgentIds = ref<string[]>([])      // 展开的 Agent
const selectedAgentTaskId = ref('')             // 正在看 prompt 的任务
const prompts = ref<AgentPrompt[]>([])          // prompt 编辑副本
let savedPromptSignature = ''                   // 已落盘的 prompt 列表签名
```

`selectedAgentTaskId` 用空串而不是 `null`，与既有 `agentId: ''` 的写法保持一致。

### 3.2 交互流

```
点 Agent 行
  → selectedAgent = agent；selectedAgentTaskId = ''；toggle 展开状态
  → 右侧：Agent 详情（原行为）

点展开的任务行
  → selectedAgentTaskId = task.id；loadPrompts()
  → 右侧：Prompt 管理面板

点「返回」
  → selectedAgentTaskId = ''
  → 右侧：Agent 详情
```

`selectedAgentTask` 用 computed 求出：Agent 的项目列表里找不到该任务时返回 `null`，
面板自动回落到 Agent 详情。**这条同时兜住了「取消关联后面板还停在已解除任务上」的情况**，
不需要额外的清理代码。

### 3.3 prompt 的保存

沿用 003 的即时保存骨架，但签名基于**过滤后的列表**：

```ts
function promptPayload(list: AgentPrompt[]) {
  return list.filter(item => item.text.trim()).map(item => ({ id: item.id, text: item.text }))
}
```

- 本地列表 `prompts` 允许含空条目（用户刚点「添加」还没输入）。
- 签名比对用 `promptPayload` 的结果，因此：
  - 点「添加」→ payload 不变 → 不触发保存（正确：后端没有可写的东西）；
  - 输入文字 → payload 变化 → 触发保存；
  - **把已有内容清空 → payload 少一条 → 触发保存，后端整体替换后该条被删除**（这是必须的，
    否则界面上删干净了、数据里还留着）。
- 保存成功后 `savedPromptSignature` 更新，并把后端返回的列表写回 `selectedAgent.prompts[taskId]`，
  保证切换出去再切回来看到的是权威数据。

本地生成的条目 ID 用 `crypto.randomUUID()`，不可用时退回时间戳 + 随机串。

### 3.4 视图变更

| 位置 | 变更 |
| --- | --- |
| `.agent-row` | 网格加一列（18px）放展开指示符；`:class` 增加 `expanded` |
| 新增 `.agent-task-list` | 展开区容器，缩进对齐 Agent 行的标题列 |
| 新增 `.agent-task-row` | 关联任务行，含任务名与 prompt 条数 |
| 新增 `.prompt-editor` / `.prompt-item` / `.prompt-add` | Prompt 管理面板 |
| 右侧面板 | 在「任务详情」与「Agent 详情」之间插入 Prompt 管理分支 |
| `App.css` | 追加上述样式，不改既有规则 |

## 4. 测试计划

| 用例 | 覆盖 |
| --- | --- |
| `TestPromptsRoundTripOnAgentTaskPair` | 增删改 + 持久化往返 |
| `TestBlankPromptsAreDropped` | 纯空白文本不落盘 |
| `TestPromptIDsAreGenerated` | 空 ID 补 UUID |
| `TestPromptsRejectUnrelatedTask` | 未关联的任务、非代码任务被拒 |
| `TestDeletingTaskClearsPrompts` | 删除任务清理各 Agent 下的 prompt |
| `TestUnlinkingKeepsPrompts` | 解除关联保留 prompt |
| `TestUpdateAgentAppKeepsPrompts` | `UpdateAgentApp` 不动 prompt |
| `TestPromptLimitsAreEnforced` | 条数与长度上限 |
