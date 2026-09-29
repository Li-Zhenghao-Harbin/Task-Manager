# 021 任务清单

## 实现

- [x] `App.css` 文件末尾「滚动归属」一节重写：
      `.output-body` 改 `display:flex;flex-direction:column;gap:22px`（保留 `overflow-y:auto` 兜底）
- [x] 新增 `.output-cards{flex:0 0 auto;margin-bottom:0}` —— KPI 卡片行固定，间距交给 `gap`
- [x] 新增 `.chart-panel,.output-list{flex:1 1 auto;min-height:200px;overflow-y:auto;overscroll-behavior:contain}`
      —— 滚动条落在白卡片边框内侧
- [x] 新增 `.output-list h2{position:sticky;top:0;background:#fff;z-index:1}`
- [x] 「滚动归属」注释重写：说明 020 → 021 改了哪一条、为什么、兜底什么时候生效
- [x] `specs/020-list-scroll/spec.md` 顶部加「已被 021 修订」注记

本轮**无 Go 改动、无 App.vue 改动**，纯 CSS（构建后 JS 哈希不变：`index.b5b8df49.js`）。

## 验证

- [x] `npx vue-tsc --noEmit && npx vite build` 通过（CSS：`index.97b1491d.css` → `index.288271e8.css`）
- [x] 无头断言矩阵（真实数据 22 任务 / 4 应用，`--dump-dom` 抓横幅文本汇总）：

| 场景 | 窗口 | 结果 | 关键读数 |
| --- | --- | --- | --- |
| 产出概览·全部产出 | 1920x1040（1898x942）`tall` | **11/11 PASS** | 滚动对象 `.chart-panel`；卡片 362 溢出 70；`.output-body` **789/789 无滚动条** |
| 产出概览·全部产出 | 1440x900（1418x802）`tall` | **11/11 PASS** | 卡片 222 溢出 210；`.output-body` **649/649** |
| 产出概览·代码 | 1440x900 `tall` | **10/10 PASS** | 滚动对象 `.output-list`；卡片 502 未超出 |
| 产出概览·文档 | 1440x900 `tall` | **10/10 PASS** | 卡片 500/500 未超出 |
| 产出概览·实施 | 1440x900 `tall` | **10/10 PASS** | 卡片 502 未超出 |
| 产出概览·文档 | 1440x620（1418x522） | **10/10 PASS** | 卡片 222 溢出 261；兜底 **0px** |
| 产出概览·全部产出 | 1440x620 | **10/10 PASS** | 卡片压到 min-height 200，溢出 232；兜底 258（设计内退路） |
| 产出概览·全部产出 | 1024x680（1002x582） | **10/10 PASS** | 卡片 200 溢出 247；兜底 198 |
| Agent应用 | 1440x620 | **7/7 PASS** | 面板 367/641 溢出 274，滚动条同样在卡片里 |
| Agent应用 | 1920x1040 | **7/7 PASS** | 面板 787/787（内容未超出，只记录） |
| **[对照] 产出概览·全部产出** | 1920x1040 `regress=1` | **8/11**（3 条 FAIL） | 滚动对象退回 `.output-body`；`.output-body` 837>789；KPI 卡片 `top 121→73` |
| **[对照] 产出概览·文档** | 1440x620 `regress=1` | **9/11**（2 条 FAIL） | 滚动对象退回 `.output-body`；KPI 卡片 `top 121→-118` |

  正向 10 个场景全绿（合计 105 条断言）；对照组精确翻红 5 条，
  全部落在「滚动条长在白卡片里 / 内容区自己没有滚动条 / KPI 卡片固定」这三条上。

- [x] 截图 10 张存 `specs/021-output-panel-scroll/screenshots/`
      （全部产出 top/bottom、文档 top/bottom、Agent、对照组、断言横幅、小窗、最小窗、窄屏）

## 未做（明确排除）

- 未动 `.agents-panel` / `.units-panel` / `.task-list` / `.calendar-scroll` / `.prompt-board` / `.sidebar`
  —— 020 之后它们的形态就是「滚轮长在内容块自己的卡片里」，本轮不需要改。
- 未动 `aside.details`（右侧详情面板）的滚动。
- 未动 `max-width:1000px` / `1100px` 断点与其隐藏规则。
- 未引入自定义滚动条样式。
- 未改 KPI 卡片的栅格列数（仍是 3 列）——「7 张卡片 3 行 405px」这个高度决定了兜底阈值，
  但压缩列数会让卡片看起来更挤，不是本轮的范围。
