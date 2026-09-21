# 002 Agent 模块实施任务

- [x] 编写功能规格与验收标准（spec.md）。
- [x] 编写技术设计：数据模型、接口、自动重置与引用完整性（design.md）。
- [x] 后端：新增 `AgentApp` / `AgentAppInput` 模型与 `DataStore.Agents` 持久化。
- [x] 后端：实现 `ListAgentApps` / `CreateAgentApp` / `UpdateAgentApp` / `DeleteAgentApp`。
- [x] 后端：实现校验规则与「到点自动恢复可用」。
- [x] 后端：`Task.AgentID` 字段、代码任务约束、删除时的引用清理。
- [x] 前端：左侧菜单新增 Agent 分组与 Agent应用子菜单。
- [x] 前端：Agent 应用列表视图（应用名-模型名、可用徽标、重置时间、关联项目）。
- [x] 前端：右侧详情面板编辑 Agent 应用，新增模态框、删除二次确认。
- [x] 前端：任务详情与创建表单在代码任务时显示「关联 Agent」单选。
- [x] 前端：任务列表行展示所关联的 Agent。
- [x] 重建 wailsjs 绑定，运行 vue-tsc 类型检查与前端构建。
- [x] 运行 go vet 与 Wails Windows 构建，产出 EXE。
- [x] 补充 `agent_test.go`（11 项用例）覆盖校验、自动重置、引用清理与持久化。
- [x] 无头浏览器截图核对界面（见 screenshots/）。

## 顺带修复

- `load()` 原先以「`tasks` 或 `units` 是否非空」判断数据文件格式，导致只含 Agent（或只含设置）的数据文件被误判为旧版数组格式而读取失败，进而在下次保存时丢失数据。改为按 JSON 顶层首个有效字符判断是 `[` 还是 `{`，`ImportData()` 同步修复。由 `TestAgentPersistsAcrossReload` 覆盖。
