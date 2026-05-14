# wcorefx 架构优化计划

> 2026-05-14
> 定位：内部工具库，先消除重复模式，不为抽象而加抽象

---

## 已完成

### 迭代 1（2026-05-14）

#### P0：`internal/winapi` 基础设施包

```
internal/winapi/
├── proc.go         — Proc{Call, CallRet} 自动处理三种调用约定
├── buffer.go       — BufferStrategy + BufferQuery + FuncStrategy
└── errors.go       — ErrInsufficientBuffer + IsErrInsufficientBuffer
```

三个核心抽象：

1. **`winapi.Proc`** — 封装 DLL 加载与 `syscall.SyscallN`，自动处理调用约定差异
   - `ConvWin32`（默认，r1 == 0 表示失败）
   - `ConvErrnoReturn`（r1 直接是错误码）
   - `ConvNTSTATUS`（int32(r1) < 0 表示错误）
   - 调用方始终写 `proc.Call(args...)` 或 `proc.CallRet(args...)`，纯 Go error 语义

2. **`winapi.BufferQuery`** — 自动处理两段式 buffer 查询
   - 接受 `BufferStrategy` 接口（单方法 `Fill([]byte) (int, error)`）
   - 自动 4096 起始 → 检测 `ErrInsufficientBuffer` → 扩容重试（最多 5 次）
   - API 特定的 buffer-too-small 信号由 `Fill` 函数自己判断并转为 `ErrInsufficientBuffer`

3. **`winapi.ErrInsufficientBuffer`** — 统一的 buffer 不足信号
   - `IsErrInsufficientBuffer(err) bool` 同时识别自定义类型和系统 errno

### 试点迁移：`netapi` 包

**`iphlpapi.go`**：64 行 → 13 行。移除 3 个 wrapper 函数。
**`fwpuclnt.go`**：函数体简化，直接用 `proc.Call`。
**`network.go`**：所有两段式查询改为 `BufferQuery`，所有 `SyscallN` 移除。

---

## 剩余架构候选（按优先级）

### P1：候选 3 推广 — `internal/winapi` 覆盖其余包

将 `winapi.Proc` + `BufferQuery` 推广到所有包，消除全部 `syscall.SyscallN` 调用。

**涉及包与改动量评估：**

| 包 | 文件 | proc 数 | 两段式查询 | 复杂度 | 状态 |
|----|------|---------|-----------|--------|------|
| `obj` | `ntdll.go` + `psapi.go` | 6 | 2 | 低 | ✅ |
| `ps` | `ntdll.go` + `psapi.go` | 5 | 1 | 低 | ✅ |
| `sec` | `advapi32.go` + `ntdll.go` | 6 | 3 | 低 | ✅ |
| `svc` | `winapi.go` | 4 | 2 | 低 | ✅ |
| `os` | `kernel32.go` + `user32.go` + `wtsapi32.go` | 12 | 0 | 低 | ✅ |
| `evtx` | `wevtapi.go` | 24 | 4 | 高 | ✅ |
| `event` | — | 0 | 0 | 无（纯数据模型） | - |
| `dns` | `winapi.go` | 1 | 0 | 低 | ✅ |
| `fs` | — | 0 | 0 | 无 | - |
| `reg` | — | 0 | 0 | 无（使用 `x/sys/windows/registry`） | - |
| `task` | — | 0 | 0 | 无 | - |
| `wmi` | — | 0 | 0 | 无（使用 go-ole COM） | - |

**策略：按依赖顺序逐步迁移**

```
Phase 1（无内部依赖）: obj → ps → sec → svc → dns
Phase 2（依赖 chain） : os（基础）→ obj→ps（依赖 os）
Phase 3（最复杂）    : evtx（24 个 proc，4 个两段式，错误码最多）
Phase 4（清理）      : 删除 event.InsufficientBufferError，统一为 winapi
```

每个 Phase 的可验证出口：`go test ./<pkg>/ -count=1` 全绿。

---

### P2：候选 2 — 消除 `ps → obj → os` 依赖链

`ps/process.go` 中 `Path()` 函数的 fallback 路径调用了 `obj.NativePathToDosPath()`，造成 `ps → obj → os` 链。

三种方案：

| 方案 | 工作量 | 风险 | 收益 |
|------|--------|------|------|
| **A：将函数下沉到 `os`** | 小（移 1 个函数） | 低 | 依赖链变为 `ps → os`，更直接 |
| **B：`ps` 内联实现** | 小（复制 ~20 行） | 中（重复逻辑） | 零内部依赖 |
| **C：提取到独立 `ntpath` 包** | 中 | 低 | 最干净但新增一个包 |

