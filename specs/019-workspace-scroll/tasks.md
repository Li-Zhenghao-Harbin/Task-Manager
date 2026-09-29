# 019 任务清单

> **⚠️ 已被 [020](../020-list-scroll/spec.md) 修订（2026-09-29）**：
> 第 1 条（`.workspace` 恢复 `overflow:auto`）被推翻 —— 滚轮落在最外层会把页头一起滑走。
> 020 让 `.workspace` 回到 `overflow:hidden`，改为各视图内部内容块自滚。
> 其余各条（`flex:0 0 auto`、任务列表/日历 flex 高度、`.sidebar` 自滚）仍然有效。

## 实现

- [x] `App.css` 第 7 行 `.workspace` 删掉 `overflow:hidden`（恢复第 1 行的 `overflow:auto`）
- [x] `App.css` 新增 `.workspace>*{flex:0 0 auto}`
- [x] `App.css` `.workspace-full` → `.workspace.workspace-full{grid-column:2 / -1;overflow:hidden}`
- [x] `App.css` `.task-list`：`max-height:calc(100vh - 190px)` → `flex:0 1 auto;min-height:0`
- [x] `App.css` `.calendar-view`：`height:calc(100vh - 185px)` → `flex:0 1 auto;min-height:0`
- [x] `App.css` 追加 `.sidebar{overflow-y:auto}`

本轮**无 Go 改动**，无绑定重建。

## 验证

### 构建链

- [x] `gofmt -l .` 无输出
- [x] `go test ./...` → `ok TaskManager 5.218s`（41 项，本轮未新增）
- [x] `npm --prefix frontend run build` → `vue-tsc --noEmit` 无报错 + vite 产物
      `assets/index.fce5da67.css`（JS 哈希未变，印证只改了样式）
- [x] `wails build -o TaskManager-verify.exe` → `Built in 17.055s`（验证产物已删除）

### 无头界面验证（1418x522，矮窗口）

| 视图 | 结果 | 关键读数 |
| --- | --- | --- |
| 产出概览 | **8/8** | `overflow-y=auto`，`scrollHeight=1012 > clientHeight=522`，`scrollTop=490/490`，容器高=内容高=940 |
| 产出概览（对照组） | 5/7 | **FAIL** `overflow-y=hidden`；**FAIL** `scrollTop=0/490` |
| Agent应用 | **7/7** | `scrollHeight=1144`，`scrollTop=622/622`，末尾模块 bottom=490 ≤ 可视底 522 |
| Agent应用（对照组） | 4/6 | **FAIL** `overflow-y=hidden`；**FAIL** `scrollTop=0/622` |
| 任务列表 | **3/3** | `workspace scrollHeight=522 == clientHeight=522`（原为 549，27px 冗余已消除） |
| 日历 | **4/4** | 日历 bottom=490 ≤ 可视底 522（原溢出 21px），月格子区 `scrollTop=865/865` |
| Prompt 视图 | **3/3** | 外层 `overflow-y=hidden`、看板 `overflow-y=auto`、横跨两列 |

### 其它尺寸回归

| 视图 | 尺寸 | 结果 |
| --- | --- | --- |
| 任务列表 | 1418x802 | 3/3，`scrollTop` 上限 = 0 → **内容不足时不撑满**，白卡片贴合内容 |
| 任务列表 | 978x602 | 3/3，窄屏右侧详情面板按断点隐藏 |
| 产出概览 | 878x502 | 8/8 |
| Agent应用 | 878x502 | 7/7 |
| Prompt 视图 | 878x502 | 3/3 |

### 对照组做法（防自证）

不用另建一份"修复前"的构建：在探针里注入

```css
.workspace{overflow:hidden !important}
.workspace>*{flex-shrink:1 !important}
```

即复现修复前的关键规则。同一构建、同一套断言，唯一变量就是 `overflow` 的取值，
结果必须从全绿变成两条 FAIL —— 这才证明断言真的在测这件事。

（附带发现：`flex-shrink:1` 注入后子项**没有**被压扁，因为 flex 子项默认
`min-height:auto` 已经挡住了压缩。所以真正敏感的是 `overflow` 那两条断言。）

### 截图

`specs/019-workspace-scroll/screenshots/`（13 张，含两张对照组）。
带断言横幅的图直接读出通过数；`*-clean.png` 是无横幅的干净版式图。

## 遗留

- [ ] `outline.md` 同步（已在本次完成）
- [x] 项目记忆（当日日志 + `MEMORY.md`）
