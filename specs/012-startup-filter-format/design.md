# 012 · 设计

## 1. 最大化：为什么必须用 `WindowStartState`

Wails（Windows 实现）创建主窗口时用的是 `WS_OVERLAPPEDWINDOW` **不带 `WS_VISIBLE`**：
窗口先以隐藏状态存在，随后按 `WindowStartState` 决定怎么第一次显示它。

```go
switch f.frontendOptions.WindowStartState {
case options.Maximised:  win32.ShowWindowMaximised(h)   // ShowWindow(SW_MAXIMIZE)
default:                 win32.ShowWindow(h)            // ShowWindow(SW_SHOW)
}
```

于是：

| 做法 | 结果 |
| --- | --- |
| `OnDomReady` → `WindowMaximise`（旧） | 窗口已经 `SW_SHOW` 出来了，再最大化 → 用户看见它长大 |
| `WindowStartState: options.Maximised`（新） | 窗口从未以普通尺寸显示过，第一帧就是最大化 |

`OnDomReady` 触发点在导航完成回调里，**先于**这次显示；旧代码还要再等 250ms，
所以那段时间窗口一直是普通尺寸。删掉它没有任何副作用——`OnStartup`（`app.startup`）仍在，负责数据加载与轮询。

`WindowStartState` 只影响首次显示；用户手动还原后回到 `Width × Height`。

```
appOptions(app)
├─ Title/Width/Height/MinWidth/MinHeight   窗口标题与尺寸（还原尺寸）
├─ WindowStartState: options.Maximised      ← 本次改动
├─ AssetServer / BackgroundColour
├─ OnStartup: app.startup                  （OnDomReady 已删除）
└─ Bind: []interface{}{app}
```

## 2. 通知：控制台为什么闪

```
GUI 进程 (Task Manager.exe)
   └─ exec.Command("powershell.exe", ...)     ← 控制台程序
         Windows 为它分配一个 console        ← 闪的就是这一下
         -WindowStyle Hidden 只影响 PowerShell 自己创建的宿主窗口，挡不住分配
```

`SysProcAttr{CreationFlags: CREATE_NO_WINDOW}`（`0x08000000`）直接告诉
`CreateProcess` **不要创建控制台**；`HideWindow: true` 再把可能出现的窗口也压住。
两者都设上是最稳的组合，代价为零。

平台相关的拆分是为了让 `notify.go` 保持无 build tag：

| 文件 | 内容 |
| --- | --- |
| `proc_windows.go` | `hiddenProcessAttr()` 返回 `{HideWindow: true, CreationFlags: createNoWindow}` |
| `proc_other.go` | `hiddenProcessAttr()` 返回 `nil`（非 Windows 无控制台可抑制） |

## 3. 单位筛选

筛选条件全部集中在 `filteredTasks` 里，新增一个与既有条件同构的分支：

```ts
const filteredTasks = computed(() =>
  tasks.value.filter(
    (task) =>
      (status.value === 'all' || task.status === status.value) &&
      (kindFilter.value === 'all' || task.kind === kindFilter.value) &&
      (typeFilter.value === 'all' || task.taskType === typeFilter.value) &&
      (unitFilter.value === 'all' || task.unit === unitFilter.value) &&   // ← 新增
      `${task.name} ${task.issuer}`.toLowerCase().includes(query.value.toLowerCase()),  // ← 去掉 task.unit
  ),
)
```

- **按名字匹配而不是 id**：`Task.Unit` 存的就是单位名（新建任务时 `select` 的 value 是 `unit.name`），
  没有 `unitId` 这个字段，也不需要为此改数据模型。
- 下拉框放在搜索框右侧、层级左侧，复用 `.toolbar select` 现有样式，不新增 CSS。
- 日历视图读同一个 `filteredTasks`，所以筛选自动生效。
- `refresh()` 里的兜底把「选中的单位被删掉/改名」收敛回 `全部单位`，
  否则 `<select>` 会显示空白但筛选仍然按旧名字算，列表变空且看不出原因。

## 4. 代码换行

用 prettier 3 按项目既有风格重排（不新增依赖到 `package.json`，直接调用已在那里的可执行文件）：

```bash
cd frontend && node node_modules/prettier/bin/prettier.cjs --write src/App.vue \
  --single-quote --no-semi --print-width 120 --tab-width 2
```

`print-width 120` 是折中：脚本里的链式调用大多能保持一行，模板的属性则逐个换行。
**不要开 `--html-whitespace-sensitivity ignore`**：本项目的模板靠 `</button><button` 这种紧邻写法
避免 inline-block 之间的空隙，`ignore` 会动这类空白。

### 必须一起改掉的模板写法

prettier 会把 `@click="a; b"` 拆成：

```html
@click="
  a
  b
"
```

Vue 的表达式解析器只接受**单条表达式**，换行后 `a` 与 `b` 两条语句直接报
`Error parsing JavaScript expression: Unexpected token, expected ","`（`vite build` 失败）。
本次把 4 处都改成具名函数：

| 原来 | 现在 |
| --- | --- |
| `section = 'tasks'; status = item.id` | `openStatus(item.id)` |
| `section = 'output'; outputFilter = item.id` | `openOutput(item.id)` |
| `selected = task; section = 'tasks'` | `openTask(task)` |
| `settings.backgroundColor = color; applyBackground()` | `applyColor(color)` |

> `vue-tsc --noEmit` **不会**报这个错（它不按 vite 的模板编译路径解析内联表达式），
> 只有 `vite build` 会。所以「类型检查过了」不能代表模板合法。

### 怎么证明「只是换了行」

重排前后各构建一次，用 Chrome 无头 `--dump-dom` 导出渲染后的 DOM，把资源哈希归一化后逐行 diff：

```bash
chrome --headless=new --dump-dom "http://127.0.0.1:8731/?view=tasks&pos=none" > before.html
# 重新格式化 + 重新构建后再导一次，然后
sed -e 's/index\.[a-z0-9]*\.js/index.HASH.js/' before.html | tr '>' '>\n' > b.txt
diff b.txt a.txt
```

本次 `tasks` 与 `agent 列表` 两个视图的 diff 都是**空**——重排没有改变任何一处渲染结果
（白空格、文本节点、属性、内联样式全部一致）。
