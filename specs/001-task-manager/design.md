# 技术设计

## 架构

React 前端通过 Wails 绑定调用 Go `App` 服务。Go 将任务集合读写到 Wails 的用户数据目录内 `tasks.json`，写入采用临时文件再替换的方式，降低中断时损坏数据的风险。

```text
React UI → Wails binding → App service → JSON repository → tasks.json
```

## 数据结构

`Task` 使用 UUID 标识，字段为：名称、层级、父任务、下发人、发布日期、截止日期、类型、状态、产出物、总结及创建/更新时间。日期以 `YYYY-MM-DD` 文本保存，保证前端日期控件直接可用。

## 接口

- `ListTasks() []Task`
- `CreateTask(TaskInput) (Task, error)`
- `UpdateTask(Task) (Task, error)`
- `UpdateTaskStatus(id, status) (Task, error)`

所有写接口都会验证字段、日期及主支线关联规则。
