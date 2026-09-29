# 018 任务清单

- [x] 排查：确认 `settings.autoStart == true` 但注册表无该项（并排除 32 位视图 / RunOnce / StartupFolder / HKLM）
- [x] 新增 `autostart_windows.go`：注册表 API 直写 + 写后回读 + 按值名参数化
- [x] 新增 `autostart_other.go`：非 Windows 平台的同名桩
- [x] `app.go`：删除基于 `reg.exe` 的 `setAutoStart`；`startup` 中新增 `restoreAutoStart()` 自愈调用；移除 `os/exec` 导入
- [x] 新增 `autostart_windows_test.go`：7 项测试，全部真实读写注册表，使用独立值名 `TaskManagerAutostartProbe`
- [x] `go mod tidy`（`golang.org/x/sys` 由间接依赖提为直接依赖）
- [x] `gofmt -l` / `go vet` / `go test ./...` 全过（41 项）
- [x] `wails build -o TaskManager-verify.exe` 整体打包验证
- [x] 端到端验证自愈：清空注册表项 → 启动新程序 → 项被自动补回（PASS）
- [x] 修复用户当前状态：手动写入注册表项，指向 `build\bin\Task Manager.exe`，读回校验一致
- [x] 规格：`specs/018-autostart-reliability/`（本轮无界面改动，无截图）
- [x] `outline.md` 同步
- [x] 项目记忆（当日日志 + `MEMORY.md`）

## 验证证据

| 项目 | 结果 |
| --- | --- |
| Go 测试 | 41 项通过（含本轮新增 7 项） |
| `go vet` / `gofmt` | 通过 |
| 打包 | `wails build -o TaskManager-verify.exe` 27.6s，成功 |
| 注册表往返（测试） | 写入 → 读回一致 → 删除 → 读回为空 |
| 陈旧路径修复（测试） | 指向 `C:\somewhere else\...` 的项被改写成当前 exe 路径 |
| 启动自愈（端到端） | 项被清空后启动程序，6 秒内被补回为当前 exe 路径 → PASS |
| 用户当前注册表项 | `"D:\projects\TaskManager\build\bin\Task Manager.exe"`（与运行中的实例一致） |

## 未验证 / 留给用户确认

- **真正的「开机自启」需要重启一次才能看到**。本轮只能验证到「注册表项存在且格式正确」，
  这一层与 Windows 实际读取的完全一致，但最终的「重启后是否弹出来」由用户在下次开机时确认。
- 用户当前运行的是**修复前**的 exe（PID 33252）：注册表项已由本轮直接写好，所以**本次重启即可生效**；
  要拿到自愈与明确的错误提示，需关掉程序后重新 `wails build`。
