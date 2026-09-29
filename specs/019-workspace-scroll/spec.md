# 019 主内容区滚动（产出概览 / Agent 应用）

> **⚠️ 已被 [020](../020-list-scroll/spec.md) 修订（2026-09-29）**
>
> 019 的定位没错（`.workspace` 第 7 行的 `overflow:hidden` 覆盖了第 1 行的 `overflow:auto`），
> 但修法过头了：它把滚动还给**整个主内容区**，于是矮窗口下滚轮落在最外层，
> **连页头（标题 + 右上角主按钮）一起被滑走**——正是用户在 020 里要求改掉的。
>
> 020 改为「页头固定 + 内容块自滚」：`.workspace` 重新回到 `overflow:hidden`，
> 滚轮只挂在各视图内部的内容块上（产出概览 `.output-body`、Agent `.agents-panel`、单位 `.units-panel`）。
> 019 的第 3–5 条（子项不参与 flex 收缩、任务列表/日历改 flex 高度、侧边栏可滚）**保持不变**，
> 仍是有效结论。下文为 019 当时的原文，保留存档。

## 背景

把窗口缩小（尤其是变**矮**）后，【产出概览】与【Agent应用】的界面显示不全，而且没有滚动条，
无法滑动查看剩余内容。

## 问题定位

`.workspace` 有两条同特异性（0,1,0）的规则：

| 位置 | 规则 |
| --- | --- |
| `App.css` 第 1 行 | `.workspace{padding:32px 34px;overflow:auto}` |
| `App.css` 第 7 行 | `.workspace{overflow:hidden;display:flex;flex-direction:column}` |

第 7 行在文件靠后，把整区滚动**覆盖成了 `hidden`**。它当初是为 Prompt 视图写的
（Prompt 的滚动要发生在视图内部的 `.prompt-board` 上，外层必须不能滚），
但副作用是所有视图一起失去了整区滚动：

- 【产出概览】【Agent应用】【单位管理】的内容超出可视高度后**直接被裁掉**，没有滚动条；
- 【任务列表】【日历】靠 `max-height:calc(100vh - 190px)` / `height:calc(100vh - 185px)`
  自己算高度，还能内部滚动，所以这两个视图一直是好的 —— 也正因如此，问题被掩盖了很久。

## 做什么

1. **主内容区恢复纵向滚动**：删掉第 7 行的 `overflow:hidden`，让第 1 行的 `overflow:auto` 生效。
2. **只有 Prompt 视图关掉整区滚动**：`.workspace.workspace-full{overflow:hidden}`。
3. **主内容区子项不参与 flex 收缩**：`.workspace>*{flex:0 0 auto}`，
   避免空间不足时子项被压扁（压扁不产生滚动条，只会让卡片/图表挤出容器，比滚动更难排查）。
4. **任务列表 / 日历改用 flex 自适应剩余高度**，替换 `calc(100vh - N px)` 的硬编码估算
   （那个估算比真实可用高度小 27px / 21px，会给这两个视图凭空加一条只能滚一点点缝隙的滚动条）。
5. **侧边栏自身可滚**：矮窗口下菜单 + 底部【设置】【关于】会超出 100vh，原先被 `.shell` 的
   `overflow:hidden` 裁掉，按钮点不到。

## 不做什么

- **不动宽度方向的响应式断点**。`@media(max-width:1000px)`（收窄侧栏、隐藏右侧详情面板）与
  `@media(max-width:1100px)`（Agent 行隐藏日期列）都保持原样。
- **不引入自定义滚动条样式**，用系统默认。
- **不改任何 Go 代码、绑定签名、数据结构、JSON 字段**。本轮是纯 CSS。
- 不给其它视图加"回到顶部"之类的附加控件。

## 验收标准

在**矮窗口**（1440x620 → 实际内宽 1418x522）下：

1. 【产出概览】主内容区 `overflow-y=auto`，`scrollHeight=1012 > clientHeight=522`，
   滚轮能滑到末尾（`scrollTop=490`，上限 490），滑到底后末尾模块（饼图面板）完整可见。
2. 【Agent应用】主内容区 `scrollHeight=1144 > 522`，能滑到末尾（`scrollTop=622`，上限 622）。
3. 【任务列表】与【日历】**不出现多余的外层滚动条**（`workspace scrollHeight == clientHeight`），
   滚动只发生在列表 / 日历格子区内部。
4. Prompt 视图行为不变：外层 `overflow-y=hidden`，`.prompt-board` 内部 `overflow-y=auto`。
5. **内容不足时不撑满**：1440x802 下任务列表 `scrollTop` 上限为 0（白卡片高度贴合内容，
   不会被拉满到屏幕底部）。
6. **对照组必须失败**：用同一套探针、把 `overflow` 注入回修复前的 `hidden`，
   「主内容区允许纵向滚动」与「滚轮能滑到内容末尾」两条必须 FAIL ——
   否则断言等于自证，证明不了它真能捕捉这个缺陷。

## 影响面

纯渲染层，无数据影响。所有视图共享同一套滚动规则，`specs/017` 起的版本号链路、
`specs/018` 的自启逻辑均不受影响。
