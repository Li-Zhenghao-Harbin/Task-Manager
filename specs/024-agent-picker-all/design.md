# 024 设计：把下拉的数据源从「可用的」放开到「全部」

## 定位

下拉框只有一个数据源 —— `App.vue` 里的 `candidateAgents()`：

```ts
// 改前
function candidateAgents(task: { agentId: string } | null) {
  if (!task) return agents.value.filter((agent) => agent.available)
  const list = agents.value.filter((agent) => agent.available)
  if (task.agentId && !list.some((agent) => agent.id === task.agentId)) {
    const current = agentById(task.agentId)
    if (current) return [...list, current]   // ← 已关联的那一个作为特例补回来
  }
  return list
}
```

两处调用：

| 调用点 | 原代码 | 症结 |
| --- | --- | --- |
| 新建任务弹窗（`draft`） | `candidateAgents(draft)` | `draft.agentId` 恒为 `''`，永远走 `return list` 那一支 → 只看到可用项 |
| 任务详情面板（`selected`） | `candidateAgents(selected)` | 已关联的不可用项靠「特例」补回来；但**换**成一个不可用的就不行了 |

所以用户遇到的现象只出现在新建任务上 —— 那条特例让旧任务看起来「没问题」。

## 改动

```ts
// 改后
function candidateAgents() {
  return agents.value
}
```

两处调用改成 `candidateAgents()`。

保留这个函数（而不是在模板里直接写 `v-for="agent in agents"`）的原因：两处下拉共用同一个口径，
以后若要加「可用的排前面」之类的规则，只改一处。

注释里写清了「为什么当初只列可用的，以及为什么现在不需要那条特例」——
全量列表天然涵盖「已关联的即使不可用也要保留」这件事，特例失去了存在意义。

## 为什么不需要动服务端

`validateLocked()` 对 AgentID 的唯一约束是存在性：

```go
if input.TaskType == "code" && input.AgentID != "" && !a.agentExistsLocked(input.AgentID) {
    return fmt.Errorf("关联的 Agent 不存在")
}
```

没有 `Available` 判断。这条**本来**就是对的 —— 关联与可用是两件事 —— 只是以前被前端挡在门外，
没人发现服务端其实一直是放行的。为了不再依赖「碰巧没人加校验」，补了
`TestUnavailableAgentStillAttachesToTask` 把它固定住：创建、改派、派生视图（`taskIds`）三条路径都走一遍。

## 验证方法

光数 option 个数证明不了「选了能存下来」，所以探针走完整条路：

```
点「＋ 新建任务」→ 默认任务类型就是「代码」→ 读「关联 Agent」下拉
  → 选中那个不可用的 → 填任务名 / 勾「本人」→ 点「创建任务」
  → 读 window.__calls.CreateTask[0].agentId
  → 等弹窗关闭、列表刷新 → 读右侧详情面板「关联 Agent」的当前值
```

对照组没法用「注入 CSS」复现（这次的过滤在 JS 里），改成**把旧实现的输出还原到 DOM 上**：
`?old=1` 时把不可用项从两个下拉里摘掉，再跑同一套断言。这是「现象级对照」而不是「代码级对照」，
spec 里已写明它的局限。

## 两个探针细节

1. **用 label 自己的文本定位控件**。`label.textContent` 会把下属 `option` 的文字一起算进来，
   任何含「关联 Agent」字样的选项都能被误命中；只能累加直接子文本节点。
2. **原生 `<select>` 的弹出层抓不到**。`--screenshot` 拍不到展开的下拉，
   所以 `?expand=1` 时把 select 临时改成列表框（`size=6`）再截图 —— 那只是给肉眼看的佐证，
   读数一律以断言为准，文件名里也标了 `expanded`。
