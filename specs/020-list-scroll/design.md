# 020 设计：滚动归属

## 一、为什么不能再让 `.workspace` 滚

`.workspace` 是「页头 + 内容」的**共同父节点**。它一滚，页头就是内容的一部分，必然被带走。
要「滚列表不滚页头」，滚动容器必须**位于页头之下**，且该容器与页头是兄弟关系。

```
.shell (grid, 100vh, overflow:hidden)
├─ .sidebar                          ← 自滚（019）
├─ .workspace  overflow:hidden       ← 不滚（本轮改回）
│   └─ .output-dashboard  flex:1, min-height:0, column
│       ├─ header                    ← 固定：吃 flex:0 0 auto
│       └─ .output-body  flex:1, min-height:0, overflow-y:auto   ← 只有它滚
├─ aside.details  overflow:auto      ← 自滚（第三个栅格列，与 .workspace 无关）
```

## 二、flex 的两个必要条件

内容块要能滚，光写 `overflow-y:auto` 不够，必须同时满足：

```css
.output-body{flex:1 1 auto;min-height:0;overflow-y:auto}
```

- **`min-height:0` 不能省。** flex 子项的 `min-height` 默认是 `auto`（content-based minimum），
  它会拒绝被压到内容高度以下 —— 于是容器只长不缩，永远不比内容矮，滚动条**永远不出现**，
  内容只是被裁掉。这一条是本项目反复踩的坑（`.task-list`、`.agent-task-options` 同理）。
- **`flex:1 1 auto`** 而不是 `flex:0 0 auto`：内容块要吃掉页头之外的剩余高度，
  这样「滚轮的可视区」正好是屏幕上的空白区，不会在列表下方留一条空白带。
  （`.workspace>*{flex:0 0 auto}` 会把默认值覆成不伸不缩，所以必须在文件末尾用同特异性、
  更靠后的规则改回来。）

## 三、为什么【产出概览】要多包一层

需求是「页头不滚、下面的卡片和图表滚」。而原结构里 `.output-cards` 与
`section.chart-panel | section.output-list` 是 `<header>` 的**兄弟**，三者平级：

```
.output-dashboard
├─ header
├─ .output-cards
└─ section.chart-panel / section.output-list
```

没有中间容器，就无法「让卡片与图表一起滚、页头不滚」——只能逐个加 `overflow`
（两个滚动条）或者连页头一起滚（回到 019）。所以加一层 `.output-body` 包住后两者。
Agent / 单位视图本来就有面板容器（`.agents-panel` / `.units-panel`），不需要新包裹层。

## 四、为什么单位管理也一起改

用户只点名了【产出概览】和【AGENT】。但【单位管理】与它们同构（页头 + 面板列表），
`.workspace` 一旦回到 `overflow:hidden`，它就会退回 019 之前的状态（矮窗口下被裁掉、无法滑动）。
所以三个视图统一处理，避免「修好两个、坏掉一个」。

## 五、探针设计（无头验证）

用 `wails-vue-ui-verify` 技能：mock 绑定 + 静态预览 + Chrome 无头。

数据量刻意放大，保证矮窗口下三个视图都真实超长：

| 视图 | 内容高（1418x522） | 可视高 |
| --- | --- | --- |
| 产出概览 `.output-body` | 859 | 369 |
| Agent应用 `.agents-panel` | 583 | 367 |
| 单位管理 `.units-panel` | 788 | 367 |

断言分三层，逐层排除误判：

1. **整页层**：`document.scrollingElement` 不可滚 —— 直接对应「不要给整个页面加滚轮」。
2. **整区层**：`.workspace` `overflow-y=hidden` 且滚不动 —— 对应「不要给整区加滚轮」。
3. **内容层**：内容块可滚 + 页头 `top` 不变 —— 对应「滚轮在列表上、页头留在原位」。

关键实现细节（都是踩过坑才写对的）：

- **页头基准值必须在任何滚动之前取。** 先量 `header.top`，再做「整区滚不动」的检查
  （那一步会把 `.workspace.scrollTop` 设成 `scrollHeight`）。若顺序反了，对照组里
  「页头是否跟着走」的前后两次读数都已经带了同样的偏移，断言恒真、失去灵敏度。
- **对照组不重建产物**：`?regress=1` 注入两条 CSS 复现修复前状态，同一构建、同一套断言，
  只有 7 条应当翻红，用来证明断言真在测这件事。
- **`?fit=1` 场景要按内容真实高度选窗口尺寸**：产出概览内容 859px，
  1440x900（可视 647）**装不下**，用 `fit=1` 跑必然假失败；改成 1440x1200（可视 947）才成立。
  估算高度不如先量一次。
- **探针里的变量名别叫 `scroll`**：`window.scroll` 是浏览器遗留函数，
  `params.get('scroll') === 'top'` 与它比较前若漏了 `params` 前缀，会静默拿到函数、分支永不成立。
  症状是「两张本该不同的干净图字节数完全相同」。
- **断言结果渲染进悬浮横幅**（红/绿），截图自带结论；另拍 `?pos=none` 的干净版式图。
  横幅文本也在 DOM 里，所以能用 `--dump-dom` 抓回来做批量汇总，比逐张读图便宜得多。

## 六、风险

- 面板变成固定高度的滚动视口后，矮窗口下白色面板会**撑满剩余高度**（`flex:1`）。
  这是列表型 UI 的常见观感，且 019 之后本来就已如此（`flex:0 1 auto` 只影响任务列表的贴合）。
- `overscroll-behavior:contain` 防止滚到边界后把滚动传给父级 —— 因为父级 `.workspace`
  已经不可滚，这里属于防御性设置。
