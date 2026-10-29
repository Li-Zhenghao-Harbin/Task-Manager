# 024 任务清单与验证矩阵

## 实现

| # | 改动 | 文件 |
| --- | --- | --- |
| 1 | `candidateAgents()` 去掉 `available` 过滤与「已关联特例」，直接返回全部 Agent | `frontend/src/App.vue` |
| 2 | 新建任务弹窗调用点 `candidateAgents(draft)` → `candidateAgents()` | 同上 |
| 3 | 任务详情面板调用点 `candidateAgents(selected)` → `candidateAgents()` | 同上 |
| 4 | 注释重写：说明为什么以前只列可用的、为什么现在不需要那条特例 | 同上 |
| 5 | 新增 `TestUnavailableAgentStillAttachesToTask`（创建 / 改派 / 派生视图三条路径） | `agent_test.go` |
| 6 | 探针新增 `view=create`（含 `?old=1` 对照、`?expand=1` 截图辅助） | `.workbuddy/verify-ui/` |
| 7 | mock 的 `CreateTask` 改成「转发 + 回显」，让创建后能读到新任务 | 同上 |

模板未改：「（不可用）」后缀本来就在，只是以前永远命中不到。

## 验证矩阵

真值数据：4 个 Agent（`Codex-GPT` / `WorkBuddy-Hy` 不可用 / `TRAE-Deepseek` / `WorkBuddy-Deepseek`）、22 个任务。

### 正向

| 场景 | 命令 | 结果 |
| --- | --- | --- |
| 新建任务下拉 | `create 1920,1040 tall=1` | **12/12 PASS** |
| 全流程（选中 → 提交 → 刷新后详情面板） | 同上 | 同上 |
| 实施列表贴合（023 回归） | `output 1440,900 filter=实施 tall=1` | 13/13 PASS |
| 全部产出吃满（023 回归） | `output 1920,1040 tall=1` | 12/12 PASS |
| 最小窗口兜底（023 回归） | `output 1024,680` | 11/11 PASS |
| Agent 面板（023 回归） | `agent 1920,1040 tall=1` | 5/5 PASS |
| 单位管理（023 回归） | `units 1920,1040 tall=1` | 5/5 PASS |
| 全部任务参照物（023 回归） | `tasks 1920,1040 tall=1` | 7/7 PASS |

### 对照组（`?old=1`：把「只列可用项」的输出还原到 DOM）

`create 1920,1040 tall=1&old=1` → **10 条失败 / 12 条断言**

| 判据 | 修复后 | 对照 |
| --- | --- | --- |
| 选项数 = 全部 Agent + 「不关联」 | 5 项 | **4 项** |
| 全部 Agent 都在列表里 | 共 4 个 | **缺 1 个：WorkBuddy** |
| 不可用项带「（不可用）」后缀 | `WorkBuddy-Hy（不可用）` | **(缺)** |
| 不可用的 Agent 能被选中 | `select.value=a2f7b462…` | **`(空)`** |
| 提交的 payload 带那个 id | `agentId=a2f7b462…` | **`(空)`** |
| 详情面板仍挂着它 | `value=a2f7b462…` | **`(空)`** |
| 详情面板标注为（不可用） | `WorkBuddy-Hy（不可用）` | **`不关联`** |
| 详情面板选项数 / 全列出 / 带后缀 | 5 项 / 共 4 个 / 有 | **4 项 / 缺 1 / (缺)** |

仍保持绿色的两条是「可用项不带后缀」—— 旧实现也不会误标，这条本来就抓不到差异，属于预期。

## 截图

`specs/024-agent-picker-all/screenshots/`

| 文件 | 内容 |
| --- | --- |
| `1440x900-create-picker-expanded.png` | 下拉展开（`size=6`，仅为截图）：5 项，`WorkBuddy-Hy（不可用）`被选中 |
| `1440x900-create-picker-expanded-old.png` | 同位置对照：4 项，没有那一项 |
| `1600x1000-create-verify.png` | 断言横幅 12 条全绿 |
| `1600x1000-create-verify-old.png` | 对照组横幅 10 条红 |
| `1440x900-created-task-details.png` | 创建后：任务列表 +1，详情面板 `关联 Agent = WorkBuddy-Hy（不可用）` |

## 构建链

- `gofmt -l .` 无输出、`go vet ./...` 无输出、`go test ./...` ok
- `vue-tsc --noEmit` 通过、`prettier --check src/App.vue` 通过
- `vite build` 通过（JS 哈希 `b5b8df49` → `6f6daf8c`，说明 `App.vue` 确实改了）
- `wails build -s` 17.6s 成功；exe 内**已无** `.filter(o=>o.available)`（旧产物有 2 处，exe 与 0 处），
  且含新代码上下文片段

## 未做

- 没给下拉加「可用 / 不可用」分组或置顶排序 —— 用户只要求「能选」，排序保持 Agent 列表顺序。
- 没做「选了不可用 Agent 时的额外提示」（如保存时提醒一句）。选项文字里已有「（不可用）」。
- 没动 Agent 列表页、可用情况的自动恢复与通知。
