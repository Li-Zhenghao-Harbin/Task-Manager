# 022 设计：只改一个 flex-grow

## 问题定位

021 定稿的规则（022 只动最后一行）：

```css
.output-body{flex:1 1 auto;min-height:0;display:flex;flex-direction:column;gap:22px;overflow-y:auto}
.output-cards{flex:0 0 auto}
.chart-panel,.output-list{flex:1 1 auto;min-height:200px;overflow-y:auto}   ← 改这里
```

`.output-body` 是纵向 flex 容器，卡片带 `flex-grow:1`，所以**只要有富余空间，卡片就会把它吃光**。
内容少时表现为「卡片被拉得很长、最后一行下面一大段空白」。

## 为什么只改 flex-grow 就够

flex 的三件事分工正交，改一处只动一个自由度：

| 属性 | 什么时候起作用 | 本轮处置 |
| --- | --- | --- |
| `flex-grow` | 有**富余空间**时 | 1 → **0**（唯一改动） |
| `flex-shrink` | **空间不足**时 | 保持 1：内容超长时仍要能被压缩、露出滚动条 |
| `flex-basis` | 初始主轴尺寸 | 保持 `auto`：以内容高度为基准 |

推论 —— 也是本轮最省事的地方：**压缩侧读数一字不变**。空间不足时 `flex-grow` 根本不参与分配，
所以 1920x1040 下卡片仍是 362 高、溢出 70，与 021 的报告完全相同（已复测）。

## 为什么参照 `.task-list` 而不是自己发明数值

`.task-list`（【全部任务】的列表）是全项目唯一被反复验证过的「页头 + 列表」写法：

```css
.task-list{flex:0 1 auto;min-height:0;overflow-y:auto;overscroll-behavior:contain}
```

同一套语义，所以直接照搬，不改数值。断言也没有停在「`flex-grow` 等于 0」，而是**从 CSSOM 里读出
`.task-list` 声明的 flex 值、与卡片的计算值比对** —— 把用户那句「参考【全部任务】」直接变成可执行判据。

> 读 CSSOM 要注意：`.task-list` 出现在两条规则里（第 1 行那条是一长串选择器、不含 flex），
> 必须挑「真的声明了该属性」的那条。见 `view.js` 的 `declared()`。

## `min-height:200px` 为什么保留（实测，不能照抄 `.task-list`）

「参考 `.task-list`」很容易顺手把 `min-height` 也改成 0（`.task-list` 就是 0），**但这里不能照抄**：

- `.output-body` 在最小窗口（1024x680）下只有 **429px** 高，而 KPI 卡片行固定占 **405px**，
  再加 `gap:22px`，留给卡片的只剩几像素。
- 实测（`?floor0=1` 只摘掉 `min-height`）：饼图卡片被压到 **38px**，`clientHeight` 只剩 **21px**
  —— 等于看不到也点不到。

所以保留 200px 下限，让 `.output-body` 的兜底滚动接管（该窗口下兜底滚 **198px**）。
两条规则各管一段：**下限管「不许被压没」，`flex-grow:0` 管「不许被拉长」**。

## 验证方法

复用 021 搭好的无头预览页（已从 `.workbuddy/verify-021/` 改名为 `.workbuddy/verify-ui/`，
因为这套脚手架已不只服务 021）。探针改动：

- `?regress=1` → 注入 `.chart-panel,.output-list{flex:1 1 auto !important}`，复现 022 之前；
- `?regress020=1` → 保留 021 的对照组（滚动挂 `.output-body`），让旧证据仍可复现；
- `?floor0=1` → 只摘掉 `min-height:200px`，量这把下限是否真的必需；
- 核心判据用「卡片下方还剩多少」：`freeBelow = body.bottom - card.bottom`。
  富余空间被卡片吃掉时是 0，留在卡片外时 `> 0`。这个量比直接量卡片高度更能表达「撑没撑满」。

### 踩到的三个坑

1. **断言里没出现「被测对象」，等于自证。** 第一版把「与 `.task-list` 同一套 flex」写成
   `listFlex === '0 1 auto'` —— 比的是 `.task-list` **自己**的声明值，与卡片无关，于是对照组里
   它照样绿。改成 `cardFlex === listFlex` 后，对照组从 2 条红变成 3 条红。**这是最值得记住的一条。**
2. **横幅里的「N 条断言」注水。** 原来写 `lines.length`，把 `note` 读数也算成了断言
   （显示 17，实际只有 11 条）。已单独计数，现在写「11 条断言（另有 6 条读数）」。
3. **截图视口 ≠ 探针视口（老坑复用）。** `--window-size=1440,900` 时探针读到 1418x802，截图却是
   1440x900。所以截图里卡片下方的留白看着比断言读数（19px）大得多 —— **不矛盾**：卡片高度是内容
   驱动的（两种视口下都是 483），视口越高留白越多。结论以读数为准，截图只作版式参考。
