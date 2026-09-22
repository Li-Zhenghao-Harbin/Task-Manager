# 013 · 设计说明

## 1. 入口：复用 `header > .actions`，不自造一套

「全部任务」与「Prompt 管理」两个视图的页头右侧都是同一个结构：

```html
<div class="actions"><button class="primary">＋ xxx</button></div>
```

单位管理页照抄这一份结构，于是对齐是**结构带来的**，不是靠调数值凑出来的：

- `.workspace` 的右内边距（`padding:32px 34px`）决定按钮右边缘 —— 实测两个按钮到内容区右边缘
  的距离都是 `0.00px`；
- `.primary` 决定按钮外观 —— 实测高度都是 `41px`；
- `header{align-items:center}` 决定垂直位置 —— 实测相对 header 顶边的偏移 21.5px vs 22.5px
  （1px 差来自「全部任务」页头左侧多了一个视图切换控件，header 略高）。

**没有为此新增任何 CSS 规则。**

## 2. 弹窗：窄弹窗 + `wide-field` 单字段

```html
<div v-if="showUnitCreate" class="modal-backdrop">
  <form class="modal unit-modal" @submit.prevent="addUnit">
    <div class="modal-title"><h2>新增单位</h2><button class="close" …>×</button></div>
    <div class="form-grid">
      <label class="wide-field">单位名<input v-model="unitDraft" placeholder="例如 数据中台部" required /></label>
    </div>
    <div class="actions">
      <button type="button" …>取消</button><button class="primary">创建单位</button>
    </div>
  </form>
</div>
```

- `.form-grid` 是两列栅格，单字段用 `wide-field`（`grid-column:span 2`）跨满两列，
  与任务弹窗里「关联 Agent」的写法一致 —— 这样输入框仍然是「标签在上、输入框在下、撑满整宽」，
  不需要为单字段另写样式。
- 只需一条新规则把宽度收窄：`.unit-modal{width:min(430px,100%)}`。
  默认 `.modal` 是 `min(620px,100%)`，任务弹窗（560）、Agent 弹窗（560）、设置（500）、
  确认框（390）各有自己的宽度，单字段的 430 落在这个区间里，符合既有梯度。

## 3. 为什么删掉 `.unit-list` 的 `margin-top` 与 `border-top`

删掉输入行后，`.unit-list` 成了面板里的第一个元素：

- `border-top:1px solid var(--line)` 原本是「表单与列表之间的分隔线」。上方没有东西可分隔时，
  它会在卡片顶部 22px 处横一条线，看起来像漏了一行。
- `.units-workspace .unit-list{margin-top:18px}` 原本是「表单下方的间距」，
  留在那里就是在面板 `padding:22px` 之外再多出 18px 空白。

两条一并去掉后，首行顶边距面板顶边正好 22px，由探针钉住（实测 23.00px = 1px 边框 + 22px 内边距）。

## 4. 命名：`newUnit` → `unitDraft`

弹窗里的输入框绑定的不再是「临时的新单位名」，而是**弹窗草稿**，与 `draft`（任务）、
`agentDraft`（Agent）是同一个角色，所以改名为 `unitDraft`。`addUnit()` 保持原名与职责
（提交 + 关弹窗 + 刷新），只是提交的源从 `newUnit` 换成 `unitDraft`。

## 5. 保持「提交即刷新列表」的既有节奏

`addUnit` 依然是「调接口 → 关弹窗 → `refresh()`」：列表数据来自后端，
不在前端做乐观插入，避免与「单位被重命名/删除后 `unitFilter` 要回落到全部单位」那条兜底逻辑打架。

## 6. 验证方式（无头探针）

`wails-vue-ui-verify` 那套：mock 绑定 + Chrome 无头 + 断言横幅。本轮额外踩到一个坑并已记入技能：

> **跨视图做几何对比，必须在离开该视图之前把 rect 换算成标量。**
> 先量「全部任务」的 `＋ 新建任务`，点左侧导航切到单位管理之后再拿那个按钮节点量 —— 节点已经被
> Vue 从 DOM 上摘掉，`getBoundingClientRect()` 返回零矩形，于是「右边缘差 1834px、高度差 41px」
> 这类**探针 bug** 会被当成应用回归。改成在测量当场算好 `inset / height / topInHeader` 三个数，
> 再与新页面的同名字段比。**判据里不要长期持有跨渲染的元素引用。**

覆盖到的断言：面板 15 项、弹窗结构 22 项、完整交互流程 32 项（宽屏与 1000px 窄屏各跑一次全绿）、
空态 5 项。
