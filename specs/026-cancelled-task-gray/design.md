# 026 设计：状态色与「完成」绿底同构，只补第三档

## 定位

列表行只有一处渲染，但有三份模板（主任务 / 子任务 / 孤儿子任务），三处都得挂同一个类：

```
App.vue:1096  class="task-row"               :class="{ completed: …, context: … }"   ← 主任务
App.vue:1115  class="task-row child"         :class="{ completed: … }"               ← 子任务
App.vue:1134  class="task-row child orphan-child"  :class="{ completed: … }"          ← 孤儿
```

样式侧唯一的先例是：

```css
.task-row.completed{background:#e9f8f0}
.task-row.completed:hover,.task-row.completed.selected{background:#d9f3e4}
```

`.completed` 是**双职**的：它既是 `.badge.completed` 的徽标色（`#e9f8f0` 底 / `#18794e` 字），
又是 `.task-row.completed` 的行底色。这次照抄这个结构，让 `.cancelled` 也双职。

## 改动

```diff
-:class="{ completed: group.parent.status === 'completed', context: group.context }"
+:class="{ completed: group.parent.status === 'completed', cancelled: group.parent.status === 'cancelled', context: group.context }"
```

```diff
-:class="{ completed: child.status === 'completed' }"
+:class="{ completed: child.status === 'completed', cancelled: child.status === 'cancelled' }"
```

（子任务与孤儿子任务两处同样处理。）

```diff
 .task-row.completed{background:#e9f8f0}.task-row.completed:hover,.task-row.completed.selected{background:#d9f3e4}
+.task-row.cancelled{background:#f2f4f7;color:var(--muted)}
+.task-row.cancelled:hover,.task-row.cancelled.selected{background:#e2e4e7}
+.task-row.cancelled .badge.cancelled{background:#e3e8ef}
```

## 为什么这样切

**行底色而不是只改文字色**：只把文字转灰（像 `.task-row.context` 那样）的话，一屏 20 行里
灰字与正文的对比度差异很小，扫视时仍然要靠读徽标。整行铺灰底是「作废」最省辨认成本的表达，
也与已有的「完成」绿底形成同一套语法：**底色 = 状态**。

**灰底选中态不回到蓝色**：`.task-row:hover,.task-row.selected{background:#f2f7ff}` 会把灰行冲成蓝白，
鼠标一划过就「复活」了。所以照 `.completed` 的做法补一条更高优先级的两态覆盖，
并且选的 `#e2e4e7` 是同一灰阶的深一档（亮度差 ≈ 5%），不引入新色相。

**`color:var(--muted)` 放在行上而不是 `.task-name` 上**：名次、日期、徽标都在行内，
`.task-name` 是 `inherit`、`.date` 自己已经是 `var(--muted)`。改行上的 `color` 一次覆盖到位，
也顺手把 `◆`/`◇` 层级符号一起压灰 —— 那一列本来就跟状态无关，但压灰后整行的「作废感」更完整。

**徽标要单独压深**：徽标自己的 `.cancelled` 底是 `#eef1f5`，行底 `#f2f4f7`，两者差 3 个色阶，
在 100% 缩放下几乎看不出边界。压到 `#e3e8ef` 后差值拉大到 15，徽标轮廓重新出现。
这条不做不会「错」，但会像「样式没写全」。

**不碰 `.task-row.context`**：`.task-row.context{color:var(--muted)}` 与新的
`.task-row.cancelled{color:var(--muted)}` 同值，叠加无副作用；左侧竖条 `#dbe6f5` 是层级线索，
一个已取消的上下文任务仍然需要那条竖条来说明「它为什么在列表里」。

## 被否掉的方案

| 方案 | 否决原因 |
| --- | --- |
| 整行 `opacity:.55` | 会连带把徽标、边框一起淡掉，`opacity` 还会建立新的层叠上下文，影响 021–023 的滚动/贴合判据（`scrollHeight` 不变，但 `position:sticky` 的表现会变） |
| `text-decoration:line-through` 删除线 | 中文任务名带删除线可读性差，且「已取消」不等于「已完成」，删除线更像「划掉不要了」 |
| `filter:grayscale(1)` | 对 `.task-row` 用 filter 会强制合成层，滚动时容易出现模糊；且它会把蓝色徽标变成灰，与 `.badge.cancelled` 的既有色冲突 |
| 在 `.badge` 上做更强的灰 | 用户看的是整行，徽标已经是灰的了，加强它等于没改 |
