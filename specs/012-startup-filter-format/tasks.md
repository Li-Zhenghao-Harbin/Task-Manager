# 012 · 任务分解

## T1 规格

- [x] `spec.md` / `design.md` / `tasks.md`

## T2 启动即最大化

- [x] `main.go`：窗口配置提取为 `appOptions(app)`
- [x] `main.go`：加 `WindowStartState: options.Maximised`，删 `OnDomReady` 里的延迟最大化与为此引入的 `context`/`time`/`runtime` 依赖
- [x] `startup_test.go`：钉住 `WindowStartState == options.Maximised`、`OnDomReady == nil`、还原尺寸仍是 1280×800

## T3 通知不闪控制台

- [x] `proc_windows.go`：`hiddenProcessAttr()` = `HideWindow + CREATE_NO_WINDOW`
- [x] `proc_other.go`：非 Windows 返回 `nil`
- [x] `notify.go`：`command.SysProcAttr = hiddenProcessAttr()`
- [x] `proc_windows_test.go`：钉住两个标志位

## T4 单位下拉筛选

- [x] `App.vue`：新增 `unitFilter` ref（按单位名匹配）
- [x] `App.vue`：`filteredTasks` 去掉 `task.unit`、加 `unitFilter` 条件
- [x] `App.vue`：工具栏加「全部单位 + units」下拉，placeholder 改「搜索任务名或下发人」
- [x] `App.vue`：`refresh()` 里单位消失时回到「全部单位」

## T5 代码换行

- [x] prettier 重排 `App.vue`（单引号 / 无分号 / 120 列；`App.css` 不动）
- [x] 4 处多语句内联处理器改为 `openStatus` / `openOutput` / `openTask` / `applyColor`
- [x] DOM diff 证明重排前后渲染结果一致（tasks + agent 列表，diff 为空）

## T6 构建与验证

- [x] `gofmt -l .` / `go vet ./...` / `go test ./...`
- [x] `npm --prefix frontend run build`（含 `vue-tsc`）
- [x] `wails build`
- [x] 无头界面断言：
  - [x] 筛选栏 9/9：3 个下拉、placeholder 不含单位、搜「数据中台部」0 条、单位筛选 1 条 / 2 条、与层级叠加
  - [x] 011 回归 16/16：Agent 列表可用行不渲染「下次可用时间」，栅格与对齐不变

## T7 文档与记忆

- [x] `outline.md`：搜索/筛选描述、窗口与通知说明、界面草图
- [x] `.workbuddy/memory/2026-09-22.md`（追加）与 `MEMORY.md`（窗口、通知、筛选、格式化约定）
- [x] `wails-vue-ui-verify` 技能：补「prettier 拆坏内联处理器」与「dump-dom 做重排回归」两条
