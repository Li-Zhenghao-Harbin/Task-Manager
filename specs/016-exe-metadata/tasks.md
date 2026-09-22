# 016 任务清单

- [x] `wails.json`：删除 `author` 块（`name` / `email`）
- [x] `wails.json`：新增 `info` 块（`companyName` / `productName` / `productVersion` / `copyright` / `comments`）
- [x] `build/windows/info.json`：`fixed` 段补 `product_version`（原缺，导致固定块产品版本恒为 `0.0.0.0`）
- [x] `build/windows/info.json`：`info` 段补 `FileVersion` / `InternalName` / `OriginalFilename`
- [x] 溯源确认 `author` 无用：`github.com/wailsapp/wails/v2@v2.12.0` 全仓库只有 `cmd/wails/init.go`
      与 `pkg/templates/*` 引用它，`pkg/commands/build/` 与 `pkg/buildassets/` 零引用
- [x] 溯源确认 `info.json` schema：`github.com/tc-hib/winres@v0.3.1/version/json.go`
- [x] `wails build -o TaskManager-verify.exe` 出独立产物（不动正在运行的 `Task Manager.exe`），构建通过
- [x] 解析 PE `VS_VERSION_INFO` 实测新旧对比：9 项字段符合验收标准，中文与 `©` 写入正常
- [x] 确认 `Comments` 为空（旧值 `Built using Wails (https://wails.io)` 消失）
- [x] 确认构建后项目根目录无 `Task_Manager-res.syso` 残留，`.gitignore` 无需改动
- [x] 删除验证用产物 `build/bin/TaskManager-verify.exe`
- [x] 规格：`specs/016-exe-metadata/` 三件套（本轮无界面改动，无截图）
- [x] `outline.md`：新增「9. 构建配置与 exe 元数据」（配置分工 / info 块取值表 / author 说明）
- [x] 项目记忆（当日日志 + `MEMORY.md`「exe 元数据与构建配置（016 起）」）
- [x] 新建用户级技能 `wails-exe-metadata`（含 `scripts/pe_version_info.py`，把 `.rsrc` 节归属判据与空值误配对两个坑写进逻辑）
- [x] `go test ./...`（33 项）通过、`gofmt -l` 无输出
