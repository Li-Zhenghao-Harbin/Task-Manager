# 010 · 设计说明

## 1. Agent 列表行

行结构从 5 个 span 收敛为 4 个：

```
▾  应用名-模型名   下次可用时间：2026-09-22 15:00   [可用]
└16px┘ └─ 弹性 ─┘ └──── max-content ────┘        └78px┘
```

- 栅格：`16px minmax(0,1fr) max-content 78px`。
  日期列用 `max-content` 而不是固定宽度——它要容纳「下次可用时间：2026-09-22 15:00」这类不定长文本，
  固定值要么浪费宽度要么被 `white-space:nowrap` 顶出格子。（011 起该列只在应用不可用时渲染，
  可用行改用 `.agent-row.available` 的 3 列栅格。）
- 窄屏断点 `@media(max-width:1100px)` 原本就把日期列 `display:none`，此时只剩 3 个子元素，
  栅格同步改为 `16px minmax(0,1fr) 78px`。**列数必须跟着子元素数量走**，
  否则多出的列会让最后一个元素换行到第二排。

## 2. 面板字段的间距

面板里 `<label>文字<input></label>` 这种写法，输入框默认是 `inline-block`，
所以只要没被 `.fields` 的 `width:100%` 撑到下一行，它就会紧贴文字右侧。
「应用名 / 模型名」在 `.fields` 栅格里所以自带换行，而这个日期框是独立 label，缺了这条：

```css
.agent-details input[type=datetime-local]{display:block;width:100%;margin-top:8px}
```

- 用 `display:block` 而不是靠 `width:100%` 隐式换行，语义明确、不受父容器宽度影响。
- 作用域限 `.agent-details`，不碰弹窗里的 `setting-row`（那一行是左右布局，输入框在右侧，不需要换行）。

## 3. 导航对齐

`.nav-sub{padding-left:22px}` 是 002 为了让子项看起来「缩进一级」加的。
但「Agent应用」在本项目里是 Agent 分组的**唯一**子项，孤零零缩进反而让它的左边缘与其他同级项错开，
用户读到的就是「没对齐」。直接删掉该规则，让它继承 `.nav-item` 的 `padding:11px 10px`。

`.nav-sub` 这个类名保留在 DOM 上：它是「Agent 分组子项」的语义钩子，界面验证探针也依赖
`.nav-item.nav-sub` 定位这一项，删掉类名会连带弄坏验证脚本。

## 影响面

- 后端零改动，不需要 `wails generate module`。
- 涉及文件：`frontend/src/App.vue`（1 行行模板 + 2 处 label 文案 + 删除 1 个函数）、
  `frontend/src/App.css`（栅格列数 ×2、删 1 条规则、删 1 条规则、加 1 条规则）。
- 删除的 CSS：`.agent-projects`（对应的 span 已不存在）、`.nav-sub`。
- 删除的 TS：`projectNames()`（唯一调用点就是被删掉的项目名列；`strict` 但未开 `noUnusedLocals`，
  编译器不会报，但留着是死代码）。
