# 004 设计说明

本次改动全部集中在前端 `frontend/src/App.vue` 与 `frontend/src/App.css`，后端 `app.go`、`notify.go` 无改动。

## 1. 移除保存状态提示

### 变更

| 位置 | 变更 |
| --- | --- |
| `flushSaves()` | 去掉 `saveState` / `saveHint` 的状态流转；校验未过时改为 `message.value = issue` |
| 脚本 | 删除 `saveState`、`saveHint` 两个 ref 与 `saveStateText` computed |
| 任务详情面板标题 | 去掉 `<span class="save-state">`，`.heading-status` 只留状态徽标 |
| Agent 详情面板标题 | 同上，只留 `可用` / `不可用` 胶囊 |
| `App.css` | 删除 `.save-state` 系列规则（含 `.saving` / `.saved` / `.invalid`） |

### 为什么把校验错误改走 toast

原先 `saveState='invalid'` 会让胶囊变红并把错误文案显示在里面，属于常驻提示。
胶囊整体被删除后，错误需要一个出口：应用里所有其它错误都已经走底部 toast（`message` + 3s 自动消失），
沿用同一个出口比新造一个面板内提示更一致。

副作用（已知并接受）：校验失败后提示 3 秒后消失，而字段仍为空。此时该次改动不会落盘，
直到用户把字段改回合法值才整体保存。

## 2. 列表分组与 context 上级任务

### 数据

新增两个 computed，替换原先写在模板里的内联过滤：

```ts
type TaskGroup = { parent: Task; context: boolean; children: Task[] }

taskGroups     // 上级任务分组：命中筛选，或「未命中但有子任务命中」
orphanChildren // 子任务但 parentId 在任务表中不存在
```

`taskGroups` 按 `tasks.value` 的原始顺序遍历，保证分组顺序与「全部任务」下的自然顺序一致。

判定逻辑：

```ts
for (const task of tasks.value) {
  if (task.kind === 'child') continue
  const children = visible.filter(c => c.kind === 'child' && c.parentId === task.id)
  if (!visibleIds.has(task.id) && !children.length) continue
  groups.push({ parent: task, context: !visibleIds.has(task.id), children })
}
```

- 上级命中筛选 → 正常分组头（行为与改前一致）。
- 上级未命中、但有子任务命中 → `context: true`，作为分组头带出来。
- 上级未命中且没有子任务命中 → 不出现（与改前一致）。

`orphanChildren` 的判据从「上级不在 filteredTasks 里」改为「上级在 tasks 里根本找不到」。
这是必要的语义修正：改前那种判据会把「上级被筛掉」的子任务也当成孤儿，
现在这类子任务由 `taskGroups` 负责，孤儿只留给真正悬空的数据。

### 视图

- 分组头：`.task-row` + `context` 类；子任务：`.task-row.child`（缩进 38px，与改前一致）。
- 孤儿子任务沿用 `.orphan-child`，文案由「子任务」改为「子任务（上级已删除）」。
- 空状态判据改为 `!taskGroups.length && !orphanChildren.length`。

### 样式

```css
.task-row.context{position:relative;color:var(--muted)}
.task-row.context::before{content:'';position:absolute;left:0;top:0;bottom:0;width:3px;background:#dbe6f5}
.task-row.context .task-name{font-weight:600}
.task-row.context .date{color:#93a1b6}
```

- 用 `::before` 画左侧色条而不是 `border-left`，避免 3px 边框把行内容整体右推、与其它行错位。
- **刻意不覆盖背景色**：`已完成` 的上级会保留 `.task-row.completed` 的绿色底，绿色本身就是「它已完成」的信息。
- context 行仍可点击、仍可右键建子任务（`showContextMenu` 对非子任务一律放行）。

## 3. 「5小时后可用」按钮

### 行为

```ts
function setAgentCooldown(target: { available:boolean; resetAt:string }, hours: number) {
  const at = new Date(Date.now() + hours * 3600 * 1000)
  const pad = (v:number) => String(v).padStart(2, '0')
  target.available = false
  target.resetAt = `${at.getFullYear()}-${pad(at.getMonth()+1)}-${pad(at.getDate())}T${pad(at.getHours())}:${pad(at.getMinutes())}`
}
```

- 时间格式必须与后端 `agentResetLayout = "2006-01-02T15:04"` 对齐：本地时间、分钟精度、`T` 分隔。
  用 `toISOString()` 会得到 UTC + 秒精度，与后端 `ParseInLocation(..., time.Local)` 语义不符。
- 只改 `available` 与 `resetAt` 两个字段，因此复用已有的 `watch(selectedAgent, …, {deep:true})`
  → `queueAgent` → 防抖 → `UpdateAgentApp` 链路，不需要新接口，也不需要手动 `refresh()`。

### 视图

按钮**放在 `<label>` 之外**。`可用情况` 是一个包裹着 checkbox 的 `<label>`，
把 button 塞进 label 内部属于无效的交互式内容嵌套，点击按钮可能顺带触发 label 的标签控件（checkbox）激活，
造成「点一下冷却、勾选状态又翻回来」。因此单独起一个 `.cooldown-row` 兄弟节点：

```html
<label>可用情况<span class="inline-toggle"><input type="checkbox" …><b>当前可用</b></span></label>
<div class="cooldown-row">
  <button type="button" class="cooldown-button" @click="setAgentCooldown(selectedAgent!, 5)">5小时后可用</button>
  <small class="hint">点击后立即置为不可用，5 小时后自动恢复并发送通知</small>
</div>
<label v-if="!selectedAgent.available">下次重置时间<input v-model="selectedAgent.resetAt" type="datetime-local"></label>
```

### 样式

```css
.cooldown-row{display:flex;align-items:center;gap:10px;margin:10px 0 4px}
.cooldown-button{border:1px solid #dbe5f2;background:#f8fbff;color:var(--blue);border-radius:8px;padding:8px 13px;font-size:12px;font-weight:700;cursor:pointer;white-space:nowrap}
.cooldown-button:hover{background:var(--blue-50);border-color:#bfdbfe}
.cooldown-row .hint{margin:0}
```

配色沿用 `.status-buttons button` 的「浅蓝底 + 品牌蓝字」次级按钮样式，不用 `primary`，
因为它是面板内的高频操作而非页面主操作。

## 验证策略

后端零改动，因此本次只用前端无 GUI 验证：

1. `npm run build`（`vue-tsc --noEmit` + `vite build`）确认类型与产物。
2. mock 预览页 + Chrome 无头截图，四个视图各带一条行为断言横幅：
   - `?view=list`：筛「进行中」后，`旧版系统迁移`（已完成）必须是 `task-row completed context`，
     且紧邻 `历史数据回填` 上方；`遗留脚本整改` 必须带 `orphan-child`；页面上 `save-state` 元素数为 0。
   - `?view=task`：任务面板无 `save-state`、无「保存」按钮。
   - `?view=agent`：快捷按钮文案为 `5小时后可用`、无 `save-state`、无「保存」按钮。
   - `?view=cooldown`：点按钮后 `UpdateAgentApp` 收到 `available=false`、`resetAt` 分钟精度、
     且距点击时刻恰为 300 分钟；面板徽标变为「不可用」；`下次重置时间` 字段出现并与 payload 一致。
3. `go test ./...` / `wails build` 确认后端未受影响、可出包。
