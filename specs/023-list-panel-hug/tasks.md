# 023 任务清单与验证矩阵

## 实现清单

- [x] `frontend/src/App.css`：`.output-list` 从 `.chart-panel,.output-list` 那条里拆出来，`min-height` 由
      `200px` 改为 `0`；`.chart-panel` 单独一条保留 `min-height:200px`。
- [x] `frontend/src/App.css`：`.agents-panel,.units-panel` 的 `flex:1 1 auto` 改为 `flex:0 1 auto`。
- [x] `frontend/src/App.css`：顶部「滚动归属」注释重写 —— 补上 023 的结论
      （卡片一律贴合内容、三张列表卡与 `.task-list` 逐字相同、`.chart-panel` 是唯一例外），
      并修正 022 注释里含糊的兜底算式（卡片行 7 张 = 405px，列表筛选只有 1~3 张 = 125px）。
- [x] `.workbuddy/verify-ui/view.js`：探针增加「卡片**内部**空白」判据、`view=tasks`（参照物），
      并修正两处失效的溢出判据 + 一个假空白选择器（见 design.md 第 5 节）。
- [x] `specs/023-list-panel-hug/{spec,design,tasks}.md` + `screenshots/`（12 张）。
- [x] `specs/022-output-panel-hug/spec.md` 顶部加注「已被 023 修订」。
- [x] `outline.md`「滚动归属」一节同步到 023 定稿；`.workbuddy/memory/` 同步。
- [x] 技能 `wails-vue-ui-verify` 补「内部空白」判据与三个新坑。

## 验证矩阵（正向 17 个场景 / 159 条断言，0 失败）

| view | 窗口 | 筛选 | 断言 | 关键读数 |
| --- | --- | --- | --- | --- |
| output | 1440x900 | **实施** | 13/13 | 卡片 **133**、内容 132、**内部空白 1**、外留白 369 |
| output | 1440x900 | 代码 | 13/13 | 卡片 475、内容 474、内部空白 1、外留白 27 |
| output | 1440x900 | 文档 | 13/13 | 卡片 483、内容 482、内部空白 1、外留白 19 |
| output | 1920x1040 | 全部产出 | 12/12 | 饼图卡片 362（溢出 69）、内部空白 −69、`.output-body` 789/789 |
| output | 1920x1200 | 全部产出 | 11/11 | 饼图卡片 432、内部空白 1、外留白 90 |
| output | 1440x900 | 全部产出 | 12/12 | 饼图卡片 222（压缩）、`.output-body` 649/649 不溢出 |
| output | 1024x680 | 文档 | 13/13 | 卡片 282（压缩出内滚 255）、`.output-body` 429/429 |
| output | 1024x680 | 全部产出 | 11/11 | 饼图卡片 **200 = 下限**、`.output-body` 429/627 兜底 198 |
| output | 1440x760 | 文档 | 13/13 | 卡片压缩 + 内部滚动 |
| **agent** | 1920x1040 | — | 5/5 | 面板 **643**、内容 634、**内部空白 1**、外留白 146（改前 789 / 147 / 0） |
| agent | 1440x900 | — | 5/5 | 面板 643、内部空白 1、外留白 6 |
| agent | 1440x620 | — | 7/7 | 面板压到 369、内部溢出 273、能滚到末尾、页头固定 |
| agent | 1024x680 | — | 7/7 | 同上（更小窗口） |
| **units** | 1920x1040 | — | 5/5 | 面板 **170**、内部空白 1、外留白 619（改前 789 / 620 / 0） |
| units | 1024x680 | — | 5/5 | 面板 170、外留白 259 |
| tasks（参照物） | 1920x1040 | 全部任务 | 7/7 | `.task-list` 725、内容 1447、内部滚动 |
| tasks（参照物） | 1440x620 | 全部任务 | 7/7 | `.task-list` 305、内部滚动 |

每个场景都同时断言：三张卡的 `flex` 与 `min-height` 逐项等于 `.task-list` 的声明值、卡片是带边框/圆角/
底色的滚动容器、整区不溢出、页头滚动前后 `top` 不变。

## 对照组（7 个场景，精确翻红）

| 注入的旧规则 | view / 场景 | 结果 | 翻红的判据 |
| --- | --- | --- | --- |
| **022**：`.output-list{min-height:200px}` | output 1440x900 · 实施 | **2 失败** / 13 | 内部空白回到 **68**；`min-height` 与 `.task-list` 不一致 |
| **022**：同上 | output 1440x900 · 文档 | 1 失败 / 13 | 只剩 `min-height` 一条（文档本来装不下限，可见 022 只对短列表有害） |
| **021**：`flex:1 1 auto` | agent 1920x1040 | **2 失败** / 5 | `flex=[1 1 auto]`；内部空白回到 **147**，外留白 0 |
| **021**：同上 | units 1920x1040 | **2 失败** / 5 | 内部空白回到 **620** |
| **021**：同上 | output 1440x900 · 实施 | **5 失败** / 13 | 内部空白 **370**、`flex-grow=1`、`flex` 不一致、`min-height` 不一致、外留白 |
| **020**：滚动挂 `.output-body` | output 1920x1040 · 全部产出 | **4 失败** / 11 | 滚动对象退回 `.output-body`、饼图下限被抹、`.output-body` 837>789、KPI 行 `top 121→73` |
| **020**：同上 | output 1440x900 · 文档 | 1 失败 / 13 | 滚动对象退回 `.output-body`（旧证据仍在） |

## 构建链路

`gofmt -l .` 无输出 · `go vet ./...` 无输出 · `go test ./...` ok · `npx vue-tsc --noEmit` + `npx vite build` 通过 ·
`wails build -s -nocolour` 成功 · exe 内嵌 CSS 已 `grep -a` 核对为 023 版本
（`.output-list,.agents-panel,.units-panel{flex:0 1 auto;min-height:0`，且已无 `flex:1 1 auto` 的面板规则）。

## 未做项

- `.calendar-scroll` / `.prompt-board` 未纳入本轮判据（它们不是「列表卡」，本轮没动）。
- 未改单位管理列表的交互（「修改」仍是 `prompt()`，「设为默认/删除」仍是行内按钮）。
