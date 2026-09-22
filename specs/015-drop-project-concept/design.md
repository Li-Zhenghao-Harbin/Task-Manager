# 015 设计

## 为什么这次要动字段名与类名（014 的判断被推翻）

014 里我给的结论是「`projectIds` 与 `.agent-projects-*` 用户看不到，改名收益为零、回归面大」。
这个判断在**单看 014 那一轮**时成立，但放在「项目这个概念不存在」的前提下就不成立了：
字段名和类名是后来人读代码时唯一的语境来源，留着 `projectIds` 就等于把这个概念又教给下一个人。
真正的成本差异只是「机械改名 + 一次全量验证」，而回归面由三条测试线兜住（Go 31→33 项、
前端 `vue-tsc`、无头断言 40/40），可控。

## 改名映射

| 旧 | 新 | 位置 |
| --- | --- | --- |
| `AgentApp.ProjectIDs` / `json:"projectIds"` | `TaskIDs` / `json:"taskIds"` | `app.go` |
| `AgentAppInput.ProjectIDs` | `TaskIDs` | `app.go` |
| `syncAgentProjectsLocked` | `syncAgentTasksLocked` | `app.go`（4 处调用） |
| `assignAgentProjectsLocked` | `assignAgentTasksLocked` | `app.go`（2 处调用） |
| `normalizeProjectIDs` | `normalizeTaskIDs` | `app.go` |
| `maxAgentProjects` | `maxAgentTasks` | `app.go` |
| `agent.projectIds` / `agentDraft.projectIds` … | `taskIds` | `App.vue`（12 处） |
| `codeProjects` | `codeTasks` | `App.vue` |
| `mirrorAgentProjects` | `mirrorAgentTasks` | `App.vue` |
| `.agent-projects-editor` | `.agent-tasks-editor` | `App.vue` / `App.css` |
| `.agent-project-list` | `.agent-task-options` | `App.vue` / `App.css` |
| `.agent-project-item` | `.agent-task-option` | `App.vue` / `App.css` |

**类名为什么不用 `.agent-task-list` / `.agent-task-item`**：`.agent-task-list` 与 `.agent-task-row`
是列表行展开区（006 起）的既有类名，`.details.agent-details .agent-task-list` 不是个选择器，
但 `agent-task-list` 一旦和编辑区的列表同名，将来任何「按类名找列表」的代码都会同时命中两处。
所以编辑区用 `-editor` / `-options` / `-option` 这一组，与列表区的 `-list` / `-row` 明确分开。

## 旧存档迁移

关系本来就有两层保险：`Task.AgentID` 是唯一事实来源，`TaskIDs` 只是派生视图。
但 `repairAgentLinksLocked()` **会读 Agent 侧的列表去回填任务的 `AgentID`**（003 那次数据修复的路径），
所以直接把 JSON 键换成 `taskIds` 会让「只写过 projectIds 的老文件」丢掉这层回填。

做法是加一个只在加载时读一次的迁移，**用局部结构体读原始 JSON**，而不是在 `AgentApp` 上留字段：

```go
func (a *App) migrateAgentTaskIDsLocked(raw []byte) {
	var legacy struct {
		Agents []struct {
			ID         string   `json:"id"`
			TaskIDs    []string `json:"taskIds"`
			ProjectIDs []string `json:"projectIds"`
		} `json:"agents"`
	}
	...
	if idx < 0 || len(item.ProjectIDs) == 0 || len(a.agents[idx].TaskIDs) > 0 {
		continue
	}
	a.agents[idx].TaskIDs = append([]string{}, item.ProjectIDs...)
}
```

- **挂在 `AgentApp` 上会写回去**：结构体上的字段会被 `json.Marshal` 序列化，下一次保存就又把
  `projectIds` 写进文件了，等于迁移永远不收敛。局部的「只读投影」没有这个隐患。
- **新键非空就不动它**：迁移只在 `taskIds` 为空时生效，避免旧键覆盖新键。
- `load()` 与 `ImportData()` 都要调用（导入的可能是老文件），且在 `repairAgentLinksLocked()` **之前**
  ——回填要靠 `TaskIDs` 已经就位。

## 验证

沿用 `wails-vue-ui-verify`：mock 绑定 + Chrome 无头 + 断言横幅。

本轮判据上的新认识（已写进技能 1.13.0）：

1. **想验证「A 取消关联后 B 的变化」，别去点详情面板，直接看列表行的小字**。
   任务行会追加 `· <Agent 名>`，取消关联后这一小段消失，是最贴近用户视角的证据。
   我最初写的是「点任务行 → 看面板里『关联 Agent』下拉是否回到『不关联』」，结果拿到的却是
   另一个任务（`selected` 已经被自动选中为第一条任务，而 `waitFor` 返回的行在点击前被 Vue 重渲染过），
   白拿一个假失败——**`waitFor` 拿到的节点到 `click()` 之间又发生了一次渲染**，这是「别跨渲染持有元素引用」
   的另一种形态。
2. **断言「负载键名」比断言「界面上没有某个词」更能钉住字段改名**。
   直接断言 `Object.keys(payload)` 等于 `available,model,name,resetAt,taskIds`，
   以及 `!('projectIds' in payload)`，就把「协议里的键名」钉死在测试里。

### 结果

| 场景 | 断言 |
| --- | --- |
| 1440×900（面板 → 取消勾选 → 任务页 → 弹窗创建） | **40/40** |
| 1000×700 | **4/4** |
| Go | **33/33**（新增 `TestLoadMigratesLegacyProjectKey`） |

截图见 `screenshots/`：`01-banner`（宽屏断言证据）、`02-panel-clean`（面板干净版式）、
`03-modal-clean`（弹窗干净版式）、`04-narrow-1000`（窄屏断言证据）。
