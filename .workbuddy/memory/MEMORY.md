# 项目长期记忆：TaskManager

## 技术栈事实（以代码为准，不要照抄 outline.md）

- Wails v2 + Go 1.26（后端单文件 `app.go`）+ **Vue 3 `<script setup>`**（`frontend/src/App.vue`）。
  `outline.md` / `setup.md` 里写的「React + TypeScript」是早期规划，**实际是 Vue**，不要被误导。
- 数据持久化：`DataStore` 结构体（`units` / `defaultUnitId` / `tasks` / `agents` / `settings`）序列化到可执行文件同目录的 `task-manager-data.json`，写盘走「临时文件 + rename」。无 SQLite。
- 前端路由是单一 `section` ref（`'tasks' | 'output' | 'agent' | 'units'`），**没有 vue-router**。
- `App.vue` 与 `App.css` 刻意保持紧凑：模板大量压成超长单行。改样式前先确认是否已有同类 class，追加到 `App.css` 末尾即可。
- 样式变量统一用 `:root` 里的 `--blue / --blue-50 / --line / --ink / --muted / --page-background`，配色为蓝白主题。

## 约定

- SDD：规格放 `specs/NNN-<slug>/`，每号三件套 `spec.md`（做什么/不做什么/验收标准）、`design.md`、`tasks.md`。
- 领域词汇：`TaskType = code | document | support`（界面文案分别是 代码 / 文档 / 实施）。
- Agent 模块相关规则见 `specs/002-agent-module/spec.md`。

## 构建与验证

- 改了 Go 方法签名后必须重建绑定，否则前端类型报错：`wails generate module`。
- 日常验证顺序：`go test ./...` → `npm --prefix frontend run build`（含 vue-tsc 类型检查）→ `wails build`。
