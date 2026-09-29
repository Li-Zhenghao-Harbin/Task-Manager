# 018 设计：把启动项写对，并且写不进去时要知道

## 为什么会「设置里是开、开机却不启动」

原实现只有 8 行：

```go
func setAutoStart(enabled bool) error {
	const key = `HKCU\Software\Microsoft\Windows\CurrentVersion\Run`
	if enabled {
		exePath, err := os.Executable()
		if err != nil { return err }
		return exec.Command("reg", "add", key, "/v", "Task Manager", "/t", "REG_SZ",
			"/d", `"`+exePath+`"`, "/f").Run()
	}
	cmd := exec.Command("reg", "delete", key, "/v", "Task Manager", "/f")
	if err := cmd.Run(); err != nil { return nil }
	return nil
}
```

它把「写注册表」这件事整个外包给了 `reg.exe` 子进程，于是继承了三个弱点：

| 弱点 | 后果 |
| --- | --- |
| `Run()` 丢弃 stdout/stderr | 失败只剩一句 `exit status 1`，用户与开发者都无从判断原因 |
| 写入后不回读 | 调用被放行但没落盘时（安全软件拦启动项很常见），程序以为成功 |
| GUI 进程启动控制台程序 | 每次设置都闪一个黑框（012 已在通知上修过同类问题） |

第三个弱点在本机有直接旁证：这台机器的执行策略**禁止 `reg.exe` 启动**（连 Go 测试里 `fork/exec` 都被拒）。
在这种环境下，旧实现**必然静默失败或报一句无用错误**；而新实现走注册表 API，实测 7 项测试全部真实读写成功。
换句话说，这个改动把「能不能写」从「取决于外挂程序是否被放行」变成了「取决于当前用户对 HKCU 的权限」。

## 实现

### 1. 直写注册表（`autostart_windows.go`）

```go
key, _, err := registry.CreateKey(registry.CURRENT_USER, autoStartKeyPath,
	registry.SET_VALUE|registry.QUERY_VALUE)
```

用 `golang.org/x/sys/windows/registry`（Wails 已间接依赖，本次提为直接依赖），不再起任何子进程。

- 无控制台窗口闪烁；
- 错误具体（拒绝访问 / 键不存在）；
- 不依赖 PATH，也不受「拦截 reg.exe」这类策略影响。

### 2. 写后回读（关键）

```go
if err = key.SetStringValue(valueName, want); err != nil { return ... }
got, _, err := key.GetStringValue(valueName)   // ← 回读
if got != want {
	return fmt.Errorf("开机启动项未生效（写入 %s，读回 %s），可能被安全软件拦截", want, got)
}
```

这是本次最要紧的一处。**「静默成功」比「明确失败」危险得多** —— 前者会让用户带着一个假象去等一次永远不会发生的开机启动。
回读把这种状态变成一条能看见的错误。

### 3. 读取与删除的边界

- 读一个不存在的值返回 `""` 而不是错误，调用方才分得清「没设过」与「读失败」；
- 删除时 `ErrNotExist` 视为成功（开关从「开」拨到「关」时项可能本就不在）。

### 4. 启动自愈（`restoreAutoStart`）

```
startup() → load() → restoreAutoStart() → syncAutoStart(settings.AutoStart)
                                            │
                                   值已等于当前 exe 路径？ ── 是 ──> 什么都不做
                                            │ 否
                                            └──> 写入
```

覆盖两种现实情况：

- **项缺失**：早先写入被拦、或被清理工具删掉 —— 用户机器上正是这个；
- **项陈旧**：程序是便携的（数据文件放在 exe 旁边），用户移动目录后注册表仍指向旧路径。

只在值不一致时才写，正常启动不产生注册表写入。失败不阻塞启动（`_ =`），
由设置弹窗在用户下次保存开关时给出明确错误。

### 5. 值名参数化以便测试

`writeAutoStartValue` / `syncAutoStartValue` 接收值名，`setAutoStart` / `syncAutoStart` 用常量
`"Task Manager"` 包一层。测试用 `TaskManagerAutostartProbe` 这个独立值名，
**绝不触碰用户真实依赖的那一项** —— 单元测试不该决定用户的程序是否跟着 Windows 启动。

## 保留的语义

`UpdateSettings` 仍是「自启写入失败 → 整体返回错误、不落盘」。这样「设置里的值 == 注册表的实际状态」
这个不变量成立：失败时设置不会停留在「开」而实际没开。前端的 `saveSettings` 已经有 `catch`
把错误显示成提示，无需改动。

## 验证

- 7 项新测试全部**真实读写注册表**（写入→读回→删除、幂等、陈旧路径修复、关闭时不动注册表）；
- 全量 41 项通过；
- 代码路径不再含任何子进程，`go vet` 通过；
- 端到端：用注册表脚本直接写入用户真实项（指向 `build\bin\Task Manager.exe`）并读回校验一致，
  用户重启即可验证自启是否生效。
