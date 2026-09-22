# 017 设计：版本号从构建配置进入运行时

## 核心问题

`wails.json` 是**构建期**配置，`wails build` 读完它就走，运行时进程手里没有这个文件
（exe 可以拷到任何目录单独运行，旁边不会有 `wails.json`）。所以「运行时读取 json」必须先解决
「值怎么进到二进制里」的问题。

## 方案：编译期内嵌 + 单例解析 + 绑定暴露

```
wails.json ──go:embed──> wailsJSON []byte ──json.Unmarshal(一次)──> productVersion string
                                                                          │
                                                    GetAppVersion() ──────┘
                                                            │
                                         前端 onMounted 调用 → appVersion ref → 模板
```

### 1. 内嵌（`version.go`）

```go
//go:embed wails.json
var wailsJSON []byte
```

`wails.json` 与 `app.go` 同目录（`package main`），所以 embed 路径合法。
embed 在**编译期**把文件字节写进二进制，exe 独立分发时不需要 `wails.json` 在旁边。

### 2. 只解析一次

用 `sync.OnceValue` 而不是每次调用都 `Unmarshal`，也不在 `init()` 里做（`init` 里失败无从反馈）：

```go
var productVersion = sync.OnceValue(func() string { ... })
```

解析失败或字段缺失时返回空串 —— 由前端兜底显示 `—`，不 panic、不伪造一个默认版本号。
**不设默认值**是刻意的：宁可界面出现 `—` 让人发现配置丢了，也不要用 `1.0.0` 掩盖问题。

### 3. 为什么不用另外两种做法

| 做法 | 否决理由 |
| --- | --- |
| 前端 `vite.config.ts` 里读 `../wails.json` 用 `define` 注入 | 值会被打进 JS 包，`wails build` 之外的单跑 `npm run build` 拿到的是快照；且版本号出现在两个构建系统里，仍可能不一致 |
| 运行时用 `os.ReadFile("wails.json")` | exe 分发时旁边没有这个文件，直接失效 |
| `-ldflags -X main.version=...` | 需要额外维护一条构建命令，且 `wails build` 默认不传；embed 零配置、零额外命令 |

### 4. 前端接入

```ts
const appVersion = ref('')

onMounted(() => {
  try {
    GetAppVersion().then((version) => (appVersion.value = version)).catch(() => {})
  } catch {
    /* 无 Wails 运行时的预览环境没有绑定 */
  }
})
```

```html
<div><small>版本号</small><strong>{{ appVersion || '—' }}</strong></div>
```

- `try/catch` 是必需的：预览环境（`wails-vue-ui-verify` 那种纯静态页）里 `GetAppVersion` 可能是
  `undefined`，调用会**同步**抛 `TypeError`，`.catch()` 接不住；`.catch()` 只负责接住 Promise 拒绝。
- 降级显示 `—` 而不是空字符串，避免「版本号」那一格看上去像掉了数据。
- 弹窗是按需渲染的（`v-if="showAbout"`），但值在 `onMounted` 就取好了，所以打开弹窗不会有
  先显示 `—` 再跳变成数字的闪烁。

## 与 016 的关系

两者读**同一个字段** `wails.json → info.productVersion`，只是走两条通道：

- exe 属性：`wails build` 把 `info` 渲染进 `build/windows/info.json` 模板 → winres 编译成 PE 资源。
- 关于弹窗：`//go:embed wails.json` → 运行时读 → 绑定 → 界面。

因此「改一处、两处同步」成立，这也是本轮的目的。注意 exe 属性页的**文件版本**会被补成 4 段
（`1.0.0` → `1.0.0.0`），那是 Windows 版本资源的结构要求；关于弹窗显示的是 json 里的原始三段式。

## 测试

- `TestGetAppVersionMatchesWailsJSON`：把返回值与**磁盘上**的 `wails.json` 逐字比对 —— 这条测试
  同时证明了 embed 拿到的不是陈旧副本。
- `TestGetAppVersionUsesThreeSegments`：正则 `^\d+\.\d+\.\d+$`，钉住 `x.x.x` 的约定。
