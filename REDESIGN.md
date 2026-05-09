# WCoreFx 重设计划

## 目标

KISS + Unix 原则重构，参考 UNONE/KNONE (BlackINT3/none) 架构，清理依赖，扁平化结构。

---

## 1. 设计决策总表

| # | 决策 | 结论 |
|---|------|------|
| 1 | beats 依赖 | **去掉**，自建精简 map[string]interface{} |
| 2 | 包结构 | **平铺**，无嵌套层级 |
| 3 | syscall 层 | **取消独立 sys/ 子包**，高层包直接 import x/sys/windows，私有 gap-filling syscall 按 DLL 分文件放在各自包内 |
| 4 | API 风格 | **PID 容器 + 无状态方法**，每方法独立开/关 handle |
| 5 | event 拆分 | `event/`（类型+XML）+ `evtx/`（读取器） |
| 6 | 命名 | **全英文**，去掉 `Get`/`Proc`/`Ps` 前缀 |
| 7 | `common/` | **解散**，内容分别迁入相关包 |
| 8 | `internal/` | **删除** |
| 9 | `winapi/` | **删除**，syscall 移入各高层包 |
| 10 | 枚举回调 | **不支持提前终止**，返回切片 |

---

## 2. 包对照 (Old → New)

### 2.1 删除的包

| 旧包 | 处理 |
|------|------|
| `common/` | `MAXPATH` → 各使用方自行定义；`MapStr` → 删；`exc/` → 删 |
| `common/exc/` | 改用 `fmt.Errorf` |
| `winapi/common/` | `Boo2Ptr` → 内联到唯一调用方 |
| `winapi/dll/advapi32/` | 迁入 `sec/` |
| `winapi/dll/kernel32/` | 迁入 `os/`, `ps/`, `fs/` |
| `winapi/dll/ntdll/` | 迁入 `ps/`, `sec/` |
| `winapi/dll/psapi/` | 迁入 `obj/` |
| `winapi/dll/iphlpapi/` | 迁入 `net/` |
| `winapi/dll/fwpuclnt/` | 迁入 `net/` |
| `winapi/dll/wevtapi/` | 迁入 `evtx/` |
| `winapi/dll/wtsapi32/` | 迁入 `os/` |
| `internal/` | 删除（空壳） |
| `mem/` | 删除（空目录） |

### 2.2 新建/改写的包

#### `ps/` — 进程模块

| 文件 | 来源 | 说明 |
|------|------|------|
| `process.go` | 重写 | `Process{PID}` 类型, `List()`, `Find()`, `CommandLine()`, `MemoryInfo()`, `Times()`, `Path()`, `User()`, `IsTerminated()`, `ParentID()`, `SessionID()`, `Modules()`, `OpenToken()` |
| `ntdll.go` | 重写 | `NtQueryObject`, `NtDuplicateObject`, `GetHandleType`, `GetHandleName`（私有） |

- 参考 UNONE `Ps*` 函数族但不加前缀：`ps.List()` = `PsGetAllProcess`
- `GetOpenFiles` 注释掉的代码不再保留
- 去掉 `treemap` 依赖（emirpasic/gods）

#### `fs/` — 文件模块

| 文件 | 来源 | 说明 |
|------|------|------|
| `file.go` | 改写 | 文件时间/版本信息/资源信息，去掉 beats 依赖 |

- `NativePathToDosPath` → 移入 `obj/path.go`
- 所有函数用 x/sys/windows

#### `netapi/` — 网络模块

| 文件 | 来源 | 说明 |
|------|------|------|
| `network.go` | 改写 | TCP/UDP 端点查询（4/6），去掉冗余边界检查代码 |
| `iphlpapi.go` | 迁入 | `GetExtendedTcpTable`, `GetExtendedUdpTable` |
| `fwpuclnt.go` | 迁入 | `FwpmEngineOpen`, `FwpmCalloutEnum`, `FwpmFilterEnum` 等 |

#### `sec/` — 安全模块

| 文件 | 来源 | 说明 |
|------|------|------|
| `security.go` | 合并改写 | `IsAdmin()`, `TokenElevation()`, `LookupSIDAccount()`, `VerifyFileSignature()`, `EnableDebugPrivilege()`, `DisableDebugPrivilege()`, `EnablePrivilege()`, `DisablePrivilege()`, `GetTokenGroupsAndPrivileges()`, `GetTokenPrivilegeNames()`, `GetDomainJoinInfo()` |
| `ntdll.go` | 迁入 | `RtlAdjustPrivilege` |
| `format.go` | 保留 | `FormatSIDAttributes`, `FormatPrivilegeStatus`, `FormatJoinStatus` |

- 统一返回 `error`，去掉 `fmt.Println` 内部打印
- `DisplayTokenAccount`（空实现）→ 删
- `exc.New` → `fmt.Errorf`

#### `os/` — 系统信息模块

