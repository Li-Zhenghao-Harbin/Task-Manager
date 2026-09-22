# 007 任务拆解

## T1 模板：Agent 详情面板加类名

- [ ] `App.vue`：`<aside v-else-if="selectedAgent" class="details">` → `class="details agent-details"`
- [ ] 确认字符串唯一（全文件 1 处）

验收：`grep -c 'details agent-details'` 为 1。

## T2 样式：Agent 详情面板纵向 flex + 列表撑满 + 删除置底

- [ ] 在 `App.css` 末尾追加：

```css
/* Agent 详情面板：关联项目列表占满剩余高度，删除应用固定在最下方 */
.details.agent-details{display:flex;flex-direction:column;overflow:hidden}
.details.agent-details>*{flex:0 0 auto}
.details.agent-details .agent-projects-editor{flex:1 1 auto;min-height:0;display:flex;flex-direction:column}
.details.agent-details .agent-project-list{flex:1;min-height:96px;max-height:none;overflow:auto;border:1px solid var(--line);border-radius:10px;background:#fbfdff;padding:2px 10px}
.details.agent-details .detail-actions{margin-top:16px;padding-top:14px;border-top:1px solid var(--line)}
@media(max-width:1000px){.details.agent-details{display:none}}
```

验收：
- `max-height:220px` 的老规则仍在（对新增弹窗生效），但被 `.details.agent-details .agent-project-list{max-height:none}` 覆盖。
- 窄屏媒体查询补偿已写入，且位置在新规则之后。

## T3 模板：Prompt 卡片改横向两栏

- [ ] 用 `.prompt-side` 包裹 `Prompt N` + 复制 + 删除，textarea 保持同级
- [ ] 删除原 `.prompt-card-head` / `.prompt-actions` 结构

验收：`grep -c 'prompt-card-head'` 为 0，`grep -c 'prompt-side'` 为 1。

## T4 样式：Prompt 卡片一行一条

- [ ] `.prompt-grid` 改 `display:flex;flex-direction:column;gap:14px`
- [ ] `.prompt-card` 改 `display:flex;align-items:stretch;gap:16px`
- [ ] 新增 `.prompt-side` / `.prompt-label`
- [ ] `.prompt-card textarea` 改 `flex:1;min-width:0;width:auto`
- [ ] 删除 `.prompt-card-head`、`.prompt-actions` 两条死规则

验收：`grep -c 'auto-fill'` 为 0（该栅格已移除）。

## T5 回归与打包

- [ ] `gofmt -l .` 无输出、`go vet ./...`、`go test ./...` 全过
- [ ] `npm --prefix frontend run build`（含 `vue-tsc`）通过
- [ ] `wails build` 产出 `build/bin/Task Manager.exe`

## T6 无头界面验证

按技能 `wails-vue-ui-verify`：mock 绑定 + 静态预览 + Chrome 无头截图，断言结果渲染进页面横幅。

- [ ] 视图 `panel`：Agent 详情面板——列表高度 > 220px；列表底边等于删除区顶边；
      删除按钮是面板最后一个可见元素且贴底；新增弹窗的列表仍为 220px
- [ ] 视图 `prompt`：一行一条（卡片宽度 ≈ board 宽度）；卡片内 `Prompt N`/复制/删除 三者左边缘相同、
      纵向依次排列；复制与删除等宽；textarea 右边缘与卡片内右边距对齐

产出：`specs/007-agent-panel-layout/screenshots/*.png`

## T7 文档同步

- [ ] `outline.md`：补 007 的界面布局说明
- [ ] `.workbuddy/memory/`：当日日志 + `MEMORY.md`（若有新约定）
