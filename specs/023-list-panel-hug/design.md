# 023 设计：为什么「贴合内容」要靠 flex 三件事同时到位

## 1. 三类视图的 DOM / flex 结构

```
.workspace                       display:flex; flex-direction:column; overflow:hidden; padding:32px 34px
│                                .workspace>*{flex:0 0 auto}
├─ header                        页头（标题 + 右上角主按钮）—— 永远 flex:0 0 auto
├─ (任务视图) .toolbar
├─ .task-list                    ← 参照物：flex:0 1 auto; min-height:0; overflow-y:auto
├─ .output-dashboard             flex:1 1 auto; min-height:0; display:flex; column
│   ├─ header
│   └─ .output-body              flex:1 1 auto; display:flex; column; gap:22px; overflow-y:auto（仅兜底）
│       ├─ .output-cards         flex:0 0 auto（7 张卡 3 列 3 行 = 405px；列表筛选只有 1~3 张 = 125px）
│       ├─ .chart-panel          flex:0 1 auto; min-height:200px   ← 唯一保留下限的卡片
│       └─ .output-list          flex:0 1 auto; min-height:0       ← 023 改掉 200px
├─ .agents-workspace             flex:1 1 auto; min-height:0; display:flex; column
│   ├─ header
│   └─ .agents-panel             flex:0 1 auto; min-height:0       ← 023 由 1 1 auto 改来
└─ .units-workspace              flex:1 1 auto; min-height:0; display:flex; column
    ├─ header
    └─ .units-panel              flex:0 1 auto; min-height:0       ← 023 由 1 1 auto 改来
```

「贴合内容」= 卡片同时满足：

1. `flex-grow:0` —— 有富余空间也不长高（022 已做）；
2. `min-height:0` —— 盒子不被自己的下限撑到内容之上（023 补的）；
3. `flex-shrink:1` + `overflow-y:auto` —— 空间不足时压缩并自己出滚动条（021/019 已做）。

三者缺一都会有用户可见的毛病：

| 缺哪一条 | 症状 | 实测 |
| --- | --- | --- |
| flex-grow（=1） | 装得下也被拉满 | Agent 面板 789 / 内容 634 |
| min-height（=200px） | 短列表被下限撑高，**卡片内部**留白 | 实施卡片 200 / 内容 132 → 内部空白 68 |
| shrink / overflow | 空间不足时被压扁且无法滚动 | 饼图卡片会压到 38px |

## 2. 为什么 022 的判据漏掉了这些

022 的判据是「卡片**下方**还剩多少空间」（`body.bottom - card.bottom`）：

- 【实施】卡片 200 高、下方留白 302px → 「留白在卡片外」✓ 通过；
- Agent 面板 789 高、下方留白 0px → 被 `overflowing` 分支判为「吃满可用高度」✓ 通过。

两条都被判对了**位置**，却没判**内容**：卡片内部那 68px / 147px 是空的。
023 换成量「最后一个子元素的底边到卡片底边的距离，再扣掉卡片自己的 `padding-bottom`」：

```js
function innerGap(cardEl, lastSel) {
  var rect = cardEl.getBoundingClientRect()
  var kids = qa(lastSel, cardEl)
  var last = kids[kids.length - 1]
  var pb = parseFloat(getComputedStyle(cardEl).paddingBottom) || 0
  var raw = rect.bottom - last.getBoundingClientRect().bottom
  return { raw: raw, padding: pb, gap: raw - pb, content: last.getBoundingClientRect().bottom - rect.top }
}
```

- 贴合 → `gap` = 1（只剩 1px 下边框）；
- 被撑高 → `gap` = 68 / 147 / 620。

**不能用 `scrollHeight` 当内容高**：盒子比内容高时 `scrollHeight` 返回的是 `clientHeight`，
会自证成功（实施卡片 200 高、内容 132，`scrollHeight` 照样报 200）。

## 3. CSS 改动（`frontend/src/App.css`）

`022` → `023` 的净变化只有两条声明：

```diff
-.chart-panel,.output-list{flex:0 1 auto;min-height:200px;overflow-y:auto;overscroll-behavior:contain}
-.agents-panel,.units-panel{flex:1 1 auto;min-height:0;overflow-y:auto;overscroll-behavior:contain}
+.chart-panel{flex:0 1 auto;min-height:200px;overflow-y:auto;overscroll-behavior:contain}
+.output-list,.agents-panel,.units-panel{flex:0 1 auto;min-height:0;overflow-y:auto;overscroll-behavior:contain}
```

