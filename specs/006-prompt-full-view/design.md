# 006 设计：Prompt 独立视图、默认展开与复制

## 1. 改动范围

只动两个文件：`frontend/src/App.vue`（脚本 + 模板）、`frontend/src/App.css`。
**后端零改动** —— prompt 的存储、接口、生命周期在 005 已经定稿，本规格纯界面/交互层。

## 2. 布局：视图如何「占满主内容区 + 右侧面板」

`.shell` 是三列栅格（`230px | minmax(430px,1fr) | 350px`）：侧栏 / 主内容 / 详情。

```text
section !== 'agent' 或未打开 Prompt：   [侧栏][workspace][details]
打开 Prompt 视图：                      [侧栏][     prompt 视图     ]
```

做法：

- `.workspace` 是栅格项，Prompt 视图打开时加 `workspace-full` class → `grid-column: 2 / -1`，横跨第 2、3 列。
- 详情区的三个 `aside` 整体包进 `<template v-if="!promptViewOpen">`。
  **必须整块不渲染**：只把它隐藏（`display:none`）或让它渲染成空的 `.details` 占位，
  都会留下一个栅格项，栅格自动布局会把横跨两列的 workspace 挤到下一行。

用 `<template>` 包裹不引入额外 DOM 层（片段），不影响栅格项的位置。

打开条件是一个 computed，而不是新的布尔状态：

```ts
// 视图是否打开 = 当前在 Agent 区 + 选中了 Agent + 选中了它的一个关联任务
const promptViewOpen = computed(() => section.value === 'agent' && !!selectedAgent.value && !!selectedAgentTask.value)
```

复用 005 的 `selectedAgentTask`（内含 `agent.projectIds.includes(taskId)` 校验）：
任务被解除关联时它自动变 null，视图自动关闭回落，不需要额外清理逻辑。

## 3. 展开状态：从「已展开集合」翻转为「已收起集合」

005 的 `expandedAgentIds` 默认空 = 全部收起。要求改成默认展开，最省事的做法是**取反语义**：

```ts
const collapsedAgentIds = ref<string[]>([])            // 默认空 = 全部展开
function isAgentExpanded(agent) { return !collapsedAgentIds.value.includes(agent.id) }
function toggleAgentExpand(agent) { /* 在 collapsedAgentIds 里增删 */ }
function selectAgent(agent) { selectedAgent.value = agent }   // 只选中，不动展开状态
```

好处：新创建的 Agent 天然是展开的，`deleteAgent` 只需从 `collapsedAgentIds` 里摘掉 id。

指示符从 `expanded ? '▾' : '▸'` 变成 `isAgentExpanded(agent) ? '▾' : '▸'`，语义不变。

指示符变成一个真正的按钮（`<button class="agent-caret" @click.stop="toggleAgentExpand(agent)">`）：
`.agent-row` 本身是 `<button>`，嵌套按钮在 HTML 里非法，因此**行不能是 button**。
把行改成 `<div role="button" tabindex="0">`？改动面偏大。折中：指示符仍是 `<span>`，靠 `@click.stop` 拦截，
行按钮的其它区域走选中。`@click.stop` 在 span 上同样会阻止事件冒泡到行的 click 处理器 —— 满足需要，且不动 DOM 结构。

## 4. Prompt 编辑缓冲区：按 (Agent, 任务) 排队

005 的写法是「一个全局 `prompts` 缓冲区 + 一个全局签名」：

```ts
const prompts = ref<AgentPrompt[]>([])   // 当前编辑的组合
let savedPromptSignature = ''
watch([selectedAgent, selectedAgentTaskId], () => loadPrompts())   // 选中项一变就换缓冲区
watch(prompts, () => queuePrompts(), { deep: true })
```

问题：换缓冲区是**无条件覆盖**。在防抖（500ms）到点前返回或切任务，那次编辑直接消失 ——
而且因为签名基线也跟着换了，`flushSaves` 里的 `promptDirty` 判定为 false，连报错都没有。

改成与 `dirtyTasks` / `dirtyAgents` 同构的队列：

```ts
type PendingPrompt = { agentId: string; taskId: string; list: AgentPrompt[] }
const dirtyPrompts = new Map<string, PendingPrompt>()   // key = `${agentId}::${taskId}`
let loadedKey = ''                                      // 缓冲区当前对应的组合
```

