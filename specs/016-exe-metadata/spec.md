# 016 · exe 属性元数据：用一个 JSON 配置版本号、版权等信息

> 右键 exe →「属性 → 详细信息」里显示的内容，全部来自 `build/windows/info.json` 渲染出的 PE 版本资源。
> 本轮把「值」统一提到 `wails.json` 的 `info` 块，并删掉从来没被构建读取过的 `author`。

## 做什么

1. **在 `wails.json` 新增 `info` 块**，作为 exe 元数据的唯一事实来源：

   | 键 | 本轮取值 |
   | --- | --- |
   | `companyName` | `TODO：公司名称`（占位，待用户填写） |
   | `productName` | `Task Manager` |
   | `productVersion` | `1.0.0` |
   | `copyright` | `TODO：版权信息，例如 © 2026 某某公司. All rights reserved.`（占位，待用户填写） |
   | `comments` | `""`（留空，去掉 Wails 默认的 `Built using Wails (https://wails.io)`） |

2. **补全 `build/windows/info.json` 模板**，把上述值映射到更完整的 PE 字段：
   - `fixed` 段补 `product_version`。原来只写了 `file_version`，导致固定版本块里的
     **产品版本一直是 `0.0.0.0`**（与字符串块里的 `1.0.0` 自相矛盾）。
   - `info` 段补三个原本缺失的键：`FileVersion`（Explorer「文件版本」一栏原先是空的）、
     `InternalName`（`{{.Name}}`）、`OriginalFilename`（`{{.OutputFilename}}`）。

3. **删除 `wails.json` 的 `author` 块**（`name` / `email`）——它不参与构建，见下「不做什么」。

## 不做什么

- **不改任何界面/代码**：本轮只动构建配置，`app.go`、`App.vue`、`App.css` 一行不改。
- 不改 `build/windows/icon.ico`（exe 图标）。
- **不引入版本号自动派生**：不读 git tag、不拼构建时间戳。版本号就是一个手写字符串。
- 不改 `build/windows/installer/project.nsi` 与 `wails_tools.nsh`：NSIS 安装包从同一组
  `Info.*` 取值（`wails_tools.nsh:17`），会跟着一起变；脚本里已有 `"${INFO_PRODUCTVERSION}.0"`
  兜底补第 4 段。
- 不改 `build/darwin/Info.plist`：它同样引用 `.Info.*`，值自动同步。
- 不改 `.gitignore`：`wails build` 结束后会自行删除项目根目录的 `Task_Manager-res.syso`
  （`pkg/commands/build/build.go:266-268`），不留残留（本轮已复验，构建后根目录无 syso）。

## 验收标准

以下每条都由「解析 exe 的 PE 版本资源」实测，而非看配置文件推断：

1. **`author` 确实无用**：删除后 `wails build` 依然成功（无缺字段报错、无警告），产物正常。
2. 新 exe 的 `CompanyName` / `LegalCopyright` 等于 `wails.json` 里填的值，
   **中文与 `©` 符号能正确写入**（版本资源是 UTF-16，实测可读回原串）。
3. `Comments` 为空 —— 旧 exe 里的 `Built using Wails (https://wails.io)` 消失。
4. `FileVersion` / `InternalName` / `OriginalFilename` 三个键从「缺失」变为有值，
   分别为 `1.0.0` / `Task Manager` / `Task Manager.exe`。
5. 固定版本块 `fixed.ProductVersion` 从 `0.0.0.0` 变为 `1.0.0.0`，与 `fixed.FileVersion` 一致。
6. `productName` / `productVersion` 保持 `Task Manager` / `1.0.0` 不变（回归项）。

## 交付后待办

- 把 `wails.json` 里两处 `TODO：…` 换成真实内容，再 `wails build` 一次即可生效。
- **版本号写法**：建议 3 段（`1.0.0`）或 4 段（`1.0.0.0`）。
  写 2 段（`1.0`）会让 NSIS 的 `VIProductVersion` 凑不满 4 段而构建安装包失败。
