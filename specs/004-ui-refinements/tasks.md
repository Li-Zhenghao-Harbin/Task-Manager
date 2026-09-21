# 004 实施清单

## T1 移除保存状态提示

- [x] `flushSaves()` 移除 `saveState` / `saveHint` 流转；校验未过改为 `message.value = issue`
- [x] 删除 `saveState`、`saveHint` ref 与 `saveStateText` computed
- [x] 任务详情面板标题去掉 `.save-state` 胶囊
- [x] Agent 详情面板标题去掉 `.save-state` 胶囊
- [x] `App.css` 删除 `.save-state` / `.saving` / `.saved` / `.invalid` 规则，注释改写

## T2 列表为子任务补齐上级任务

- [x] 新增 `taskGroups` computed（含 `context` 标记）与 `orphanChildren` computed
- [x] 列表模板改用 `taskGroups` / `orphanChildren` 渲染
- [x] 空状态判据改为 `!taskGroups.length && !orphanChildren.length`
- [x] 孤儿子任务文案改为「子任务（上级已删除）」
- [x] `App.css` 追加 `.task-row.context` 系列样式

## T3 Agent 快捷冷却按钮

- [x] 新增 `setAgentCooldown(target, hours)`，输出 `YYYY-MM-DDTHH:mm` 本地时间
- [x] Agent 详情面板「可用情况」下方加 `.cooldown-row`（按钮 + 说明），按钮放在 `<label>` 之外
- [x] `App.css` 追加 `.cooldown-row` / `.cooldown-button` 样式

## T4 验证

- [x] `npm run build`（`vue-tsc --noEmit` + `vite build`）通过
- [x] mock 预览页 + Chrome 无头截图四个视图，断言横幅全部绿色：
  - `?view=list`：上级 `task-row completed context` 紧邻子任务上方；`orphan-child` 正确；`save-state` 元素 = 0
  - `?view=task`：无 `save-state`、无「保存」按钮
  - `?view=agent`：按钮文案 `5小时后可用`；无 `save-state`、无「保存」按钮
  - `?view=cooldown`：`available=false`、`resetAt` 分钟精度、距点击 300 分钟、徽标转「不可用」、重置时间字段回填一致
- [x] `gofmt -l .` 干净、`go test ./...` 通过（19 项）
- [x] `wails build` 产出 `build/bin/Task Manager.exe`
- [x] 截图归档到 `specs/004-ui-refinements/screenshots/`

## 备注

- 本次未改任何 Go 代码，19 项后端测试跑的是既有用例，用于确认没有连带影响。
- 未落 SDD 编号的猜测项：无。三处需求都已在 `spec.md` 的「不做什么」里划清了边界。
