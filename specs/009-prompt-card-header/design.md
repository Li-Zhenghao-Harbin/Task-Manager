# 009 · 设计

## 结构变化

008（两段式）：

```html
<article class="prompt-card">
  <div class="prompt-body">              <!-- row：左列 + 文本框 -->
    <div class="prompt-side">            <!-- 96px 窄列，竖排三行 -->
      <span class="prompt-label">Prompt 1</span>
      <button class="prompt-copy">复制</button>
    </div>
    <textarea />
  </div>
  <div class="prompt-foot">              <!-- 分隔线 + 右对齐删除 -->
    <button class="prompt-delete">删除</button>
  </div>
</article>
```

009（头部一行 + 整行文本框）：

```html
<article class="prompt-card">            <!-- column，gap 10px -->
  <div class="prompt-head">              <!-- row -->
    <span class="prompt-label">Prompt 1</span>
    <button class="prompt-copy">复制</button>
    <button class="prompt-delete">删除</button>   <!-- margin-left:auto 推到最右 -->
  </div>
  <textarea />
</article>
```

DOM 顺序即视觉顺序：`label → 复制 → 删除`。「复制靠左、删除靠右」不靠额外的
spacer 元素，而是给【删除】加 `margin-left:auto` —— 在 flex 行里这会把该元素
右边的剩余空间全部吃掉，等于自动靠右；比 `justify-content:space-between`
少一层包裹，也比插一个空 `div` 干净。

`margin-left:auto` 的第二个好处：它同时把「复制 ↔ 删除」之间的最小间距交给
`gap`（`gap:8px` 时如果两者相邻则至少 8px），窗口再窄也不会贴到一起。
`flex-wrap` 不开：三个元素都很窄（最宽是「删除」两字），不需要换行兜底。

## 样式改动

| 规则 | 改动 |
| --- | --- |
| `.prompt-card` | 保持 `display:flex;flex-direction:column;gap:10px`（形状没变，删掉 `prompt-body` 后它直接管头部与文本框的垂直间距） |
| `.prompt-head` | **新增**：`display:flex;align-items:center;gap:8px` |
| `.prompt-label` | 不变（`font-size:11px;font-weight:700;color:var(--muted);letter-spacing:.06em`） |
| `.prompt-copy` | 不变 |
| `.prompt-delete` | 追加 `margin-left:auto` |
| `.prompt-card textarea` | `flex:1;min-width:0;width:auto` → `width:auto;flex:0 0 auto`（**必须拿下 `flex:1`**：它是从横向 flex 行继承下来的，在 column 容器里 `flex:1` 会让文本框去抢纵向剩余高度，把卡片拉高/压扁；column 子项默认 `align-self:stretch`，宽度本来就撑满，不需要显式设宽） |
| `.prompt-body` / `.prompt-side` / `.prompt-foot` | **删除**（含 `.prompt-foot` 的 `border-top` 分隔线——头部一行化后分隔线失去意义） |

## 为什么删掉而不是留着

`.prompt-body` / `.prompt-side` / `.prompt-foot` 在新结构里没有任何元素引用，
留着只会让后来改样式的人以为还有第三种布局。007 已经有过一次同类清理
（`.prompt-card-head` / `.prompt-actions`），本次沿用同样做法：**模板改完立刻删样式**。

## 影响面

只动 `frontend/src/App.vue`（一行模板）与 `frontend/src/App.css`。
后端、绑定、数据结构、其它面板均不受影响。
