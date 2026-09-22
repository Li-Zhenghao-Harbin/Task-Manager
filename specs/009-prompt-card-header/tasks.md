# 009 · 任务分解

## T1 规格

- [x] `spec.md` / `design.md` / `tasks.md`

## T2 前端实现

- [x] `App.vue`：卡片模板换成 `prompt-head`（`prompt-label` + `prompt-copy` + `prompt-delete`）+ 同级 `textarea`
- [x] `App.css`：
  - [x] 新增 `.prompt-head`
  - [x] `.prompt-delete` 追加 `margin-left:auto`
  - [x] `.prompt-card textarea` 去掉 `flex:1`（改为 `flex:0 0 auto`）
  - [x] 删除 `.prompt-body` / `.prompt-side` / `.prompt-foot`

## T3 构建与验证

- [x] `npm --prefix frontend run build`（含 `vue-tsc`）
- [x] `gofmt -l .` / `go vet ./...` / `go test ./...`
- [x] `wails build`
- [x] 无头行为断言（mock 绑定 + Chrome 截图）：
  - [x] label 与复制左边缘对齐、垂直中心齐平
  - [x] 删除与它们同一行、右边缘贴内容区右边
  - [x] `删除.left - 复制.right` > 0（有留白）
  - [x] textarea 在头部下方且横向撑满
  - [x] 复制行为：提示 + 剪贴板内容 + 不触发保存
  - [x] 删除行为：条数 -1 + 保存一次

## T4 文档与记忆

- [x] `outline.md` 3.8 节的卡片描述
- [x] `.workbuddy/memory/2026-09-22.md`（追加）与 `MEMORY.md`（卡片约定）
- [x] 技能 `wails-vue-ui-verify` 如有新坑则一并补充
