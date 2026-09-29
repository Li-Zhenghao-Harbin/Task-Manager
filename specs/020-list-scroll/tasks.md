# 020 任务清单

## 实现

- [x] `App.css` 第 1 行 `.workspace{padding:32px 34px;overflow:auto}` → `overflow:hidden`
- [x] `App.css` 第 7 行注释更新：整区不滚、滚动归属见文件末尾
- [x] `App.css` `.task-list` 注释更新（删掉「外层滚动条」的说法，点明它本来就有自己的滚轮）
- [x] `App.css` `.workspace.workspace-full` 注释更新（整区滚动已由 `.workspace` 关掉，此处为防御性）
- [x] `App.css` 文件末尾新增「滚动归属」区块：
      `.output-dashboard,.agents-workspace,.units-workspace{flex:1 1 auto;min-height:0;display:flex;flex-direction:column}`
      `.output-body{flex:1 1 auto;min-height:0;overflow-y:auto;overscroll-behavior:contain}`
      `.agents-panel,.units-panel{flex:1 1 auto;min-height:0;overflow-y:auto;overscroll-behavior:contain}`
- [x] `App.vue`【产出概览】卡片区与饼图/列表区外包一层 `<div class="output-body">`（prettier 重排过）

本轮**无 Go 改动**，无绑定重建。

## 验证

- [x] `npx vue-tsc --noEmit && vite build` 通过（`npm --prefix frontend run build`）
- [x] `go test ./...` 通过（41 项，本轮未新增）
- [x] `gofmt -l .` 无输出
- [x] `wails build -o TaskManager-verify.exe` 通过（17.2s），验完删除产物

无头断言矩阵（`--dump-dom` 抓横幅文本汇总，探针见 `design.md` 第五节）：

| 场景 | 窗口 | 结果 |
| --- | --- | --- |
| 产出概览 | 1440x620（1418x522） | **11/11 PASS** |
| Agent应用 | 1440x620 | **11/11 PASS** |
| 单位管理 | 1440x620 | **10/10 PASS** |
| 任务列表 | 1440x620 | **6/6 PASS** |
| 日历 | 1440x620 | **6/6 PASS** |
| Prompt | 1440x620 | **6/6 PASS** |
| 产出概览（高窗口） | 1440x1200（1418x1102） | **9/9 PASS**，滚动余量 0 |
| 单位管理（高窗口） | 1440x1200 | **8/8 PASS**，滚动余量 0 |
| Agent应用（高窗口） | 1440x900（1418x802） | **9/9 PASS**，滚动余量 0 |
| 任务列表（高窗口） | 1440x900 | **5/5 PASS**，滚动余量 0（不撑满） |
| 产出概览（小窗） | 900x600（878x502） | **11/11 PASS** |
| Agent应用（小窗） | 900x600 | **11/11 PASS** |
| 单位管理（小窗） | 900x600 | **10/10 PASS** |
| 任务列表（窄屏） | 1000x700（978x602） | **6/6 PASS** |
| Agent应用（窄屏） | 1000x700 | **11/11 PASS** |
| **[对照组] 产出概览** | 1440x620 `regress=1` | **4/11**（7 条 FAIL） |
| **[对照组] Agent应用** | 1440x620 `regress=1` | **4/11**（7 条 FAIL） |
| **[对照组] 单位管理** | 1440x620 `regress=1` | **3/10**（7 条 FAIL） |

正向 15/15 场景全绿（合计 130/130 条断言）；3 个对照组各精确翻红 7 条，
且失败项全部落在「整区不给滚轮 / 整区滚不动 / 根节点吃满高度 / 内容块自己有滚轮 /
滚得到末尾 / 页头不跟着内容走 / 末尾完整可见」。

对照组最有说服力的一条：**「页头不跟着内容走」的读数**
产出概览 `滚前 top=32.0px → 滚后 top=-426.0px`、Agent应用 `→ -143.0px`、单位管理 `→ -334.0px`
—— 这正是用户看到的「页头被滚走」。

- [x] 截图 24 张存 `specs/020-list-scroll/screenshots/`
      （含 header-fixed-top/bottom 对照、regress 对照组、高窗口、小窗、窄屏）
- [x] `specs/019-workspace-scroll/spec.md` 加注「已被 020 修订」并说明改了哪一条
- [x] `outline.md` 同步（界面规划 / 布局规则章节）

## 未做（明确排除）

- 未改 `aside.details`（右侧详情面板）的滚动 —— 它是 `.shell` 第三个栅格列，本就自滚。
- 未改 `max-width:1000px` / `1100px` 断点与其隐藏规则。
- 未引入自定义滚动条样式。
