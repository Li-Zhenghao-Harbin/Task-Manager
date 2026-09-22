# 012 · 启动即最大化、通知不闪控制台、单位筛选与前端代码换行

## 背景

四条互不相干的体验问题，都是「用起来别扭」而不是「功能不对」：

1. **启动时会看到窗口放大**。`main.go` 原来在 `OnDomReady` 里延迟 250ms 调 `runtime.WindowMaximise`：
   此时窗口已经以 1280×800 显示出来了，于是用户先看到小窗、再看到它长成最大化——一段多余的动画。
   Wails 本身支持「第一次显示就是最大化」，不需要事后补救。
2. **Agent 恢复可用时闪一个 cmd/控制台窗口**。桌面通知是靠起一个 `powershell.exe` 调 WinRT Toast 实现的，
   PowerShell 是控制台程序，从 GUI 进程启动它时 Windows 会为它分配一个控制台——那一下就是闪出来的黑框。
   `-WindowStyle Hidden` 只隐藏 PowerShell 自己的宿主窗口，挡不住这次控制台分配。
3. **搜索框会命中单位名**。在「全部任务」里搜「研发」会同时命中单位叫「研发」的任务，
   而这个筛选意图本该用下拉框表达；关键词与筛选混在一起后，结果不好预期也不好解释。
4. **前端代码被挤成超长单行**。`App.vue` 的模板与脚本有大量两千字符以上的行（最长近 3000 字符），
   改一处要在编辑器里横向拖动，review 也看不出改了什么。

## 做什么

### 1. 启动即最大化

- `main.go` 用 `options.App.WindowStartState = options.Maximised` 声明起始状态；
  Wails 会在窗口**首次显示时**直接以最大化状态显示它，不存在可见的放大过程。
- 删掉 `OnDomReady` 里的 `time.Sleep(250ms)` + `runtime.WindowMaximise`。
- `Width/Height`（1280×800）保留，作为取消最大化后的还原尺寸；`MinWidth/MinHeight` 不变。
- 窗口配置从 `main()` 的内联字面量提取为 `appOptions(app *App) *options.App`，便于用测试钉住。

### 2. 通知不再闪控制台窗口

- 新增平台相关的 `hiddenProcessAttr()`：Windows 下返回
  `&syscall.SysProcAttr{HideWindow: true, CreationFlags: CREATE_NO_WINDOW}`，
  其它平台返回 `nil`（保持 `notify.go` 不写 build tag）。
- `showWindowsToast` 把它设到 `command.SysProcAttr` 上。
  `CREATE_NO_WINDOW` 让 Windows**根本不为子进程分配控制台**，所以连闪一下都不会有。
- 通知本身的机制（PowerShell + WinRT Toast + `-EncodedCommand` 传中文）不变，通知样子也不变。

### 3. 单位改为下拉筛选

- 工具栏新增「单位」下拉框，放在关键词输入框与「全部层级」之间，样式与旁边两个下拉完全一致：
  首项 `全部单位`，其余项按 `units` 列表渲染（value 用**单位名**，因为任务里存的就是名字）。
- `filteredTasks` 的搜索关键词**不再包含单位**：只匹配 `任务名` 与 `下发人`；
  单位作为独立条件参与筛选（`unitFilter === 'all' || task.unit === unitFilter`）。
- 搜索框 placeholder 改为「搜索任务名或下发人」。
- 三个下拉筛选（单位 / 层级 / 类型）可叠加，并且与状态（左侧导航）筛选叠加，
  日历视图沿用同一份 `filteredTasks`，自动跟着筛。
- **兜底**：`refresh()` 里若当前选中的单位已不存在（被删除/改名），筛选值回到 `全部单位`，
  避免下拉框显示空白且列表被一个不存在的单位名筛成空。

### 4. 前端代码换行

- 用 prettier 按项目既有风格重排 `frontend/src/App.vue`（单引号、无分号、printWidth 120）：
  模板属性逐个换行，脚本里的长行按 120 列折行。**只重排格式，不改逻辑。**
- 顺带把 4 处「多语句内联事件处理器」改成具名函数（`openStatus` / `openOutput` / `openTask` / `applyColor`）：
  `@click="a; b"` 这种写法在被格式化成换行后会让 Vue 的表达式解析器直接报错
  （`Error parsing JavaScript expression: Unexpected token, expected ","`），
  写成函数既修掉这个隐患，也让模板更短。
- `App.css` **不参与格式化**：它刻意保持「一条规则一行」的紧凑写法，与本轮的诉求无关。

## 不做什么

- 不改 `Width/Height/MinWidth/MinHeight` 的数值，不改主题色与背景色。
- 不改通知的触发时机与文案（仍然是 30 秒轮询 + 启动补偿检查，文案含「应用名-模型名」）。
- 不引入通知库，不把 Toast 换成 Wails 内建通知 API（机制替换风险大，且与本次诉求无关）。
- 不给搜索框加「搜单位」开关：单位筛选已经覆盖这个用法。
- 不改单位的数据结构，不给 `Task` 加 `unitId`（单位名仍是唯一标识，改名由单位管理负责）。
- 不动 `App.css` / `main.ts` / `style.css` 的排版风格。

## 验收标准

1. 窗口第一次显示出来时就是最大化状态，全程看不到「小窗 → 放大」的过程；
   `appOptions` 的 `WindowStartState` 为 `options.Maximised` 且 `OnDomReady` 为空（单元测试钉住）。
2. 发送桌面通知的子进程带 `HideWindow` 与 `CREATE_NO_WINDOW`（单元测试钉住）；通知本身仍能弹出。
3. 「全部任务」工具栏有 **3 个下拉框**（单位/层级/类型），placeholder 不含「单位」。
4. 关键词等于某单位名时**不再命中**该单位的任务（搜「数据中台部」结果 0 条）。
5. 单位下拉筛「数据中台部」只剩该单位的任务；筛「平台研发部」剩 2 条；单位 + 层级筛选可叠加。
6. `App.vue` 重排后，各视图渲染出的 DOM 与重排前**完全一致**（逐行 diff 为空），
   且 011 的 Agent 列表断言全部继续通过。
7. `gofmt` / `go vet` / `go test ./...` / `vue-tsc` / `wails build` 全部通过。
