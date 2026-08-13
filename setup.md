# 工作任务管理器：Windows 环境安装指南

本项目计划使用 Go + Wails（React + TypeScript 前端）构建 Windows 桌面程序。以下步骤仅准备开发与打包环境；本阶段不初始化或编写应用代码。

## 1. 系统前提

- Windows 10（1803 及以上）或 Windows 11，64 位。
- 已安装或允许安装 Microsoft Edge WebView2 Runtime。多数 Windows 11 系统已自带；Wails 桌面窗口依赖它渲染界面。
- PowerShell 7（推荐）或 Windows PowerShell 5.1。

## 2. 安装必需软件

### 2.1 Go

安装当前稳定版 Go（建议 1.22 或更高）。安装完成后重新打开 PowerShell，确认：

```powershell
go version
```

### 2.2 Node.js

安装 Node.js LTS 版本（建议 20 或更高），其会包含 npm。确认：

```powershell
node --version
npm --version
```

### 2.3 Visual Studio Build Tools（Windows 编译链）

安装 **Visual Studio 2022 Build Tools**，在安装器中勾选：

- Desktop development with C++（使用 C++ 的桌面开发）
- MSVC v143 或更新的 C++ x64/x86 build tools
- Windows 10/11 SDK

安装后新开 PowerShell。可用下列命令确认已存在编译器（应输出路径）：

```powershell
Get-Command cl -ErrorAction SilentlyContinue
```

若普通 PowerShell 找不到 `cl`，可从“Developer PowerShell for VS 2022”运行构建命令；后续也可再配置普通终端环境。

### 2.4 WebView2 Runtime

若运行 Wails 程序时提示缺少 WebView2，请安装 Microsoft Edge WebView2 Evergreen Runtime，然后重试。无需安装完整 Edge 浏览器开发工具。

## 3. 安装 Wails CLI

在 PowerShell 运行：

```powershell
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

确保 Go 的 bin 目录已位于 `PATH`。通常为：

```text
%USERPROFILE%\go\bin
```

重新打开终端并验证：

```powershell
wails version
wails doctor
```

`wails doctor` 应不再报告 Go、Node、WebView2 或 Windows 编译工具缺失。

## 4. 推荐开发工具

- Visual Studio Code
- VS Code 扩展：Go、ESLint、Prettier、Tailwind CSS IntelliSense（仅在后续选用 Tailwind 时需要）
- Git（用于版本管理）

确认 Git：

```powershell
git --version
```

## 5. 项目运行与打包

在项目根目录（即含有 `wails.json` 的目录）执行：

```powershell
wails dev
wails build
```

生成的 Windows 可执行文件位于 `build\bin\Task Manager.exe`。

## 6. 本地数据说明

第一版将任务数据保存为应用数据目录中的 JSON 文件，不需要单独安装数据库服务。后续若数据量或查询需求增长，可将 repository 层替换为 SQLite，而不改变前端界面和业务接口。

## 7. 环境检查清单

在开始编码前，以下命令均应成功运行：

```powershell
go version
node --version
npm --version
git --version
wails version
wails doctor
```

如 `wails doctor` 的某一项仍失败，请保留完整输出，后续可据此定位 PATH、WebView2 或 C++ Build Tools 的配置问题。
