//go:build windows

package winapi

import (
	"errors"
	"strconv"
	"syscall"
)

// ErrInsufficientBuffer signals BufferQuery to grow and retry.
// Size is the suggested buffer size; 0 means "double and retry".
type ErrInsufficientBuffer struct {
	// Size is the suggested buffer size; 0 means "double and retry".
	Size int
	// Cause is the original error that caused the insufficient buffer condition, if any.
	Cause error
}

func (e *ErrInsufficientBuffer) Error() string {
	if e.Cause != nil {
		return e.Cause.Error()
	}
	if e.Size > 0 {
		return "insufficient buffer: need " + strconv.Itoa(e.Size)
	}
	return "insufficient buffer"
}

// RequiredSize returns the required buffer size from an ErrInsufficientBuffer.
// Returns 0 if the error is not an *ErrInsufficientBuffer.
func RequiredSize(err error) int {
	var ib *ErrInsufficientBuffer
	if errors.As(err, &ib) {
		return ib.Size
	}
	return 0
}

// IsErrInsufficientBuffer reports whether err is or wraps an ErrInsufficientBuffer
// or a system ERROR_INSUFFICIENT_BUFFER errno.
func IsErrInsufficientBuffer(err error) bool {
	if err == nil {
		return false
	}
	var ib *ErrInsufficientBuffer
	if errors.As(err, &ib) {
		return true
	}
	return errors.Is(err, syscall.Errno(122))
}
