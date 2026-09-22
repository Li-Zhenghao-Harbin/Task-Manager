# 014 任务清单

- [x] `App.vue`：Agent 详情面板关联区 `<h3>关联项目</h3>` → `<h3>关联任务</h3>`
      （按四行上下文匹配，避开弹窗里同名的那个实例）
- [x] `App.vue`：行展开区提示 `未关联任何代码任务，可在右侧「关联项目」中勾选。` → `…「关联任务」…`
- [x] `App.css`：`.details.agent-details` 上方注释里的「关联项目列表」→「关联任务列表」
- [x] `outline.md`：字段表 / 列表行说明 / Prompt 生命周期 / 详情面板布局 / 数据规则 五处界面文案改名，
      并在字段表与派生视图一段注明「界面上叫关联任务，数据字段仍是 `projectIds`」
- [x] 全仓检查残留：`grep -rn 关联项目` → 只剩 `outline.md` 那句解释性括注、
      `App.vue` 的弹窗实例（本轮范围外）、历史 specs 与后端错误文案（均不改）
- [x] 前端构建：`npm --prefix frontend run build`（含 `vue-tsc`）通过
- [x] 无头界面断言：宽屏 20/20、窄屏 5/5 全绿
- [x] 干净版式截图：面板标题已显示为【关联任务】
- [x] Go 侧未改动，仍跑一遍 `gofmt -l` / `go vet` / `go test ./...`
- [x] `wails build` 重建 `build/bin/Task Manager.exe`
- [x] 项目记忆：当日日志 + `MEMORY.md` 文案约定
- [x] 技能 `wails-vue-ui-verify`：补「块级元素 rect 量不出文字宽度」与「窄屏先判可见性」两条判据
