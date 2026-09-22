# 014 设计

## 改动面

只改文案，两处：

| 文件 | 位置 | 改动 |
| --- | --- | --- |
| `frontend/src/App.vue` | Agent 详情面板 `.agent-projects-editor > h3` | `关联项目` → `关联任务` |
| `frontend/src/App.vue` | Agent 列表行展开区 `.agent-task-list > .hint` | `…可在右侧「关联项目」中勾选。` → `…可在右侧「关联任务」中勾选。` |
| `frontend/src/App.css` | `.details.agent-details` 上方注释 | 顺带把注释里的「关联项目列表」写成「关联任务列表」，避免注释与界面两个叫法 |

定位方式：面板与弹窗各有一个 `.agent-projects-editor`，两处的 `<h3>` 文本相同、外层结构相同，
**唯一能区分的是 `v-for` 的 `:key`**（面板是 `selectedAgent.id + ':' + task.id`，弹窗是 `task.id`）。
用 h3 单行替换必然替换错，所以按「标题 + 副说明 + 列表首行 label」四行一起匹配。

## 为什么改的是文案而不是字段

界面上的东西本来就是任务：候选项是「任务类型 = 代码」的任务，勾选后写的是 `Task.AgentID`，
后端校验失败时也说「关联项目必须是**代码类型的任务**」。改成「关联任务」是让名字与事实对齐，
**与数据模型没有关系**——`projectIds` 这个字段名继续保留（它与「项目」这个概念的来历有关，
改名会牵动 `AgentApp` 结构体、`AgentAppInput`、绑定、存档兼容，收益为零）。

## 为什么类名不改

`.agent-projects-editor` 等类名只在 `App.vue` 与 `App.css` 内部出现，用户看不到。
改名要同时动两处、且必须保证 007 的 flex 规则与窄屏断点都跟着改（那些规则是 `0,2,0` 特异性的
`.details.agent-details` 后代选择器），出错风险远大于收益。**类名与界面文案解耦**是更稳的选择。

## 验证

沿用 `wails-vue-ui-verify` 那套：mock 绑定 + Chrome 无头 + 断言横幅。

本轮判据上的两个新认识（已写进技能）：

1. **块级元素的 rect 宽度量不出文字长短**。`h3` 是块级元素，`getBoundingClientRect().width`
   等于容器宽度（面板 297px / 弹窗 510px），拿它比「两个标题谁宽」是错的。
   要量文字行本身，得用 `document.createRange().selectNodeContents(el).getBoundingClientRect().width`。
   实测改名后与弹窗里那个 4 字标题的文字行宽度都是 **60.00px**，
   这才是「同为 4 个汉字、版式不受影响」的直接证据。
2. **窄屏要断言「面板按设计隐藏」而不是量它的几何**。`@media(max-width:1000px){.details{display:none}}`
   是 007 的既定行为，1000px 下量面板内部会得到一排 0.00px。探针里先判 `details` 的宽度，
   宽度为 0 就走「按设计隐藏」这条断言分支，跳过面板内部几何——否则每次窄屏都要人工解释三个假失败。

### 结果

| 场景 | 断言 |
| --- | --- |
| 1440×900 | **20/20** |
| 1000×700 | **5/5**（列表文案 + 面板按设计隐藏） |

截图见 `screenshots/`：`01-panel-banner`（宽屏断言证据）、`02-panel-clean`（干净版式，能看到标题已是【关联任务】）、
`03-narrow-1000`（窄屏断言证据）。
