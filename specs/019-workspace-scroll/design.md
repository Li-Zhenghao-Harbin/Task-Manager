# 019 设计说明

## 1. 为什么是「恢复第 1 行的 overflow:auto」而不是「给每个视图加滚动」

最省事的写法是给 `.output-dashboard` / `.agents-workspace` / `.units-workspace` 各加一份
`flex:1;min-height:0;overflow:auto`。没有采用，原因：

- 三处重复同一组规则，将来新增第四个视图还要再抄一次，漏掉就是同样的 bug 复现；
- 主内容区本来就是滚动容器（第 1 行的 `overflow:auto`），第 7 行的 `overflow:hidden` 是**例外**，
  正确的写法是把这个例外收窄到 Prompt 视图，而不是给其它视图各发一份豁免。

因此改成「默认整区滚动 + Prompt 视图单独关闭」，规则总数反而减少。

## 2. Prompt 视图为什么要显式关掉（不能只靠"它没有溢出"）

`.prompt-workspace{flex:1;min-height:0}` + `.prompt-board{overflow:auto}` 的设计是
**滚动发生在看板内部**：标题行固定在顶部，只有 prompt 卡片区滚。

如果外层 `.workspace` 也是 `auto`，在矮窗口下会出现**内外两层滚动条**：
看板已经被 `flex:1;min-height:0` 压到可用高度，外层本不该滚，
但 `min-height:0` 只影响 flex 分配，外层容器的 `scrollHeight` 仍可能因为
header 的 margin 等零头略大于 `clientHeight`，于是挂出一条 1~2px 的假滚动条。

用 `.workspace.workspace-full{overflow:hidden}` 明确表达意图，比依赖数值巧合可靠。
选择器写成 `.workspace.workspace-full`（0,2,0）而不是 `.workspace-full`（0,1,0）：
文件里有更晚出现的 `.workspace>*{flex:0 0 auto}` 等规则，提高特异性可以避免顺序依赖。

## 3. `.workspace>*{flex:0 0 auto}` 的必要性

`.workspace` 是 `display:flex;flex-direction:column`。flex 子项默认 `flex-shrink:1`，
空间不足时会被压缩。

实测这个压缩**不会产生滚动条**：子项高度从 940 压到 522，但它内部的卡片并不跟着缩，
只是溢出到容器外，被滚动容器的可视区裁掉 —— 而 `scrollHeight` 反映的是溢出量，
所以滚动条照样出现，看起来"没坏"，但容器自己有 `background`/`border` 的元素
（比如 `.task-list` 的白色卡片）只画到被压扁的高度，底部内容会跑到卡片外面。

显式 `flex:0 0 auto` 让子项高度 = 内容高度，语义清晰。

> 注意：`.task-list` 与 `.calendar-view` 需要的是**能收缩**，它们在自己的规则里
> 覆盖成 `flex:0 1 auto`（见下）。两条规则同特异性（0,1,0），靠文件顺序决定：
> `.workspace>*` 在第 7 行，`.task-list` / `.calendar-view` 在其后。

## 4. 任务列表 / 日历：去掉 `calc(100vh - N px)`

原实现：

```css
.task-list{max-height:calc(100vh - 190px)}
.calendar-view{height:calc(100vh - 185px)}
```

`190` / `185` 是「侧栏 padding + header + toolbar」高度的手工估算。
实测真实开销是 **217px**，于是：

- 任务列表：内容比可用空间少 27px → 主内容区挂出一条**只能滚 27px** 的滚动条，
  滚一下列表底部就被推出可视区；
- 日历：`100vh - 185` 比可用高度**多 21px** → 日历底部溢出，同样触发外层滚动条。

这两个数字还会随字体、DPI、`header` 内容（比如标题换行）变化，属于典型的
"看起来对、换台机器就错"的估算。

改成让 flex 算：

```css
.task-list{flex:0 1 auto;min-height:0;overflow-y:auto;overscroll-behavior:contain}
.calendar-view{...;flex:0 1 auto;min-height:0}
```

- `flex-basis:auto` → 基础高度取内容高度；
- `flex-grow:0` → **内容不足时不撑满**（白卡片贴合内容，不会拉到屏幕底部）；
- `flex-shrink:1` → 空间不足时才压缩到可用高度，溢出的部分由自身 `overflow-y` 接管；
- `min-height:0` → **必需**。flex 子项默认 `min-height:auto`（= content-based minimum），
  会拒绝被压到内容高度以下，不加这一条 `flex-shrink` 完全不生效。

这次也顺便修掉了「日历底部溢出 21px」这个此前未被发现的同类问题。

## 5. 侧边栏

`.sidebar{overflow-y:auto}`：矮窗口下「菜单 + 底部统计 + 设置/关于按钮」会超出 `100vh`，
在没有滚动条的情况下被 `.shell` 的 `overflow:hidden` 裁掉，底部按钮点不到。
`.stats{margin-top:auto}` 在滚动容器里仍然贴底（容器高度仍是 100vh），行为不变。

用文件末尾追加一条独立规则、而不是改第 1 行那条超长规则，是为了保持
`App.css`「一条规则一行、按主题追加」的既有写法。

## 6. 改动的规则清单（都在 `App.css`）

| 规则 | 变更 |
| --- | --- |
| `.workspace`（第 7 行） | 删 `overflow:hidden`，只留 `display:flex;flex-direction:column` |
| `.workspace>*`（第 7 行） | 新增 `flex:0 0 auto` |
| `.workspace.workspace-full` | `overflow:hidden`（原为 `.workspace-full`，仅 grid-column） |
| `.task-list` | `max-height:calc(100vh - 190px)` → `flex:0 1 auto;min-height:0` |
| `.calendar-view` | `height:calc(100vh - 185px)` → `flex:0 1 auto;min-height:0` |
| `.sidebar` | 新增 `overflow-y:auto` |

`App.css` 不参与 prettier 格式化（刻意保持"一条规则一行"的紧凑写法），本轮改动均在同一行内完成。