推荐方案 A。`NativePathToDosPath` 本身就和系统路径相关，放在 `os` 中语义正确。
状态：**已完成**。已迁移到 `os/path.go`，`ps` 改为 import `os`。（✅）

---

### P3：候选 5 — Session 模式（按需，非全面改造）

现有 session 模式：

| 包 | 已有模式 | 状态 |
|----|---------|------|
| `wmi` | `Connect → Query → Close` | ✅ 已有 |
| `evtx` | `NewReader → Open/Read → Close` | ✅ 已有 |

观察：`netapi` 中 `WfpCallouts()` / `WfpFilters()` 每次建立新的 WFP session，频繁调用时效率低。但当前无性能投诉，不动。

候选入口：如果后续出现高频 WFP 查询需求，为 `netapi` 加 `WfpSession`（
`Open → Callouts/Filters → Close`）。

---

### P4：候选 4 — event 包合并到 evtx

`event` 是纯数据模型，目前只有一个消费者 `evtx`。

| 做法 | 收益 | 成本 |
|------|------|------|
| 保留独立 | 可能的外部消费者可独立 import | 增加接口面积 |
| 合并到 `evtx` | `evtx` 自包含，减少包数量 | 破坏依赖 |
| **搁置** | **不花时间** | **无** |

推荐：**搁置**。内部库不需要对外暴露纯净的 event 模型，但合并没有显著维护收益，
不值得在此时投入。等 `evtx` 有实质性重构需要时再做。

---

### P5：候选 1 — 接口（永久搁置）

内部工具库不需要接口。`# Already Done` 标记：
- 现有集成测试保证功能正确
- 无外部消费者需要 mock
- `BufferStrategy` 已经是接口，内部测试点已有覆盖

---

## 迁移模板

### 标准 DLL proc 迁移

```go
// 改造前
var modFoo = windows.NewLazySystemDLL("foo.dll")
var procBar = modFoo.NewProc("Bar")
func Bar() error {
    r1, _, e1 := syscall.SyscallN(procBar.Addr())
    if r1 == 0 { return e1 }
    return nil
}

// 改造后
var procBar = winapi.NewProc("foo.dll", "Bar")
// 调用方: err := procBar.Call(args...)
// 不再需要包装函数
```

### 两段式 buffer 查询迁移

```go
// 改造前
func queryBuffer() ([]byte, uint32, error) {
    var needed uint32
    err := someAPI(nil, 0, &needed)
    if err != windows.ERROR_INSUFFICIENT_BUFFER { return nil, 0, err }
    buf := make([]byte, needed)
    err = someAPI(&buf[0], needed, &used)
    if err != nil { return nil, 0, err }
    return buf, used, nil
}

// 改造后
func queryBuffer() ([]byte, error) {
    return winapi.BufferQuery(winapi.FuncStrategy(func(buf []byte) (int, error) {
        var size uint32 = uint32(len(buf))
        p := (*byte)(nil)
        if len(buf) > 0 { p = &buf[0] }
        err := procSomeAPI.Call(
            uintptr(unsafe.Pointer(p)),
            uintptr(unsafe.Pointer(&size)),
        )
        if err != nil && winapi.IsErrInsufficientBuffer(err) {
            return int(size), &winapi.ErrInsufficientBuffer{Size: int(size)}
        }
        return int(size), err
    }))
}
```

### ConvErrnoReturn 迁移

```go
var procGetFoo = winapi.NewProc("foo.dll", "GetFoo", winapi.ConvErrnoReturn)

// 自动：r1 != 0 → syscall.Errno(r1)
err := procGetFoo.Call(args...)
```

---

## 变更清单（按实施顺序）

```mermaid
gantt
    title 架构优化实施计划
    dateFormat  YYYY-MM-DD
    section Phase 1
    obj 迁移           :p1a, 1d
    ps 迁移            :p1b, 1d
    sec 迁移           :p1c, 1d
    svc 迁移           :p1d, 1d
    dns 迁移           :p1e, 0.5d
    section Phase 2
    os 迁移            :p2a, 1d
    依赖链修复          :p2b, 0.5d
    section Phase 3
    evtx 迁移          :p3a, 2d
    共享错误清理        :p3b, 0.5d
```

每个 Phase 完成后全量测试：`go test ./... -count=1`。WMI 测试的环境依赖问题需单独处理。
