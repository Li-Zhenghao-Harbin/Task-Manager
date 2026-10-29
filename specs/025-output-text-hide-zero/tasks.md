# 025 任务清单与验证矩阵

## 实现

| # | 改动 | 文件 |
| --- | --- | --- |
| 1 | `taskOutputText()` 增加 `.filter((key) => (task.output[key] \|\| 0) !== 0)` | `frontend/src/App.vue` |
| 2 | 模板给 `<em>` 加 `v-if="taskOutputText(task)"`（全零时整段不渲染） | 同上 |
| 3 | 探针新增 `view=outtext`（按真值逐行比对）、`?old=1` 对照、`?nodetail=1` 停在列表页截图 | `.workbuddy/verify-ui/` |

CSS 未改 —— 摘掉一项后 `justify-content:space-between` 的布局自然成立。

## 验证矩阵

真值数据：代码 6 条已完成任务（其中 2 条全零）、文档 6 条（每条都有 0 项）、实施 1 条。

### 正向

| 场景 | 命令 | 结果 |
| --- | --- | --- |
| 【代码】列表 | `outtext 1920,1040 filter=代码 tall=1` | **9/9 PASS** |
| 【文档】列表 | `outtext 1920,1040 filter=文档 tall=1` | **9/9 PASS** |
| 【实施】列表 | `outtext 1920,1040 filter=实施 tall=1` | **6/6 PASS** |

【代码】逐行读数（改动后）：

| 任务 | 摘要 |
| --- | --- |
| Clou BusinessOpportunityGrabber | `应用程序 1` |
| Task Manager | `应用程序 1` |
| ClouESP 8.0 | *(无摘要)* |
| ClouESP Docker部署 | *(无摘要)* |
| Clou ReleaseFlow | `应用程序 1` |
| Meter Group Management显示优化 | `修改功能 1` |

【文档】逐行读数：`新增文档 23` / `新增文档 1` ×3 / `Skills 1` / `新增文档 3`。

### 对照组（`?old=1`：把旧的全量文本还原到 DOM）

| 场景 | 结果 | 翻红的判据 |
| --- | --- | --- |
| 【代码】 | **3 / 9 失败** | 摘要含 0（6 行全部）、逐行不一致（6 行）、全零任务仍渲染 em（2 行） |
| 【文档】 | **2 / 9 失败** | 摘要含 0（6 行全部）、逐行不一致（6 行） |

文档筛选下「全零不渲染」保持绿 —— 该筛选 6 行都有非零项，本来没有可对照的全零行，属预期。

### 回归（019–024）

| 场景 | 结果 |
| --- | --- |
| 实施列表贴合 | 13/13 |
| 全部产出吃满 | 12/12 |
| 最小窗口兜底 | 11/11 |
| Agent 面板 | 5/5 |
| 单位管理 | 5/5 |
| 全部任务参照物 | 7/7 |
| 新建任务下拉（024） | 12/12 |

## 截图

`specs/025-output-text-hide-zero/screenshots/`

| 文件 | 内容 |
| --- | --- |
| `1440x900-code-after.png` | 【代码】列表：`应用程序 1` / `修改功能 1`，两行全零无摘要 |
| `1440x900-code-before.png` | 同位置对照：`新增功能 0 \| 修改功能 0 \| 应用程序 1` |
| `1440x900-doc-after.png` | 【文档】列表：`新增文档 23` / `Skills 1` |
| `1440x900-doc-before.png` | 同位置对照：每行都拖着两个 `0` |
| `1600x1000-code-verify.png` | 断言横幅 9 条全绿 |
| `1600x1000-code-verify-old.png` | 对照组横幅 3 条红 |
| `1600x1000-doc-verify.png` | 【文档】断言横幅 9 条全绿 |

## 构建链

- `gofmt -l .` 无输出、`go vet ./...` 无输出、`go test ./...` ok
- `vue-tsc --noEmit` 通过、`prettier --check src/App.vue` 通过
- `vite build` 通过（JS 哈希 `6f6daf8c` → `31541b70`）
- `wails build -s` 成功；exe 内嵌的是
  `Ut[c.taskType].filter(o=>(c.output[o]||0)!==0).map(...)`（旧版无 `.filter`），已核对

## 未做

- 没把「全零任务」从列表里去掉 —— 它是真实任务，只是产出还没填。
- 没给「全零」补一句占位文案（如「暂无产出」）—— 用户要的是不显示，不是换一句话。
- 没动详情面板、KPI 卡片、饼图、菜单计数。
