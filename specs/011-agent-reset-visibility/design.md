# 011 · 设计

## 结构变化（Agent 列表行）

```
不可用（宽屏，4 格）
┌────┬──────────────────────────┬────────────────────────────┬──────────┐
│ ▾  │ Codex CLI-GPT-5.2        │ 下次可用时间：2026-09-22 15:00 │  [不可用] │
└────┴──────────────────────────┴────────────────────────────┴──────────┘
   16px        minmax(0,1fr)              max-content               78px

可用（宽屏，3 格：日期整格不进 DOM）
┌────┬──────────────────────────────────────────────────────────┬──────────┐
│ ▾  │ ClouESP 8.0-Claude Opus 4.1                              │  [可用]  │
└────┴──────────────────────────────────────────────────────────┴──────────┘
   16px                  minmax(0,1fr)                             78px
```

## 为什么用 `.available` 类而不是「永远渲染、内容置空」

`.agent-row` 自己是 grid 容器，**每一行独立计算列宽**。如果把 `.agent-reset` 留着只清空文本：

- 该行的第 3 列轨道仍然存在，`gap:12px` 还会在它两侧各留一份 → 标题与状态标签之间凭空多出 24px 空档；
- 列数（4）与可见子元素数（3）不一致，一旦以后有人把某列改成固定宽度或 `min-content`，末项就会被挤到第二排
  —— 010 正是被这个坑咬过一次。

所以正确做法是**让整格不渲染**（`v-if="!agent.available"`），并同时把该行的栅格降到 3 列。
两处必须一起改：只 `v-if` 不改 CSS，末项（可用情况标签）会掉到新建的第 4 列以外，
栅格按 `16px minmax(0,1fr) max-content 78px` 只有 4 条轨道，多出的子元素会被放到隐式行里换行。

## CSS

```css
/* 基规则：不可用行（日期列在场） */
.agent-row{grid-template-columns:16px minmax(0,1fr) max-content 78px}
/* 可用行：少一个子元素，栅格同步降为 3 列 */
.agent-row.available{grid-template-columns:16px minmax(0,1fr) 78px}

@media(max-width:1100px){
  /* 日期列整列隐藏；媒体查询不增加特异性，必须显式列上 .available */
  .agent-row,.agent-row.available{grid-template-columns:16px minmax(0,1fr) 78px}
  .agent-reset{display:none}
}
```

`.agent-row.available` 特异性 0,2,0 高于基规则 0,1,0，无需 `!important`；
但它同样高于媒体查询里的 `.agent-row`（0,1,0），所以断点里必须把两个选择器并列写，
否则窄屏下列定义会被它盖掉（当前两处取值相同，属于**提前把规则写明确**，避免以后改断点时踩空）。

## 模板

```html
<button class="agent-row" :class="{ selected: …, expanded: …, available: agent.available }" …>
  <span class="agent-caret">…</span>
  <span class="agent-title">{{ agentLabel(agent) }}</span>
  <span v-if="!agent.available" class="agent-reset">下次可用时间：{{ resetText(agent.resetAt) }}</span>
  <span class="agent-pill" :class="{ offline: !agent.available }">…</span>
</button>
```

- 行的 `available` 类只服务于栅格列数，不参与任何配色/交互（`.agent-row.selected` / `:hover` 不受影响）。
- `resetText()` 仍被不可用行使用；`未设置` 分支保留。

## 与面板的一致性

右侧面板本来就是 `v-if="!selectedAgent.available"`，本轮改动之后，
「可用时不出现该字段」在列表行与面板两处规则一致，不再有「一处显示 `—`、一处整块消失」的分歧。

## 验证设计

`specs/011-agent-reset-visibility/screenshots/` 下的截图由 mock 绑定 + Chrome 无头探针生成
（`?view=list|panel`、`?pos=none` 出干净版式图、`?narrow=1` 走窄屏断点断言）。
断言刻意包含两类反面对照：

1. **列数一致性**：可用行 3 轨道 / 不可用行 4 轨道，同时比对「可见子元素数」，两者必须相等；
   混排时三行的标题左缘与状态标签右缘差值 ≤1px —— 这是「列数没写错」最直接的证据。
2. **宽度回收**：可用行标题列必须明显宽于不可用行（实测 606px vs 417px），
   证明释放出的日期列宽度确实被标题列吃满，而不是留了一个空占位格。
