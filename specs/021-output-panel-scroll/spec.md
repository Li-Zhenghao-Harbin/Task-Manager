# 021 【产出概览】的滚动条要长在白卡片里

> 修订 [020](../020-list-scroll/spec.md)：020 把滚轮交给 `.output-body`，位置对了、观感不对。
> **⚠️ 本号那条 `flex:1 1 auto` 已被 [022](../022-output-panel-hug/spec.md) 修订（2026-09-29）**：
> 滚动归属不变，但卡片改成 `flex:0 1 auto`（内容未满时不再被拉满）。下面凡是提到
> `.chart-panel,.output-list{flex:1 1 auto}` 的地方，`flex-grow` 一律以 022 为准（= 0）。

## 背景

020 之后【产出概览】的滚动归属是 `.output-body` —— 那是本轮为「把页头与内容分开」而加的一层包裹：

```html
<div class="output-dashboard">
  <header>产出概览 / 全部产出</header>
  <div class="output-body">        <!-- 透明，无边框/底色/圆角 -->
    <div class="output-cards">…7 张 KPI 卡片…</div>
    <section class="chart-panel">…饼图…</section>
    <section class="output-list">…对应任务…</section>
  </div>
</div>
```

它在功能上完全正确（页头固定、整区不滚），但**滚动条画在这一层上**：既没有边框也没有底色，
`overflow-y:auto` 的滚动条就落在内容区最右边、自上而下一整条，看上去与「整页滚动条」无异 ——
而【AGENT】的滚动条长在 `.agents-panel` 这张白卡片（`border:1px + radius:13px + #fff`）的边框内侧，
一眼就知道是「列表自己的滚动条」。

用户报告：**「【AGENT】中的是对的，但是【产出概览】中的仍然是滚动整个页面而不是列表」**。
追问后确认：页头与卡片都不动，问题是**滚动条的位置**（「滚动条像整页的」），四个产出筛选下都注意到了。

## 做什么

把【产出概览】改成两段式，让滚轮落在**底部的白卡片**上（与 Agent / 单位管理同构）：

| 元素 | 021 之后 |
| --- | --- |
| 页头 `header` | 不动（020 起就固定） |
| KPI 卡片行 `.output-cards` | `flex:0 0 auto` —— 固定在页头下面，不再跟着内容滚 |
| 饼图卡片 `.chart-panel` / 对应任务卡片 `.output-list` | 自己滚：`flex:1 1 auto; min-height:200px; overflow-y:auto` |
| 包裹层 `.output-body` | 改成 `display:flex;flex-direction:column;gap:22px`，保留 `overflow-y:auto` 作为**兜底** |

具体改动（`App.css` 文件末尾「滚动归属」一节）：

1. `.output-body{flex:1 1 auto;min-height:0;display:flex;flex-direction:column;gap:22px;overflow-y:auto;overscroll-behavior:contain}`
2. `.output-cards{flex:0 0 auto;margin-bottom:0}`（间距由 `gap` 提供，避免与 `margin-bottom:22px` 叠加）
3. `.chart-panel,.output-list{flex:1 1 auto;min-height:200px;overflow-y:auto;overscroll-behavior:contain}`
4. `.output-list h2{position:sticky;top:0;background:#fff;z-index:1}` ——「对应任务」列表滚起来时标题留在卡片顶部

## 不做什么

- **不改 020 的其余结论**：`.workspace` 仍是 `overflow:hidden`、`.workspace>*{flex:0 0 auto}`、
  Agent / 单位 / 任务列表 / 日历 / Prompt / 侧边栏的滚动归属全部不动。
- **不动右侧详情面板**（`.details`）自身的滚动 —— 它是 `.shell` 的第三列，与本轮无关。
- **不改宽度断点**（`max-width:1000px` / `1100px`）与其隐藏规则。
- **不改任何 Go 代码 / 绑定 / 数据结构**。本轮是纯 CSS。
- 不引入自定义滚动条样式。

## 验收标准

断言全部是标量，在无头页面里用真实数据（`build/bin/task-manager-data.json`：22 条任务 / 4 个应用）量真值：

1. **滚动对象是白卡片**：`scroller` 不是 `.output-body`，而是 `.chart-panel`（全部产出）或 `.output-list`（代码/文档/实施）。
2. **这张卡片确实是卡片**：`border-width>0`、`border-radius≥8px`、`background` 非透明、`overflow-y=auto`
   —— 滚动条因此落在卡片边框内侧。
3. **内容区自己不再有滚动条**：`.output-body` 的 `scrollHeight == clientHeight`（窗口够高时），
   即不再出现「整页级」滚动条。
4. **页头固定**：滚动前后 `header` 的 `top` 相差 ≤ 1px。
5. **KPI 卡片行固定**：滚动前后 `.output-cards` 的 `top` 相差 ≤ 1px（020 下它会跟着滚）。
6. **卡片内容超出时能滚到末尾**：`scrollTop > 0`；内容未超出时只记录不判失败（避免假失败）。
7. **整区不给滚轮**：`.workspace` 的 `overflow-y=hidden` 且 `scrollHeight == clientHeight`。
8. **对照组必须失败**：同一构建、注入 020 的两条规则
   （`.chart-panel,.output-list{overflow:visible;min-height:0}` + `.output-body{display:block;overflow-y:auto}`），
   第 1、3、5 条必须 FAIL —— 否则断言等于自证。

**矮窗口的退路要在标准里写清楚**：KPI 卡片行本身就是 405px 高（7 张卡片 × 3 行），
当可视高度装不下「卡片行 + 200px 白卡片」时（约 CSS 高度 < 790px），`.output-body` 会兜底滚
—— 此时内容区重新出现整页级滚动条，这是设计内的退路，不是缺陷：那种尺寸下卡片行本身就超过一屏，
必须能滚。`min-height:200px` 的作用是保证这种时候下面的白卡片不会被压没（压到 0 高就点不到了）。

## 影响面

纯渲染层，无数据影响。`.output-body` 仍是【产出概览】独有元素，只是从「滚动容器」变成「布局容器 + 兜底」。
`specs/017` 版本号链路、`specs/018` 自启逻辑、`specs/019` 其余结论、`specs/020` 的 Agent / 单位 / 任务列表 /
日历 / Prompt 结论均不受影响。
