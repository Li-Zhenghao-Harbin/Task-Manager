# 016 · 设计

## 1. 数据流：值是怎么变成 exe 属性的

```
wails.json
  "info": { companyName, productName, productVersion, copyright, comments }
        │
        │  ① Wails 解析 wails.json → internal/project.Project.Info
        │     （internal/project/project.go:225 的 Info 结构体，setDefaults 补默认值）
        ▼
build/windows/info.json        ← 带 {{.Info.*}} 占位符的模板（仓库里就有这个文件）
        │
        │  ② buildassets.ReadFileWithProjectData(projectData, "windows/info.json")
        │     用 text/template 把 .Info.* / .Name / .OutputFilename 渲染成字面值
        ▼
VS_VERSION_INFO 资源
        │
        │  ③ winres.ResourceSet.SetVersionInfo() 序列化成 PE 资源
        │     写成项目根目录的 Task_Manager-res.syso（名字 = 项目名，空格替下划线）
        ▼
go build 把 syso 链进 exe
        ▼
右键 exe →「属性 → 详细信息」
```

关键点：**`build/windows/info.json` 本身就是那个「设置 exe 详细信息的 JSON」**，
它从 Wails 脚手架生成起就存在，此前一直是默认模板、且因为 `wails.json` 里没有 `info` 块，
所有值都走了 Wails 的默认值 —— 其中 `copyright` 的默认值是占位符字面量 `Copyright.........`。

## 2. 为什么「值」放 `wails.json`，而不是直接写死在 info.json

两种做法都能生效：

| 做法 | 优点 | 缺点 |
| --- | --- | --- |
| **值写 `wails.json` 的 `info` 块**（本轮采用） | 一处生效 Windows + macOS + NSIS；`info.json` 保持模板原样 | 多一层间接，需知道模板在哪 |
| 直接把 `{{...}}` 换成字面值写在 `info.json` | 直观，改一处即所见 | `build/darwin/Info.plist` 是另一份文件，macOS 要各维护一套，容易漂移 |

选前者。`build/windows/info.json` 仍然保留占位符形态，只把**能映射的键补全**。

## 3. 关于 `wails.json` 的 `author`：它对构建无用

在 `github.com/wailsapp/wails/v2@v2.12.0` 全仓库检索 `Author`：

- `internal/project/project.go:220` 定义了 `type Author struct { Name, Email }`，
  `Project.Author` 字段**没有 json tag**（靠字段名大小写不敏感匹配 `"author"`），所以它**会被解析**。
- 唯一的赋值处是**脚手架**：`cmd/wails/init.go:106 findAuthorDetails()` 读 git 配置填进去，
  再由 `pkg/templates/templates.go` 渲染 `templates/*/wails.tmpl.json` 里的
  `"author": { "name": "{{.AuthorName}}", "email": "{{.AuthorEmail}}" }`。
- `pkg/commands/build/`（构建链路）与 `pkg/buildassets/`（模板注入）里**没有任何一处读 `Author`**；
  连 `wails doctor` / `wails generate module` 也不读它。

结论：`author` 是**只写不读的元数据**，对 exe 属性、绑定生成、打包、安装包全无影响。
删掉后 `wails build` 实测通过（本轮验收标准 1）。
如果哪天想把它用起来（例如填 `companyName` / `copyright`），得自己写脚本读它，Wails 不会代劳。

> 附带影响：删掉后 `Li-Zhenghao-Harbin` / `dream7c@126.com` 这两个值在本仓库里就没有落脚点了
> （git 历史仍在）。若想保留，写进 `info.companyName` / `info.copyright` 即可。

## 4. `info.json` 支持哪些字段

Schema 来自 `github.com/tc-hib/winres@v0.3.1/version/json.go`（Wails 用它编译版本资源）：

```jsonc
{
  "fixed": {
    "file_version":    "a.b.c.d",   // 固定版本块，数值型；不足 4 段时 Sscanf 补 0
    "product_version": "a.b.c.d",
    "flags":           "Debug,Prerelease,Patched,PrivateBuild,SpecialBuild",
    "type":            "app" | "dll",
    "timestamp":       "<RFC3339>"
  },
  "info": {
    "<HEX LCID>": {                 // "0000" = 语言无关（中性）；winres 落盘时按默认语言表写出
      "Comments":         "…",
      "CompanyName":      "…",
      "FileDescription":  "…",
      "FileVersion":      "…",
      "InternalName":     "…",
      "LegalCopyright":   "…",
      "LegalTrademarks":  "…",
      "OriginalFilename": "…",
      "PrivateBuild":     "…",
      "ProductName":      "…",
      "ProductVersion":   "…",
      "SpecialBuild":     "…"
    }
  }
}
```

`info` 段里的键**按字母序写出**，与 JSON 里的书写顺序无关（实例里 `Comments` 排最前）。

## 5. 值 → 属性页显示名对照

以中文 Windows 的「属性 → 详细信息」为例（各版本措辞略有差异）：

| PE 键 | 属性页显示 | 本轮来源 |
| --- | --- | --- |
| `ProductName` | 产品名称 | `info.productName` |
| `FileDescription` | 文件说明 | `info.productName` |
| `FileVersion` | 文件版本 | `info.productVersion` |
| `ProductVersion` | 产品版本 | `info.productVersion` |
| `CompanyName` | 公司名称 | `info.companyName` |
| `LegalCopyright` | 版权 | `info.copyright` |
| `Comments` | 备注 | `info.comments`（本轮置空） |
| `InternalName` | 内部名称 | `{{.Name}}` = `Task Manager` |
| `OriginalFilename` | 原始文件名 | `{{.OutputFilename}}` = `Task Manager.exe` |

## 6. 验证手法（不依赖 GUI 点击属性页）

生成过程没有界面，截图无从下手，所以改为**直接解析 PE 的 `VS_VERSION_INFO` 资源**：

1. 按 UTF-16LE 在文件里定位所有 `VS_VERSION_INFO` 根键。
   注意：exe 里还内嵌了 WebView2Loader 等自带资源，也有同名根键，
   必须挑**键值对数量 ≥ 5 的那一个**（我们的），否则会读到 Microsoft 的那份。
2. `VS_FIXEDFILEINFO` 结构用签名 `0xFEEF04BD` 定位，偏移 +8 起是
   `dwFileVersionMS/LS` 与 `dwProductVersionMS/LS`（各 16 位拆成 4 段）。
3. 字符串块按 `\x00` 切分后「键 → 下一个非空串」配对。
   **值为空时会误配到下一个键**（本轮 `Comments` 就踩了），所以空值要单独判定。

另用 `wails build -o TaskManager-verify.exe` 出独立产物做验证：
`-o` 只覆盖 `buildOptions.OutputFile`，不碰 `ProjectData.OutputFilename`，
因此正在运行的 `Task Manager.exe` 不会被锁、`{{.OutputFilename}}` 也仍渲染成正确的原文件名。
