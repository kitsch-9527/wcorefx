package comm

import "syscall"

// 只为常见的Errno值执行一次接口分配
const (
	errnoERROR_IO_PENDING = 997 // IO操作挂起的错误码
)

var (
	// 预定义IO操作挂起错误，避免运行时重复分配
	errERROR_IO_PENDING error = syscall.Errno(errnoERROR_IO_PENDING)
	// 预定义无效参数错误，避免运行时重复分配
	errERROR_EINVAL error = syscall.EINVAL
)

// ErrnoErr 返回已封装的常见Errno错误值，以避免运行时内存分配
// 参数e为系统调用返回的错误码
// 返回值为对应的预定义错误实例或原始错误码
func ErrnoErr(e syscall.Errno) error {
	switch e {
	case 0:
		return errERROR_EINVAL // 当错误码为0时返回预定义的无效参数错误
	case errnoERROR_IO_PENDING:
		return errERROR_IO_PENDING // 当错误码为IO挂起时返回预定义的对应错误
	}
	// TODO: 在收集Windows上常见错误值的数据后，在这里添加更多case
	// （或许可以在运行all.bat时收集？）
	return e // 对于未预定义的错误码，直接返回原始错误
}
