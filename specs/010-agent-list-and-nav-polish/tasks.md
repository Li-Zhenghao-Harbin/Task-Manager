# 010 · 任务分解

## T1 规格

- [x] `spec.md` / `design.md` / `tasks.md`

## T2 前端实现

- [x] `App.vue`：删掉行内 `.agent-projects` span
- [x] `App.vue`：`.agent-reset` 文本加「下次可用时间：」前缀
- [x] `App.vue`：面板与弹窗的「下次重置时间」→「下次可用时间」
- [x] `App.vue`：删除已无引用的 `projectNames()`
- [x] `App.css`：`.agent-row` 基栅格改 4 列、`max-width:1100px` 断点改 3 列
- [x] `App.css`：删除 `.agent-projects`
- [x] `App.css`：删除 `.nav-sub{padding-left:22px}`（留注释说明类名为何保留）
- [x] `App.css`：新增 `.agent-details input[type=datetime-local]` 换行与间距

## T3 构建与验证

- [x] `npm --prefix frontend run build`（含 `vue-tsc`）
- [x] `gofmt -l .` / `go vet ./...` / `go test ./...`
- [x] `wails build`
- [x] 无头行为断言（mock 绑定 + Chrome 截图）：
  - [x] 行内无顿号、无「未关联项目」、子元素 4 个
  - [x] 日期列以「下次可用时间：」开头；不可用时带日期、可用时为 `—`
  - [x] 面板标签为「下次可用时间」、全文无「下次重置时间」
  - [x] 日期框在标签下方且间距 ≥ 6px、宽度撑满内容区
  - [x] 导航三项文字左边缘对齐（差值 ≤ 1px）

## T4 文档与记忆

- [x] `outline.md` 对应段落
- [x] `.workbuddy/memory/2026-09-22.md`（追加）与 `MEMORY.md`（如果有新约定）
