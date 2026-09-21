# 003 技术设计

## 一、关系统一

### 不变量

```text
agent.ProjectIDs == [t.ID for t in tasks where t.TaskType == "code" and t.AgentID == agent.ID]
```

按 `tasks` 的存储顺序排列，天然去重（每个任务只有一个 `agentId`）。`ProjectIDs` 仍然落盘，但**任何写入路径都不再直接信任它**，落盘前统一由任务重建。

### 三个内部函数

| 函数 | 职责 |
| --- | --- |
| `syncAgentProjectsLocked()` | 按上述不变量重建所有 Agent 的 `ProjectIDs`；有变化时刷新 `UpdatedAt` |
| `repairAgentLinksLocked()` | 一次性数据修复：遍历 Agent（按存储顺序），把项目列表里 `agentId` 为空的代码任务补记到该 Agent 名下；随后调用 `syncAgentProjectsLocked()` |
| `assignAgentProjectsLocked(agentID, ids)` | 把 `ids` 视为该 Agent 的**目标项目集合**：集合内的任务 `agentId` 置为本 Agent（从其他 Agent 转移），集合外且原本属于本 Agent 的任务清空 `agentId` |

`repairAgentLinksLocked()` 的补记规则保证了 003 验收标准 8：`agentId` 非空的任务以任务侧为准（不覆盖），为空的任务才接受 Agent 侧的补记；两个 Agent 都列了同一任务时，先出现的 Agent 胜出。修复后再 `sync`，因此标准 9 也成立。

### 写入路径

| 入口 | 处理 |
| --- | --- |
| `load()` | 反序列化后执行 `repairAgentLinksLocked()`；有修复则立即落盘 |
| `ImportData()` | 同上（导入外部文件同样需要修复） |
| `CreateTask` / `UpdateTask` | 非代码任务强制清空 `agentId`（既有），随后 `syncAgentProjectsLocked()` |
| `CreateAgentApp` | 先创建，再 `assignAgentProjectsLocked(newID, input.ProjectIDs)` |
| `UpdateAgentApp` | 先 `assignAgentProjectsLocked(id, input.ProjectIDs)`，再 `syncAgentProjectsLocked()` |
| `DeleteAgentApp` | 清空引用该 Agent 的任务 `agentId`（既有），再 `syncAgentProjectsLocked()` 把其他 Agent 的列表补齐 |
| `DeleteTask` | 既有逻辑直接删除即可；改为调用 `syncAgentProjectsLocked()` 统一处理 |

`UpdateAgentApp` 的返回值必须是 `sync` 之后的最终 `ProjectIDs`，前端据此回写本地状态。

## 二、重置通知

### 检查与通知两条路径分开

```text
applyAgentResetLocked() []string     // 只做状态恢复，返回被恢复的「应用名-模型名」
checkAgentResets()                   // 加锁 → 恢复 → 落盘 → 解锁 → 逐个通知 + 事件广播
```

把「通知」放在锁外，避免在持有写锁时启动子进程。

### 常驻检查

```text
startup():
  a.ctx = ctx
  load() → repair → save
  checkAgentResets()                 // 覆盖「关闭期间已到点」
  go a.watchAgentResets()            // ticker 30s + select on a.ctx.Done()
```

`watchAgentResets()` 只在 `a.ctx` 就绪后启动；`a.ctx.Done()` 触发即退出，不做额外清理。

### Windows 通知实现

`notifier func(title, body string) error` 作为 `App` 的字段（默认 `nil` 时回退到 `showWindowsToast`），测试中直接替换为空实现，避免测试真的弹通知。

`showWindowsToast` 走 PowerShell 的 WinRT 接口，脚本用 `-EncodedCommand`（UTF-16LE + Base64）传递，**不经临时文件、不经过 shell 引号解析**；标题与正文通过**环境变量**传入，脚本正文保持纯 ASCII，从而绕开 PowerShell 5.1 对无 BOM 脚本按 ANSI 解码导致中文乱码的问题。

