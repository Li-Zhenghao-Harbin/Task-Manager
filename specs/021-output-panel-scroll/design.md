# 021 设计

## 1. 问题定位：滚动条「画在谁身上」

020 的两段式结构是对的，问题在于**第二段选错了元素**：

```
.output-dashboard (flex column, flex:1 1 auto)
├── header                                   ← 固定 ✔
└── .output-body  overflow-y:auto            ← 滚动条画在这层 ✘
    ├── .output-cards    （7 张 KPI 卡片，405px）
    └── .chart-panel / .output-list          ← 白卡片，但它自己不滚
```

`.output-body` 是 020 为「把页头摘出去」而加的**透明包裹层**，没有任何卡片外观。
`overflow-y:auto` 的滚动条因此紧贴内容区右边缘（`.workspace` 的 `padding-right:34px` 之内），
从页头下面一直拉到窗口底部 —— 与窗口级滚动条在视觉上无法区分。

对照【AGENT】：`.agents-panel{background:#fff;border:1px solid var(--line);border-radius:13px;padding:8px}`
既是内容也是滚动容器，滚动条落在卡片边框**内侧**，边框/圆角把它「框」住了，一眼可辨。

所以修法是**把滚动交给下面那张白卡片**，`.output-body` 退化成布局容器：

```
.output-dashboard (flex column, flex:1 1 auto)
├── header                                   ← 固定
└── .output-body  （透明，flex column，gap:22px，overflow:auto 仅兜底）
    ├── .output-cards    flex:0 0 auto       ← 固定
    └── .chart-panel / .output-list          ← flex:1 1 auto; min-height:200px; overflow-y:auto ✔ 滚动条在卡片里
```

## 2. 具体改动（`App.css` 文件末尾「滚动归属」一节）

```css
.output-body{flex:1 1 auto;min-height:0;display:flex;flex-direction:column;gap:22px;overflow-y:auto;overscroll-behavior:contain}
.output-cards{flex:0 0 auto;margin-bottom:0}
.chart-panel,.output-list{flex:1 1 auto;min-height:200px;overflow-y:auto;overscroll-behavior:contain}
.output-list h2{position:sticky;top:0;background:#fff;z-index:1}
```

三点说明：

1. **`gap:22px` 顶替 `.output-cards` 的 `margin-bottom:22px`**：`.output-cards` 原本靠外边距与下面
   拉开 22px，改成 flex 后由 `gap` 提供；不把 `margin-bottom` 归零会叠成 44px。
2. **`min-height:200px` 而不是 `min-height:0`**：`.output-body` 仍可滚（兜底），如果把卡片的
   `min-height` 设成 0，矮窗口下 flex 会把它压到 0 高 —— 卡片消失、什么都点不到，
   而外层 `scrollHeight` 也不再反映它（内容已经没了）。给个下限，宁可让外层多滚一点。
3. **`.output-list h2` 用 `position:sticky`**：列表卡片自己滚之后，「对应任务」标题会跟着行滚走。
   它 `padding:18px` 且父级 `.output-list{padding:0}`，天然满宽，加个白底即可吸顶
   （与日历 `.calendar-week{position:sticky;top:0;background:#fff}` 同一手法）。
   **饼图卡片的 `h2` 不加 sticky**：`.chart-panel` 有 `padding:18px`，吸顶元素左右会露出
   18px 缝隙，内容从缝里穿过去更难看。

## 3. 兜底的量化边界

7 张 KPI 卡片 = 3 行 × 125px + 2 × 15px gap = **405px**，加 `gap:22px` + 卡片 `min-height:200px` = **627px**。
所以：

- `.output-body` 可视高 ≥ 627px（即窗口 CSS 高度 ≳ 790px）→ 只滚白卡片，内容区没有滚动条；
- 更矮 → `.output-body` 兜底滚（实测：1440x620 兜底 258px、1024x680 兜底 198px）。
  这类尺寸下卡片行本身已超过一屏，兜底不可避免；`min-height:200px` 保证白卡片仍以 200px 显示并可点。

代码/文档/实施三个筛选只有 1 行卡片（125px），627 → 347px，矮窗口下基本用不到兜底
（实测 1440x620 文档筛选：`.output-body` 369/369，兜底 0px）。

## 4. 探针与验证方法

复用用户级技能 `wails-vue-ui-verify` 的套路（mock 绑定 + 无头截图 + 断言横幅 + `--dump-dom` 抓文本），
探针放在 `.workbuddy/verify-021/`（不入库）：

- `mock.js` 的**数据来自真值**：构建脚本把 `build/bin/task-manager-data.json` 写成 `data.js`
  注入 `window.__DATA`，桩只做转发 —— 避免「界面读了桩的假数据」这种自证。
- 断言分两组：`outputFacts()`（021 的核心）与 `agentFacts()`（作为「正确形态」的基准口径）。
- `?regress=1` 注入 020 的两条规则复现修复前状态，同一构建、同一套断言必须精确翻红。
- `?tall=1` 表示窗口够高，此时才断言「内容区自己没有滚动条」。
- 「卡片内容未超出」时只 `note` 不判失败 —— 文档筛选在 1440x900 下确实没有溢出，
  硬断言「能滚到末尾」就是假失败。

### 本轮踩到并记进技能的两个坑

1. **截图里的滚动状态会「消失」**：程序化写 `scrollTop` 后，Vue 的一次重渲染（初始加载触发的
   自动保存回调）会重建节点、把滚动抹回 0。用 `setInterval` 复写仍然不稳 ——
   写很可能落在这一帧绘制之后，结果就是「`--dump-dom` 里 `scrollTop=70`、截图却完全没滚、
   top/bottom 两张图字节数一模一样」。
   改成 `requestAnimationFrame` 循环复写（写总是发生在下一帧绘制之前），且**每次重新
   `querySelector`**（旧节点已被摘掉，往游离节点写 `scrollTop` 是空操作），问题消失。
2. **探针元素不能静态插到 `body` 末尾**：给页面接一段内容会让文档凭空长出滚动条，
   污染「整页是否会给滚轮」的结论。调试用的 live 读数容器改成 `position:fixed;top:-200px;visibility:hidden`。

另外记录一条既有结论的复现：本机 `--headless=new` 的**截图视口与探针视口不一致**
（`--window-size=1920,1040` 时探针读到 `innerWidth/Height=1898x942`，截图却是 1920x1040）。
所以截图里的布局对应的是截图视口（≈1040 高），比探针读数（942）宽松；
**结论以断言读数为准，截图只作版式参考**。
