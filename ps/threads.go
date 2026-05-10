//go:build windows

package ps

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

// ThreadInfo 保存线程信息
type ThreadInfo struct {
	// ID 线程ID
	ID uint32
	// OwnerPID 所属进程ID
	OwnerPID uint32
	// BasePri 线程基础优先级
	BasePri int32
}

// Threads 返回指定进程的所有线程
//   pid - 进程ID
//   返回 - 线程信息列表
//   返回 - 错误信息
func Threads(pid uint32) ([]ThreadInfo, error) {
	h, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPTHREAD, pid)
	if err != nil {
		return nil, fmt.Errorf("CreateToolhelp32Snapshot failed: %w", err)
	}
	defer windows.CloseHandle(h)

	var te windows.ThreadEntry32
	te.Size = uint32(unsafe.Sizeof(te))
	err = windows.Thread32First(h, &te)
	if err != nil {
		return nil, fmt.Errorf("Thread32First failed: %w", err)
	}

	var threads []ThreadInfo
	for {
		if te.OwnerProcessID == pid {
			threads = append(threads, ThreadInfo{
				ID:           te.ThreadID,
				OwnerPID:     te.OwnerProcessID,
				BasePri: te.BasePri,
			})
		}
		err = windows.Thread32Next(h, &te)
		if err != nil {
			break
		}
	}
	return threads, nil
}
