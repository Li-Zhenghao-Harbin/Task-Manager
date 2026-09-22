# 015 · 去掉「项目」这个概念：所有文字与命名统一为「任务」

> 修订 014 里「不动数据字段与类名」的判断。014 只改了右侧面板的标题，本轮把剩下的「项目」全部清掉。

## 做什么

1. **界面文案**：【新增 Agent 应用】弹窗里的分区标题 `关联项目` → `关联任务`
   （014 只改了右侧面板，弹窗那个实例当时按 007 的口径留在原样）。
2. **数据字段改名**：`AgentApp.TaskIDs`（JSON 键 `taskIds`），旧键 `projectIds` 不再写出。
   - `AgentAppInput.TaskIDs` 同步改名，前端在线负载的键名随之变为 `taskIds`。
   - `wails generate module` 重建 `models.ts`。
3. **后端标识符改名**：`syncAgentProjectsLocked` → `syncAgentTasksLocked`、
   `assignAgentProjectsLocked` → `assignAgentTasksLocked`、`normalizeProjectIDs` → `normalizeTaskIDs`、
   `maxAgentProjects` → `maxAgentTasks`；相关注释里的 project 措辞一并改掉。
4. **后端报错文案**：`关联项目必须是代码类型的任务` → **`只能关联代码类型的任务`**
   （顺带把同一函数里 010 就改过的界面名统一：`下次重置时间格式无效` → `下次可用时间格式无效`）。
5. **前端命名**：`projectIds` → `taskIds`、`codeProjects` → `codeTasks`、
   `mirrorAgentProjects` → `mirrorAgentTasks`；
   类名 `.agent-projects-editor` → `.agent-tasks-editor`、`.agent-project-list` → `.agent-task-options`、
   `.agent-project-item` → `.agent-task-option`（**刻意避开 `agent-task-list` / `agent-task-row`**，
   那是列表行展开区的既有类名，撞名会串样式）。
6. **旧存档迁移**：读到只有 `projectIds` 的存档时，把它搬进 `taskIds`，**迁移不写回旧键**；
   下一次保存就是新键。旧键只在迁移函数里以局部结构体读取，不挂在 `AgentApp` 上
   （挂在结构体上就迟早会被写回去）。

## 不做什么

- **不改 `Task.AgentID` 这条唯一事实来源**：`syncAgentTasksLocked` 的重建规则、
  `assignAgentTasksLocked` 的「任务侧说了算」、`repairAgentLinksLocked` 的幂等修复全部不动，
  只改名字。
- **不改 `Task` 结构与任务侧的字段名**（任务本来就叫 task）。
- 不改任何布局与样式数值；类名只是换了名字，规则内容一字未动。
- 不改 002/003 等历史规格正文（它们记录的是当时的命名），由本规格说明被修订之处。

## 验收标准

1. 界面任何位置都不再出现「项目」二字：右侧面板、弹窗、列表行提示、左侧导航、
   任务面板的「关联 Agent」下拉，全文档级断言通过。
2. 全文档没有任何 `class` 含 `project`（旧类名彻底消失）。
3. 右侧面板与新增弹窗的关联区标题都是 `关联任务`，文字行宽度一致（同为 4 个汉字，60.00px）。
4. 保存链路的键名是 `taskIds`：`UpdateAgentApp` / `CreateAgentApp` 的入参键恰为
   `available,model,name,resetAt,taskIds`，**不含 `projectIds`**。
5. 关联行为无回归：取消勾选后任务行不再带 `· <Agent 名>`，同 Agent 下的另一条仍带着
   （说明本地镜像与 `Task.AgentID` 这条事实来源都没被改名碰坏）。
6. 旧存档兼容：只有 `projectIds` 键的文件加载后关联仍在，且**回写时用的是 `taskIds`**、
   不再出现 `projectIds`（Go 单测 `TestLoadMigratesLegacyProjectKey`）。
7. ≤1000px 时详情面板仍按设计隐藏（007 的媒体查询未受影响）。
