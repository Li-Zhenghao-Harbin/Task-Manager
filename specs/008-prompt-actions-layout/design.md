# 008 设计

## 改动范围

| 文件 | 改动 |
| --- | --- |
| `frontend/src/App.vue` | Prompt 视图模板：返回按钮文案、标题行结构、卡片结构 |
| `frontend/src/App.css` | `.prompt-title-text` / `.prompt-body` / `.prompt-foot` 等新规则 |

后端零改动。

## 1. 返回按钮

```diff
-<button type="button" class="back-button" @click="closePromptView">← 返回</button>
+<button type="button" class="back-button" @click="closePromptView">返回</button>
```

只改文案。`.back-button` 的样式不变（padding 8px 13px，去掉箭头后按钮自然变窄）。

## 2. 标题行

原结构：标题行只放按钮和 `h1`，agent 名是标题行**外面**的兄弟节点，所以必然另起一行。

```diff
-<div class="prompt-title-line">
-  <button class="back-button">← 返回</button>
-  <h1>{{ selectedAgentTask.name }}</h1>
-</div>
-<small class="agent-sub">{{ agentLabel(selectedAgent) }}</small>
+<div class="prompt-title-line">
+  <button class="back-button">返回</button>
+  <div class="prompt-title-text">
+    <h1>{{ selectedAgentTask.name }}</h1>
+    <small class="agent-sub">{{ agentLabel(selectedAgent) }}</small>
+  </div>
+</div>
```

新增一层 `prompt-title-text` 包裹 `h1` + `agent-sub`，目的是拿到**两个不同的间距**：

- 按钮 → 标题块：沿用 `.prompt-title-line` 的 `gap:14px`；
- 任务名 → agent 名：`.prompt-title-text` 自己的 `gap:8px`，让 agent 名看起来是标题的后缀而不是并列项。

对齐用 `align-items:baseline`（26px 的 `h1` 与 12px 的小字按基线对齐），比 `center` 更稳：
中文字形没有下伸部分，`center` 会让小字看起来偏上。外面那层仍是 `center` —— 按钮和标题块按中心对齐。

`.agent-sub` 原本是 `display:block; margin-top:3px`，会被这条更具体的新规则盖掉：

```css
.prompt-title-text{display:flex;align-items:baseline;gap:8px;flex-wrap:wrap}
.prompt-title-text .agent-sub{display:inline;margin-top:0}
```

`flex-wrap:wrap` 保留：任务名特别长时 agent 名仍然换行，不会把标题挤成一行溢出。

## 3. 卡片：上段内容 + 底段操作区

原来卡片本身就是 `display:flex` 的横向两栏（`.prompt-side` + `textarea`），删除按钮在左列里，
所以它必然紧挨【复制】。要让【删除】落到卡片底部，就得把卡片改成纵向、横向两栏下沉成一层：

```diff
 <article class="prompt-card">
-  <div class="prompt-side">
-    <span class="prompt-label">Prompt N</span>
-    <button class="prompt-copy">复制</button>
-    <button class="prompt-delete">删除</button>
-  </div>
-  <textarea ...></textarea>
+  <div class="prompt-body">
+    <div class="prompt-side">
+      <span class="prompt-label">Prompt N</span>
+      <button class="prompt-copy">复制</button>
+    </div>
+    <textarea ...></textarea>
+  </div>
+  <div class="prompt-foot">
+    <button class="prompt-delete">删除</button>
+  </div>
 </article>
```

```css
.prompt-card{display:flex;flex-direction:column;gap:10px;...}   /* 原来是 row + align-items:stretch */
.prompt-body{display:flex;align-items:stretch;gap:16px}          /* 承接原来的横向两栏语义 */
.prompt-foot{display:flex;justify-content:flex-end;border-top:1px solid var(--line);padding-top:10px}
```

要点：

- `.prompt-card textarea{flex:1;...}` 不用改，flex 上下文从卡片换成了 `.prompt-body`，语义相同；
- 【删除】放在底段**右端**：与左列顶部的【复制】构成一条对角线，是两个按钮之间能达到的最大距离；
- 底段加 `border-top` 分隔线，视觉上把「删除整条记录」和「编辑内容」分开，也解释了为什么它不在左列里；
- `align-items:stretch` 从卡片搬到 `.prompt-body`，保证左列与文本框仍然等高（007 的验收项之一）。

## 风险与取舍

- 没动 `.prompt-side` 的 `width:96px`：少一个按钮后左列只剩 `Prompt N` + 【复制】，宽度依旧够用，
  改动越小越好；如果以后左列只剩一个元素再收窄。
- 没把【删除】做成整宽按钮（`detail-actions` 那种）：它是行内操作而不是卡片级主操作，
  保持小按钮形态，和【复制】同一视觉量级。