```text
powershell.exe -NoProfile -NonInteractive -WindowStyle Hidden -EncodedCommand <base64>
  环境变量：TM_TOAST_APPID / TM_TOAST_XML
```

XML 在 Go 侧构造（`encoding/xml` 转义标题与正文），脚本只做 `LoadXml` + `Show`。AppID 使用 PowerShell 自身的 AUMID，无需注册开始菜单项即可显示。

通知失败只记录，不向 UI 报错：状态恢复与落盘已经完成，通知是尽力而为。

### 前端联动

恢复后调用 `runtime.EventsEmit(a.ctx, "agents:reset", names)`；前端 `EventsOn('agents:reset')` → `refresh()`。事件在无 GUI 的 mock 预览里不存在，因此前端注册时要允许 `window.runtime` 缺失（try/catch）。

## 三、前端即时保存

### 状态

```text
saveState: 'idle' | 'saving' | 'saved' | 'invalid'
saveHint: string                     // invalid 时的原因
dirtyTasks: Map<id, Task>            // 待落盘队列
dirtyAgents: Map<id, AgentApp>
savedSignatures: Map<id, string>     // 已落盘内容的签名（基线）
```

### 防回环：签名对比

```text
signature(payload) = JSON.stringify(payload)
taskPayload(task)   = id/name/kind/parentId/issuer/unit/publishDate/dueDate/taskType/status/completedAt/agentId/output
agentPayload(agent) = name/model/projectIds/available/resetAt
```

**刻意排除 `createdAt` / `updatedAt`**：服务端每次保存都会刷新 `updatedAt`，若纳入签名会导致「保存 → 回写 → 签名变化 → 再保存」的无限循环。

流程：深度 watch 选中的对象 → 计算签名 → 与基线相同则直接返回（这条同时覆盖了「切换选中项」和「服务端回写」两种情况）→ 不同则入队并重置 500ms 防抖定时器 → `flushSaves()` 落盘 → 用返回值重建基线。

### 落盘前的前端预校验

`taskIssue()` / `agentIssue()` 复刻服务端必填与日期规则。不通过时**不出队、不发请求**，置 `saveState = 'invalid'` 与 `saveHint`，用户补全后下一次 watch 自然重试。这样避免「清空输入框准备重打」时反复弹错误提示。

### 关联的双向同步（前端侧）

- 落盘任务成功后：把该任务的 `agentId` 镜像到本地 agents 的 `projectIds`（本地镜像服务端的 `sync` 逻辑，不额外发请求）。
- 落盘 Agent 成功后：按返回的 `projectIds` 镜像本地任务的 `agentId`。
- 两边都只重算**真正被改动**的对象并重建其基线，避免把用户此刻正在编辑的另一个对象误判为「已保存」。

### 视图变更

| 位置 | 变更 |
| --- | --- |
| 任务详情面板 | 移除【保存全部修改】按钮；标题区新增保存状态 |
| Agent 详情面板 | 同上 |
| 两个面板 | 删除按钮保留在 `.detail-actions` 中 |
| `App.css` | 新增 `.save-state` / `.save-state.saving` / `.save-state.invalid` |

### 关联项目勾选框

改用 Vue 复选框组的数组 `v-model`（`v-model="selectedAgent.projectIds"` + `:value="task.id"`），
替代原先手写的 `:checked` + `@change` 组合，少一处自定义逻辑，增删项目由运行时指令统一处理。

`v-for` 的 `key` 带上 Agent 身份（`selectedAgent.id + ':' + task.id`）：同一批任务在切换 Agent 时
会命中相同的 key，元素被原地复用容易留下上一个 Agent 的勾选残留；带上身份后每次切换都重建，状态必然干净。

## 四、构建与验证

```powershell
gofmt -l .
go vet ./...
go test ./...            # 含新增的关联同步、数据修复、通知回调用例
wails generate module    # GetXxx 签名未变，但保持绑定与源码一致
npm --prefix frontend run build
wails build
```

界面验证沿用 `wails-vue-ui-verify` 技能：注入 mock 绑定（本次需额外兜住 `window.runtime`）后用系统自带 Chrome 无头截图。
