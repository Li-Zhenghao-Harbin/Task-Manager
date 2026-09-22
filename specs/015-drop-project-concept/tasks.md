# 015 任务清单

- [x] `app.go`：`ProjectIDs`→`TaskIDs`、json 键 `projectIds`→`taskIds`（含 `AgentAppInput`）
- [x] `app.go`：`syncAgentProjectsLocked` / `assignAgentProjectsLocked` / `normalizeProjectIDs` / `maxAgentProjects` 改名与注释
- [x] `app.go`：报错文案 `关联项目必须是代码类型的任务` → `只能关联代码类型的任务`；
      顺带把 `下次重置时间格式无效` → `下次可用时间格式无效`（对齐 010 的界面名）
- [x] `app.go`：新增 `migrateAgentTaskIDsLocked(raw)`，在 `load()` 与 `ImportData()` 里于 `repairAgentLinksLocked()` 之前调用
- [x] `agent_test.go` / `prompt_test.go`：`ProjectIDs`→`TaskIDs`、测试名与断言文案里的 project 措辞
- [x] `agent_test.go`：新增 `TestLoadMigratesLegacyProjectKey`（只有旧键的存档仍保住关联，回写只有新键）
- [x] `App.vue`：`projectIds`→`taskIds`、`codeProjects`→`codeTasks`、`mirrorAgentProjects`→`mirrorAgentTasks`
- [x] `App.vue`：类名改名；新增弹窗标题 `关联项目` → `关联任务`
- [x] `App.css`：6 处类名选择器与注释同步
- [x] `wails generate module` 重建绑定（`models.ts` 里 `taskIds`）
- [x] `prettier` 格式化 `App.vue`（`App.css` 不参与）
- [x] `gofmt -l` / `go vet` / `go test ./...`（33 项）全过
- [x] `npm --prefix frontend run build`（含 `vue-tsc`）通过
- [x] 无头断言：宽屏 40/40、窄屏 4/4；含负载键名与「取消勾选后任务行不再带 Agent 名」两组关键断言
- [x] `wails build` 重建 `build/bin/Task Manager.exe`
- [x] 规格：`specs/014-agent-panel-task-label/` 标注被本规格修订之处
- [x] `outline.md`：去掉「数据字段仍是 `projectIds`」的说法，字段表/数据规则改 `taskIds`
- [x] 项目记忆（当日日志 + `MEMORY.md`）；技能 `wails-vue-ui-verify` 升 1.13.0
