# 027 设计：把「已完成」这个代理指标换成真正要表达的「有产出」

## 定位

```
App.vue:176  const typeKeys: Record<TaskType, OutputKey[]> = { code: [...], document: [...], support: [...] }
App.vue:293  const outputTasks = computed(() => tasks.value.filter(task => task.status === 'completed' && ...))
App.vue:416  function taskOutputText(task) { return typeKeys[task.taskType].filter(key => (task.output[key] || 0) !== 0)... }
```

`outputTasks` 只在 `outputFilter !== 'all'` 时被模板用到（`v-if="outputFilter === 'all'"` 走饼图，`v-else` 才是列表），
所以 `outputFilter.value === 'all' ||` 这一支实际是死分支，保留是为了语义完整。

## 改动

```diff
+// 「有产出」= 该类型对应的产出项里有任意一项非零。判据与 taskOutputText 同源，
+// 所以列表里不会出现「有这一行、摘要却是空的」。
+function hasOutput(task: Task) {
+  return typeKeys[task.taskType].some((key) => (task.output[key] || 0) !== 0)
+}
+
 const outputTasks = computed(() =>
   tasks.value.filter(
-    (task) => task.status === 'completed' && (outputFilter.value === 'all' || task.taskType === outputFilter.value),
+    (task) => hasOutput(task) && (outputFilter.value === 'all' || task.taskType === outputFilter.value),
   ),
 )
```

## 为什么这样切

**`status === 'completed'` 是个代理指标，代理错了。** 当初写它是想表达「产出已经发生了」，
但「完成」是任务的流程状态，「产出」是填在 `task.output` 里的数字，两者可以脱钩：
完成了没记产出（3 条 code），记了产出但还在进行（1 条 code）。用户要的是后者能被看见、前者别占位。

**判据必须挂在 `typeKeys` 上，而不是「所有 7 个 output 键任一非零」。** 一个 `code` 任务即使填了
`newDocuments`，它也不该出现在【代码】页之外的任何地方；而【代码】页的摘要按 `typeKeys.code` 只显示
三个键，若用全量键判据，就会出现「因为填了文档数而进入代码页、但摘要显示空」的行。176 行那套
`typeKeys` 本来就是「类型 → 该类型关心的产出项」的唯一真源，判据跟着它走，行、摘要、卡片三者自动一致。

**抽成函数而不是内联。** 三处（未来若加排序/统计）都要用同一个判据，内联会让「有产出」的定义
散在多处；且对照组（回归探针）需要能一眼看出改的是哪一句。

**不动菜单计数。** 菜单的 `count`（App.vue:329）已经是「该类型全部任务的产出合计」，与新的列表天然一致 ——
这正是这次改动顺带修好的数字错位（改前【代码】菜单 13、列表 9 行且其中 3 行合计 0）。

**不改排序。** `tasks` 数组是后端 `ListTasks` 的顺序（按创建/发布日期），`outputTasks` 直接沿用它。
用户没提排序诉求；自作主张按状态或日期重排会让「刚填完的一条跳到哪去」变得不可预测。

## 被否掉的方案

| 方案 | 否决原因 |
| --- | --- |
| 保留「已完成」并额外并入「进行中且有产出」 | 结果是「已完成无产出」那 3 条仍然在，问题的一半没解决 |
| 用 `taskOutputText(task) !== ''` 判据 | 等价于现在的写法，但把一个「为了显示」的字符串函数当成筛选条件，语义绕；且将来改摘要格式（比如加单位、加粗）会顺带改掉筛选 |
| 把 `result` 文本也算产出 | 与卡片统计（只累加数字）不一致；「写了结论」不等于「记了产出」，见 spec「不做什么」第 1 条 |
| 改用后端返回一个 `hasOutput` 字段 | 纯展示层派生量，过一遍持久化没有收益，还会引入「老数据没有这个字段」的兼容问题 |
