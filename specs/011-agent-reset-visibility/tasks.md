# 011 · 任务分解

## T1 规格

- [x] `spec.md` / `design.md` / `tasks.md`

## T2 前端实现

- [x] `App.vue`：`.agent-reset` 加 `v-if="!agent.available"`（可用时整格不渲染）
- [x] `App.vue`：行 class 绑定加 `available: agent.available`
- [x] `App.css`：新增 `.agent-row.available{grid-template-columns:16px minmax(0,1fr) 78px}`
- [x] `App.css`：更新 `.agent-row` 上方注释（说明日期列按可用性进出 DOM）
- [x] `App.css`：`max-width:1100px` 断点里把 `.agent-row.available` 并列写进 3 列规则
      （媒体查询不增加特异性，不写会被更高特异性的 `.agent-row.available` 盖掉）

## T3 构建与验证

- [x] `npm --prefix frontend run build`（含 `vue-tsc`）
- [x] `gofmt -l .` / `go vet ./...` / `go test ./...`（29 passed）
- [x] `wails build`
- [x] 无头行为断言（mock 绑定 + Chrome 截图，宽屏 14/14、窄屏 13/13）：
  - [x] 可用行无「下次可用时间」字样、无 `—`、无 `.agent-reset` 元素
  - [x] 不可用行文案为「下次可用时间：2026-09-22 15:00」/「下次可用时间：未设置」
  - [x] 页面 `.agent-reset` 元素数 = 不可用 Agent 数（2）
  - [x] 可用行可见子元素 3 / 栅格 3 轨道；不可用行 4 / 4 轨道
  - [x] 混排三行标题左缘、状态标签右缘对齐（差值 0.00px）
  - [x] 对照组：可用行标题列 606px > 不可用行 417px（宽度已回收）
  - [x] 面板对照：选中可用 Agent 时不出现该字段、无日期输入框
  - [x] 窄屏 1000px：两种行都 3 轨道、日期列隐藏、对齐仍成立

## T4 文档与记忆

- [x] `outline.md`：Agent 列表行描述去掉「可用时是 `—`」
- [x] `specs/010-.../spec.md`：标注其验收标准第 2 条已被 011 修订
- [x] `.workbuddy/memory/2026-09-22.md`（追加）与 `MEMORY.md`（补栅格随可用性变化的约定）
