# 008 任务清单

- [x] T1 编写 spec / design / tasks 三件套
- [x] T2 `App.vue`：返回按钮文案 `← 返回` → `返回`
- [x] T3 `App.vue`：标题行加 `prompt-title-text` 包裹层，`agent-sub` 移入并排在 `h1` 之后
- [x] T4 `App.vue`：卡片拆成 `prompt-body`（左列 `Prompt N`+【复制】、右侧 textarea）+ `prompt-foot`（【删除】）
- [x] T5 `App.css`：新增 `.prompt-title-text` / `.prompt-title-text .agent-sub` / `.prompt-body` / `.prompt-foot`，`.prompt-card` 改纵向、`align-items:stretch` 下移到 `.prompt-body`
- [x] T6 `vue-tsc` 构建 + `gofmt` / `go vet` / `go test ./...` 无回归
- [x] T7 无头界面验证（mock 预览 + 行为断言 + 截图）：返回无箭头、agent 名同行且居右、删除在卡片底部且是最后一个子元素、复制仍在左列
- [x] T8 `wails build` 打包 `build/bin/Task Manager.exe`
- [x] T9 同步 `outline.md` 与 `.workbuddy/memory`
