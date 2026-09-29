# 022 任务

## 实现

- [x] `App.css`：`.chart-panel,.output-list` 的 `flex:1 1 auto` → `flex:0 1 auto`
- [x] `App.css`：滚动归属注释补 022 段（含 `min-height:200px` 的实测依据）
- [x] 探针：`?regress=1` 改为复现 022 之前；新增 `?regress020=1`（021 对照组）与 `?floor0=1`
- [x] 探针：新增「不撑满」4 条断言 + `declared()` 读 CSSOM 比对 `.task-list`
- [x] 探针：横幅断言计数改为 `asserts`（不再把 note 读数算进去）
- [x] 脚手架目录 `.workbuddy/verify-021/` → `.workbuddy/verify-ui/`（已不只服务 021）

## 验证矩阵（真实数据 22 任务 / 4 应用 / 无头 Chrome）

正向（全部 0 失败）：

| 场景 | 窗口 | 卡片高 | 内容高 | 卡片下方留白 | 断言 |
| --- | --- | --- | --- | --- | --- |
| 全部产出 | 1920x1200 | 432 | 430 | 90 | 11/11 |
| 全部产出 | 1920x1040 | 362 | 430 | 0（吃满） | 12/12 |
| 文档 | 1440x900 | 483 | 481 | 19 | 11/11 |
| 代码 | 1440x900 | 475 | 473 | 27 | 11/11 |
| 实施 | 1440x900 | 200（下限） | 198 | 302 | 11/11 |
| 文档 | 1440x760 | 362 | 481 | 0（吃满） | 11/11 |
| 全部产出 | 1024x680 | 200 | 430 | −198（外层兜底） | 11/11 |
| 文档 | 1024x680 | 282 | 536 | 0（吃满） | 11/11 |
| Agent（未改动的守卫） | 1920x1040 | 面板 789 | 787 | 0（仍撑满） | 3/3 |

合计 **92 条断言全绿**（产出 89 + Agent 3），另有每场景 4–6 条读数。

对照组（同一构建、同一套断言，只多注入一条规则）：

| 对照组 | 注入 | 结果 | 翻红的断言 |
| --- | --- | --- | --- |
| `?regress=1`（022 之前） | `flex:1 1 auto` | 8/11 | `flex-grow=1`；与 `.task-list` 不一致；留白 0、卡片被拉到 502 |
| `?regress020=1`（021 之前） | 滚动挂 `.output-body` | 8/11 | 滚动对象退回 `.output-body`；body 837>789；KPI 行 top 121→73 |
| `?floor0=1`（读数量用） | 摘 `min-height` | — | 1024x680 下卡片只剩 38px（`clientHeight` 21） |

## 截图

`specs/022-output-panel-hug/screenshots/`

| 文件 | 内容 |
| --- | --- |
| `1440x900-output-doc-hug.png` | 修复后：文档列表贴合内容，卡片下方留白 |
| `1440x900-output-doc-stretch-regress022.png` | 对照组：卡片被拉到窗口底部（用户看到的现象） |
| `1440x900-output-support-minfloor.png` | 内容极短（198px），卡片由 200px 下限托住 |
| `1920x1200-output-all-hug.png` | 窗口够高：饼图卡片贴合（432），留白 90 |
| `1920x1040-output-all-fill.png` | 窗口不够高：饼图卡片吃满并内滚（读数与 021 一致） |
| `1440x760-output-doc-fill.png` | 中等窗口：列表超出后在卡片内滚动 |
| `1024x680-output-all-fallback.png` | 最小窗口：卡片 200px，外层兜底滚 198px |
| `1024x680-output-all-floor0-38px.png` | 摘掉 `min-height` 的后果：卡片只剩 38px |
| `1920x1040-agent-unchanged.png` | Agent 未改动（仍撑满） |
| `1440x900-output-doc-verify.png` | 断言横幅（正向：11 条断言 0 失败） |
| `1440x900-output-doc-regress022-verify.png` | 断言横幅（对照组：3 条 022 断言翻红） |

## 未做

- 未动【AGENT】【单位管理】面板的 `flex:1 1 auto`（用户明确这两处是对的，仍撑满）。
- 未把 `.task-list` 的 `min-height:0` 照搬到卡片上（理由与实测见 design.md）。