| 文件 | 来源 | 说明 |
|------|------|------|
| `os.go` | 扩充 | 版本/用户/目录/环境，用 stdlib + x/sys/windows |
| `wtsapi.go` | 新建 | `WTSQuerySessionInformation`（私有） |
| `kernel32.go` | 新建 | `GetTickCount64`（私有） |
| `user32.go` | 新建 | `ExitWindowsEx`（私有） |

**函数清单：**

```go
func Is64() bool
func IsVistaUpper() bool
func MajorVersion() uint32
func MinorVersion() uint32
func BuildNumber() uint32
func ReleaseID() string
func VersionInfo() string
func CPUCount() uint32
func TickCount() uint64
func StartupTime() time.Time
func NetBiosName() (string, error)
func HostName() (string, error)
func UserName() (string, error)
func SessionUserName(sessionID uint32) (string, error)
func WinDir() (string, error)
func SystemDir() (string, error)
func System32Dir() (string, error)
func Syswow64Dir() (string, error)
func Getenv(name string) string
func Environ() map[string]string
func DosErrorMsg(errCode uint32) string
func Reboot() error
func Poweroff() error
```

#### `obj/` — 对象模块

| 文件 | 来源 | 说明 |
|------|------|------|
| `object.go` | 保留改写 | `DriverList()`, `DriverPath()`, `DriverName()` → 去 `Get` 前缀 |
| `path.go` | 迁入 | `NativePathToDosPath`（从 old `fs/` 迁入） |

- 参考 UNONE `Ob*` 函数族

#### `reg/` — 注册表模块

| 文件 | 来源 | 说明 |
|------|------|------|
| `registry.go` | 改写 | 用 `x/sys/windows/registry` |

- 去掉 `TestRegPath`（测试代码混入生产）
- `ExportRegPathToFile` 空实现 → 删

#### `event/` — 事件类型

| 文件 | 来源 | 说明 |
|------|------|------|
| `types.go` | 迁入 | `Event`, `Provider`, `Execution`, `EventIdentifier`, `SID`, `SIDType`, `WinMeta` 等结构体 |
| `xml.go` | 迁入 | `UnmarshalXML`, `EnrichRawValuesWithNames`，去掉 beats 依赖 |

- `Fields()` → 改为纯 `map[string]interface{}` 返回，不依赖 beats `MapStr`/`AddOptional`
- `AddPairs`, `AddOptional`, `isZero`, `debugf` → 全部去掉
- 去掉 beats license header（重写后不存在 beats 代码）

#### `evtx/` — 事件日志读取

| 文件 | 来源 | 说明 |
|------|------|------|
| `reader.go` | 迁入改写 | `winEventLog` 读取器 |
| `syscall.go` | 迁入 | wevtapi 全部（`EvtExportLog`, `EvtSubscribe`, `EvtRender`, `EvtClose` 等） |
| `buffer.go` | 迁入 | `ByteBuffer`，去掉 `PtrAt` 等未使用的辅助方法 |

- 去掉 x/sys/windows 未包含的 WMI 查询（当前无用）

---

## 3. 依赖变更

### 移除

| 依赖 | 原因 |
|------|------|
| `github.com/elastic/beats/v7` | 独立自实现 |
| `github.com/emirpasic/gods` | `treemap` 不再使用 |
| `github.com/cespare/xxhash/v2` | beats 传递依赖 |
| `github.com/elastic/go-sysinfo` | beats 传递依赖 |
| `github.com/elastic/go-ucfg` | beats 传递依赖 |
| `github.com/elastic/go-windows` | beats 传递依赖 |
| `github.com/magefile/mage` | beats 传递依赖 |
| `github.com/pkg/errors` | 改用 `fmt.Errorf` |
| `go.uber.org/zap` + `multierr` + `atomic` | beats 传递依赖 |
| `gopkg.in/yaml.v2` | beats 传递依赖 |
| `howett.net/plist` | beats 传递依赖 |

### 保留

| 依赖 | 原因 |
|------|------|
| `golang.org/x/sys` | 核心 Windows API 绑定 |
| `github.com/prometheus/procfs` | 如使用（检查当前 import） |

---

## 4. 实现顺序

| 阶段 | 包 | 原因 |
|------|----|------|
| 1 | `os/` | 无外部依赖，底层基础 |
| 2 | `obj/` | 仅有 psapi syscall |
| 3 | `ps/` | 用 os, obj，中层 |
| 4 | `fs/` | 独立，薄封装 |
| 5 | `sec/` | 独立，薄封装 |
| 6 | `reg/` | 直接用 x/sys/windows/registry |
| 7 | `event/` + `evtx/` | 最复杂，最后 |
| 8 | `netapi/` | 独立，最后 |
| 9 | go.mod 清理 | 移除所有 beats 传递依赖 |
| 10 | 测试 | `ps/process_test.go` 适配新 API |

---

## 5. 待决策项

- [x] `net/` 包名与标准库冲突 → **改用 `netapi`**
