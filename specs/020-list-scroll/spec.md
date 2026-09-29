# 020 滚动条只挂在列表上（页头固定）

> **⚠️ 【产出概览】那一行已被 [021](../021-output-panel-scroll/spec.md) 修订（2026-09-29）**
>
> 020 的方向对（页头固定 + 内容块自滚），但产出概览选的滚动容器是 `.output-body` ——
> 那是**透明无边框的包裹层**，滚动条因此画在内容区最右边、上下贯通一整条，看上去就是「整页滚动条」，
> 与 Agent 的 `.agents-panel`（滚动条落在白卡片边框内侧）观感不一致。
>
> 021 把产出概览改成两段式：KPI 卡片行固定，滚动条只长在下面那张白卡片
> （`.chart-panel` 饼图 / `.output-list` 对应任务）上，`.output-body` 退化为布局容器 + 矮窗口兜底。
> **Agent / 单位管理 / 任务列表 / 日历 / Prompt 的结论不受影响**，下文其余内容仍有效。

## 背景

019 把整区滚动还给了 `.workspace`，解决了「矮窗口下内容看不到」。
但滚轮落在了**最外层**：一滚，页头（`产出概览 / 全部产出` 标题、右上角 `＋ 新增 …` 主按钮）
跟着内容一起被滑走，看不出当前在哪个视图、也点不到主操作。

用户要求：**只在列表里加滚轮，不要给整个页面加滚轮。**

## 做什么

把滚动从「整区」下移到「内容块」，页头留在原位：

| 视图 | 滚轮挂在 | 页头 |
| --- | --- | --- |
| 产出概览 | `.output-body`（新增的包裹层，包住卡片 + 饼图/任务列表） | `.output-dashboard > header` 固定 |
| Agent应用 | `.agents-panel`（整块列表面板） | `.agents-workspace > header` 固定 |
| 单位管理 | `.units-panel`（整块列表面板） | `.units-workspace > header` 固定 |
| 任务列表 | `.task-list`（019 起就是这样，不变） | 工具栏/页头固定 |
| 日历 | `.calendar-scroll`（019 起就是这样，不变） | 月份行固定 |
| Prompt | `.prompt-board`（006 起就是这样，不变） | 标题行固定 |

具体改动：

1. `.workspace` 回到 `overflow:hidden`（第 1 行），整区**不滚**；`.workspace>*{flex:0 0 auto}` 保留。
2. 三个视图的根节点 `.output-dashboard` / `.agents-workspace` / `.units-workspace`
   改 `flex:1 1 auto;min-height:0;display:flex;flex-direction:column` —— 吃满 `.workspace`
   的剩余高度并纵向排布，页头在上、内容块在下。
3. 内容块 `.output-body` / `.agents-panel` / `.units-panel` 加
   `flex:1 1 auto;min-height:0;overflow-y:auto;overscroll-behavior:contain`。
4. `App.vue`：【产出概览】的卡片区与饼图/列表区外面包一层 `<div class="output-body">`
   （页头与内容块是兄弟节点，不包一层就没法只滚下面那块）。

## 不做什么

- **不改 019 已经修好的三件事**：`.workspace>*{flex:0 0 auto}`、
  任务列表/日历的 flex 自适应高度（替掉 `calc(100vh - Npx)`）、`.sidebar{overflow-y:auto}`。
- **不动右侧详情面板**（`.details`）自身的滚动 —— 它是 `.shell` 的独立栅格列，一直自滚，与本轮无关。
- **不改宽度断点**（`max-width:1000px` / `1100px`）与其隐藏规则。
- **不改任何 Go 代码 / 绑定 / 数据结构**。纯 CSS + 一个包裹层。
- 不引入自定义滚动条样式。

## 验收标准

断言全部是标量，在无头页面里量真值（探针见 `design.md`）。矮窗口 1440x620（实际内区 1418x522）：

1. **整页本身没有滚轮**：`document.scrollingElement` 的 `scrollHeight == innerHeight`，
   把它的 `scrollTop` 设成 99999 仍是 0。
2. **主内容区不给滚轮**：`.workspace` 的 `overflow-y=hidden`，`scrollTop` 设成 `scrollHeight` 后仍为 0。
3. **视图根节点吃满剩余高度**：`root.bottom` 与「内容区下边缘（`.workspace` bottom 减 paddingBottom）」
   相差 ≤ 2px。
4. **内容块自己有滚轮**：`overflow-y=auto`、`scrollHeight > clientHeight`，滚得到末尾（`scrollTop > 0`）。
5. **页头不跟着内容走**：滚前 / 滚后 `header` 的 `top` 相差 ≤ 1px。
6. **滑到底后末尾模块完整可见**：末位元素 `bottom ≤ 内容块 bottom`。
7. **窗口够高时不撑出滚动余量**：内容不超出时滚动余量为 0（内容块是 `flex:1` 的**滚动视口**，
   但它不会把内容拉高；`.task-list` 仍保持 `flex:0 1 auto`，贴合内容高度）。
8. **对照组必须失败**：同一构建、注入「修复前」的两条规则
   （`.workspace{overflow:auto}` + 内容块 `overflow:visible`），
   第 2/3/4/5/6 条必须 FAIL —— 否则断言等于自证。

覆盖矩阵（15 个正向场景 + 3 个对照组，见 `tasks.md`）：产出概览 / Agent应用 / 单位管理 /
任务列表 / 日历 / Prompt × 矮窗口、高窗口、小窗、窄屏。

## 影响面

纯渲染层。`.output-body` 是新增元素，只出现在【产出概览】。
`specs/017` 的版本号链路、`specs/018` 的自启逻辑、`specs/019` 的其余结论均不受影响。
