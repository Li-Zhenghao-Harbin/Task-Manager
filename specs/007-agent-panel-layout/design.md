# 007 设计：Agent 面板与 Prompt 卡片布局

## 1. 现状

```
.shell{display:grid;grid-template-columns:230px minmax(430px,1fr) 350px;height:100vh;overflow:hidden}

.details{background:#fff;border-left:1px solid var(--line);padding:32px 26px;overflow:auto}
```

`.details` 是一个**整体滚动**的普通块元素。其直接子元素依次为
`.details-heading` → `.fields` → 可用情况 `label` → `.cooldown-row` → 重置时间 `label`
→ `.agent-projects-editor` → `.detail-actions`。

```
.agent-project-list{margin-top:10px;max-height:220px;overflow:auto}
```

列表被钉死在 220px；因为父级是普通块，`.detail-actions` 只能紧跟在它后面，
下方空白无法被列表占用。

Prompt 侧：

```
.prompt-grid{display:grid;grid-template-columns:repeat(auto-fill,minmax(360px,1fr));gap:14px}
.prompt-card{...;padding:12px 14px}
.prompt-card-head{display:flex;align-items:center;justify-content:space-between;gap:10px;margin-bottom:9px}
.prompt-card textarea{width:100%;min-height:132px;resize:vertical;...}
```

卡片内是「头部一行（左 Prompt N、右两个按钮）+ 下方整宽 textarea」的竖排结构，
卡片本身被栅格切成 360px 左右的小格子。

## 2. 方案

### 2.1 Agent 详情面板：改成纵向 flex，中间一段吃掉剩余高度

只给 **Agent 详情这一个 aside** 加类名 `agent-details`，其余规则按该类名作用域，避免波及任务详情面板
和新增弹窗：

```css
.details.agent-details{display:flex;flex-direction:column;overflow:hidden}
.details.agent-details>*{flex:0 0 auto}                                /* 顶部字段、删除区不被压缩 */
.details.agent-details .agent-projects-editor{flex:1 1 auto;min-height:0;display:flex;flex-direction:column}
.details.agent-details .agent-project-list{flex:1;min-height:96px;max-height:none;overflow:auto;
  border:1px solid var(--line);border-radius:10px;background:#fbfdff;padding:2px 10px}
.details.agent-details .detail-actions{margin-top:16px;padding-top:14px;border-top:1px solid var(--line)}
```

要点：

- **`overflow:hidden` 是必须的**。`.details` 原本整体 `overflow:auto`，若不在 Agent 面板上关掉，
  外层会出现滚动条而不是让 `flex:1` 生效后的内层滚动，等于白改。
- `.details.agent-details>*{flex:0 0 auto}` 把顶部的字段区锁成自然高度：flex 子项默认 `flex-shrink:1`，
  不锁定的话窗口变矮时会先把「应用名/模型名」等输入框压扁，而不是让列表先缩。
- 列表加 `border + background`：列表区域被放大后会形成大片空白，加个浅色边框让这块区域读作
  「一个列表容器」而不是「面板漏了背景」。
- `min-height:96px` 是兜底：窗口极矮时保证列表仍可用，不会缩成 0。
- **必须补一条 `@media(max-width:1000px){.details.agent-details{display:none}}`。**
  文件里已有 `@media(max-width:1000px){.shell{...}.details{display:none}}`，
  它靠 `.details`（0,1,0）生效；而新规则 `.details.agent-details`（0,2,0）写在文件更靠后的位置，
  特异性更高、位置更后，会**盖掉**那条媒体查询，导致窄屏下 Agent 面板不再隐藏。
  媒体查询不增加特异性，所以只能显式重述一次。

模板改动仅一处——给 aside 加类名：

```html
<aside v-else-if="selectedAgent" class="details agent-details">
```

`.agent-projects-editor` 内部结构（h3 / hint / list）不变，靠 flex 自然分配。
【删除应用】本来就在最后，配合 `flex-direction:column` 即自动贴底，无需 `position:absolute`。

### 2.2 Prompt 卡片：栅格改单列 flex，卡片内部横向两栏

```css
.prompt-grid{display:flex;flex-direction:column;gap:14px}                 /* 一行一条 */
.prompt-card{display:flex;align-items:stretch;gap:16px;...;padding:12px 14px}
.prompt-side{flex:0 0 auto;width:96px;display:flex;flex-direction:column;align-items:stretch;gap:8px}
.prompt-label{font-size:11px;font-weight:700;color:var(--muted);letter-spacing:.06em}
.prompt-copy,.prompt-delete{text-align:center}
.prompt-card textarea{flex:1;min-width:0;width:auto;min-height:132px;resize:vertical;font-size:13px;line-height:1.55}
```

要点：

- 用 `flex-direction:column` 而不是 `grid-template-columns:1fr`：单列 grid 也成立，但 flex 下
  `gap` 语义和 `.prompt-board` 已有的纵向流一致，且少一层隐式轨道。
- `width:96px` 固定左列宽度，`align-items:stretch` 让两个按钮撑满该列宽度 → 【复制】【删除】等宽对齐。
- textarea 必须写 `min-width:0`。flex 子项默认 `min-width:auto`，textarea 有内在宽度，
  不置零会把卡片撑破、把左列挤扁（这也是原规则里的 `width:100%` 必须改成 `width:auto` 的原因，
  否则 flex-basis 会取到卡片内容宽度的 100%）。
- 头部容器 `.prompt-card-head` 与 `.prompt-actions` 不再使用，从 CSS 中删除，避免死规则。

模板改动：

```html
<article v-for="(prompt, index) in prompts" :key="prompt.id" class="prompt-card">
  <div class="prompt-side">
    <span class="prompt-label">Prompt {{ index + 1 }}</span>
    <button type="button" class="prompt-copy" @click="copyPrompt(prompt)">复制</button>
    <button type="button" class="prompt-delete" @click="removePrompt(index)">删除</button>
  </div>
  <textarea v-model="prompt.text" rows="6" placeholder="输入这个任务要用的提示词"></textarea>
</article>
```

按钮的点击处理函数、`v-model` 绑定、`:key` 全部不变——本次是纯样式/结构重排，
prompt 的增删改与即时保存链路零改动。

## 3. 影响面

| 文件 | 改动 |
| --- | --- |
| `frontend/src/App.vue` | aside 加 `agent-details` 类；prompt 卡片模板改横排 |
| `frontend/src/App.css` | 新增 Agent 面板 flex 一组规则；改写 `.prompt-grid`/`.prompt-card`，删除 `.prompt-card-head`/`.prompt-actions` |

后端、绑定、数据结构均不动，因此不需要 `wails generate module`。

## 4. 风险

- **媒体查询被盖**：见 2.1，已显式补偿。
- **窄面板下顶栏过高**：350px 宽面板里「可用情况」那一行带说明文字的 `.cooldown-row` 会折行。
  已按 `MinHeight:680` 估算剩余高度仍 >200px，且列表有 96px 兜底，不会出现删除按钮被裁掉的情况。
- **旧 `.prompt-card textarea` 的 `width:100%` 残留**：若不改会让 flex-basis 取满宽，必须一并覆盖。
