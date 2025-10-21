# WCoreFx - Windows Core Framework for Go

[![Go Version](https://img.shields.io/badge/Go-1.20+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

## 简介

WCoreFx 是一个专为 Go 语言设计的 Windows 系统核心框架库。"W" 代表 Windows，"Core" 对应核心（kernel），"Fx" 代表框架（Framework）。

该库提供了 Go 标准库之外的 Windows 系统 API 封装，帮助开发者更方便地开发基于 Windows 系统的应用程序。

## 功能特性

WCoreFx 库包含以下核心功能模块：

### 🔐 安全与权限 (se)
- 用户权限检查
- 管理员权限验证
- Windows 安全令牌处理

### ⚙️ 进程管理 (ps)
- 进程枚举和信息获取
- 进程树结构分析
- 进程权限和状态查询

### 📁 文件系统 (fs)
- 文件资源信息获取
- 文件版本信息查询
- 文件属性和权限操作

### 🌐 网络管理 (net)
- 网络连接枚举
- TCP/UDP 端点信息
- 网络接口状态查询

### 🖥️ 系统对象 (ob)
- 系统驱动程序管理
- 内核对象操作
- 设备驱动信息

### 📋 注册表 (reg)
- 注册表键值读写
- 注册表项枚举
- 系统配置管理

### 📊 事件管理 (eve)
- Windows 事件日志处理
- 系统事件监控
- 安全事件分析

### 🔧 操作系统 (os)
- 系统信息获取
- 操作系统版本检查
- 系统资源状态

### 🛠️ 通用工具 (comm)
- 通用数据结构
- 错误处理
- 常量定义

## 安装

```bash
go get github.com/kitsch-9527/wcorefx
```

## 快速开始

```go
package main

import (
    "fmt"
    "github.com/kitsch-9527/wcorefx/se"
    "github.com/kitsch-9527/wcorefx/ps"
)

func main() {
    // 检查管理员权限
    isAdmin, err := se.CheckAdmin()
    if err != nil {
        panic(err)
    }
    fmt.Printf("当前用户是否为管理员: %t\n", isAdmin)

    // 枚举进程
    processes, err := ps.EnumProcessMap()
    if err != nil {
        panic(err)
    }
    fmt.Printf("系统进程总数: %d\n", processes.Size())
}
```

## 模块使用示例

### 安全权限检查

```go
import "github.com/kitsch-9527/wcorefx/sec"

// 检查当前用户是否具有管理员权限
isAdmin, err := se.CheckAdmin()
if err != nil {
    log.Fatal(err)
}
```

### 进程管理

```go
import "github.com/kitsch-9527/wcorefx/ps"

// 获取所有进程信息
processMap, err := ps.EnumProcessMap()
if err != nil {
    log.Fatal(err)
}
```

### 文件系统操作

```go
import "github.com/kitsch-9527/wcorefx/fs"

// 获取文件版本信息
info, err := fs.GetFileVersionInfo("C:\\Windows\\System32\\kernel32.dll")
if err != nil {
    log.Fatal(err)
}
```

### 注册表操作

```go
import "github.com/kitsch-9527/wcorefx/reg"

// 读取注册表字符串值
value, err := reg.GetSValue("HKEY_LOCAL_MACHINE\\SOFTWARE\\Microsoft\\Windows\\CurrentVersion", "ProductName")
if err != nil {
    log.Fatal(err)
}
```

## 系统要求

- **操作系统**: Windows 7 及以上版本
- **Go 版本**: 1.20 或更高版本
- **架构**: 支持 amd64 和 386

## 依赖项

- `golang.org/x/sys/windows` - Windows 系统调用包

## 编译标签

本库仅支持 Windows 平台，所有源文件都使用了以下编译标签：

```go
//go:build windows
// +build windows
```

## 贡献

欢迎贡献代码！请遵循以下步骤：

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

## 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件。

## 作者

- **kitsch-9527** - *初始作者* - [GitHub](https://github.com/kitsch-9527)

## 致谢

- 感谢 Go 社区提供的优秀工具和库
- 感谢所有为这个项目做出贡献的开发者

---

如果您觉得这个项目对您有帮助，请给我们一个 ⭐️！