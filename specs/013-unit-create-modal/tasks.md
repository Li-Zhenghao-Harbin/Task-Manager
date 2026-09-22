# 013 · 任务清单

## 前端

- [x] `App.vue`：`newUnit` → `unitDraft`，新增 `showUnitCreate`
- [x] `App.vue`：`openUnitCreate()`（重置草稿 + 打开弹窗）
- [x] `App.vue`：`addUnit()` 改为提交 `unitDraft`、成功后关弹窗
- [x] `App.vue`：单位管理页 `<header>` 增加 `.actions > button.primary`（`＋ 新增单位`）
- [x] `App.vue`：删除面板里的 `.unit-add` 表单
- [x] `App.vue`：新增 `.unit-modal` 弹窗，只含「单位名」一个字段
- [x] `App.vue`：空态文案改为「暂无单位，点击右上角新增。」
- [x] `App.css`：删除 `.unit-add` / `.unit-add input` / `.unit-list` 三条规则
- [x] `App.css`：删除 `.units-workspace .unit-list{margin-top:18px}`
- [x] `App.css`：新增 `.unit-modal{width:min(430px,100%)}`（含说明注释）
- [x] `App.vue` 用 prettier 复排（单引号 / 无分号 / printWidth 120）；`App.css` 不参与

## 文档

- [x] `specs/013-unit-create-modal/`：`spec.md` / `design.md` / `tasks.md`
- [x] `outline.md`：新增「3.9 单位管理」小节；界面规划补一条单位页新建入口的说明

## 验证

- [x] `npm --prefix frontend run build`（`vue-tsc --noEmit` + `vite build`）
- [x] `gofmt -l` / `go vet ./...` / `go test ./...`（31 项）
- [x] 无头探针 —— 面板静态断言 **15/15**（宽屏 1440×900）
- [x] 无头探针 —— 弹窗结构断言 **22/22**（`?view=unitmodalopen`）
- [x] 无头探针 —— 完整交互流程 **32/32**（取消 / 空名 / 创建 / 列表刷新）
- [x] 无头探针 —— 同一套流程在 **1000×700 窄屏 32/32**
- [x] 无头探针 —— 空态 **5/5**（`?empty=1`）
- [x] 「全部任务」页回归截图（工具栏 3 个下拉、`＋ 新建任务` 位置不变）
- [x] `wails build`

## 证据截图

| 文件 | 内容 |
| --- | --- |
| `01-units-banner.png` | 单位面板断言 15/15（含与「＋ 新建任务」的几何对比） |
| `02-unit-modal-banner.png` | 弹窗结构断言 22/22（弹窗保持打开） |
| `03-unit-modal-flow.png` | 完整流程 32/32，列表末尾出现新建的「财务管理部」 |
| `04-units-clean.png` | 单位面板干净版式（无横幅） |
| `05-unit-modal-clean.png` | 弹窗干净版式（可见 `取消` / `创建单位`） |
| `06-empty-state.png` | 无单位时的空态与页头入口 |
| `07-narrow-1000.png` | 1000px 窄屏跑完 32 项断言 |
| `08-tasks-regression.png` | 「全部任务」页未受影响 |
