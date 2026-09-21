# 002 Agent 模块技术设计

## 架构

沿用现有分层：Vue 单文件前端 → Wails 绑定 → Go `App` 服务 → JSON 仓储（`task-manager-data.json`）。

```text
App.vue (section='agent') → Wails binding → App service → DataStore.Agents → task-manager-data.json
App.vue (任务表单/详情)   → Wails binding → App service → Task.AgentID
```

## 数据结构

### AgentApp

```text
id          string / UUID
name        string       应用名
model       string       模型名（自由文本）
projectIds  []string     关联项目 = 任务类型为 code 的任务 ID
available   bool         可用情况
resetAt     string       下次重置时间，格式 "YYYY-MM-DDTHH:MM"，可空
createdAt   datetime     RFC3339
updatedAt   datetime     RFC3339
```

`resetAt` 采用 `datetime-local` 控件的原生格式（无时区、精确到分钟），解析用 `time.ParseInLocation("2006-01-02T15:04", ..., time.Local)`。

### Task 变更

新增字段 `agentId string`（可空）。**仅当 `taskType == "code"` 时可非空**，由服务端强制。

### DataStore 变更

```text
{
  units, defaultUnitId, tasks, settings,
  agents: AgentApp[]        // 新增
}
```

向后兼容：旧数据文件没有 `agents` 字段时读出为空数组；`load()` 的兼容分支（纯数组格式）保持不变。

## 接口

| 方法 | 签名 | 说明 |
| --- | --- | --- |
| ListAgentApps | `() []AgentApp` | 读列表；读取时先执行到点自动重置 |
| CreateAgentApp | `(AgentAppInput) (AgentApp, error)` | 校验后创建 |
| UpdateAgentApp | `(id string, AgentAppInput) (AgentApp, error)` | 校验后更新 |
| DeleteAgentApp | `(id string) error` | 删除，并清空任务上的引用 |

`AgentAppInput` 字段：`name / model / projectIds / available / resetAt`。

已有接口变更：`CreateTask`、`UpdateTask` 接收并保存 `agentId`。

### 校验规则（服务端）

1. `name`、`model` 去空格后不得为空 → 否则 `请填写应用名和模型名`。
2. `resetAt` 非空时必须能被 `2006-01-02T15:04` 解析 → 否则 `下次重置时间格式无效`。
3. `projectIds` 每个 ID 必须指向一个已存在且 `taskType == "code"` 的任务 → 否则 `关联项目必须是代码任务`。
4. `projectIds` 去重、剔除空串。
5. 任务侧：`agentId` 非空时必须指向已存在的 Agent 应用 → 否则 `关联的 Agent 不存在`。

### 自动重置

`applyAgentResetLocked()`：遍历 agents，对 `available == false && resetAt != ""` 且 `now >= resetAt` 的项，置 `available = true`、`resetAt = ""`、刷新 `updatedAt`，返回是否发生变化。

- `ListAgentApps()` 持写锁执行该函数，有变化则落盘再返回。
- `startup()` 加载数据后执行一次，避免「重置时间已过但一直没打开列表」时状态陈旧。

### 引用完整性

- `DeleteAgentApp` → 把所有 `task.agentId == id` 的任务置为 `""`。
- `DeleteTask` → 把所有 agent 的 `projectIds` 中该任务 ID 移除。
- `ExportData` / `ImportData` 的 `DataStore` 字面量补上 `Agents`。
- `CreateTask` / `UpdateTask` 在校验通过后，若 `taskType != "code"` 则强制 `agentId = ""`。

## 前端设计

### 状态

- `section` 联合类型扩展为 `'tasks' | 'output' | 'agent' | 'units'`。
- 新增 `agents = ref<AgentApp[]>([])`、`selectedAgent = ref<AgentApp | null>(null)`。
- 新增 `agentDraft` 仅用于「新增 Agent 应用」模态框；右侧详情面板直接编辑 `selectedAgent`。
- `refresh()` 中并行拉取 `ListAgentApps()`，并保持 `selectedAgent` 的选中项不丢。

### 候选 Agent 计算

```text
candidateAgents(task) = agents.filter(a => a.available)
                        ∪ (当前 task.agentId 对应的 agent，即使不可用)
```

显示文案统一为 `agentLabel(a) = `${a.name}-${a.model}``。

### 视图

| 位置 | 变更 |
| --- | --- |
| 左侧菜单 | `Agent` 分组标题（`side-title menu-divider`）+ `Agent应用` 子条目（新增 `.nav-sub` 缩进） |
| workspace | 新增 `section === 'agent'` 分支：标题区 + 「新增 Agent 应用」按钮 + 列表 + 空状态 |
| 右侧详情 | 新增 `section === 'agent' && selectedAgent` 分支，字段可编辑并带「保存 / 删除」 |
| 任务详情 | `taskType === 'code'` 时显示「关联 Agent」下拉 |
| 创建任务表单 | 同上 |
| 任务列表行 | 代码任务且有关联时，在 `small` 行追加「应用名-模型名」 |

### 样式

在 `App.css` 追加：`.nav-sub`、`.agents-workspace`、`.agents-panel`、`.agent-row`、`.agent-title`、`.agent-meta`、`.agent-projects`、`.agent-pill`、`.agent-pill.offline`、`.agent-reset`。配色沿用现有 CSS 变量（`--blue`、`--line`、`--muted`），不可用徽标沿用 `--danger` 系红色。

## 构建与绑定

Go 方法变化后需重建 `frontend/wailsjs` 绑定：

```powershell
wails generate module
npm --prefix frontend run build
go vet ./...
wails build
```