- `commitPromptBuffer()`：把缓冲区里**过滤掉空条目**的内容与「模型里已知的该组合内容」比较，
  不一致就写进 `dirtyPrompts`，一致就把这个 key 从队列里删掉。
  （模型里的 `agent.prompts[taskId]` 就是服务端权威内容，不需要再单独维护一份签名基线。）
- `loadPrompts()` 先 `commitPromptBuffer()`（此时缓冲区还是**旧的**那一份，因为 Vue 的 watch 回调里 `prompts.value` 尚未被覆盖），
  再按新选中的组合重建缓冲区。顺序不能反。
- `flushSaves()` 里先 `commitPromptBuffer()`，把正在编辑的那份也收进队列，然后整队提交。
  提交成功后把服务端返回写回 `agent.prompts[taskId]`；如果这个组合正是当前缓冲区，
  再用「服务端结果 + 本地空白条目」重建缓冲区（保留用户刚点开还没输入的空行，并为已有条目补上后端生成的 ID）。

对「点添加不发请求」这条 005 行为的保持：空条目被 `promptPayloadOf` 过滤，
所以只点「添加」缓冲区内容仍与模型一致 → 不入队 → 不发请求。

## 5. 复制

剪贴板走三级降级，任一成功即止：

1. `ClipboardSetText`（`wailsjs/runtime/runtime`，Windows 原生剪贴板，最可靠）；
2. `navigator.clipboard.writeText`（WebView2 的 `http://wails.localhost` 属安全上下文，可用）；
3. 临时 `<textarea>` + `document.execCommand('copy')`（老兜底）。

全失败才提示「复制失败，请手动选中文本复制」。

三级都要包 `try/catch`：预览/非 Wails 环境下 `window.runtime` 不存在，`ClipboardSetText` 会直接抛。
组件里因此 import 运行时函数，`mock.js` 必须补 `window.runtime` 桩（见技能文档第 3 节）。

## 6. 样式

复用现有变量与观感，新增：

| class | 作用 |
| --- | --- |
| `.workspace-full` | `grid-column: 2 / -1`，让 workspace 横跨两列 |
| `.prompt-workspace` | 视图容器，flex 纵向、`flex:1;min-height:0`（内部滚动而不是整页滚动） |
| `.prompt-title-line` / `.back-button` | 标题行与返回按钮（浅蓝描边、品牌蓝文字，与 `.cooldown-button` 同族） |
| `.prompt-board` | 白色圆角面板，`flex:1;overflow:auto` |
| `.prompt-grid` | `repeat(auto-fill,minmax(380px,1fr))`，宽屏两列、窄屏一列 |
| `.prompt-card` / `.prompt-card-head` / `.prompt-actions` / `.prompt-copy` | 卡片与「复制 / 删除」按钮组 |

005 的 `.prompt-list` / `.prompt-item` / `.prompt-item-head` / `.prompt-editor` / `.prompt-add` / `.prompt-footnote`
只服务于右侧面板里的旧版式，随面板一起删除；`.prompt-delete` 保留（卡片里继续用）。

`.agent-caret` 从纯展示变成可点：加 `cursor:pointer`、`padding`、hover 态。

## 7. 验证方式

无 GUI 环境，按 `wails-vue-ui-verify` 技能走 mock 预览页 + Chrome 无头截图，
把行为断言渲染成顶部横幅：

| 视图 | 断言 |
| --- | --- |
| `agent`（默认态，不点任何东西） | 所有 `.agent-row` 的指示符都是 `▾`、`.agent-task-list` 数量 = Agent 数量 |
| `row`（只点行，不点指示符） | 行被选中（`.selected`）、展开区还在（没被收起） |
| `caret`（点指示符） | 该行收起（`.agent-task-list` 少一个）、指示符变 `▸` |
| `prompt`（点任务行） | `.workspace-full` 存在、右侧 `aside.details` **不存在**、有 `.back-button`、标题含任务名 |
| `back`（点返回） | `.workspace-full` 消失、`.details` 回来、展开状态与进入前一致 |
| `copy` | 空 prompt 点复制 → 调用 0 次且有提示；有内容点复制 → 传入的文本等于该条内容 |
| `switch` | 输入后立刻返回（不等防抖）→ 保存仍发生，payload 含新内容（第 4 节修复的回归点） |

断言里带上「不该发生的事」：右侧面板不出现、点复制不越界取到别的 prompt 的文本。
