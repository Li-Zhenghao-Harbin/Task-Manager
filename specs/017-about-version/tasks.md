# 017 任务清单

- [x] `version.go`：`//go:embed wails.json` + `sync.OnceValue` 解析 + `GetAppVersion()`
- [x] `version_test.go`：返回值 == 磁盘 `wails.json`；格式 `x.x.x`
- [x] `wails generate module` 重建绑定（`App.d.ts` / `App.js` 出现 `GetAppVersion`）
- [x] `App.vue`：导入 `GetAppVersion`、新增 `appVersion` ref、`onMounted` 取值（带预览环境兜底）
- [x] `App.vue` 模板：`beta` → `{{ appVersion || '—' }}`
- [x] prettier 排版（单引号 / 无分号 / printWidth 120 / tabWidth 2）
- [x] `gofmt -l` / `go vet` / `go test ./...` 全过（34 项）
- [x] `npm --prefix frontend run build`（`vue-tsc --noEmit` + vite build）通过
- [x] `wails build -o TaskManager-verify.exe` 整体打包验证（不动正在运行的实例）
- [x] 无头界面验证：宽屏 `?pos=top` 断言 4/4；`?pos=none` 干净图
- [x] 对照组：把注入值改成 `9.9.9` 后界面随之变化（证明非硬编码）
- [x] 规格：`specs/017-about-version/` 三件套
- [x] `outline.md` 同步
- [x] 项目记忆（当日日志 + `MEMORY.md`）

## 验证证据

| 项目 | 结果 |
| --- | --- |
| Go 测试 | 34 项通过（含本轮 2 项） |
| 前端构建 | `vue-tsc` 无类型错误，vite 产物 113.25 KiB |
| 界面断言（宽屏 1440×900） | 4/4 PASS |
| 弹窗显示 | `1.0.0`（与 `wails.json` 一致） |
| 对照组（注入 9.9.9） | 弹窗显示 `9.9.9` |
| `beta` 残留 | 界面中已不存在 |
