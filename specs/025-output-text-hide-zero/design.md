# 025 设计：在唯一的拼接点过滤，而不是在模板里做条件渲染

## 定位

产出摘要只有一个生成点，也只有一个消费点：

```
App.vue:414  function taskOutputText(task: Task)         ← 唯一生成点
App.vue:1225 <em>{{ taskOutputText(task) }}</em>          ← 唯一消费点（产出概览 · 对应任务）
```

「全部产出」那一档不渲染 `.output-list`（走饼图卡片），所以这次改动的作用范围就是
【代码】/【文档】/【实施】三个筛选下的任务列表。

## 改动

```diff
-function taskOutputText(task: Task) {
-  return typeKeys[task.taskType].map((key) => `${outputLabels[key]} ${task.output[key] || 0}`).join(' | ')
-}
+function taskOutputText(task: Task) {
+  return typeKeys[task.taskType]
+    .filter((key) => (task.output[key] || 0) !== 0)
+    .map((key) => `${outputLabels[key]} ${task.output[key] || 0}`)
+    .join(' | ')
+}
```

```diff
-  ><em>{{ taskOutputText(task) }}</em>
+  ><em v-if="taskOutputText(task)">{{ taskOutputText(task) }}</em>
```

## 为什么这样切

**放在函数里而不是模板里**：模板不需要知道「哪个字段算产出、怎么算零」。若在模板里写
`v-if="task.output.newFeatures"`，等于把字段语义复制一份到模板，三个类型九个字段都要维护，
还会漏掉「分隔符也得跟着消失」这件事（这正是用户抱怨的形态：不是字段没藏，而是藏了字段留下管道符）。

**`!== 0` 而不是 `> 0`**：后端已把负数 clamp 到 0（`UpdateTask` 里 `if task.Output.X < 0 { = 0 }`），
正常情况下不会出现负数。用 `!== 0` 的语义是「值为 0 就不显示」，与需求字面一致；万一导入了脏数据里的负数，
显示出来反而是提示，比静默吞掉更安全。

**全零返回空串 + `v-if`**：空串在模板里会让 `<em>` 变成一个空盒子。虽然 `.output-task` 是
`justify-content:space-between`、空 `em` 宽度为 0 看不出差别，但留一个空元素迟早会被人当成
「为什么这里有个空 em」的谜题。所以直接不渲染。

`taskOutputText` 在模板里被调用两次（`v-if` 一次、插值一次），22 条任务量级下没有性能问题；
真嫌重复可以改成 computed 缓存的 Map，属于过度设计。

## 影响面的自我约束

右侧详情面板的产出区用的是另一套模板：

```html
<div v-if="selected.taskType !== 'support'" class="quantity-grid">
  <template v-for="key in typeKeys[selected.taskType]" :key="key">
    <label>{{ outputLabels[key] }}<OutputStepper ... /></label>
  </template>
</div>
```

**它没有走 `taskOutputText`，所以天然不受影响** —— 这是运气好，不是设计保证。为了防止以后有人
「统一一下」，验证里加了一条显式判据：打开那个「部分字段为 0」的任务，详情面板必须仍是 3 个数量字段。

## 验证方法

判据不是「有没有出现 0」（容易被「0 被别的字符代替」晃过去），而是**按真值逐行比对期望文本**：

```js
// 期望值从 window.__DATA 现算，不写死
KEYS_OF[type].filter(k => valOf(task, k) !== 0)
             .map(k => LABEL_OF[k] + ' ' + valOf(task, k))
             .join(' | ')
```

漏项、错序、多一个空格、残留 `|`，任何一种都会让比对失败。

对照组（`?old=1`）用同一份期望值，但把按「不过滤」算出的全量文本写回 DOM —— 旧版连全零任务也渲染
`<em>`，所以还要给缺失的行补一个，才能完整复现旧输出。

## 一个探针时序坑（已回灌技能）

点任务行后不能只等 `.quantity-grid` 出现：`waitFor` 的第一次检查是**同步**的，而 Vue 改完响应式数据要等
一拍才更新 DOM，于是立刻命中的是**点击前那块面板**，读到的是上一个任务的字段 —— 表现为
「点了代码任务，面板却报新增文档/修改文档/Skills」。修法是把「面板里显示的就是我刚点的任务名」
写进等待条件。这类断言最危险的地方在于它**看起来是绿的**，直到你把断言文字当真。