拆成两条的原因：`.output-list` 与两个面板的声明**完全相同**，且都等于 `.task-list`；
`.chart-panel` 是例外，必须留 `min-height:200px`。合成一条会把例外一起抹掉。

为什么 `.chart-panel` 能保留下限而 `.output-list` 不能：饼图的内容高度是**常数**（`.pie-layout{min-height:310px}`），
下限 200px 永远小于内容，只在极端窗口下起「防压扁」作用；列表的内容高度是**变量**（1 行到 22 行），
一旦下限大于内容，下限就变成「撑高」。

## 4. 量化边界（本机实测，视口 = 窗口高 − 98）

| 场景 | 卡片行 | 卡片 | 内容 | 卡片内空白 | 卡片外留白 | `.output-body` |
| --- | --- | --- | --- | --- | --- | --- |
| 实施 @1440x900（改前） | 125 | 200 | 132 | **68** | 302 | 649/649 不溢出 |
| 实施 @1440x900（改后） | 125 | **133** | 132 | **1** | 369 | 649/649 |
| 代码 @1440x900 | 125 | 475 | 474 | 1 | 27 | 649/649 |
| 文档 @1440x900 | 125 | 483 | 482 | 1 | 19 | 649/649 |
| 全部产出 @1920x1040 | 405 | 362（溢出 69） | 413 | 溢出 | 0 | 789/789 |
| 全部产出 @1920x1200 | 405 | 432 | 413 | 1 | 90 | 949/949 |
| 文档 @1024x680 | 125 | 282 | 537 | 溢出 255 | 0 | 429/429 |
| 全部产出 @1024x680 | 405 | **200（下限）** | 413 | 溢出 213 | −198 | 429/627 兜底滚 198 |
| Agent @1920x1040（改前） | — | 789 | 634 | **147** | 0 | — |
| Agent @1920x1040（改后） | — | **643** | 634 | **1** | 146 | — |
| Agent @1440x620（改后） | — | 369 | 634 | 溢出 273 | 0 | — |
| 单位 @1920x1040（改前） | — | 789 | 147 | **620** | 0 | — |
| 单位 @1920x1040（改后） | — | **170** | 147 | **1** | 619 | — |
| 全部任务 @1920x1040（参照物） | — | 725 | 1447 | 溢出 722 | 0 | — |

`.output-body` 的兜底阈值（不变）：卡片行 405 + gap 22 + 饼图下限 200 = 627，
加 `.workspace` 内边距与页头占掉的 153px ⇒ 窗口 CSS 高 < 约 780px 时生效。

## 5. 探针与两个新坑

- 复用 `.workbuddy/verify-ui/`（mock 绑定 + `window.runtime` 桩 + Chrome 无头截图 + 断言横幅 +
  `--dump-dom` 抓断言文本）。本轮新增 `view=tasks`，把参照物 `.task-list` 也纳入同一组判据。
- **坑 1：选择器必须命中真正排在最下面的那一项。** Agent 面板里 `template v-for` 的展开块
  `.agent-task-list` 排在 `.agent-row` **之后**，只量最后一个 `.agent-row` 会凭空多出 49px「假空白」
  （第一版就这么误判成「贴合失败」）。改成「容器的最后一个子元素」（`.agent-list > *`）。
- **坑 2：`rootRect.bottom` 不等于可用高度底边。** `view=tasks` 的根节点就是 `.workspace`，
  自带 32px 下内边距，不扣掉的话列表明明已经吃满仍报「留白 32px」。
- **坑 3：溢出判据写成 `height > content` 会整条失效。** 被压缩的卡片满足
  `height < content`，会被判成「未超出」而走进「贴合」分支（022 遗产，本轮两处都修了）。
- 产物校验：`App.vue` 未动 ⇒ JS 哈希仍为 `index.b5b8df49.js`；CSS 注释被 minifier 剥掉 ⇒
  只改注释时 CSS 哈希也不变（`index.403a7970.css`），可用它反证「注释改动没有行为影响」。
