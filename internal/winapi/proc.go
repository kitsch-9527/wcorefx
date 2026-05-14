//go:build windows

package winapi

import (
	"syscall"

	"golang.org/x/sys/windows"
)

// Convention specifies how a Windows API returns errors.
type Convention int

const (
	ConvWin32        Convention = iota // r1 == 0 signals error, use GetLastError()
	ConvErrnoReturn                     // r1 is the error code (0 = success)
	ConvNTSTATUS                        // int32(r1) < 0 signals error
)

// Proc wraps a Windows DLL procedure, normalising call conventions into Go error semantics.
type Proc struct {
	addr uintptr
	conv Convention
}

// NewProc loads a system DLL procedure and returns a caller-ready Proc.
// conv defaults to ConvWin32 if omitted.
func NewProc(dll, name string, conv ...Convention) *Proc {
	p := &Proc{
		addr: windows.NewLazySystemDLL(dll).NewProc(name).Addr(),
	}
	if len(conv) > 0 {
		p.conv = conv[0]
	}
	return p
}

// Call invokes the procedure with the given arguments.
func (p *Proc) Call(args ...uintptr) error {
	_, err := p.call(args)
	return err
}

// CallRet invokes the procedure and returns the r1 value alongside the error.
// Use when the caller needs the return value (e.g. an EvtHandle).
func (p *Proc) CallRet(args ...uintptr) (uintptr, error) {
	return p.call(args)
}

func (p *Proc) call(args []uintptr) (uintptr, error) {
	r1, _, e1 := syscall.SyscallN(p.addr, args...)
	switch p.conv {
	case ConvErrnoReturn:
		if r1 != 0 {
			return r1, syscall.Errno(r1)
		}
	case ConvNTSTATUS:
		if int32(r1) < 0 {
			return r1, syscall.Errno(r1)
		}
	default: // ConvWin32
		if r1 == 0 {
			if e1 != 0 {
				return 0, e1
			}
			return 0, syscall.EINVAL
		}
	}
	return r1, nil
}
