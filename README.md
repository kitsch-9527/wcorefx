# WCoreFx - Windows Core Framework for Go

[![Go Version](https://img.shields.io/badge/Go-1.20+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

## 简介

WCoreFx 是一个专为 Go 语言设计的 Windows 系统核心框架库。"W" 代表 Windows，"Core" 对应核心（kernel），"Fx" 代表框架（Framework）。

该库提供了 Go 标准库之外的 Windows 系统 API 封装，涵盖系统信息、进程管理、安全权限、文件版本、网络端点、注册表操作、事件日志等多个领域。

## 包功能总览

| 包 | 功能领域 | 核心能力 |
|----|---------|---------|
| [os](#os--系统信息模块) | 系统信息 | 版本、用户、目录、环境变量、系统运行时间、关机控制 |
| [ps](#ps--进程模块) | 进程管理 | 进程枚举、命令行、内存、时间、模块、令牌查询 |
| [sec](#sec--安全模块) | 安全权限 | 管理员检测、令牌提权、权限调整、签名验证、SID 查询 |
| [fs](#fs--文件模块) | 文件信息 | 文件时间戳、版本资源信息 |
| [obj](#obj--对象模块) | 内核对象 | 设备驱动枚举、NT 路径转换 |
| [reg](#reg--注册表模块) | 注册表 | 注册表路径检查、值查询（支持 HKLM/HKCU 等根键缩写） |
| [netapi](#netapi--网络模块) | 网络 | TCP/UDP 端点查询、WFP 调用/过滤枚举、地址转换 |
| [event](#event--事件类型模块) | 事件类型 | Windows Event Log XML 解析、SID 解析、WinMeta 元数据 |
| [evtx](#evtx--事件日志读取模块) | 事件读取 | 实时订阅、历史查询、日志通道枚举、书签、渲染 |

---

## os — 系统信息模块

提供 Windows 操作系统版本、用户、目录、环境变量和运行状态信息。

```go
import "github.com/kitsch-9527/wcorefx/os"
```

| 函数 | 说明 |
|------|------|
| `Is64()` | 返回操作系统是否为 64 位 |
| `IsVistaUpper()` | 返回系统版本是否为 Vista 或更高 |
| `MajorVersion()` | 返回内核主版本号 |
| `MinorVersion()` | 返回内核副版本号 |
| `BuildNumber()` | 返回内核构建号 |
| `ReleaseID()` | 返回 Windows 发行标识符（如 "22H2"） |
| `VersionInfo()` | 返回人类可读的 Windows 版本字符串（如 "Windows 11"） |
| `CPUCount()` | 返回逻辑处理器数量 |
| `TickCount()` | 返回系统运行时间（毫秒） |
| `StartupTime()` | 返回系统启动时间 |
| `NetBiosName()` | 返回 NetBIOS 计算机名 |
| `HostName()` | 返回 DNS 主机名 |
| `UserName()` | 返回当前用户完整域名\用户名 |
| `SessionUserName(sessionID)` | 返回指定会话 ID 的用户名 |
| `WinDir()` | 返回 Windows 目录（如 C:\Windows） |
| `SystemDir()` | 返回系统目录（64 位系统返回 SysWOW64） |
| `System32Dir()` | 返回 System32 目录 |
| `Syswow64Dir()` | 返回 SysWOW64 目录 |
| `Getenv(name)` | 返回环境变量值（支持 %PATH% 展开） |
| `Environ()` | 返回全部环境变量（map[string]string） |
| `DosErrorMsg(code)` | 返回 Windows 错误码对应的描述文本 |
| `Reboot()` | 重启系统（需 SE_SHUTDOWN_NAME 权限） |
| `Poweroff()` | 关闭系统（需 SE_SHUTDOWN_NAME 权限） |

---

## ps — 进程模块

提供进程枚举、查询和诊断信息。

```go
import "github.com/kitsch-9527/wcorefx/ps"
```

| 函数 | 说明 |
|------|------|
| `List()` | 返回所有正在运行的进程快照 |
| `Find(name)` | 根据可执行文件名模糊匹配查找进程 |
| `CommandLine(pid)` | 返回指定 PID 进程的命令行参数 |
| `MemoryInfo(pid)` | 返回进程内存计数器信息 |
| `Times(pid)` | 返回进程创建/退出/内核/用户时间 |
| `Path(pid)` | 返回进程可执行文件完整路径 |
| `User(pid)` | 返回进程所属域\用户名 |
| `IsTerminated(pid)` | 检查进程是否已终止 |
| `ParentID(pid)` | 返回父进程 ID |
| `SessionID(pid)` | 返回进程所属会话 ID |
| `Modules(pid)` | 返回进程加载的模块列表 |

---

## sec — 安全模块

提供权限提升、令牌操作、数字签名验证等安全相关功能。

```go
import "github.com/kitsch-9527/wcorefx/sec"
```

| 函数 | 说明 |
|------|------|
| `IsAdmin()` | 检测当前进程是否以管理员权限运行 |
| `TokenElevation(procHandle)` | 获取指定进程的令牌提权状态 |
| `EnableDebugPrivilege(useNative)` | 启用 SeDebugPrivilege（API 或原生路径） |
| `DisableDebugPrivilege(useNative)` | 禁用 SeDebugPrivilege |
| `EnablePrivilege(name, number)` | 启用指定权限（名称或编号方式） |
| `DisablePrivilege(name, number)` | 禁用指定权限 |
| `GetTokenGroupsAndPrivileges(token)` | 获取令牌的组和权限信息 |
| `GetTokenPrivilegeNames(groups)` | 获取令牌中权限的名称和状态列表 |
| `GetTokenAccountSIDs(type, count, sids)` | 获取令牌中 SID 对应的账户名列表 |
| `LookupSIDAccount(sid)` | 将 SID 解析为域名\账户名 |
| `LookupPrivilegeNameByLUID(luid)` | 将 LUID 权限值解析为权限名称 |
| `GetTokenInformation(token, class)` | 获取令牌的原始信息数据 |
| `VerifyFileSignature(path)` | 验证文件的数字签名（WinVerifyTrust） |
| `GetDomainJoinInfo()` | 获取系统域加入状态信息 |
| `CheckTokenMembership(handle, sid)` | 检查令牌是否属于指定 SID 组 |
| `RtlAdjustPrivilege(privilege, enable, current)` | 通过 ntdll 原生 API 调整权限 |
| `FormatSIDAttributes(label, attr)` | 格式化 SID 属性标志为可读文本 |
| `FormatPrivilegeStatus(attr)` | 格式化权限属性标志为可读文本 |
| `FormatJoinStatus(status)` | 格式化域加入状态为可读文本 |

---

## fs — 文件模块

提供文件时间戳和版本资源信息查询。

```go
import "github.com/kitsch-9527/wcorefx/fs"
```

| 函数 | 说明 |
|------|------|
| `CreateTime(path)` | 返回文件创建时间（Unix 时间戳） |
| `AccessTime(path)` | 返回文件最后访问时间（Unix 时间戳） |
| `ModifyTime(path)` | 返回文件最后修改时间（Unix 时间戳） |
| `VersionInfo(path)` | 返回文件版本字符串（major.minor.build.revision） |
| `Info(path, infoType, translation)` | 返回文件版本资源中指定的信息字段 |

支持的版本信息类型（`InfoType`）：
- `FileDescription`、`CompanyName`、`OriginalFileName`
- `LegalCopyright`、`ProductName`、`ProductVersion`

---

## obj — 对象模块

提供内核对象查询和 NT 路径转换功能。

```go
import "github.com/kitsch-9527/wcorefx/obj"
```

| 函数 | 说明 |
|------|------|
| `DriverList()` | 返回所有已加载设备驱动的基础地址列表 |
| `DriverName(driver)` | 返回指定驱动地址对应的驱动文件名 |
| `DriverPath(driver)` | 返回指定驱动地址对应的驱动文件路径 |
| `NativePathToDosPath(nativePath)` | 将 NT 原生路径转换为 DOS 路径（如 \SystemRoot → C:\Windows） |

---

## reg — 注册表模块

提供 Windows 注册表查询功能，支持所有标准根键缩写。

```go
import "github.com/kitsch-9527/wcorefx/reg"
```

支持的根键：`HKLM` / `HKCU` / `HKU` / `HKCR` / `HKCC`（含全称和缩写）

| 函数 | 说明 |
|------|------|
| `CheckPath(path)` | 检查注册表路径是否存在 |
| `GetValue(path, key)` | 获取注册表指定路径下键的字符串值 |

---

## netapi — 网络模块

提供 TCP/UDP 端点查询和 Windows Filtering Platform 枚举。

```go
import "github.com/kitsch-9527/wcorefx/netapi"
```

| 函数 | 说明 |
|------|------|
| `Tcp4Endpoints()` | 返回所有 IPv4 TCP 端点 |
| `Tcp6Endpoints()` | 返回所有 IPv6 TCP 端点 |
| `Udp4Endpoints()` | 返回所有 IPv4 UDP 端点 |
| `Udp6Endpoints()` | 返回所有 IPv6 UDP 端点 |
| `TCPState(state)` | 将 TCP 状态码转换为可读字符串 |
| `InetNtoa(addr)` | 将 uint32 网络字节序地址转为 IPv4 字符串 |
| `InetNtoa6(addr)` | 将 16 字节数组转为 IPv6 字符串 |
| `Ntohs(port)` | 将网络字节序端口转为主机字节序 |
| `WfpCallouts()` | 枚举 WFP 标注（Callout） |
| `WfpFilters()` | 枚举 WFP 过滤规则 |

WFP 底层函数（`fwpuclnt.go`）：
- `FwpmEngineOpen/Close` — 打开/关闭 WFP 引擎会话
- `FwpmCalloutCreateEnumHandle/Enum/GetByKey` — 标注枚举与查询
- `FwpmFilterCreateEnumHandle/Enum/GetByKey` — 过滤规则枚举与查询
- `FwpmFreeMemory` — 释放 WFP 内存

---

## event — 事件类型模块

提供 Windows Event Log XML 解析、SID 解析和元数据映射（WinMeta）。

```go
import "github.com/kitsch-9527/wcorefx/event"
```

| 类型 | 说明 |
|------|------|
| `Event` | 事件完整结构，包含 System/EventData/RenderingInfo 等 |
| `Record` | `Event` 包装，附加 API 来源和原始 XML |
| `Provider` | 事件提供商标识（Name + GUID） |
| `Execution` | 进程 ID、线程 ID、会话 ID 等执行上下文 |
| `SID` / `SIDType` | 安全标识符及其类型枚举 |
| `WinMeta` | 事件元数据（Keywords/Opcodes/Levels/Tasks）映射 |

| 函数 | 说明 |
|------|------|
| `UnmarshalXML(rawXML)` | 解析原始 XML 为 `Event` 结构体 |
| `EnrichRawValuesWithNames(meta, event)` | 将原始值（KeywordsRaw/LevelRaw 等）填充为人可读名称 |
| `PopulateAccount(sid)` | 通过系统查找填充 SID 的账户名和类型 |
| `RemoveWindowsLineEndings(s)` | 将 CRLF 替换为 LF 并去除尾部换行 |
| `Format(records, fn)` | 将 `Record` 列表格式化为精简输出结构 |

等级原值映射（LevelRaw → Level）：0=Information、1=Critical、2=Error、3=Warning、4=Information、5=Verbose

---

## evtx — 事件日志读取模块

提供 Windows Event Log 的实时订阅、历史查询、日志通道枚举和 XML 渲染。

```go
import "github.com/kitsch-9527/wcorefx/evtx"
```

| 函数 | 说明 |
|------|------|
| `NewReader(target, eventID)` | 创建事件日志读取器（支持通道名或 .evtx 文件） |
| `Reader.Open(recordNumber)` | 打开日志开始读取 |
| `Reader.Read()` | 读取一批事件记录 |
| `Reader.Close()` | 关闭读取器 |
| `GetEvents(target, eventID)` | 一次性查询获取事件记录列表 |
| `Subscribe(session, signal, path, query, bookmark, flags)` | 实时订阅事件日志 |
| `EventHandles(subscription, max)` | 从订阅中获取事件句柄 |
| `RenderEventXML(eventHandle, buf, out)` | 将事件句柄渲染为 XML |
| `RenderEvent(eventHandle, lang, buf, out)` | 使用指定语言渲染事件 |
| `FormatEventString(flag, handle, ...)` | 格式化事件消息字符串 |
| `IsAvailable()` | 检测 wevtapi 是否可用 |
| `EvtQuery(session, path, query, flags)` | 执行事件日志查询 |
| `EvtOpenLog(session, path, flags)` | 打开事件日志 |
| `Channels()` | 枚举所有事件日志通道 |
| `Publishers()` | 枚举所有事件发布者 |
| `OpenPublisherMetadata(session, name, lang)` | 打开发布者元数据 |
| `CreateBookmarkFromXML(xml)` | 从 XML 创建书签 |
| `CreateRenderContext(paths, flag)` | 创建渲染上下文 |
| `NewByteBuffer(initialSize)` | 创建动态字节缓冲区（用于渲染） |
| `EvtVariantData(variant, buf)` | 将 `EvtVariant` 解析为 Go 类型 |

支持的查询标志：`EvtQueryReverseDirection`、`EvtQueryTolerateQueryErrors` 等
支持的订阅标志：`EvtSubscribeToFutureEvents`、`EvtSubscribeStartAtOldestRecord`、`EvtSubscribeStartAfterBookmark` 等

---

## 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件。
